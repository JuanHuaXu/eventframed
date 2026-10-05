package researchwindowjournal

import (
	"errors"
	"math"
)

func (m *Model) initializeIndividual(i int, b float64) error {
	weights := [21]float64{}
	weights[0] = 1
	for j := 0; j < 20; j++ {
		weights[0] *= (1 - b + float64(j)) / (1 + float64(j))
	}
	for z := 0; z < 20; z++ {
		weights[z+1] = weights[z] * float64(20-z) / float64(z+1) * (b + float64(z)) / (1 - b + float64(19-z))
	}
	sum := 0.
	for _, w := range weights {
		sum += w
	}
	if !finite(sum) || sum <= 0 {
		return errors.New("local prior normalization")
	}
	m.localRates[i][0], m.localPriors[i][0] = b, .8
	mean := .8 * b
	for z, w := range weights {
		m.localRates[i][z+1] = float64(z) / 20
		m.localPriors[i][z+1] = .2 * w / sum
		if !finite(m.localPriors[i][z+1]) || m.localPriors[i][z+1] <= 0 {
			return errors.New("local prior support")
		}
		mean += m.localRates[i][z+1] * m.localPriors[i][z+1]
	}
	if math.Abs(mean-b) > 2e-14 {
		return errors.New("local prior mean")
	}
	for h, eta := range noise {
		for z, p := range m.localRates[i] {
			q := eta + (1-2*eta)*p
			m.localFactors[i][h][z][0], m.localFactors[i][h][z][1] = math.Log1p(-q), math.Log(q)
			for a := 0; a < 2; a++ {
				for b := 0; b < 2; b++ {
					one, two := eta, eta
					if a == 1 {
						one = 1 - eta
					}
					if b == 1 {
						two = 1 - eta
					}
					m.localFactors[i][h][z][2+2*a+b] = math.Log(p*one*two + (1-p)*(1-one)*(1-two))
				}
			}
		}
	}
	return m.refreshIndividual(&m.individuals[i], i)
}

func (m *Model) refreshIndividual(x *individual, i int) error {
	total := 0
	for _, c := range x.counts {
		if c < 0 {
			return errors.New("local member count")
		}
		total += c
	}
	if total > MaxTrials || total > m.cfg.Window {
		return errors.New("local member window")
	}
	for h := 0; h < 3; h++ {
		var terms [LocalAtoms]float64
		maximum := math.Inf(-1)
		for z := range terms {
			if !finite(x.finite[h][z]) || !finite(x.compensation[h][z]) || x.zeros[h][z] < 0 || x.zeros[h][z] > total {
				return errors.New("local member finite/support fault")
			}
			terms[z] = math.Log(m.localPriors[i][z]) + x.finite[h][z]
			if x.zeros[h][z] > 0 {
				terms[z] = math.Inf(-1)
			}
			maximum = math.Max(maximum, terms[z])
			x.posterior[h][z] = m.localPriors[i][z]
		}
		x.evidence[h], x.mean[h] = math.Inf(-1), m.base[i]
		if !finite(maximum) {
			continue
		} // Zero-mass noise branches are not normalized.
		sum := 0.
		for z, v := range terms {
			x.posterior[h][z] = math.Exp(v - maximum)
			sum += x.posterior[h][z]
		}
		if !finite(sum) || sum <= 0 {
			return errors.New("local member normalization")
		}
		x.mean[h] = 0
		for z := range terms {
			x.posterior[h][z] /= sum
			x.mean[h] += x.posterior[h][z] * m.localRates[i][z]
		}
		x.evidence[h] = maximum + math.Log(sum)
		if total == 0 {
			x.evidence[h] = 0
		} // Empty products are exactly one.
		if !finite(x.mean[h]) || x.mean[h] < 0 || x.mean[h] > 1+2e-14 {
			return errors.New("local member mean")
		}
	}
	return nil
}

func addIndividual(x *individual, h, z int, v float64, sign int) error {
	if math.IsInf(v, -1) {
		x.zeros[h][z] += sign
		return nil
	}
	if !finite(v) {
		return errors.New("local member factor")
	}
	y := float64(sign)*v - x.compensation[h][z]
	next := x.finite[h][z] + y
	x.compensation[h][z] = (next - x.finite[h][z]) - y
	x.finite[h][z] = next
	return nil
}
func (m *Model) prepareIndividual(x *individual, i, old, next int) error {
	if old >= 0 {
		x.counts[old]--
	}
	if next >= 0 {
		x.counts[next]++
	}
	for h := 0; h < 3; h++ {
		for z := 0; z < LocalAtoms; z++ {
			if old >= 0 {
				if e := addIndividual(x, h, z, m.localFactors[i][h][z][old], -1); e != nil {
					return e
				}
			}
			if next >= 0 {
				if e := addIndividual(x, h, z, m.localFactors[i][h][z][next], 1); e != nil {
					return e
				}
			}
		}
	}
	return m.refreshIndividual(x, i)
}

// Product evidence can lose and regain noise support as individual factors
// change or expire. Keep integer zero counts separate from removable finite logs.
func addProduct(x *node, h int, v float64, sign int) error {
	if math.IsInf(v, -1) {
		x.localZeros[h] += sign
		return nil
	}
	if !finite(v) {
		return errors.New("local product evidence")
	}
	y := float64(sign)*v - x.localComp[h]
	next := x.localFinite[h] + y
	x.localComp[h] = (next - x.localFinite[h]) - y
	x.localFinite[h] = next
	return nil
}
