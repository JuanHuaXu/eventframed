package observationgate

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

func reserveRunArm(tape [learnedSteps]learnedTapeTick, base *[512]float64,
	meanUncertainty float64, policy string, rng *rand.Rand, correctAlternative, delay int) (learnedArm, error) {
	r := learnedArm{Policy: policy}
	state := newLearnedState()
	var due [learnedSteps][]int
	for clock, event := range tape {
		if clock == 384 && r.Nominated != 80 {
			return r, errors.New("first reserve stage did not spend exactly 80 requests")
		}
		pb := base[event.X]
		p := state.forecast(event.X, pb)
		if p <= 0 || p >= 1 || math.IsNaN(p) {
			return r, errors.New("invalid reserve working forecast")
		}
		alt := state.bestAlternative()
		tick := learnedArmTick{P: p, Alternate: alt}
		loss := (p - boolFloat(event.Y)) * (p - boolFloat(event.Y))
		r.FullBrier += loss / learnedSteps
		if clock >= 256 {
			r.PostBrier += loss / 256
			if clock < 320 {
				r.EarlyBrier += loss / 64
			}
		}
		target, end := 80, 384
		if clock >= 384 {
			target, end = learnedBudget, learnedSteps
		}
		if learnedNominate(policy, event.X, pb, meanUncertainty, &state,
			target-r.Nominated, end-clock, rng) {
			tick.Selected = true
			r.Nominated++
			if clock >= 256 {
				r.NominatedPost++
				if alt == correctAlternative {
					r.CorrectPost++
				}
			}
			if !event.Missing {
				arrival := clock + delay
				if arrival < learnedSteps {
					due[arrival] = append(due[arrival], clock)
				} else {
					r.Pending++
				}
			}
		}
		for _, origin := range due[clock] {
			if origin > clock || tape[origin].Missing {
				return r, errors.New("future or missing reserve label")
			}
			r.Arrived++
			tick.Delivered = append(tick.Delivered, origin)
			if err := state.observe(tape[origin].X, tape[origin].Y, base); err != nil {
				return r, err
			}
		}
		r.Ticks = append(r.Ticks, tick)
	}
	if r.Nominated != learnedBudget || r.Arrived+r.Pending > learnedBudget {
		return r, errors.New("invalid reserve acquisition accounting")
	}
	return r, nil
}

type reserveRow struct {
	Kind, Split, Scenario string
	Fit, Stream           int
	Tape                  [learnedSteps]v3Tick
	Arms                  [6]transferArm
}

func TestLearnedContrastReserve(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_RESERVE_OUT"), os.Getenv("EVENTFRAME_RESERVE_SPLIT")
	if path == "" && split == "" {
		t.Skip("opt-in research study")
	}
	if path == "" || (split != "design" && split != "confirmation") {
		t.Fatal("output path and design|confirmation split required")
	}
	seed, fitOffset := int64(2026102601), 1600
	if split == "confirmation" {
		seed, fitOffset = 2026102602, 1700
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := make(map[string]string)
	for _, name := range []string{
		"docs/experiments/mmm-learned-contrast-reserve-v1-protocol.md",
		"internal/observationgate/learned_contrast_reserve_test.go",
		"internal/observationgate/learned_contrast_transfer_test.go",
		"internal/observationgate/learned_contrast_v2_test.go",
		"internal/observationgate/learned_contrast_v3_test.go",
		"internal/observationlearners/experiment.go", "internal/observation/forecast.go",
	} {
		b, e := os.ReadFile(filepath.Join(root, name))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer gz.Close()
	enc := json.NewEncoder(gz)
	if err := enc.Encode(map[string]any{"kind": "manifest", "split": split,
		"seedBase": seed, "fitOffset": fitOffset, "trialsPerCase": 256,
		"budget": learnedBudget, "stageBudgets": [2]int{80, 48},
		"stateBytes": unsafe.Sizeof(learnedState{}), "hashes": hashes}); err != nil {
		t.Fatal(err)
	}
	for caseID, c := range transferCases {
		for fit := 0; fit < 16; fit++ {
			baseScenario := observationlearners.Scenarios[0]
			if c.Noise == .20 {
				baseScenario = observationlearners.Scenarios[1]
			}
			base, e := observationlearners.Base(baseScenario, caseID, fit+fitOffset)
			if e != nil {
				t.Fatal(e)
			}
			var probability [512]float64
			meanUncertainty := 0.
			for x := range probability {
				p, e := base.ForecastObserved(511, uint16(x))
				if e != nil || p <= 0 || p >= 1 {
					t.Fatalf("base forecast %s x=%d: %v", c.Name, x, e)
				}
				probability[x] = p
				meanUncertainty += 4 * p * (1 - p) / 512
			}
			for stream := 0; stream < 16; stream++ {
				tape := transferTape(seed, c, caseID, fit, stream)
				var oldTape [learnedSteps]learnedTapeTick
				for clock, event := range tape {
					oldTape[clock] = learnedTapeTick{X: event.X, Y: event.Y, Missing: event.Missing}
				}
				row := reserveRow{Kind: "trial", Split: split, Scenario: c.Name,
					Fit: fit, Stream: stream, Tape: tape}
				correct := -1
				if !c.Stable && !c.OOD {
					correct = 3
				}
				for policy, name := range learnedPolicies {
					selectorSeed := learnedSeed(seed, caseID, fit, stream, 10+policy)
					oldRNG := rand.New(rand.NewSource(selectorSeed))
					old, e := learnedRunArm(oldTape, &probability, meanUncertainty,
						name, oldRNG, correct, c.Delay)
					if e != nil {
						t.Fatalf("old %s/%d/%d/%s: %v", c.Name, fit, stream, name, e)
					}
					row.Arms[policy] = transferMeasure(old, tape, c)
					reserveRNG := rand.New(rand.NewSource(selectorSeed))
					reserved, e := reserveRunArm(oldTape, &probability, meanUncertainty,
						name, reserveRNG, correct, c.Delay)
					if e != nil {
						t.Fatalf("reserve %s/%d/%d/%s: %v", c.Name, fit, stream, name, e)
					}
					row.Arms[policy+3] = transferMeasure(reserved, tape, c)
				}
				if err := enc.Encode(row); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}

func TestLearnedContrastReserveQuota(t *testing.T) {
	var base [512]float64
	for x := range base {
		base[x] = .5
	}
	tape := transferTape(2026102601, transferCases[0], 0, 0, 0)
	var oldTape [learnedSteps]learnedTapeTick
	for clock, event := range tape {
		oldTape[clock] = learnedTapeTick{X: event.X, Y: event.Y, Missing: event.Missing}
	}
	for policy, name := range learnedPolicies {
		rng := rand.New(rand.NewSource(int64(20261026 + policy)))
		arm, err := reserveRunArm(oldTape, &base, 1, name, rng, 3, 0)
		if err != nil {
			t.Fatal(err)
		}
		first := 0
		for _, tick := range arm.Ticks[:384] {
			if tick.Selected {
				first++
			}
		}
		if first != 80 || arm.Nominated != 128 {
			t.Fatalf("%s quota %d/%d", name, first, arm.Nominated)
		}
	}
}
