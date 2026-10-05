package observationlearners

import (
	"errors"
	"math"
	"math/bits"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type booleanSpecialist struct {
	weights, agreement, evidence, predictions [512]float64
}

// fitBooleanSpecialist is a bounded research model, not a general parity solver.
// Each subset has one Beta-distributed agreement rate, unlike the separate
// conditional-cell rates in fitSubset. Both evidence and prediction use it.
func fitBooleanSpecialist(samples []observation.Sample) (*booleanSpecialist, error) {
	if len(samples) == 0 || len(samples) > 256 {
		return nil, errors.New("invalid Boolean specialist sample count")
	}
	for _, sample := range samples {
		if sample.Bits >= 512 {
			return nil, errors.New("invalid Boolean specialist input")
		}
	}
	m := new(booleanSpecialist)
	maximum := math.Inf(-1)
	for mask := range m.weights {
		count := 0
		for n, sample := range samples {
			agrees := (bits.OnesCount16(sample.Bits&uint16(mask))%2 == 1) == sample.Outcome
			p := (float64(count) + .5) / (float64(n) + 1)
			if !agrees {
				p = 1 - p
			}
			m.evidence[mask] += math.Log(p)
			if agrees {
				count++
			}
		}
		m.agreement[mask] = (float64(count) + .5) / (float64(len(samples)) + 1)
		k := float64(bits.OnesCount(uint(mask)))
		m.weights[mask] = k*math.Log(1./3) + (9-k)*math.Log(2./3) + m.evidence[mask]
		maximum = math.Max(maximum, m.weights[mask])
	}
	total := 0.
	for mask, w := range m.weights {
		m.weights[mask] = math.Exp(w - maximum)
		total += m.weights[mask]
	}
	for mask := range m.weights {
		m.weights[mask] /= total
	}
	for x := range m.predictions {
		for mask, w := range m.weights {
			p := m.agreement[mask]
			if bits.OnesCount(uint(x&mask))%2 == 0 {
				p = 1 - p
			}
			m.predictions[x] += w * p
		}
	}
	return m, nil
}

// NewBooleanConditional freezes a noisy-rule posterior predictive under uniform
// independent input bits. Supplied samples must already be eligible evidence;
// this snapshot builder neither retrieves labels nor authorizes their use.
func NewBooleanConditional(samples []observation.Sample) (*ConditionalForest, error) {
	fitted, err := fitBooleanSpecialist(samples)
	if err != nil {
		return nil, err
	}
	m := new(ConditionalForest)
	for x, p := range fitted.predictions {
		m.cells[partialIndex(511, uint16(x))] = conditionalCell{1, p}
	}
	// Keep frozen predecessor constructors unchanged for archived replay. This
	// descending ternary pass marginalizes the same complete joint law.
	for i := len(m.cells) - 1; i >= 0; i-- {
		v, place := i, 1
		for bit := 0; bit < 9; bit++ {
			if v%3 == 0 {
				a, b := m.cells[i+place], m.cells[i+2*place]
				m.cells[i] = conditionalCell{a.mass + b.mass, a.weighted + b.weighted}
				break
			}
			v /= 3
			place *= 3
		}
	}
	return m, nil
}
