package observationlearners

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
)

var PartialArms = []string{"fixed_mmm", "rolling_mmm", "frozen_guide", "breadth"}

// PartialRecord uses the first four slots of the existing trace/score contract.
// The remaining three slots are explicitly unused, never counted as arms.
type PartialRecord struct {
	Record
	Views       [][4]observation.Result
	TreeUpdates int
	TreeNodes   int
	TreeNS      int64
}
type PartialOutput struct {
	Hashes      map[string]string
	Records     []PartialRecord
	Comparisons []Comparison
	Verdicts    []Verdict
}

func RunPartialStream(base *observation.Model, split string, j, fit, stream int, seed int64) (PartialRecord, error) {
	s := Scenarios[j]
	r := PartialRecord{Record: Record{Split: split, Scenario: s.Name, Fit: fit, Stream: stream}}
	for i := range r.Recovery {
		r.Recovery[i] = -1
	}
	rng := rand.New(rand.NewSource(Seed(seed, j, fit, stream, 0)))
	ar := rand.New(rand.NewSource(Seed(seed, j, fit, stream, 1)))
	mr := rand.New(rand.NewSource(Seed(seed, j, fit, stream, 2)))
	rollingSeed := Seed(seed, j, fit, stream, 4)
	forest := NewForest(rollingSeed)
	var short, long *observation.Model
	var mixes [4]bayes.ForecastMix
	var audits []observation.Sample
	var queue []packet
	for t := 0; t < 512; t++ {
		tick := Tick{Audit: ar.Float64() < .25, Missing: mr.Float64() < s.Missing}
		x := uint16(rng.Intn(512))
		rd := observationexperiment.Frames(x, fmt.Sprintf("v6-%s-%s-%d-%d-%d", split, s.Name, fit, stream, t))
		var views [4]observation.Result
		for a := 0; a < 4; a++ {
			w := mixes[a].Weights
			if w == [4]float64{} {
				w = [4]float64{.7, .1, .1, .1}
			}
			guide := 0
			if a < 2 {
				if short != nil && w[1] > w[guide] {
					guide = 1
				}
				if long != nil && w[2] > w[guide] {
					guide = 2
				}
			}
			var view observation.Result
			var e error
			if a == 1 && guide == 1 {
				view, e = RunForestObserver(forest, rd, rd.Epoch())
			} else {
				model := base
				if guide == 1 {
					model = short
				}
				if guide == 2 {
					model = long
				}
				policy := "mmm"
				if a == 3 {
					policy = "breadth"
				}
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
			if short != nil {
				if a == 0 {
					c, e = short.ForecastObserved(mask, values)
				} else {
					c, e = forest.ForecastPartial(mask, values)
				}
				if e != nil {
					return r, e
				}
			}
			ex := [4]float64{b, c, l, .5}
			tick.Predictions[a] = Prediction{P: mixes[a].Forecast(ex), Experts: ex}
		}
		r.Views = append(r.Views, views)
		tick.Outcome = truth(x, t, s, rng)
		if !tick.Missing {
			queue = append(queue, packet{Origin: t, Due: t + s.Delay, X: x, Y: tick.Outcome, Audit: tick.Audit, Predictions: tick.Predictions})
		}
		for len(queue) > 0 && queue[0].Due <= t {
			q := queue[0]
			queue = queue[1:]
			r.Available++
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
			forest = NewForest(rollingSeed)
			for _, v := range audits[max(0, len(audits)-64):] {
				forest.Update(v.Bits, v.Outcome)
				r.TreeUpdates++
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
	r.TreeNodes = forest.Nodes()
	return r, nil
}

func RunPartial() (PartialOutput, error) {
	o := PartialOutput{}
	for k, split := range []string{"design", "confirmation"} {
		for j, s := range Scenarios {
			for fit := 0; fit < 3; fit++ {
				rng := rand.New(rand.NewSource(2026092801*1000000 + int64(j*1000+fit)))
				samples := make([]observation.Sample, 4096)
				for i := range samples {
					x := uint16(rng.Intn(512))
					samples[i] = observation.Sample{Bits: x, Outcome: truth(x, -1, s, rng)}
				}
				base, e := observation.Fit(samples)
				if e != nil {
					return o, e
				}
				for stream := 0; stream < 8; stream++ {
					r, e := RunPartialStream(base, split, j, fit, stream, int64(2026092802+k))
					if e != nil {
						return o, e
					}
					o.Records = append(o.Records, r)
				}
			}
		}
	}
	SummarizePartial(&o)
	return o, nil
}

func SummarizePartial(o *PartialOutput) {
	// Reuse the frozen paired-comparison calculation by mapping fixed to slot3
	// and the three candidates to slots4..6, then rename its output labels.
	x := Output{}
	for _, v := range o.Records {
		r := v.Record
		r.Full, r.Post = [7]Metrics{}, [7]Metrics{}
		for a := 0; a < 4; a++ {
			r.Full[a+3] = v.Full[a]
			r.Post[a+3] = v.Post[a]
		}
		x.Records = append(x.Records, r)
	}
	Summarize(&x)
	for i := range x.Comparisons {
		for j := 4; j < 7; j++ {
			if x.Comparisons[i].Arm == Arms[j] {
				x.Comparisons[i].Arm = PartialArms[j-3]
				break
			}
		}
	}
	for i := range x.Verdicts {
		x.Verdicts[i].Arm = PartialArms[i+1]
	}
	o.Comparisons, o.Verdicts = x.Comparisons, x.Verdicts
}
