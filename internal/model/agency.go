package model

import "time"

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
	NotBefore          time.Time    `json:"not_before"`
	ScheduledFor       *time.Time   `json:"scheduled_for,omitempty"`
	ExpiresAt          time.Time    `json:"expires_at"`
	IdempotencyKey     string       `json:"idempotency_key"`
	CausalChainID      string       `json:"causal_chain_id"`
	ParentProposalID   string       `json:"parent_proposal_id,omitempty"`
	CausalChainDepth   int          `json:"causal_chain_depth"`
	CreatedAt          time.Time    `json:"created_at"`
	ContractVersion    uint64       `json:"contract_version"`
}

type AgencyProposalDraft struct {
	ID               string       `json:"id"`
	TenantID         string       `json:"tenant_id"`
	SessionID        string       `json:"session_id"`
	Action           AgencyAction `json:"action"`
	Reason           string       `json:"reason"`
	EvidenceIDs      []string     `json:"evidence_ids"`
	ExpectedUtility  float64      `json:"expected_utility"`
	Priority         float64      `json:"priority"`
	NotBefore        time.Time    `json:"not_before"`
	ScheduledFor     *time.Time   `json:"scheduled_for,omitempty"`
	ExpiresAt        time.Time    `json:"expires_at"`
	IdempotencyKey   string       `json:"idempotency_key"`
	CausalChainID    string       `json:"causal_chain_id"`
	ParentProposalID string       `json:"parent_proposal_id,omitempty"`
	CausalChainDepth int          `json:"causal_chain_depth"`
}

type SignedAgencyProposal struct {
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
	KeyID     string `json:"key_id"`
}

type AgencyProposalStatus string

const (
	AgencyPending  AgencyProposalStatus = "pending"
	AgencyClaimed  AgencyProposalStatus = "claimed"
	AgencyApproved AgencyProposalStatus = "approved"
	AgencyRejected AgencyProposalStatus = "rejected"
	AgencyExpired  AgencyProposalStatus = "expired"
)

type AgencyProposalRecord struct {
	Proposal         AgencyProposal       `json:"proposal"`
	Signed           SignedAgencyProposal `json:"signed"`
	Status           AgencyProposalStatus `json:"status"`
	ClaimedBy        string               `json:"claimed_by,omitempty"`
	LeaseUntil       time.Time            `json:"lease_until,omitempty"`
	ResolutionReason string               `json:"resolution_reason,omitempty"`
	ExecutionRef     string               `json:"execution_ref,omitempty"`
	ResolvedAt       time.Time            `json:"resolved_at,omitempty"`
}

type IssueAgencyProposalRequest struct {
	ProtocolVersion string              `json:"protocol_version"`
	IssuerToken     string              `json:"issuer_token"`
	Proposal        AgencyProposalDraft `json:"proposal"`
}

type IssueAgencyProposalResponse struct {
	ProtocolVersion string               `json:"protocol_version"`
	Duplicate       bool                 `json:"duplicate"`
	Record          AgencyProposalRecord `json:"record"`
	Snapshot        Snapshot             `json:"snapshot"`
}

type ClaimAgencyProposalsRequest struct {
	ProtocolVersion string `json:"protocol_version"`
	AuthorityToken  string `json:"authority_token"`
	TenantID        string `json:"tenant_id"`
	ConsumerID      string `json:"consumer_id"`
	Limit           int    `json:"limit"`
}

type ClaimAgencyProposalsResponse struct {
	ProtocolVersion string                 `json:"protocol_version"`
	Records         []AgencyProposalRecord `json:"records"`
	Snapshot        Snapshot               `json:"snapshot"`
}

type ResolveAgencyProposalRequest struct {
	ProtocolVersion string               `json:"protocol_version"`
	AuthorityToken  string               `json:"authority_token"`
	TenantID        string               `json:"tenant_id"`
	ProposalID      string               `json:"proposal_id"`
	ConsumerID      string               `json:"consumer_id"`
	Decision        AgencyProposalStatus `json:"decision"`
	Reason          string               `json:"reason"`
	ExecutionRef    string               `json:"execution_ref,omitempty"`
}

type ResolveAgencyProposalResponse struct {
	ProtocolVersion string               `json:"protocol_version"`
	Duplicate       bool                 `json:"duplicate"`
	Record          AgencyProposalRecord `json:"record"`
	Snapshot        Snapshot             `json:"snapshot"`
}
