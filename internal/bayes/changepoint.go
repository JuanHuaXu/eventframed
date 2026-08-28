package bayes

import (
	"math"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

type ChangePolicy struct {
	Hazard    float64
	Threshold float64
	MaxRun    int
}

func (policy ChangePolicy) Valid() bool {
	return policy.Hazard > 0 && policy.Hazard < 1 && policy.Threshold > 0 && policy.Threshold <= 1 && policy.MaxRun > 1
}

// UpdateRunLength performs one capped Bayesian online changepoint update for a
// Bernoulli usefulness observation. The cap makes update cost and stored state
// independent of trajectory length.
func UpdateRunLength(state model.BayesianRunLengthState, success bool, policy ChangePolicy) (model.BayesianRunLengthState, float64) {
	if !policy.Valid() {
		return state, 0
	}
	if len(state.Probabilities) == 0 || len(state.Alpha) != len(state.Probabilities) || len(state.Beta) != len(state.Probabilities) {
		state = model.BayesianRunLengthState{Probabilities: []float64{1}, Alpha: []float64{1}, Beta: []float64{1}}
	}
	length := min(len(state.Probabilities)+1, policy.MaxRun+1)
	next := model.BayesianRunLengthState{Probabilities: make([]float64, length), Alpha: make([]float64, length), Beta: make([]float64, length)}
	x := 0.0
	if success {
		x = 1
	}
	priorPredictive := .5
	for index, probability := range state.Probabilities {
		predictive := (state.Alpha[index]*x + state.Beta[index]*(1-x)) / (state.Alpha[index] + state.Beta[index])
		next.Probabilities[0] += probability * policy.Hazard * priorPredictive
		growth := min(index+1, length-1)
		next.Probabilities[growth] += probability * (1 - policy.Hazard) * predictive
	}
	normalizer := 0.0
	for _, probability := range next.Probabilities {
		normalizer += probability
	}
	if normalizer <= 0 || math.IsNaN(normalizer) || math.IsInf(normalizer, 0) {
		return model.BayesianRunLengthState{Probabilities: []float64{1}, Alpha: []float64{1 + x}, Beta: []float64{2 - x}}, 1
	}
	for index := range next.Probabilities {
		next.Probabilities[index] /= normalizer
	}
	next.Alpha[0], next.Beta[0] = 1+x, 2-x
	for index := 1; index < length; index++ {
		source := min(index-1, len(state.Alpha)-1)
		next.Alpha[index] = state.Alpha[source] + x
		next.Beta[index] = state.Beta[source] + 1 - x
	}
	return next, next.Probabilities[0]
}

func ApplyOutcome(posterior model.BayesianPosterior, success bool, weight float64, policy ChangePolicy) (model.BayesianPosterior, bool) {
	state, probability := UpdateRunLength(posterior.RunLengthState, success, policy)
	posterior.RunLengthState = state
	posterior.ChangePointProbability = probability
	triggered := probability >= policy.Threshold && len(state.Probabilities) > 1
	if triggered {
		posterior.Alpha, posterior.Beta, posterior.EffectiveSupport = 1, 1, 0
	}
	if success {
		posterior.Alpha += weight
	} else {
		posterior.Beta += weight
	}
	posterior.EffectiveSupport += weight
	return posterior, triggered
}
