package model

import "time"

type ChainTranslationClass string

const (
	ChainHigherOrderInvariant  ChainTranslationClass = "higher_order_invariant_candidate"
	ChainPredictiveTranslation ChainTranslationClass = "predictive_translation_candidate"
	ChainDivergence            ChainTranslationClass = "divergence"
)

type ChainTrajectory struct {
	BaselineEventIDs []string `json:"baseline_event_ids"`
	RevealedEventIDs []string `json:"revealed_event_ids"`
}

// ChainStageMap declares the coordinate and value correspondence used at one
// aligned stage. Values are exact because normalization belongs to the frozen
// map-generating procedure, not to this audit endpoint.
type ChainStageMap struct {
	Stage            int       `json:"stage"`
	DomainAField     FuzzField `json:"domain_a_field"`
	DomainBField     FuzzField `json:"domain_b_field"`
	DomainABefore    string    `json:"domain_a_before"`
	DomainAAfter     string    `json:"domain_a_after"`
	DomainBBefore    string    `json:"domain_b_before"`
	DomainBAfter     string    `json:"domain_b_after"`
	CorrespondenceID string    `json:"correspondence_id"`
	ValidityEvidence string    `json:"validity_evidence"`
}

type ChainTranslationRequest struct {
	ProtocolVersion      string          `json:"protocol_version"`
	TenantID             string          `json:"tenant_id"`
	Query                string          `json:"query"`
	AsOf                 time.Time       `json:"as_of"`
	BaseSnapshot         Snapshot        `json:"base_snapshot"`
	DomainA              ChainTrajectory `json:"domain_a"`
	DomainB              ChainTrajectory `json:"domain_b"`
	StageMaps            []ChainStageMap `json:"stage_maps"`
	InvariantThreshold   float64         `json:"invariant_threshold"`
	TranslationThreshold float64         `json:"translation_threshold"`
}

type ChainStageResult struct {
	Stage               int       `json:"stage"`
	DomainAField        FuzzField `json:"domain_a_field"`
	DomainBField        FuzzField `json:"domain_b_field"`
	DomainAChanged      bool      `json:"domain_a_changed"`
	DomainBChanged      bool      `json:"domain_b_changed"`
	DomainALocal        bool      `json:"domain_a_local"`
	DomainBLocal        bool      `json:"domain_b_local"`
	BeforeValuesAligned bool      `json:"before_values_aligned"`
	AfterValuesAligned  bool      `json:"after_values_aligned"`
	EdgewiseCommuting   bool      `json:"edgewise_commuting"`
	CorrespondenceID    string    `json:"correspondence_id"`
}

type ChainTranslationResponse struct {
	ProtocolVersion          string                `json:"protocol_version"`
	Classification           ChainTranslationClass `json:"classification"`
	PredictorKind            string                `json:"predictor_kind"`
	OutputFunctional         string                `json:"output_functional"`
	PredictionEvaluated      bool                  `json:"prediction_evaluated"`
	DomainAMovement          float64               `json:"domain_a_movement"`
	DomainBMovement          float64               `json:"domain_b_movement"`
	PredictionEffectDefect   float64               `json:"prediction_effect_defect"`
	InvariantThreshold       float64               `json:"invariant_threshold"`
	TranslationThreshold     float64               `json:"translation_threshold"`
	LocalitySatisfied        bool                  `json:"locality_satisfied"`
	EdgewiseCommuting        bool                  `json:"edgewise_commuting"`
	FullyPropagated          bool                  `json:"fully_propagated"`
	TerminalAgreement        bool                  `json:"terminal_agreement"`
	Stages                   []ChainStageResult    `json:"stages"`
	ClaimScope               string                `json:"claim_scope"`
	CausalClaim              bool                  `json:"causal_claim"`
	PublishesGraphOrGrouping bool                  `json:"publishes_graph_or_grouping"`
	Snapshot                 Snapshot              `json:"snapshot"`
}
