package observationlearners

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
)

var RetainedArms = []string{"fixed_mmm", "replacement_mmm", "adaptive_retained", "static_retained"}

// RetainedRecord uses the first four slots of the existing trace/score contract.
// The remaining three slots are explicitly unused, never counted as arms.
type RetainedRecord struct {
	Record
	Views       [][4]observation.Result
	Inner       [][4][4]float64
	TreeUpdates int
	TreeNodes   int
	TreeNS      int64
}
type RetainedOutput struct {
	Hashes      map[string]string
	Records     []RetainedRecord
	Comparisons []Comparison
	Verdicts    []Verdict
}

func RunRetainedStream(base *observation.Model, split string, j, fit, stream int, seed int64) (RetainedRecord, error) {
	s := Scenarios[j]
	r := RetainedRecord{Record: Record{Split: split, Scenario: s.Name, Fit: fit, Stream: stream}}
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
	var inner bayes.ForecastMix
	var audits []observation.Sample
	var queue []packet
	for t := 0; t < 512; t++ {
		tick := Tick{Audit: ar.Float64() < .25, Missing: mr.Float64() < s.Missing}
		x := uint16(rng.Intn(512))
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
			useTree := a == 1
			if a == 2 {
				iw := inner.Weights
				if iw == [4]float64{} {
					iw = [4]float64{.7, .1, .1, .1}
				}
				useTree = iw[1]+iw[2]+iw[3] > iw[0]
			}
			if useTree && guide == 1 {
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
				tp, e = forest.ForecastPartial(mask, values)
				if e != nil {
					return r, e
				}
			}
			innerTrace[a] = [4]float64{sp, tp, tp, tp}
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
		tick.Outcome = truth(x, t, s, rng)
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

func RunRetained() (RetainedOutput, error) {
	o := RetainedOutput{}
	for k, split := range []string{"design", "confirmation"} {
		for j, s := range Scenarios {
			for fit := 0; fit < 3; fit++ {
				rng := rand.New(rand.NewSource(2026093001*1000000 + int64(j*1000+fit)))
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
					r, e := RunRetainedStream(base, split, j, fit, stream, int64(2026093002+k))
					if e != nil {
						return o, e
					}
					o.Records = append(o.Records, r)
				}
			}
		}
	}
	SummarizeRetained(&o)
	return o, nil
}

func SummarizeRetained(o *RetainedOutput) {
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
				x.Comparisons[i].Arm = RetainedArms[j-3]
				break
			}
		}
	}
	for i := range x.Verdicts {
		x.Verdicts[i].Arm = RetainedArms[i+1]
	}
	o.Comparisons, o.Verdicts = x.Comparisons, x.Verdicts
}
