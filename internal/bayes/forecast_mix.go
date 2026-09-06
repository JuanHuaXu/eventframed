package bayes

import (
	"github.com/JuanHuaXu/eventframed/internal/model"
	"math"
)

// ForecastMix is a fixed-size online selector of complete forecast laws. Its
// weights express relative predictive performance, not factual truth.
type ForecastMix struct {
	Weights [4]float64 `json:"weights"`
}

// Called inside the outcome transaction, after replay detection and before
// publication. A revealing changepoint clears stale performance evidence; its
// old-regime forecasts do not initialize the new regime's selector weights.
func UpdateForecastWeights(weights [4]float64, record model.ExpertForecasts, useful bool, weight float64, reset, enabled bool, snapshot model.Snapshot) [4]float64 {
	if reset {
		return [4]float64{}
	}
	if !enabled || !record.Enabled || record.PolicyVersion != snapshot.PolicyVersion || record.EvidenceEpoch != snapshot.EvidenceEpoch {
		return weights
	}
	return (ForecastMix{Weights: weights}).Observe(record.Probabilities, useful, weight).Weights
}

func mixPrior() [4]float64 { return [4]float64{.7, .1, .1, .1} }

func (s ForecastMix) predictiveWeights() [4]float64 {
	prior := mixPrior()
	sum := 0.
	for _, w := range s.Weights {
		if math.IsNaN(w) || math.IsInf(w, 0) || w < 0 || w > 1 {
			return prior
		}
		sum += w
	}
	if math.Abs(sum-1) > 1e-9 {
		return prior
	}
	for j := range prior {
		prior[j] = .998*s.Weights[j]/sum + .002*prior[j]
	}
	return prior
}

func ValidForecastExperts(experts [4]float64) bool {
	for _, p := range experts {
		if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
			return false
		}
	}
	return true
}

// Forecast does not advance share on reads. Each expert is already a complete
// Bernoulli law; calibration/shrinkage must not be applied again to the mixture.
func (s ForecastMix) Forecast(experts [4]float64) float64 {
	if !ValidForecastExperts(experts) {
		return .5
	}
	w := s.predictiveWeights()
	p := 0.
	for j, v := range experts {
		p += w[j] * max(1e-6, min(1-1e-6, v))
	}
	return p
}

// Observe must receive journaled PRE-outcome forecasts, never regenerated ones.
// Delayed feedback updates current weights in arrival order; it is not claimed
// to be the immediate-feedback Bayesian posterior for the original time order.
func (s ForecastMix) Observe(experts [4]float64, useful bool, weight float64) ForecastMix {
	if !ValidForecastExperts(experts) || math.IsNaN(weight) || math.IsInf(weight, 0) || weight <= 0 {
		return s
	}
	w := s.predictiveWeights()
	weight = min(1, weight)
	sum := 0.
	for j, p := range experts {
		p = max(1e-6, min(1-1e-6, p))
		if !useful {
			p = 1 - p
		}
		w[j] *= math.Pow(p, weight)
		sum += w[j]
	}
	for j := range w {
		w[j] /= sum
	}
	return ForecastMix{Weights: w}
}
