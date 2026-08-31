package model

import "time"

type FuzzField string
type FuzzValidationKind string

const (
	FuzzWho   FuzzField = "who"
	FuzzWhat  FuzzField = "what"
	FuzzWhere FuzzField = "where"
	FuzzWhen  FuzzField = "when"
	FuzzWhy   FuzzField = "why"
	FuzzHow   FuzzField = "how"
)

const (
	FuzzValidationDeclaredRelocation FuzzValidationKind = "declared_context_relocation"
	FuzzValidationSourceEventBundle  FuzzValidationKind = "source_event_semantic_bundle"
)

func (field FuzzField) Valid() bool {
	switch field {
	case FuzzWho, FuzzWhat, FuzzWhere, FuzzWhen, FuzzWhy, FuzzHow:
		return true
	default:
		return false
	}
}

// FuzzPerturbation is a declared partial-map application. Replacements are
// atomic: either every field validates and is applied, or the trial is rejected.
type FuzzPerturbation struct {
	ID             string              `json:"id"`
	PropertyID     string              `json:"property_id"`
	EventID        string              `json:"event_id"`
	ValidityRuleID string              `json:"validity_rule_id"`
	ValidationKind FuzzValidationKind  `json:"validation_kind"`
	SourceEventID  string              `json:"source_event_id,omitempty"`
	Replacements   map[FuzzField]Field `json:"replacements"`
}

type FuzzSensitivityRequest struct {
	ProtocolVersion           string             `json:"protocol_version"`
	TenantID                  string             `json:"tenant_id"`
	Query                     string             `json:"query"`
	AsOf                      time.Time          `json:"as_of"`
	BaseSnapshot              Snapshot           `json:"base_snapshot"`
	EventIDs                  []string           `json:"event_ids"`
	Perturbations             []FuzzPerturbation `json:"perturbations"`
	StabilityThreshold        float64            `json:"stability_threshold"`
	RequiredStableProbability float64            `json:"required_stable_probability"`
	ConfidenceLevel           float64            `json:"confidence_level"`
	MinTrials                 int                `json:"min_trials"`
}

type FuzzTrialResult struct {
	ID                  string             `json:"id"`
	PropertyID          string             `json:"property_id"`
	EventID             string             `json:"event_id"`
	Fields              []FuzzField        `json:"fields"`
	ValidityRuleID      string             `json:"validity_rule_id"`
	ValidationKind      FuzzValidationKind `json:"validation_kind"`
	SourceEventID       string             `json:"source_event_id,omitempty"`
	TotalVariation      float64            `json:"total_variation"`
	MaxAbsoluteMovement float64            `json:"max_absolute_movement"`
	MaxMovedEventID     string             `json:"max_moved_event_id"`
	Stable              bool               `json:"stable"`
}

type FuzzPropertyReport struct {
	PropertyID           string  `json:"property_id"`
	Trials               int     `json:"trials"`
	StableTrials         int     `json:"stable_trials"`
	StableFraction       float64 `json:"stable_fraction"`
	StableProbabilityLCB float64 `json:"stable_probability_lcb"`
	BoundConfidenceLevel float64 `json:"bound_confidence_level"`
	MeanTotalVariation   float64 `json:"mean_total_variation"`
	MaxTotalVariation    float64 `json:"max_total_variation"`
	ConditionalInvariant bool    `json:"conditional_invariant"`
}

type FuzzSensitivityResponse struct {
	ProtocolVersion           string               `json:"protocol_version"`
	PredictorKind             string               `json:"predictor_kind"`
	OutputFunctional          string               `json:"output_functional"`
	DistanceKind              string               `json:"distance_kind"`
	StabilityThreshold        float64              `json:"stability_threshold"`
	RequiredStableProbability float64              `json:"required_stable_probability"`
	ConfidenceLevel           float64              `json:"confidence_level"`
	BaselineLaw               map[string]float64   `json:"baseline_law"`
	Trials                    []FuzzTrialResult    `json:"trials"`
	Properties                []FuzzPropertyReport `json:"properties"`
	Snapshot                  Snapshot             `json:"snapshot"`
	CausalClaim               bool                 `json:"causal_claim"`
}
