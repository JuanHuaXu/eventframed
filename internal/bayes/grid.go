package bayes

import (
	"math"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const gridShare = .02

func gridProbability(j int) float64 {
	return max(.01, min(.99, float64(j)/20))
}

// State holds the posterior AFTER the last outcome. Prediction and the next
// update each derive the same reset-HMM prior, without modifying stored state.
// Hazard runs on admitted observations, not wall time or rejected replays.
func gridPrior(state *model.WorkingBelief, reset bool, p WorkingPolicy) [21]float64 {
	var weights [21]float64
	valid := !reset && state != nil && state.PolicyID == p.ID()
	sum := 0.0
	if valid {
		for _, w := range state.GridWeights {
			if math.IsNaN(w) || math.IsInf(w, 0) || w < 0 || w > 1 {
				valid = false
			}
			sum += w
		}
		valid = valid && math.Abs(sum-1) <= 1e-9
	}
	for j := range weights {
		weights[j] = 1.0 / 21
		if valid {
			weights[j] = (1-gridShare)*state.GridWeights[j]/sum + gridShare/21
		}
	}
	return weights
}

func gridMean(weights [21]float64) float64 {
	p := 0.0
	for j, w := range weights {
		p += w * gridProbability(j)
	}
	return p
}

// Unit weight is ordinary Bayes for the finite reset model; fractional weights
// are generalized Bayes. This does not establish selection ignorability or truth.
func updateGrid(previous *model.WorkingBelief, success bool, weight float64, reset bool, p WorkingPolicy) *model.WorkingBelief {
	weights := gridPrior(previous, reset, p)
	if math.IsNaN(weight) || math.IsInf(weight, 0) {
		weight = 0
	}
	weight = max(0, min(1, weight))
	sum := 0.0
	for j := range weights {
		likelihood := gridProbability(j)
		if !success {
			likelihood = 1 - likelihood
		}
		if weight != 1 {
			likelihood = math.Pow(likelihood, weight)
		}
		weights[j] *= likelihood
		sum += weights[j]
	}
	for j := range weights {
		weights[j] /= sum
	}
	state := &model.WorkingBelief{PolicyID: p.ID(), GridWeights: weights}
	state.PredictiveUseful = gridMean(gridPrior(state, false, p))
	return state
}
