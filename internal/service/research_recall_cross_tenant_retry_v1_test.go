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

func TestResearchRecallCrossTenantRetryV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_CROSS_TENANT_RETRY_V1") != "1" {
		t.Skip("opt-in cross-tenant journal retry probe")
	}
	for _, persistent := range []bool{false, true} {
		for _, visible := range []bool{false, true} {
			t.Run(fmt.Sprintf("persistent-%v/visible-%v", persistent, visible), func(t *testing.T) {
				ctx := context.Background()
				asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
				em, err := embed.NewHashEmbedder(8)
				if err != nil {
					t.Fatal(err)
				}
				var db store.EventStore = memorystore.New()
				if persistent {
					db, err = libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/cross-tenant.libravdb", Dimension: 8, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
					if err != nil {
						t.Fatal(err)
					}
				}
				wrapped := &journalInterleaver{EventStore: db}
				s, err := New(wrapped, em, Config{DefaultRecallK: 50, DefaultPackK: 50, DefaultTokenBudget: 10000})
				if err != nil {
					db.Close()
					t.Fatal(err)
				}
				defer s.Close()
				seed := testutil.Event("tenant-a-seed", "public query", asOf.Add(-time.Minute))
				if _, err := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: seed.ID, Event: seed}); err != nil {
					t.Fatal(err)
				}
				before := db.Snapshot(ctx)
				wrapped.before = func(attempt int) {
					if attempt != 1 {
						return
					}
					at := asOf.Add(time.Second)
					if visible {
						at = asOf.Add(-time.Second)
					}
					other := testutil.Event("tenant-b-write", "public query", at)
					other.TenantID = "tenant-b"
					if _, err := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: other.ID, Event: other}); err != nil {
						t.Fatal(err)
					}
				}
				packet, err := s.Recall(ctx, model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "query", Query: "public query", AsOf: asOf, RecallK: 50, PackK: 50, TokenBudget: 10000})
				if err != nil {
					t.Fatal(err)
				}
				journal, err := db.GetBayesianJournal(ctx, "tenant-a", packet.BayesianShadow.JournalID)
				if err != nil {
					t.Fatal(err)
				}
				wantAttempts := 1
				if visible {
					wantAttempts = 2
				}
				t.Logf("visible=%v attempts=%d before_version=%d packet_version=%d journal_version=%d candidates=%d", visible, wrapped.attempts, before.RuntimeVersion, packet.Snapshot.RuntimeVersion, journal.Snapshot.RuntimeVersion, len(packet.Candidates))
				if wrapped.attempts != wantAttempts || packet.Snapshot != journal.Snapshot || len(packet.Candidates) != 1 || packet.Candidates[0].Event.ID != seed.ID {
					t.Fatal("unrelated write had unexpected retry or candidate effect")
				}
			})
		}
	}
}
