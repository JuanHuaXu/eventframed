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

type logV109Journal interface {
	publish(*observationExperts) error
	predict(uint64, observation.Reader, uint64) (observation.Result, error)
	deliver(uint64, bool) error
	expireBefore(uint64) error
	statsSnapshot() routedJournalStats
	selectorCount() uint64
}
type prefixV109Journal struct{ *routedFeedbackJournal }

func (g prefixV109Journal) statsSnapshot() routedJournalStats { return g.stats }
func (g prefixV109Journal) selectorCount() uint64             { return g.stats.Applied + g.stats.BankOnly }

type logV109Adapter struct{ *arrivalRoutedJournal }

func (g logV109Adapter) statsSnapshot() routedJournalStats { return g.journal.stats }
func (g logV109Adapter) selectorCount() uint64             { return g.received }

type logV109Step struct {
	X        uint16
	Q        float64
	Y        bool
	Delay    int
	Missing  bool
	P        [7]float64
	Mask     [7]uint16
	Cost     [7]int
	Stats    [6]routedJournalStats
	Selector [6]uint64
}
type logV109Fit struct {
	Clock   int
	Origins [2][]int
}
type logV109Record struct {
	Phase, Case     string
	Index, Schedule int
	Rules           [2]uint16
	Metrics         [7][2]stackV93Metric
	Realized        [7]float64
	LogLoss         [7][2]float64
	RealizedLog     [7]float64
	Steps           []logV109Step
	Fits            []logV109Fit
	Final           [6]routedJournalStats
	FinalSelector   [6]uint64
	Arrived         int
}

type agedV109Adapter struct{ *agedAdviceJournal }

func (g agedV109Adapter) statsSnapshot() routedJournalStats { return g.journal.stats }
func (g agedV109Adapter) selectorCount() uint64             { return g.received }

type likelihoodV109Adapter struct{ *logAdviceJournal }

func (g likelihoodV109Adapter) statsSnapshot() routedJournalStats { return g.journal.stats }
func (g likelihoodV109Adapter) selectorCount() uint64             { return g.received }

func logV109Run(phase, scenario, index, schedule int, base int64) (logV109Record, error) {
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
	journals := [6]logV109Journal{prefixV109Journal{newRoutedFeedbackJournal()}, prefixV109Journal{newRoleRoutedFeedbackJournal()}, logV109Adapter{newArrivalRoutedJournal()}, agedV109Adapter{newAgedAdviceJournal(true, false)}, likelihoodV109Adapter{newLogAdviceJournal(false)}, likelihoodV109Adapter{newLogAdviceJournal(true)}}
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

func TestLogV109Compatibility(t *testing.T) {
	for _, scenario := range []int{0, 3, 10, 11} {
		for schedule := 0; schedule < 2; schedule++ {
			old, err := adviceV108Run(0, scenario, 0, schedule, 2120110800)
			if err != nil {
				t.Fatal(err)
			}
			got, err := logV109Run(0, scenario, 0, schedule, 2120110800)
			if err != nil {
				t.Fatal(err)
			}
			if old.Rules != got.Rules || old.Arrived != got.Arrived {
				t.Fatal("control identity")
			}
			for j, f := range old.Fits {
				if f.Clock != got.Fits[j].Clock || !reflect.DeepEqual(f.Origins, got.Fits[j].Origins) {
					t.Fatal("as-of fits")
				}
			}
			for arm := 0; arm < 5; arm++ {
				if old.Metrics[arm] != got.Metrics[arm] || old.Realized[arm] != got.Realized[arm] {
					t.Fatal("control metrics")
				}
			}
			for i, a := range old.Steps {
				b := got.Steps[i]
				if a.X != b.X || a.Q != b.Q || a.Y != b.Y || a.Delay != b.Delay || a.Missing != b.Missing {
					t.Fatal("latent control")
				}
				for arm := 0; arm < 5; arm++ {
					if a.P[arm] != b.P[arm] || a.Mask[arm] != b.Mask[arm] || a.Cost[arm] != b.Cost[arm] {
						t.Fatal("control forecast", scenario, schedule, i, arm)
					}
				}
				for arm := 0; arm < 4; arm++ {
					if a.Stats[arm] != b.Stats[arm] || a.Selector[arm] != b.Selector[arm] {
						t.Fatal("control accounting")
					}
				}
			}
			for arm := 0; arm < 4; arm++ {
				if old.Final[arm] != got.Final[arm] || old.FinalSelector[arm] != got.FinalSelector[arm] {
					t.Fatal("final control")
				}
			}
		}
	}
}

func TestLogV109Seeds(t *testing.T) {
	used := map[int64]bool{}
	for phase := int64(0); phase < 2; phase++ {
		for scenario := int64(0); scenario < 12; scenario++ {
			for index := int64(0); index < 32; index++ {
				for role := int64(0); role < 5; role++ {
					seed := (2130110900 + phase*1000000 + scenario*10000 + index*10 + role) % 2147483647
					if used[seed] {
						t.Fatal("within-run seed collision")
					}
					used[seed] = true
				}
			}
		}
	}
	for _, base := range []int64{2026119000, 2030119100, 2034119200, 2038119300, 2042119400, 2046119500, 2050119600, 2054119700, 2062119900, 2066110000, 2078110100, 2082110200, 2090110300, 2100110400, 2110110600, 2120110800} {
		// Conservatively include 128 indices and five roles for all earlier runs.
		for phase := int64(0); phase < 2; phase++ {
			for scenario := int64(0); scenario < 12; scenario++ {
				for index := int64(0); index < 128; index++ {
					for role := int64(0); role < 5; role++ {
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

type logV109Artifact struct {
	Version, Go, OS, Arch string
	SeedBase              int64
	PerCase               int
	Hashes                map[string]string
	Records               []logV109Record
}

func TestLogV109(t *testing.T) {
	path := os.Getenv("EVENTFRAME_LOG_V109")
	if path == "" {
		t.Skip("explicit artifact required")
	}
	if os.Getenv("EVENTFRAME_LOG_V109_REPLAY") == "1" {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var a logV109Artifact
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
			got, err := logV109Run(phase, scenario, r.Index, r.Schedule, a.SeedBase)
			if err != nil || !reflect.DeepEqual(got, r) {
				t.Fatal("replay", r.Phase, r.Case, r.Index, r.Schedule, err)
			}
		}
		return
	}
	a := logV109Artifact{Version: "v109", Go: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, SeedBase: 2130110900, PerCase: 32, Hashes: map[string]string{}}
	paths := []string{"internal/observationlearners/advice_v108_test.go", "internal/observationlearners/log_advice.go", "internal/observationlearners/log_advice_test.go", "research/log-advice-component-contract.md", "internal/observationlearners/arrival_v106_test.go", "internal/observationlearners/aged_advice.go", "internal/observationlearners/aged_advice_journal.go", "internal/observationlearners/aged_advice_test.go", "internal/observationlearners/aged_advice_journal_test.go", "research/aged-advice-component-contract.md", "go.mod", "go.sum", "docs/experiments/mmm-log-v109-protocol.md", "research/log-v109-summary.mjs", "internal/observation/controller.go", "internal/observationlearners/log_v109_test.go", "internal/observationlearners/replication_v102_test.go", "internal/observationlearners/stacking_v93_test.go", "internal/observationlearners/subset.go", "internal/observationlearners/conditional.go", "internal/observationlearners/boolean_specialist.go", "internal/observationlearners/brier_bank.go", "internal/observationlearners/forecast_falsification.go", "internal/observationlearners/comparative_falsification.go", "internal/observationlearners/evidence_routing.go", "internal/observationlearners/joint_observation.go", "internal/observationlearners/joint_observer.go", "internal/observationlearners/joint_observation_test.go", "internal/observationlearners/arrival_routed_journal.go", "internal/observationlearners/arrival_routed_journal_test.go", "internal/observationlearners/routed_feedback_journal.go", "internal/observationlearners/routed_feedback_journal_test.go", "internal/observationlearners/conditional_observer.go"}
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
				for schedule := 0; schedule < 2; schedule++ {
					r, err := logV109Run(phase, scenario, index, schedule, a.SeedBase)
					if err != nil {
						t.Fatal(err)
					}
					a.Records = append(a.Records, r)
				}
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
