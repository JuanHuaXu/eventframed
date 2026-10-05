package observationgate

import (
	"errors"
	"math"
)

// rateMixture averages correlated evidence wealth, never alarm bits or products.
// Fixed weights reserve half the initial evidence for the fixed-rate component.
type rateMixture struct {
	components [3]predictiveGate
	alert      bool
}

func componentLogWealth(g predictiveGate) float64 {
	var starts [8]float64
	for j := range starts {
		starts[j] = logMean(g.base.logs[j][:])
	}
	return logMean(starts[:])
}

func (g rateMixture) logWealth() float64 {
	var weighted [3]float64
	for i, w := range []float64{.5, .25, .25} {
		weighted[i] = math.Log(w) + componentLogWealth(g.components[i])
	}
	// logMean subtracts log(3); restore it to obtain a weighted sum.
	return logMean(weighted[:]) + math.Log(3)
}

func (g *rateMixture) observe(d float64, channel int, long, short predictiveBet) (bool, error) {
	if long.q != short.q || long.m != short.m || long.c != short.c || long.eta != short.eta {
		return g.alert, errors.New("mixture snapshot dependency mismatch")
	}
	fixed := long
	fixed.rate = [2]float64{.25, .25}
	// All components commit together, including rejection of an invalid last
	// component. This bounded value copy prevents a partially consumed outcome.
	next := *g
	for i, s := range [3]predictiveBet{fixed, long, short} {
		if _, e := next.components[i].observe(d, channel, s); e != nil {
			return g.alert, e
		}
	}
	next.alert = next.alert || next.logWealth() >= math.Log(100)
	*g = next
	return g.alert, nil
}
