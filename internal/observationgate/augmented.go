package observationgate

import (
	"errors"
	"math"
)

// residualAllocation uses only past signed observations to estimate conditional
// means and variances. It cannot see the simulator's channel law or future tape.
type residualAllocation struct {
	values      [4][32]float64
	count, next [4]int
}

func (p *residualAllocation) proposal() (q, m [4]float64) {
	total := 0.
	for i := range q {
		sum, squares := 0., 0.
		for j := 0; j < p.count[i]; j++ {
			d := p.values[i][j]
			sum += d
			squares += d * d
		}
		n := float64(p.count[i] + 2)
		m[i] = sum / n
		q[i] = math.Sqrt((1+squares)/n - m[i]*m[i])
		total += q[i]
	}
	for i := range q {
		q[i] = .1 + .6*q[i]/total
	}
	return
}

func (p *residualAllocation) observe(c int, d float64) error {
	if c < 0 || c >= 4 || math.IsNaN(d) || math.IsInf(d, 0) || math.Abs(d) > 1 {
		return errors.New("invalid residual allocation evidence")
	}
	p.values[c][p.next[c]] = d
	p.next[c] = (p.next[c] + 1) % 32
	if p.count[c] < 32 {
		p.count[c]++
	}
	return nil
}

// augmentation limits a predictable coefficient, not the revealing outcome.
// Correct propensities make the control term mean-zero even for a wrong model.
func augmentation(q, m [4]float64) (eta float64, c [4]float64, err error) {
	total, average := 0., 0.
	for i, v := range q {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < .1 || v > .7 || math.IsNaN(m[i]) || math.IsInf(m[i], 0) || math.Abs(m[i]) > 1 {
			return 0, c, errors.New("invalid augmentation contract")
		}
		total += v
		average += m[i] / 4
	}
	if math.Abs(total-1) > 1e-12 {
		return 0, c, errors.New("unnormalized proposal")
	}
	eta = 1
	for i := range c {
		w := .25 / q[i]
		c[i] = average - w*m[i]
		if c[i] != 0 {
			eta = math.Min(eta, (3.53-w)/math.Abs(c[i]))
		}
	}
	return
}

type augmentedGate struct{ base importanceGate }

func (g *augmentedGate) observe(d float64, channel int, q, m [4]float64) (bool, error) {
	if channel < 0 || channel >= 4 || math.IsNaN(d) || math.IsInf(d, 0) || math.Abs(d) > 1 {
		return g.base.alert, errors.New("invalid augmented outcome")
	}
	eta, c, err := augmentation(q, m)
	if err != nil {
		return g.base.alert, err
	}
	z := .25*d/q[channel] + eta*c[channel]
	var starts [8]float64
	for j := range starts {
		if g.base.next < j*64 {
			continue
		}
		for k, s := range []float64{1, -1} {
			g.base.logs[j][k] += math.Log1p(.25 * (s*z - .15))
		}
		starts[j] = logMean(g.base.logs[j][:])
	}
	g.base.alert = g.base.alert || logMean(starts[:]) >= math.Log(100)
	g.base.next++
	return g.base.alert, nil
}
