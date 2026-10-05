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

type adviceV108Journal interface {
	publish(*observationExperts) error
	predict(uint64, observation.Reader, uint64) (observation.Result, error)
	deliver(uint64, bool) error
	expireBefore(uint64) error
	statsSnapshot() routedJournalStats
	selectorCount() uint64
}
type prefixV108Journal struct{ *routedFeedbackJournal }

func (g prefixV108Journal) statsSnapshot() routedJournalStats { return g.stats }
func (g prefixV108Journal) selectorCount() uint64             { return g.stats.Applied + g.stats.BankOnly }

type adviceV108Adapter struct{ *arrivalRoutedJournal }

func (g adviceV108Adapter) statsSnapshot() routedJournalStats { return g.journal.stats }
func (g adviceV108Adapter) selectorCount() uint64             { return g.received }

type adviceV108Step struct {
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
type adviceV108Fit struct {
	Clock   int
	Origins [2][]int
}
type adviceV108Record struct {
	Phase, Case     string
	Index, Schedule int
	Rules           [2]uint16
	Metrics         [7][2]stackV93Metric
	Realized        [7]float64
	Steps           []adviceV108Step
	Fits            []adviceV108Fit
	Final           [6]routedJournalStats
	FinalSelector   [6]uint64
	Arrived         int
}

type agedV108Adapter struct{ *agedAdviceJournal }

func (g agedV108Adapter) statsSnapshot() routedJournalStats { return g.journal.stats }
func (g agedV108Adapter) selectorCount() uint64             { return g.received }

func adviceV108Run(phase, scenario, index, schedule int, base int64) (adviceV108Record, error) {
	r := adviceV108Record{Schedule: schedule, Phase: []string{"design", "confirmation"}[phase], Case: replicationV102Cases[scenario], Index: index}
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
	journals := [6]adviceV108Journal{prefixV108Journal{newRoutedFeedbackJournal()}, prefixV108Journal{newRoleRoutedFeedbackJournal()}, adviceV108Adapter{newArrivalRoutedJournal()}, agedV108Adapter{newAgedAdviceJournal(true, false)}, agedV108Adapter{newAgedAdviceJournal(false, true)}, agedV108Adapter{newAgedAdviceJournal(true, true)}}
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
			r.Fits = append(r.Fits, adviceV108Fit{Clock: clock, Origins: [2][]int{append([]int(nil), origins...), append([]int(nil), shortOrigins...)}})
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
		s := adviceV108Step{X: x}
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
			if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 || s.Cost[arm] < 1 || s.Cost[arm] > 6 || s.Cost[arm] != bits.OnesCount16(s.Mask[arm]) {
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

func TestAdviceV108Compatibility(t *testing.T) {
	for _, scenario := range []int{0, 3, 10, 11} {
		for schedule := 0; schedule < 2; schedule++ {
			old, err := arrivalV106Run(0, scenario, 0, schedule, 2110110600)
			if err != nil {
				t.Fatal(err)
			}
			got, err := adviceV108Run(0, scenario, 0, schedule, 2110110600)
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
			for arm := 0; arm < 4; arm++ {
				if old.Metrics[arm] != got.Metrics[arm] || old.Realized[arm] != got.Realized[arm] {
					t.Fatal("control metrics")
				}
			}
			for i, a := range old.Steps {
				b := got.Steps[i]
				if a.X != b.X || a.Q != b.Q || a.Y != b.Y || a.Delay != b.Delay || a.Missing != b.Missing {
					t.Fatal("latent control")
				}
				for arm := 0; arm < 4; arm++ {
					if a.P[arm] != b.P[arm] || a.Mask[arm] != b.Mask[arm] || a.Cost[arm] != b.Cost[arm] {
						t.Fatal("control forecast", scenario, schedule, i, arm)
					}
				}
				for arm := 0; arm < 3; arm++ {
					if a.Stats[arm] != b.Stats[arm] || a.Selector[arm] != b.Selector[arm] {
						t.Fatal("control accounting")
					}
				}
			}
			for arm := 0; arm < 3; arm++ {
				if old.Final[arm] != got.Final[arm] || old.FinalSelector[arm] != got.FinalSelector[arm] {
					t.Fatal("final control")
				}
			}
		}
	}
}

func TestAdviceV108Seeds(t *testing.T) {
	used := map[int64]bool{}
	for phase := int64(0); phase < 2; phase++ {
		for scenario := int64(0); scenario < 12; scenario++ {
			for index := int64(0); index < 32; index++ {
				for role := int64(0); role < 5; role++ {
					seed := (2120110800 + phase*1000000 + scenario*10000 + index*10 + role) % 2147483647
					if used[seed] {
						t.Fatal("within-run seed collision")
					}
					used[seed] = true
				}
			}
		}
	}
	for _, base := range []int64{2026119000, 2030119100, 2034119200, 2038119300, 2042119400, 2046119500, 2050119600, 2054119700, 2062119900, 2066110000, 2078110100, 2082110200, 2090110300, 2100110400, 2110110600} {
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

type adviceV108Artifact struct {
	Version, Go, OS, Arch string
	SeedBase              int64
	PerCase               int
	Hashes                map[string]string
	Records               []adviceV108Record
}

func TestAdviceV108(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ADVICE_V108")
	if path == "" {
		t.Skip("explicit artifact required")
	}
	if os.Getenv("EVENTFRAME_ADVICE_V108_REPLAY") == "1" {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var a adviceV108Artifact
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
			got, err := adviceV108Run(phase, scenario, r.Index, r.Schedule, a.SeedBase)
			if err != nil || !reflect.DeepEqual(got, r) {
				t.Fatal("replay", r.Phase, r.Case, r.Index, r.Schedule, err)
			}
		}
		return
	}
	a := adviceV108Artifact{Version: "v108", Go: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, SeedBase: 2120110800, PerCase: 32, Hashes: map[string]string{}}
	paths := []string{"internal/observationlearners/arrival_v106_test.go", "internal/observationlearners/aged_advice.go", "internal/observationlearners/aged_advice_journal.go", "internal/observationlearners/aged_advice_test.go", "internal/observationlearners/aged_advice_journal_test.go", "research/aged-advice-component-contract.md", "go.mod", "go.sum", "docs/experiments/mmm-advice-v108-protocol.md", "research/advice-v108-summary.mjs", "internal/observation/controller.go", "internal/observationlearners/advice_v108_test.go", "internal/observationlearners/replication_v102_test.go", "internal/observationlearners/stacking_v93_test.go", "internal/observationlearners/subset.go", "internal/observationlearners/conditional.go", "internal/observationlearners/boolean_specialist.go", "internal/observationlearners/brier_bank.go", "internal/observationlearners/forecast_falsification.go", "internal/observationlearners/comparative_falsification.go", "internal/observationlearners/evidence_routing.go", "internal/observationlearners/joint_observation.go", "internal/observationlearners/joint_observer.go", "internal/observationlearners/joint_observation_test.go", "internal/observationlearners/arrival_routed_journal.go", "internal/observationlearners/arrival_routed_journal_test.go", "internal/observationlearners/routed_feedback_journal.go", "internal/observationlearners/routed_feedback_journal_test.go", "internal/observationlearners/conditional_observer.go"}
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
					r, err := adviceV108Run(phase, scenario, index, schedule, a.SeedBase)
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
