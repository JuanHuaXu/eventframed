package bayes_test

import (
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestBoundedChangepointDetectsAbruptRegimeShift(t *testing.T) {
	policy := bayes.ChangePolicy{Hazard: .05, Threshold: .30, MaxRun: 32}
	var posterior model.BayesianPosterior
	triggered := false
	for index := 0; index < 30; index++ {
		posterior, triggered = bayes.ApplyOutcome(posterior, true, 1, policy)
		if triggered {
			t.Fatalf("false changepoint at stable observation %d", index)
		}
		if len(posterior.RunLengthState.Probabilities) > policy.MaxRun+1 {
			t.Fatal("run-length state exceeded cap")
		}
	}
	for index := 0; index < 5 && !triggered; index++ {
		posterior, triggered = bayes.ApplyOutcome(posterior, false, 1, policy)
	}
	if !triggered {
		t.Fatalf("abrupt shift was not detected; probability=%f", posterior.ChangePointProbability)
	}
	if posterior.EffectiveSupport != 1 || posterior.Beta != 2 {
		t.Fatalf("posterior was not reset onto triggering observation: %+v", posterior)
	}
}
