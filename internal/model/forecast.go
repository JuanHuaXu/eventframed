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
	ModelKind        string           `json:"model_kind"`
	HorizonKey       string           `json:"horizon_key"`
	BaseLaw          BernoulliLaw     `json:"base_law"`
	BeliefLaw        *BernoulliLaw    `json:"belief_law,omitempty"`
	PreResidualLaw   BernoulliLaw     `json:"pre_residual_law"`
	CorrectedLaw     BernoulliLaw     `json:"corrected_law"`
	Template         ForecastTemplate `json:"template"`
	PosteriorKey     string           `json:"posterior_key,omitempty"`
	PosteriorVersion uint64           `json:"posterior_version"`
	ResidualApplied  bool             `json:"residual_applied"`
	ResidualRecordID string           `json:"residual_record_id,omitempty"`
}
