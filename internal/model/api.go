package model

import "time"

const ProtocolVersion = "eventframe.v1alpha1"

type Snapshot struct {
	RuntimeVersion  uint64 `json:"runtime_version"`
	PolicyVersion   uint64 `json:"policy_version"`
	ContractVersion uint64 `json:"contract_version"`
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

type RecallRequest struct {
	ProtocolVersion string    `json:"protocol_version"`
	TenantID        string    `json:"tenant_id"`
	SessionID       string    `json:"session_id"`
	Query           string    `json:"query"`
	Embedding       []float32 `json:"embedding,omitempty"`
	AsOf            time.Time `json:"as_of"`
	RecallK         int       `json:"recall_k"`
	PackK           int       `json:"pack_k"`
	TokenBudget     int       `json:"token_budget"`
}

type Candidate struct {
	Event           Event   `json:"event"`
	Similarity      float64 `json:"similarity"`
	Score           float64 `json:"score"`
	EstimatedTokens int     `json:"estimated_tokens"`
}

type ContextPacket struct {
	ProtocolVersion string      `json:"protocol_version"`
	Candidates      []Candidate `json:"candidates"`
	Recalled        int         `json:"recalled"`
	Eligible        int         `json:"eligible"`
	Packed          int         `json:"packed"`
	UsedTokens      int         `json:"used_tokens"`
	Snapshot        Snapshot    `json:"snapshot"`
}

type AgencyAction string

const (
	AgencyWake     AgencyAction = "wake"
	AgencySchedule AgencyAction = "schedule"
	AgencyNotify   AgencyAction = "notify"
	AgencyRemember AgencyAction = "remember"
	AgencySuppress AgencyAction = "suppress"
)

// AgencyProposal is deliberately data-only. The OpenClaw adapter remains the
// authority that validates capabilities and executes or rejects the proposal.
type AgencyProposal struct {
	ID                 string       `json:"id"`
	TenantID           string       `json:"tenant_id"`
	SessionID          string       `json:"session_id"`
	Action             AgencyAction `json:"action"`
	Reason             string       `json:"reason"`
	EvidenceIDs        []string     `json:"evidence_ids"`
	ExpectedUtility    float64      `json:"expected_utility"`
	Priority           float64      `json:"priority"`
	RequiredCapability string       `json:"required_capability"`
	ExpiresAt          time.Time    `json:"expires_at"`
	IdempotencyKey     string       `json:"idempotency_key"`
	CausalChainID      string       `json:"causal_chain_id"`
	ContractVersion    uint64       `json:"contract_version"`
}

type HealthResponse struct {
	ProtocolVersion string   `json:"protocol_version"`
	Status          string   `json:"status"`
	Store           string   `json:"store"`
	Dimension       int      `json:"dimension"`
	Quantization    string   `json:"quantization"`
	Snapshot        Snapshot `json:"snapshot"`
}

type ErrorResponse struct {
	ProtocolVersion string `json:"protocol_version"`
	Code            string `json:"code"`
	Message         string `json:"message"`
}
