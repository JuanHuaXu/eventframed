package claimsexperiment

import (
	"context"
	"fmt"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
)

var experimentAnchor = time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC)

func observeEvent(ctx context.Context, runtime *service.Service, event model.Event) error {
	_, err := runtime.Observe(ctx, model.ObserveRequest{
		ProtocolVersion: model.ProtocolVersion,
		IdempotencyKey:  event.ID,
		Event:           event,
	})
	return err
}

func submitOutcome(ctx context.Context, runtime *service.Service, packet model.ContextPacket, eventID string, useful bool, id string, availableAt time.Time) error {
	_, err := runtime.ObserveBayesianOutcome(ctx, model.BayesianOutcomeRequest{
		ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id,
		TenantID: packetTenant(packet), JournalID: packet.BayesianShadow.JournalID,
		EventID: eventID, Useful: useful, ObservedAt: availableAt, AvailableAt: availableAt,
		Source: model.OutcomeFullStream, InclusionProbability: 1,
	})
	return err
}

func packetTenant(packet model.ContextPacket) string {
	if len(packet.Candidates) == 0 {
		return ""
	}
	return packet.Candidates[0].Event.TenantID
}

func candidateByID(packet model.ContextPacket, eventID string) (model.Candidate, error) {
	for _, candidate := range packet.Candidates {
		if candidate.Event.ID == eventID {
			return candidate, nil
		}
	}
	return model.Candidate{}, fmt.Errorf("candidate %q not found", eventID)
}

func square(value float64) float64 { return value * value }
