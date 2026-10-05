package observationgate

import (
	"errors"
	"math"
)

// prepareWindowBet changes rate estimation only. Sampling and augmentation
// retain their longer history, so the observation correction uses the actual q.
// This isolated research implementation preserves the archived v76 source.
func prepareWindowBet(p residualAllocation, window int) (out predictiveBet, err error) {
	if window < 1 || window > 32 {
		return out, errors.New("invalid rate window")
	}
	out.q, out.m = p.proposal()
	out.eta, out.c, err = augmentation(out.q, out.m)
	if err != nil {
		return
	}
	var probabilities [4][3]float64
	for c := range probabilities {
		probabilities[c] = [3]float64{.5, 1, .5}
		n := min(p.count[c], window)
		for j := 0; j < n; j++ {
			// next points to the oldest slot only when full; indexing backward
			// selects the same recent observations before and after wrapping.
			d := p.values[c][(p.next[c]-1-j+32)%32]
			if d != -1 && d != 0 && d != 1 {
				return out, errors.New("rate model requires ternary evidence")
			}
			probabilities[c][int(d)+1]++
		}
		for j := range probabilities[c] {
			probabilities[c][j] /= float64(n + 2)
		}
	}
	for k, sign := range []float64{1, -1} {
		var xs, ps [12]float64
		lower := math.Inf(1)
		for c := range probabilities {
			for j := 0; j < 3; j++ {
				i := c*3 + j
				xs[i] = sign*(.25*float64(j-1)/out.q[c]+out.eta*out.c[c]) - .15
				ps[i] = out.q[c] * probabilities[c][j]
				lower = math.Min(lower, xs[i])
			}
		}
		cap := .8
		if lower < 0 {
			cap = math.Min(cap, .92/-lower)
		}
		derivative := func(rate float64) float64 {
			v := 0.
			for i, x := range xs {
				v += ps[i] * x / (1 + rate*x)
			}
			return v
		}
		if derivative(0) <= 0 {
			out.rate[k] = 0
			continue
		}
		if derivative(cap) >= 0 {
			out.rate[k] = cap
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
		out.rate[k] = (lo + hi) / 2
	}
	return
}
