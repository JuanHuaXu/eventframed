package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
)

func TestResearchDelayedFeedbackAfterDurableMotionRestart(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	em, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	cfg := libravdbstore.Config{Path: filepath.Join(root, "events.libravdb"), Dimension: 8, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true}
	path := filepath.Join(root, "lineage.sqlite")
	open := func(durable, create bool, tap *ResearchFrontierTap) *Service {
		t.Helper()
		backend, err := libravdbstore.Open(cfg)
		if err != nil {
			t.Fatal(err)
		}
		var wrapped store.EventStore
		if durable {
			wrapped, err = researchpublicationstore.WrapWithDurableLineage(backend, path, create)
		} else {
			wrapped, err = researchpublicationstore.Wrap(backend)
		}
		if err != nil {
			_ = backend.Close()
			t.Fatal(err)
		}
		s, err := New(wrapped, em, Config{ResearchFrontier: tap, DefaultRecallK: 3, DefaultPackK: 2, DefaultTokenBudget: 1000})
		if err != nil {
			_ = wrapped.Close()
			t.Fatal(err)
		}
		return s
	}
	tap, err := NewResearchFrontierTap("tenant-a", 4)
	if err != nil {
		t.Fatal(err)
	}
	defer tap.Close()
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	s := open(true, true, tap)
	putTemporalFixture(t, s, "old", now.Add(-time.Minute))
	in := takeTemporalFixture(t, s, now)
	if len(in.Candidates) != 1 || in.Candidates[0].EventID != "old" {
		t.Fatalf("unexpected original frontier: %+v", in.Candidates)
	}
	candidate := in.Candidates[0]
	request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: em.ModelKey(), AsOf: now, RecallK: 3, PackK: 2, TokenBudget: 1000}
	journal, err := s.store.GetBayesianJournal(ctx, "tenant-a", in.JournalID)
	if err != nil {
		t.Fatal(err)
	}
	key := []byte("0123456789abcdef0123456789abcdef")
	durablePath := filepath.Join(root, "labels.sqlite")
	d, err := researchmemory.OpenDurable(ctx, durablePath, "tenant-a", "motion-stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	binding := researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: "old", Snapshot: in.Snapshot}
	preview := researchmemory.RecordedPrediction{Contract: researchmemory.RecordContract, Seed: 42, Prediction: researchmemory.Prediction{ID: 1, Probability: candidate.Baseline, Features: candidate.Features, Epoch: 1}, At: in.AsOf, Outer: [4]float64{candidate.Baseline}, Binding: &binding}
	err = s.WithValidatedResearchAdmission(ctx, preview, request, func() error {
		events, readErr := s.store.GetEvents(ctx, "tenant-a", []string{"old"}, in.AsOf)
		if readErr != nil {
			return readErr
		}
		witness, witnessErr := researchmemory.NewSourceWitness("lab-v1", key, binding, journal.QueryDigest, events[0], candidate.Features, candidate.Baseline)
		if witnessErr != nil {
			return witnessErr
		}
		_, _, admitErr := d.AdmitBoundWithWitness(ctx, 1, candidate.Features, candidate.Baseline, in.AsOf, binding, witness)
		return admitErr
	})
	if err != nil {
		t.Fatal(err)
	}
	original, err := d.Admission(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	putTemporalFixture(t, s, "future", now.Add(time.Hour))
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = open(true, false, nil)
	d, err = researchmemory.OpenDurable(ctx, durablePath, "tenant-a", "motion-stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if valid, err := s.ValidateResearchTransferSource(ctx, s.store.Snapshot(ctx), original, "lab-v1", key, now.Add(2*time.Second)); err != nil || !valid {
		t.Fatalf("original source failed after future-only restart: valid=%v err=%v", valid, err)
	}
	guardCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	err = s.WithValidatedResearchAsOfAdmission(guardCtx, original, request, func() error {
		_, writeErr := d.Feedback(guardCtx, 1, true, now.Add(time.Second))
		return writeErr
	})
	cancel()
	if err != nil {
		t.Fatalf("future-only old prediction rejected after restart: %v", err)
	}
	waitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	if err := d.WaitProcessed(waitCtx, 1); err != nil {
		t.Fatal(err)
	}
	cancel()
	if completed, _, _, _ := d.Counts(); completed != 1 {
		t.Fatal("reopened delayed label was not processed")
	}
	slot := &ResearchBoundSlot{}
	plan := ResearchBoundPlan{Target: s.store.Snapshot(ctx), Tenant: "tenant-a", Epoch: 2, Seed: 43, Cutoff: now.Add(2 * time.Second), KeyID: "lab-v1", Key: key}
	prepared, err := s.PrepareResearchBound(ctx, slot, d, plan)
	if err != nil || prepared == nil || prepared.candidate.Retained != 1 {
		t.Fatalf("future-only bound preparation: prepared=%v err=%v", prepared, err)
	}
	guardCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
	if err := s.PublishPreparedResearchBound(guardCtx, prepared); err != nil {
		t.Fatal(err)
	}
	cancel()
	scoreCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	if _, err := slot.Score(scoreCtx, s, candidate.Features, candidate.Baseline, now); err == nil {
		t.Fatal("published model used a label before its availability")
	}
	cancel()
	putTemporalFixture(t, s, "later", now.Add(2*time.Hour))
	scoreCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
	if _, err := slot.Score(scoreCtx, s, candidate.Features, candidate.Baseline, now.Add(2*time.Second)); err != nil {
		t.Fatalf("future-only ingestion revoked earlier as-of score: %v", err)
	}
	cancel()
	scoreCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
	if _, err := slot.Score(scoreCtx, s, candidate.Features, candidate.Baseline, now.Add(2*time.Hour)); err == nil {
		t.Fatal("post-publication event at its availability left old model scoreable")
	}
	cancel()
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = open(false, false, nil)
	defer s.Close()
	guardCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
	called := false
	err = s.WithValidatedResearchAsOfAdmission(guardCtx, original, request, func() error {
		called = true
		return nil
	})
	cancel()
	if err == nil || called {
		t.Fatal("default wrapper invented pre-restart future motion")
	}
}
