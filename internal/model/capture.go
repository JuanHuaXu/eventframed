package model

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
)

const (
	maxTurnTextBytes  = 256 << 10
	maxCaptureIDBytes = 1024
	maxRetrievedIDs   = 64
)

// TurnCapture is the raw transport envelope accepted from an agent adapter.
// Semantic EventFrame fields are deliberately absent from this contract.
type TurnCapture struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	SessionID string `json:"session_id"`
	Sequence  uint64 `json:"sequence"`
	RunID     string `json:"run_id,omitempty"`
	AgentID   string `json:"agent_id,omitempty"`
	// Identity keys come from the trusted adapter/roster, never chat text. These
	// identify accounts, not an assertion that the account's statements are true.
	UserID             string    `json:"user_id,omitempty"`
	ParticipantUserIDs []string  `json:"participant_user_ids,omitempty"`
	PreviousTurnID     string    `json:"previous_turn_id,omitempty"`
	UserText           string    `json:"user_text"`
	AssistantText      string    `json:"assistant_text"`
	RetrievedIDs       []string  `json:"retrieved_ids,omitempty"`
	OccurredAt         time.Time `json:"occurred_at"`
	ObservedAt         time.Time `json:"observed_at"`
	AvailableAt        time.Time `json:"available_at"`
}

type CaptureTurnRequest struct {
	ProtocolVersion string      `json:"protocol_version"`
	IdempotencyKey  string      `json:"idempotency_key"`
	Turn            TurnCapture `json:"turn"`
}

func (c TurnCapture) Validate() error {
	for name, value := range map[string]string{
		"id": c.ID, "tenant_id": c.TenantID, "session_id": c.SessionID,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
		if len(value) > maxCaptureIDBytes {
			return fmt.Errorf("%s exceeds %d bytes", name, maxCaptureIDBytes)
		}
	}
	if strings.TrimSpace(c.UserText) == "" || strings.TrimSpace(c.AssistantText) == "" {
		return errors.New("user_text and assistant_text are required")
	}
	if len(c.UserText) > maxTurnTextBytes || len(c.AssistantText) > maxTurnTextBytes {
		return fmt.Errorf("turn text exceeds %d bytes per participant", maxTurnTextBytes)
	}
	if len(c.RunID) > maxCaptureIDBytes || len(c.AgentID) > maxCaptureIDBytes {
		return fmt.Errorf("run_id and agent_id must not exceed %d bytes", maxCaptureIDBytes)
	}
	for _, id := range []string{c.UserID, c.PreviousTurnID} {
		if len(id) > maxCaptureIDBytes || (id != "" && strings.TrimSpace(id) != id) || strings.IndexFunc(id, unicode.IsControl) >= 0 {
			return errors.New("identity and previous-turn keys must be bounded, nonblank identifiers")
		}
	}
	if c.PreviousTurnID == c.ID {
		return errors.New("previous_turn_id cannot refer to the current turn")
	}
	if len(c.ParticipantUserIDs) > 7 {
		return errors.New("participant_user_ids exceeds 7 entries")
	}
	participants := make(map[string]bool)
	for _, id := range c.ParticipantUserIDs {
		if id == "" || strings.TrimSpace(id) != id || len(id) > maxCaptureIDBytes || participants[id] || strings.IndexFunc(id, unicode.IsControl) >= 0 {
			return errors.New("participant_user_ids must contain unique bounded identifiers")
		}
		participants[id] = true
	}
	if len(c.RetrievedIDs) > maxRetrievedIDs {
		return fmt.Errorf("retrieved_ids exceeds %d entries", maxRetrievedIDs)
	}
	seen := make(map[string]struct{}, len(c.RetrievedIDs))
	for _, id := range c.RetrievedIDs {
		if strings.TrimSpace(id) == "" || len(id) > maxCaptureIDBytes {
			return errors.New("retrieved_ids contains an invalid identifier")
		}
		if _, duplicate := seen[id]; duplicate {
			return errors.New("retrieved_ids must be unique")
		}
		seen[id] = struct{}{}
	}
	if c.OccurredAt.IsZero() || c.ObservedAt.IsZero() || c.AvailableAt.IsZero() {
		return errors.New("occurred_at, observed_at, and available_at are required")
	}
	if c.ObservedAt.Before(c.OccurredAt) {
		return errors.New("observed_at cannot precede occurred_at")
	}
	if c.AvailableAt.Before(c.ObservedAt) {
		return errors.New("available_at cannot precede observed_at")
	}
	return nil
}
