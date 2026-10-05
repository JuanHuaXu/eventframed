package observationgate

import (
	"errors"
	"math"
)

// allocationPolicy requests evidence, not authority to change a group. Its
// proposal is computed entirely from previously observed squared differences.
type allocationPolicy struct {
	values      [4][32]float64
	count, next [4]int
}

func (p *allocationPolicy) proposal() [4]float64 {
	var q [4]float64
	total := 0.
	for i := range q {
		sum := 0.
		for j := 0; j < p.count[i]; j++ {
			sum += p.values[i][j]
		}
		q[i] = math.Sqrt((1 + sum) / float64(2+p.count[i]))
		total += q[i]
	}
	for i := range q {
		q[i] = .1 + .6*q[i]/total
	}
	return q
}

func (p *allocationPolicy) observe(channel int, d float64) error {
	if channel < 0 || channel >= 4 || math.IsNaN(d) || math.IsInf(d, 0) || math.Abs(d) > 1 {
		return errors.New("invalid allocation observation")
	}
	p.values[channel][p.next[channel]] = d * d
	p.next[channel] = (p.next[channel] + 1) % 32
	if p.count[channel] < 32 {
		p.count[channel]++
	}
	return nil
}

// importanceGate is research-only and requires the actual randomized propensity,
// not a fitted relevance score. The fixed target mass is one quarter per channel.
// At the declared floor, weighting preserves the null and positive bet factors.
type importanceGate struct {
	logs  [8][2]float64
	next  int
	alert bool
}

func (g *importanceGate) observe(d, q float64) (bool, error) {
	if math.IsNaN(d) || math.IsInf(d, 0) || math.Abs(d) > 1 || math.IsNaN(q) || math.IsInf(q, 0) || q < .1 || q > .7 {
		return g.alert, errors.New("invalid importance evidence")
	}
	z := .25 * d / q
	var starts [8]float64
	for j := range starts {
		if g.next < j*64 {
			continue
		}
		for k, s := range []float64{1, -1} {
			g.logs[j][k] += math.Log1p(.25 * (s*z - .15))
		}
		starts[j] = logMean(g.logs[j][:])
	}
	g.alert = g.alert || logMean(starts[:]) >= math.Log(100)
	g.next++
	return g.alert, nil
}
