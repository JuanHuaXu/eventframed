package residual_test

import (
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
)

func TestLawResidualPreservesMassAndRequiresCertifiedMotion(t *testing.T) {
	policy := residual.Policy{Clip: .2, MinSupport: 2, MinConfidence: .5, ConfidenceDelta: .05, MotionLimit: .1, MaxAge: time.Hour, ImprovementDelta: .001}
	now := time.Now().UTC()
	snapshot := model.Snapshot{PolicyVersion: 2, EvidenceEpoch: 3, ResidualVersion: 4}
	observation := model.ResidualObservation{HorizonKey: model.RetrievalUsefulnessHorizon, BaseProbability: .4, Useful: true, ValidationEligible: true, EventID: "event", JournalID: "journal", AvailableAt: now}
	var record model.ResidualRecord
	for index := 0; index < 80; index++ {
		record = residual.Update(record, observation, model.ResidualExact, "key", "tenant", 1, snapshot, policy)
	}
	if !residual.Eligible(record, .45, snapshot, now, policy) {
		t.Fatalf("record should be eligible: %+v", record)
	}
	corrected := residual.Apply(model.BernoulliLaw{Useful: .45, NotUseful: .55}, record, policy)
	if corrected.Useful <= .45 || corrected.Useful+corrected.NotUseful != 1 {
		t.Fatalf("corrected law = %+v", corrected)
	}
	if residual.Eligible(record, .7, snapshot, now, policy) {
		t.Fatal("record survived motion beyond its fixed reference limit")
	}
}
