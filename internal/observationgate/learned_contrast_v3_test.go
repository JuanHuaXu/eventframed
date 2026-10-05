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

var v3Cases = [...]string{"stable", "bit2_shift", "bit2_delayed", "bit0_shift", "interaction_shift", "null", "majority_ood"}
var v3Prior = [...]float64{.85, .01, .01, .01, .01, .01, .01, .01, .01, .01, .01, .05}

type v3Tick struct {
	X       uint16
	Y       bool
	Missing bool
	PTrue   float64
}

func v3Truth(scenario int, x uint16, clock int) float64 {
	if scenario == 5 {
		return .5
	}
	bit := ((x&(1<<6) != 0) != (x&(1<<7) != 0)) != (x&(1<<8) != 0)
	if clock >= 256 {
		switch scenario {
		case 1, 2:
			bit = x&(1<<2) != 0
		case 3:
			bit = x&1 != 0
		case 4:
			bit = ((x&1 != 0) != (x&2 != 0)) != (x&4 != 0)
		case 6:
			count := 0
			for j := 0; j < 3; j++ {
				if x&(1<<j) != 0 {
					count++
				}
			}
			bit = count >= 2
		}
	}
	if bit {
		return .95
	}
	return .05
}

func v3Tape(seed int64, scenario, fit, stream int) [learnedSteps]v3Tick {
	rng := rand.New(rand.NewSource(learnedSeed(seed, scenario, fit, stream, 0)))
	missing := rand.New(rand.NewSource(learnedSeed(seed, scenario, fit, stream, 1)))
	var tape [learnedSteps]v3Tick
	for clock := range tape {
		x := uint16(rng.Intn(512))
		p := v3Truth(scenario, x, clock)
		missRate := .20
		if scenario == 2 {
			missRate = .25
		}
		tape[clock] = v3Tick{X: x, Y: rng.Float64() < p, Missing: missing.Float64() < missRate, PTrue: p}
	}
	return tape
}

type v3State struct {
	History [32]learnedEvidence
	N, Next int
	Weight  [12]float64
}

func newV3State() v3State { return v3State{Weight: v3Prior} }

func v3Rule(h int, x uint16, base float64) float64 {
	if h == 11 {
		return .5
	}
	return learnedRule(h, x, base)
}

func (s *v3State) forecast(x uint16, base float64) float64 {
	p := 0.
	for h, w := range s.Weight {
		p += w * v3Rule(h, x, base)
	}
	return p
}

func (s *v3State) bestKnown() int {
	best := 1
	for h := 2; h <= 10; h++ {
		if s.Weight[h] > s.Weight[best] {
			best = h
		}
	}
	return best
}

func (s *v3State) observe(x uint16, y bool, base *[512]float64) error {
	s.History[s.Next] = learnedEvidence{X: x, Y: y}
	s.Next = (s.Next + 1) % len(s.History)
	if s.N < len(s.History) {
		s.N++
	}
	var logw [12]float64
	maximum := math.Inf(-1)
	for h, prior := range v3Prior {
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
		return errors.New("invalid v3 hypothesis normalization")
	}
	for h := range s.Weight {
		s.Weight[h] /= total
	}
	return nil
}

func v3Nominate(policy string, x uint16, base, meanUncertainty float64,
	state *v3State, remaining, clocks int, rng *rand.Rand) bool {
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
		if state.Weight[11] <= .5 {
			alt := state.bestKnown()
			if (v3Rule(alt, x, base) >= .5) == (base >= .5) {
				return false
			}
			rate *= 2
		}
	}
	return rng.Float64() < math.Min(1, rate)
}

func v3RunArm(tape [learnedSteps]learnedTapeTick, base *[512]float64,
	meanUncertainty float64, policy string, rng *rand.Rand, correctAlternative int,
	delay int) (learnedArm, error) {
	r := learnedArm{Policy: policy}
	state := newV3State()
	var due [learnedSteps][]int
	for clock, event := range tape {
		pb := base[event.X]
		p := state.forecast(event.X, pb)
		if p <= 0 || p >= 1 || math.IsNaN(p) {
			return r, errors.New("invalid v3 forecast")
		}
		alt := state.bestKnown()
		tick := learnedArmTick{P: p, Alternate: alt}
		loss := (p - boolFloat(event.Y)) * (p - boolFloat(event.Y))
		r.FullBrier += loss / learnedSteps
		if clock >= 256 {
			r.PostBrier += loss / 256
			if clock < 320 {
				r.EarlyBrier += loss / 64
			}
		}
		if v3Nominate(policy, event.X, pb, meanUncertainty, &state,
			learnedBudget-r.Nominated, learnedSteps-clock, rng) {
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
				return r, errors.New("invalid v3 due label")
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
		return r, errors.New("invalid v3 acquisition accounting")
	}
	return r, nil
}

type v3Arm struct {
	Data           learnedArm
	PostExpected   float64
	RecoveryDelay  int
	MissedRecovery bool
}

// Expected loss is evaluator-only and is never passed into the selector.
func v3Measure(data learnedArm, tape [learnedSteps]v3Tick, changed bool) v3Arm {
	r := v3Arm{Data: data, RecoveryDelay: -1}
	var expected [learnedSteps]float64
	for clock, tick := range data.Ticks {
		pTrue := tape[clock].PTrue
		expected[clock] = pTrue*(1-pTrue) + (tick.P-pTrue)*(tick.P-pTrue)
		if clock >= 256 {
			r.PostExpected += expected[clock] / 256
		}
	}
	if !changed {
		return r
	}
	consecutive := 0
	for clock := 287; clock < learnedSteps; clock++ {
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
			r.RecoveryDelay = clock - 256
			return r
		}
	}
	r.RecoveryDelay, r.MissedRecovery = 256, true
	return r
}

type v3Row struct {
	Kind, Split, Scenario string
	Fit, Stream           int
	Tape                  [learnedSteps]v3Tick
	Arms                  [6]v3Arm
}

func TestLearnedContrastV3(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_LEARNED_V3_OUT"), os.Getenv("EVENTFRAME_LEARNED_V3_SPLIT")
	if path == "" && split == "" {
		t.Skip("opt-in research study")
	}
	if path == "" || (split != "design" && split != "confirmation") {
		t.Fatal("output path and design|confirmation split required")
	}
	seed, fitOffset := int64(2026102301), 600
	if split == "confirmation" {
		seed, fitOffset = 2026102302, 700
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := make(map[string]string)
	for _, name := range []string{"docs/experiments/mmm-learned-contrast-v3-protocol.md",
		"internal/observationgate/learned_contrast_v3_test.go",
		"internal/observationgate/learned_contrast_v2_test.go",
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
		"budget": learnedBudget, "prior": v3Prior, "stateBytes": unsafe.Sizeof(v3State{}),
		"hashes": hashes}); err != nil {
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
				r := v3Row{Kind: "trial", Split: split, Scenario: name, Fit: fit, Stream: stream, Tape: tape}
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
				for policy, name := range learnedPolicies {
					selectorSeed := learnedSeed(seed, scenario, fit, stream, 10+policy)
					rngOld := rand.New(rand.NewSource(selectorSeed))
					old, e := learnedRunArm(oldTape, &probability, meanUncertainty, name, rngOld, correct, delay)
					if e != nil {
						t.Fatalf("old %s/%d/%d/%s: %v", v3Cases[scenario], fit, stream, name, e)
					}
					r.Arms[policy] = v3Measure(old, tape, scenario != 0 && scenario != 5)
					rngNew := rand.New(rand.NewSource(selectorSeed))
					current, e := v3RunArm(oldTape, &probability, meanUncertainty, name, rngNew, correct, delay)
					if e != nil {
						t.Fatalf("new %s/%d/%d/%s: %v", v3Cases[scenario], fit, stream, name, e)
					}
					r.Arms[policy+3] = v3Measure(current, tape, scenario != 0 && scenario != 5)
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

func TestLearnedContrastV3CausalBoundary(t *testing.T) {
	seed := int64(2026102301)
	base, err := observationlearners.Base(observationlearners.Scenarios[0], 2, 600)
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
	original := v3Tape(seed, 2, 0, 0)
	var tape [learnedSteps]learnedTapeTick
	for clock, tick := range original {
		tape[clock] = learnedTapeTick{X: tick.X, Y: tick.Y, Missing: tick.Missing}
	}
	var arms [3]learnedArm
	for policy, name := range learnedPolicies {
		rng := rand.New(rand.NewSource(learnedSeed(seed, 2, 0, 0, 10+policy)))
		arms[policy], err = v3RunArm(tape, &probability, meanUncertainty, name, rng, 3, 16)
		if err != nil {
			t.Fatal(err)
		}
	}
	for clock := 0; clock <= 16; clock++ {
		for policy := 1; policy < 3; policy++ {
			if arms[policy].Ticks[clock].P != arms[0].Ticks[clock].P {
				t.Fatalf("unarrived outcome changed clock=%d policy=%d", clock, policy)
			}
		}
	}
	state := newV3State()
	for i := 0; i < 512; i++ {
		x := uint16(i)
		if err := state.observe(x, i%2 == 0, &probability); err != nil {
			t.Fatal(err)
		}
		if state.Weight[11] < 0 || state.Weight[11] > 1 {
			t.Fatal("invalid null posterior")
		}
	}
}

func TestLearnedContrastV3NullAndKnownRules(t *testing.T) {
	var base [512]float64
	for x := range base {
		base[x] = v3Truth(0, uint16(x), 0)
	}
	xRNG := rand.New(rand.NewSource(17))
	yRNG := rand.New(rand.NewSource(29))
	null := newV3State()
	for i := 0; i < 32; i++ {
		if err := null.observe(uint16(xRNG.Intn(512)), yRNG.Intn(2) == 1, &base); err != nil {
			t.Fatal(err)
		}
	}
	if null.Weight[11] <= .5 {
		t.Fatalf("null branch did not dominate uninformative evidence: %g", null.Weight[11])
	}
	known := newV3State()
	for x := uint16(0); x < 512; x += 17 {
		if err := known.observe(x, x&(1<<2) != 0, &base); err != nil {
			t.Fatal(err)
		}
	}
	if known.bestKnown() != 3 || known.Weight[11] >= .5 {
		t.Fatalf("known bit2 rule was lost: bit2=%g null=%g best=%d",
			known.Weight[3], known.Weight[11], known.bestKnown())
	}
}

func TestLearnedContrastV3RecoveryClock(t *testing.T) {
	var tape [learnedSteps]v3Tick
	arm := learnedArm{Ticks: make([]learnedArmTick, learnedSteps)}
	for clock := range tape {
		x := uint16(clock % 512)
		p := v3Truth(1, x, clock)
		tape[clock] = v3Tick{X: x, PTrue: p}
		arm.Ticks[clock] = learnedArmTick{P: p}
	}
	perfect := v3Measure(arm, tape, true)
	if perfect.MissedRecovery || perfect.RecoveryDelay != 46 {
		t.Fatalf("wrong earliest recovery: %+v", perfect)
	}
	for clock := range arm.Ticks {
		arm.Ticks[clock].P = .5
	}
	unrecovered := v3Measure(arm, tape, true)
	if !unrecovered.MissedRecovery || unrecovered.RecoveryDelay != 256 {
		t.Fatalf("wrong unrecovered cap: %+v", unrecovered)
	}
}
