package researchpartition

import (
	"errors"
	"math"
)

// TrainingLOO integrates the held-out member as unobserved, removing BOTH its
// leaf count and its effect on partition mass. A post-label mean is not LOO.
func (m *Model) TrainingLOO(i int) (float64, error) {
	if i < 0 || i >= len(m.base) || !m.seen[i] {
		return 0, errors.New("leave-one-out requires an observed member")
	}
	mass, numerator := 0., 0.
	for k, w := range m.weights {
		q, likelihood := m.base[i], m.base[i]
		if !m.useful[i] {
			likelihood = 1 - likelihood
		}
		if k > 0 {
			node := m.partitions[k-1].node[m.bin[i]]
			s, f := float64(m.successes[node]), float64(m.failures[node])
			denominator := 1 + s + f
			if m.useful[i] {
				q, likelihood = s/denominator, s/denominator
			} else {
				q, likelihood = (1+s)/denominator, f/denominator
			}
		}
		if likelihood <= 0 || math.IsNaN(likelihood) || math.IsInf(likelihood, 0) {
			return 0, errors.New("invalid held-out likelihood")
		}
		v := w / likelihood
		mass += v
		numerator += v * q
	}
	if mass <= 0 || math.IsNaN(mass) || math.IsInf(mass, 0) || math.IsNaN(numerator) || math.IsInf(numerator, 0) {
		return 0, errors.New("invalid leave-one-out normalization")
	}
	return numerator / mass, nil
}
