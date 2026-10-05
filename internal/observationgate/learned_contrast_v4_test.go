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

var v4Prior = [...]float64{.855, .0095, .0095, .0095, .0095, .0095, .0095, .0095, .0095, .0095, .0095, .05}

type v4State struct {
	History [32]learnedEvidence
	N, Next int
	Weight  [12]float64
}

func newV4State() v4State { return v4State{Weight: v4Prior} }

func (s *v4State) bestKnown() int {
	best := 1
	for h := 2; h <= 10; h++ {
		if s.Weight[h] > s.Weight[best] {
			best = h
		}
	}
	return best
}

func (s *v4State) forecast(x uint16, base float64) (float64, bool, error) {
	wMax := 0.
	for h := 0; h <= 10; h++ {
		wMax = math.Max(wMax, s.Weight[h])
	}
	if s.N >= 8 && s.Weight[11] >= 100*wMax {
		return .5, true, nil
	}
	known := 1 - s.Weight[11]
	if known <= 0 || math.IsNaN(known) {
		return 0, false, errors.New("empty known-rule conditional law")
	}
	p := 0.
	for h := 0; h <= 10; h++ {
		p += s.Weight[h] * v3Rule(h, x, base) / known
	}
	if p <= 0 || p >= 1 || math.IsNaN(p) {
		return 0, false, errors.New("invalid gated forecast")
	}
	return p, false, nil
}

func (s *v4State) observe(x uint16, y bool, base *[512]float64) error {
	s.History[s.Next] = learnedEvidence{X: x, Y: y}
	s.Next = (s.Next + 1) % len(s.History)
	if s.N < len(s.History) {
		s.N++
	}
	var logw [12]float64
	maximum := math.Inf(-1)
	for h, prior := range v4Prior {
		logw[h] = math.Log(prior)
		for i := 0; i < s.N; i++ {
			e := s.History[i]
			p := v3Rule(h, e.X, base[e.X])
			if e.Y {
				logw[h] += math.Log(p)
			} else {
				logw[h] += math.Log1p(-p)
			}
		}
		maximum = math.Max(maximum, logw[h])
	}
	total := 0.
	for h := range logw {
		s.Weight[h] = math.Exp(logw[h] - maximum)
		total += s.Weight[h]
	}
	if total <= 0 || math.IsNaN(total) || math.IsInf(total, 0) {
		return errors.New("invalid v4 likelihood normalization")
	}
	for h := range s.Weight {
		s.Weight[h] /= total
	}
	return nil
}

func v4Nominate(policy string, x uint16, base, meanUncertainty float64, alt, remaining, clocks int, rng *rand.Rand) bool {
	if remaining == 0 {
		return false
	}
	if remaining >= clocks {
		return true
	}
	rate := float64(remaining) / float64(clocks)
	switch policy {
	case "uncertainty":
		rate *= 4 * base * (1 - base) / meanUncertainty
	case "learned_disagreement":
		if (learnedRule(alt, x, base) >= .5) == (base >= .5) {
			return false
		}
		rate *= 2
	}
	return rng.Float64() < math.Min(1, rate)
}

type v4Arm struct {
	Data      v3Arm
	GateAt    []bool `json:",omitempty"`
	GateCount int
	FirstGate int
}

func v4RunArm(tape [learnedSteps]learnedTapeTick, base *[512]float64,
	meanUncertainty float64, policy string, rng *rand.Rand, correctAlternative, delay int,
	gated bool) (v4Arm, error) {
	r := v4Arm{FirstGate: -1}
	data := learnedArm{Policy: policy}
	var old v3State
	var current v4State
	selector := newLearnedState()
	if gated {
		current = newV4State()
		r.GateAt = make([]bool, learnedSteps)
	} else {
		old = newV3State()
	}
	var due [learnedSteps][]int
	for clock, event := range tape {
		pb := base[event.X]
		p, gate, alt := 0., false, 0
		if gated {
			var err error
			p, gate, err = current.forecast(event.X, pb)
			if err != nil {
				return r, err
			}
			r.GateAt[clock] = gate
			if gate {
				r.GateCount++
				if r.FirstGate < 0 {
					r.FirstGate = clock
				}
			}
		} else {
			p = old.forecast(event.X, pb)
		}
		alt = selector.bestAlternative()
		tick := learnedArmTick{P: p, Alternate: alt}
		loss := (p - boolFloat(event.Y)) * (p - boolFloat(event.Y))
		data.FullBrier += loss / learnedSteps
		if clock >= 256 {
			data.PostBrier += loss / 256
			if clock < 320 {
				data.EarlyBrier += loss / 64
			}
		}
		if v4Nominate(policy, event.X, pb, meanUncertainty, alt,
			learnedBudget-data.Nominated, learnedSteps-clock, rng) {
			tick.Selected = true
			data.Nominated++
			if clock >= 256 {
				data.NominatedPost++
				if alt == correctAlternative {
					data.CorrectPost++
				}
			}
			if !event.Missing {
				arrival := clock + delay
				if arrival < learnedSteps {
					due[arrival] = append(due[arrival], clock)
				} else {
					data.Pending++
				}
			}
		}
		for _, origin := range due[clock] {
			if origin > clock || tape[origin].Missing {
				return r, errors.New("future or missing v4 label")
			}
			data.Arrived++
			tick.Delivered = append(tick.Delivered, origin)
			observed := tape[origin]
			var err error
			if gated {
				err = current.observe(observed.X, observed.Y, base)
			} else {
				err = old.observe(observed.X, observed.Y, base)
			}
			if err != nil {
				return r, err
			}
			if err := selector.observe(observed.X, observed.Y, base); err != nil {
				return r, err
			}
		}
		data.Ticks = append(data.Ticks, tick)
	}
	if data.Nominated != learnedBudget || data.Arrived+data.Pending > learnedBudget {
		return r, errors.New("invalid v4 acquisition accounting")
	}
	r.Data.Data = data
	return r, nil
}

type v4Row struct {
	Kind, Split, Scenario string
	Fit, Stream           int
	Tape                  [learnedSteps]v3Tick
	Arms                  [8]v4Arm
}

func TestLearnedContrastV4(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_LEARNED_V4_OUT"), os.Getenv("EVENTFRAME_LEARNED_V4_SPLIT")
	if path == "" && split == "" {
		t.Skip("opt-in research study")
	}
	if path == "" || (split != "design" && split != "confirmation") {
		t.Fatal("output path and design|confirmation split required")
	}
	seed, fitOffset := int64(2026102401), 800
	if split == "confirmation" {
		seed, fitOffset = 2026102402, 900
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := make(map[string]string)
	for _, name := range []string{"docs/experiments/mmm-learned-contrast-v4-protocol.md",
		"internal/observationgate/learned_contrast_v4_test.go",
		"internal/observationgate/learned_contrast_v2_test.go",
		"internal/observationgate/learned_contrast_v3_test.go",
		"internal/observationlearners/experiment.go", "internal/observation/forecast.go"} {
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
		"budget": learnedBudget, "prior": v4Prior,
		"stateBytes": unsafe.Sizeof(v4State{}) + unsafe.Sizeof(learnedState{}),
		"hashes":     hashes}); err != nil {
		t.Fatal(err)
	}
	for scenario, name := range v3Cases {
		for fit := 0; fit < 16; fit++ {
			base, e := observationlearners.Base(observationlearners.Scenarios[0], scenario, fit+fitOffset)
			if e != nil {
				t.Fatal(e)
			}
			var probability [512]float64
			meanUncertainty := 0.
			for x := range probability {
				p, e := base.ForecastObserved(511, uint16(x))
				if e != nil || p <= 0 || p >= 1 {
					t.Fatalf("base forecast x=%d: %v", x, e)
				}
				probability[x] = p
				meanUncertainty += 4 * p * (1 - p) / 512
			}
			for stream := 0; stream < 16; stream++ {
				tape := v3Tape(seed, scenario, fit, stream)
				r := v4Row{Kind: "trial", Split: split, Scenario: name, Fit: fit, Stream: stream, Tape: tape}
				var oldTape [learnedSteps]learnedTapeTick
				for clock, tick := range tape {
					oldTape[clock] = learnedTapeTick{X: tick.X, Y: tick.Y, Missing: tick.Missing}
				}
				correct, delay := -1, 0
				switch scenario {
				case 1:
					correct = 3
				case 2:
					correct, delay = 3, 16
				case 3:
					correct = 1
				case 4:
					correct = 10
				}
				for policy, policyName := range learnedPolicies {
					selectorSeed := learnedSeed(seed, scenario, fit, stream, 10+policy)
					oldRNG := rand.New(rand.NewSource(selectorSeed))
					old, e := learnedRunArm(oldTape, &probability, meanUncertainty, policyName, oldRNG, correct, delay)
					if e != nil {
						t.Fatalf("old %s/%d/%d/%s: %v", name, fit, stream, policyName, e)
					}
					r.Arms[policy].Data = v3Measure(old, tape, scenario != 0 && scenario != 5)
					if policy == 2 {
						fallbackRNG := rand.New(rand.NewSource(selectorSeed))
						fallback, e := v3RunArm(oldTape, &probability, meanUncertainty,
							policyName, fallbackRNG, correct, delay)
						if e != nil {
							t.Fatal(e)
						}
						r.Arms[3].Data = v3Measure(fallback, tape, scenario != 0 && scenario != 5)
						matchedRNG := rand.New(rand.NewSource(selectorSeed))
						r.Arms[4], e = v4RunArm(oldTape, &probability, meanUncertainty,
							policyName, matchedRNG, correct, delay, false)
						if e != nil {
							t.Fatal(e)
						}
						r.Arms[4].Data = v3Measure(r.Arms[4].Data.Data, tape, scenario != 0 && scenario != 5)
					}
					gatedRNG := rand.New(rand.NewSource(selectorSeed))
					r.Arms[policy+5], e = v4RunArm(oldTape, &probability, meanUncertainty,
						policyName, gatedRNG, correct, delay, true)
					if e != nil {
						t.Fatal(e)
					}
					r.Arms[policy+5].Data = v3Measure(r.Arms[policy+5].Data.Data, tape, scenario != 0 && scenario != 5)
				}
				for _, pair := range [][2]int{{0, 5}, {1, 6}, {2, 4}, {2, 7}} {
					for clock := 0; clock < learnedSteps; clock++ {
						if r.Arms[pair[0]].Data.Data.Ticks[clock].Selected !=
							r.Arms[pair[1]].Data.Data.Ticks[clock].Selected {
							t.Fatalf("matched acquisition diverged %s/%d/%d arms=%v clock=%d",
								name, fit, stream, pair, clock)
						}
					}
				}
				if e := enc.Encode(r); e != nil {
					t.Fatal(e)
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

func TestLearnedContrastV4GateUnit(t *testing.T) {
	var base [512]float64
	for x := range base {
		base[x] = v3Truth(0, uint16(x), 0)
	}
	state := newV4State()
	old := newLearnedState()
	for x := uint16(0); x < 512; x += 17 {
		p, active, err := state.forecast(x, base[x])
		if err != nil || active || math.Abs(p-old.forecast(x, base[x])) > 1e-12 {
			t.Fatalf("inactive gate differs from old law at %d: %g %t %v", x, p, active, err)
		}
		if err := state.observe(x, x&(1<<2) != 0, &base); err != nil {
			t.Fatal(err)
		}
		if err := old.observe(x, x&(1<<2) != 0, &base); err != nil {
			t.Fatal(err)
		}
	}
	if state.bestKnown() != 3 {
		t.Fatal("lost known bit2 alternative")
	}
	null := newV4State()
	xRNG := rand.New(rand.NewSource(17))
	yRNG := rand.New(rand.NewSource(29))
	for i := 0; i < 32; i++ {
		if err := null.observe(uint16(xRNG.Intn(512)), yRNG.Intn(2) == 1, &base); err != nil {
			t.Fatal(err)
		}
	}
	if p, active, err := null.forecast(7, base[7]); err != nil || !active || p != .5 {
		t.Fatalf("null evidence did not activate gate: %g %t %v", p, active, err)
	}
}

func TestLearnedContrastV4MatchedSelection(t *testing.T) {
	base, err := observationlearners.Base(observationlearners.Scenarios[0], 0, 800)
	if err != nil {
		t.Fatal(err)
	}
	var probability [512]float64
	meanUncertainty := 0.
	for x := range probability {
		probability[x], err = base.ForecastObserved(511, uint16(x))
		if err != nil {
			t.Fatal(err)
		}
		meanUncertainty += 4 * probability[x] * (1 - probability[x]) / 512
	}
	seed := int64(2026102401)
	tape := v3Tape(seed, 0, 0, 0)
	var oldTape [learnedSteps]learnedTapeTick
	for clock, tick := range tape {
		oldTape[clock] = learnedTapeTick{X: tick.X, Y: tick.Y, Missing: tick.Missing}
	}
	selectorSeed := learnedSeed(seed, 0, 0, 0, 12)
	oldRNG := rand.New(rand.NewSource(selectorSeed))
	old, err := learnedRunArm(oldTape, &probability, meanUncertainty,
		"learned_disagreement", oldRNG, -1, 0)
	if err != nil {
		t.Fatal(err)
	}
	newRNG := rand.New(rand.NewSource(selectorSeed))
	current, err := v4RunArm(oldTape, &probability, meanUncertainty,
		"learned_disagreement", newRNG, -1, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	for clock := 0; clock < learnedSteps; clock++ {
		a, b := old.Ticks[clock], current.Data.Data.Ticks[clock]
		if a.Selected != b.Selected {
			t.Fatalf("first selection divergence clock=%d x=%d oldAlt=%d newAlt=%d oldSelected=%t newSelected=%t oldDelivered=%v newDelivered=%v",
				clock, tape[clock].X, a.Alternate, b.Alternate,
				a.Selected, b.Selected, a.Delivered, b.Delivered)
		}
	}
}
