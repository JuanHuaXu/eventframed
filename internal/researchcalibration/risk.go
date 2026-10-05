package researchcalibration

import (
	"errors"
	"math"
)

// RiskScores gives one-step expected reduction in weighted future Bernoulli
// Brier risk under this model, not an external-law or packet-utility guarantee.
// Each target is a future label, including another label for the queried event.
// Seen events remain targets but cannot supply another training observation.
// The covariance Gram shortcut costs O(n*M*M) time and O(n*M+M*M) scratch.
func (m *Model) RiskScores(priority []float64) ([]float64, error) {
	n := len(m.base)
	if len(priority) != n {
		return nil, errors.New("priority frontier mismatch")
	}
	normalizer := 0.0
	for _, v := range priority {
		if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, errors.New("invalid target priority")
		}
		normalizer += v
	}
	if normalizer <= 0 || math.IsInf(normalizer, 0) {
		return nil, errors.New("invalid priority mass")
	}
	mean := make([]float64, n)
	centered := make([]float64, n*Hypotheses)
	var gram [Hypotheses * Hypotheses]float64
	for j := 0; j < n; j++ {
		q, err := m.Predict(j)
		if err != nil {
			return nil, err
		}
		mean[j] = q
		row := centered[j*Hypotheses : (j+1)*Hypotheses]
		for h := range row {
			p := m.p[j*Hypotheses+h]
			if m.seen[j] {
				y := 0.0
				if m.useful[j] {
					y = 1
				}
				p = (2*p + y) / 3
			}
			row[h] = p - q
		}
		v := priority[j] / normalizer
		for h, dh := range row {
			for k, dk := range row {
				gram[h*Hypotheses+k] += v * dh * dk
			}
		}
	}
	scores := make([]float64, n)
	for e := 0; e < n; e++ {
		if m.seen[e] {
			scores[e] = -1
			continue
		}
		row := centered[e*Hypotheses : (e+1)*Hypotheses]
		var x [Hypotheses]float64
		between, within := 0.0, 0.0
		for h, w := range m.weights {
			x[h] = w * row[h]
			between += w * row[h] * row[h]
			p := m.p[e*Hypotheses+h]
			within += w * p * (1 - p) / 3
		}
		gain := 0.0
		for h, xh := range x {
			for k, xk := range x {
				gain += xh * gram[h*Hypotheses+k] * xk
			}
		}
		// Other events share theta only; the queried event also shares phi_e
		// with its future label. Restore that conditional Beta variance term.
		gain += priority[e] / normalizer * (2*between*within + within*within)
		variance := mean[e] * (1 - mean[e])
		if variance <= 0 || math.IsNaN(variance) {
			return nil, errors.New("degenerate predictive variance")
		}
		scores[e] = math.Max(0, gain/variance)
	}
	return scores, nil
}

func (m *Model) SelectRisk(priority []float64) (int, float64, error) {
	if m.count == len(m.base) {
		return -1, 0, errors.New("frontier exhausted")
	}
	scores, err := m.RiskScores(priority)
	if err != nil {
		return -1, 0, err
	}
	best, score := -1, -1.0
	for i, v := range scores {
		if v > score {
			best, score = i, v
		}
	}
	return best, score, nil
}
