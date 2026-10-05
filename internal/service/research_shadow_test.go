package service

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func shadowService(t *testing.T, p ResearchShadowPolicy) (*Service, time.Time) {
	t.Helper()
	em, e := embed.NewHashEmbedder(8)
	if e != nil {
		t.Fatal(e)
	}
	s, e := New(memorystore.New(), em, Config{DefaultRecallK: 3, DefaultPackK: 2, DefaultTokenBudget: 1000, ResearchShadow: p})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { s.Close() })
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	for i, id := range []string{"shadow-a", "shadow-b", "shadow-c"} {
		ev := testutil.Event(id, "public test fact", now.Add(-time.Minute))
		ev.Embedding = []float32{1, 0, 0, 0, 0, 0, 0, 0}
		ev.EmbeddingModel = em.ModelKey()
		ev.Priority = float64(i) / 3
		if _, e = s.Observe(context.Background(), model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: ev}); e != nil {
			t.Fatal(e)
		}
	}
	return s, now
}
func shadowRecall(t *testing.T, s *Service, now time.Time) model.ContextPacket {
	t.Helper()
	p, e := s.Recall(context.Background(), model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: s.embedder.ModelKey(), AsOf: now, RecallK: 3, PackK: 2, TokenBudget: 1000})
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func awaitShadow(t *testing.T, s *Service, done func(ResearchShadowStatus) bool) ResearchShadowStatus {
	t.Helper()
	end := time.Now().Add(2 * time.Second)
	for time.Now().Before(end) {
		v := s.ResearchShadowStatus()
		if done(v) {
			return v
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("shadow did not reach expected state")
	return ResearchShadowStatus{}
}
func TestResearchShadowPacketIsolation(t *testing.T) {
	base, now := shadowService(t, ResearchShadowPolicy{})
	seen := make(chan ResearchShadowInput, 1)
	s, _ := shadowService(t, ResearchShadowPolicy{Enabled: true, Capacity: 2, MaxAge: time.Second, Process: func(_ context.Context, in ResearchShadowInput) (float64, error) {
		seen <- in
		in.Scores[0] = 999
		return 1, nil
	}})
	a, b := shadowRecall(t, base, now), shadowRecall(t, s, now)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("shadow changed served packet")
	}
	in := <-seen
	if in.Count != b.Packed || in.Snapshot != b.Snapshot {
		t.Fatal("incorrect handoff")
	}
	v := awaitShadow(t, s, func(v ResearchShadowStatus) bool { return v.Completed == 1 })
	if !v.HasResult || v.Value != 1 {
		t.Fatal("missing shadow result")
	}
	s.researchShadow.Close()
	if closed := s.ResearchShadowStatus(); closed.HasResult || closed.Value != 0 {
		t.Fatal("closed worker exposed its old scalar")
	}
}
func TestResearchShadowBackpressureAndCancellation(t *testing.T) {
	started := make(chan struct{})
	s, _ := shadowService(t, ResearchShadowPolicy{Enabled: true, Capacity: 1, MaxAge: time.Second, Process: func(ctx context.Context, _ ResearchShadowInput) (float64, error) {
		close(started)
		<-ctx.Done()
		return 0, ctx.Err()
	}})
	p := model.ContextPacket{Snapshot: s.store.Snapshot(context.Background())}
	s.nominateResearchShadow(p)
	<-started
	s.nominateResearchShadow(p)
	s.nominateResearchShadow(p)
	v := s.ResearchShadowStatus()
	if v.Accepted != 2 || v.Dropped != 1 || v.Depth != 1 {
		t.Fatal("backpressure failure", v)
	}
	s.researchShadow.Close()
	v = s.ResearchShadowStatus()
	if !v.Closed || v.HasResult || v.Depth != 0 {
		t.Fatal("shutdown retained jobs/results", v)
	}
}
func TestResearchShadowStaleInFlight(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	s, now := shadowService(t, ResearchShadowPolicy{Enabled: true, Capacity: 1, MaxAge: time.Second, Process: func(ctx context.Context, _ ResearchShadowInput) (float64, error) {
		close(started)
		select {
		case <-release:
			return 1, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}})
	s.nominateResearchShadow(model.ContextPacket{Snapshot: s.store.Snapshot(context.Background())})
	<-started
	ev := testutil.Event("new-shadow-event", "public changed fact", now)
	if _, e := s.Observe(context.Background(), model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: ev.ID, Event: ev}); e != nil {
		t.Fatal(e)
	}
	close(release)
	v := awaitShadow(t, s, func(v ResearchShadowStatus) bool { return v.Stale == 1 })
	if v.HasResult || v.Completed != 0 {
		t.Fatal("published stale result")
	}
}
func TestResearchShadowFailureIsolation(t *testing.T) {
	for _, kind := range []string{"panic", "nan", "error", "timeout"} {
		t.Run(kind, func(t *testing.T) {
			s, _ := shadowService(t, ResearchShadowPolicy{Enabled: true, Capacity: 1, MaxAge: 10 * time.Millisecond, Process: func(ctx context.Context, _ ResearchShadowInput) (float64, error) {
				switch kind {
				case "panic":
					panic("private failure details")
				case "nan":
					return math.NaN(), nil
				case "error":
					return 0, errors.New("private error")
				default:
					<-ctx.Done()
					return 0, ctx.Err()
				}
			}})
			s.nominateResearchShadow(model.ContextPacket{Snapshot: s.store.Snapshot(context.Background())})
			v := awaitShadow(t, s, func(v ResearchShadowStatus) bool { return v.Failed+v.Stale == 1 })
			if v.HasResult {
				t.Fatal("failed job published")
			}
		})
	}
}
