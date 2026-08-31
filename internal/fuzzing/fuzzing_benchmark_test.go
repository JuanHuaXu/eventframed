package fuzzing_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/fuzzing"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

func BenchmarkEvaluateEmbeddingNomination50x64x256(b *testing.B) {
	embedder, _ := embed.NewHashEmbedder(256)
	events := make([]model.Event, 50)
	eventIDs := make([]string, 50)
	for index := range events {
		events[index] = testEvent(fmt.Sprintf("event-%02d", index), fmt.Sprintf("public fact %d", index), testEvents()[0].AvailableAt)
		eventIDs[index] = events[index].ID
	}
	perturbations := make([]model.FuzzPerturbation, 64)
	for index := range perturbations {
		perturbations[index] = model.FuzzPerturbation{
			ID: fmt.Sprintf("trial-%02d", index), PropertyID: "context-envelope", EventID: events[index%len(events)].ID,
			ValidityRuleID: "benchmark-valid-v1", ValidationKind: model.FuzzValidationDeclaredRelocation,
			Replacements: map[model.FuzzField]model.Field{
				model.FuzzWho: synthetic(fmt.Sprintf("alternate participant %d", index), "benchmark declaration"),
			},
		}
	}
	request := model.FuzzSensitivityRequest{
		TenantID: "public", Query: "correct public fact", AsOf: testEvents()[0].AvailableAt.AddDate(0, 0, 1),
		EventIDs: eventIDs, Perturbations: perturbations, StabilityThreshold: .05,
		RequiredStableProbability: .9, ConfidenceLevel: .95, MinTrials: 24,
	}
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		predictor, err := fuzzing.NewEmbeddingNominationPredictor(context.Background(), embedder, "correct public fact")
		if err != nil {
			b.Fatal(err)
		}
		if _, err := fuzzing.Evaluate(context.Background(), request, events, predictor); err != nil {
			b.Fatal(err)
		}
	}
}
