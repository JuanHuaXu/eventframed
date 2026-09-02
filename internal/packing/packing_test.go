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
	candidates[0].Event.What.Value = "same repeated event"
	candidates[1].Event.What.Value = "same repeated event"
	candidates[2].Event.What.Value = "different material"
	candidates[0].Score, candidates[1].Score, candidates[2].Score = .9, .89, .88
	policy := DefaultPolicy()
	policy.DiversityEnabled = true
	keys := map[string]string{candidates[0].Event.ID: "ap:a", candidates[1].Event.ID: "ap:b"}
	result := Select(candidates, keys, 2, 3, 1000, policy)
	if result.Candidates[1].Event.ID != candidates[1].Event.ID {
		t.Fatalf("distinct Anti-Pigeon bucket was suppressed: %+v", result.Candidates)
	}
}

func TestDiversityIgnoresRawContent(t *testing.T) {
	candidates := fixtureCandidates(3)
	candidates[0].Event.Content = "RAW_ALPHA"
	candidates[1].Event.Content = "RAW_BETA"
	candidates[0].Event.What.Value = "same event frame"
	candidates[1].Event.What.Value = "same event frame"
	candidates[2].Event.What.Value = "different event frame"
	candidates[0].Score, candidates[1].Score, candidates[2].Score = .9, .89, .88
	policy := DefaultPolicy()
	policy.DiversityEnabled = true
	policy.DiversityPenalty = .2
	result := Select(candidates, nil, 2, 3, 1000, policy)
	if result.Candidates[1].Event.ID != candidates[2].Event.ID {
		t.Fatalf("raw payload affected diversity ordering: %+v", result.Candidates)
	}
}

func TestEvidenceOccupancySuppressesRepeatedClaimAndLineage(t *testing.T) {
	candidates := fixtureCandidates(3)
	candidates[0].EvidenceGroupKey = "repeated"
	candidates[1].EvidenceGroupKey = "repeated"
	candidates[2].EvidenceGroupKey = "independent"
	result := Select(candidates, nil, 2, 3, 1000, DefaultPolicy())
	if len(result.Candidates) != 2 || result.Candidates[0].Event.ID != "event-0" || result.Candidates[1].Event.ID != "event-2" {
		t.Fatalf("packed correlated records: %+v", result.Candidates)
	}
	if result.CorrelatedSuppressed != 1 {
		t.Fatalf("correlated suppressed = %d", result.CorrelatedSuppressed)
	}
}

func TestEvidenceOccupancyRespectsDistinctAntiPigeonBuckets(t *testing.T) {
	candidates := fixtureCandidates(2)
	candidates[0].EvidenceGroupKey = "shared-lineage"
	candidates[1].EvidenceGroupKey = "shared-lineage"
	keys := map[string]string{candidates[0].Event.ID: "ap:left", candidates[1].Event.ID: "ap:right"}
	result := Select(candidates, keys, 2, 2, 1000, DefaultPolicy())
	if len(result.Candidates) != 2 || result.CorrelatedSuppressed != 0 {
		t.Fatalf("Anti-Pigeon buckets were collapsed: %+v", result)
	}
}

func fixtureCandidates(count int) []model.Candidate {
	result := make([]model.Candidate, count)
	for index := range result {
		probability := .7 - float64(index)*.001
		result[index] = model.Candidate{
			Event: model.Event{ID: fmt.Sprintf("event-%d", index), Content: fmt.Sprintf("content %d", index), What: model.Field{Value: fmt.Sprintf("event %d", index)}},
			Score: probability, EstimatedTokens: 1,
			Forecast: model.ForecastBundle{CorrectedLaw: model.BernoulliLaw{Useful: probability, NotUseful: 1 - probability}},
		}
	}
	return result
}
