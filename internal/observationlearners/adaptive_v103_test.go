package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"math/bits"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type adaptiveV103Step struct {
	X               uint16
	Q               float64
	Y               bool
	P               [6]float64
	Mask            [6]uint16
	Cost            [6]int
	Nodes, Branches int
}
type adaptiveV103Record struct {
	Phase, Case string
	Index       int
	Rules       [2]uint16
	Metrics     [6][2]stackV93Metric
	Realized    [6]float64
	Steps       []adaptiveV103Step
}

func adaptiveV103Random(m observationMixture, x uint16, rng *rand.Rand) (uint16, error) {
	mask := uint16(1)
	for bits.OnesCount16(mask) < 6 {
		p, err := m.Forecast(mask, x&mask)
		if err != nil {
			return 0, err
		}
		if p <= .1 || p >= .9 {
			break
		}
		var choices []uint16
		for s := 0; s < 3; s++ {
			for d := 0; d < 3; d++ {
				v := observation.View{Scope: s, Depth: d}.Mask()
				if v&^mask != 0 && bits.OnesCount16(mask|v) <= 6 {
					choices = append(choices, v)
				}
			}
		}
		if len(choices) == 0 {
			break
		}
		mask |= choices[rng.Intn(len(choices))]
	}
	return mask, nil
}

func adaptiveV103Run(phase, scenario, index int, base int64) (adaptiveV103Record, error) {
	r := adaptiveV103Record{Phase: []string{"design", "confirmation"}[phase], Case: replicationV102Cases[scenario], Index: index}
	seed := base + int64(phase*1000000+scenario*10000+index*10)
	rules, xs, ys, views := rand.New(rand.NewSource(seed)), rand.New(rand.NewSource(seed+1)), rand.New(rand.NewSource(seed+2)), rand.New(rand.NewSource(seed+3))
	arities := [12]int{1, 2, 3, 4, 4, 3, 3, 0, 0, 4, 3, 4}
	for k := range r.Rules {
		a := arities[scenario]
		if scenario >= 10 && k == 1 {
			a = 7 - a
		}
		if a > 0 {
			pool := stackV93Masks(a, phase)
			r.Rules[k] = pool[rules.Intn(len(pool))]
		}
	}
	truth := func(x uint16, step int) float64 {
		name, rule := r.Case, r.Rules[0]
		if scenario >= 10 {
			name = "majority3"
			if scenario == 11 {
				name = "parity4"
			}
			if step >= 128 {
				rule = r.Rules[1]
				if scenario == 10 {
					name = "parity4"
				} else {
					name = "majority3"
				}
			}
		}
		return stackV93Truth(x, rule, name)
	}
	history := make([]observation.Sample, 0, 272)
	for step := -16; step < 0; step++ {
		x := stackV93Input(uint16(xs.Intn(512)), r.Case)
		history = append(history, observation.Sample{Bits: x, Outcome: ys.Float64() < truth(x, step)})
	}
	var states [5]*routedObservationState
	for j := range states {
		states[j] = newRoutedObservationState()
	}
	var weights [512]float64
	for j := range weights {
		weights[j] = 1
	}
	var publication *observationExperts
	for step := 0; step < 256; step++ {
		if step%32 == 0 {
			window := history
			if len(window) > 64 {
				window = window[len(window)-64:]
			}
			short := window
			if len(short) > 32 {
				short = short[len(short)-32:]
			}
			var models [4]*ConditionalForest
			var err error
			models[0], err = NewSubsetConditional(window, weights)
			if err != nil {
				return r, err
			}
			models[1], err = NewBooleanConditional(window)
			if err != nil {
				return r, err
			}
			models[2], err = NewSubsetConditional(short, weights)
			if err != nil {
				return r, err
			}
			models[3], err = NewBooleanConditional(short)
			if err != nil {
				return r, err
			}
			publication, err = newObservationExperts(models)
			if err != nil {
				return r, err
			}
		}
		x := stackV93Input(uint16(xs.Intn(512)), r.Case)
		entry := adaptiveV103Step{X: x}
		generic, err := RunConditionalObserver(&publication.models[0], &jointReader{x: x, epoch: 1}, 1)
		if err != nil {
			return r, err
		}
		entry.P[0], entry.Mask[0], entry.Cost[0] = generic.Probability, generic.Trace[len(generic.Trace)-1].Observed, generic.Cost
		for j, state := range states {
			arm := j + 1
			if arm == 2 {
				got, err := state.predict(uint64(step), publication, &jointReader{x: x, epoch: 1}, 1)
				if err != nil {
					return r, err
				}
				entry.P[arm], entry.Mask[arm], entry.Cost[arm] = got.Probability, got.Trace[len(got.Trace)-1].Observed, got.Cost
			} else if arm == 3 {
				got, err := state.predictLookahead(uint64(step), publication, &jointReader{x: x, epoch: 1}, 1)
				if err != nil {
					return r, err
				}
				entry.P[arm], entry.Mask[arm], entry.Cost[arm] = got.Probability, got.Trace[len(got.Trace)-1].Observed, got.Cost
				entry.Nodes, entry.Branches = got.Nodes, got.Branches
			} else {
				m, err := state.preview(uint64(step), publication)
				if err != nil {
					return r, err
				}
				mask := uint16(63)
				if arm == 1 {
					mask = entry.Mask[0]
				}
				if arm == 4 {
					mask, err = adaptiveV103Random(m, x, views)
					if err != nil {
						return r, err
					}
				}
				f, err := state.finish(uint64(step), m, mask, x&mask)
				if err != nil {
					return r, err
				}
				entry.P[arm], entry.Mask[arm], entry.Cost[arm] = f.P, mask, bits.OnesCount16(mask)
			}
		}
		// Truth and outcome are evaluated only after all acquired forecasts exist.
		entry.Q = truth(x, step)
		entry.Y = ys.Float64() < entry.Q
		y := 0.
		if entry.Y {
			y = 1
		}
		for arm, p := range entry.P {
			if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 || entry.Cost[arm] < 1 || entry.Cost[arm] > 6 || entry.Cost[arm] != bits.OnesCount16(entry.Mask[arm]) {
				return r, fmt.Errorf("invalid scored bundle")
			}
			brier := (p-entry.Q)*(p-entry.Q) + entry.Q*(1-entry.Q)
			acc := 1 - entry.Q
			if p >= .5 {
				acc = entry.Q
			}
			r.Metrics[arm][0].Brier += brier / 256
			r.Metrics[arm][0].Accuracy += acc / 256
			if step >= 128 {
				r.Metrics[arm][1].Brier += brier / 128
				r.Metrics[arm][1].Accuracy += acc / 128
			}
			r.Realized[arm] += (p - y) * (p - y) / 256
		}
		r.Steps = append(r.Steps, entry)
		for _, state := range states {
			if err := state.observe(uint64(step), entry.Y); err != nil {
				return r, err
			}
		}
		history = append(history, observation.Sample{Bits: x, Outcome: entry.Y})
	}
	return r, nil
}

func TestAdaptiveV103Compatibility(t *testing.T) {
	for _, scenario := range []int{0, 3, 10, 11} {
		old, err := replicationV102Run(0, scenario, 0)
		if err != nil {
			t.Fatal(err)
		}
		got, err := adaptiveV103Run(0, scenario, 0, 2082110200)
		if err != nil {
			t.Fatal(err)
		}
		if got.Rules != old.Rules || got.Metrics[5] != old.Metrics[11][1] {
			t.Fatal("fixed-mask v102 mismatch", scenario, got.Metrics[5], old.Metrics[11][1])
		}
	}
}

func TestAdaptiveV103Seeds(t *testing.T) {
	used := map[int64]bool{}
	for phase := int64(0); phase < 2; phase++ {
		for scenario := int64(0); scenario < 12; scenario++ {
			for index := int64(0); index < 32; index++ {
				for role := int64(0); role < 4; role++ {
					seed := (2090110300 + phase*1000000 + scenario*10000 + index*10 + role) % 2147483647
					if used[seed] {
						t.Fatal("within-run seed collision")
					}
					used[seed] = true
				}
			}
		}
	}
	for _, base := range []int64{2026119000, 2030119100, 2034119200, 2038119300, 2042119400, 2046119500, 2050119600, 2054119700, 2062119900, 2066110000, 2078110100, 2082110200} {
		// Conservatively include 128 indices and four roles for all earlier runs.
		for phase := int64(0); phase < 2; phase++ {
			for scenario := int64(0); scenario < 12; scenario++ {
				for index := int64(0); index < 128; index++ {
					for role := int64(0); role < 4; role++ {
						if used[(base+phase*1000000+scenario*10000+index*10+role)%2147483647] {
							t.Fatal("archived learner seed collision", base)
						}
					}
				}
			}
		}
	}
	for _, base := range []int64{3070119900, 3090110000, 3110110100, 3130110200} {
		for mode := int64(0); mode < 2; mode++ {
			for family := int64(0); family < 4096; family++ {
				for test := int64(0); test < 64; test++ {
					if used[(base+mode*10000000+family*1000+test)%2147483647] {
						t.Fatal("archived null seed collision", base)
					}
				}
			}
		}
	}
}

type adaptiveV103Artifact struct {
	Version, Go, OS, Arch string
	SeedBase              int64
	PerCase               int
	Hashes                map[string]string
	Records               []adaptiveV103Record
}

func TestAdaptiveV103(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ADAPTIVE_V103")
	if path == "" {
		t.Skip("explicit artifact required")
	}
	if os.Getenv("EVENTFRAME_ADAPTIVE_V103_REPLAY") == "1" {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var a adaptiveV103Artifact
		if err := json.Unmarshal(raw, &a); err != nil {
			t.Fatal(err)
		}
		for _, r := range a.Records {
			phase := 0
			if r.Phase == "confirmation" {
				phase = 1
			}
			scenario := -1
			for i, name := range replicationV102Cases {
				if name == r.Case {
					scenario = i
				}
			}
			if scenario < 0 {
				t.Fatal("unknown case")
			}
			got, err := adaptiveV103Run(phase, scenario, r.Index, a.SeedBase)
			if err != nil || !reflect.DeepEqual(got, r) {
				t.Fatal("replay", r.Phase, r.Case, r.Index, err)
			}
		}
		return
	}
	a := adaptiveV103Artifact{Version: "v103", Go: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, SeedBase: 2090110300, PerCase: 32, Hashes: map[string]string{}}
	paths := []string{"go.mod", "go.sum", "docs/experiments/mmm-adaptive-v103-protocol.md", "research/adaptive-v103-summary.mjs", "internal/observation/controller.go", "internal/observationlearners/adaptive_v103_test.go", "internal/observationlearners/replication_v102_test.go", "internal/observationlearners/stacking_v93_test.go", "internal/observationlearners/subset.go", "internal/observationlearners/conditional.go", "internal/observationlearners/boolean_specialist.go", "internal/observationlearners/brier_bank.go", "internal/observationlearners/forecast_falsification.go", "internal/observationlearners/comparative_falsification.go", "internal/observationlearners/evidence_routing.go", "internal/observationlearners/joint_observation.go", "internal/observationlearners/joint_observer.go", "internal/observationlearners/joint_observation_test.go", "internal/observationlearners/joint_lookahead.go", "internal/observationlearners/joint_lookahead_test.go", "internal/observationlearners/conditional_observer.go"}
	for _, name := range paths {
		raw, err := os.ReadFile(filepath.Join("../..", name))
		if err != nil {
			t.Fatal(err)
		}
		a.Hashes[name] = fmt.Sprintf("%x", sha256.Sum256(raw))
	}
	for phase := 0; phase < 2; phase++ {
		for scenario := range replicationV102Cases {
			for index := 0; index < 32; index++ {
				r, err := adaptiveV103Run(phase, scenario, index, a.SeedBase)
				if err != nil {
					t.Fatal(err)
				}
				a.Records = append(a.Records, r)
			}
		}
	}
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(raw); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
