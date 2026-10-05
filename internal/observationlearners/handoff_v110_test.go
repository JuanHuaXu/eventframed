package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"math/bits"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type handoffV110Record struct {
	Record logV109Record
	Trace  handoffV110Trace
}
type handoffV110Artifact struct {
	Version, ParentSHA256 string
	Hashes                map[string]string
	Records               []handoffV110Record
}

func TestHandoffV110Smoke(t *testing.T) {
	for _, scenario := range []int{10, 11} {
		parent, err := logV109Run(0, scenario, 0, 1, 2130110900)
		if err != nil {
			t.Fatal(err)
		}
		var trace handoffV110Trace
		got, err := tracedV110Run(0, scenario, 0, 1, 2130110900, &parent, &trace)
		if err != nil || !reflect.DeepEqual(got, parent) {
			t.Fatal("instrumentation changed parent", err)
		}
		if len(trace.Profiles) != 8 || len(trace.Arms[0].Steps) != 256 || len(trace.Arms[1].Updates) != parent.Arrived {
			t.Fatal("trace coverage")
		}
	}
}

func tracedV110Run(phase, scenario, index, schedule int, base int64, parent *logV109Record, trace *handoffV110Trace) (logV109Record, error) {
	r := logV109Record{Schedule: schedule, Phase: []string{"design", "confirmation"}[phase], Case: replicationV102Cases[scenario], Index: index}
	seed := base + int64(phase*1000000+scenario*10000+index*10)
	rules, xs, ys, delays := rand.New(rand.NewSource(seed)), rand.New(rand.NewSource(seed+1)), rand.New(rand.NewSource(seed+2)), rand.New(rand.NewSource(seed+3))
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

	misses := rand.New(rand.NewSource(seed + 4))
	journals := [6]logV109Journal{prefixV109Journal{newRoutedFeedbackJournal()}, prefixV109Journal{newRoleRoutedFeedbackJournal()}, logV109Adapter{newArrivalRoutedJournal()}, agedV109Adapter{newAgedAdviceJournal(true, false)}, newTracingV110Journal(false, parent, &trace.Arms[0], &trace.Profiles), newTracingV110Journal(true, parent, &trace.Arms[1], nil)}
	var weights [512]float64
	for j := range weights {
		weights[j] = 1
	}
	var publication *observationExperts
	var arrived [256]bool
	release := func(clock int) error {
		for origin, s := range r.Steps {
			if !arrived[origin] && !s.Missing && origin+s.Delay <= clock {
				for _, g := range journals {
					if err := g.deliver(uint64(origin), s.Y); err != nil {
						return err
					}
				}
				arrived[origin] = true
			}
		}
		if clock >= 32 {
			for _, g := range journals {
				if err := g.expireBefore(uint64(clock - 31)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for clock := 0; clock < 288; clock++ {
		for _, g := range journals {
			if timed, ok := g.(interface{ setClock(uint64) error }); ok {
				if err := timed.setClock(uint64(clock)); err != nil {
					return r, err
				}
			}
		}
		if err := release(clock); err != nil {
			return r, err
		}
		if clock >= 256 {
			continue
		}
		if clock%32 == 0 {
			var eligible []observation.Sample
			var origins []int
			eligible = append(eligible, history[:16]...)
			for i := -16; i < 0; i++ {
				origins = append(origins, i)
			}
			for i := 0; i < clock; i++ {
				if arrived[i] {
					eligible = append(eligible, history[i+16])
					origins = append(origins, i)
				}
			}
			if len(eligible) > 64 {
				eligible = eligible[len(eligible)-64:]
				origins = origins[len(origins)-64:]
			}
			short, shortOrigins := eligible, origins
			if len(short) > 32 {
				short = short[len(short)-32:]
				shortOrigins = shortOrigins[len(shortOrigins)-32:]
			}
			r.Fits = append(r.Fits, logV109Fit{Clock: clock, Origins: [2][]int{append([]int(nil), origins...), append([]int(nil), shortOrigins...)}})
			var models [4]*ConditionalForest
			var err error
			models[0], err = NewSubsetConditional(eligible, weights)
			if err != nil {
				return r, err
			}
			models[1], err = NewBooleanConditional(eligible)
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
			for _, g := range journals {
				if err := g.publish(publication); err != nil {
					return r, err
				}
			}
		}
		x := stackV93Input(uint16(xs.Intn(512)), r.Case)
		s := logV109Step{X: x}
		generic, err := RunConditionalObserver(&publication.models[0], &jointReader{x: x, epoch: 1}, 1)
		if err != nil {
			return r, err
		}
		s.P[0], s.Mask[0], s.Cost[0] = generic.Probability, generic.Trace[len(generic.Trace)-1].Observed, generic.Cost
		for j, g := range journals {
			got, err := g.predict(uint64(clock), &jointReader{x: x, epoch: 1}, 1)
			if err != nil {
				return r, err
			}
			s.P[j+1], s.Mask[j+1], s.Cost[j+1] = got.Probability, got.Trace[len(got.Trace)-1].Observed, got.Cost
		}
		// The current outcome is unavailable until after all seven forecasts.
		s.Q = truth(x, clock)
		s.Y = ys.Float64() < s.Q
		d, missing := delays.Intn(32), misses.Float64() < .2
		if schedule == 1 {
			s.Delay, s.Missing = d, missing
		}
		target := 0.
		if s.Y {
			target = 1
		}
		for arm, p := range s.P {
			if math.IsNaN(p) || math.IsInf(p, 0) || p <= 0 || p >= 1 || s.Cost[arm] < 1 || s.Cost[arm] > 6 || s.Cost[arm] != bits.OnesCount16(s.Mask[arm]) {
				return r, fmt.Errorf("invalid issued bundle")
			}
			loss := (p-s.Q)*(p-s.Q) + s.Q*(1-s.Q)
			acc := 1 - s.Q
			if p >= .5 {
				acc = s.Q
			}
			r.Metrics[arm][0].Brier += loss / 256
			r.Metrics[arm][0].Accuracy += acc / 256
			if clock >= 128 {
				r.Metrics[arm][1].Brier += loss / 128
				r.Metrics[arm][1].Accuracy += acc / 128
			}
			r.Realized[arm] += (p - target) * (p - target) / 256
			logLoss := -s.Q*math.Log(p) - (1-s.Q)*math.Log1p(-p)
			r.LogLoss[arm][0] += logLoss / 256
			if clock >= 128 {
				r.LogLoss[arm][1] += logLoss / 128
			}
			r.RealizedLog[arm] += (-target*math.Log(p) - (1-target)*math.Log1p(-p)) / 256
		}
		r.Steps = append(r.Steps, s)
		history = append(history, observation.Sample{Bits: x, Outcome: s.Y})
		if !s.Missing && s.Delay == 0 {
			for _, g := range journals {
				if err := g.deliver(uint64(clock), s.Y); err != nil {
					return r, err
				}
			}
			arrived[clock] = true
		}
		for j, g := range journals {
			z := g.statsSnapshot()
			if z.Issued != z.Applied+z.BankOnly+z.Stale+z.Censored+z.Pending || z.Pending > 64 {
				return r, fmt.Errorf("journal accounting")
			}
			r.Steps[clock].Stats[j] = z
			r.Steps[clock].Selector[j] = g.selectorCount()
		}
	}
	for j, g := range journals {
		r.Final[j] = g.statsSnapshot()
		r.FinalSelector[j] = g.selectorCount()
		if g.statsSnapshot().Pending != 0 {
			return r, fmt.Errorf("unsettled final journal")
		}
	}
	for _, v := range arrived {
		if v {
			r.Arrived++
		}
	}
	return r, nil
}

func TestHandoffV110(t *testing.T) {
	output := os.Getenv("EVENTFRAME_HANDOFF_V110")
	if output == "" {
		t.Skip("explicit consumed diagnostic artifact required")
	}
	raw, err := os.ReadFile("../../docs/experiments/mmm-log-v109.json")
	if err != nil {
		t.Fatal(err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(raw))
	if hash != "6173c8d4c0232608876af25aaeba5e1ffbfddcf541c7c0870d51720ea03ca1ad" {
		t.Fatal("parent changed")
	}
	var parent logV109Artifact
	if err = json.Unmarshal(raw, &parent); err != nil {
		t.Fatal(err)
	}
	a := handoffV110Artifact{Version: "v110", ParentSHA256: hash, Hashes: map[string]string{}}
	for k, v := range parent.Hashes {
		a.Hashes[k] = v
	}
	for _, name := range []string{"internal/observationlearners/handoff_v110_test.go", "internal/observationlearners/handoff_v110_trace_test.go", "research/handoff-v110-protocol.md", "research/handoff-v110-summary.mjs"} {
		b, err := os.ReadFile(filepath.Join("../..", name))
		if err != nil {
			t.Fatal(err)
		}
		a.Hashes[name] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	for name, hash := range a.Hashes {
		b, err := os.ReadFile(filepath.Join("../..", name))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatal("source changed", name, err)
		}
	}
	for _, r := range parent.Records {
		if r.Schedule != 1 || (r.Case != "majority_to_parity" && r.Case != "parity_to_majority") {
			continue
		}
		phase, scenario := 0, 10
		if r.Phase == "confirmation" {
			phase = 1
		}
		if r.Case == "parity_to_majority" {
			scenario = 11
		}
		var d handoffV110Record
		d.Record, err = tracedV110Run(phase, scenario, r.Index, r.Schedule, parent.SeedBase, &r, &d.Trace)
		if err != nil || !reflect.DeepEqual(d.Record, r) {
			t.Fatal("trace changed original", r.Phase, r.Case, r.Index, err)
		}
		if len(d.Trace.Profiles) != 8 {
			t.Fatal("profiles")
		}
		for _, arm := range d.Trace.Arms {
			if len(arm.Steps) != 256 || len(arm.Updates) != r.Arrived {
				t.Fatal("trace coverage")
			}
		}
		a.Records = append(a.Records, d)
	}
	if len(a.Records) != 128 {
		t.Fatal("record count")
	}
	if os.Getenv("EVENTFRAME_HANDOFF_V110_REPLAY") == "1" {
		b, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		var old handoffV110Artifact
		if err = json.Unmarshal(b, &old); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(old, a) {
			t.Fatal("trace replay mismatch")
		}
		return
	}
	raw, err = json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write(raw); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
}
