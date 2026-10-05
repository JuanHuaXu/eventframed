package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

func (s *Service) enrichTurn(ctx context.Context, turn model.TurnCapture) (model.Event, error) {
	if err := ctx.Err(); err != nil {
		return model.Event{}, err
	}
	var previous *model.Event
	if turn.PreviousTurnID != "" {
		events, err := s.store.GetEvents(ctx, turn.TenantID, []string{turn.PreviousTurnID}, turn.OccurredAt)
		if err != nil && !errors.Is(err, store.ErrEventNotFound) {
			return model.Event{}, err
		}
		if len(events) == 1 && events[0].TenantID == turn.TenantID && events[0].SessionID == turn.SessionID && events[0].Kind == "agent_turn" && events[0].Sequence < turn.Sequence && events[0].OccurredAt.Before(turn.OccurredAt) && !events[0].AvailableAt.After(turn.OccurredAt) {
			previous = &events[0]
		}
	}
	reader := s.config.UserCards
	if reader == nil {
		reader, _ = s.index.(retrieval.UserCardReader)
	}
	// Bounded exact-key lookups at ingestion only. No vector search, name-based
	// roster discovery, global profile scan, or user-card writes occur here.
	lookupCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	var cards []frame.IdentityCard
	aliasesComplete := true
	statuses := make(map[string]string)
	ids := append([]string{turn.UserID}, turn.ParticipantUserIDs...)
	seen := make(map[string]bool)
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		if reader == nil {
			aliasesComplete = false
			statuses[id] = "reader-unavailable"
			continue
		}
		response, err := reader.GetUserCard(lookupCtx, turn.TenantID, retrieval.GetUserCardRequest{UserID: id})
		if err != nil {
			aliasesComplete = false
			statuses[id] = "lookup-unavailable"
			continue
		}
		if response.CardJSON == "" {
			statuses[id] = "absent"
			continue
		}
		updated := time.UnixMilli(response.UpdatedAt)
		if response.Version < 1 || response.UpdatedAt <= 0 {
			aliasesComplete = false
			statuses[id] = "unversioned"
			continue
		}
		// GetUserCard exposes only the latest version, not historical versions.
		// A replay cannot borrow aliases learned after this frame was available.
		if updated.After(turn.AvailableAt) {
			aliasesComplete = false
			statuses[id] = "future-version"
			continue
		}
		card, err := frame.ParseIdentityCard(id, response.CardJSON, response.Version, updated)
		if err != nil {
			aliasesComplete = false
			statuses[id] = "invalid-card"
			continue
		}
		cards = append(cards, card)
		statuses[id] = "available"
	}
	if err := ctx.Err(); err != nil {
		return model.Event{}, err
	}
	event := frame.FromTurnWithIdentities(turn, cards, previous, aliasesComplete)
	if turn.PreviousTurnID != "" && previous == nil {
		event.Attributes["identity_context_status"] = "unavailable-or-ineligible"
	}
	if len(statuses) > 0 {
		payload, _ := json.Marshal(statuses)
		event.Attributes["user_card_lookup_status"] = string(payload)
	}
	if len(cards) > 0 {
		// Speaker and subject may be different accounts. Keep all accepted
		// profile snapshots auditable without storing their private prose.
		links := make([]frame.WhoResolution, 0, len(cards))
		for _, card := range cards {
			links = append(links, frame.WhoResolution{Status: "linked-card", UserID: card.UserID, CardVersion: card.Version, CardHash: card.Hash, CardUpdatedAt: card.UpdatedAt})
		}
		payload, _ := json.Marshal(links)
		event.Attributes["user_card_links"] = string(payload)
	}
	return event, nil
}
