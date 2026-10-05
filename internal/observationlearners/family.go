package observationlearners

import (
	"math"
	"math/bits"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// familyLogEvidence integrates a family's alternative subset hypotheses under
// its normalized inclusion prior. It does not treat them as independent data.
func familyLogEvidence(evidence [512]float64) float64 {
	var terms [512]float64
	maximum := math.Inf(-1)
	for mask, e := range evidence {
		k := float64(bits.OnesCount(uint(mask)))
		terms[mask] = e + k*math.Log(1./3) + (9-k)*math.Log(2./3)
		maximum = math.Max(maximum, terms[mask])
	}
	total := 0.
	for _, term := range terms {
		total += math.Exp(term - maximum)
	}
	return maximum + math.Log(total)
}

func fitFamilyConditional(samples []observation.Sample) (*ConditionalForest, [2]float64, error) {
	g, err := fitSubset(samples)
	if err != nil {
		return nil, [2]float64{}, err
	}
	p, err := fitBooleanSpecialist(samples)
	if err != nil {
		return nil, [2]float64{}, err
	}
	logs := [2]float64{familyLogEvidence(g.evidence), familyLogEvidence(p.evidence)}
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

// NewFamilyConditional is a finite research posterior predictive over generic
// and parity models. Callers must supply only already eligible evidence; family
// probabilities do not authorize causal claims, sharing or residual reuse.
func NewFamilyConditional(samples []observation.Sample) (*ConditionalForest, error) {
	m, _, err := fitFamilyConditional(samples)
	return m, err
}
