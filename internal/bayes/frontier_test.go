package bayes_test

import (
	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"testing"
)

func TestEvaluateIsBoundedShadowOnlyAndPrioritySensitive(t *testing.T) {
	policy := bayes.Policy{VectorWeight: .6, NeighborWeight: .1, NoveltyWeight: .2, IndependenceWeight: .1, Threshold: .75, CriticalThreshold: .55, AuditProbability: .2, MaxActive: 1, AuditSeed: "fixed"}
	report := bayes.Evaluate([]bayes.Candidate{
		{EventID: "ordinary", VectorRelevance: .7, Novelty: .3, SourceIndependence: 1, EvidenceReady: true},
		{EventID: "critical", VectorRelevance: .7, Novelty: .3, SourceIndependence: 1, Priority: .9, EvidenceReady: true},
		{EventID: "hypothesis", VectorRelevance: 1, Novelty: 1, SourceIndependence: 1, Priority: 1, EvidenceReady: false},
	}, 4, policy)
	if report.Mode != "shadow" || report.SelectionSupportCertified {
		t.Fatalf("report = %+v", report)
	}
	if report.Activated != 1 {
		t.Fatalf("activated = %d", report.Activated)
	}
	if report.Decisions[0].EventID != "hypothesis" || report.Decisions[0].Activated {
		t.Fatalf("evidence readiness failed: %+v", report.Decisions[0])
	}
}

func TestAuditSelectionIsDeterministic(t *testing.T) {
	policy := bayes.Policy{AuditProbability: .5, MaxActive: 1, AuditSeed: "fixed"}
	candidates := []bayes.Candidate{{EventID: "a"}, {EventID: "b"}}
	first := bayes.Evaluate(candidates, 7, policy)
	second := bayes.Evaluate(candidates, 7, policy)
	for index := range first.Decisions {
		if first.Decisions[index].AuditSelected != second.Decisions[index].AuditSelected {
			t.Fatal("audit selection changed")
		}
	}
}
