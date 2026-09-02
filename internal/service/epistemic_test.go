package service

import (
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

func TestNewPreservesConfiguredEvidenceOccupancyWithDefaultPackingPolicy(t *testing.T) {
	embedder, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := New(memorystore.New(), embedder, Config{
		DefaultRecallK: 10, DefaultPackK: 5, DefaultTokenBudget: 100,
		PackingPolicy: packing.Policy{EvidenceOccupancyLimit: 2, EvidenceSimilarity: .9},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close()
	if runtime.config.PackingPolicy.EvidenceOccupancyLimit != 2 || runtime.config.PackingPolicy.EvidenceSimilarity != .9 {
		t.Fatalf("evidence policy was replaced by defaults: %+v", runtime.config.PackingPolicy)
	}
}

func TestNewRejectsFailOpenEvidencePolicies(t *testing.T) {
	embedder, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	for name, policy := range map[string]packing.Policy{
		"disabled occupancy": {MaxPack: 5, EvidenceOccupancyLimit: -1, EvidenceSimilarity: .85},
		"NaN similarity":     {MaxPack: 5, EvidenceOccupancyLimit: 1, EvidenceSimilarity: math.NaN()},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := New(memorystore.New(), embedder, Config{DefaultRecallK: 10, DefaultPackK: 5, DefaultTokenBudget: 100, PackingPolicy: policy}); err == nil {
				t.Fatal("fail-open evidence policy was accepted")
			}
		})
	}
}

func TestMarkCorrelatedBayesianDecisionsPreservesUsefulnessUpdates(t *testing.T) {
	report := model.BayesianShadowReport{Activated: 2, DeepReviewed: 2, Decisions: []model.BayesianDecision{
		{EventID: "best", EvidenceGroupKey: "same", PosteriorKey: "best", Activated: true, CheapUpdate: true, DeepReview: true, ActivationProbability: 1, TotalSelectionProbabilityLowerBound: .5},
		{EventID: "copy", EvidenceGroupKey: "same", PosteriorKey: "copy", Activated: true, CheapUpdate: true, DeepReview: true, ActivationProbability: 1, TotalSelectionProbabilityLowerBound: .5},
	}}
	markCorrelatedBayesianDecisions(&report)
	if report.Activated != 2 || report.DeepReviewed != 2 || !report.Decisions[1].CorrelatedSuppressed || !report.Decisions[1].Activated {
		t.Fatalf("correlated evidence marker changed exhaustive usefulness updates: %+v", report)
	}
}

func TestSuppressCorrelatedBayesianDecisionsDefersToAntiPigeonSplit(t *testing.T) {
	report := model.BayesianShadowReport{Activated: 2, DeepReviewed: 2, Decisions: []model.BayesianDecision{
		{EventID: "left", EvidenceGroupKey: "same", PosteriorKey: "ap:left", Activated: true, CheapUpdate: true, DeepReview: true},
		{EventID: "right", EvidenceGroupKey: "same", PosteriorKey: "ap:right", Activated: true, CheapUpdate: true, DeepReview: true},
	}}
	markCorrelatedBayesianDecisions(&report)
	if report.Activated != 2 || report.Decisions[1].CorrelatedSuppressed {
		t.Fatalf("Anti-Pigeon split lost authority: %+v", report)
	}
}

func TestCorrelatedDecisionCannotClaimIndependentSelectedOutcome(t *testing.T) {
	decision := model.BayesianDecision{Activated: true, CorrelatedSuppressed: true, TotalSelectionProbabilityLowerBound: .5}
	report := model.BayesianShadowReport{SelectionSupportCertified: true}
	request := model.BayesianOutcomeRequest{Source: model.OutcomeSelected, InclusionProbability: .5}
	if _, err := bayesianOutcomeWeight(request, decision, report); err == nil {
		t.Fatal("correlated record was accepted as independent selective evidence")
	}
	request.Source = model.OutcomeFullStream
	request.InclusionProbability = 1
	if weight, err := bayesianOutcomeWeight(request, decision, report); err != nil || weight != 1 {
		t.Fatalf("exhaustive usefulness outcome was incorrectly suppressed: weight=%v err=%v", weight, err)
	}
}
