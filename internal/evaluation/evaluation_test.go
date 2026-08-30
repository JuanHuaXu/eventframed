package evaluation_test

import (
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/evaluation"
)

func TestEvaluatePairedChronologicalConfirmation(t *testing.T) {
	dataset := mechanismDataset()
	report, err := evaluation.Evaluate(dataset)
	if err != nil {
		t.Fatal(err)
	}
	comparison := report.Comparisons["eventframe"]
	if report.Cases != 4 || report.Trajectories != 2 {
		t.Fatalf("report dimensions = %+v", report)
	}
	if comparison.BrierGain <= 0 || comparison.PriorityWeightedBrierGain <= 0 || comparison.ClusterBootstrapLower95 <= 0 {
		t.Fatalf("synthetic improvement was not recovered: %+v", comparison)
	}
	if comparison.RecallAt10Gain != 0 || comparison.RecallAt10ClusterLower95 != 0 || comparison.RecallAt10ClusterUpper95 != 0 {
		t.Fatalf("three-candidate fixture should have saturated Recall@10: %+v", comparison)
	}
	if report.Variants["eventframe"].RecallAt10 != 1 || report.Variants["eventframe"].HighPriorityMissRate10 != 0 {
		t.Fatalf("eventframe metrics = %+v", report.Variants["eventframe"])
	}
}

func TestValidateRejectsFutureVisibleCandidate(t *testing.T) {
	dataset := mechanismDataset()
	item := &dataset.Cases[0]
	forecast := item.Variants["eventframe"]
	forecast.Candidates[0].SourceAvailableAt = item.PredictedAt.Add(time.Nanosecond)
	item.Variants["eventframe"] = forecast
	if _, err := evaluation.Evaluate(dataset); err == nil || !strings.Contains(err.Error(), "future event") {
		t.Fatalf("future-visible candidate error = %v", err)
	}
}

func TestValidateRejectsIncomparableUniverse(t *testing.T) {
	dataset := mechanismDataset()
	item := &dataset.Cases[0]
	forecast := item.Variants["eventframe"]
	forecast.Candidates = forecast.Candidates[:2]
	item.Variants["eventframe"] = forecast
	if _, err := evaluation.Evaluate(dataset); err == nil || !strings.Contains(err.Error(), "fixed universe") {
		t.Fatalf("incomparable-universe error = %v", err)
	}
}

func TestValidateRejectsUnfrozenConfirmation(t *testing.T) {
	dataset := mechanismDataset()
	dataset.Cases[0].PredictedAt = dataset.PolicyFrozenAt.Add(time.Duration(dataset.EmbargoSeconds-1) * time.Second)
	if _, err := evaluation.Evaluate(dataset); err == nil || !strings.Contains(err.Error(), "embargo") {
		t.Fatalf("embargo error = %v", err)
	}
}

func TestPriorityConstraintBlocksAggregateGainWithHighPriorityMisses(t *testing.T) {
	frozen := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	dataset := evaluation.Dataset{
		SchemaVersion: evaluation.SchemaVersion, EvaluationBlock: "confirmation", BaselineVariant: "baseline",
		PolicyFrozenAt: frozen, EmbargoSeconds: 1, PriorityWeightScale: 4, BootstrapSamples: 500, BootstrapSeed: 91,
		PriorityConstraint: &evaluation.PriorityConstraint{MaxHighPriorityMissIncrease: 0, MaxOverallRecallLoss: .01, MinTrajectories: 3},
	}
	universe := make([]string, 12)
	for index := range universe {
		universe[index] = "event-" + string(rune('a'+index))
	}
	for trajectory := 0; trajectory < 3; trajectory++ {
		predicted := frozen.Add(time.Duration(trajectory+1) * time.Minute)
		baselineOrder := append([]string(nil), universe...)
		candidateOrder := append([]string(nil), universe[1:]...)
		candidateOrder = append(candidateOrder, universe[0])
		dataset.Cases = append(dataset.Cases, evaluation.Case{
			ID: "priority-" + string(rune('a'+trajectory)), TrajectoryID: "trajectory-" + string(rune('a'+trajectory)),
			PredictedAt: predicted, OutcomeAvailableAt: predicted.Add(time.Minute), Priority: 1,
			UniverseEventIDs: universe, RelevantEventIDs: []string{universe[0]},
			Variants: map[string]evaluation.VariantForecast{
				"baseline":  priorityForecast(predicted, baselineOrder, universe[0], .6, .5),
				"candidate": priorityForecast(predicted, candidateOrder, universe[0], .9, .05),
			},
		})
	}
	report, err := evaluation.Evaluate(dataset)
	if err != nil {
		t.Fatal(err)
	}
	comparison := report.Comparisons["candidate"]
	if comparison.PriorityWeightedBrierGain <= 0 || !comparison.PriorityConstraintEvaluated || comparison.PriorityConstraintPassed {
		t.Fatalf("priority gate failed to block aggregate-only improvement: %+v", comparison)
	}
}

func TestPriorityConstraintRemainsUnevaluatedWithoutHighPriorityCases(t *testing.T) {
	dataset := mechanismDataset()
	dataset.PriorityConstraint = &evaluation.PriorityConstraint{MaxHighPriorityMissIncrease: 0, MaxOverallRecallLoss: .01, MinTrajectories: 2}
	for index := range dataset.Cases {
		dataset.Cases[index].Priority = .2
	}
	report, err := evaluation.Evaluate(dataset)
	if err != nil {
		t.Fatal(err)
	}
	if report.Comparisons["eventframe"].PriorityConstraintEvaluated {
		t.Fatal("priority constraint was evaluated without high-priority cases")
	}
}

func TestRepresentationAblationRequiresMatchedContractsAndEvidence(t *testing.T) {
	dataset := mechanismDataset()
	for index := range dataset.Cases {
		base := dataset.Cases[index].Variants["baseline"]
		dataset.Cases[index].Variants = map[string]evaluation.VariantForecast{"raw": base, "structured": base, "combined": base, "shuffled": base}
	}
	dataset.BaselineVariant = "raw"
	dataset.RepresentationAblation = &evaluation.RepresentationAblationContract{InterpretabilityRaters: 2, RatingsBlinded: true, Variants: map[string]evaluation.RepresentationVariantContract{}}
	for _, role := range []string{"raw", "structured", "combined", "shuffled"} {
		dataset.RepresentationAblation.Variants[role] = evaluation.RepresentationVariantContract{Role: role, SourceDigest: "same-source", EmbeddingModel: "model-v1", EmbeddingBudget: 768, RankingContract: "libravdb-rank-v1"}
	}
	report, err := evaluation.Evaluate(dataset)
	if err != nil {
		t.Fatal(err)
	}
	if report.RepresentationAblationReady || report.InterpretabilityReady {
		t.Fatalf("two trajectories were incorrectly promoted: %+v", report)
	}
	mismatched := *dataset.RepresentationAblation
	mismatched.Variants = make(map[string]evaluation.RepresentationVariantContract, len(dataset.RepresentationAblation.Variants))
	for name, variant := range dataset.RepresentationAblation.Variants {
		mismatched.Variants[name] = variant
	}
	variant := mismatched.Variants["structured"]
	variant.EmbeddingModel = "different-model"
	mismatched.Variants["structured"] = variant
	dataset.RepresentationAblation = &mismatched
	if _, err := evaluation.Evaluate(dataset); err == nil {
		t.Fatal("mismatched representation ablation was accepted")
	}
}

func priorityForecast(predicted time.Time, order []string, relevant string, relevantProbability, otherProbability float64) evaluation.VariantForecast {
	candidates := make([]evaluation.CandidateForecast, 0, len(order))
	for _, eventID := range order {
		probability := otherProbability
		if eventID == relevant {
			probability = relevantProbability
		}
		candidates = append(candidates, evaluation.CandidateForecast{EventID: eventID, SourceAvailableAt: predicted.Add(-time.Minute), Probability: probability, RankScore: probability, Nominated: true})
	}
	return evaluation.VariantForecast{StateAsOf: predicted, Candidates: candidates}
}

func mechanismDataset() evaluation.Dataset {
	frozen := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	dataset := evaluation.Dataset{
		SchemaVersion: evaluation.SchemaVersion, EvaluationBlock: "confirmation", BaselineVariant: "baseline",
		PolicyFrozenAt: frozen, EmbargoSeconds: 60, PriorityWeightScale: 4, BootstrapSamples: 500, BootstrapSeed: 73,
	}
	for index := 0; index < 4; index++ {
		predicted := frozen.Add(time.Duration(index+2) * time.Minute)
		relevant := "a"
		if index%2 == 1 {
			relevant = "c"
		}
		dataset.Cases = append(dataset.Cases, evaluation.Case{
			ID: "case-" + string(rune('a'+index)), TrajectoryID: "trajectory-" + string(rune('a'+index/2)),
			PredictedAt: predicted, OutcomeAvailableAt: predicted.Add(time.Minute), Priority: .4 + float64(index)*.2,
			UniverseEventIDs: []string{"a", "b", "c"}, RelevantEventIDs: []string{relevant},
			Variants: map[string]evaluation.VariantForecast{
				"baseline":   forecast(predicted, []string{"b", "a", "c"}, relevant, .4, false),
				"eventframe": forecast(predicted, []string{relevant, "b", other(relevant)}, relevant, .8, true),
			},
		})
	}
	return dataset
}

func forecast(predicted time.Time, order []string, relevant string, relevantProbability float64, activated bool) evaluation.VariantForecast {
	candidates := make([]evaluation.CandidateForecast, 0, len(order))
	for _, id := range order {
		probability := .15
		if id == relevant {
			probability = relevantProbability
		}
		candidates = append(candidates, evaluation.CandidateForecast{
			EventID: id, SourceAvailableAt: predicted.Add(-time.Minute), Probability: probability,
			Nominated: true, Activated: activated && id == relevant,
		})
	}
	return evaluation.VariantForecast{StateAsOf: predicted, Candidates: candidates}
}

func other(relevant string) string {
	if relevant == "a" {
		return "c"
	}
	return "a"
}
