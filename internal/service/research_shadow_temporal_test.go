package service

import (
	"context"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
	"testing"
	"time"
)

func TestResearchShadowTemporalFallback(t *testing.T) {
	s, now := shadowService(t, ResearchShadowPolicy{})
	ctx := context.Background()
	snapshot := s.store.Snapshot(ctx)
	q := &researchShadowQueue{service: s, policy: ResearchShadowPolicy{TemporalReuse: true}}
	ev := testutil.Event("future-fallback", "public fact", now.Add(time.Second))
	if _, e := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: ev.ID, Event: ev}); e != nil {
		t.Fatal(e)
	}
	if !q.compatible(ctx, snapshot, now) {
		t.Fatal("future compatible change rejected")
	}
	if q.compatible(ctx, snapshot, time.Time{}) {
		t.Fatal("zero cutoff reused")
	}
	s.store = struct{ store.EventStore }{s.store}
	if q.compatible(ctx, snapshot, now) {
		t.Fatal("missing optional interface did not fail closed")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if q.compatible(cancelled, s.store.Snapshot(ctx), now) {
		t.Fatal("cancelled exact snapshot accepted")
	}
}

func TestResearchShadowTemporalReuse(t *testing.T) {
	for _, future := range []bool{false, true} {
		t.Run(map[bool]string{false: "backfill", true: "future"}[future], func(t *testing.T) {
			started, release := make(chan struct{}), make(chan struct{})
			s, now := shadowService(t, ResearchShadowPolicy{Enabled: true, TemporalReuse: true, Capacity: 1, MaxAge: time.Second, Process: func(ctx context.Context, in ResearchShadowInput) (float64, error) {
				close(started)
				select {
				case <-release:
					return 1, nil
				case <-ctx.Done():
					return 0, ctx.Err()
				}
			}})
			s.nominateResearchShadow(model.ContextPacket{Snapshot: s.store.Snapshot(context.Background())}, now)
			<-started
			at := now
			if future {
				at = now.Add(time.Second)
			}
			ev := testutil.Event("temporal-change", "public fact", at)
			if _, e := s.Observe(context.Background(), model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: ev.ID, Event: ev}); e != nil {
				t.Fatal(e)
			}
			close(release)
			v := awaitShadow(t, s, func(v ResearchShadowStatus) bool { return v.Completed+v.Stale == 1 })
			if future && (!v.HasResult || v.Completed != 1) {
				t.Fatal("future invalidated diagnostic")
			}
			if !future && (v.HasResult || v.Stale != 1) {
				t.Fatal("backfill accepted")
			}
		})
	}
}
