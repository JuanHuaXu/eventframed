package observationlearners

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
)

// Codes: count0, subset64=1, subset16=2, forest64=3.
func windowGuide(w [4]float64, mode int) int {
	if w == [4]float64{} {
		w = [4]float64{.7, .1, .1, .1}
	}
	weights := []float64{w[0], w[1] + w[2] + w[3]}
	codes := []int{0, 1}
	if mode == 1 {
		weights = []float64{w[0], w[1], w[2] + w[3]}
		codes = []int{0, 2, 1}
	}
	if mode == 2 {
		weights = []float64{w[0], w[1], w[2], w[3]}
		codes = []int{0, 3, 2, 1}
	}
	best := 0
	for i := 1; i < len(weights); i++ {
		if weights[i] > weights[best] {
			best = i
		}
	}
	return codes[best]
}

func RunWindowBankStream(base *observation.Model, split string, j, fit, stream int, seed int64, generator, mode, family int) (BreadthRecord, error) {
	if family < 0 || family > 1 || mode < 0 || mode > 2 || generator < 0 || generator > 2 || j < 0 || j >= len(Scenarios) || base == nil {
		return BreadthRecord{}, fmt.Errorf("invalid dependent experiment input")
	}
	s := Scenarios[j]
	r := BreadthRecord{AcquisitionRecord: AcquisitionRecord{DependentRecord: DependentRecord{RetainedRecord: RetainedRecord{Record: Record{Split: split, Scenario: s.Name, Fit: fit, Stream: stream}}, Generator: []string{"fair", "biased", "clustered"}[generator]}, Mode: []string{"subset64", "window_bank", "retained_bank"}[mode]}, Family: []string{"majority", "multiplexer"}[family]}
	for i := range r.Recovery {
		r.Recovery[i] = -1
	}
	rng := rand.New(rand.NewSource(Seed(seed, 30*family+10*generator+j, fit, stream, 0)))
	ar := rand.New(rand.NewSource(Seed(seed, 30*family+10*generator+j, fit, stream, 1)))
	mr := rand.New(rand.NewSource(Seed(seed, 30*family+10*generator+j, fit, stream, 2)))
	rollingSeed := Seed(seed, 30*family+10*generator+j, fit, stream, 4)
	forest := NewForest(rollingSeed)
	var conditional, narrow *ConditionalForest
	var short, long *observation.Model
	var mixes [4]bayes.ForecastMix
	var inner bayes.ForecastMix
	var audits []observation.Sample
	var queue []packet
	for t := 0; t < 512; t++ {
		tick := Tick{Audit: ar.Float64() < .25, Missing: mr.Float64() < s.Missing}
		x := dependentInput(rng, generator)
		r.Inputs = append(r.Inputs, x)
		rd := observationexperiment.Frames(x, fmt.Sprintf("v6-%s-%s-%d-%d-%d", split, s.Name, fit, stream, t))
		var views [4]observation.Result
		var innerTrace [4][4]float64
		for a := 0; a < 4; a++ {
			w := mixes[a].Weights
			if w == [4]float64{} {
				w = [4]float64{.7, .1, .1, .1}
			}
			guide := 0
			{
				if short != nil && w[1] > w[guide] {
					guide = 1
				}
				if long != nil && w[2] > w[guide] {
					guide = 2
				}
			}
			var view observation.Result
			var e error

			chosen := 1
			if a == 2 {
				chosen = windowGuide(inner.Weights, mode)
			}
			if a == 0 || a == 3 {
				chosen = 0
			}
			if chosen != 0 && guide == 1 {
				switch chosen {
				case 1:
					view, e = RunConditionalObserver(conditional, rd, rd.Epoch())
				case 2:
					view, e = RunConditionalObserver(narrow, rd, rd.Epoch())
				case 3:
					view, e = RunForestObserver(forest, rd, rd.Epoch())
				}

			} else {
				model := base
				if guide == 1 {
					model = short
				}
				if guide == 2 {
					model = long
				}
				policy := "mmm"

				view, e = observation.Run(model, rd, rd.Epoch(), policy, 0)
			}
			if e != nil {
				return r, e
			}
			views[a] = view
			last := view.Trace[len(view.Trace)-1]
			mask, values := last.Observed, last.Values
			b, e := base.ForecastObserved(mask, values)
			if e != nil {
				return r, e
			}
			l, c := .5, .5
			if long != nil {
				l, e = long.ForecastObserved(mask, values)
				if e != nil {
					return r, e
				}
			}
			sp, tp := .5, .5
			if short != nil {
				sp, e = short.ForecastObserved(mask, values)
				if e != nil {
					return r, e
				}
				tp, e = conditional.Forecast(mask, values)
				if e != nil {
					return r, e
				}
			}
			innerTrace[a] = [4]float64{sp, tp, tp, tp}
			if a == 2 && mode != 0 && short != nil {
				np, err := narrow.Forecast(mask, values)
				if err != nil {
					return r, err
				}
				if mode == 1 {
					innerTrace[a] = [4]float64{sp, np, tp, tp}
				} else {
					fp, err := forest.ForecastPartial(mask, values)
					if err != nil {
						return r, err
					}
					innerTrace[a] = [4]float64{sp, fp, np, tp}
				}
			}
			switch a {
			case 0:
				c = sp
			case 1:
				c = tp
			case 2:
				c = inner.Forecast(innerTrace[a])
			case 3:
				c = .5 * (sp + tp)
			}
			ex := [4]float64{b, c, l, .5}
			tick.Predictions[a] = Prediction{P: mixes[a].Forecast(ex), Experts: ex}
		}
		r.Views = append(r.Views, views)
		r.Inner = append(r.Inner, innerTrace)
		tick.Outcome = breadthTruth(x, t, s, family, rng)
		if !tick.Missing {
			queue = append(queue, packet{Origin: t, Due: t + s.Delay, X: x, Y: tick.Outcome, Audit: tick.Audit, Predictions: tick.Predictions})
		}
		for len(queue) > 0 && queue[0].Due <= t {
			q := queue[0]
			queue = queue[1:]
			r.Available++
			inner = inner.Observe(r.Inner[q.Origin][2], q.Y, 1)
			tick.Delivered = append(tick.Delivered, q.Origin)
			for a := range mixes {
				mixes[a] = mixes[a].Observe(q.Predictions[a].Experts, q.Y, 1)
			}
			if !q.Audit {
				continue
			}
			r.Audits++
			v := observation.Sample{Bits: q.X, Outcome: q.Y}
			audits = append(audits, v)
			if len(audits) > 256 {
				audits = audits[1:]
			}
			if r.Audits < 32 || r.Audits%16 != 0 {
				continue
			}
			start := time.Now()
			var e error
			short, e = observation.Fit(audits[max(0, len(audits)-64):])
			if e != nil {
				return r, e
			}
			long, e = observation.Fit(audits)
			if e != nil {
				return r, e
			}
			r.FitNS += int64(time.Since(start))
			r.Fits += 2
			start = time.Now()
			if mode == 2 {
				forest = NewForest(rollingSeed)
				for _, v := range audits[max(0, len(audits)-64):] {
					forest.Update(v.Bits, v.Outcome)
					r.TreeUpdates++
				}

			}

			{
				var weights [512]float64
				for x := range weights {
					weights[x] = 1. / 512
				}
				for _, v := range audits {
					weights[v.Bits]++
				}
				conditional, e = NewSubsetConditional(audits[max(0, len(audits)-64):], weights)
				if e == nil && mode != 0 {
					narrow, e = NewSubsetConditional(audits[max(0, len(audits)-16):], weights)
				}
				if e != nil {
					return r, e
				}
			}

			r.TreeNS += int64(time.Since(start))
		}
		tick.Audits = r.Audits
		r.Ticks = append(r.Ticks, tick)
	}
	r.Pending = len(queue)
	post := s.Change
	if post >= 512 {
		post = 256
	}
	for a := 0; a < 4; a++ {
		r.Full[a] = score(r.Ticks, a, 0)
		r.Post[a] = score(r.Ticks, a, post)
	}
	if mode == 2 {
		r.TreeNodes = forest.Nodes()
	}
	return r, nil
}

func RunWindowBank(emit func(BreadthRecord) error) error {
	if emit == nil {
		return fmt.Errorf("missing output callback")
	}
	for splitIndex, split := range []string{"design", "confirmation"} {
		for family := 0; family < 2; family++ {
			for generator := 0; generator < 3; generator++ {
				for _, j := range []int{0, 2, 6, 7} {
					for fit := 0; fit < 6; fit++ {
						rng := rand.New(rand.NewSource(Seed(2026108201, 30*family+10*generator+j, fit, 0, 0)))
						samples := make([]observation.Sample, 4096)
						for i := range samples {
							x := dependentInput(rng, generator)
							samples[i] = observation.Sample{Bits: x, Outcome: breadthTruth(x, -1, Scenarios[j], family, rng)}
						}
						base, e := observation.Fit(samples)
						if e != nil {
							return e
						}
						for stream := 0; stream < 2; stream++ {
							for mode := 0; mode < 3; mode++ {
								r, e := RunWindowBankStream(base, split, j, fit, stream, int64(2026108202+splitIndex), generator, mode, family)
								if e != nil {
									return e
								}
								if e = emit(r); e != nil {
									return e
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}
