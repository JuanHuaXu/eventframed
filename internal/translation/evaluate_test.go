package translation

import (
	"context"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type fixedPredictor map[string][]float64

func (fixedPredictor) Kind() string { return "fixed-chain-test" }
func (p fixedPredictor) Predict(_ context.Context, events []model.Event) ([]float64, error) {
	return append([]float64(nil), p[events[0].ID]...), nil
}

func TestEvaluateClassifiesEdgewisePredictiveTranslation(t *testing.T) {
	request, chains := fixture(t)
	predictor := fixedPredictor{
		"a0-0": {.70, .20, .10}, "a1-0": {.60, .25, .15},
		"b0-0": {.70, .20, .10}, "b1-0": {.60, .25, .15},
	}
	response, err := Evaluate(context.Background(), request, chains, predictor)
	if err != nil {
		t.Fatal(err)
	}
	if response.Classification != model.ChainPredictiveTranslation || !response.LocalitySatisfied || !response.EdgewiseCommuting || !response.TerminalAgreement || response.CausalClaim || response.PublishesGraphOrGrouping {
		t.Fatalf("response = %+v", response)
	}
}

func TestEvaluateGivesInvariantPrecedenceWhenPredictionErasesPerturbation(t *testing.T) {
	request, chains := fixture(t)
	last := len(request.StageMaps) - 1
	chains.DomainARevealed[last].What.Value = chains.DomainABaseline[last].What.Value
	chains.DomainBRevealed[last].What.Value = chains.DomainBBaseline[last].What.Value
	request.StageMaps[last].DomainAAfter = request.StageMaps[last].DomainABefore
	request.StageMaps[last].DomainBAfter = request.StageMaps[last].DomainBBefore
	predictor := fixedPredictor{
		"a0-0": {.70, .20, .10}, "a1-0": {.70, .20, .10},
		"b0-0": {.70, .20, .10}, "b1-0": {.70, .20, .10},
	}
	response, err := Evaluate(context.Background(), request, chains, predictor)
	if err != nil {
		t.Fatal(err)
	}
	if response.Classification != model.ChainHigherOrderInvariant {
		t.Fatalf("classification = %q", response.Classification)
	}
}

func TestEvaluateDoesNotCallLowMovementInvariantWhenTerminalChanged(t *testing.T) {
	request, chains := fixture(t)
	predictor := fixedPredictor{
		"a0-0": {.70, .20, .10}, "a1-0": {.70, .20, .10},
		"b0-0": {.70, .20, .10}, "b1-0": {.70, .20, .10},
	}
	response, err := Evaluate(context.Background(), request, chains, predictor)
	if err != nil {
		t.Fatal(err)
	}
	if response.Classification != model.ChainPredictiveTranslation {
		t.Fatalf("classification = %q", response.Classification)
	}
}

func TestEvaluateRejectsEndpointOnlyMatchWhenIntermediateLocalityFails(t *testing.T) {
	request, chains := fixture(t)
	chains.DomainBRevealed[1].Where.Value = "an undeclared second movement"
	predictor := fixedPredictor{
		"a0-0": {.70, .20, .10}, "a1-0": {.60, .25, .15},
		"b0-0": {.70, .20, .10}, "b1-0": {.60, .25, .15},
	}
	response, err := Evaluate(context.Background(), request, chains, predictor)
	if err != nil {
		t.Fatal(err)
	}
	if response.Classification != model.ChainDivergence || response.LocalitySatisfied || response.EdgewiseCommuting {
		t.Fatalf("response = %+v", response)
	}
	if response.PredictionEvaluated {
		t.Fatal("structural divergence unnecessarily evaluated the predictor")
	}
}

func TestEvaluateStructuralDivergenceDoesNotConstructPredictor(t *testing.T) {
	request, chains := fixture(t)
	chains.DomainBRevealed[1].Where.Value = "undeclared movement"
	called := false
	response, err := EvaluateWithFactory(context.Background(), request, chains, func(context.Context) (Predictor, error) {
		called = true
		return fixedPredictor{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if called || response.PredictionEvaluated || response.Classification != model.ChainDivergence {
		t.Fatalf("factory called=%t response=%+v", called, response)
	}
}

func TestEvaluateRejectsUnmappedObservedValue(t *testing.T) {
	request, chains := fixture(t)
	request.StageMaps[1].DomainBAfter = "wrong"
	predictor := fixedPredictor{
		"a0-0": {.70, .20, .10}, "a1-0": {.60, .25, .15},
		"b0-0": {.70, .20, .10}, "b1-0": {.60, .25, .15},
	}
	response, err := Evaluate(context.Background(), request, chains, predictor)
	if err != nil {
		t.Fatal(err)
	}
	if response.Classification != model.ChainDivergence || response.Stages[1].AfterValuesAligned {
		t.Fatalf("response = %+v", response)
	}
}

func TestEvaluateFailsClosedOnDuplicateEvidence(t *testing.T) {
	request, chains := fixture(t)
	chains.DomainBRevealed[2] = chains.DomainBBaseline[2]
	_, err := Evaluate(context.Background(), request, chains, fixedPredictor{})
	if err == nil {
		t.Fatal("expected duplicate evidence rejection")
	}
}

func TestEvaluateRejectsShuffledTrajectory(t *testing.T) {
	request, chains := fixture(t)
	chains.DomainARevealed[0], chains.DomainARevealed[1] = chains.DomainARevealed[1], chains.DomainARevealed[0]
	request.DomainA.RevealedEventIDs[0], request.DomainA.RevealedEventIDs[1] = request.DomainA.RevealedEventIDs[1], request.DomainA.RevealedEventIDs[0]
	_, err := Evaluate(context.Background(), request, chains, fixedPredictor{})
	if err == nil {
		t.Fatal("expected shuffled trajectory rejection")
	}
}

func fixture(t *testing.T) (model.ChainTranslationRequest, Chains) {
	t.Helper()
	now := time.Now().UTC()
	makeChain := func(prefix string, values []string) []model.Event {
		result := make([]model.Event, len(values))
		for i, value := range values {
			event := testutil.Event(prefix+"-"+string(rune('0'+i)), value, now.Add(-time.Duration(10-i)*time.Minute))
			event.What.Value = value
			event.When.Value = "stage-time"
			result[i] = event
		}
		return result
	}
	aBefore := []string{"low input", "low mechanism", "low outcome"}
	aAfter := []string{"high input", "high mechanism", "high outcome"}
	bBefore := []string{"cold input", "cold mechanism", "cold outcome"}
	bAfter := []string{"hot input", "hot mechanism", "hot outcome"}
	chains := Chains{
		DomainABaseline: makeChain("a0", aBefore), DomainARevealed: makeChain("a1", aAfter),
		DomainBBaseline: makeChain("b0", bBefore), DomainBRevealed: makeChain("b1", bAfter),
	}
	maps := make([]model.ChainStageMap, 3)
	for i := range maps {
		maps[i] = model.ChainStageMap{Stage: i, DomainAField: model.FuzzWhat, DomainBField: model.FuzzWhat,
			DomainABefore: aBefore[i], DomainAAfter: aAfter[i], DomainBBefore: bBefore[i], DomainBAfter: bAfter[i],
			CorrespondenceID: "map-" + string(rune('0'+i)), ValidityEvidence: "frozen test correspondence"}
	}
	request := model.ChainTranslationRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", Query: "outcome",
		AsOf: now, DomainA: ids(chains.DomainABaseline, chains.DomainARevealed),
		DomainB: ids(chains.DomainBBaseline, chains.DomainBRevealed), StageMaps: maps,
		InvariantThreshold: .01, TranslationThreshold: .001,
	}
	return request, chains
}

func ids(baseline, revealed []model.Event) model.ChainTrajectory {
	result := model.ChainTrajectory{BaselineEventIDs: make([]string, len(baseline)), RevealedEventIDs: make([]string, len(revealed))}
	for i := range baseline {
		result.BaselineEventIDs[i], result.RevealedEventIDs[i] = baseline[i].ID, revealed[i].ID
	}
	return result
}
