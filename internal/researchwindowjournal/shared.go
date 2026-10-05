package researchwindowjournal

import "errors"

// Immutable tables are package-private and built from a COPY of caller input.
// Bank siblings also share the raw observation journal. Only the bank may
// publish journal transitions; each sibling retains independent statistics.
func (m *Model) newSibling(window int, epoch uint64, cap int) (*Model, error) {
	if window < 1 || window > len(m.base)*MaxTrials || epoch == 0 || cap < 1 || cap > 2*len(m.base)*MaxTrials {
		return nil, errors.New("bank sibling contract")
	}
	next := &Model{bankOwned: true, base: m.base, cfg: Config{Depth: m.cfg.Depth, Window: window}, epoch: epoch, cap: cap,
		issued: make([]int, len(m.base)), ranks: m.ranks, trials: m.trials, seqSlots: m.seqSlots,
		rates: m.rates, factors: m.factors, individuals: make([]individual, len(m.base)), localRates: m.localRates, localPriors: m.localPriors,
		localFactors: m.localFactors, leafMembers: m.leafMembers}
	for i := range next.base {
		if e := next.refreshIndividual(&next.individuals[i], i); e != nil {
			return nil, e
		}
	}
	for n := (1 << (next.cfg.Depth + 1)) - 2; n >= 0; n-- {
		if e := next.refresh(&next.nodes[n], n, nil); e != nil {
			return nil, e
		}
	}
	var e error
	next.weights, e = next.normalize(nil)
	return next, e
}

// LatentJoint derives Y,W1,W2 from the ACTUAL current tree/noise posterior.
// It is a next-trial kernel, not a joint for an already-issued old trial.
func (m *Model) LatentJoint(i int) ([8]float64, error) {
	var out [8]float64
	if i < 0 || i >= len(m.base) {
		return out, errors.New("bank joint member")
	}
	for h, w := range m.weights {
		if w == 0 {
			continue
		}
		n := (1 << m.cfg.Depth) - 1 + m.ranks[i]
		x := &m.nodes[n]
		mean := 0.
		for z, p := range x.posterior[h] {
			mean += p * m.rates[i][z]
		}
		mean = x.stop[h]*mean + (1-x.stop[h])*m.individuals[i].mean[h]
		for n > 0 {
			n = (n - 1) / 2
			x = &m.nodes[n]
			v := 0.
			for z, p := range x.posterior[h] {
				v += p * m.rates[i][z]
			}
			mean = x.stop[h]*v + (1-x.stop[h])*mean
		}
		if !finite(mean) || mean < 0 || mean > 1+2e-14 {
			return [8]float64{}, errors.New("bank joint branch mean")
		}
		// Tiny arithmetic endpoint excess is bounded by the same branch tolerance.
		if mean > 1 {
			mean = 1
		}
		for y := 0; y < 2; y++ {
			for a := 0; a < 2; a++ {
				for b := 0; b < 2; b++ {
					mass := w * mean
					if y == 0 {
						mass = w * (1 - mean)
					}
					one, two := noise[h], noise[h]
					if a == y {
						one = 1 - noise[h]
					}
					if b == y {
						two = 1 - noise[h]
					}
					out[4*y+2*a+b] += mass * one * two
				}
			}
		}
	}
	return out, nil
}
