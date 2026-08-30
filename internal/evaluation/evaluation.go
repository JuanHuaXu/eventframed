package evaluation

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

const SchemaVersion = "eventframe.evaluation.v1"

type Dataset struct {
	SchemaVersion          string                          `json:"schema_version"`
	EvaluationBlock        string                          `json:"evaluation_block"`
	BaselineVariant        string                          `json:"baseline_variant"`
	PolicyFrozenAt         time.Time                       `json:"policy_frozen_at"`
	EmbargoSeconds         int64                           `json:"embargo_seconds"`
	PriorityWeightScale    float64                         `json:"priority_weight_scale"`
	BootstrapSamples       int                             `json:"bootstrap_samples"`
	BootstrapSeed          int64                           `json:"bootstrap_seed"`
	PriorityConstraint     *PriorityConstraint             `json:"priority_constraint,omitempty"`
	RepresentationAblation *RepresentationAblationContract `json:"representation_ablation,omitempty"`
	Cases                  []Case                          `json:"cases"`
}

type RepresentationAblationContract struct {
	Variants               map[string]RepresentationVariantContract `json:"variants"`
	InterpretabilityRaters int                                      `json:"interpretability_raters"`
	RatingsBlinded         bool                                     `json:"ratings_blinded"`
}

type RepresentationVariantContract struct {
	Role            string `json:"role"`
	SourceDigest    string `json:"source_digest"`
	EmbeddingModel  string `json:"embedding_model"`
	EmbeddingBudget int    `json:"embedding_budget"`
	RankingContract string `json:"ranking_contract"`
}

type PriorityConstraint struct {
	MaxHighPriorityMissIncrease float64 `json:"max_high_priority_miss_increase"`
	MaxOverallRecallLoss        float64 `json:"max_overall_recall_loss"`
	MinTrajectories             int     `json:"min_trajectories"`
}

type Case struct {
	ID                 string                     `json:"id"`
	TrajectoryID       string                     `json:"trajectory_id"`
	PredictedAt        time.Time                  `json:"predicted_at"`
	OutcomeAvailableAt time.Time                  `json:"outcome_available_at"`
	Priority           float64                    `json:"priority"`
	UniverseEventIDs   []string                   `json:"universe_event_ids"`
	RelevantEventIDs   []string                   `json:"relevant_event_ids"`
	Variants           map[string]VariantForecast `json:"variants"`
}

type VariantForecast struct {
	StateAsOf        time.Time           `json:"state_as_of"`
	Candidates       []CandidateForecast `json:"candidates"`
	PackedCount      int                 `json:"packed_count"`
	UsedTokens       int                 `json:"used_tokens"`
	AdaptiveExpanded bool                `json:"adaptive_expanded"`
}

type CandidateForecast struct {
	EventID           string    `json:"event_id"`
	SourceAvailableAt time.Time `json:"source_available_at"`
	Probability       float64   `json:"probability"`
	RankScore         float64   `json:"rank_score"`
	Nominated         bool      `json:"nominated"`
	Activated         bool      `json:"activated"`
	Packed            bool      `json:"packed"`
}

type Report struct {
	SchemaVersion               string                `json:"schema_version"`
	EvaluationBlock             string                `json:"evaluation_block"`
	Cases                       int                   `json:"cases"`
	Trajectories                int                   `json:"trajectories"`
	Variants                    map[string]Metrics    `json:"variants"`
	Comparisons                 map[string]Comparison `json:"comparisons"`
	RepresentationAblationReady bool                  `json:"representation_ablation_ready"`
	InterpretabilityReady       bool                  `json:"interpretability_ready"`
}

type Metrics struct {
	Brier                    float64 `json:"brier"`
	ExpectedCalibrationError float64 `json:"expected_calibration_error"`
	RecallAt10               float64 `json:"recall_at_10"`
	RecallAt50               float64 `json:"recall_at_50"`
	MeanReciprocalRank       float64 `json:"mean_reciprocal_rank"`
	PriorityWeightedBrier    float64 `json:"priority_weighted_brier"`
	PriorityWeightedRecall10 float64 `json:"priority_weighted_recall_at_10"`
	HighPriorityMissRate10   float64 `json:"high_priority_miss_rate_at_10"`
	HighPriorityCases        int     `json:"high_priority_cases"`
	NominationRecall         float64 `json:"nomination_recall"`
	ActivationRate           float64 `json:"activation_rate"`
	PackedRecall             float64 `json:"packed_recall"`
	MeanPackedCount          float64 `json:"mean_packed_count"`
	MeanUsedTokens           float64 `json:"mean_used_tokens"`
	AdaptiveExpansionRate    float64 `json:"adaptive_expansion_rate"`
}

type Comparison struct {
	Baseline                        string  `json:"baseline"`
	Candidate                       string  `json:"candidate"`
	ClusterInferenceValid           bool    `json:"cluster_inference_valid"`
	BrierGain                       float64 `json:"brier_gain"`
	PriorityWeightedBrierGain       float64 `json:"priority_weighted_brier_gain"`
	ClusterBootstrapLower95         float64 `json:"cluster_bootstrap_lower_95"`
	ClusterBootstrapUpper95         float64 `json:"cluster_bootstrap_upper_95"`
	RecallAt10Gain                  float64 `json:"recall_at_10_gain"`
	RecallAt10ClusterLower95        float64 `json:"recall_at_10_cluster_lower_95"`
	RecallAt10ClusterUpper95        float64 `json:"recall_at_10_cluster_upper_95"`
	PackedRecallGain                float64 `json:"packed_recall_gain"`
	PackedRecallClusterLower95      float64 `json:"packed_recall_cluster_lower_95"`
	PackedRecallClusterUpper95      float64 `json:"packed_recall_cluster_upper_95"`
	MeanUsedTokensDelta             float64 `json:"mean_used_tokens_delta"`
	HighPriorityMissIncrease        float64 `json:"high_priority_miss_increase"`
	HighPriorityMissIncreaseUpper95 float64 `json:"high_priority_miss_increase_upper_95"`
	PriorityConstraintEvaluated     bool    `json:"priority_constraint_evaluated"`
	PriorityConstraintPassed        bool    `json:"priority_constraint_passed"`
}

type caseScore struct {
	trajectory       string
	priority         float64
	brier            float64
	recall10         float64
	packedRecall     float64
	highPriority     bool
	highPriorityMiss float64
}

func Evaluate(dataset Dataset) (Report, error) {
	if err := Validate(dataset); err != nil {
		return Report{}, err
	}
	variantNames := sortedVariantNames(dataset.Cases[0].Variants)
	scores := make(map[string][]caseScore, len(variantNames))
	report := Report{
		SchemaVersion: SchemaVersion, EvaluationBlock: dataset.EvaluationBlock,
		Cases: len(dataset.Cases), Variants: make(map[string]Metrics, len(variantNames)),
		Comparisons: make(map[string]Comparison, len(variantNames)-1),
	}
	trajectories := make(map[string]struct{})
	for _, item := range dataset.Cases {
		trajectories[item.TrajectoryID] = struct{}{}
	}
	report.Trajectories = len(trajectories)
	if dataset.RepresentationAblation != nil {
		report.RepresentationAblationReady = report.Trajectories >= 3
		report.InterpretabilityReady = report.RepresentationAblationReady && dataset.RepresentationAblation.RatingsBlinded && dataset.RepresentationAblation.InterpretabilityRaters >= 2
	}
	for _, name := range variantNames {
		metrics, variantScores := evaluateVariant(dataset, name)
		report.Variants[name], scores[name] = metrics, variantScores
	}
	baseline := report.Variants[dataset.BaselineVariant]
	for _, name := range variantNames {
		if name == dataset.BaselineVariant {
			continue
		}
		candidate := report.Variants[name]
		lower, upper, recallLower, recallUpper, packedLower, packedUpper, highMissUpper := clusteredGainIntervals(scores[dataset.BaselineVariant], scores[name], dataset.PriorityWeightScale, dataset.BootstrapSamples, dataset.BootstrapSeed)
		comparison := Comparison{
			Baseline: dataset.BaselineVariant, Candidate: name,
			ClusterInferenceValid:     report.Trajectories >= 2,
			BrierGain:                 baseline.Brier - candidate.Brier,
			PriorityWeightedBrierGain: baseline.PriorityWeightedBrier - candidate.PriorityWeightedBrier,
			ClusterBootstrapLower95:   lower, ClusterBootstrapUpper95: upper,
			RecallAt10Gain:           candidate.RecallAt10 - baseline.RecallAt10,
			RecallAt10ClusterLower95: recallLower, RecallAt10ClusterUpper95: recallUpper,
			PackedRecallGain:           candidate.PackedRecall - baseline.PackedRecall,
			PackedRecallClusterLower95: packedLower, PackedRecallClusterUpper95: packedUpper,
			MeanUsedTokensDelta:             candidate.MeanUsedTokens - baseline.MeanUsedTokens,
			HighPriorityMissIncrease:        candidate.HighPriorityMissRate10 - baseline.HighPriorityMissRate10,
			HighPriorityMissIncreaseUpper95: highMissUpper,
		}
		if constraint := dataset.PriorityConstraint; constraint != nil && report.Trajectories >= constraint.MinTrajectories && baseline.HighPriorityCases > 0 && candidate.HighPriorityCases > 0 {
			comparison.PriorityConstraintEvaluated = true
			comparison.PriorityConstraintPassed = highMissUpper <= constraint.MaxHighPriorityMissIncrease && comparison.RecallAt10Gain >= -constraint.MaxOverallRecallLoss
		}
		report.Comparisons[name] = comparison
	}
	return report, nil
}

func Validate(dataset Dataset) error {
	if dataset.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema_version must be %q", SchemaVersion)
	}
	if dataset.EvaluationBlock != "design" && dataset.EvaluationBlock != "confirmation" {
		return errors.New("evaluation_block must be design or confirmation")
	}
	if dataset.BaselineVariant == "" || dataset.PolicyFrozenAt.IsZero() || dataset.EmbargoSeconds < 0 || dataset.PriorityWeightScale < 0 || math.IsNaN(dataset.PriorityWeightScale) || math.IsInf(dataset.PriorityWeightScale, 0) {
		return errors.New("baseline, freeze time, non-negative embargo, and finite priority scale are required")
	}
	if dataset.BootstrapSamples <= 0 {
		return errors.New("bootstrap_samples must be positive")
	}
	if constraint := dataset.PriorityConstraint; constraint != nil {
		if constraint.MinTrajectories < 2 || constraint.MaxHighPriorityMissIncrease < 0 || constraint.MaxOverallRecallLoss < 0 || !finite(constraint.MaxHighPriorityMissIncrease) || !finite(constraint.MaxOverallRecallLoss) {
			return errors.New("priority constraint requires finite non-negative ceilings and at least two trajectories")
		}
	}
	if len(dataset.Cases) == 0 {
		return errors.New("evaluation dataset is empty")
	}
	if err := validateRepresentationAblation(dataset); err != nil {
		return err
	}
	freezeBoundary := dataset.PolicyFrozenAt.Add(time.Duration(dataset.EmbargoSeconds) * time.Second)
	caseIDs := make(map[string]struct{}, len(dataset.Cases))
	var variantNames []string
	lastPrediction := make(map[string]time.Time)
	for index, item := range dataset.Cases {
		if item.ID == "" || item.TrajectoryID == "" || item.PredictedAt.IsZero() || item.OutcomeAvailableAt.IsZero() {
			return fmt.Errorf("case %d lacks identity or times", index)
		}
		if _, duplicate := caseIDs[item.ID]; duplicate {
			return fmt.Errorf("duplicate case id %q", item.ID)
		}
		caseIDs[item.ID] = struct{}{}
		if item.PredictedAt.Before(freezeBoundary) {
			return fmt.Errorf("case %q violates the frozen-policy embargo", item.ID)
		}
		if !item.OutcomeAvailableAt.After(item.PredictedAt) {
			return fmt.Errorf("case %q outcome is not strictly post-prediction", item.ID)
		}
		if previous := lastPrediction[item.TrajectoryID]; !previous.IsZero() && !item.PredictedAt.After(previous) {
			return fmt.Errorf("trajectory %q is not in strict chronological order", item.TrajectoryID)
		}
		lastPrediction[item.TrajectoryID] = item.PredictedAt
		if item.Priority < 0 || item.Priority > 1 || len(item.UniverseEventIDs) == 0 || len(item.RelevantEventIDs) == 0 {
			return fmt.Errorf("case %q has invalid priority or empty target sets", item.ID)
		}
		universe, err := uniqueSet(item.UniverseEventIDs)
		if err != nil {
			return fmt.Errorf("case %q universe: %w", item.ID, err)
		}
		for _, relevant := range item.RelevantEventIDs {
			if _, ok := universe[relevant]; !ok {
				return fmt.Errorf("case %q relevant event %q is outside the universe", item.ID, relevant)
			}
		}
		currentNames := sortedVariantNames(item.Variants)
		if index == 0 {
			variantNames = currentNames
			if len(variantNames) < 2 {
				return errors.New("evaluation requires a baseline and at least one candidate variant")
			}
			if _, ok := item.Variants[dataset.BaselineVariant]; !ok {
				return errors.New("baseline variant is absent")
			}
		} else if fmt.Sprint(currentNames) != fmt.Sprint(variantNames) {
			return fmt.Errorf("case %q has a different variant family", item.ID)
		}
		for name, forecast := range item.Variants {
			if forecast.StateAsOf.IsZero() || forecast.StateAsOf.After(item.PredictedAt) {
				return fmt.Errorf("case %q variant %q uses future state", item.ID, name)
			}
			if len(forecast.Candidates) != len(universe) {
				return fmt.Errorf("case %q variant %q does not cover the fixed universe", item.ID, name)
			}
			seen := make(map[string]struct{}, len(forecast.Candidates))
			packedCount := 0
			for _, candidate := range forecast.Candidates {
				if _, ok := universe[candidate.EventID]; !ok {
					return fmt.Errorf("case %q variant %q contains event outside the universe", item.ID, name)
				}
				if _, duplicate := seen[candidate.EventID]; duplicate {
					return fmt.Errorf("case %q variant %q duplicates event %q", item.ID, name, candidate.EventID)
				}
				seen[candidate.EventID] = struct{}{}
				if candidate.SourceAvailableAt.IsZero() || candidate.SourceAvailableAt.After(item.PredictedAt) {
					return fmt.Errorf("case %q variant %q exposes future event %q", item.ID, name, candidate.EventID)
				}
				if candidate.Probability < 0 || candidate.Probability > 1 || math.IsNaN(candidate.Probability) || math.IsInf(candidate.Probability, 0) {
					return fmt.Errorf("case %q variant %q has invalid probability", item.ID, name)
				}
				if candidate.RankScore < 0 || candidate.RankScore > 1 || math.IsNaN(candidate.RankScore) || math.IsInf(candidate.RankScore, 0) {
					return fmt.Errorf("case %q variant %q has invalid rank score", item.ID, name)
				}
				if candidate.Activated && !candidate.Nominated {
					return fmt.Errorf("case %q variant %q activates an unnominated event", item.ID, name)
				}
				if candidate.Packed {
					packedCount++
				}
			}
			if packedCount != forecast.PackedCount || forecast.UsedTokens < 0 {
				return fmt.Errorf("case %q variant %q has inconsistent packing metadata", item.ID, name)
			}
		}
	}
	return nil
}

func validateRepresentationAblation(dataset Dataset) error {
	contract := dataset.RepresentationAblation
	if contract == nil {
		return nil
	}
	if len(contract.Variants) != 4 || contract.InterpretabilityRaters < 0 {
		return errors.New("representation ablation requires exactly four variants and non-negative rater count")
	}
	roles := make(map[string]struct{}, 4)
	var reference RepresentationVariantContract
	first := true
	for name, variant := range contract.Variants {
		if _, exists := dataset.Cases[0].Variants[name]; !exists {
			return fmt.Errorf("representation ablation variant %q is absent", name)
		}
		switch variant.Role {
		case "raw", "structured", "combined", "shuffled":
		default:
			return fmt.Errorf("representation ablation variant %q has invalid role", name)
		}
		if _, duplicate := roles[variant.Role]; duplicate {
			return errors.New("representation ablation roles must be unique")
		}
		roles[variant.Role] = struct{}{}
		if variant.SourceDigest == "" || variant.EmbeddingModel == "" || variant.EmbeddingBudget <= 0 || variant.RankingContract == "" {
			return errors.New("representation ablation variants require source, embedding, budget, and ranking contracts")
		}
		if first {
			reference, first = variant, false
			continue
		}
		if variant.SourceDigest != reference.SourceDigest || variant.EmbeddingModel != reference.EmbeddingModel || variant.EmbeddingBudget != reference.EmbeddingBudget || variant.RankingContract != reference.RankingContract {
			return errors.New("representation ablation variants are not matched")
		}
	}
	return nil
}

func evaluateVariant(dataset Dataset, name string) (Metrics, []caseScore) {
	var metrics Metrics
	var totalPairs, relevantTotal, nominatedRelevant, activatedTotal int
	var brierSum, reciprocalRank, recall10Sum, recall50Sum, packedRecallSum, packedCountSum, usedTokenSum, expansionSum, weightedBrier, weightedRecall, weightSum float64
	var highPriorityCases, highPriorityMisses int
	bins := make([]struct {
		count                int
		probability, outcome float64
	}, 10)
	scores := make([]caseScore, 0, len(dataset.Cases))
	for _, item := range dataset.Cases {
		forecast := item.Variants[name]
		relevant, _ := uniqueSet(item.RelevantEventIDs)
		caseBrier := 0.0
		firstRank := 0
		hits10, hits50 := 0, 0
		packedHits := 0
		for rank, candidate := range forecast.Candidates {
			outcome := 0.0
			if _, ok := relevant[candidate.EventID]; ok {
				outcome = 1
				if firstRank == 0 {
					firstRank = rank + 1
				}
				if rank < 10 {
					hits10++
				}
				if rank < 50 {
					hits50++
				}
				if candidate.Nominated {
					nominatedRelevant++
				}
				if candidate.Packed {
					packedHits++
				}
			}
			if candidate.Activated {
				activatedTotal++
			}
			error := candidate.Probability - outcome
			caseBrier += error * error
			bin := min(int(candidate.Probability*10), 9)
			bins[bin].count++
			bins[bin].probability += candidate.Probability
			bins[bin].outcome += outcome
		}
		caseBrier /= float64(len(forecast.Candidates))
		recall10 := float64(hits10) / float64(len(relevant))
		recall50 := float64(hits50) / float64(len(relevant))
		packedRecall := float64(packedHits) / float64(len(relevant))
		weight := 1 + dataset.PriorityWeightScale*item.Priority
		brierSum += caseBrier
		recall10Sum += recall10
		recall50Sum += recall50
		packedRecallSum += packedRecall
		packedCountSum += float64(forecast.PackedCount)
		usedTokenSum += float64(forecast.UsedTokens)
		if forecast.AdaptiveExpanded {
			expansionSum++
		}
		weightedBrier += weight * caseBrier
		weightedRecall += weight * recall10
		weightSum += weight
		if firstRank > 0 {
			reciprocalRank += 1 / float64(firstRank)
		}
		if item.Priority >= .8 {
			highPriorityCases++
			if hits10 == 0 {
				highPriorityMisses++
			}
		}
		totalPairs += len(forecast.Candidates)
		relevantTotal += len(relevant)
		highPriority := item.Priority >= .8
		highPriorityMiss := 0.0
		if highPriority && hits10 == 0 {
			highPriorityMiss = 1
		}
		scores = append(scores, caseScore{trajectory: item.TrajectoryID, priority: item.Priority, brier: caseBrier, recall10: recall10, packedRecall: packedRecall, highPriority: highPriority, highPriorityMiss: highPriorityMiss})
	}
	metrics.Brier = brierSum / float64(len(dataset.Cases))
	metrics.RecallAt10 = recall10Sum / float64(len(dataset.Cases))
	metrics.RecallAt50 = recall50Sum / float64(len(dataset.Cases))
	metrics.MeanReciprocalRank = reciprocalRank / float64(len(dataset.Cases))
	metrics.PriorityWeightedBrier = weightedBrier / weightSum
	metrics.PriorityWeightedRecall10 = weightedRecall / weightSum
	metrics.NominationRecall = float64(nominatedRelevant) / float64(relevantTotal)
	metrics.ActivationRate = float64(activatedTotal) / float64(totalPairs)
	metrics.PackedRecall = packedRecallSum / float64(len(dataset.Cases))
	metrics.MeanPackedCount = packedCountSum / float64(len(dataset.Cases))
	metrics.MeanUsedTokens = usedTokenSum / float64(len(dataset.Cases))
	metrics.AdaptiveExpansionRate = expansionSum / float64(len(dataset.Cases))
	if highPriorityCases > 0 {
		metrics.HighPriorityMissRate10 = float64(highPriorityMisses) / float64(highPriorityCases)
	}
	metrics.HighPriorityCases = highPriorityCases
	for _, bin := range bins {
		if bin.count == 0 {
			continue
		}
		metrics.ExpectedCalibrationError += float64(bin.count) / float64(totalPairs) * math.Abs(bin.probability/float64(bin.count)-bin.outcome/float64(bin.count))
	}
	return metrics, scores
}

func clusteredGainIntervals(baseline, candidate []caseScore, scale float64, samples int, seed int64) (float64, float64, float64, float64, float64, float64, float64) {
	byTrajectory := make(map[string][]int)
	for index := range baseline {
		byTrajectory[baseline[index].trajectory] = append(byTrajectory[baseline[index].trajectory], index)
	}
	keys := make([]string, 0, len(byTrajectory))
	for key := range byTrajectory {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	random := rand.New(rand.NewSource(seed))
	brierGains := make([]float64, samples)
	recallGains := make([]float64, samples)
	packedGains := make([]float64, samples)
	highMissIncreases := make([]float64, samples)
	for sample := range brierGains {
		var brierGain, weightSum, recallGain, packedGain, highMissIncrease float64
		var caseCount, highPriorityCount int
		for range keys {
			key := keys[random.Intn(len(keys))]
			for _, index := range byTrajectory[key] {
				weight := 1 + scale*baseline[index].priority
				brierGain += weight * (baseline[index].brier - candidate[index].brier)
				weightSum += weight
				recallGain += candidate[index].recall10 - baseline[index].recall10
				packedGain += candidate[index].packedRecall - baseline[index].packedRecall
				if baseline[index].highPriority {
					highMissIncrease += candidate[index].highPriorityMiss - baseline[index].highPriorityMiss
					highPriorityCount++
				}
				caseCount++
			}
		}
		brierGains[sample] = brierGain / weightSum
		recallGains[sample] = recallGain / float64(caseCount)
		packedGains[sample] = packedGain / float64(caseCount)
		if highPriorityCount > 0 {
			highMissIncreases[sample] = highMissIncrease / float64(highPriorityCount)
		}
	}
	sort.Float64s(brierGains)
	sort.Float64s(recallGains)
	sort.Float64s(packedGains)
	sort.Float64s(highMissIncreases)
	return brierGains[percentile(len(brierGains), .025)], brierGains[percentile(len(brierGains), .975)], recallGains[percentile(len(recallGains), .025)], recallGains[percentile(len(recallGains), .975)], packedGains[percentile(len(packedGains), .025)], packedGains[percentile(len(packedGains), .975)], highMissIncreases[percentile(len(highMissIncreases), .95)]
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func percentile(count int, probability float64) int {
	return min(max(int(float64(count-1)*probability), 0), count-1)
}

func uniqueSet(values []string) (map[string]struct{}, error) {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			return nil, errors.New("empty event id")
		}
		if _, duplicate := out[value]; duplicate {
			return nil, fmt.Errorf("duplicate event id %q", value)
		}
		out[value] = struct{}{}
	}
	return out, nil
}

func sortedVariantNames(variants map[string]VariantForecast) []string {
	names := make([]string, 0, len(variants))
	for name := range variants {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
