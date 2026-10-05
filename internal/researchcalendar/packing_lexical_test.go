package researchcalendar

import (
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"reflect"
	"testing"
)

func TestLexicalPackingBoundaries(t *testing.T) {
	candidates := []model.Candidate{
		{Event: model.Event{ID: "a"}, Score: .8, EstimatedTokens: 1, Forecast: model.ForecastBundle{CorrectedLaw: model.BernoulliLaw{Useful: 1}}},
		{Event: model.Event{ID: "b"}, Score: .2, EstimatedTokens: 1, Forecast: model.ForecastBundle{CorrectedLaw: model.BernoulliLaw{NotUseful: 1}}},
		{Event: model.Event{ID: "c"}, Score: .4, EstimatedTokens: 1, Forecast: model.ForecastBundle{CorrectedLaw: model.BernoulliLaw{NotUseful: 1}}},
	}
	original := append([]model.Candidate(nil), candidates...)
	plan := PriorityPlan{Method: "research/task-role-v1", CalibrationStatus: "not_evaluated", Order: []int{0, 1, 2}, Decisions: []TemporalDecision{{EventID: "a", State: "unknown"}, {EventID: "b", State: "unknown"}, {EventID: "c", State: "unknown"}}}
	lex := LexicalResult{Method: "what-lexical-v2", Order: []int{1, 2, 0}, Scores: []float64{.1, .9, .8}}
	for _, diversity := range []bool{false, true} {
		policy := packing.DefaultPolicy()
		policy.DiversityEnabled = diversity
		policy.AdaptiveEnabled = true
		got, err := PackTaskLexical(candidates, plan, lex, nil, 1, 3, 100, policy)
		if err != nil || got.Expanded || len(got.Candidates) != 1 || !reflect.DeepEqual(got.Candidates[0], original[1]) {
			t.Fatal("original boundary/record violated", got, err)
		}
		if !reflect.DeepEqual(candidates, original) {
			t.Fatal("input mutated")
		}
	}
	policy := packing.DefaultPolicy()
	candidates[1].EstimatedTokens = 10
	got, err := PackTaskLexical(candidates, plan, lex, nil, 1, 3, 2, policy)
	if err != nil || len(got.Candidates) != 1 || got.Candidates[0].Event.ID != "c" {
		t.Fatal("budget bypass", got, err)
	}
	candidates[1].EstimatedTokens = 1
	candidates[1].EvidenceGroupKey = "shared"
	candidates[2].EvidenceGroupKey = "shared"
	got, err = PackTaskLexical(candidates, plan, lex, nil, 3, 3, 100, policy)
	if err != nil || got.CorrelatedSuppressed < 1 {
		t.Fatal("occupancy bypass", got, err)
	}
	lex.Order = []int{1, 1, 0}
	if _, err = PackTaskLexical(candidates, plan, lex, nil, 1, 3, 100, policy); err == nil {
		t.Fatal("invalid permutation")
	}
}
