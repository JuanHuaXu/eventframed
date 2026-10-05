package researchnoisemoment

import (
	"errors"
	"math"
)

// LogEvidence is the full observed-label marginal under ONE shared family.
// Private families need a product of member marginals, a different model.
func (m *Model) LogEvidence() (float64, error) {
	if !m.cfg.Shared {
		return 0, errors.New("shared evidence requires shared family")
	}
	prior := familyWeights(m.h)
	maximum := math.Inf(-1)
	var logs [MaxFamilies]float64
	for h := 0; h < m.h; h++ {
		logs[h] = math.Log(prior[h]) + m.global[h]
		maximum = math.Max(maximum, logs[h])
	}
	sum := 0.
	for h := 0; h < m.h; h++ {
		sum += math.Exp(logs[h] - maximum)
	}
	q := maximum + math.Log(sum)
	if !finite(q) {
		return 0, errors.New("nonfinite marginal evidence")
	}
	return q, nil
}

func (m *Model) ObservedPredict(i int) (float64, error) {
	q, e := m.Predict(i)
	return m.cfg.Noise + (1-2*m.cfg.Noise)*q, e
}
