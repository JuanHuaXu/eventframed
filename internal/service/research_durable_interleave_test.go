package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type admissionInterleaveStore struct {
	store.EventStore
	mutate bool
	reads  int
}

func (s *admissionInterleaveStore) GetEvents(ctx context.Context, tenant string, ids []string, at time.Time) ([]model.Event, error) {
	s.reads++
	events, e := s.EventStore.GetEvents(ctx, tenant, ids, at)
	if e != nil {
		return nil, e
	}
	if s.mutate {
		if _, e = s.EventStore.BindBayesianPolicy(ctx, "interleaved-policy"); e != nil {
			return nil, e
		}
	}
	return events, nil
}

func TestResearchAdmissionRejectsInterleavedPolicyChange(t *testing.T) {
	for _, persistent := range []bool{false, true} {
		name := "memory"
		if persistent {
			name = "persistent"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			em, e := embed.NewHashEmbedder(8)
			if e != nil {
				t.Fatal(e)
			}
			var backend store.EventStore = memorystore.New()
			if persistent {
				backend, e = libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/validation.libravdb", Dimension: 8, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
				if e != nil {
					t.Fatal(e)
				}
			}
			tap, e := NewResearchFrontierTap("tenant-a", 1)
			if e != nil {
				t.Fatal(e)
			}
			defer tap.Close()
			s, e := New(backend, em, Config{DefaultRecallK: 3, DefaultPackK: 2, DefaultTokenBudget: 1000, ResearchFrontier: tap})
			if e != nil {
				backend.Close()
				t.Fatal(e)
			}
			defer s.Close()
			now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
			ev := testutil.Event("public-event", "public test fact", now.Add(-time.Minute))
			ev.Embedding = []float32{1, 0, 0, 0, 0, 0, 0, 0}
			ev.EmbeddingModel = em.ModelKey()
			if _, e = s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: ev.ID, Event: ev}); e != nil {
				t.Fatal(e)
			}
			shadowRecall(t, s, now)
			wait, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			in, e := tap.Take(wait)
			if e != nil || len(in.Candidates) != 1 {
				t.Fatal(in, e)
			}
			c := in.Candidates[0]
			a := researchmemory.New(1, 42)
			p, e := a.Predict(c.Features, c.Baseline, 1, in.AsOf)
			if e != nil {
				t.Fatal(e)
			}
			r, e := a.Record(p.ID)
			if e != nil {
				t.Fatal(e)
			}
			r.Binding = &researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: c.EventID, Snapshot: in.Snapshot}
			req := model.RecallRequest{TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: ev.Embedding, EmbeddingModel: em.ModelKey(), AsOf: now}
			wrapped := &admissionInterleaveStore{EventStore: backend}
			s.store = wrapped
			if e = s.ValidateResearchAdmission(ctx, r, req); e != nil || wrapped.reads != 1 {
				t.Fatal("control did not read real event", e, wrapped.reads)
			}
			wrapped.mutate = true
			e = s.ValidateResearchAdmission(ctx, r, req)
			if e == nil || !strings.Contains(e.Error(), "dependencies changed during validation") || wrapped.reads != 2 {
				t.Fatal("interleave not rejected at final boundary", e, wrapped.reads)
			}
		})
	}
}
