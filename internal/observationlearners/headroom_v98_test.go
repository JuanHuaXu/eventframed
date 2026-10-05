package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

type headroomV98Step struct {
	Step    int
	X       uint16
	Q       float64
	Y       bool
	P       [9][2]float64
	Weights [2][4]float64
}
type headroomV98Record struct {
	Original windowV97Record
	Trace    []headroomV98Step
	Blocks   [8][2]headroomMetric
}

func headroomV98Run(phase, scenario, index int) (headroomV98Record, error) {
	var trace []headroomV98Step
	r := windowV97Record{Phase: []string{"design", "confirmation"}[phase], Case: windowV97Cases[scenario], Index: index}
	seed := int64(2054119700 + phase*1000000 + scenario*10000 + index*10)
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
	var banks [2]*brierBank
	for view := range banks {
		banks[view], _ = newBrierBank([]float64{.95, .05 / 3, .05 / 3, .05 / 3}, .001)
	}
	var models [5]*ConditionalForest
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
				return headroomV98Record{}, e
			}
			models[1], e = NewBooleanConditional(window)
			if e != nil {
				return headroomV98Record{}, e
			}
			models[2], e = NewPriorConditional(window, .1)
			if e != nil {
				return headroomV98Record{}, e
			}
			short := window
			if len(short) > 32 {
				short = short[len(short)-32:]
			}
			models[3], e = NewSubsetConditional(short, weights)
			if e != nil {
				return headroomV98Record{}, e
			}
			models[4], e = NewBooleanConditional(short)
			if e != nil {
				return headroomV98Record{}, e
			}
			pub := windowV97Publication{Step: step, Counts: [2]int{len(window), len(short)}, Origins: [2][2]int{{step - len(window), step - 1}, {step - len(short), step - 1}}}
			for view, b := range banks {
				pub.Weights[view] = b.weights()
			}
			r.Publications = append(r.Publications, pub)
		}
		x := stackV93Input(uint16(xs.Intn(512)), r.Case)
		var forecasts [2][2]brierForecast
		var predictions [9][2]float64
		for view, mask := range []uint16{511, 63} {
			for modelIndex, m := range models {
				arm := modelIndex
				if modelIndex >= 3 {
					arm = modelIndex + 3
				}
				p, e := m.Forecast(mask, x&mask)
				if e != nil {
					return headroomV98Record{}, e
				}
				predictions[arm][view] = p
			}
			for arm, s := range states[view] {
				f, e := s.predict(uint64(step), [2]float64{predictions[0][view], predictions[1][view]})
				if e != nil {
					return headroomV98Record{}, e
				}
				forecasts[view][arm] = f
				predictions[3+arm][view] = f.P
			}
		}
		for view, state := range intervals {
			p, e := state.predict(uint64(step), [2]float64{predictions[0][view], predictions[1][view]})
			if e != nil {
				return headroomV98Record{}, e
			}
			predictions[5][view] = p
		}
		for view, bank := range banks {
			f, e := bank.predict(uint64(step), []float64{predictions[0][view], predictions[1][view], predictions[6][view], predictions[7][view]})
			if e != nil {
				return headroomV98Record{}, e
			}
			predictions[8][view] = f.P
		}
		// Outcome sampling and oracle scoring occur only after every forecast exists.
		q := truth(x, step)
		y := ys.Float64() < q
		snapshot := headroomV98Step{Step: step, X: x, Q: q, Y: y, P: predictions}
		for view, bank := range banks {
			snapshot.Weights[view] = bank.pending.Weights
		}
		trace = append(trace, snapshot)
		target := 0.
		if y {
			target = 1
		}
		raw, _ := json.Marshal(struct {
			Step int
			X    uint16
			P    [9][2]float64
			Y    bool
		}{step, x, predictions, y})
		tape.Write(raw)
		for arm := range predictions {
			for view, p := range predictions[arm] {
				if math.IsNaN(p) || p < 0 || p > 1 {
					return headroomV98Record{}, fmt.Errorf("invalid forecast")
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
					return headroomV98Record{}, e
				}
			}
		}
		for _, state := range intervals {
			if e := state.observe(uint64(step), y); e != nil {
				return headroomV98Record{}, e
			}
		}
		for view, bank := range banks {
			d := r.Realized[8][view] - r.Realized[0][view] - bank.genericBound(uint64(step+1))
			r.BankViolation[view] = math.Max(r.BankViolation[view], d)
			if e := bank.observe(uint64(step), y); e != nil {
				return headroomV98Record{}, e
			}
		}
		history = append(history, observation.Sample{Bits: x, Outcome: y})
	}
	r.Tape = fmt.Sprintf("%x", tape.Sum(nil))
	result := headroomV98Record{Original: r, Trace: trace}
	for block := 0; block < 8; block++ {
		for view := 0; view < 2; view++ {
			var points []headroomPoint
			for _, s := range trace[block*32 : (block+1)*32] {
				points = append(points, headroomPoint{Q: s.Q, Experts: [4]float64{s.P[0][view], s.P[1][view], s.P[6][view], s.P[7][view]}, Bank: s.P[8][view]})
			}
			m, e := scoreHeadroom(points)
			if e != nil {
				return headroomV98Record{}, e
			}
			result.Blocks[block][view] = m
		}
	}
	return result, nil
}

func TestHeadroomV98Smoke(t *testing.T) {
	r, e := headroomV98Run(0, 11, 0)
	if e != nil {
		t.Fatal(e)
	}
	old, e := windowV97Run(0, 11, 0)
	if e != nil || !reflect.DeepEqual(old, r.Original) {
		t.Fatal("parent changed", e)
	}
	if len(r.Trace) != 256 {
		t.Fatal("trace")
	}
}
func TestHeadroomV98(t *testing.T) {
	out, replay := os.Getenv("EVENTFRAME_HEADROOM_V98_OUT"), os.Getenv("EVENTFRAME_HEADROOM_V98_REPLAY")
	if out == "" && replay == "" {
		t.Skip("explicit artifact required")
	}
	if out != "" && replay != "" {
		t.Fatal("choose one")
	}
	parentPath := "../../docs/experiments/mmm-window-bank-v97.json"
	raw, e := os.ReadFile(parentPath)
	if e != nil {
		t.Fatal(e)
	}
	const expected = "6b4559a63d7b6c1e1b2603513a90dcf4c82281b27e48430405522a1fb48380e5"
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != expected {
		t.Fatal("parent hash")
	}
	var parent struct {
		Hashes  map[string]string
		Records []windowV97Record
	}
	if e = json.Unmarshal(raw, &parent); e != nil {
		t.Fatal(e)
	}
	if len(parent.Records) != 768 {
		t.Fatal("parent count")
	}
	type artifact struct {
		Protocol, Runtime, ParentSHA256 string
		Hashes                          map[string]string
		Records                         []headroomV98Record
	}
	a := artifact{Protocol: "mmm-window-headroom-v98", Runtime: runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH, ParentSHA256: expected, Hashes: map[string]string{}}
	for p, h := range parent.Hashes {
		b, e := os.ReadFile(filepath.Join("../..", p))
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != h {
			t.Fatal("parent source", p, e)
		}
		a.Hashes[p] = h
	}
	for _, p := range []string{"internal/observationlearners/headroom_v98_test.go", "internal/observationlearners/headroom_math_test.go", "docs/experiments/mmm-window-headroom-v98-protocol.md"} {
		b, e := os.ReadFile(filepath.Join("../..", p))
		if e != nil {
			t.Fatal(e)
		}
		a.Hashes[p] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	var previous artifact
	if replay != "" {
		b, e := os.ReadFile(replay)
		if e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(b, &previous); e != nil {
			t.Fatal(e)
		}
		if previous.Protocol != a.Protocol || previous.Runtime != a.Runtime || previous.ParentSHA256 != expected || !reflect.DeepEqual(previous.Hashes, a.Hashes) || len(previous.Records) != 128 {
			t.Fatal("manifest")
		}
	}
	for phase := 0; phase < 2; phase++ {
		for scenario := 10; scenario < 12; scenario++ {
			for index := 0; index < 32; index++ {
				r, e := headroomV98Run(phase, scenario, index)
				if e != nil {
					t.Fatal(e)
				}
				position := (phase*12+scenario)*32 + index
				if !reflect.DeepEqual(r.Original, parent.Records[position]) {
					t.Fatal("original prediction changed", phase, scenario, index)
				}
				if replay != "" && !reflect.DeepEqual(r, previous.Records[len(a.Records)]) {
					t.Fatal("diagnostic replay")
				}
				a.Records = append(a.Records, r)
			}
			t.Log(phase, scenario, "complete")
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
