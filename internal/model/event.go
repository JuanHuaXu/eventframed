package model

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type FieldSource string

const SemanticRepresentationVersion = "eventframe-5w1h-v1"

const (
	maxSemanticFieldBytes = 2048
	maxFieldEvidenceBytes = 1024
)

const (
	SourceObserved  FieldSource = "observed"
	SourceInferred  FieldSource = "inferred"
	SourceSynthetic FieldSource = "synthetic"
)

type Field struct {
	Value      string      `json:"value"`
	Source     FieldSource `json:"source"`
	Confidence float64     `json:"confidence"`
	Evidence   string      `json:"evidence,omitempty"`
}

type Provenance struct {
	Producer       string   `json:"producer"`
	SourceEventIDs []string `json:"source_event_ids,omitempty"`
	RetrievedIDs   []string `json:"retrieved_ids,omitempty"`
	ToolCallID     string   `json:"tool_call_id,omitempty"`
	RunID          string   `json:"run_id,omitempty"`
}

const HigherOrderEventKind = "higher_order"

// Composition binds a derived higher-order EventFrame to the constituent
// evidence that remains independently retrievable.
type Composition struct {
	MemberEventIDs          []string `json:"member_event_ids"`
	RepresentativeEventID   string   `json:"representative_event_id"`
	RuleID                  string   `json:"rule_id"`
	Resolution              string   `json:"resolution"`
	Confidence              float64  `json:"confidence"`
	AntiPigeonCertificateID string   `json:"anti_pigeon_certificate_id"`
	EvidenceEpoch           uint64   `json:"evidence_epoch"`
}

// Event is the durable, availability-time-aware envelope exchanged across the
// plugin/daemon boundary. Content is untrusted historical data, never an instruction.
type Event struct {
	ID             string            `json:"id"`
	TenantID       string            `json:"tenant_id"`
	SessionID      string            `json:"session_id"`
	Sequence       uint64            `json:"sequence"`
	Kind           string            `json:"kind"`
	Content        string            `json:"content"`
	OccurredAt     time.Time         `json:"occurred_at"`
	ObservedAt     time.Time         `json:"observed_at"`
	AvailableAt    time.Time         `json:"available_at"`
	Who            Field             `json:"who"`
	What           Field             `json:"what"`
	Where          Field             `json:"where"`
	When           Field             `json:"when"`
	Why            Field             `json:"why"`
	How            Field             `json:"how"`
	Priority       float64           `json:"priority"`
	Tags           []string          `json:"tags,omitempty"`
	Provenance     Provenance        `json:"provenance"`
	Attributes     map[string]string `json:"attributes,omitempty"`
	Embedding      []float32         `json:"embedding,omitempty"`
	EmbeddingModel string            `json:"embedding_model"`
	Composition    *Composition      `json:"composition,omitempty"`
}

func (e Event) Validate(dimension int) error {
	if strings.TrimSpace(e.ID) == "" {
		return errors.New("event id is required")
	}
	if strings.TrimSpace(e.TenantID) == "" {
		return errors.New("tenant id is required")
	}
	if strings.TrimSpace(e.SessionID) == "" {
		return errors.New("session id is required")
	}
	if strings.TrimSpace(e.Kind) == "" {
		return errors.New("event kind is required")
	}
	if strings.TrimSpace(e.Content) == "" {
		return errors.New("event content is required")
	}
	if strings.TrimSpace(e.Provenance.Producer) == "" {
		return errors.New("provenance producer is required")
	}
	if e.OccurredAt.IsZero() || e.ObservedAt.IsZero() || e.AvailableAt.IsZero() {
		return errors.New("occurred_at, observed_at, and available_at are required")
	}
	if e.ObservedAt.Before(e.OccurredAt) {
		return errors.New("observed_at cannot precede occurred_at")
	}
	if e.AvailableAt.Before(e.ObservedAt) {
		return errors.New("available_at cannot precede observed_at")
	}
	if e.Priority < 0 || e.Priority > 1 {
		return errors.New("priority must be in [0,1]")
	}
	if len(e.Embedding) != 0 && len(e.Embedding) != dimension {
		return fmt.Errorf("embedding dimension %d does not match configured dimension %d", len(e.Embedding), dimension)
	}
	if len(e.Embedding) != 0 && strings.TrimSpace(e.EmbeddingModel) == "" {
		return errors.New("embedding_model is required with an explicit embedding")
	}
	for name, field := range map[string]Field{
		"who": e.Who, "what": e.What, "where": e.Where,
		"when": e.When, "why": e.Why, "how": e.How,
	} {
		if err := field.Validate(); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	if strings.TrimSpace(e.What.Value) == "" {
		return errors.New("what is required for an EventFrame")
	}
	if e.Composition == nil && e.Kind == HigherOrderEventKind {
		return errors.New("higher-order event requires composition metadata")
	}
	if e.Composition != nil {
		if e.Kind != HigherOrderEventKind {
			return errors.New("composition metadata is reserved for higher-order events")
		}
		if err := e.Composition.Validate(e.ID, e.Provenance.SourceEventIDs); err != nil {
			return fmt.Errorf("composition: %w", err)
		}
	}
	return nil
}

func (composition Composition) Validate(eventID string, sourceEventIDs []string) error {
	if len(composition.MemberEventIDs) < 2 || len(composition.MemberEventIDs) > 64 {
		return errors.New("member_event_ids must contain between 2 and 64 events")
	}
	if strings.TrimSpace(composition.RepresentativeEventID) == "" ||
		strings.TrimSpace(composition.RuleID) == "" ||
		strings.TrimSpace(composition.Resolution) == "" ||
		strings.TrimSpace(composition.AntiPigeonCertificateID) == "" {
		return errors.New("representative_event_id, rule_id, resolution, and anti_pigeon_certificate_id are required")
	}
	if math.IsNaN(composition.Confidence) || math.IsInf(composition.Confidence, 0) || composition.Confidence < 0 || composition.Confidence > 1 {
		return errors.New("confidence must be in [0,1]")
	}
	members := make(map[string]struct{}, len(composition.MemberEventIDs))
	for _, memberID := range composition.MemberEventIDs {
		if strings.TrimSpace(memberID) == "" || memberID == eventID {
			return errors.New("member event ids must be non-empty and cannot contain the composite event")
		}
		if _, duplicate := members[memberID]; duplicate {
			return errors.New("member event ids must be unique")
		}
		members[memberID] = struct{}{}
	}
	if _, ok := members[composition.RepresentativeEventID]; !ok {
		return errors.New("representative event must be a composition member")
	}
	if len(sourceEventIDs) != len(members) {
		return errors.New("provenance source_event_ids must equal the composition members")
	}
	sources := make(map[string]struct{}, len(sourceEventIDs))
	for _, sourceID := range sourceEventIDs {
		if _, ok := members[sourceID]; !ok {
			return errors.New("provenance source_event_ids must equal the composition members")
		}
		if _, duplicate := sources[sourceID]; duplicate {
			return errors.New("provenance source_event_ids must be unique")
		}
		sources[sourceID] = struct{}{}
	}
	return nil
}

func (f Field) Validate() error {
	if f.Confidence < 0 || f.Confidence > 1 {
		return errors.New("confidence must be in [0,1]")
	}
	if f.Source != "" && f.Source != SourceObserved && f.Source != SourceInferred && f.Source != SourceSynthetic {
		return fmt.Errorf("unknown source %q", f.Source)
	}
	if strings.TrimSpace(f.Value) != "" && f.Source == "" {
		return errors.New("non-empty field requires a source")
	}
	if len(f.Value) > maxSemanticFieldBytes || len(f.Evidence) > maxFieldEvidenceBytes {
		return errors.New("field value or evidence exceeds the semantic bound")
	}
	return nil
}

// FrameText is the canonical semantic representation used by retrieval,
// ranking, and correction machinery. Raw Content is deliberately excluded.
func (e Event) FrameText() string {
	fields := [...]struct {
		name  string
		value string
	}{
		{"who", e.Who.Value}, {"what", e.What.Value}, {"where", e.Where.Value},
		{"when", e.When.Value}, {"why", e.Why.Value}, {"how", e.How.Value},
	}
	parts := []string{"representation: " + SemanticRepresentationVersion}
	for _, field := range fields {
		if value := strings.Join(strings.Fields(field.value), " "); value != "" {
			parts = append(parts, field.name+": "+value)
		}
	}
	return strings.Join(parts, "\n")
}

func (e Event) MeanFieldConfidence() float64 {
	fields := [...]Field{e.Who, e.What, e.Where, e.When, e.Why, e.How}
	var total float64
	var count float64
	for _, field := range fields {
		if field.Value == "" {
			continue
		}
		total += field.Confidence
		count++
	}
	if count == 0 {
		return 0.5
	}
	return total / count
}
