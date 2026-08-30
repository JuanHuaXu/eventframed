package bayes

import (
	"math"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

type ChangePolicy struct {
	Hazard           float64
	Threshold        float64
	MaxRun           int
	RecentWindow     int
	RecentThreshold  float64
	FastRate         float64
	SlowRate         float64
	DriftThreshold   float64
	DriftPersistence int
	MinSamples       int
	CUSUMSlack       float64
	CUSUMThreshold   float64
	CooldownSamples  int
}

func (policy ChangePolicy) Valid() bool {
	coreValid := policy.Hazard > 0 && policy.Hazard < 1 && policy.Threshold > 0 && policy.Threshold <= 1 && policy.MaxRun > 1
	recentDisabled := policy.RecentWindow == 0 && policy.RecentThreshold == 0
	recentValid := policy.RecentWindow > 0 && policy.RecentWindow <= policy.MaxRun && policy.RecentThreshold > 0 && policy.RecentThreshold <= 1
	driftDisabled := policy.FastRate == 0 && policy.SlowRate == 0 && policy.DriftThreshold == 0 && policy.DriftPersistence == 0 && policy.MinSamples == 0 && policy.CUSUMSlack == 0 && policy.CUSUMThreshold == 0 && policy.CooldownSamples == 0
	cusumDisabled := policy.CUSUMSlack == 0 && policy.CUSUMThreshold == 0
	cusumValid := policy.CUSUMSlack > 0 && policy.CUSUMSlack < 1 && policy.CUSUMThreshold > 0
	driftValid := policy.FastRate > 0 && policy.FastRate <= 1 && policy.SlowRate > 0 && policy.SlowRate < policy.FastRate && policy.DriftThreshold > 0 && policy.DriftThreshold <= 1 && policy.DriftPersistence > 0 && policy.MinSamples > 1 && policy.CooldownSamples >= 0 && (cusumDisabled || cusumValid)
	return coreValid && (recentDisabled || recentValid) && (driftDisabled || driftValid)
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
	return ApplyOutcomeAuthorized(posterior, success, weight, policy, true)
}

// ApplyOutcomeAuthorized updates detector state for every observation but only
// commits a posterior reset when resetAuthorized is true. This lets selected
// evidence monitor a certified shared group without invalidating it by itself.
func ApplyOutcomeAuthorized(posterior model.BayesianPosterior, success bool, weight float64, policy ChangePolicy, resetAuthorized bool) (model.BayesianPosterior, bool) {
	coolingDown := posterior.DriftState.Cooldown > 0
	state, probability := UpdateRunLength(posterior.RunLengthState, success, policy)
	posterior.RunLengthState = state
	posterior.ChangePointProbability = probability
	posterior.RecentChangeProbability = 0
	recentTriggered := false
	if policy.RecentWindow > 0 {
		posterior.RecentChangeProbability = recentRunProbability(state, policy.RecentWindow)
		recentTriggered = len(state.Probabilities) > 2*policy.RecentWindow && posterior.RecentChangeProbability >= policy.RecentThreshold
	}
	driftState, driftTriggered := updateDrift(posterior.DriftState, success, policy)
	posterior.DriftState = driftState
	detected := !coolingDown && len(state.Probabilities) > 1 && (probability >= policy.Threshold || recentTriggered || driftTriggered)
	triggered := detected && resetAuthorized
	if triggered {
		posterior.Alpha, posterior.Beta, posterior.EffectiveSupport = 1, 1, 0
		posterior.CalibrationWeight, posterior.BrierLossSum, posterior.ForecastUsefulSum, posterior.ObservedUsefulSum = 0, 0, 0, 0
		posterior.RunLengthState = resetRunLength(success)
		posterior.DriftState = resetDrift(success)
		posterior.DriftState.Cooldown = policy.CooldownSamples
	}
	if success {
		posterior.Alpha += weight
	} else {
		posterior.Beta += weight
	}
	posterior.EffectiveSupport += weight
	return posterior, triggered
}

func recentRunProbability(state model.BayesianRunLengthState, window int) float64 {
	limit := min(window+1, len(state.Probabilities))
	probability := 0.0
	for index := 0; index < limit; index++ {
		probability += state.Probabilities[index]
	}
	return probability
}

func updateDrift(state model.BayesianDriftState, success bool, policy ChangePolicy) (model.BayesianDriftState, bool) {
	if policy.FastRate == 0 {
		return state, false
	}
	x := 0.0
	if success {
		x = 1
	}
	if state.Samples == 0 {
		return resetDrift(success), false
	}
	if state.Cooldown > 0 {
		state.Cooldown--
	}
	if policy.CUSUMThreshold > 0 && state.Samples < policy.MinSamples {
		state.Samples++
		state.SlowMean += (x - state.SlowMean) / float64(state.Samples)
		state.FastMean = state.SlowMean
		state.UpwardCUSUM, state.DownwardCUSUM = 0, 0
		return state, false
	}
	residual := x - state.SlowMean
	state.FastMean += policy.FastRate * (x - state.FastMean)
	state.SlowMean += policy.SlowRate * (x - state.SlowMean)
	state.Samples++
	if policy.CUSUMThreshold > 0 {
		state.UpwardCUSUM = math.Max(0, state.UpwardCUSUM+residual-policy.CUSUMSlack)
		state.DownwardCUSUM = math.Min(0, state.DownwardCUSUM+residual+policy.CUSUMSlack)
		triggered := state.UpwardCUSUM >= policy.CUSUMThreshold || -state.DownwardCUSUM >= policy.CUSUMThreshold
		return state, triggered
	}
	difference := state.FastMean - state.SlowMean
	direction := 0
	if difference >= policy.DriftThreshold {
		direction = 1
	} else if difference <= -policy.DriftThreshold {
		direction = -1
	}
	if state.Samples < policy.MinSamples || direction == 0 {
		state.Streak, state.Direction = 0, 0
		return state, false
	}
	if state.Direction == direction {
		state.Streak++
	} else {
		state.Direction, state.Streak = direction, 1
	}
	return state, state.Streak >= policy.DriftPersistence
}

func resetRunLength(success bool) model.BayesianRunLengthState {
	alpha, beta := 1.0, 2.0
	if success {
		alpha, beta = 2, 1
	}
	return model.BayesianRunLengthState{Probabilities: []float64{1}, Alpha: []float64{alpha}, Beta: []float64{beta}}
}

func resetDrift(success bool) model.BayesianDriftState {
	mean := 0.0
	if success {
		mean = 1
	}
	return model.BayesianDriftState{FastMean: mean, SlowMean: mean, Samples: 1}
}
