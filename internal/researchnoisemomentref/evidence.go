package researchnoisemomentref

import (
	"errors"
	"math"
)

// Global evidence is recomputed from EVERY member's independently integrated
// marginal. No candidate subtraction cache or candidate log-evidence is used.
func (r *Reference) LogEvidence() (float64, error) {
	if !r.cfg.Shared {
		return 0, errors.New("reference shared evidence only")
	}
	var logs [28]float64
	maximum := math.Inf(-1)
	for h := 0; h < r.h; h++ {
		logs[h] = math.Log(r.weights[h])
		for i := range r.issued {
			logs[h] += r.logs[i*r.h+h]
		}
		maximum = math.Max(maximum, logs[h])
	}
	sum := 0.
	for h := 0; h < r.h; h++ {
		sum += math.Exp(logs[h] - maximum)
	}
	return maximum + math.Log(sum), nil
}
func (r *Reference) ObservedPredict(i int) (float64, error) {
	q, e := r.Predict(i)
	return r.cfg.Noise + (1-2*r.cfg.Noise)*q, e
}
