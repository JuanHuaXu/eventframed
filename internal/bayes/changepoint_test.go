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

func TestBoundedChangepointDetectsPersistentGradualDrift(t *testing.T) {
	policy := bayes.ChangePolicy{
		Hazard: .01, Threshold: .95, MaxRun: 64, RecentWindow: 4, RecentThreshold: .99,
		FastRate: .25, SlowRate: .025, DriftThreshold: .20, DriftPersistence: 3, MinSamples: 20,
	}
	var posterior model.BayesianPosterior
	triggered := false
	for index := 0; index < 35; index++ {
		posterior, triggered = bayes.ApplyOutcome(posterior, true, 1, policy)
		if triggered {
			t.Fatalf("false changepoint during stable prefix at %d", index)
		}
	}
	for index := 0; index < 20 && !triggered; index++ {
		// A deterministic ramp is represented by progressively denser failures.
		success := index%4 > index/7
		posterior, triggered = bayes.ApplyOutcome(posterior, success, 1, policy)
	}
	if !triggered {
		t.Fatalf("persistent drift was not detected: %+v", posterior.DriftState)
	}
}

func TestChangepointResetClearsDetectorMomentum(t *testing.T) {
	policy := bayes.ChangePolicy{Hazard: .05, Threshold: .30, MaxRun: 32}
	var posterior model.BayesianPosterior
	for index := 0; index < 30; index++ {
		posterior, _ = bayes.ApplyOutcome(posterior, true, 1, policy)
	}
	posterior, triggered := bayes.ApplyOutcome(posterior, false, 1, policy)
	if !triggered {
		t.Fatal("expected clean reversal to trigger")
	}
	if len(posterior.RunLengthState.Probabilities) != 1 || posterior.DriftState.Samples != 1 {
		t.Fatalf("detector momentum survived reset: %+v", posterior)
	}
	posterior, triggered = bayes.ApplyOutcome(posterior, false, 1, policy)
	if triggered {
		t.Fatal("stale detector momentum caused an immediate repeat trigger")
	}
}

func TestCUSUMWarmupIgnoresInitialSampleBiasAndDetectsReversal(t *testing.T) {
	policy := bayes.ChangePolicy{
		Hazard: .05, Threshold: 1, MaxRun: 64,
		FastRate: .25, SlowRate: .025, DriftThreshold: .30, DriftPersistence: 12, MinSamples: 20,
		CUSUMSlack: .10, CUSUMThreshold: 8, CooldownSamples: 20,
	}
	var posterior model.BayesianPosterior
	for index := 0; index < 80; index++ {
		// The unlucky first failure is followed by a stable 0.8-rate sequence.
		success := index > 0 && index%5 != 0
		var triggered bool
		posterior, triggered = bayes.ApplyOutcome(posterior, success, 1, policy)
		if triggered {
			t.Fatalf("stable sequence triggered at %d: %+v", index, posterior.DriftState)
		}
	}
	triggered := false
	for index := 0; index < 20 && !triggered; index++ {
		posterior, triggered = bayes.ApplyOutcome(posterior, false, 1, policy)
	}
	if !triggered {
		t.Fatalf("sustained reversal was not detected: %+v", posterior.DriftState)
	}
}

func TestApplyOutcomeAuthorizedMonitorsWithoutResetAuthority(t *testing.T) {
	policy := bayes.ChangePolicy{Hazard: .05, Threshold: .30, MaxRun: 32}
	posterior := model.BayesianPosterior{Alpha: 101, Beta: 1, EffectiveSupport: 100}
	for index := 0; index < 12; index++ {
		var reset bool
		posterior, reset = bayes.ApplyOutcomeAuthorized(posterior, false, 1, policy, false)
		if reset {
			t.Fatal("unauthorized evidence reset the posterior")
		}
	}
	if posterior.Alpha != 101 || posterior.Beta != 13 || posterior.EffectiveSupport != 112 {
		t.Fatalf("unauthorized monitor discarded ordinary updates: %+v", posterior)
	}
}
