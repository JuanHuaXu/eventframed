package translation

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// The formal chain is bounded; the runtime freezes that bound at 32 stages to
// keep structural validation and four predictive evaluations resource-bounded.
const maxChainStages = 32

type Predictor interface {
	Kind() string
	Predict(ctx context.Context, events []model.Event) ([]float64, error)
}

type PredictorFactory func(context.Context) (Predictor, error)

type Chains struct {
	DomainABaseline []model.Event
	DomainARevealed []model.Event
	DomainBBaseline []model.Event
	DomainBRevealed []model.Event
}

func Evaluate(ctx context.Context, request model.ChainTranslationRequest, chains Chains, predictor Predictor) (model.ChainTranslationResponse, error) {
	return EvaluateWithFactory(ctx, request, chains, func(context.Context) (Predictor, error) {
		if predictor == nil {
			return nil, errors.New("chain predictor is required")
		}
		return predictor, nil
	})
}

// EvaluateWithFactory defers predictor construction until structural checks
// establish that a prediction can affect the classification.
func EvaluateWithFactory(ctx context.Context, request model.ChainTranslationRequest, chains Chains, factory PredictorFactory) (model.ChainTranslationResponse, error) {
	// Translation is a whole-chain condition: the mapped upstream change must
	// commute through every aligned stage while non-target 5W1H fields stay fixed.
	// Matching only the terminal outcome is insufficient.
	if err := validate(request, chains); err != nil {
		return model.ChainTranslationResponse{}, err
	}
	stages := make([]model.ChainStageResult, len(request.StageMaps))
	locality, commuting, fullyPropagated := true, true, true
	for i, stageMap := range request.StageMaps {
		result := evaluateStage(stageMap,
			chains.DomainABaseline[i], chains.DomainARevealed[i],
			chains.DomainBBaseline[i], chains.DomainBRevealed[i])
		stages[i] = result
		locality = locality && result.DomainALocal && result.DomainBLocal
		commuting = commuting && result.EdgewiseCommuting
		fullyPropagated = fullyPropagated && result.DomainAChanged && result.DomainBChanged
	}
	terminalErased := !stages[len(stages)-1].DomainAChanged && !stages[len(stages)-1].DomainBChanged
	base := model.ChainTranslationResponse{
		ProtocolVersion: model.ProtocolVersion, Classification: model.ChainDivergence,
		PredictorKind: "not-evaluated/structural-divergence", OutputFunctional: "structural-chain-compatibility", PredictionEvaluated: false,
		InvariantThreshold: request.InvariantThreshold, TranslationThreshold: request.TranslationThreshold,
		LocalitySatisfied: locality, EdgewiseCommuting: commuting, FullyPropagated: fullyPropagated,
		Stages: stages, ClaimScope: "predictive", CausalClaim: false, PublishesGraphOrGrouping: false,
	}
	if !locality || !commuting || (!terminalErased && !fullyPropagated) {
		return base, nil
	}
	if factory == nil {
		return model.ChainTranslationResponse{}, errors.New("chain predictor factory is required")
	}
	predictor, err := factory(ctx)
	if err != nil {
		return model.ChainTranslationResponse{}, err
	}
	if predictor == nil {
		return model.ChainTranslationResponse{}, errors.New("chain predictor is required")
	}
	a0, err := predict(ctx, predictor, chains.DomainABaseline, "domain A baseline")
	if err != nil {
		return model.ChainTranslationResponse{}, err
	}
	a1, err := predict(ctx, predictor, chains.DomainARevealed, "domain A revealed")
	if err != nil {
		return model.ChainTranslationResponse{}, err
	}
	b0, err := predict(ctx, predictor, chains.DomainBBaseline, "domain B baseline")
	if err != nil {
		return model.ChainTranslationResponse{}, err
	}
	b1, err := predict(ctx, predictor, chains.DomainBRevealed, "domain B revealed")
	if err != nil {
		return model.ChainTranslationResponse{}, err
	}
	aMovement, bMovement, defect := predictionMetrics(a0, a1, b0, b1)
	// Keep terminal agreement explicit because an aggregate effect defect can
	// hide a large final-stage mismatch behind agreement at earlier stages.
	terminalAgreement := math.Abs((a1[len(a1)-1]-a0[len(a0)-1])-(b1[len(b1)-1]-b0[len(b0)-1])) <= request.TranslationThreshold

	classification := model.ChainDivergence
	if locality && commuting && terminalErased && aMovement <= request.InvariantThreshold && bMovement <= request.InvariantThreshold {
		classification = model.ChainHigherOrderInvariant
	} else if locality && commuting && fullyPropagated && terminalAgreement && defect <= request.TranslationThreshold {
		classification = model.ChainPredictiveTranslation
	}

	base.Classification = classification
	base.PredictorKind = predictor.Kind()
	base.OutputFunctional = "aligned-chain-embedding-nomination-effect"
	base.PredictionEvaluated = true
	base.DomainAMovement, base.DomainBMovement, base.PredictionEffectDefect = aMovement, bMovement, defect
	base.TerminalAgreement = terminalAgreement
	return base, nil
}

func validate(request model.ChainTranslationRequest, chains Chains) error {
	if strings.TrimSpace(request.TenantID) == "" || strings.TrimSpace(request.Query) == "" || request.AsOf.IsZero() {
		return errors.New("tenant_id, query, and as_of are required")
	}
	n := len(request.StageMaps)
	if n < 2 || n > maxChainStages {
		return fmt.Errorf("chain audit requires between 2 and %d aligned stages", maxChainStages)
	}
	lengths := []int{
		len(request.DomainA.BaselineEventIDs), len(request.DomainA.RevealedEventIDs),
		len(request.DomainB.BaselineEventIDs), len(request.DomainB.RevealedEventIDs),
		len(chains.DomainABaseline), len(chains.DomainARevealed),
		len(chains.DomainBBaseline), len(chains.DomainBRevealed),
	}
	for _, length := range lengths {
		if length != n {
			return errors.New("all four chains and stage_maps must have equal length")
		}
	}
	if !finiteUnit(request.InvariantThreshold) || request.InvariantThreshold == 0 || !finiteUnit(request.TranslationThreshold) || request.TranslationThreshold == 0 {
		return errors.New("invariant_threshold and translation_threshold must be in (0,1]")
	}
	requested := [][]string{
		request.DomainA.BaselineEventIDs, request.DomainA.RevealedEventIDs,
		request.DomainB.BaselineEventIDs, request.DomainB.RevealedEventIDs,
	}
	ids := make(map[string]struct{}, 4*n)
	chainSets := [][]model.Event{chains.DomainABaseline, chains.DomainARevealed, chains.DomainBBaseline, chains.DomainBRevealed}
	for chainIndex, chain := range chainSets {
		for eventIndex, event := range chain {
			if event.ID != requested[chainIndex][eventIndex] {
				return errors.New("resolved event order does not match requested chains")
			}
			if err := event.Validate(len(event.Embedding)); err != nil {
				return fmt.Errorf("invalid chain EventFrame %q: %w", event.ID, err)
			}
			if event.TenantID != request.TenantID || event.AvailableAt.After(request.AsOf) {
				return errors.New("chain contains cross-tenant or future-available evidence")
			}
			if _, duplicate := ids[event.ID]; duplicate {
				return errors.New("the four observed chains must use distinct event ids")
			}
			ids[event.ID] = struct{}{}
			if eventIndex > 0 {
				previous := chain[eventIndex-1]
				if !event.OccurredAt.After(previous.OccurredAt) || event.AvailableAt.Before(previous.AvailableAt) {
					return errors.New("each chain must be strictly occurrence-ordered and nondecreasing in availability")
				}
			}
		}
	}
	correspondences := make(map[string]struct{}, n)
	for i, stageMap := range request.StageMaps {
		if stageMap.Stage != i {
			return errors.New("stage_maps must be contiguous and ordered from zero")
		}
		if !stageMap.DomainAField.Valid() || !stageMap.DomainBField.Valid() {
			return fmt.Errorf("stage %d has an unknown field", i)
		}
		if strings.TrimSpace(stageMap.CorrespondenceID) == "" || strings.TrimSpace(stageMap.ValidityEvidence) == "" {
			return fmt.Errorf("stage %d requires correspondence_id and validity_evidence", i)
		}
		if _, duplicate := correspondences[stageMap.CorrespondenceID]; duplicate {
			return errors.New("correspondence_id values must be unique within an audit")
		}
		correspondences[stageMap.CorrespondenceID] = struct{}{}
		aChanged := stageMap.DomainABefore != stageMap.DomainAAfter
		bChanged := stageMap.DomainBBefore != stageMap.DomainBAfter
		if aChanged != bChanged {
			return fmt.Errorf("stage %d correspondence must change or preserve both mapped values together", i)
		}
		if i == 0 && !aChanged {
			return errors.New("stage zero must contain the aligned upstream perturbation")
		}
	}
	return nil
}

func evaluateStage(m model.ChainStageMap, a0, a1, b0, b1 model.Event) model.ChainStageResult {
	aChanged := m.DomainABefore != m.DomainAAfter
	bChanged := m.DomainBBefore != m.DomainBAfter
	aLocal := localTransition(a0, a1, m.DomainAField, aChanged)
	bLocal := localTransition(b0, b1, m.DomainBField, bChanged)
	aBefore, aAfter := fieldValue(a0, m.DomainAField).Value, fieldValue(a1, m.DomainAField).Value
	bBefore, bAfter := fieldValue(b0, m.DomainBField).Value, fieldValue(b1, m.DomainBField).Value
	before := aBefore == m.DomainABefore && bBefore == m.DomainBBefore
	after := aAfter == m.DomainAAfter && bAfter == m.DomainBAfter
	return model.ChainStageResult{Stage: m.Stage, DomainAField: m.DomainAField, DomainBField: m.DomainBField,
		DomainAChanged: aChanged, DomainBChanged: bChanged,
		DomainALocal: aLocal, DomainBLocal: bLocal, BeforeValuesAligned: before, AfterValuesAligned: after,
		EdgewiseCommuting: aLocal && bLocal && before && after, CorrespondenceID: m.CorrespondenceID}
}

func localTransition(before, after model.Event, allowed model.FuzzField, expectedChange bool) bool {
	changed := false
	for _, field := range []model.FuzzField{model.FuzzWho, model.FuzzWhat, model.FuzzWhere, model.FuzzWhen, model.FuzzWhy, model.FuzzHow} {
		equal := fieldValue(before, field).Value == fieldValue(after, field).Value
		if field == allowed {
			changed = !equal
		} else if !equal {
			return false
		}
	}
	return changed == expectedChange
}

func fieldValue(event model.Event, field model.FuzzField) model.Field {
	switch field {
	case model.FuzzWho:
		return event.Who
	case model.FuzzWhat:
		return event.What
	case model.FuzzWhere:
		return event.Where
	case model.FuzzWhen:
		return event.When
	case model.FuzzWhy:
		return event.Why
	case model.FuzzHow:
		return event.How
	default:
		return model.Field{}
	}
}

func predict(ctx context.Context, predictor Predictor, events []model.Event, label string) ([]float64, error) {
	law, err := predictor.Predict(ctx, events)
	if err != nil {
		return nil, fmt.Errorf("%s prediction: %w", label, err)
	}
	if len(law) != len(events) {
		return nil, fmt.Errorf("%s prediction dimension does not match chain", label)
	}
	total := 0.0
	for _, value := range law {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
			return nil, fmt.Errorf("%s prediction contains invalid probability", label)
		}
		total += value
	}
	if math.Abs(total-1) > 1e-9 {
		return nil, fmt.Errorf("%s prediction mass is %.12f, want 1", label, total)
	}
	return law, nil
}

func predictionMetrics(a0, a1, b0, b1 []float64) (aMovement, bMovement, defect float64) {
	for i := range a0 {
		aEffect := a1[i] - a0[i]
		bEffect := b1[i] - b0[i]
		aMovement += math.Abs(aEffect) / 2
		bMovement += math.Abs(bEffect) / 2
		defect += math.Abs(aEffect-bEffect) / 2
	}
	return aMovement, bMovement, defect
}

func finiteUnit(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}
