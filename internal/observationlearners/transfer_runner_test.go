package observationlearners

import (
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/transfergenerator"
	"math"
	"math/bits"
	"reflect"
	"testing"
)

type transferStep struct {
	X        uint16
	Q        float64
	Y        bool
	Delay    int
	Missing  bool
	P        [3]float64
	Mask     [3]uint16
	Cost     [3]int
	Stats    [2]routedJournalStats
	Selector [2]uint64
}
type transferFit struct {
	Clock   int
	Origins [2][]int
}
type transferRecord struct {
	Schedule      int
	Teacher       transfergenerator.Description
	Metrics       [3][2]stackV93Metric
	Realized      [3]float64
	LogLoss       [3][2]float64
	RealizedLog   [3]float64
	Steps         []transferStep
	Fits          []transferFit
	Final         [2]routedJournalStats
	FinalSelector [2]uint64
	Arrived       int
}

func transferRunEvidence(data transfergenerator.Evidence, teacher transfergenerator.Teacher, schedule int) (transferRecord, error) {
	r := transferRecord{Schedule: schedule, Teacher: teacher.Describe()}
	history := make([]observation.Sample, 0, 272)
	for _, row := range data.Initial {
		history = append(history, observation.Sample{Bits: row.X, Outcome: row.Y})
	}
	journals := [2]snapshotV115Journal{likelihoodV115Adapter{newLogAdviceJournal(false)}, snapshotV115FilterAdapter{newMarkovAdviceJournal()}}
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
			r.Fits = append(r.Fits, transferFit{Clock: clock, Origins: [2][]int{append([]int(nil), origins...), append([]int(nil), shortOrigins...)}})
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
		x := data.Frames[clock].X
		s := transferStep{X: x}
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
		// Evaluator truth and current feedback enter only after all forecasts.
		s.Q, err = teacher.Truth(x, clock)
		if err != nil {
			return r, err
		}
		s.Y = data.Frames[clock].Y
		d, missing := int(data.Frames[clock].Delay), data.Frames[clock].Missing
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
			if clock >= 192 {
				r.Metrics[arm][1].Brier += loss / 64
				r.Metrics[arm][1].Accuracy += acc / 64
			}
			r.Realized[arm] += (p - target) * (p - target) / 256
			logLoss := -s.Q*math.Log(p) - (1-s.Q)*math.Log1p(-p)
			r.LogLoss[arm][0] += logLoss / 256
			if clock >= 192 {
				r.LogLoss[arm][1] += logLoss / 64
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

func TestTransferRunnerAsOf(t *testing.T) {
	for f := transfergenerator.Additive; f <= transfergenerator.LocalTable; f++ {
		for mode := transfergenerator.Stationary; mode <= transfergenerator.Gradual; mode++ {
			spec := transfergenerator.Spec{SeedBase: transfergenerator.ValidationSeedBase, Family: f, Mode: mode}
			data, teacher, err := transfergenerator.Generate(spec)
			if err != nil {
				t.Fatal(err)
			}
			var runs [2]transferRecord
			for schedule := 0; schedule < 2; schedule++ {
				r, err := transferRunEvidence(data, teacher, schedule)
				if err != nil {
					t.Fatal(err)
				}
				runs[schedule] = r
				arrivals := 0
				for origin, row := range data.Frames {
					if schedule == 0 || !row.Missing {
						arrivals++
					}
					if r.Steps[origin].X != row.X || r.Steps[origin].Y != row.Y {
						t.Fatal("packet mismatch")
					}
				}
				if arrivals != r.Arrived {
					t.Fatal("arrival count")
				}
				for _, fit := range r.Fits {
					for window, origins := range fit.Origins {
						var want []int
						for i := -16; i < fit.Clock; i++ {
							if i < 0 || schedule == 0 || (!data.Frames[i].Missing && i+int(data.Frames[i].Delay) <= fit.Clock) {
								want = append(want, i)
							}
						}
						cap := 64
						if window == 1 {
							cap = 32
						}
						if len(want) > cap {
							want = want[len(want)-cap:]
						}
						if !reflect.DeepEqual(want, origins) {
							t.Fatal("future or missing fit label", spec, schedule, fit.Clock)
						}
					}
				}
			}
			for i := range data.Frames {
				a, b := runs[0].Steps[i], runs[1].Steps[i]
				if a.X != b.X || a.Y != b.Y || a.Q != b.Q {
					t.Fatal("unpaired schedules")
				}
			}
			altered := data
			altered.Frames[160].X ^= 511
			altered.Frames[160].Y = !altered.Frames[160].Y
			altered.Frames[160].Missing = !altered.Frames[160].Missing
			changed, err := transferRunEvidence(altered, teacher, 1)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(runs[1].Steps[:160], changed.Steps[:160]) {
				t.Fatal("future packet leaked")
			}
			// An unrevealed current label and evaluator-only oracle must not
			// influence the issued law, view, acquisition cost or journal state.
			altered = data
			altered.Frames[160].Y = !altered.Frames[160].Y
			labelChanged, err := transferRunEvidence(altered, teacher, 1)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i <= 160; i++ {
				a, b := runs[1].Steps[i], labelChanged.Steps[i]
				if a.P != b.P || a.Mask != b.Mask || a.Cost != b.Cost {
					t.Fatal("current label leaked into forecast")
				}
			}
			otherSpec := spec
			otherSpec.Index = 1
			_, otherTeacher, err := transfergenerator.Generate(otherSpec)
			if err != nil {
				t.Fatal(err)
			}
			oracleChanged, err := transferRunEvidence(data, otherTeacher, 1)
			if err != nil {
				t.Fatal(err)
			}
			for i, a := range runs[1].Steps {
				b := oracleChanged.Steps[i]
				if a.P != b.P || a.Mask != b.Mask || a.Cost != b.Cost || a.Stats != b.Stats || a.Selector != b.Selector {
					t.Fatal("oracle changed policy")
				}
			}
		}
	}
}
