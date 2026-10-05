package researchcalendar

import (
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
)

func packingFixture() ([]model.Candidate, PriorityPlan) {
	c := []model.Candidate{{Event: model.Event{ID: "bad", What: model.Field{Value: "Rosetta arrival"}}, Score: .99, EstimatedTokens: 1}, {Event: model.Event{ID: "good", What: model.Field{Value: "Rosetta launch"}}, Score: .01, EstimatedTokens: 1}}
	p := PriorityPlan{Method: "research/calendar-priority-v1", CalibrationStatus: "not_evaluated", Order: []int{1, 0}, Decisions: []TemporalDecision{{EventID: "bad", State: "contradicted"}, {EventID: "good", State: "compatible"}}}
	return c, p
}

func TestPriorityPackingPreservesDiversityAndScores(t *testing.T) {
	c, p := packingFixture()
	policy := packing.DefaultPolicy()
	policy.DiversityEnabled = true
	naive := packing.Select([]model.Candidate{c[1], c[0]}, nil, 1, 2, 100, policy)
	if naive.Candidates[0].Event.ID != "bad" {
		t.Fatal("diversity counterexample missing")
	}
	out, err := PackPriority(c, p, nil, 1, 2, 100, policy)
	if err != nil || len(out.Candidates) != 1 || !reflect.DeepEqual(out.Candidates[0], c[1]) {
		t.Fatal(out, err)
	}
	if c[1].Score != .01 {
		t.Fatal("input mutation")
	}
}

func TestPriorityPackingRespectsBudgetAndOccupancy(t *testing.T) {
	c, p := packingFixture()
	policy := packing.DefaultPolicy()
	policy.DiversityEnabled = true
	c[1].EstimatedTokens = 101
	out, err := PackPriority(c, p, nil, 1, 2, 100, policy)
	if err != nil || out.Candidates[0].Event.ID != "bad" || out.UsedTokens != 1 {
		t.Fatal("budget bypass", out, err)
	}
	c[1].EstimatedTokens = 1
	c[0].EvidenceGroupKey = "same"
	c[1].EvidenceGroupKey = "same"
	out, err = PackPriority(c, p, nil, 2, 2, 100, policy)
	if err != nil || len(out.Candidates) != 1 || out.Candidates[0].Event.ID != "good" || out.CorrelatedSuppressed != 1 {
		t.Fatal("occupancy bypass", out, err)
	}
}

func TestPriorityPackingExpansionUsesOriginalScores(t *testing.T) {
	c, p := packingFixture()
	c[0].Score = .51
	c[1].Score = .5
	policy := packing.DefaultPolicy()
	policy.AdaptiveEnabled = true
	policy.DiversityEnabled = true
	control := packing.Select(c, nil, 1, 2, 100, policy)
	out, err := PackPriority(c, p, nil, 1, 2, 100, policy)
	if err != nil || out.Expanded != control.Expanded || !out.Expanded || len(out.Candidates) != 2 {
		t.Fatal(out, control, err)
	}
	if out.Candidates[0].Score != .5 || out.Candidates[1].Score != .51 {
		t.Fatal("selection scores escaped", out)
	}
	p.Decisions[1].State = "contradicted"
	p.AllContradicted = true
	out, err = PackPriority(c, p, nil, 1, 2, 100, policy)
	if err != nil || !reflect.DeepEqual(out, control) {
		t.Fatal("all contradicted changed fallback", out, err)
	}
}
