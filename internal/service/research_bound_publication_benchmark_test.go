package service

import (
	"context"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

var researchBoundBenchmarkSink float64

func boundPublicationBenchmarkFixture(b *testing.B) (*Service, *researchmemory.Durable, *ResearchBoundSlot, *PreparedResearchBound, ResearchBoundPlan, uint16, float64) {
	b.Helper()
	ctx := context.Background()
	em, err := embed.NewHashEmbedder(8)
	if err != nil {
		b.Fatal(err)
	}
	wrapped, err := researchpublicationstore.Wrap(memorystore.New())
	if err != nil {
		b.Fatal(err)
	}
	tap, err := NewResearchFrontierTap("tenant-a", 4)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(tap.Close)
	s, err := New(wrapped, em, Config{ResearchFrontier: tap, DefaultRecallK: 3, DefaultPackK: 2, DefaultTokenBudget: 1000})
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = s.Close() })
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	event := testutil.Event("source", "public query fixture", now.Add(-time.Minute))
	if _, err := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event}); err != nil {
		b.Fatal(err)
	}
	request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: em.ModelKey(), AsOf: now, RecallK: 3, PackK: 2, TokenBudget: 1000}
	if _, err := s.Recall(ctx, request); err != nil {
		b.Fatal(err)
	}
	takeCtx, cancel := context.WithTimeout(ctx, time.Second)
	in, err := tap.Take(takeCtx)
	cancel()
	if err != nil || len(in.Candidates) != 1 {
		b.Fatalf("missing benchmark frontier: %v", err)
	}
	candidate := in.Candidates[0]
	journal, err := s.store.GetBayesianJournal(ctx, "tenant-a", in.JournalID)
	if err != nil {
		b.Fatal(err)
	}
	d, err := researchmemory.OpenDurable(ctx, filepath.Join(b.TempDir(), "labels.sqlite"), "tenant-a", "benchmark", 1, 42)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = d.Close() })
	key := []byte("0123456789abcdef0123456789abcdef")
	binding := researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: candidate.EventID, Snapshot: in.Snapshot}
	preview := researchmemory.RecordedPrediction{Contract: researchmemory.RecordContract, Seed: 42, Prediction: researchmemory.Prediction{ID: 1, Probability: candidate.Baseline, Features: candidate.Features, Epoch: 1}, At: in.AsOf, Outer: [4]float64{candidate.Baseline}, Binding: &binding}
	err = s.WithValidatedResearchAdmission(ctx, preview, request, func() error {
		witness, witnessErr := researchmemory.NewSourceWitness("lab-v1", key, binding, journal.QueryDigest, event, candidate.Features, candidate.Baseline)
		if witnessErr != nil {
			return witnessErr
		}
		_, _, admitErr := d.AdmitBoundWithWitness(ctx, 1, candidate.Features, candidate.Baseline, in.AsOf, binding, witness)
		return admitErr
	})
	if err != nil {
		b.Fatal(err)
	}
	original, err := d.Admission(ctx, 1)
	if err != nil {
		b.Fatal(err)
	}
	guardCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	err = s.WithValidatedResearchAsOfAdmission(guardCtx, original, request, func() error {
		_, writeErr := d.Feedback(guardCtx, 1, true, now.Add(time.Second))
		return writeErr
	})
	cancel()
	if err != nil {
		b.Fatal(err)
	}
	waitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	if err := d.WaitProcessed(waitCtx, 1); err != nil {
		b.Fatal(err)
	}
	cancel()
	plan := ResearchBoundPlan{Target: s.store.Snapshot(ctx), Tenant: "tenant-a", Epoch: 2, Seed: 43, Cutoff: now.Add(2 * time.Second), KeyID: "lab-v1", Key: key}
	slot := &ResearchBoundSlot{}
	prepared, err := s.PrepareResearchBound(ctx, slot, d, plan)
	if err != nil {
		b.Fatal(err)
	}
	return s, d, slot, prepared, plan, candidate.Features, candidate.Baseline
}

func BenchmarkResearchBoundPublication(b *testing.B) {
	b.Run("prepare", func(b *testing.B) {
		s, d, slot, _, plan, _, _ := boundPublicationBenchmarkFixture(b)
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := s.PrepareResearchBound(ctx, slot, d, plan); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("publish", func(b *testing.B) {
		s, _, _, prepared, _, _, _ := boundPublicationBenchmarkFixture(b)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			copyPrepared := *prepared
			copyPrepared.slot = &ResearchBoundSlot{}
			if err := s.PublishPreparedResearchBound(ctx, &copyPrepared); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("score", func(b *testing.B) {
		s, _, slot, prepared, plan, features, baseline := boundPublicationBenchmarkFixture(b)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.PublishPreparedResearchBound(ctx, prepared); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			score, err := slot.Score(ctx, s, features, baseline, plan.Cutoff)
			if err != nil {
				b.Fatal(err)
			}
			researchBoundBenchmarkSink = score
		}
	})
	b.Run("score-fitted", func(b *testing.B) {
		s, _, slot, prepared, plan, features, baseline := boundPublicationBenchmarkFixture(b)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.PublishPreparedResearchBound(ctx, prepared); err != nil {
			b.Fatal(err)
		}
		adapter := researchmemory.New(plan.Epoch, plan.Seed)
		for i := 0; i < 32; i++ {
			at := plan.Cutoff.Add(time.Duration(i) * time.Second)
			prediction, err := adapter.Predict(uint16(i*7%512), baseline, plan.Epoch, at)
			if err != nil {
				b.Fatal(err)
			}
			if err := adapter.Feedback(prediction.ID, i%2 == 0, plan.Epoch, at.Add(time.Millisecond)); err != nil {
				b.Fatal(err)
			}
		}
		published := *slot.current.Load()
		published.model = adapter.Freeze()
		slot.current.Store(&published)
		at := plan.Cutoff.Add(33 * time.Second)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			score, err := slot.Score(ctx, s, features, baseline, at)
			if err != nil || math.IsNaN(score) {
				b.Fatalf("fitted score failed: %v", err)
			}
			researchBoundBenchmarkSink = score
		}
	})
}
