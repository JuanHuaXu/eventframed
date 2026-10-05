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
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestResearchDurableLineageRebuildAcrossRestart(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	em, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	cfg := libravdbstore.Config{Path: filepath.Join(root, "events.libravdb"), Dimension: 8, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true}
	lineagePath := filepath.Join(root, "lineage.sqlite")
	open := func(create bool, tap *ResearchFrontierTap) *Service {
		t.Helper()
		backend, err := libravdbstore.Open(cfg)
		if err != nil {
			t.Fatal(err)
		}
		wrapped, err := researchpublicationstore.WrapWithDurableLineage(backend, lineagePath, create)
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
	s := open(true, tap)
	for _, id := range []string{"a", "b"} {
		putTemporalFixture(t, s, id, now.Add(-time.Minute))
	}
	in := takeTemporalFixture(t, s, now)
	if len(in.Candidates) != 2 {
		t.Fatalf("expected A/B source candidates, got %d", len(in.Candidates))
	}
	request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: em.ModelKey(), AsOf: now, RecallK: 3, PackK: 2, TokenBudget: 1000}
	journal, err := s.store.GetBayesianJournal(ctx, "tenant-a", in.JournalID)
	if err != nil {
		t.Fatal(err)
	}
	key := []byte("0123456789abcdef0123456789abcdef")
	durablePath := filepath.Join(root, "labels.sqlite")
	d, err := researchmemory.OpenDurable(ctx, durablePath, "tenant-a", "witness-stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	labels := make([]researchmemory.BoundLabel, 0, 2)
	originals := make([]researchmemory.RecordedPrediction, 0, 2)
	for i, candidate := range in.Candidates {
		id := uint64(i + 1)
		binding := researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: candidate.EventID, Snapshot: in.Snapshot}
		preview := researchmemory.RecordedPrediction{Contract: researchmemory.RecordContract, Seed: 42, Prediction: researchmemory.Prediction{ID: id, Probability: candidate.Baseline, Features: candidate.Features, Epoch: 1}, At: in.AsOf, Outer: [4]float64{candidate.Baseline}, Binding: &binding}
		err := s.WithValidatedResearchAdmission(ctx, preview, request, func() error {
			events, readErr := s.store.GetEvents(ctx, "tenant-a", []string{candidate.EventID}, in.AsOf)
			if readErr != nil {
				return readErr
			}
			witness, witnessErr := researchmemory.NewSourceWitness("lab-v1", key, binding, journal.QueryDigest, events[0], candidate.Features, candidate.Baseline)
			if witnessErr != nil {
				return witnessErr
			}
			_, _, admitErr := d.AdmitBoundWithWitness(ctx, id, candidate.Features, candidate.Baseline, in.AsOf, binding, witness)
			return admitErr
		})
		if err != nil {
			t.Fatal(err)
		}
		original, err := d.Admission(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		originals = append(originals, original)
	}
	for i, candidate := range in.Candidates {
		original := originals[i]
		id := original.Prediction.ID
		feedback := researchmemory.RecordedFeedback{ID: id, Useful: candidate.EventID == "b", Available: now.Add(time.Duration(i+1) * time.Second)}
		guardCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = s.WithValidatedResearchAsOfAdmission(guardCtx, original, request, func() error {
			_, writeErr := d.Feedback(guardCtx, feedback.ID, feedback.Useful, feedback.Available)
			return writeErr
		})
		cancel()
		if err != nil {
			t.Fatal(err)
		}
		labels = append(labels, researchmemory.BoundLabel{Prediction: original, Feedback: feedback})
	}
	waitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	if err := d.WaitProcessed(waitCtx, 2); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	d, err = researchmemory.OpenDurable(ctx, durablePath, "tenant-a", "witness-stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if completed, _, _, _ := d.Counts(); completed != 2 {
		t.Fatal("durable labels did not replay")
	}
	s = open(false, nil)
	if deleted, err := s.Delete(ctx, model.DeleteRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", EventID: "a"}); err != nil || !deleted.Deleted {
		t.Fatalf("delete a: %+v %v", deleted, err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = open(false, nil)
	cutoff := now.Add(3 * time.Second)
	validate := func(ctx context.Context, target model.Snapshot, label researchmemory.BoundLabel) (bool, error) {
		return s.ValidateResearchTransferSource(ctx, target, label.Prediction, "lab-v1", key, cutoff)
	}
	_, retained, err := researchmemory.RebuildFromBoundLabels(ctx, s.store.Snapshot(ctx), "tenant-a", 2, 43, cutoff, labels, validate)
	if err != nil || retained != 1 {
		t.Fatalf("restarted A-deleted rebuild: retained=%d err=%v", retained, err)
	}
	if deleted, err := s.Delete(ctx, model.DeleteRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", EventID: "b"}); err != nil || !deleted.Deleted {
		t.Fatalf("delete b: %+v %v", deleted, err)
	}
	if _, err := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: "b", Event: testutil.Event("b", "public query fixture", now.Add(-time.Minute))}); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = open(false, nil)
	defer s.Close()
	_, retained, err = researchmemory.RebuildFromBoundLabels(ctx, s.store.Snapshot(ctx), "tenant-a", 3, 44, cutoff, labels, validate)
	if err != nil || retained != 0 {
		t.Fatalf("restarted same-ID rebuild: retained=%d err=%v", retained, err)
	}
}
