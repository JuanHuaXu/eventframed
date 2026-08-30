package claimsexperiment

import (
	"context"
	"fmt"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func RunAntiPigeonGranularity(ctx context.Context) (AntiPigeonReport, error) {
	const training, evaluation = 30, 20
	runners := make([]antiPigeonRunner, 0, 3)
	for _, name := range []string{"separate", "naive_shared", "oracle_ap"} {
		runner, err := newAntiPigeonRunner(ctx, name)
		if err != nil {
			return AntiPigeonReport{}, err
		}
		runners = append(runners, runner)
	}
	report := AntiPigeonReport{TrainingTurns: training, EvaluationTurns: evaluation, Variants: make(map[string]AntiPigeonVariant)}
	for _, runner := range runners {
		variant, err := runAntiPigeonVariant(ctx, runner, training, evaluation)
		if err != nil {
			return AntiPigeonReport{}, err
		}
		report.Variants[runner.name] = variant
	}
	return report, nil
}

func runAntiPigeonVariant(ctx context.Context, runner antiPigeonRunner, training, evaluation int) (AntiPigeonVariant, error) {
	var loss float64
	var last model.ContextPacket
	for turn := 0; turn < training+evaluation; turn++ {
		asOf := experimentAnchor.Add(time.Duration(turn) * time.Minute)
		packet, err := runner.runtime.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "anti-pigeon-experiment", SessionID: "trajectory", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: "feature-hash-v1:d8:repr=eventframe-5w1h-v1", AsOf: asOf, RecallK: 4, PackK: 4, TokenBudget: 400})
		if err != nil {
			return AntiPigeonVariant{}, err
		}
		if turn >= training {
			for _, candidate := range packet.Candidates {
				target := 0.0
				if isPositive(candidate.Event.ID) {
					target = 1
				}
				loss += square(target - candidate.Forecast.CorrectedLaw.Useful)
			}
		}
		for _, candidate := range packet.Candidates {
			if err := submitOutcome(ctx, runner.runtime, packet, candidate.Event.ID, isPositive(candidate.Event.ID), fmt.Sprintf("%s-%03d-%s", runner.name, turn, candidate.Event.ID), asOf.Add(time.Second)); err != nil {
				return AntiPigeonVariant{}, err
			}
		}
		last = packet
	}
	keys := make(map[string]struct{})
	eventKeys := make(map[string]string)
	for _, decision := range last.BayesianShadow.Decisions {
		keys[decision.PosteriorKey] = struct{}{}
		eventKeys[decision.EventID] = decision.PosteriorKey
	}
	return AntiPigeonVariant{Brier: loss / float64(evaluation*len(antiPigeonIDs)), FalseMergeRate: falseMergeRate(eventKeys), PosteriorKeys: len(keys)}, nil
}

func falseMergeRate(keys map[string]string) float64 {
	falseMerges, crossPairs := 0, 0
	for _, positive := range antiPigeonIDs[:2] {
		for _, negative := range antiPigeonIDs[2:] {
			crossPairs++
			if keys[positive] == keys[negative] {
				falseMerges++
			}
		}
	}
	return float64(falseMerges) / float64(crossPairs)
}

func isPositive(id string) bool { return len(id) >= 8 && id[:8] == "positive" }
