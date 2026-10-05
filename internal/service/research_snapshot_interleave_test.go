package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type journalInterleaver struct {
	store.EventStore
	attempts int
	before   func(int)
}

func (s *journalInterleaver) PutBayesianJournal(ctx context.Context, e model.BayesianJournalEntry) error {
	s.attempts++
	s.before(s.attempts)
	return s.EventStore.PutBayesianJournal(ctx, e)
}

// Deliberately place an ingestion between forecast assembly and journal commit.
// This identifies the temporal cause without a timing-sensitive concurrency race.
func TestResearchSnapshotInterleavings(t *testing.T) {
	for _, persistent := range []bool{false, true} {
		for _, tc := range []struct {
			name             string
			offset           time.Duration
			writes, attempts int
			stale            bool
		}{
			{"future", time.Second, 5, 1, false},
			{"past-sustained", -time.Second, 5, 5, true},
			{"past-settles", -time.Second, 3, 4, false},
		} {
			t.Run(fmt.Sprintf("persistent%v/%s", persistent, tc.name), func(t *testing.T) {
				ctx := context.Background()
				now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
				em, e := embed.NewHashEmbedder(8)
				if e != nil {
					t.Fatal(e)
				}
				var db store.EventStore = memorystore.New()
				if persistent {
					db, e = libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/interleave.libravdb", Dimension: 8, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
					if e != nil {
						t.Fatal(e)
					}
				}
				s, e := New(db, em, Config{DefaultRecallK: 50, DefaultPackK: 50, DefaultTokenBudget: 10000})
				if e != nil {
					t.Fatal(e)
				}
				defer s.Close()
				put := func(id string, at time.Time) {
					ev := testutil.Event(id, "public query", at)
					if _, e := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: ev}); e != nil {
						t.Fatal(e)
					}
				}
				put("seed", now.Add(-time.Minute))
				wrapped := &journalInterleaver{EventStore: db, before: func(attempt int) {
					if attempt <= tc.writes {
						put(fmt.Sprintf("inserted-%d", attempt), now.Add(tc.offset))
					}
				}}
				s.store = wrapped
				packet, e := s.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "query", Query: "public query", AsOf: now, RecallK: 50, PackK: 50, TokenBudget: 10000})
				if errors.Is(e, store.ErrStaleSnapshot) != tc.stale || wrapped.attempts != tc.attempts {
					t.Fatalf("attempts=%d error=%v", wrapped.attempts, e)
				}
				if !tc.stale && e != nil {
					t.Fatal(e)
				}
				if !tc.stale {
					want := 1
					if tc.offset < 0 {
						want += tc.writes
					}
					if len(packet.Candidates) != want {
						t.Fatalf("visible records=%d want%d", len(packet.Candidates), want)
					}
					for _, c := range packet.Candidates {
						if c.Event.AvailableAt.After(now) {
							t.Fatal("future evidence leaked")
						}
					}
				}
			})
		}
	}
}
