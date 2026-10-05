package observationgate

import "errors"

// pairedInvestigator requests scrutiny of one reliability pair, not permission
// to share. Four identical rate-history slots represent a single scalar law.
// They never multiply support: one feedback pair advances evidence once.
type pairedInvestigator struct {
	history  residualAllocation
	evidence rateMixture
	next     int
}

func (g *pairedInvestigator) observe(sequence int, reference, live bool) (bool, error) {
	if sequence != g.next {
		return g.evidence.alert, errors.New("out-of-order reliability pair")
	}
	for c := 1; c < 4; c++ {
		if g.history.values[c] != g.history.values[0] || g.history.count[c] != g.history.count[0] || g.history.next[c] != g.history.next[0] {
			return g.evidence.alert, errors.New("non-scalar reliability history")
		}
	}
	long, e := preparePredictiveBet(g.history)
	if e != nil {
		return g.evidence.alert, e
	}
	short, e := prepareWindowBet(g.history, 8)
	if e != nil {
		return g.evidence.alert, e
	}
	// Materialize the exact scalar law instead of testing floating-point
	// normalization for equality. The correction is identically zero; m=0
	// expresses that without a rounded mean subtraction. Rates remain based
	// only on the past scalar observations, not the revealing pair below.
	for _, s := range []*predictiveBet{&long, &short} {
		s.q = [4]float64{.25, .25, .25, .25}
		s.m = [4]float64{}
		s.c = [4]float64{}
		s.eta = 1
	}
	d := 0.
	if reference {
		d++
	}
	if live {
		d--
	}
	next := *g
	request, e := next.evidence.observe(d, 0, long, short)
	if e != nil {
		return g.evidence.alert, e
	}
	for c := 0; c < 4; c++ {
		if e := next.history.observe(c, d); e != nil {
			return g.evidence.alert, e
		}
	}
	next.next++
	*g = next
	return request, nil
}
