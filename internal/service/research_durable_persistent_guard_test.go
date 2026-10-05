package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestResearchGuardedAdmissionPersistentBoundary(t *testing.T) {
	for _, scenario := range []struct{ persistent, future bool }{{false, false}, {true, false}, {false, true}, {true, true}} {
		name := "memory"
		if scenario.persistent {
			name = "persistent"
		}
		if scenario.future {
			name += "/future-ingest"
		} else {
			name += "/policy"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			em, err := embed.NewHashEmbedder(8)
			if err != nil {
				t.Fatal(err)
			}
			var backend store.EventStore = memorystore.New()
			if scenario.persistent {
				backend, err = libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/guard.libravdb", Dimension: 8, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
				if err != nil {
					t.Fatal(err)
				}
			}
			wrapped, err := researchpublicationstore.Wrap(backend)
			if err != nil {
				backend.Close()
				t.Fatal(err)
			}
			tap, err := NewResearchFrontierTap("tenant-a", 1)
			if err != nil {
				wrapped.Close()
				t.Fatal(err)
			}
			defer tap.Close()
			s, err := New(wrapped, em, Config{DefaultRecallK: 3, DefaultPackK: 2, DefaultTokenBudget: 1000, ResearchFrontier: tap})
			if err != nil {
				wrapped.Close()
				t.Fatal(err)
			}
			defer s.Close()
			now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
			event := testutil.Event("guard-public-event", "public test fact", now.Add(-time.Minute))
			event.Embedding = []float32{1, 0, 0, 0, 0, 0, 0, 0}
			event.EmbeddingModel = em.ModelKey()
			if _, err = s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event}); err != nil {
				t.Fatal(err)
			}
			shadowRecall(t, s, now)
			in, err := tap.Take(ctx)
			if err != nil || len(in.Candidates) != 1 {
				t.Fatal(in, err)
			}
			candidate := in.Candidates[0]
			a := researchmemory.New(1, 42)
			prediction, err := a.Predict(candidate.Features, candidate.Baseline, 1, in.AsOf)
			if err != nil {
				t.Fatal(err)
			}
			record, err := a.Record(prediction.ID)
			if err != nil {
				t.Fatal(err)
			}
			record.Binding = &researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: candidate.EventID, Snapshot: in.Snapshot}
			request := model.RecallRequest{TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: event.Embedding, EmbeddingModel: em.ModelKey(), AsOf: now}
			path := t.TempDir() + "/admission.sqlite"
			durable, err := researchmemory.OpenDurable(ctx, path, "tenant-a", "stream", 1, 42)
			if err != nil {
				t.Fatal(err)
			}
			defer durable.Close()
			started, done := make(chan struct{}), make(chan error, 1)
			called := false
			err = s.WithValidatedResearchAdmission(ctx, record, request, func() error {
				called = true
				go func() {
					close(started)
					var writeErr error
					if scenario.future {
						future := testutil.Event("future-guard-event", "public future fact", now.Add(time.Hour))
						_, writeErr = s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: future.ID, Event: future})
					} else {
						_, writeErr = s.store.BindBayesianPolicy(ctx, "after-persistent-guard")
					}
					done <- writeErr
				}()
				<-started
				select {
				case writeErr := <-done:
					return errors.Join(errors.New("writer completed inside guard"), writeErr)
				case <-time.After(20 * time.Millisecond):
				}
				if s.store.Snapshot(ctx) != record.Binding.Snapshot {
					return errors.New("snapshot changed before ledger commit")
				}
				_, _, writeErr := durable.AdmitBound(ctx, prediction.ID, candidate.Features, candidate.Baseline, in.AsOf, *record.Binding)
				return writeErr
			})
			if err != nil || !called {
				t.Fatal("guarded commit failed", err)
			}
			select {
			case err = <-done:
				if err != nil {
					t.Fatal(err)
				}
			case <-ctx.Done():
				t.Fatal("writer did not resume", ctx.Err())
			}
			if s.store.Snapshot(ctx) == record.Binding.Snapshot {
				t.Fatal("writer did not advance snapshot")
			}
			if scenario.future {
				// Future-only motion preserves as-of validity, but the current
				// exact-snapshot guard deliberately has a stricter acceptance rule.
				if err = s.ValidateResearchAdmission(ctx, record, request); err != nil {
					t.Fatal("future-only point validation rejected", err)
				}
			}
			if err = durable.Close(); err != nil {
				t.Fatal(err)
			}
			reopened, err := researchmemory.OpenDurable(ctx, path, "tenant-a", "stream", 1, 42)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			saved, err := reopened.Admission(ctx, prediction.ID)
			if err != nil || !reflect.DeepEqual(saved, record) {
				t.Fatal("reopen changed original bound record", err, saved, record)
			}
			called = false
			err = s.WithValidatedResearchAdmission(ctx, saved, request, func() error { called = true; return nil })
			if err == nil || called {
				t.Fatal("reopened ledger granted stale service authority", err)
			}
			called = false
			err = s.WithValidatedResearchAsOfAdmission(ctx, saved, request, func() error { called = true; return nil })
			if scenario.future {
				if err != nil || !called {
					t.Fatal("as-of guarded admission rejected proven future motion", err)
				}
				wrongTime := request
				wrongTime.AsOf = now.Add(time.Minute)
				if err = s.WithValidatedResearchAsOfAdmission(ctx, saved, wrongTime, func() error { t.Error("wrong request time accepted"); return nil }); err == nil {
					t.Fatal("request time binding missing")
				}
			} else if err == nil || called {
				t.Fatal("as-of guard accepted policy change", err)
			}
			called = false
			err = s.WithValidatedResearchBatchAdmission(ctx, []researchmemory.RecordedPrediction{saved}, request, func() error { called = true; return nil })
			if scenario.future {
				if err != nil || !called {
					t.Fatal("batch guard rejected future history", err)
				}
				bad := saved
				bad.Prediction.Features ^= 1
				if err = s.WithValidatedResearchBatchAdmission(ctx, []researchmemory.RecordedPrediction{bad}, request, func() error { t.Error("invalid batch callback entered"); return nil }); err == nil {
					t.Fatal("bad batch accepted")
				}
			} else if err == nil || called {
				t.Fatal("batch guard accepted policy change", err)
			}
		})
	}
}
