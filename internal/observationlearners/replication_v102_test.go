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

var replicationV102Cases = []string{"parity1", "parity2", "parity3", "parity4", "complement4", "majority3", "mux3", "constant", "null", "dependent4", "majority_to_parity", "parity_to_majority"}

type replicationV102Record struct {
	Phase, Case string
	Index       int
	Rules       [2]uint16
	// Arm: generic64, Boolean64, skeptical BMA, rho0, rho.001, interval, generic32, Boolean32, bank, neutral-gated, comparative-gated, evidence-routed.
	// View: full, mask63. Segment: all256, late128. Values are means.
	Metrics                               [12][2][2]stackV93Metric
	Realized                              [12][2]float64
	MaxViolation                          [2][2]float64
	Masked, Neutral                       [2]int
	ComparativeMasked, ComparativeNeutral [2]int
	ComparativeRejectAt                   [8][2][4]int
	RejectAt                              [8][2][4]int
	BankViolation                         [2]float64
	Publications                          []replicationV102Publication
	Tape                                  string
	BlockMetrics                          [12][8][2]stackV93Metric
	RoutedWeights                         [8][2][5]float64
	RawBlockWeights                       [8][2][4]float64
	RoutedNeutral                         [2]int
}

type replicationV102Publication struct {
	Step    int
	Counts  [2]int
	Origins [2][2]int
	Weights [2][4]float64
}

func replicationV102Run(phase, scenario, index int) (replicationV102Record, error) {
	r := replicationV102Record{Phase: []string{"design", "confirmation"}[phase], Case: replicationV102Cases[scenario], Index: index}
	for v := range r.RejectAt {
		for view := range r.RejectAt[v] {
			for k := range r.RejectAt[v][view] {
				r.RejectAt[v][view][k] = -1
				r.ComparativeRejectAt[v][view][k] = -1
			}
		}
	}
	seed := int64(2082110200 + phase*1000000 + scenario*10000 + index*10)
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
	var gates [2]forecastFalsification
	var comparativeGates [2]comparativeFalsification
	var routedGates [2]evidenceRouting
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
			short := window
			if len(short) > 32 {
				short = short[len(short)-32:]
			}
			models[3], e = NewSubsetConditional(short, weights)
			if e != nil {
				return r, e
			}
			models[4], e = NewBooleanConditional(short)
			if e != nil {
				return r, e
			}
			pub := replicationV102Publication{Step: step, Counts: [2]int{len(window), len(short)}, Origins: [2][2]int{{step - len(window), step - 1}, {step - len(short), step - 1}}}
			for view, b := range banks {
				pub.Weights[view] = b.weights()
			}
			r.Publications = append(r.Publications, pub)
		}
		x := stackV93Input(uint16(xs.Intn(512)), r.Case)
		var forecasts [2][2]brierForecast
		var predictions [12][2]float64
		for view, mask := range []uint16{511, 63} {
			for modelIndex, m := range models {
				arm := modelIndex
				if modelIndex >= 3 {
					arm = modelIndex + 3
				}
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
		for view, bank := range banks {
			f, e := bank.predict(uint64(step), []float64{predictions[0][view], predictions[1][view], predictions[6][view], predictions[7][view]})
			if e != nil {
				return r, e
			}
			predictions[8][view] = f.P
		}
		for view := range gates {
			f, e := gates[view].predict(uint64(step), [4]float64{predictions[0][view], predictions[1][view], predictions[6][view], predictions[7][view]}, banks[view].pending.Weights)
			if e != nil {
				return r, e
			}
			predictions[9][view] = f.P
			masked := false
			for _, ok := range f.Accepted {
				masked = masked || !ok
			}
			if masked {
				r.Masked[view]++
			}
			if f.Neutral {
				r.Neutral[view]++
			}
		}
		for view := range comparativeGates {
			f, e := comparativeGates[view].predict(uint64(step), [4]float64{predictions[0][view], predictions[1][view], predictions[6][view], predictions[7][view]}, banks[view].pending.Weights)
			if e != nil {
				return r, e
			}
			predictions[10][view] = f.P
			masked := false
			for _, ok := range f.Accepted {
				masked = masked || !ok
			}
			if masked {
				r.ComparativeMasked[view]++
			}
			if f.Neutral {
				r.ComparativeNeutral[view]++
			}
		}
		for view := range routedGates {
			f, e := routedGates[view].predict(uint64(step), [4]float64{predictions[0][view], predictions[1][view], predictions[6][view], predictions[7][view]}, banks[view].pending.Weights)
			if e != nil {
				return r, e
			}
			predictions[11][view] = f.P
			for j, w := range f.Weights {
				r.RoutedWeights[step/32][view][j] += w / 32
			}
			for j, w := range banks[view].pending.Weights {
				r.RawBlockWeights[step/32][view][j] += w / 32
			}
			if f.Neutral {
				r.RoutedNeutral[view]++
			}
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
			P    [12][2]float64
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
				r.BlockMetrics[arm][step/32][view].Brier += expected / 32
				r.BlockMetrics[arm][step/32][view].Accuracy += acc / 32
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
		for view, bank := range banks {
			d := r.Realized[8][view] - r.Realized[0][view] - bank.genericBound(uint64(step+1))
			r.BankViolation[view] = math.Max(r.BankViolation[view], d)
			if e := bank.observe(uint64(step), y); e != nil {
				return r, e
			}
		}
		for view := range gates {
			if e := gates[view].observe(uint64(step), y); e != nil {
				return r, e
			}
			for k, m := range gates[view].tests {
				if m.Rejected && r.RejectAt[step/32][view][k] < 0 {
					r.RejectAt[step/32][view][k] = step
				}
			}
		}
		for view := range comparativeGates {
			if e := comparativeGates[view].observe(uint64(step), y); e != nil {
				return r, e
			}
			for k, m := range comparativeGates[view].tests {
				if m.Rejected && r.ComparativeRejectAt[step/32][view][k] < 0 {
					r.ComparativeRejectAt[step/32][view][k] = step
				}
			}
		}
		for view := range routedGates {
			if e := routedGates[view].observe(uint64(step), y); e != nil {
				return r, e
			}
			if routedGates[view].gate != comparativeGates[view] {
				return r, fmt.Errorf("routing changed the comparative tests")
			}
		}
		history = append(history, observation.Sample{Bits: x, Outcome: y})
	}
	r.Tape = fmt.Sprintf("%x", tape.Sum(nil))
	return r, nil
}

func TestReplicationV102Smoke(t *testing.T) {
	for _, scenario := range []int{3, 10, 11} {
		a, e := replicationV102Run(0, scenario, 0)
		if e != nil {
			t.Fatal(e)
		}
		b, e := replicationV102Run(0, scenario, 0)
		if e != nil || !reflect.DeepEqual(a, b) {
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

func TestReplicationV102(t *testing.T) {
	out, replay := os.Getenv("EVENTFRAME_REPLICATION_V102_OUT"), os.Getenv("EVENTFRAME_REPLICATION_V102_REPLAY")
	if out == "" && replay == "" {
		t.Skip("explicit artifact required")
	}
	if out != "" && replay != "" {
		t.Fatal("choose one")
	}
	type artifact struct {
		Protocol, Runtime string
		Hashes            map[string]string
		Records           []replicationV102Record
		Nulls             [2][2]int
	}
	a := artifact{Protocol: "mmm-replication-v102", Runtime: runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH, Hashes: map[string]string{}}
	paths := []string{"go.mod", "go.sum", "internal/observationlearners/replication_v102_test.go", "internal/observationlearners/brier_online.go", "internal/observationlearners/brier_bank.go", "internal/observationlearners/forecast_falsification.go", "internal/observationlearners/comparative_falsification.go", "internal/observationlearners/evidence_routing.go", "internal/observationlearners/evidence_routing_test.go", "internal/observationlearners/comparative_falsification_test.go", "internal/observationlearners/forecast_falsification_test.go", "internal/observationlearners/brier_bank_test.go", "internal/observationlearners/interval_brier.go", "internal/observationlearners/interval_brier_test.go", "internal/observationlearners/subset.go", "internal/observationlearners/conditional.go", "internal/observationlearners/boolean_specialist.go", "internal/observationlearners/family.go", "internal/observationlearners/family_prior.go", "internal/observationlearners/stacking_v93_test.go", "docs/experiments/mmm-replication-v102-protocol.md"}
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
		if old.Protocol != a.Protocol || old.Runtime != a.Runtime || !reflect.DeepEqual(old.Hashes, a.Hashes) || len(old.Records) != 3072 {
			t.Fatal("manifest")
		}
	}
	for phase := 0; phase < 2; phase++ {
		for scenario := range replicationV102Cases {
			for index := 0; index < 128; index++ {
				r, e := replicationV102Run(phase, scenario, index)
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
				if replay != "" && !reflect.DeepEqual(r, old.Records[len(a.Records)]) {
					t.Fatal("replay record")
				}
				for _, d := range r.BankViolation {
					if d > 1e-8 {
						t.Fatal("bank bound")
					}
				}
				a.Records = append(a.Records, r)
			}
			t.Log(phase, replicationV102Cases[scenario], "complete")
		}
	}
	for mode := 0; mode < 2; mode++ {
		for family := 0; family < 4096; family++ {
			any := [2]bool{}
			tests := 64
			if mode == 1 {
				tests = 8
			}
			for test := 0; test < tests; test++ {
				rng := rand.New(rand.NewSource(int64(3130110200 + mode*10000000 + family*1000 + test)))
				var neutral forecastTest
				var comparative comparativeTest
				previous := false
				for step := 0; step < 32; step++ {
					p := []float64{.05, .2, .5, .8, .95}[(test+step)%5]
					q := [4]float64{.5, p, p, p}
					if mode == 0 {
						alt := .2
						if previous {
							alt = .8
						}
						q = [4]float64{.5, alt, 1 - p, p}
					}
					// Both the conditional null and alternatives precede this outcome.
					y := rng.Float64() < p
					neutral.update(p, y)
					comparative.update(p, q, y)
					previous = y
				}
				any[0] = any[0] || neutral.Rejected
				any[1] = any[1] || comparative.Rejected
			}
			for method, rejected := range any {
				if rejected {
					a.Nulls[mode][method]++
				}
			}
		}
	}
	if replay != "" && a.Nulls != old.Nulls {
		t.Fatal("null replay")
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

func BenchmarkReplicationV102Stream(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, e := replicationV102Run(0, 3, 0); e != nil {
			b.Fatal(e)
		}
	}
}
