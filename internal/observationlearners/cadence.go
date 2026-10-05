package observationlearners

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
)

var CadenceArms = []string{"fixed", "immediate", "batched", "rolling"}

// CadenceRecord uses the first four slots of the existing trace/score contract.
// The remaining three slots are explicitly unused, never counted as arms.
type CadenceRecord struct {
	Record
	TreeUpdates [3]int
	TreeNodes   [3]int
	TreeNS      [3]int64
}
type CadenceOutput struct {
	Hashes      map[string]string
	Records     []CadenceRecord
	Comparisons []Comparison
	Verdicts    []Verdict
}

func RunCadenceStream(base *observation.Model, split string, j, fit, stream int, seed int64) (CadenceRecord, error) {
	s := Scenarios[j]
	r := CadenceRecord{Record: Record{Split: split, Scenario: s.Name, Fit: fit, Stream: stream}}
	for i := range r.Recovery {
		r.Recovery[i] = -1
	}
	rng := rand.New(rand.NewSource(Seed(seed, j, fit, stream, 0)))
	ar := rand.New(rand.NewSource(Seed(seed, j, fit, stream, 1)))
	mr := rand.New(rand.NewSource(Seed(seed, j, fit, stream, 2)))
	forestSeed := Seed(seed, j, fit, stream, 3)
	rollingSeed := Seed(seed, j, fit, stream, 4)
	trees := [3]*Forest{NewForest(forestSeed), NewForest(forestSeed), NewForest(rollingSeed)}
	var short, long *observation.Model
	var mixes [4]bayes.ForecastMix
	var audits []observation.Sample
	var pendingAudits []observation.Sample
	var queue []packet
	for t := 0; t < 512; t++ {
		tick := Tick{Audit: ar.Float64() < .25, Missing: mr.Float64() < s.Missing}
		x := uint16(rng.Intn(512))
		rd := observationexperiment.Frames(x, fmt.Sprintf("v6-%s-%s-%d-%d-%d", split, s.Name, fit, stream, t))
		var mask, values uint16
		for scope := 0; scope < 3; scope++ {
			m, v, e := rd.Read(observation.View{Scope: scope, Depth: 2})
			if e != nil {
				return r, e
			}
			mask |= m
			values |= v
		}
		if mask != 511 {
			return r, fmt.Errorf("incomplete full observation")
		}
		b, l := forecast(base, values), forecast(long, values)
		challengers := [4]float64{forecast(short, values), trees[0].Predict(values), trees[1].Predict(values), trees[2].Predict(values)}
		for a, c := range challengers {
			ex := [4]float64{b, c, l, .5}
			tick.Predictions[a] = Prediction{P: mixes[a].Forecast(ex), Experts: ex}
		}
		tick.Outcome = truth(x, t, s, rng)
		if !tick.Missing {
			queue = append(queue, packet{Origin: t, Due: t + s.Delay, X: values, Y: tick.Outcome, Audit: tick.Audit, Predictions: tick.Predictions})
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
			pendingAudits = append(pendingAudits, v)
			start := time.Now()
			trees[0].Update(v.Bits, v.Outcome)
			r.TreeNS[0] += int64(time.Since(start))
			r.TreeUpdates[0]++
			if r.Audits < 32 || r.Audits%16 != 0 {
				continue
			}
			start = time.Now()
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
			for _, v := range pendingAudits {
				trees[1].Update(v.Bits, v.Outcome)
				r.TreeUpdates[1]++
			}
			pendingAudits = nil
			r.TreeNS[1] += int64(time.Since(start))
			start = time.Now()
			trees[2] = NewForest(rollingSeed)
			for _, v := range audits[max(0, len(audits)-64):] {
				trees[2].Update(v.Bits, v.Outcome)
				r.TreeUpdates[2]++
			}
			r.TreeNS[2] += int64(time.Since(start))
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
	for a, f := range trees {
		r.TreeNodes[a] = f.Nodes()
	}
	return r, nil
}

func RunCadence() (CadenceOutput, error) {
	o := CadenceOutput{}
	for k, split := range []string{"design", "confirmation"} {
		for j, s := range Scenarios {
			for fit := 0; fit < 3; fit++ {
				rng := rand.New(rand.NewSource(2026092601*1000000 + int64(j*1000+fit)))
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
					r, e := RunCadenceStream(base, split, j, fit, stream, int64(2026092602+k))
					if e != nil {
						return o, e
					}
					o.Records = append(o.Records, r)
				}
			}
		}
	}
	SummarizeCadence(&o)
	return o, nil
}

func SummarizeCadence(o *CadenceOutput) {
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
				x.Comparisons[i].Arm = CadenceArms[j-3]
				break
			}
		}
	}
	for i := range x.Verdicts {
		x.Verdicts[i].Arm = CadenceArms[i+1]
	}
	o.Comparisons, o.Verdicts = x.Comparisons, x.Verdicts
}
