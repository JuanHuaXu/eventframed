package model

const RetrievalUsefulnessHorizon = "retrieval-usefulness-v1"

// BernoulliLaw is the runtime specialization of the paper's full outcome law.
// Both branches are explicit so a correction cannot silently ignore a negative
// usefulness outcome.
type BernoulliLaw struct {
	Useful    float64 `json:"useful"`
	NotUseful float64 `json:"not_useful"`
}

type ForecastTemplate struct {
	EventID         string  `json:"event_id"`
	PredictedUseful bool    `json:"predicted_useful"`
	Confidence      float64 `json:"confidence"`
}

// ForecastBundle preserves the canonical Phase 4 order. BaseLaw is the
// declared plug-in baseline; BeliefLaw is present only when a certified
// posterior predictive was accepted. PreResidualLaw is the law presented to
// residual selection, and CorrectedLaw is the complete law returned afterward.
type ForecastBundle struct {
	ExpertMixture          ExpertForecasts  `json:"expert_mixture,omitzero"`
	ModelKind              string           `json:"model_kind"`
	RankScore              float64          `json:"rank_score"`
	HorizonKey             string           `json:"horizon_key"`
	BaseLaw                BernoulliLaw     `json:"base_law"`
	BeliefLaw              *BernoulliLaw    `json:"belief_law,omitempty"`
	PreResidualLaw         BernoulliLaw     `json:"pre_residual_law"`
	CorrectedLaw           BernoulliLaw     `json:"corrected_law"`
	Template               ForecastTemplate `json:"template"`
	PosteriorKey           string           `json:"posterior_key,omitempty"`
	PosteriorVersion       uint64           `json:"posterior_version"`
	ResidualApplied        bool             `json:"residual_applied"`
	ResidualShadowEligible bool             `json:"residual_shadow_eligible"`
	ResidualRecordID       string           `json:"residual_record_id,omitempty"`
}

// ExpertForecasts is an immutable, pre-outcome journal commitment. Versions
// prevent delayed feedback from training a selector under a different contract.
type ExpertForecasts struct {
	Enabled       bool       `json:"enabled"`
	Probabilities [4]float64 `json:"probabilities"`
	PolicyVersion uint64     `json:"policy_version"`
	EvidenceEpoch uint64     `json:"evidence_epoch"`
}
