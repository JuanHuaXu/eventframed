package researchpairedvalueref

import (
	"errors"
	"math"
)

// PredictionValue independently integrates unnormalized full histories and
// computes risk subtraction, rather than the learner's squared movements.
func (r *Reference) PredictionValue(i, ordinal int, targets []float64) (float64, float64, error) {
	if i < 0 || i >= r.n || ordinal < 1 || ordinal > r.issued[i] || len(targets) != r.n {
		return 0, 0, errors.New("value reference identity")
	}
	sum := 0.
	for _, w := range targets {
		if math.IsNaN(w) || math.IsInf(w, 0) || w < 0 || w > 1 {
			return 0, 0, errors.New("value reference target")
		}
		sum += w
	}
	if math.Abs(sum-1) > 1e-12 {
		return 0, 0, errors.New("value reference normalization")
	}
	k := i*64 + ordinal - 1
	x := r.trials[k]
	if !x.known1 || x.requested {
		return 0, 0, errors.New("value reference availability")
	}
	priorWeights, e := r.Weights()
	if e != nil {
		return 0, 0, e
	}
	current := make([]float64, r.n)
	risk := 0.
	for j := range current {
		for c, w := range priorWeights {
			current[j] += w * r.next[j*84+c]
		}
		risk += targets[j] * current[j] * (1 - current[j])
	}
	expected, observed := 0., 0.
	tower := make([]float64, r.n)
	x.known2 = true
	for branch := 0; branch < 2; branch++ {
		x.y2 = branch == 0
		var logs, own, posterior [84]float64
		probability := 0.
		maximum := math.Inf(-1)
		for c := range logs {
			logs[c], own[c] = r.mass(i, c, k, x)
			if priorWeights[c] > 0 {
				probability += priorWeights[c] * math.Exp(logs[c]-r.logs[i*84+c])
			}
			v := math.Log(componentPrior(c))
			for j := 0; j < r.n; j++ {
				if j == i {
					v += logs[c]
				} else {
					v += r.logs[j*84+c]
				}
			}
			posterior[c] = v
			maximum = math.Max(maximum, v)
		}
		if math.IsNaN(probability) || math.IsInf(probability, 0) || probability < 0 || probability > 1+2e-10 {
			return 0, 0, errors.New("value reference branch probability")
		}
		probability = math.Min(1, probability)
		if branch == 0 {
			observed = probability
		}
		if probability == 0 {
			continue
		}
		if math.IsNaN(maximum) || math.IsInf(maximum, 0) {
			return 0, 0, errors.New("value reference posterior support")
		}
		norm := 0.
		for c := range posterior {
			posterior[c] = math.Exp(posterior[c] - maximum)
			norm += posterior[c]
		}
		for c := range posterior {
			posterior[c] /= norm
		}
		for j := 0; j < r.n; j++ {
			q := 0.
			for c, w := range posterior {
				mean := r.next[j*84+c]
				if j == i {
					mean = own[c]
				}
				q += w * mean
			}
			tower[j] += probability * q
			expected += probability * targets[j] * q * (1 - q)
		}
	}
	for j := range tower {
		if math.Abs(tower[j]-current[j]) > 2e-10 {
			return 0, 0, errors.New("value reference tower defect")
		}
	}
	value := risk - expected
	if math.IsNaN(value) || math.IsInf(value, 0) || value < -2e-10 || value > .25+2e-10 {
		return 0, 0, errors.New("value reference risk bound")
	}
	return observed, math.Max(0, value), nil
}
