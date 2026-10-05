package researchcalibration

import (
	"errors"
	"math"
)

// TrainingLOO predicts the first label of an observed member after removing
// its likelihood. It does not return the repeated-member future prediction.
func (m *Model) TrainingLOO(i int) (float64, error) {
	if i < 0 || i >= len(m.base) || !m.seen[i] {
		return 0, errors.New("leave-one-out requires an observed member")
	}
	row := m.p[i*Hypotheses : (i+1)*Hypotheses]
	mass, numerator := 0., 0.
	for h, w := range m.weights {
		likelihood := row[h]
		if !m.useful[i] {
			likelihood = 1 - likelihood
		}
		if likelihood <= 0 || math.IsNaN(likelihood) || math.IsInf(likelihood, 0) {
			return 0, errors.New("invalid held-out likelihood")
		}
		v := w / likelihood
		mass += v
		numerator += v * row[h]
	}
	if mass <= 0 || math.IsNaN(mass) || math.IsInf(mass, 0) || math.IsNaN(numerator) || math.IsInf(numerator, 0) {
		return 0, errors.New("invalid leave-one-out normalization")
	}
	return numerator / mass, nil
}
