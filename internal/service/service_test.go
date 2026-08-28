package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestObserveIsIdempotentAndRejectsConflicts(t *testing.T) {
	runtime := newMemoryService(t)
	event := testutil.Event("stable-id", "first", time.Now().UTC())
	request := model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event}

	first, err := runtime.Observe(context.Background(), request)
	if err != nil || first.Duplicate {
		t.Fatalf("first observe = %+v, %v", first, err)
	}
	second, err := runtime.Observe(context.Background(), request)
	if err != nil || !second.Duplicate {
		t.Fatalf("duplicate observe = %+v, %v", second, err)
	}
	request.Event.Content = "changed"
	if _, err := runtime.Observe(context.Background(), request); !errors.Is(err, store.ErrIdempotencyConflict) {
		t.Fatalf("conflicting observe error = %v", err)
	}
}

func TestObserveRejectsSameIDWithDifferentExplicitVector(t *testing.T) {
	runtime := newMemoryService(t)
	event := testutil.Event("vector-id", "same", time.Now().UTC())
	event.Embedding = []float32{1, 0, 0, 0, 0, 0, 0, 0}
	event.EmbeddingModel = "feature-hash-v1:d8"
	request := model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event}
	if _, err := runtime.Observe(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	request.Event.Embedding = []float32{0, 1, 0, 0, 0, 0, 0, 0}
	if _, err := runtime.Observe(context.Background(), request); !errors.Is(err, store.ErrIdempotencyConflict) {
		t.Fatalf("error = %v", err)
	}
}

func TestRecallExcludesUnavailableEventsBeforeCandidateLimit(t *testing.T) {
	runtime := newMemoryService(t)
	now := time.Now().UTC()
	for index := 0; index < 30; index++ {
		event := testutil.Event(fmt.Sprintf("future-%02d", index), "same query", now.Add(time.Hour))
		observe(t, runtime, event)
	}
	past := testutil.Event("eligible", "same query", now.Add(-time.Minute))
	observe(t, runtime, past)

	packet, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a",
		Query: "same query", AsOf: now, RecallK: 3, PackK: 3, TokenBudget: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(packet.Candidates) != 1 || packet.Candidates[0].Event.ID != "eligible" {
		t.Fatalf("candidates = %+v", packet.Candidates)
	}
}

func TestRerankingRunsBeforeIndependentPackingCap(t *testing.T) {
	now := time.Now().UTC()
	results := make([]store.SearchResult, 20)
	for index := range results {
		event := testutil.Event(fmt.Sprintf("ordinary-%02d", index), "ordinary", now.Add(-time.Minute))
		event.Priority = 0
		results[index] = store.SearchResult{Event: event, Similarity: 0.5}
	}
	answer := testutil.Event("answer", "the needed answer", now.Add(-time.Minute))
	answer.Priority = 1
	results[15] = store.SearchResult{Event: answer, Similarity: 0.5}
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(&fixedStore{results: results}, embedder, service.Config{
		DefaultRecallK: 20, DefaultPackK: 3, DefaultTokenBudget: 100, OverfetchMultiplier: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	packet, err := runtime.Recall(context.Background(), model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "other-session",
		Query: "query", AsOf: now, RecallK: 20, PackK: 3, TokenBudget: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if packet.Candidates[0].Event.ID != "answer" {
		t.Fatalf("answer was dropped before reranking: %+v", packet.Candidates)
	}
}

func newMemoryService(t *testing.T) *service.Service {
	t.Helper()
	embedder, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := service.New(memorystore.New(), embedder, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2_000, OverfetchMultiplier: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	return runtime
}

func observe(t *testing.T, runtime *service.Service, event model.Event) {
	t.Helper()
	_, err := runtime.Observe(context.Background(), model.ObserveRequest{
		ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event,
	})
	if err != nil {
		t.Fatal(err)
	}
}

type fixedStore struct{ results []store.SearchResult }

func (s *fixedStore) Put(context.Context, model.Event, []float32, string) (store.PutResult, error) {
	return store.PutResult{}, nil
}
func (s *fixedStore) Delete(context.Context, string, string) (store.DeleteResult, error) {
	return store.DeleteResult{}, nil
}
func (s *fixedStore) DeleteBefore(context.Context, string, time.Time, int) (store.RetentionResult, error) {
	return store.RetentionResult{}, nil
}
func (s *fixedStore) Backup(context.Context, string) error { return nil }
func (s *fixedStore) Compact(context.Context) error        { return nil }
func (s *fixedStore) Search(_ context.Context, _ string, _ []float32, _ time.Time, limit int) ([]store.SearchResult, error) {
	return s.results[:min(limit, len(s.results))], nil
}
func (s *fixedStore) Stats(context.Context) (store.Stats, error) {
	return store.Stats{Backend: "fixed"}, nil
}
func (s *fixedStore) Snapshot(context.Context) model.Snapshot { return model.Snapshot{} }
func (s *fixedStore) Close() error                            { return nil }
