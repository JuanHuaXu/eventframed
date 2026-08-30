package packing

import (
	"fmt"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestAdaptivePackingExpandsOnSmallBoundaryMargin(t *testing.T) {
	candidates := fixtureCandidates(12)
	policy := DefaultPolicy()
	policy.AdaptiveEnabled = true
	result := Select(candidates, nil, 5, 12, 1000, policy)
	if !result.Expanded || len(result.Candidates) != 10 {
		t.Fatalf("adaptive result = %+v", result)
	}
}

func TestDiversityDoesNotSuppressDistinctAntiPigeonBuckets(t *testing.T) {
	candidates := fixtureCandidates(3)
	candidates[0].Event.Content = "same repeated content"
	candidates[1].Event.Content = "same repeated content"
	candidates[2].Event.Content = "different material"
	candidates[0].Score, candidates[1].Score, candidates[2].Score = .9, .89, .88
	policy := DefaultPolicy()
	policy.DiversityEnabled = true
	keys := map[string]string{candidates[0].Event.ID: "ap:a", candidates[1].Event.ID: "ap:b"}
	result := Select(candidates, keys, 2, 3, 1000, policy)
	if result.Candidates[1].Event.ID != candidates[1].Event.ID {
		t.Fatalf("distinct Anti-Pigeon bucket was suppressed: %+v", result.Candidates)
	}
}

func fixtureCandidates(count int) []model.Candidate {
	result := make([]model.Candidate, count)
	for index := range result {
		probability := .7 - float64(index)*.001
		result[index] = model.Candidate{
			Event: model.Event{ID: fmt.Sprintf("event-%d", index), Content: fmt.Sprintf("content %d", index)},
			Score: probability, EstimatedTokens: 1,
			Forecast: model.ForecastBundle{CorrectedLaw: model.BernoulliLaw{Useful: probability, NotUseful: 1 - probability}},
		}
	}
	return result
}
