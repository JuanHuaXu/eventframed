package service

import (
	"context"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func boundWorkerBenchmarkFixture(b *testing.B) (*Service, *ResearchBoundWorker, *ResearchFrontierTap, time.Time) {
	b.Helper()
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	em, err := embed.NewHashEmbedder(8)
	if err != nil {
		b.Fatal(err)
	}
	tap, err := NewResearchFrontierTap("tenant-a", 4)
	if err != nil {
		b.Fatal(err)
	}
	s, err := New(memorystore.New(), em, Config{ResearchFrontier: tap, DefaultRecallK: 3, DefaultPackK: 2, DefaultTokenBudget: 1000})
	if err != nil {
		b.Fatal(err)
	}
	if _, err = s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: "bench-source", Event: testutil.Event("bench-source", "public query fixture", now.Add(-time.Minute))}); err != nil {
		b.Fatal(err)
	}
	wrapper, err := researchpublicationstore.Wrap(s.store)
	if err != nil {
		b.Fatal(err)
	}
	s.store = wrapper
	source, err := researchmemory.OpenDurable(ctx, b.TempDir()+"/source.sqlite", "tenant-a", "source", 1, 42)
	if err != nil {
		b.Fatal(err)
	}
	plan := ResearchBoundPlan{Target: s.store.Snapshot(ctx), Tenant: "tenant-a", Epoch: 2, Seed: 43, Cutoff: now.Add(time.Second), KeyID: "bench-key", Key: []byte("0123456789abcdef0123456789abcdef")}
	prepared, err := s.PrepareResearchBoundWorker(ctx, source, plan, "bound")
	if err != nil {
		b.Fatal(err)
	}
	openCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	worker, err := s.OpenPreparedResearchBoundWorker(openCtx, prepared, b.TempDir()+"/bound.sqlite")
	cancel()
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = worker.Close(); _ = source.Close(); tap.Close(); _ = s.Close() })
	return s, worker, tap, now
}

func boundWorkerBenchmarkObservation(b *testing.B, s *Service, tap *ResearchFrontierTap, at time.Time) (model.RecallRequest, ResearchFrontierObservation) {
	b.Helper()
	request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: s.embedder.ModelKey(), AsOf: at, RecallK: 3, PackK: 2, TokenBudget: 1000}
	if _, err := s.Recall(context.Background(), request); err != nil {
		b.Fatal(err)
	}
	takeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	observed, err := tap.Take(takeCtx)
	if err != nil || len(observed.Candidates) == 0 {
		b.Fatalf("missing benchmark candidate: %v", err)
	}
	return request, observed
}

func BenchmarkResearchBoundWorkerGuardedOperations(b *testing.B) {
	if b.N > 128 {
		b.Skip("bounded research benchmark: use -benchtime=100x")
	}
	for _, operation := range []string{"admit", "feedback"} {
		b.Run(operation, func(b *testing.B) {
			if b.N > 128 {
				b.Skip("bounded research benchmark: use -benchtime=100x")
			}
			s, worker, tap, now := boundWorkerBenchmarkFixture(b)
			b.ResetTimer()
			for i := range b.N {
				b.StopTimer()
				at := now.Add(time.Duration(2*i+2) * time.Second)
				request, observed := boundWorkerBenchmarkObservation(b, s, tap, at)
				candidate := observed.Candidates[0]
				if operation == "feedback" {
					admitCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					_, _, err := worker.Admit(admitCtx, request, observed, candidate)
					cancel()
					if err != nil {
						b.Fatal(err)
					}
				}
				opCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				b.StartTimer()
				var err error
				if operation == "admit" {
					_, _, err = worker.Admit(opCtx, request, observed, candidate)
				} else {
					_, err = worker.Feedback(opCtx, request, observed.JournalID, candidate.EventID, true, at.Add(time.Second))
				}
				b.StopTimer()
				cancel()
				if err != nil {
					b.Fatal(err)
				}
				if operation == "admit" {
					feedbackCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					_, err = worker.Feedback(feedbackCtx, request, observed.JournalID, candidate.EventID, true, at.Add(time.Second))
					cancel()
					if err != nil {
						b.Fatal(err)
					}
				}
				waitCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				if err = worker.WaitProcessed(waitCtx, uint64(i+1)); err != nil {
					b.Fatal(err)
				}
				cancel()
				b.StartTimer()
			}
		})
	}
}
