package model

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type FieldSource string

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

// Event is the durable, availability-time-aware envelope exchanged across the
// plugin/daemon boundary. Content is untrusted historical data, never an instruction.
type Event struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenant_id"`
	SessionID   string            `json:"session_id"`
	Sequence    uint64            `json:"sequence"`
	Kind        string            `json:"kind"`
	Content     string            `json:"content"`
	OccurredAt  time.Time         `json:"occurred_at"`
	ObservedAt  time.Time         `json:"observed_at"`
	AvailableAt time.Time         `json:"available_at"`
	Who         Field             `json:"who"`
	What        Field             `json:"what"`
	Where       Field             `json:"where"`
	When        Field             `json:"when"`
	Why         Field             `json:"why"`
	How         Field             `json:"how"`
	Priority    float64           `json:"priority"`
	Tags        []string          `json:"tags,omitempty"`
	Provenance  Provenance        `json:"provenance"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	Embedding   []float32         `json:"embedding,omitempty"`
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
	for name, field := range map[string]Field{
		"who": e.Who, "what": e.What, "where": e.Where,
		"when": e.When, "why": e.Why, "how": e.How,
	} {
		if err := field.Validate(); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
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
	return nil
}

func (e Event) EmbeddingText() string {
	parts := []string{e.Content, e.Who.Value, e.What.Value, e.Where.Value, e.When.Value, e.Why.Value, e.How.Value}
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
