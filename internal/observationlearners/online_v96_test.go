package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

var onlineV96Cases = []string{"parity1", "parity2", "parity3", "parity4", "complement4", "majority3", "mux3", "constant", "null", "dependent4", "majority_to_parity", "parity_to_majority"}

type onlineV96Record struct {
	Phase, Case string
	Index       int
	Rules       [2]uint16
	// Arm: generic, parity, skeptical BMA, online rho0, online rho.001, interval.
	// View: full, mask63. Segment: all256, late128. Values are means.
	Metrics      [6][2][2]stackV93Metric
	Realized     [6][2]float64
	MaxViolation [2][2]float64
	Tape         string
}

func onlineV96Run(phase, scenario, index int) (onlineV96Record, error) {
	r := onlineV96Record{Phase: []string{"design", "confirmation"}[phase], Case: onlineV96Cases[scenario], Index: index}
	seed := int64(2050119600 + phase*1000000 + scenario*10000 + index*10)
	rules, xs, ys := rand.New(rand.NewSource(seed)), rand.New(rand.NewSource(seed+1)), rand.New(rand.NewSource(seed+2))
	arities := [12]int{1, 2, 3, 4, 4, 3, 3, 0, 0, 4, 3, 4}
	for k := range r.Rules {
		arity := arities[scenario]
		if scenario >= 10 && k == 1 {
			arity = 7 - arity
		}
		if arity > 0 {
			pool := stackV93Masks(arity, phase)
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
	var states [2][2]*brierAggregator
	for view := range states {
		for arm, rho := range []float64{0, .001} {
			states[view][arm], _ = newBrierAggregator(rho)
		}
	}
	var intervals [2]*intervalBrier
	for view := range intervals {
		intervals[view], _ = newIntervalBrier(256)
	}
	var models [3]*ConditionalForest
	var weights [512]float64
	for k := range weights {
		weights[k] = 1
	}
	tape := sha256.New()
	for step := 0; step < 256; step++ {
		if step%32 == 0 {
			start := len(history) - 64
			if start < 0 {
				start = 0
			}
			window := history[start:]
			raw, _ := json.Marshal(window)
			tape.Write(raw)
			var e error
			models[0], e = NewSubsetConditional(window, weights)
			if e != nil {
				return r, e
			}
			models[1], e = NewBooleanConditional(window)
			if e != nil {
				return r, e
			}
			models[2], e = NewPriorConditional(window, .1)
			if e != nil {
				return r, e
			}
		}
		x := stackV93Input(uint16(xs.Intn(512)), r.Case)
		var forecasts [2][2]brierForecast
		var predictions [6][2]float64
		for view, mask := range []uint16{511, 63} {
			for arm, m := range models {
				p, e := m.Forecast(mask, x&mask)
				if e != nil {
					return r, e
				}
				predictions[arm][view] = p
			}
			for arm, s := range states[view] {
				f, e := s.predict(uint64(step), [2]float64{predictions[0][view], predictions[1][view]})
				if e != nil {
					return r, e
				}
				forecasts[view][arm] = f
				predictions[3+arm][view] = f.P
			}
		}
		for view, state := range intervals {
			p, e := state.predict(uint64(step), [2]float64{predictions[0][view], predictions[1][view]})
			if e != nil {
				return r, e
			}
			predictions[5][view] = p
		}
		// Outcome sampling and oracle scoring occur only after every forecast exists.
		q := truth(x, step)
		y := ys.Float64() < q
		target := 0.
		if y {
			target = 1
		}
		raw, _ := json.Marshal(struct {
			Step int
			X    uint16
			P    [6][2]float64
			Y    bool
		}{step, x, predictions, y})
		tape.Write(raw)
		for arm := range predictions {
			for view, p := range predictions[arm] {
				if math.IsNaN(p) || p < 0 || p > 1 {
					return r, fmt.Errorf("invalid forecast")
				}
				expected := (p-q)*(p-q) + q*(1-q)
				acc := 1 - q
				if p >= .5 {
					acc = q
				}
				r.Metrics[arm][view][0].Brier += expected / 256
				r.Metrics[arm][view][0].Accuracy += acc / 256
				if step >= 128 {
					r.Metrics[arm][view][1].Brier += expected / 128
					r.Metrics[arm][view][1].Accuracy += acc / 128
				}
				r.Realized[arm][view] += (p - target) * (p - target)
			}
		}
		for view := range states {
			for arm, s := range states[view] {
				rho := []float64{0, .001}[arm]
				defect := r.Realized[3+arm][view] - r.Realized[0][view] - brierGenericBound(uint64(step+1), rho)
				r.MaxViolation[arm][view] = math.Max(r.MaxViolation[arm][view], defect)
				if e := s.observe(forecasts[view][arm].Sequence, y); e != nil {
					return r, e
				}
			}
		}
		for _, state := range intervals {
			if e := state.observe(uint64(step), y); e != nil {
				return r, e
			}
		}
		history = append(history, observation.Sample{Bits: x, Outcome: y})
	}
	r.Tape = fmt.Sprintf("%x", tape.Sum(nil))
	return r, nil
}

func TestOnlineV96Smoke(t *testing.T) {
	for _, scenario := range []int{3, 10, 11} {
		a, e := onlineV96Run(0, scenario, 0)
		if e != nil {
			t.Fatal(e)
		}
		b, e := onlineV96Run(0, scenario, 0)
		if e != nil || a != b {
			t.Fatal("replay", e)
		}
		for _, row := range a.MaxViolation {
			for _, d := range row {
				if d > 1e-8 {
					t.Fatal("bound", d)
				}
			}
		}
	}
}

func TestOnlineV96(t *testing.T) {
	out, replay := os.Getenv("EVENTFRAME_ONLINE_V96_OUT"), os.Getenv("EVENTFRAME_ONLINE_V96_REPLAY")
	if out == "" && replay == "" {
		t.Skip("explicit artifact required")
	}
	if out != "" && replay != "" {
		t.Fatal("choose one")
	}
	type artifact struct {
		Protocol, Runtime string
		Hashes            map[string]string
		Records           []onlineV96Record
	}
	a := artifact{Protocol: "mmm-online-interval-v96", Runtime: runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH, Hashes: map[string]string{}}
	paths := []string{"go.mod", "go.sum", "internal/observationlearners/online_v96_test.go", "internal/observationlearners/brier_online.go", "internal/observationlearners/interval_brier.go", "internal/observationlearners/interval_brier_test.go", "internal/observationlearners/subset.go", "internal/observationlearners/conditional.go", "internal/observationlearners/boolean_specialist.go", "internal/observationlearners/family.go", "internal/observationlearners/family_prior.go", "internal/observationlearners/stacking_v93_test.go", "docs/experiments/mmm-online-interval-v96-protocol.md"}
	for _, p := range paths {
		b, e := os.ReadFile(filepath.Join("../..", p))
		if e != nil {
			t.Fatal(e)
		}
		a.Hashes[p] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	var old artifact
	if replay != "" {
		b, e := os.ReadFile(replay)
		if e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(b, &old); e != nil {
			t.Fatal(e)
		}
		if old.Protocol != a.Protocol || old.Runtime != a.Runtime || !reflect.DeepEqual(old.Hashes, a.Hashes) || len(old.Records) != 768 {
			t.Fatal("manifest")
		}
	}
	for phase := 0; phase < 2; phase++ {
		for scenario := range onlineV96Cases {
			for index := 0; index < 32; index++ {
				r, e := onlineV96Run(phase, scenario, index)
				if e != nil {
					t.Fatal(e)
				}
				for _, row := range r.MaxViolation {
					for _, d := range row {
						if d > 1e-8 {
							t.Fatal("bound")
						}
					}
				}
				if replay != "" && r != old.Records[len(a.Records)] {
					t.Fatal("replay record")
				}
				a.Records = append(a.Records, r)
			}
			t.Log(phase, onlineV96Cases[scenario], "complete")
		}
	}
	if out != "" {
		f, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal(e)
		}
		e = json.NewEncoder(f).Encode(a)
		ce := f.Close()
		if e != nil || ce != nil {
			t.Fatal(e, ce)
		}
	}
}

func BenchmarkOnlineV96Stream(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, e := onlineV96Run(0, 3, 0); e != nil {
			b.Fatal(e)
		}
	}
}
