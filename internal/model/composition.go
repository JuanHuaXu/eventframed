package model

import "time"

type RecallResolution string

const (
	RecallResolutionAuto   RecallResolution = "auto"
	RecallResolutionCoarse RecallResolution = "coarse"
	RecallResolutionFine   RecallResolution = "fine"
)

func (resolution RecallResolution) Valid() bool {
	return resolution == "" || resolution == RecallResolutionAuto || resolution == RecallResolutionCoarse || resolution == RecallResolutionFine
}

type ComposeInvariantRequest struct {
	ProtocolVersion         string    `json:"protocol_version"`
	ID                      string    `json:"id"`
	TenantID                string    `json:"tenant_id"`
	SessionID               string    `json:"session_id"`
	MemberEventIDs          []string  `json:"member_event_ids"`
	RepresentativeEventID   string    `json:"representative_event_id"`
	Label                   string    `json:"label"`
	RuleID                  string    `json:"rule_id"`
	Resolution              string    `json:"resolution"`
	Confidence              float64   `json:"confidence"`
	AntiPigeonCertificateID string    `json:"anti_pigeon_certificate_id"`
	PublishedAt             time.Time `json:"published_at"`
	BaseSnapshot            Snapshot  `json:"base_snapshot"`
}

type ComposeInvariantResponse struct {
	ProtocolVersion string   `json:"protocol_version"`
	Event           Event    `json:"event"`
	Duplicate       bool     `json:"duplicate"`
	Snapshot        Snapshot `json:"snapshot"`
}

type DecomposeInvariantRequest struct {
	ProtocolVersion string `json:"protocol_version"`
	TenantID        string `json:"tenant_id"`
	EventID         string `json:"event_id"`
	Reason          string `json:"reason"`
}

type DecomposeInvariantResponse struct {
	ProtocolVersion        string   `json:"protocol_version"`
	EventID                string   `json:"event_id"`
	Deleted                bool     `json:"deleted"`
	RestoredMemberEventIDs []string `json:"restored_member_event_ids"`
	Snapshot               Snapshot `json:"snapshot"`
}

type CompositionTombstone struct {
	TenantID                string    `json:"tenant_id"`
	EventID                 string    `json:"event_id"`
	MemberEventIDs          []string  `json:"member_event_ids"`
	AntiPigeonCertificateID string    `json:"anti_pigeon_certificate_id"`
	Reason                  string    `json:"reason"`
	DecomposedAt            time.Time `json:"decomposed_at"`
}
