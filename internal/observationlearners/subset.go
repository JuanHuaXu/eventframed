package observationlearners

import (
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"math/bits"
)

type subsetCell struct{ n, yes int }
type subsetModel struct{ predictions, weights, evidence [512]float64 }

// fitSubset averages conditional-label models. Covariates are conditioned on;
// likelihoods are for the labeled sequence, so no binomial coefficient belongs.
func fitSubset(samples []observation.Sample) (*subsetModel, error) {
	if len(samples) == 0 || len(samples) > 256 {
		return nil, errors.New("invalid subset sample count")
	}
	m := new(subsetModel)
	var counts [19683]subsetCell
	for _, sample := range samples {
		if sample.Bits >= 512 {
			return nil, errors.New("invalid subset input")
		}
		for mask := uint16(0); mask < 512; mask++ {
			c := &counts[partialIndex(mask, sample.Bits&mask)]
			p := (float64(c.yes) + .5) / (float64(c.n) + 1)
			if !sample.Outcome {
				p = 1 - p
			}
			m.evidence[mask] += math.Log(p)
			c.n++
			if sample.Outcome {
				c.yes++
			}
		}
	}
	maximum := math.Inf(-1)
	for mask := range m.weights {
		k := float64(bits.OnesCount(uint(mask)))
		m.weights[mask] = k*math.Log(1./3) + (9-k)*math.Log(2./3) + m.evidence[mask]
		maximum = math.Max(maximum, m.weights[mask])
	}
	total := 0.
	for i, w := range m.weights {
		m.weights[i] = math.Exp(w - maximum)
		total += m.weights[i]
	}
	for i := range m.weights {
		m.weights[i] /= total
	}
	for x := uint16(0); x < 512; x++ {
		for mask, w := range m.weights {
			c := counts[partialIndex(uint16(mask), x&uint16(mask))]
			m.predictions[x] += w * (float64(c.yes) + .5) / (float64(c.n) + 1)
		}
	}
	return m, nil
}

// NewSubsetConditional freezes a conditional outcome model and a separate
// declared input estimate. It uses no current query or future outcome.
func NewSubsetConditional(samples []observation.Sample, weights [512]float64) (*ConditionalForest, error) {
	fitted, e := fitSubset(samples)
	if e != nil {
		return nil, e
	}
	m := new(ConditionalForest)
	for x, w := range weights {
		if math.IsNaN(w) || math.IsInf(w, 0) || w <= 0 || w > 1e6 {
			return nil, errors.New("invalid subset input weight")
		}
		m.cells[partialIndex(511, uint16(x))] = conditionalCell{w, w * fitted.predictions[x]}
	}
	for i := len(m.cells) - 1; i >= 0; i-- {
		v, p := i, 1
		for bit := 0; bit < 9; bit++ {
			if v%3 == 0 {
				a, b := m.cells[i+p], m.cells[i+2*p]
				m.cells[i] = conditionalCell{a.mass + b.mass, a.weighted + b.weighted}
				break
			}
			v /= 3
			p *= 3
		}
	}
	return m, nil
}
