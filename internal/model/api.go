package model

import "time"

const ProtocolVersion = "eventframe.v1alpha1"
const ContractVersion uint64 = 16

type Snapshot struct {
	RuntimeVersion     uint64 `json:"runtime_version"`
	PolicyVersion      uint64 `json:"policy_version"`
	ContractVersion    uint64 `json:"contract_version"`
	GraphVersion       uint64 `json:"graph_version"`
	PosteriorVersion   uint64 `json:"posterior_version"`
	ResidualVersion    uint64 `json:"residual_version"`
	AbstractionVersion uint64 `json:"abstraction_version"`
	AgencyVersion      uint64 `json:"agency_version"`
	EvidenceEpoch      uint64 `json:"evidence_epoch"`
}

type ObserveRequest struct {
	ProtocolVersion string `json:"protocol_version"`
	IdempotencyKey  string `json:"idempotency_key"`
	Event           Event  `json:"event"`
}

type ObserveResponse struct {
	ProtocolVersion string   `json:"protocol_version"`
	EventID         string   `json:"event_id"`
	Duplicate       bool     `json:"duplicate"`
	Snapshot        Snapshot `json:"snapshot"`
}

type DeleteRequest struct {
	ProtocolVersion string `json:"protocol_version"`
	TenantID        string `json:"tenant_id"`
	EventID         string `json:"event_id"`
}
type DeleteResponse struct {
	ProtocolVersion string   `json:"protocol_version"`
	EventID         string   `json:"event_id"`
	Deleted         bool     `json:"deleted"`
	Snapshot        Snapshot `json:"snapshot"`
}
type RetentionRequest struct {
	ProtocolVersion string    `json:"protocol_version"`
	TenantID        string    `json:"tenant_id"`
	Before          time.Time `json:"before"`
	Limit           int       `json:"limit"`
}
type RetentionResponse struct {
	ProtocolVersion string   `json:"protocol_version"`
	DeletedIDs      []string `json:"deleted_ids"`
	Snapshot        Snapshot `json:"snapshot"`
}
type BackupRequest struct {
	ProtocolVersion string `json:"protocol_version"`
	Destination     string `json:"destination"`
}
type MaintenanceResponse struct {
	ProtocolVersion string   `json:"protocol_version"`
	Operation       string   `json:"operation"`
	Snapshot        Snapshot `json:"snapshot"`
}

type RecallRequest struct {
	ProtocolVersion              string              `json:"protocol_version"`
	TenantID                     string              `json:"tenant_id"`
	SessionID                    string              `json:"session_id"`
	Query                        string              `json:"query"`
	Embedding                    []float32           `json:"embedding,omitempty"`
	EmbeddingModel               string              `json:"embedding_model,omitempty"`
	AsOf                         time.Time           `json:"as_of"`
	RecallK                      int                 `json:"recall_k"`
	PackK                        int                 `json:"pack_k"`
	TokenBudget                  int                 `json:"token_budget"`
	RetrievalCollections         []string            `json:"retrieval_collections,omitempty"`
	RetrievalExcludeByCollection map[string][]string `json:"retrieval_exclude_by_collection,omitempty"`
	Resolution                   RecallResolution    `json:"resolution,omitempty"`
}

type Candidate struct {
	Event                          Event   `json:"event"`
	Similarity                     float64 `json:"similarity"`
	RecencyScore                   float64 `json:"recency_score"`
	GraphCompatibility             float64 `json:"graph_compatibility"`
	GraphApplied                   bool    `json:"graph_applied"`
	BaselineScore                  float64 `json:"baseline_score"`
	PredictiveScore                float64 `json:"predictive_score"`
	RetrievalScore                 float64 `json:"retrieval_score"`
	RankDelta                      float64 `json:"rank_delta"`
	RankDeltaScale                 float64 `json:"rank_delta_scale,omitempty"`
	RankDeltaAnswerCertainty       float64 `json:"rank_delta_answer_certainty,omitempty"`
	RankDeltaCorrectionReliability float64 `json:"rank_delta_correction_reliability,omitempty"`
	// RankDeltaConfidence is a deprecated alias for answer certainty.
	RankDeltaConfidence float64 `json:"rank_delta_confidence,omitempty"`
	RankDeltaBasis      string  `json:"rank_delta_basis,omitempty"`
	ResolutionRankDelta float64 `json:"resolution_rank_delta,omitempty"`
	RetrievalContract   string  `json:"retrieval_contract"`
	RetrievalMetadata   []byte  `json:"-"`
	Score               float64 `json:"score"`
	BayesianProbability float64 `json:"bayesian_probability,omitempty"`
	BayesianApplied     bool    `json:"bayesian_applied"`
	// EvidenceGroupKey is an opaque daemon-derived claim/lineage identifier.
	// It exposes no source text and prevents repeated records from monopolizing
	// a context packet while preserving every durable EventFrame.
	EvidenceGroupKey string         `json:"-"`
	Forecast         ForecastBundle `json:"forecast"`
	EstimatedTokens  int            `json:"estimated_tokens"`
}

type ContextPacket struct {
	ProtocolVersion       string               `json:"protocol_version"`
	Candidates            []Candidate          `json:"candidates"`
	Recalled              int                  `json:"recalled"`
	Eligible              int                  `json:"eligible"`
	Packed                int                  `json:"packed"`
	UsedTokens            int                  `json:"used_tokens"`
	AdaptiveExpanded      bool                 `json:"adaptive_expanded"`
	CorrelatedSuppressed  int                  `json:"correlated_suppressed"`
	PacketConfidence      float64              `json:"packet_confidence"`
	PacketAnswerCertainty float64              `json:"packet_answer_certainty"`
	RetrievalContract     string               `json:"retrieval_contract"`
	NominationContract    string               `json:"nomination_contract"`
	Snapshot              Snapshot             `json:"snapshot"`
	BayesianShadow        BayesianShadowReport `json:"bayesian_shadow"`
}

type BayesianDecision struct {
	EventID                             string         `json:"event_id"`
	ActivationScore                     float64        `json:"activation_score"`
	Activated                           bool           `json:"activated"`
	CheapUpdate                         bool           `json:"cheap_update"`
	DeepReview                          bool           `json:"deep_review"`
	EvidenceReady                       bool           `json:"evidence_ready"`
	AuditSelected                       bool           `json:"audit_selected"`
	AuditProbability                    float64        `json:"audit_probability"`
	NominationProbabilityLowerBound     float64        `json:"nomination_probability_lower_bound"`
	ActivationProbability               float64        `json:"activation_probability"`
	TotalSelectionProbabilityLowerBound float64        `json:"total_selection_probability_lower_bound"`
	PosteriorKey                        string         `json:"posterior_key"`
	ParentPosteriorKey                  string         `json:"parent_posterior_key,omitempty"`
	EvidenceGroupKey                    string         `json:"-"`
	CorrelatedSuppressed                bool           `json:"correlated_suppressed"`
	Forecast                            ForecastBundle `json:"forecast"`
}

type BayesianShadowReport struct {
	Mode                      string             `json:"mode"`
	JournalID                 string             `json:"journal_id"`
	JournalDurable            bool               `json:"journal_durable"`
	Nominated                 int                `json:"nominated"`
	Activated                 int                `json:"activated"`
	DeepReviewed              int                `json:"deep_reviewed"`
	SelectionSupportCertified bool               `json:"selection_support_certified"`
	OmittedInfluenceCertified bool               `json:"omitted_influence_certified"`
	ResidualApplied           int                `json:"residual_applied"`
	ResidualShadowEligible    int                `json:"residual_shadow_eligible"`
	Decisions                 []BayesianDecision `json:"decisions"`
}

type AgencyAction string

const (
	AgencyWake     AgencyAction = "wake"
	AgencySchedule AgencyAction = "schedule"
	AgencyNotify   AgencyAction = "notify"
	AgencyRemember AgencyAction = "remember"
	AgencySuppress AgencyAction = "suppress"
)

type HealthResponse struct {
	ProtocolVersion        string   `json:"protocol_version"`
	Status                 string   `json:"status"`
	Store                  string   `json:"store"`
	Dimension              int      `json:"dimension"`
	Quantization           string   `json:"quantization"`
	NominationContract     string   `json:"nomination_contract"`
	RetrievalContract      string   `json:"retrieval_contract"`
	ExternalCandidateIndex bool     `json:"external_candidate_index"`
	Snapshot               Snapshot `json:"snapshot"`
}

type ReadinessResponse struct {
	ProtocolVersion        string   `json:"protocol_version"`
	Status                 string   `json:"status"`
	StoreReady             bool     `json:"store_ready"`
	ExternalContractsReady bool     `json:"external_contracts_ready"`
	ExternalError          string   `json:"external_error,omitempty"`
	Snapshot               Snapshot `json:"snapshot"`
}

type ErrorResponse struct {
	ProtocolVersion string `json:"protocol_version"`
	Code            string `json:"code"`
	Message         string `json:"message"`
}
