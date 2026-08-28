package model

import "time"

type CompatibilityNode struct {
	ID             string   `json:"id"`
	Kind           string   `json:"kind"`
	MemberEventIDs []string `json:"member_event_ids"`
	PosteriorKeys  []string `json:"posterior_keys"`
	LawSpace       string   `json:"law_space"`
}

type CompatibilityEdge struct {
	ID            string  `json:"id"`
	From          string  `json:"from"`
	To            string  `json:"to"`
	ComparisonMap string  `json:"comparison_map"`
	Weight        float64 `json:"weight"`
}

type PredictiveGraph struct {
	TenantID     string              `json:"tenant_id"`
	Version      uint64              `json:"version"`
	Nodes        []CompatibilityNode `json:"nodes"`
	Edges        []CompatibilityEdge `json:"edges"`
	PublishedAt  time.Time           `json:"published_at"`
	SourceSnapID string              `json:"source_snap_id"`
}

type ComparisonObligation struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Weight float64 `json:"weight"`
}

type SnapBucketCertificate struct {
	NodeID            string  `json:"node_id"`
	FutureDiameterUCB float64 `json:"future_diameter_ucb"`
	DiameterLimit     float64 `json:"diameter_limit"`
	EffectiveSupport  float64 `json:"effective_support"`
}

type SnapEdgeCertificate struct {
	EdgeID      string  `json:"edge_id"`
	DefectUCB   float64 `json:"defect_ucb"`
	DefectLimit float64 `json:"defect_limit"`
}

type PredictiveSnapRequest struct {
	ProtocolVersion            string                  `json:"protocol_version"`
	ID                         string                  `json:"id"`
	TenantID                   string                  `json:"tenant_id"`
	BaseSnapshot               Snapshot                `json:"base_snapshot"`
	Candidate                  PredictiveGraph         `json:"candidate"`
	Obligations                []ComparisonObligation  `json:"obligations"`
	BucketCertificates         []SnapBucketCertificate `json:"bucket_certificates"`
	EdgeCertificates           []SnapEdgeCertificate   `json:"edge_certificates"`
	DesignStart                time.Time               `json:"design_start"`
	DesignEnd                  time.Time               `json:"design_end"`
	ConfirmationStart          time.Time               `json:"confirmation_start"`
	ConfirmationEnd            time.Time               `json:"confirmation_end"`
	CandidateFamilySize        int                     `json:"candidate_family_size"`
	UnchangedCandidateIncluded bool                    `json:"unchanged_candidate_included"`
	PriorityGainLCB            float64                 `json:"priority_gain_lcb"`
	ResourceCostUCB            float64                 `json:"resource_cost_ucb"`
	ProperRiskIncreaseUCB      float64                 `json:"proper_risk_increase_ucb"`
	SimultaneousCoverage       float64                 `json:"simultaneous_coverage"`
	Procedure                  string                  `json:"procedure"`
	Issuer                     string                  `json:"issuer"`
	ExternalAudit              bool                    `json:"external_audit"`
}

type DependencyClosure struct {
	NodeIDs       []string `json:"node_ids"`
	EdgeIDs       []string `json:"edge_ids"`
	EventIDs      []string `json:"event_ids"`
	PosteriorKeys []string `json:"posterior_keys"`
}

type PredictiveSnapRecord struct {
	ID                   string            `json:"id"`
	TenantID             string            `json:"tenant_id"`
	PreviousGraph        PredictiveGraph   `json:"previous_graph"`
	PublishedGraph       PredictiveGraph   `json:"published_graph"`
	Closure              DependencyClosure `json:"closure"`
	UnresolvedBurden     float64           `json:"unresolved_burden"`
	SimultaneousCoverage float64           `json:"simultaneous_coverage"`
	Procedure            string            `json:"procedure"`
	Issuer               string            `json:"issuer"`
	PublishedAt          time.Time         `json:"published_at"`
	RolledBack           bool              `json:"rolled_back"`
	RollbackReason       string            `json:"rollback_reason,omitempty"`
}

type PredictiveSnapResponse struct {
	ProtocolVersion string            `json:"protocol_version"`
	Accepted        bool              `json:"accepted"`
	Reason          string            `json:"reason"`
	Graph           PredictiveGraph   `json:"graph"`
	Closure         DependencyClosure `json:"closure"`
	Snapshot        Snapshot          `json:"snapshot"`
}

type PredictiveGraphResponse struct {
	ProtocolVersion string          `json:"protocol_version"`
	Graph           PredictiveGraph `json:"graph"`
	Snapshot        Snapshot        `json:"snapshot"`
}

type RollbackSnapRequest struct {
	ProtocolVersion string `json:"protocol_version"`
	TenantID        string `json:"tenant_id"`
	SnapID          string `json:"snap_id"`
	Reason          string `json:"reason"`
}
