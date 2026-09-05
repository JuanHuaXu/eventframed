package bayes

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// WorkingPolicy declares two Bernoulli hypotheses about usefulness, not truth.
// Forgetting and capped evidence make this a working filter, not exact Bayes.
type WorkingPolicy struct {
	Enabled                                                    bool
	Retention, MaxLogOdds, MaxLogFactor, LowUseful, HighUseful float64
}

func DefaultWorkingPolicy() WorkingPolicy {
	return WorkingPolicy{true, .98, 6, 2, .2, .8}
}

func (p WorkingPolicy) Valid() bool {
	if !p.Enabled {
		return p == (WorkingPolicy{})
	}
	for _, v := range []float64{p.Retention, p.MaxLogOdds, p.MaxLogFactor, p.LowUseful, p.HighUseful} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return p.Retention >= 0 && p.Retention <= 1 && p.MaxLogOdds > 0 && p.MaxLogOdds <= 30 && p.MaxLogFactor > 0 && p.MaxLogFactor <= 30 && p.LowUseful > 0 && p.LowUseful < p.HighUseful && p.HighUseful < 1
}

func (p WorkingPolicy) ID() string {
	b, _ := json.Marshal(p)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func UpdateWorking(previous *model.WorkingBelief, success bool, weight float64, reset bool, p WorkingPolicy) *model.WorkingBelief {
	if !p.Enabled || !p.Valid() {
		return nil
	}
	id := p.ID()
	odds := 0.0 // Equal prior odds; no imported Beta certainty after a split.
	if !reset && previous != nil && previous.PolicyID == id && !math.IsNaN(previous.LogOdds) && !math.IsInf(previous.LogOdds, 0) {
		odds = max(-p.MaxLogOdds, min(p.MaxLogOdds, previous.LogOdds))
	}
	factor := math.Log((1 - p.HighUseful) / (1 - p.LowUseful))
	if success {
		factor = math.Log(p.HighUseful / p.LowUseful)
	}
	// Importance weights cannot manufacture multiple independent observations.
	if math.IsNaN(weight) || math.IsInf(weight, 0) {
		weight = 0
	}
	factor = max(-p.MaxLogFactor, min(p.MaxLogFactor, factor))
	odds = max(-p.MaxLogOdds, min(p.MaxLogOdds, p.Retention*odds+max(0, min(1, weight))*factor))
	return &model.WorkingBelief{PolicyID: id, LogOdds: odds, PredictiveUseful: workingPredictive(odds, p)}
}

func workingPredictive(odds float64, p WorkingPolicy) float64 {
	return p.LowUseful + (p.HighUseful-p.LowUseful)/(1+math.Exp(-odds))
}

// PredictiveMean maps hypothesis belief to the outcome law, not to P(H_high).
func PredictiveMean(posterior model.BayesianPosterior, p WorkingPolicy) float64 {
	if !p.Enabled {
		return posterior.Mean()
	}
	state := posterior.WorkingBelief
	if state == nil || state.PolicyID != p.ID() || math.IsNaN(state.LogOdds) || math.IsInf(state.LogOdds, 0) {
		return workingPredictive(0, p)
	}
	return workingPredictive(max(-p.MaxLogOdds, min(p.MaxLogOdds, state.LogOdds)), p)
}
