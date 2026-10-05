package observationgate

import (
	"errors"
	"math"
)

// predictiveBet freezes the query/model/rate before the revealing observation.
// Its model chooses a bet; it does not certify the model or replace provenance.
type predictiveBet struct {
	q, m, c [4]float64
	eta     float64
	rate    [2]float64
}

func preparePredictiveBet(p residualAllocation) (out predictiveBet, err error) {
	out.q, out.m = p.proposal()
	out.eta, out.c, err = augmentation(out.q, out.m)
	if err != nil {
		return
	}
	var probabilities [4][3]float64
	for i := range probabilities {
		probabilities[i] = [3]float64{.5, 1, .5}
		for j := 0; j < p.count[i]; j++ {
			d := p.values[i][j]
			if d != -1 && d != 0 && d != 1 {
				return out, errors.New("predictive bet requires ternary evidence")
			}
			probabilities[i][int(d)+1]++
		}
		for j := range probabilities[i] {
			probabilities[i][j] /= float64(p.count[i] + 2)
		}
	}
	for signIndex, sign := range []float64{1, -1} {
		var xs, ps [12]float64
		lower := math.Inf(1)
		for i := range out.q {
			for j := 0; j < 3; j++ {
				k := i*3 + j
				xs[k] = sign*(.25*float64(j-1)/out.q[i]+out.eta*out.c[i]) - .15
				ps[k] = out.q[i] * probabilities[i][j]
				lower = math.Min(lower, xs[k])
			}
		}
		cap := .8
		if lower < 0 {
			cap = math.Min(cap, .92/(-lower))
		}
		derivative := func(rate float64) float64 {
			v := 0.
			for k, x := range xs {
				v += ps[k] * x / (1 + rate*x)
			}
			return v
		}
		if derivative(0) <= 0 {
			out.rate[signIndex] = 0
			continue
		}
		if derivative(cap) >= 0 {
			out.rate[signIndex] = cap
			continue
		}
		lo, hi := 0., cap
		for j := 0; j < 24; j++ {
			mid := (lo + hi) / 2
			if derivative(mid) > 0 {
				lo = mid
			} else {
				hi = mid
			}
		}
		out.rate[signIndex] = (lo + hi) / 2
	}
	return
}

type predictiveGate struct{ base importanceGate }

func (g *predictiveGate) observe(d float64, channel int, snapshot predictiveBet) (bool, error) {
	if channel < 0 || channel >= 4 || (d != -1 && d != 0 && d != 1) {
		return g.base.alert, errors.New("invalid predictive evidence")
	}
	// The snapshot is private and constructed by preparePredictiveBet. It must
	// be retained with the query, never rebuilt from later feedback history.
	eta, c, err := augmentation(snapshot.q, snapshot.m)
	if err != nil || eta != snapshot.eta || c != snapshot.c {
		return g.base.alert, errors.New("invalid predictive snapshot")
	}
	for k, s := range []float64{1, -1} {
		rate := snapshot.rate[k]
		if math.IsNaN(rate) || math.IsInf(rate, 0) || rate < 0 || rate > .8 {
			return g.base.alert, errors.New("invalid predictable rate")
		}
		for i := range c {
			for _, endpoint := range []float64{-1, 1} {
				x := s*(.25*endpoint/snapshot.q[i]+eta*c[i]) - .15
				if 1+rate*x < .08-1e-12 {
					return g.base.alert, errors.New("unsafe predictable rate")
				}
			}
		}
	}
	z := .25*d/snapshot.q[channel] + snapshot.eta*snapshot.c[channel]
	var starts [8]float64
	for j := range starts {
		if g.base.next < j*64 {
			continue
		}
		for k, s := range []float64{1, -1} {
			g.base.logs[j][k] += math.Log1p(snapshot.rate[k] * (s*z - .15))
		}
		starts[j] = logMean(g.base.logs[j][:])
	}
	g.base.alert = g.base.alert || logMean(starts[:]) >= math.Log(100)
	g.base.next++
	return g.base.alert, nil
}
