package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type searchSnapshotInterleaverV1 struct {
	store.EventStore
	afterSearch     func()
	attempts        int
	pinBeforeSearch bool
	pinned          model.Snapshot
	pending         bool
}

func (s *searchSnapshotInterleaverV1) Search(ctx context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]store.SearchResult, error) {
	if s.pinBeforeSearch {
		s.pinned = s.EventStore.Snapshot(ctx)
		s.pending = true
	}
	results, err := s.EventStore.Search(ctx, tenantID, vector, availableBy, limit)
	s.attempts++
	if err == nil && s.afterSearch != nil {
		after := s.afterSearch
		s.afterSearch = nil
		after()
	}
	return results, err
}

func (s *searchSnapshotInterleaverV1) Snapshot(ctx context.Context) model.Snapshot {
	if s.pending {
		s.pending = false
		return s.pinned
	}
	return s.EventStore.Snapshot(ctx)
}

func TestResearchRecallSearchSnapshotV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_SEARCH_SNAPSHOT_V1") != "1" {
		t.Skip("opt-in search/snapshot interleaving diagnostic")
	}
	for _, persistent := range []bool{false, true} {
		t.Run(fmt.Sprintf("persistent-%v", persistent), func(t *testing.T) {
			ctx := context.Background()
			asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
			em, err := embed.NewHashEmbedder(8)
			if err != nil {
				t.Fatal(err)
			}
			var db store.EventStore = memorystore.New()
			if persistent {
				db, err = libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/search-snapshot.libravdb", Dimension: 8, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
				if err != nil {
					t.Fatal(err)
				}
			}
			wrapped := &searchSnapshotInterleaverV1{EventStore: db}
			s, err := New(wrapped, em, Config{DefaultRecallK: 50, DefaultPackK: 50, DefaultTokenBudget: 10000})
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			put := func(id string) {
				event := testutil.Event(id, "public query", asOf.Add(-time.Minute))
				if _, err := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: event}); err != nil {
					t.Fatal(err)
				}
			}
			put("seed")
			before := db.Snapshot(ctx)
			wrapped.afterSearch = func() { put("interleaved") }
			packet, err := s.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "query", Query: "public query", AsOf: asOf, RecallK: 50, PackK: 50, TokenBudget: 10000})
			if err != nil {
				t.Fatal(err)
			}
			journal, err := db.GetBayesianJournal(ctx, "tenant-a", packet.BayesianShadow.JournalID)
			if err != nil {
				t.Fatal(err)
			}
			seen := make(map[string]bool)
			for _, candidate := range packet.Candidates {
				seen[candidate.Event.ID] = true
			}
			t.Logf("search_attempts=%d before_version=%d packet_version=%d journal_version=%d candidates=%v", wrapped.attempts, before.RuntimeVersion, packet.Snapshot.RuntimeVersion, journal.Snapshot.RuntimeVersion, seen)
			if packet.Snapshot.RuntimeVersion > before.RuntimeVersion && !seen["interleaved"] {
				t.Error("packet and durable frontier journal claim a newer snapshot than their searched candidate set")
			}
		})
	}
}

func TestResearchRecallPreSearchSnapshotV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_PRESEARCH_SNAPSHOT_V1") != "1" {
		t.Skip("opt-in pre-search snapshot ordering probe")
	}
	for _, persistent := range []bool{false, true} {
		t.Run(fmt.Sprintf("persistent-%v", persistent), func(t *testing.T) {
			ctx := context.Background()
			asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
			em, err := embed.NewHashEmbedder(8)
			if err != nil {
				t.Fatal(err)
			}
			var db store.EventStore = memorystore.New()
			if persistent {
				db, err = libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/presearch-snapshot.libravdb", Dimension: 8, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
				if err != nil {
					t.Fatal(err)
				}
			}
			wrapped := &searchSnapshotInterleaverV1{EventStore: db, pinBeforeSearch: true}
			s, err := New(wrapped, em, Config{DefaultRecallK: 50, DefaultPackK: 50, DefaultTokenBudget: 10000})
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			put := func(id string) {
				event := testutil.Event(id, "public query", asOf.Add(-time.Minute))
				if _, err := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: event}); err != nil {
					t.Fatal(err)
				}
			}
			put("seed")
			wrapped.afterSearch = func() { put("interleaved") }
			packet, err := s.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "query", Query: "public query", AsOf: asOf, RecallK: 50, PackK: 50, TokenBudget: 10000})
			if err != nil {
				t.Fatal(err)
			}
			journal, err := db.GetBayesianJournal(ctx, "tenant-a", packet.BayesianShadow.JournalID)
			if err != nil {
				t.Fatal(err)
			}
			seen := make(map[string]bool)
			for _, candidate := range packet.Candidates {
				seen[candidate.Event.ID] = true
			}
			t.Logf("search_attempts=%d packet_version=%d journal_version=%d candidates=%v", wrapped.attempts, packet.Snapshot.RuntimeVersion, journal.Snapshot.RuntimeVersion, seen)
			if wrapped.attempts != 2 || len(seen) != 2 || !seen["seed"] || !seen["interleaved"] ||
				packet.Snapshot != db.Snapshot(ctx) || journal.Snapshot != packet.Snapshot {
				t.Fatal("pre-search pin did not retry into a coherent candidate snapshot")
			}
		})
	}
}
