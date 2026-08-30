package invariant

import (
	"fmt"
	"strings"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const Producer = "eventframed-invariant-seeker"

func Compose(request model.ComposeInvariantRequest, members []model.Event) model.Event {
	occurredAt, latestOccurredAt := members[0].OccurredAt, members[0].OccurredAt
	observedAt, sequence, priority := members[0].ObservedAt, members[0].Sequence, members[0].Priority
	for _, member := range members[1:] {
		if member.OccurredAt.Before(occurredAt) {
			occurredAt = member.OccurredAt
		}
		if member.OccurredAt.After(latestOccurredAt) {
			latestOccurredAt = member.OccurredAt
		}
		if member.ObservedAt.After(observedAt) {
			observedAt = member.ObservedAt
		}
		sequence = max(sequence, member.Sequence)
		priority = max(priority, member.Priority)
	}
	memberIDs := append([]string(nil), request.MemberEventIDs...)
	return model.Event{
		ID: request.ID, TenantID: request.TenantID, SessionID: request.SessionID,
		Sequence: sequence + 1, Kind: model.HigherOrderEventKind, Content: request.Label,
		OccurredAt: occurredAt, ObservedAt: observedAt, AvailableAt: request.PublishedAt.UTC(),
		Who:      commonField(members, func(event model.Event) model.Field { return event.Who }),
		What:     model.Field{Value: request.Label, Source: model.SourceInferred, Confidence: request.Confidence, Evidence: "higher-order invariant " + request.RuleID},
		Where:    commonField(members, func(event model.Event) model.Field { return event.Where }),
		When:     model.Field{Value: intervalLabel(occurredAt, latestOccurredAt), Source: model.SourceInferred, Confidence: request.Confidence, Evidence: "constituent time envelope"},
		Why:      commonField(members, func(event model.Event) model.Field { return event.Why }),
		How:      commonField(members, func(event model.Event) model.Field { return event.How }),
		Priority: priority, Tags: []string{"higher-order", request.Resolution},
		Provenance: model.Provenance{Producer: Producer, SourceEventIDs: memberIDs},
		Composition: &model.Composition{
			MemberEventIDs: memberIDs, RepresentativeEventID: request.RepresentativeEventID,
			RuleID: request.RuleID, Resolution: request.Resolution, Confidence: request.Confidence,
			AntiPigeonCertificateID: request.AntiPigeonCertificateID, EvidenceEpoch: request.BaseSnapshot.EvidenceEpoch,
		},
	}
}

func commonField(members []model.Event, selectField func(model.Event) model.Field) model.Field {
	first := selectField(members[0])
	canonical := strings.Join(strings.Fields(strings.ToLower(first.Value)), " ")
	if canonical == "" {
		return model.Field{}
	}
	confidence := first.Confidence
	for _, member := range members[1:] {
		field := selectField(member)
		if strings.Join(strings.Fields(strings.ToLower(field.Value)), " ") != canonical {
			return model.Field{}
		}
		confidence = min(confidence, field.Confidence)
	}
	return model.Field{Value: first.Value, Source: model.SourceInferred, Confidence: confidence, Evidence: fmt.Sprintf("shared across %d constituent EventFrames", len(members))}
}

func intervalLabel(start, end time.Time) string {
	if start.Equal(end) {
		return start.UTC().Format(time.RFC3339Nano)
	}
	return start.UTC().Format(time.RFC3339Nano) + " / " + end.UTC().Format(time.RFC3339Nano)
}
