package model

import "time"

type ResidualScope string

const (
	ResidualExact   ResidualScope = "exact_action"
	ResidualGeneral ResidualScope = "general_posterior"
)

// ResidualRecord is law-only in contract version 4. Correction shifts the
// useful branch of a Bernoulli law; the not-useful branch receives the opposite
// shift so total mass remains one.
type ResidualRecord struct {
	ID                      string        `json:"id"`
	TenantID                string        `json:"tenant_id"`
	Scope                   ResidualScope `json:"scope"`
	Key                     string        `json:"key"`
	HorizonKey              string        `json:"horizon_key"`
	Correction              float64       `json:"correction"`
	EffectiveSupport        float64       `json:"effective_support"`
	ImprovementAlpha        float64       `json:"improvement_alpha"`
	ImprovementBeta         float64       `json:"improvement_beta"`
	ReferenceProbability    float64       `json:"reference_probability"`
	MotionLimit             float64       `json:"motion_limit"`
	ApproximationErrorBound float64       `json:"approximation_error_bound"`
	PolicyVersion           uint64        `json:"policy_version"`
	EvidenceEpoch           uint64        `json:"evidence_epoch"`
	ResidualVersion         uint64        `json:"residual_version"`
	SourceJournalID         string        `json:"source_journal_id"`
	SourceEventID           string        `json:"source_event_id"`
	PosteriorKey            string        `json:"posterior_key"`
	CreatedAt               time.Time     `json:"created_at"`
	UpdatedAt               time.Time     `json:"updated_at"`
	Active                  bool          `json:"active"`
}

func (record ResidualRecord) ImprovementConfidence() float64 {
	if record.ImprovementAlpha <= 0 || record.ImprovementBeta <= 0 {
		return 0
	}
	return record.ImprovementAlpha / (record.ImprovementAlpha + record.ImprovementBeta)
}

type ResidualObservation struct {
	ActionKey            string    `json:"action_key"`
	GeneralKey           string    `json:"general_key"`
	HorizonKey           string    `json:"horizon_key"`
	BaseProbability      float64   `json:"base_probability"`
	CommittedProbability float64   `json:"committed_probability"`
	Useful               bool      `json:"useful"`
	ValidationEligible   bool      `json:"validation_eligible"`
	EventID              string    `json:"event_id"`
	JournalID            string    `json:"journal_id"`
	PosteriorKey         string    `json:"posterior_key"`
	AvailableAt          time.Time `json:"available_at"`
}

type ResidualCandidates struct {
	Exact   *ResidualRecord
	General *ResidualRecord
}
