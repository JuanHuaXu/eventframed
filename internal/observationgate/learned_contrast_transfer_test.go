package observationgate

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

type transferCase struct {
	Name              string
	Change            int
	Noise, Missing    float64
	Delay             int
	Skew, Stable, OOD bool
}

var transferCases = [...]transferCase{
	{"early_bit2", 128, .05, .20, 0, false, false, false},
	{"late_bit2", 384, .05, .20, 0, false, false, false},
	{"noisy_bit2", 256, .20, .20, 0, false, false, false},
	{"delayed_bit2", 256, .05, .25, 32, false, false, false},
	{"skew_bit2", 256, .05, .20, 0, true, false, false},
	{"stable20", 512, .20, .20, 0, false, true, false},
	{"majority_ood", 256, .05, .20, 0, false, false, true},
}

func transferTruth(c transferCase, x uint16, clock int) float64 {
	bit := ((x&(1<<6) != 0) != (x&(1<<7) != 0)) != (x&(1<<8) != 0)
	if clock >= c.Change && !c.Stable {
		if c.OOD {
			count := 0
			for j := 0; j < 3; j++ {
				if x&(1<<j) != 0 {
					count++
				}
			}
			bit = count >= 2
		} else {
			bit = x&(1<<2) != 0
		}
	}
	if bit {
		return 1 - c.Noise
	}
	return c.Noise
}

func transferTape(seed int64, c transferCase, caseID, fit, stream int) [learnedSteps]v3Tick {
	rng := rand.New(rand.NewSource(learnedSeed(seed, caseID, fit, stream, 0)))
	missing := rand.New(rand.NewSource(learnedSeed(seed, caseID, fit, stream, 1)))
	var tape [learnedSteps]v3Tick
	for clock := range tape {
		x := uint16(rng.Intn(512))
		if c.Skew {
			if rng.Float64() < .10 {
				x |= 1 << 2
			} else {
				x &^= 1 << 2
			}
		}
		p := transferTruth(c, x, clock)
		tape[clock] = v3Tick{X: x, Y: rng.Float64() < p,
			Missing: missing.Float64() < c.Missing, PTrue: p}
	}
	return tape
}

type transferArm struct {
	Data                                                    learnedArm
	FullExpected, PostExpected, FirstExpected, PostRealized float64
	RecoveryDelay                                           int
	MissedRecovery                                          bool
}

func transferMeasure(data learnedArm, tape [learnedSteps]v3Tick, c transferCase) transferArm {
	r := transferArm{Data: data, RecoveryDelay: -1}
	var expected [learnedSteps]float64
	postStart := c.Change
	if c.Stable {
		postStart = 256
	}
	for clock, tick := range data.Ticks {
		pTrue := tape[clock].PTrue
		expected[clock] = pTrue*(1-pTrue) + math.Pow(tick.P-pTrue, 2)
		r.FullExpected += expected[clock] / learnedSteps
		if clock >= postStart {
			r.PostExpected += expected[clock] / float64(learnedSteps-postStart)
			r.PostRealized += math.Pow(tick.P-boolFloat(tape[clock].Y), 2) / float64(learnedSteps-postStart)
			if clock < postStart+64 {
				r.FirstExpected += expected[clock] / 64
			}
		}
	}
	if c.Stable || c.OOD || c.Noise >= .20 {
		return r
	}
	consecutive := 0
	for clock := c.Change + 31; clock < learnedSteps; clock++ {
		mean := 0.
		for j := clock - 31; j <= clock; j++ {
			mean += expected[j] / 32
		}
		if mean <= .12 {
			consecutive++
		} else {
			consecutive = 0
		}
		if consecutive == 16 {
			r.RecoveryDelay = clock - c.Change
			return r
		}
	}
	r.RecoveryDelay, r.MissedRecovery = learnedSteps-c.Change, true
	return r
}

type transferRow struct {
	Kind, Split, Scenario string
	Fit, Stream           int
	Tape                  [learnedSteps]v3Tick
	Arms                  [3]transferArm
}

func TestLearnedContrastTransfer(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_TRANSFER_OUT"), os.Getenv("EVENTFRAME_TRANSFER_SPLIT")
	if path == "" && split == "" {
		t.Skip("opt-in research study")
	}
	if path == "" || (split != "design" && split != "confirmation") {
		t.Fatal("output path and design|confirmation split required")
	}
	seed, fitOffset := int64(2026102501), 1200
	if split == "confirmation" {
		seed, fitOffset = 2026102502, 1300
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := make(map[string]string)
	for _, name := range []string{
		"docs/experiments/mmm-learned-contrast-transfer-v1-protocol.md",
		"internal/observationgate/learned_contrast_transfer_test.go",
		"internal/observationgate/learned_contrast_v2_test.go",
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
		"budget": learnedBudget, "stateBytes": unsafe.Sizeof(learnedState{}), "hashes": hashes}); err != nil {
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
				row := transferRow{Kind: "trial", Split: split, Scenario: c.Name, Fit: fit, Stream: stream, Tape: tape}
				correct := -1
				if !c.Stable && !c.OOD {
					correct = 3
				}
				for policy, name := range learnedPolicies {
					rng := rand.New(rand.NewSource(learnedSeed(seed, caseID, fit, stream, 10+policy)))
					data, e := learnedRunArm(oldTape, &probability, meanUncertainty, name, rng, correct, c.Delay)
					if e != nil {
						t.Fatalf("%s/%d/%d/%s: %v", c.Name, fit, stream, name, e)
					}
					row.Arms[policy] = transferMeasure(data, tape, c)
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

func TestLearnedContrastTransferTruth(t *testing.T) {
	if transferTruth(transferCases[0], 0, 0) != .05 ||
		transferTruth(transferCases[0], 4, 128) != .95 ||
		transferTruth(transferCases[5], 0, 511) != .20 ||
		transferTruth(transferCases[6], 7, 256) != .95 ||
		transferTruth(transferCases[6], 1, 256) != .05 {
		t.Fatal("transfer truth branch mismatch")
	}
}
