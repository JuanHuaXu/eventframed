package testutil

import (
	"fmt"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func Event(id, content string, availableAt time.Time) model.Event {
	observedAt := availableAt.Add(-time.Second)
	occurredAt := observedAt.Add(-time.Second)
	observed := func(value string) model.Field {
		return model.Field{Value: value, Source: model.SourceObserved, Confidence: 1}
	}
	return model.Event{
		ID: id, TenantID: "tenant-a", SessionID: "session-a", Sequence: uint64(max(0, occurredAt.UnixMilli())),
		Kind: "test", Content: content, OccurredAt: occurredAt, ObservedAt: observedAt, AvailableAt: availableAt,
		Who: observed("tester"), What: observed(content), Where: observed("test"), When: observed(occurredAt.Format(time.RFC3339Nano)),
		Why: model.Field{Value: "test fixture", Source: model.SourceSynthetic, Confidence: 1},
		How: observed("test"), Priority: 0.5,
		// Each fixture ID represents an independently sourced observation unless
		// a test deliberately replaces this provenance to model repetition.
		Provenance: model.Provenance{Producer: "test", SourceEventIDs: []string{id}},
		Attributes: map[string]string{"fixture": fmt.Sprintf("%s:%s", id, content)},
	}
}
