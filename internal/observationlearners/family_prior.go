package observationlearners

import (
	"errors"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func fitPriorConditional(samples []observation.Sample, parityPrior float64) (*ConditionalForest, [2]float64, error) {
	if math.IsNaN(parityPrior) || math.IsInf(parityPrior, 0) || parityPrior < 0 || parityPrior > 1 {
		return nil, [2]float64{}, errors.New("invalid parity family prior")
	}
	g, err := fitSubset(samples)
	if err != nil {
		return nil, [2]float64{}, err
	}
	p, err := fitBooleanSpecialist(samples)
	if err != nil {
		return nil, [2]float64{}, err
	}
	logs := [2]float64{math.Log1p(-parityPrior) + familyLogEvidence(g.evidence), math.Log(parityPrior) + familyLogEvidence(p.evidence)}
	maximum := math.Max(logs[0], logs[1])
	w := [2]float64{math.Exp(logs[0] - maximum), math.Exp(logs[1] - maximum)}
	total := w[0] + w[1]
	w[0], w[1] = w[0]/total, w[1]/total
	m := new(ConditionalForest)
	for x := range g.predictions {
		prediction := w[0]*g.predictions[x] + w[1]*p.predictions[x]
		m.cells[partialIndex(511, uint16(x))] = conditionalCell{1, prediction}
	}
	// The input law is identical in both families, so marginalizing this mixture
	// is coherent without changing family weights after a partial observation.
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
	return m, w, nil
}

// NewPriorConditional freezes a declared family prior and its coherent posterior.
// Endpoints retain one family; interior priors do not cap posterior influence.
// This is a finite research posterior predictive over generic
// and parity models. Callers must supply only already eligible evidence; family
// probabilities do not authorize causal claims, sharing or residual reuse.
func NewPriorConditional(samples []observation.Sample, parityPrior float64) (*ConditionalForest, error) {
	m, _, err := fitPriorConditional(samples, parityPrior)
	return m, err
}
