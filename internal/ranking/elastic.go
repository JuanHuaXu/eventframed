package ranking

import "math"

// ElasticDeltaPolicy scales a raw correction using rank-domain answer certainty
// and an independent correction-reliability gate.
type ElasticDeltaPolicy struct {
	Enabled  bool
	MinScale float64
	MaxScale float64
}

func DefaultElasticDeltaPolicy() ElasticDeltaPolicy {
	return ElasticDeltaPolicy{Enabled: true, MinScale: .5, MaxScale: 2.5}
}

func (policy ElasticDeltaPolicy) Valid() bool {
	return !math.IsNaN(policy.MinScale) && !math.IsNaN(policy.MaxScale) &&
		!math.IsInf(policy.MinScale, 0) && !math.IsInf(policy.MaxScale, 0) &&
		policy.MinScale >= 0 && policy.MaxScale >= policy.MinScale && policy.MaxScale <= 10
}

func (policy ElasticDeltaPolicy) Scale(answerCertainty, correctionReliability float64) float64 {
	correctionReliability = clamp(correctionReliability)
	if !policy.Enabled {
		return correctionReliability
	}
	answerCertainty = clamp(answerCertainty)
	plasticity := policy.MinScale + (policy.MaxScale-policy.MinScale)*(1-answerCertainty)
	return plasticity * correctionReliability
}
