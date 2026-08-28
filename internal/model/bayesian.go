package model

import "time"

// BayesianJournalEntry is the durable evidence record for one frontier
// nomination and activation decision. QueryDigest binds the decision to its
// input without retaining user query text in the Bayesian audit collection.
type BayesianJournalEntry struct {
	ID          string               `json:"id"`
	TenantID    string               `json:"tenant_id"`
	SessionID   string               `json:"session_id"`
	AsOf        time.Time            `json:"as_of"`
	QueryDigest string               `json:"query_digest"`
	Snapshot    Snapshot             `json:"snapshot"`
	Report      BayesianShadowReport `json:"report"`
}

type SelectionSupportCertificate struct {
	ID                      string    `json:"id"`
	TenantID                string    `json:"tenant_id"`
	PolicyVersion           uint64    `json:"policy_version"`
	EvidenceEpoch           uint64    `json:"evidence_epoch"`
	MinSelectionProbability float64   `json:"min_selection_probability"`
	SimultaneousCoverage    float64   `json:"simultaneous_coverage"`
	Procedure               string    `json:"procedure"`
	Issuer                  string    `json:"issuer"`
	ExternalAudit           bool      `json:"external_audit"`
	ValidFrom               time.Time `json:"valid_from"`
	ValidUntil              time.Time `json:"valid_until"`
}

type AntiPigeonCertificate struct {
	ID                   string    `json:"id"`
	TenantID             string    `json:"tenant_id"`
	MemberEventIDs       []string  `json:"member_event_ids"`
	HorizonKey           string    `json:"horizon_key"`
	GraphVersion         uint64    `json:"graph_version"`
	EvidenceEpoch        uint64    `json:"evidence_epoch"`
	TargetDiameterUCB    float64   `json:"target_diameter_ucb"`
	DiameterLimit        float64   `json:"diameter_limit"`
	EffectiveSupport     float64   `json:"effective_support"`
	MinEffectiveSupport  float64   `json:"min_effective_support"`
	SimultaneousCoverage float64   `json:"simultaneous_coverage"`
	Procedure            string    `json:"procedure"`
	Issuer               string    `json:"issuer"`
	ExternalAudit        bool      `json:"external_audit"`
	ValidUntil           time.Time `json:"valid_until"`
}

type OmittedInfluenceCertificate struct {
	ID                   string    `json:"id"`
	TenantID             string    `json:"tenant_id"`
	PolicyVersion        uint64    `json:"policy_version"`
	EvidenceEpoch        uint64    `json:"evidence_epoch"`
	DivergenceUCB        float64   `json:"divergence_ucb"`
	DivergenceLimit      float64   `json:"divergence_limit"`
	AuditProbability     float64   `json:"audit_probability"`
	SimultaneousCoverage float64   `json:"simultaneous_coverage"`
	Procedure            string    `json:"procedure"`
	Issuer               string    `json:"issuer"`
	ExternalAudit        bool      `json:"external_audit"`
	ValidUntil           time.Time `json:"valid_until"`
}

type PublishSelectionCertificateRequest struct {
	ProtocolVersion string                      `json:"protocol_version"`
	Certificate     SelectionSupportCertificate `json:"certificate"`
}

type PublishAntiPigeonCertificateRequest struct {
	ProtocolVersion string                `json:"protocol_version"`
	Certificate     AntiPigeonCertificate `json:"certificate"`
}

type PublishOmittedInfluenceCertificateRequest struct {
	ProtocolVersion string                      `json:"protocol_version"`
	Certificate     OmittedInfluenceCertificate `json:"certificate"`
}

type CertificateResponse struct {
	ProtocolVersion string   `json:"protocol_version"`
	CertificateID   string   `json:"certificate_id"`
	Snapshot        Snapshot `json:"snapshot"`
}

type BayesianOutcomeSource string

const (
	OutcomeFullStream       BayesianOutcomeSource = "full_stream"
	OutcomeIndependentAudit BayesianOutcomeSource = "independent_audit"
	OutcomeSelected         BayesianOutcomeSource = "selected"
)

type BayesianOutcomeRequest struct {
	ProtocolVersion      string                `json:"protocol_version"`
	IdempotencyKey       string                `json:"idempotency_key"`
	TenantID             string                `json:"tenant_id"`
	JournalID            string                `json:"journal_id"`
	EventID              string                `json:"event_id"`
	Useful               bool                  `json:"useful"`
	ObservedAt           time.Time             `json:"observed_at"`
	AvailableAt          time.Time             `json:"available_at"`
	Source               BayesianOutcomeSource `json:"source"`
	InclusionProbability float64               `json:"inclusion_probability"`
}

type BayesianPosterior struct {
	TenantID               string                 `json:"tenant_id"`
	PosteriorKey           string                 `json:"posterior_key"`
	Alpha                  float64                `json:"alpha"`
	Beta                   float64                `json:"beta"`
	EffectiveSupport       float64                `json:"effective_support"`
	EvidenceEpoch          uint64                 `json:"evidence_epoch"`
	Certified              bool                   `json:"certified"`
	UpdatedAt              time.Time              `json:"updated_at"`
	ChangePointProbability float64                `json:"change_point_probability"`
	RunLengthState         BayesianRunLengthState `json:"run_length_state"`
	CalibrationWeight      float64                `json:"calibration_weight"`
	BrierLossSum           float64                `json:"brier_loss_sum"`
	ForecastUsefulSum      float64                `json:"forecast_useful_sum"`
	ObservedUsefulSum      float64                `json:"observed_useful_sum"`
}

type BayesianRunLengthState struct {
	Probabilities []float64 `json:"probabilities"`
	Alpha         []float64 `json:"alpha"`
	Beta          []float64 `json:"beta"`
}

func (posterior BayesianPosterior) Mean() float64 {
	if posterior.Alpha <= 0 || posterior.Beta <= 0 {
		return 0.5
	}
	return posterior.Alpha / (posterior.Alpha + posterior.Beta)
}

type BayesianOutcomeResponse struct {
	ProtocolVersion string            `json:"protocol_version"`
	Duplicate       bool              `json:"duplicate"`
	ChangePoint     bool              `json:"change_point"`
	Posterior       BayesianPosterior `json:"posterior"`
	Snapshot        Snapshot          `json:"snapshot"`
}
