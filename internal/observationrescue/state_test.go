package observationrescue

import (
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
)

func TestProductionRevisionSeparatesSplitAndCommonChange(t *testing.T) {
	if !ChangePolicy().Valid() || !GroupPolicy().Valid() {
		t.Fatal("invalid frozen policy")
	}
	for _, divergent := range []bool{false, true} {
		var m Monitor
		split, reset := false, false
		for i := 0; i < 180; i++ {
			live := i < 100
			ref := live
			if divergent {
				ref = true
			}
			r, _ := m.Observe(ref, live, true)
			split = split || bayes.RevisionSplits(r.Action)
			reset = reset || r.Action == model.BayesianRevisionSharedReset
		}
		if divergent && !split {
			t.Fatal("divergent members never split")
		}
		if !divergent && (split || !reset) {
			t.Fatal("common change should reset, not split")
		}
	}
}
func TestSelectedOnlyEvidenceCannotInvalidate(t *testing.T) {
	var m Monitor
	for i := 0; i < 180; i++ {
		r, cp := m.Observe(true, i < 100, false)
		if cp || r.Action != model.BayesianRevisionRetain {
			t.Fatal("selected stream gained authority")
		}
	}
}
func TestFeedbackIsPrequentialAndDuplicateSafe(t *testing.T) {
	base, err := observationexperiment.FitFamily(2)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(base, "audit_learning")
	if err != nil {
		t.Fatal(err)
	}
	reader := observationexperiment.Frames(511, "test")
	if _, err := s.Observe(Feedback{}); err == nil {
		t.Fatal("feedback without prediction")
	}
	for i := 0; i < 32; i++ {
		forecast, err := s.Predict(reader, i)
		if err != nil {
			t.Fatal(err)
		}
		if i == 31 && forecast.Epoch != 1 {
			t.Fatal("revealing outcome leaked into current forecast")
		}
		if _, err := s.Predict(reader, i); err == nil {
			t.Fatal("double prediction accepted")
		}
		f := Feedback{Sequence: i, Epoch: forecast.Epoch, ReferenceCorrect: true, LiveCorrect: true, ValidationEligible: true, Audit: &observation.Sample{Bits: 511, Outcome: true}}
		bad := f
		bad.Epoch++
		if _, err := s.Observe(bad); err == nil {
			t.Fatal("stale epoch accepted")
		}
		if _, err := s.Observe(f); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Observe(f); err == nil {
			t.Fatal("duplicate feedback accepted")
		}
	}
	if s.FitCount != 1 || len(s.audits) != 32 || base.Support() != 4096 {
		t.Fatal("incorrect learning/support")
	}
	f, err := s.Predict(reader, 32)
	if err != nil {
		t.Fatal(err)
	}
	if f.Epoch != 2 || f.Support != 32 {
		t.Fatal("new model did not apply to next forecast")
	}
}
func TestRepeatedOldDivergenceDoesNotEraseRecovery(t *testing.T) {
	base, _ := observationexperiment.FitFamily(2)
	s, _ := New(base, "ap_audit")
	reader := observationexperiment.Frames(511, "test")
	revocations := 0
	for i := 0; i < 256; i++ {
		f, err := s.Predict(reader, i)
		if err != nil {
			t.Fatal(err)
		}
		tr, err := s.Observe(Feedback{Sequence: i, Epoch: f.Epoch, ReferenceCorrect: true, LiveCorrect: i < 100, ValidationEligible: true, Audit: &observation.Sample{Bits: 511, Outcome: true}})
		if err != nil {
			t.Fatal(err)
		}
		if tr.Invalidated {
			revocations++
		}
	}
	if revocations != 1 || s.FitCount == 0 || len(s.audits) > AuditCapacity {
		t.Fatal("revocation loop or missing recovery")
	}
}
