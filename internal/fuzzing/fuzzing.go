package fuzzing

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const (
	maxContextEvents = 64
	maxPerturbations = 512
)

type Predictor interface {
	Kind() string
	Predict(ctx context.Context, events []model.Event) ([]float64, error)
}

func Evaluate(ctx context.Context, request model.FuzzSensitivityRequest, events []model.Event, predictor Predictor) (model.FuzzSensitivityResponse, error) {
	if predictor == nil {
		return model.FuzzSensitivityResponse{}, errors.New("fuzz predictor is required")
	}
	if err := validateRequest(request, events); err != nil {
		return model.FuzzSensitivityResponse{}, err
	}
	baseline, err := predictor.Predict(ctx, events)
	if err != nil {
		return model.FuzzSensitivityResponse{}, fmt.Errorf("baseline prediction: %w", err)
	}
	if err := validateLaw(baseline, len(events)); err != nil {
		return model.FuzzSensitivityResponse{}, fmt.Errorf("baseline prediction: %w", err)
	}
	indexes := make(map[string]int, len(events))
	baselineLaw := make(map[string]float64, len(events))
	for index, event := range events {
		indexes[event.ID] = index
		baselineLaw[event.ID] = baseline[index]
	}

	trials := make([]model.FuzzTrialResult, 0, len(request.Perturbations))
	for _, perturbation := range request.Perturbations {
		perturbed, fields, err := apply(events, indexes[perturbation.EventID], perturbation)
		if err != nil {
			return model.FuzzSensitivityResponse{}, fmt.Errorf("perturbation %q: %w", perturbation.ID, err)
		}
		law, err := predictor.Predict(ctx, perturbed)
		if err != nil {
			return model.FuzzSensitivityResponse{}, fmt.Errorf("perturbation %q prediction: %w", perturbation.ID, err)
		}
		if err := validateLaw(law, len(events)); err != nil {
			return model.FuzzSensitivityResponse{}, fmt.Errorf("perturbation %q prediction: %w", perturbation.ID, err)
		}
		tv, maxMovement, maxIndex := lawDistance(baseline, law)
		trial := model.FuzzTrialResult{
			ID: perturbation.ID, PropertyID: perturbation.PropertyID, EventID: perturbation.EventID,
			Fields: fields, ValidityRuleID: perturbation.ValidityRuleID, ValidationKind: perturbation.ValidationKind,
			SourceEventID:  perturbation.SourceEventID,
			TotalVariation: tv, MaxAbsoluteMovement: maxMovement,
			MaxMovedEventID: events[maxIndex].ID, Stable: tv <= request.StabilityThreshold,
		}
		trials = append(trials, trial)
	}

	properties := Summarize(trials, request.RequiredStableProbability, request.ConfidenceLevel, request.MinTrials)

	return model.FuzzSensitivityResponse{
		ProtocolVersion: model.ProtocolVersion, PredictorKind: predictor.Kind(),
		OutputFunctional: "normalized-context-embedding-nomination-law", DistanceKind: "total-variation",
		StabilityThreshold: request.StabilityThreshold, RequiredStableProbability: request.RequiredStableProbability,
		ConfidenceLevel: request.ConfidenceLevel, BaselineLaw: baselineLaw, Trials: trials, Properties: properties,
		CausalClaim: false,
	}, nil
}

func Summarize(trials []model.FuzzTrialResult, requiredStableProbability, confidenceLevel float64, minTrials int) []model.FuzzPropertyReport {
	groups := make(map[string][]model.FuzzTrialResult)
	for _, trial := range trials {
		groups[trial.PropertyID] = append(groups[trial.PropertyID], trial)
	}
	if len(groups) == 0 {
		return nil
	}
	boundConfidence := 1 - (1-confidenceLevel)/float64(len(groups))
	properties := make([]model.FuzzPropertyReport, 0, len(groups))
	for propertyID, propertyTrials := range groups {
		report := model.FuzzPropertyReport{PropertyID: propertyID, Trials: len(propertyTrials)}
		for _, trial := range propertyTrials {
			report.MeanTotalVariation += trial.TotalVariation
			report.MaxTotalVariation = math.Max(report.MaxTotalVariation, trial.TotalVariation)
			if trial.Stable {
				report.StableTrials++
			}
		}
		report.MeanTotalVariation /= float64(report.Trials)
		report.StableFraction = float64(report.StableTrials) / float64(report.Trials)
		report.BoundConfidenceLevel = boundConfidence
		report.StableProbabilityLCB = wilsonLowerBound(report.StableTrials, report.Trials, boundConfidence)
		report.ConditionalInvariant = report.Trials >= minTrials && report.StableProbabilityLCB >= requiredStableProbability
		properties = append(properties, report)
	}
	sort.Slice(properties, func(i, j int) bool { return properties[i].PropertyID < properties[j].PropertyID })
	return properties
}

func validateRequest(request model.FuzzSensitivityRequest, events []model.Event) error {
	if strings.TrimSpace(request.TenantID) == "" || strings.TrimSpace(request.Query) == "" || request.AsOf.IsZero() {
		return errors.New("tenant_id, query, and as_of are required")
	}
	if len(events) < 2 || len(events) > maxContextEvents || len(request.EventIDs) != len(events) {
		return fmt.Errorf("fuzz context must contain between 2 and %d events", maxContextEvents)
	}
	if len(request.Perturbations) == 0 || len(request.Perturbations) > maxPerturbations {
		return fmt.Errorf("fuzz request must contain between 1 and %d perturbations", maxPerturbations)
	}
	if math.IsNaN(request.StabilityThreshold) || math.IsInf(request.StabilityThreshold, 0) || request.StabilityThreshold <= 0 || request.StabilityThreshold > 1 {
		return errors.New("stability_threshold must be in (0,1]")
	}
	if math.IsNaN(request.RequiredStableProbability) || math.IsInf(request.RequiredStableProbability, 0) || request.RequiredStableProbability <= 0 || request.RequiredStableProbability > 1 {
		return errors.New("required_stable_probability must be in (0,1]")
	}
	if request.ConfidenceLevel != .90 && request.ConfidenceLevel != .95 && request.ConfidenceLevel != .99 {
		return errors.New("confidence_level must be one of 0.90, 0.95, or 0.99")
	}
	if request.MinTrials < 1 || request.MinTrials > maxPerturbations {
		return fmt.Errorf("min_trials must be in [1,%d]", maxPerturbations)
	}
	eventIDs := make(map[string]struct{}, len(events))
	for index, event := range events {
		if err := event.Validate(len(event.Embedding)); err != nil {
			return fmt.Errorf("event %q is not a valid EventFrame: %w", event.ID, err)
		}
		if event.ID != request.EventIDs[index] {
			return errors.New("resolved event order does not match event_ids")
		}
		if event.TenantID != request.TenantID || event.AvailableAt.After(request.AsOf) {
			return errors.New("fuzz context contains cross-tenant or future-available evidence")
		}
		if _, duplicate := eventIDs[event.ID]; duplicate {
			return errors.New("event_ids must be unique")
		}
		eventIDs[event.ID] = struct{}{}
	}
	perturbationIDs := make(map[string]struct{}, len(request.Perturbations))
	for _, perturbation := range request.Perturbations {
		if strings.TrimSpace(perturbation.ID) == "" || strings.TrimSpace(perturbation.PropertyID) == "" || strings.TrimSpace(perturbation.ValidityRuleID) == "" {
			return errors.New("every perturbation requires id, property_id, and validity_rule_id")
		}
		if _, duplicate := perturbationIDs[perturbation.ID]; duplicate {
			return errors.New("perturbation ids must be unique")
		}
		perturbationIDs[perturbation.ID] = struct{}{}
		if _, ok := eventIDs[perturbation.EventID]; !ok {
			return fmt.Errorf("perturbation %q targets an event outside the context", perturbation.ID)
		}
		if len(perturbation.Replacements) == 0 || len(perturbation.Replacements) > 6 {
			return fmt.Errorf("perturbation %q must replace between 1 and 6 fields", perturbation.ID)
		}
		for field, replacement := range perturbation.Replacements {
			if !field.Valid() {
				return fmt.Errorf("perturbation %q uses unknown field %q", perturbation.ID, field)
			}
			if replacement.Source != model.SourceSynthetic || strings.TrimSpace(replacement.Evidence) == "" {
				return fmt.Errorf("perturbation %q replacements must be synthetic and carry validity evidence", perturbation.ID)
			}
			if err := replacement.Validate(); err != nil {
				return fmt.Errorf("perturbation %q field %q: %w", perturbation.ID, field, err)
			}
		}
		if err := validatePerturbationAuthority(perturbation, events, eventIDs); err != nil {
			return err
		}
	}
	return nil
}

func validatePerturbationAuthority(perturbation model.FuzzPerturbation, events []model.Event, eventIDs map[string]struct{}) error {
	switch perturbation.ValidationKind {
	case model.FuzzValidationDeclaredRelocation:
		if perturbation.SourceEventID != "" {
			return fmt.Errorf("perturbation %q declared relocation cannot name a source event", perturbation.ID)
		}
		for field := range perturbation.Replacements {
			if field != model.FuzzWho && field != model.FuzzWhere && field != model.FuzzWhen {
				return fmt.Errorf("perturbation %q declared relocation may change only who, where, or when", perturbation.ID)
			}
		}
	case model.FuzzValidationSourceEventBundle:
		if perturbation.SourceEventID == "" || perturbation.SourceEventID == perturbation.EventID {
			return fmt.Errorf("perturbation %q semantic bundle requires a distinct source_event_id", perturbation.ID)
		}
		if _, ok := eventIDs[perturbation.SourceEventID]; !ok {
			return fmt.Errorf("perturbation %q source event is outside the as-of context", perturbation.ID)
		}
		if len(perturbation.Replacements) != 3 {
			return fmt.Errorf("perturbation %q semantic bundle must atomically replace what, why, and how", perturbation.ID)
		}
		for _, field := range []model.FuzzField{model.FuzzWhat, model.FuzzWhy, model.FuzzHow} {
			if _, ok := perturbation.Replacements[field]; !ok {
				return fmt.Errorf("perturbation %q semantic bundle must atomically replace what, why, and how", perturbation.ID)
			}
		}
		var source model.Event
		for _, event := range events {
			if event.ID == perturbation.SourceEventID {
				source = event
				break
			}
		}
		for field, replacement := range perturbation.Replacements {
			if replacement.Value != fieldValue(source, field).Value {
				return fmt.Errorf("perturbation %q replacement %q does not match its source EventFrame", perturbation.ID, field)
			}
		}
	default:
		return fmt.Errorf("perturbation %q has unknown validation_kind %q", perturbation.ID, perturbation.ValidationKind)
	}
	return nil
}

func apply(events []model.Event, index int, perturbation model.FuzzPerturbation) ([]model.Event, []model.FuzzField, error) {
	result := append([]model.Event(nil), events...)
	event := result[index]
	fields := make([]model.FuzzField, 0, len(perturbation.Replacements))
	changed := false
	for field, replacement := range perturbation.Replacements {
		fields = append(fields, field)
		current := fieldValue(event, field)
		if current.Value != replacement.Value {
			changed = true
		}
		setField(&event, field, replacement)
	}
	if !changed {
		return nil, nil, errors.New("replacement must change at least one field")
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i] < fields[j] })
	event.Embedding = nil
	event.EmbeddingModel = ""
	if err := event.Validate(0); err != nil {
		return nil, nil, fmt.Errorf("perturbed EventFrame is invalid: %w", err)
	}
	result[index] = event
	return result, fields, nil
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

func setField(event *model.Event, field model.FuzzField, replacement model.Field) {
	switch field {
	case model.FuzzWho:
		event.Who = replacement
	case model.FuzzWhat:
		event.What = replacement
	case model.FuzzWhere:
		event.Where = replacement
	case model.FuzzWhen:
		event.When = replacement
	case model.FuzzWhy:
		event.Why = replacement
	case model.FuzzHow:
		event.How = replacement
	}
}

func validateLaw(law []float64, size int) error {
	if len(law) != size || len(law) == 0 {
		return errors.New("predictor law dimension does not match the context")
	}
	total := 0.0
	for _, probability := range law {
		if math.IsNaN(probability) || math.IsInf(probability, 0) || probability < 0 || probability > 1 {
			return errors.New("predictor law contains an invalid probability")
		}
		total += probability
	}
	if math.Abs(total-1) > 1e-9 {
		return fmt.Errorf("predictor law mass is %.12f, want 1", total)
	}
	return nil
}

func lawDistance(left, right []float64) (totalVariation, maxMovement float64, maxIndex int) {
	for index := range left {
		movement := math.Abs(left[index] - right[index])
		totalVariation += movement / 2
		if movement > maxMovement {
			maxMovement, maxIndex = movement, index
		}
	}
	return totalVariation, maxMovement, maxIndex
}

func wilsonLowerBound(successes, trials int, confidence float64) float64 {
	if trials == 0 {
		return 0
	}
	z := math.Sqrt2 * math.Erfinv(2*confidence-1)
	n := float64(trials)
	p := float64(successes) / n
	z2 := z * z
	center := p + z2/(2*n)
	radius := z * math.Sqrt(p*(1-p)/n+z2/(4*n*n))
	return math.Max(0, (center-radius)/(1+z2/n))
}
