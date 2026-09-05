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
	RuntimeEstimated     bool      `json:"runtime_estimated"`
	QueryDigest          string    `json:"query_digest,omitempty"`
	PopulationDigest     string    `json:"population_digest,omitempty"`
	ValidUntil           time.Time `json:"valid_until"`
}

type OmittedInfluenceAuditObservation struct {
	EventID     string       `json:"event_id"`
	ExpandedLaw BernoulliLaw `json:"expanded_law"`
}

type EstimateOmittedInfluenceRequest struct {
	ProtocolVersion    string                             `json:"protocol_version"`
	ID                 string                             `json:"id"`
	TenantID           string                             `json:"tenant_id"`
	JournalID          string                             `json:"journal_id"`
	PopulationEventIDs []string                           `json:"population_event_ids"`
	LocalLaw           BernoulliLaw                       `json:"local_law"`
	Observations       []OmittedInfluenceAuditObservation `json:"observations"`
	AuditSequence      uint64                             `json:"audit_sequence"`
	ConfidenceDelta    float64                            `json:"confidence_delta"`
	DivergenceLimit    float64                            `json:"divergence_limit"`
	ValidUntil         time.Time                          `json:"valid_until"`
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
	Attestation          *EvidenceAttestation  `json:"attestation,omitempty"`
	ProtocolVersion      string                `json:"protocol_version"`
	IdempotencyKey       string                `json:"idempotency_key"`
	TenantID             string                `json:"tenant_id"`
	JournalID            string                `json:"journal_id"`
	EventID              string                `json:"event_id"`
	Useful               bool                  `json:"useful"`
	Signals              OutcomeSignals        `json:"signals,omitempty"`
	ObservedAt           time.Time             `json:"observed_at"`
	AvailableAt          time.Time             `json:"available_at"`
	Source               BayesianOutcomeSource `json:"source"`
	InclusionProbability float64               `json:"inclusion_probability"`
}

type OutcomeSignals struct {
	ExplicitUseful       *bool `json:"explicit_useful,omitempty"`
	Packed               bool  `json:"packed"`
	Cited                bool  `json:"cited"`
	SuccessfulDownstream bool  `json:"successful_downstream"`
	Correction           bool  `json:"correction"`
	Rejected             bool  `json:"rejected"`
}

func (signals OutcomeSignals) Resolve(fallback bool) (useful, evidence bool) {
	if signals.Rejected || signals.Correction {
		return false, true
	}
	if signals.Cited || signals.SuccessfulDownstream {
		return true, true
	}
	if signals.ExplicitUseful != nil {
		return *signals.ExplicitUseful, true
	}
	if signals.Packed {
		return false, false
	}
	return fallback, true
}

type BayesianPosterior struct {
	EvidenceTrust           string                            `json:"evidence_trust,omitempty"`
	WorkingBelief           *WorkingBelief                    `json:"working_belief,omitempty"`
	TenantID                string                            `json:"tenant_id"`
	PosteriorKey            string                            `json:"posterior_key"`
	Alpha                   float64                           `json:"alpha"`
	Beta                    float64                           `json:"beta"`
	EffectiveSupport        float64                           `json:"effective_support"`
	EvidenceEpoch           uint64                            `json:"evidence_epoch"`
	Certified               bool                              `json:"certified"`
	UpdatedAt               time.Time                         `json:"updated_at"`
	ChangePointProbability  float64                           `json:"change_point_probability"`
	RecentChangeProbability float64                           `json:"recent_change_probability"`
	RunLengthState          BayesianRunLengthState            `json:"run_length_state"`
	DriftState              BayesianDriftState                `json:"drift_state"`
	CalibrationWeight       float64                           `json:"calibration_weight"`
	BrierLossSum            float64                           `json:"brier_loss_sum"`
	ForecastUsefulSum       float64                           `json:"forecast_useful_sum"`
	ObservedUsefulSum       float64                           `json:"observed_useful_sum"`
	MemberEvidence          map[string]BayesianMemberEvidence `json:"member_evidence,omitempty"`
}

// EvidenceAttestation authenticates an issuer's outcome assertion, not its truth.
// Parent-bearing assertions are derived evidence, not new independent trials.
type EvidenceAttestation struct {
	Issuer        string   `json:"issuer"`
	KeyID         string   `json:"key_id"`
	ObservationID string   `json:"observation_id"`
	Parents       []string `json:"parents,omitempty"`
	Signature     []byte   `json:"signature"`
}

// WorkingBelief is a bounded two-hypothesis filter, separate from the retained
// Beta/member sufficient statistics used by the Anti-Pigeon diagnostic.
type WorkingBelief struct {
	PolicyID         string  `json:"policy_id"`
	LogOdds          float64 `json:"log_odds"`
	PredictiveUseful float64 `json:"predictive_useful"`
}

type BayesianMemberEvidence struct {
	UsefulWeight    float64 `json:"useful_weight"`
	NotUsefulWeight float64 `json:"not_useful_weight"`
}

type BayesianRunLengthState struct {
	Probabilities []float64 `json:"probabilities"`
	Alpha         []float64 `json:"alpha"`
	Beta          []float64 `json:"beta"`
}

type BayesianDriftState struct {
	FastMean      float64 `json:"fast_mean"`
	SlowMean      float64 `json:"slow_mean"`
	Samples       int     `json:"samples"`
	Streak        int     `json:"streak"`
	Direction     int     `json:"direction"`
	UpwardCUSUM   float64 `json:"upward_cusum"`
	DownwardCUSUM float64 `json:"downward_cusum"`
	Cooldown      int     `json:"cooldown"`
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
	Revision        BayesianRevision  `json:"revision"`
	Posterior       BayesianPosterior `json:"posterior"`
	Snapshot        Snapshot          `json:"snapshot"`
}

type BayesianRevisionAction string

const (
	BayesianRevisionRetain          BayesianRevisionAction = "retain"
	BayesianRevisionIndividualReset BayesianRevisionAction = "individual_reset"
	BayesianRevisionSharedReset     BayesianRevisionAction = "shared_reset"
	BayesianRevisionSplit           BayesianRevisionAction = "split"
	BayesianRevisionSplitReset      BayesianRevisionAction = "split_reset"
)

// BayesianRevision records the structural consequence of an outcome update.
// A split revokes the named Anti-Pigeon sharing certificate; it does not certify
// a replacement grouping.
type BayesianRevision struct {
	Action          BayesianRevisionAction  `json:"action"`
	CertificateID   string                  `json:"certificate_id,omitempty"`
	TriggerEventID  string                  `json:"trigger_event_id,omitempty"`
	Comparison      BayesianGroupComparison `json:"comparison,omitempty"`
	ValidationBasis string                  `json:"validation_basis,omitempty"`
}

type BayesianGroupComparisonRequest struct {
	ProtocolVersion string   `json:"protocol_version"`
	TenantID        string   `json:"tenant_id"`
	MemberEventIDs  []string `json:"member_event_ids"`
}

type BayesianGroupMember struct {
	EventID             string  `json:"event_id"`
	UsefulWeight        float64 `json:"useful_weight"`
	NotUsefulWeight     float64 `json:"not_useful_weight"`
	CurrentPosteriorKey string  `json:"current_posterior_key"`
}

type BayesianGroupComparison struct {
	ProtocolVersion                 string                `json:"protocol_version"`
	TenantID                        string                `json:"tenant_id"`
	HorizonKey                      string                `json:"horizon_key"`
	Members                         []BayesianGroupMember `json:"members"`
	SharedLogEvidence               float64               `json:"shared_log_evidence"`
	SplitLogEvidence                float64               `json:"split_log_evidence"`
	SplitPosteriorProbability       float64               `json:"split_posterior_probability"`
	EquivalenceProbability          float64               `json:"equivalence_probability"`
	BorrowingWeight                 float64               `json:"borrowing_weight"`
	Recommendation                  string                `json:"recommendation"`
	SufficientSupport               bool                  `json:"sufficient_support"`
	RequiresAntiPigeonCertification bool                  `json:"requires_anti_pigeon_certification"`
	Snapshot                        Snapshot              `json:"snapshot"`
}
