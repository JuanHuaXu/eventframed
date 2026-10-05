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

	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

const learnedSteps = 512
const learnedBudget = 128
const learnedTrials = 256

var learnedCases = [...]string{"stable", "bit2_shift", "bit2_delayed", "bit0_shift", "interaction_shift", "null"}
var learnedPolicies = [...]string{"random", "uncertainty", "learned_disagreement"}
var learnedPrior = [...]float64{.90, .01, .01, .01, .01, .01, .01, .01, .01, .01, .01}

type learnedTapeTick struct {
	X       uint16
	Y       bool
	Missing bool
}

type learnedEvidence struct {
	X uint16
	Y bool
}

type learnedState struct {
	History [32]learnedEvidence
	N, Next int
	Weight  [11]float64
}

func newLearnedState() learnedState { return learnedState{Weight: learnedPrior} }

func learnedRule(rule int, x uint16, base float64) float64 {
	if rule == 0 {
		return base
	}
	bit := false
	if rule <= 9 {
		bit = x&(1<<(rule-1)) != 0
	} else {
		bit = ((x&1 != 0) != (x&2 != 0)) != (x&4 != 0)
	}
	if bit {
		return .95
	}
	return .05
}

func (s *learnedState) forecast(x uint16, base float64) float64 {
	p := 0.
	for h, weight := range s.Weight {
		p += weight * learnedRule(h, x, base)
	}
	return p
}

func (s *learnedState) bestAlternative() int {
	best := 1
	for h := 2; h < len(s.Weight); h++ {
		if s.Weight[h] > s.Weight[best] {
			best = h
		}
	}
	return best
}

// Evidence enters this working likelihood only after its request has arrived.
func (s *learnedState) observe(x uint16, y bool, base *[512]float64) error {
	s.History[s.Next] = learnedEvidence{X: x, Y: y}
	s.Next = (s.Next + 1) % len(s.History)
	if s.N < len(s.History) {
		s.N++
	}
	var logw [11]float64
	maximum := math.Inf(-1)
	for h, prior := range learnedPrior {
		logw[h] = math.Log(prior)
		for i := 0; i < s.N; i++ {
			v := s.History[i]
			p := learnedRule(h, v.X, base[v.X])
			if v.Y {
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
		return errors.New("invalid learned hypothesis normalization")
	}
	for h := range s.Weight {
		s.Weight[h] /= total
	}
	return nil
}

func learnedSeed(base int64, scenario, fit, stream, role int) int64 {
	return base*100000000 + int64(scenario)*1000000 + int64(fit)*10000 + int64(stream)*100 + int64(role)
}

func learnedTape(seed int64, scenario, fit, stream int) [learnedSteps]learnedTapeTick {
	rng := rand.New(rand.NewSource(learnedSeed(seed, scenario, fit, stream, 0)))
	missing := rand.New(rand.NewSource(learnedSeed(seed, scenario, fit, stream, 1)))
	var tape [learnedSteps]learnedTapeTick
	for clock := range tape {
		x := uint16(rng.Intn(512))
		p := learnedRule(0, x, .05)
		if scenario != 5 && clock >= 256 {
			switch scenario {
			case 1, 2:
				p = learnedRule(3, x, .05)
			case 3:
				p = learnedRule(1, x, .05)
			case 4:
				p = learnedRule(10, x, .05)
			}
		}
		if scenario == 0 || clock < 256 {
			bit := ((x&(1<<6) != 0) != (x&(1<<7) != 0)) != (x&(1<<8) != 0)
			if bit {
				p = .95
			} else {
				p = .05
			}
		}
		if scenario == 5 {
			p = .5
		}
		missRate := .20
		if scenario == 2 {
			missRate = .25
		}
		tape[clock] = learnedTapeTick{X: x, Y: rng.Float64() < p, Missing: missing.Float64() < missRate}
	}
	return tape
}

type learnedArmTick struct {
	P         float64
	Selected  bool
	Alternate int
	Delivered []int `json:",omitempty"`
}

type learnedArm struct {
	Policy                           string
	Ticks                            []learnedArmTick
	FullBrier, PostBrier, EarlyBrier float64
	Nominated, Arrived, Pending      int
	CorrectPost, NominatedPost       int
}

type learnedRow struct {
	Kind, Split, Scenario string
	Fit, Stream           int
	Tape                  [learnedSteps]learnedTapeTick
	Arms                  [3]learnedArm
}

func learnedNominate(policy string, x uint16, base, meanUncertainty float64,
	state *learnedState, remaining, clocks int, rng *rand.Rand) bool {
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
		alt := state.bestAlternative()
		if (learnedRule(alt, x, base) >= .5) == (base >= .5) {
			return false
		}
		rate *= 2
	}
	return rng.Float64() < math.Min(1, rate)
}

func learnedRunArm(tape [learnedSteps]learnedTapeTick, base *[512]float64,
	meanUncertainty float64, policy string, rng *rand.Rand, correctAlternative int,
	delay int) (learnedArm, error) {
	r := learnedArm{Policy: policy}
	state := newLearnedState()
	var due [learnedSteps][]int
	for clock, event := range tape {
		pb := base[event.X]
		p := state.forecast(event.X, pb)
		if p <= 0 || p >= 1 || math.IsNaN(p) {
			return r, errors.New("invalid working forecast")
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
		if learnedNominate(policy, event.X, pb, meanUncertainty, &state,
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
				return r, errors.New("invalid due label")
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
		return r, errors.New("invalid learned acquisition accounting")
	}
	return r, nil
}

func boolFloat(v bool) float64 {
	if v {
		return 1
	}
	return 0
}

func TestLearnedContrastV2(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_LEARNED_CONTRAST_OUT"), os.Getenv("EVENTFRAME_LEARNED_CONTRAST_SPLIT")
	if path == "" && split == "" {
		t.Skip("opt-in research study")
	}
	if path == "" || (split != "design" && split != "confirmation") {
		t.Fatal("output path and design|confirmation split required")
	}
	seed := int64(2026102201)
	fitOffset := 400
	if split == "confirmation" {
		seed, fitOffset = 2026102202, 500
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := make(map[string]string)
	for _, name := range []string{"docs/experiments/mmm-learned-contrast-v2-protocol.md",
		"internal/observationgate/learned_contrast_v2_test.go", "internal/observationlearners/experiment.go",
		"internal/observation/forecast.go"} {
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
	if err := enc.Encode(map[string]any{"kind": "manifest", "split": split, "seedBase": seed,
		"fitOffset": fitOffset, "trialsPerCase": learnedTrials, "budget": learnedBudget,
		"hypothesisPrior": learnedPrior, "hashes": hashes}); err != nil {
		t.Fatal(err)
	}
	for scenario, name := range learnedCases {
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
				tape := learnedTape(seed, scenario, fit, stream)
				r := learnedRow{Kind: "trial", Split: split, Scenario: name, Fit: fit, Stream: stream, Tape: tape}
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
				for arm, policy := range learnedPolicies {
					rng := rand.New(rand.NewSource(learnedSeed(seed, scenario, fit, stream, 10+arm)))
					r.Arms[arm], e = learnedRunArm(tape, &probability, meanUncertainty, policy, rng, correct, delay)
					if e != nil {
						t.Fatalf("%s fit=%d stream=%d policy=%s: %v", name, fit, stream, policy, e)
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

func TestLearnedContrastV2ArrivalBoundary(t *testing.T) {
	seed := int64(2026102201)
	tape := learnedTape(seed, 2, 0, 0)
	base, err := observationlearners.Base(observationlearners.Scenarios[0], 2, 400)
	if err != nil {
		t.Fatal(err)
	}
	var probabilities [512]float64
	meanUncertainty := 0.
	for x := range probabilities {
		probabilities[x], err = base.ForecastObserved(511, uint16(x))
		if err != nil {
			t.Fatal(err)
		}
		meanUncertainty += 4 * probabilities[x] * (1 - probabilities[x]) / 512
	}
	var arms [3]learnedArm
	for arm, policy := range learnedPolicies {
		rng := rand.New(rand.NewSource(learnedSeed(seed, 2, 0, 0, 10+arm)))
		arms[arm], err = learnedRunArm(tape, &probabilities, meanUncertainty, policy, rng, 3, 16)
		if err != nil {
			t.Fatal(err)
		}
	}
	for clock := 0; clock <= 16; clock++ {
		for arm := 1; arm < 3; arm++ {
			if arms[arm].Ticks[clock].P != arms[0].Ticks[clock].P {
				t.Fatalf("unarrived outcome changed clock=%d arm=%d", clock, arm)
			}
		}
	}
}

func TestLearnedContrastV2KnownAlternative(t *testing.T) {
	var base [512]float64
	for x := range base {
		parity := ((x&(1<<6) != 0) != (x&(1<<7) != 0)) != (x&(1<<8) != 0)
		if parity {
			base[x] = .95
		} else {
			base[x] = .05
		}
	}
	s := newLearnedState()
	for x := uint16(0); x < 512; x += 17 {
		if err := s.observe(x, x&(1<<2) != 0, &base); err != nil {
			t.Fatal(err)
		}
	}
	if s.bestAlternative() != 3 || s.Weight[3] <= s.Weight[0] {
		t.Fatalf("known bit2 alternative not learned: baseline=%g bit2=%g best=%d",
			s.Weight[0], s.Weight[3], s.bestAlternative())
	}
	for x := uint16(0); x < 512; x += 53 {
		p := s.forecast(x, base[x])
		if (p >= .5) != (x&(1<<2) != 0) {
			t.Fatalf("learned forecast disagrees with known bit2 rule at %d: %g", x, p)
		}
	}
}
