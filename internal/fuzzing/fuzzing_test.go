package fuzzing_test

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/fuzzing"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestEvaluateReportsConditionalInvariantWithoutMutatingContext(t *testing.T) {
	events := testEvents()
	original := append([]model.Event(nil), events...)
	request := validRequest(events)
	for index := 0; index < 64; index++ {
		request.Perturbations = append(request.Perturbations, model.FuzzPerturbation{
			ID: fmt.Sprintf("stable-%02d", index), PropertyID: "wording", EventID: "relevant",
			ValidityRuleID: "valid-context-relocation-v1", ValidationKind: model.FuzzValidationDeclaredRelocation,
			Replacements: map[model.FuzzField]model.Field{
				model.FuzzWho: synthetic("alternate participant", "public context relocation"),
			},
		})
	}
	response, err := fuzzing.Evaluate(context.Background(), request, events, scriptedPredictor{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(events, original) {
		t.Fatal("fuzzing mutated the durable context")
	}
	if len(response.Properties) != 1 || !response.Properties[0].ConditionalInvariant || response.Properties[0].StableProbabilityLCB < .95 {
		t.Fatalf("property report = %+v", response.Properties)
	}
	if response.CausalClaim || response.DistanceKind != "total-variation" {
		t.Fatalf("response semantics = %+v", response)
	}
}

func TestEvaluateDetectsSensitiveBundleAndRequiresConfidenceCoverage(t *testing.T) {
	events := testEvents()
	request := validRequest(events)
	request.MinTrials = 2
	request.Perturbations = []model.FuzzPerturbation{{
		ID: "meaning-change", PropertyID: "semantic-bundle", EventID: "relevant", ValidityRuleID: "public-complete-bundle-v1",
		ValidationKind: model.FuzzValidationSourceEventBundle, SourceEventID: "distractor",
		Replacements: map[model.FuzzField]model.Field{
			model.FuzzWhat: synthetic(events[1].What.Value, "complete public alternative"),
			model.FuzzWhy:  synthetic(events[1].Why.Value, "complete public alternative"),
			model.FuzzHow:  synthetic(events[1].How.Value, "complete public alternative"),
		},
	}}
	response, err := fuzzing.Evaluate(context.Background(), request, events, scriptedPredictor{})
	if err != nil {
		t.Fatal(err)
	}
	if response.Trials[0].Stable || response.Trials[0].TotalVariation < .5 || response.Properties[0].ConditionalInvariant {
		t.Fatalf("sensitive report = %+v", response)
	}
}

func TestEvaluateRejectsUndeclaredOrNoOpPerturbations(t *testing.T) {
	events := testEvents()
	request := validRequest(events)
	request.Perturbations = []model.FuzzPerturbation{{
		ID: "bad", PropertyID: "wording", EventID: "relevant", ValidityRuleID: "",
		ValidationKind: model.FuzzValidationDeclaredRelocation,
		Replacements:   map[model.FuzzField]model.Field{model.FuzzWho: synthetic("changed", "evidence")},
	}}
	if _, err := fuzzing.Evaluate(context.Background(), request, events, scriptedPredictor{}); err == nil {
		t.Fatal("undeclared validity rule was accepted")
	}
	request.Perturbations[0].ValidityRuleID = "rule"
	request.Perturbations[0].Replacements[model.FuzzWho] = synthetic(events[0].Who.Value, "same semantic value")
	if _, err := fuzzing.Evaluate(context.Background(), request, events, scriptedPredictor{}); err == nil {
		t.Fatal("no-op perturbation was accepted")
	}
}

func TestEvaluateRejectsUnverifiedSemanticBundleAndNaNProbability(t *testing.T) {
	events := testEvents()
	request := validRequest(events)
	request.Perturbations = []model.FuzzPerturbation{{
		ID: "forged", PropertyID: "semantic", EventID: "relevant", ValidityRuleID: "claimed-source",
		ValidationKind: model.FuzzValidationSourceEventBundle, SourceEventID: "distractor",
		Replacements: map[model.FuzzField]model.Field{
			model.FuzzWhat: synthetic("forged meaning", "claimed source"),
			model.FuzzWhy:  synthetic(events[1].Why.Value, "claimed source"),
			model.FuzzHow:  synthetic(events[1].How.Value, "claimed source"),
		},
	}}
	if _, err := fuzzing.Evaluate(context.Background(), request, events, scriptedPredictor{}); err == nil {
		t.Fatal("semantic replacement not present in the source EventFrame was accepted")
	}
	request.Perturbations = []model.FuzzPerturbation{{
		ID: "relocate", PropertyID: "context", EventID: "relevant", ValidityRuleID: "relocation",
		ValidationKind: model.FuzzValidationDeclaredRelocation,
		Replacements:   map[model.FuzzField]model.Field{model.FuzzWho: synthetic("other", "declared")},
	}}
	request.RequiredStableProbability = math.NaN()
	if _, err := fuzzing.Evaluate(context.Background(), request, events, scriptedPredictor{}); err == nil {
		t.Fatal("NaN required stability probability was accepted")
	}
}

func TestEmbeddingNominationPredictorProducesNormalizedMovement(t *testing.T) {
	embedder, _ := embed.NewHashEmbedder(64)
	predictor, err := fuzzing.NewEmbeddingNominationPredictor(context.Background(), embedder, "relevant answer")
	if err != nil {
		t.Fatal(err)
	}
	events := testEvents()
	baseline, err := predictor.Predict(context.Background(), events)
	if err != nil || math.Abs(baseline[0]+baseline[1]-1) > 1e-12 {
		t.Fatalf("baseline law = %v, %v", baseline, err)
	}
	events[0].What = synthetic("unrelated answer", "valid test replacement")
	perturbed, err := predictor.Predict(context.Background(), events)
	if err != nil || reflect.DeepEqual(baseline, perturbed) {
		t.Fatalf("perturbed law = %v, %v", perturbed, err)
	}
}

func TestEmbeddingNominationPredictorFromVectorCopiesAndValidatesInput(t *testing.T) {
	embedder, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	vector := []float32{1, 0, 0, 0, 0, 0, 0, 0}
	predictor, err := fuzzing.NewEmbeddingNominationPredictorFromVector(embedder, vector, "query-key", "snapshot", nil)
	if err != nil {
		t.Fatal(err)
	}
	vector[0] = 0
	if law, predictErr := predictor.Predict(context.Background(), testEvents()); predictErr != nil || len(law) != 2 {
		t.Fatalf("vector predictor = %v, %v", law, predictErr)
	}
	if _, err := fuzzing.NewEmbeddingNominationPredictorFromVector(embedder, []float32{1}, "query-key", "snapshot", nil); err == nil {
		t.Fatal("mismatched background query vector was accepted")
	}
	invalid := []float32{1, 0, 0, 0, 0, 0, 0, float32(math.Inf(1))}
	if _, err := fuzzing.NewEmbeddingNominationPredictorFromVector(embedder, invalid, "query-key", "snapshot", nil); err == nil {
		t.Fatal("non-finite background query vector was accepted")
	}
}

type scriptedPredictor struct{}

func (scriptedPredictor) Kind() string { return "test-scripted" }
func (scriptedPredictor) Predict(_ context.Context, events []model.Event) ([]float64, error) {
	switch events[0].What.Value {
	case "unrelated answer":
		return []float64{.1, .9}, nil
	case "relevant answer paraphrase":
		return []float64{.79, .21}, nil
	default:
		return []float64{.8, .2}, nil
	}
}

func validRequest(events []model.Event) model.FuzzSensitivityRequest {
	return model.FuzzSensitivityRequest{
		TenantID: "public", Query: "relevant answer", AsOf: time.Now().UTC(), EventIDs: []string{events[0].ID, events[1].ID},
		StabilityThreshold: .05, RequiredStableProbability: .95, ConfidenceLevel: .95, MinTrials: 1,
	}
}

func testEvents() []model.Event {
	now := time.Now().UTC().Add(-time.Hour)
	return []model.Event{
		testEvent("relevant", "relevant answer", now),
		testEvent("distractor", "unrelated answer", now.Add(time.Minute)),
	}
}

func testEvent(id, what string, at time.Time) model.Event {
	field := func(value string) model.Field {
		return model.Field{Value: value, Source: model.SourceObserved, Confidence: 1}
	}
	return model.Event{
		ID: id, TenantID: "public", SessionID: "fixture", Kind: "public-fact", Content: what,
		OccurredAt: at, ObservedAt: at, AvailableAt: at, Who: field("source"), What: field(what),
		Where: field("fixture"), When: field(at.Format(time.RFC3339)), Why: field("testing"), How: field("public record"),
		Provenance: model.Provenance{Producer: "test"},
	}
}

func synthetic(value, evidence string) model.Field {
	return model.Field{Value: value, Source: model.SourceSynthetic, Confidence: 1, Evidence: evidence}
}
