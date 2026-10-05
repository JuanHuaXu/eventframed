package service

import (
	"context"
	"errors"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
)

func motionExitService(t *testing.T, root string, create bool) (*Service, *ResearchFrontierTap) {
	t.Helper()
	em, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	backend, err := libravdbstore.Open(libravdbstore.Config{Path: filepath.Join(root, "events.libravdb"), Dimension: 8, EmbeddingModel: em.ModelKey(), Quantization: "none", MemoryMapping: true})
	if err != nil {
		t.Fatal(err)
	}
	wrapped, err := researchpublicationstore.WrapWithDurableLineage(backend, filepath.Join(root, "lineage.sqlite"), create)
	if err != nil {
		_ = backend.Close()
		t.Fatal(err)
	}
	tap, err := NewResearchFrontierTap("tenant-a", 4)
	if err != nil {
		_ = wrapped.Close()
		t.Fatal(err)
	}
	s, err := New(wrapped, em, Config{ResearchFrontier: tap, DefaultRecallK: 3, DefaultPackK: 2, DefaultTokenBudget: 1000})
	if err != nil {
		tap.Close()
		_ = wrapped.Close()
		t.Fatal(err)
	}
	return s, tap
}

func motionExitRequest(s *Service, at time.Time) model.RecallRequest {
	return model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: s.embedder.ModelKey(), AsOf: at, RecallK: 3, PackK: 2, TokenBudget: 1000}
}

func motionExitCandidate(t *testing.T, observed ResearchFrontierObservation) ResearchFrontierCandidate {
	t.Helper()
	for _, candidate := range observed.Candidates {
		if candidate.EventID == "seed" {
			return candidate
		}
	}
	t.Fatal("seed source not nominated")
	return ResearchFrontierCandidate{}
}

func TestResearchMotionBoundWorkerExitHelper(t *testing.T) {
	root := os.Getenv("EVENTFRAME_MOTION_EXIT_ROOT")
	if root == "" {
		t.Skip("child only")
	}
	mode := os.Getenv("EVENTFRAME_MOTION_EXIT_MODE")
	if mode != "" && mode != "fit" {
		t.Fatal("invalid motion exit mode")
	}
	ctx := context.Background()
	at := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	key := []byte("0123456789abcdef0123456789abcdef")
	s, _ := motionExitService(t, root, true)
	putTemporalFixture(t, s, "seed", at.Add(-time.Minute))
	source, err := researchmemory.OpenDurable(ctx, filepath.Join(root, "source.sqlite"), "tenant-a", "source-stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	observed := takeTemporalFixture(t, s, at)
	candidate := motionExitCandidate(t, observed)
	request := motionExitRequest(s, at)
	binding := researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: observed.JournalID, EventID: candidate.EventID, Snapshot: observed.Snapshot}
	preview := researchmemory.RecordedPrediction{Contract: researchmemory.RecordContract, Seed: 42,
		Prediction: researchmemory.Prediction{ID: 1, Probability: candidate.Baseline, Features: candidate.Features, Epoch: 1},
		At:         at, Outer: [4]float64{candidate.Baseline}, Binding: &binding}
	guardCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	err = s.WithValidatedResearchAsOfAdmission(guardCtx, preview, request, func() error {
		journal, readErr := s.store.GetBayesianJournal(guardCtx, binding.Tenant, binding.JournalID)
		if readErr != nil {
			return readErr
		}
		events, readErr := s.store.GetEvents(guardCtx, binding.Tenant, []string{binding.EventID}, at)
		if readErr != nil {
			return readErr
		}
		if len(events) != 1 {
			return errors.New("source event unavailable")
		}
		witness, witnessErr := researchmemory.NewSourceWitness("motion-exit", key, binding, journal.QueryDigest, events[0], candidate.Features, candidate.Baseline)
		if witnessErr != nil {
			return witnessErr
		}
		_, _, admitErr := source.AdmitBoundWithWitness(guardCtx, 1, candidate.Features, candidate.Baseline, at, binding, witness)
		return admitErr
	})
	cancel()
	if err != nil {
		t.Fatal(err)
	}
	original, err := source.Admission(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	guardCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
	err = s.WithValidatedResearchAsOfAdmission(guardCtx, original, request, func() error {
		_, labelErr := source.Feedback(guardCtx, 1, true, at.Add(time.Second))
		return labelErr
	})
	cancel()
	if err != nil {
		t.Fatal(err)
	}
	waitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	if err = source.WaitProcessed(waitCtx, 1); err != nil {
		t.Fatal(err)
	}
	cancel()
	plan := ResearchBoundPlan{Target: s.store.Snapshot(ctx), Tenant: "tenant-a", Epoch: 2, Seed: 43, Cutoff: at.Add(2 * time.Second), KeyID: "motion-exit", Key: key}
	path := filepath.Join(root, "motion.sqlite")
	prepared, err := s.PrepareResearchMotionBoundWorker(ctx, source, plan, "motion-stream", path)
	if err != nil {
		t.Fatal(err)
	}
	openCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	worker, err := s.OpenPreparedResearchBoundWorker(openCtx, prepared, path)
	cancel()
	if err != nil {
		t.Fatal(err)
	}
	labels := 1
	if mode == "fit" {
		labels = 31
	}
	for i := range labels {
		laterAt := at.Add(time.Duration(3+2*i) * time.Second)
		later := takeTemporalFixture(t, s, laterAt)
		laterRequest := motionExitRequest(s, laterAt)
		guardCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
		first, retry, admitErr := worker.Admit(guardCtx, laterRequest, later, motionExitCandidate(t, later))
		cancel()
		if admitErr != nil || retry || first.ID != uint64(i+1) {
			t.Fatalf("continuing admission %d: %+v retry=%v err=%v", i, first, retry, admitErr)
		}
		guardCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
		_, err = worker.Feedback(guardCtx, laterRequest, later.JournalID, "seed", false, laterAt.Add(time.Second))
		cancel()
		if err != nil {
			t.Fatal(err)
		}
	}
	waitCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
	if err = worker.WaitProcessed(waitCtx, uint64(1+labels)); err != nil {
		t.Fatal(err)
	}
	cancel()
	if mode == "fit" {
		probeAt := at.Add(65 * time.Second)
		probe := takeTemporalFixture(t, s, probeAt)
		probeCandidate := motionExitCandidate(t, probe)
		admitCtx, release := context.WithTimeout(ctx, 2*time.Second)
		prediction, retry, admitErr := worker.Admit(admitCtx, motionExitRequest(s, probeAt), probe, probeCandidate)
		release()
		if admitErr != nil || retry || prediction.ID != 32 || math.Abs(prediction.Probability-probeCandidate.Baseline) <= .01 {
			t.Fatalf("nontrivial post-fit probe unavailable: %+v baseline=%g retry=%v err=%v", prediction, probeCandidate.Baseline, retry, admitErr)
		}
	}
	putTemporalFixture(t, s, "future-exit", at.Add(time.Hour))
	os.Exit(27)
}

func TestResearchMotionBoundWorkerAcknowledgedAbruptExit(t *testing.T) {
	runMotionBoundWorkerAbruptExit(t, "")
}

func TestResearchMotionBoundWorkerFittedAbruptExit(t *testing.T) {
	runMotionBoundWorkerAbruptExit(t, "fit")
}

func runMotionBoundWorkerAbruptExit(t *testing.T, mode string) {
	t.Helper()
	root := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestResearchMotionBoundWorkerExitHelper$")
	cmd.Env = append(os.Environ(), "EVENTFRAME_MOTION_EXIT_ROOT="+root, "EVENTFRAME_MOTION_EXIT_MODE="+mode)
	output, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 27 {
		t.Fatalf("abrupt exit boundary not reached: err=%v output=%s", err, output)
	}
	s, tap := motionExitService(t, root, false)
	defer s.Close()
	defer tap.Close()
	ctx = context.Background()
	source, err := researchmemory.OpenDurable(ctx, filepath.Join(root, "source.sqlite"), "tenant-a", "source-stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	at := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	plan := ResearchBoundPlan{Target: s.store.Snapshot(ctx), Tenant: "tenant-a", Epoch: 2, Seed: 43, Cutoff: at.Add(2 * time.Second), KeyID: "motion-exit", Key: []byte("0123456789abcdef0123456789abcdef")}
	path := filepath.Join(root, "motion.sqlite")
	prepared, err := s.PrepareResearchMotionBoundWorker(ctx, source, plan, "motion-stream", path)
	if err != nil {
		t.Fatal(err)
	}
	openCtx, release := context.WithTimeout(ctx, 2*time.Second)
	worker, err := s.OpenPreparedResearchBoundWorker(openCtx, prepared, path)
	release()
	if err != nil {
		t.Fatal(err)
	}
	defer worker.Close()
	wantCompleted, wantPending := uint64(2), 0
	if mode == "fit" {
		wantCompleted, wantPending = 32, 1
	}
	if completed, failed, pending, queued := worker.Counts(); completed != wantCompleted || failed != 0 || pending != wantPending || queued != 0 {
		t.Fatalf("abrupt exit replay counts: %d/%d/%d/%d", completed, failed, pending, queued)
	}
	original, err := worker.durable.Admission(ctx, 1)
	if err != nil || original.Binding == nil || original.Witness == nil {
		t.Fatalf("missing recovered original: %+v err=%v", original, err)
	}
	bySource, err := worker.durable.AdmissionBySource(ctx, original.Binding.JournalID, original.Binding.EventID)
	if err != nil || bySource.Prediction != original.Prediction {
		t.Fatalf("source index did not recover: %+v err=%v", bySource, err)
	}
	nextAt, wantID := at.Add(5*time.Second), uint64(2)
	if mode == "fit" {
		nextAt, wantID = at.Add(65*time.Second), 33
	}
	next := takeTemporalFixture(t, s, nextAt)
	nextRequest := motionExitRequest(s, nextAt)
	nextCandidate := motionExitCandidate(t, next)
	var probe researchmemory.RecordedPrediction
	if mode == "fit" {
		probe, err = worker.durable.Admission(ctx, 32)
		if err != nil || probe.Binding == nil || probe.Witness == nil || probe.Prediction.Features != nextCandidate.Features || probe.Outer[0] != nextCandidate.Baseline || math.Abs(probe.Prediction.Probability-probe.Outer[0]) <= .01 {
			t.Fatalf("pre-exit fitted probe not comparable: %+v candidate=%+v err=%v", probe, nextCandidate, err)
		}
	}
	admitCtx, release := context.WithTimeout(ctx, 2*time.Second)
	nextPrediction, retry, err := worker.Admit(admitCtx, nextRequest, next, nextCandidate)
	release()
	if err != nil || retry || nextPrediction.ID != wantID {
		t.Fatalf("post-exit next admission: %+v retry=%v err=%v", nextPrediction, retry, err)
	}
	if mode == "fit" && nextPrediction.Probability != probe.Prediction.Probability {
		t.Fatalf("fitted prediction changed after abrupt exit: before=%g after=%g", probe.Prediction.Probability, nextPrediction.Probability)
	}
	if mode == "fit" {
		t.Logf("fitted probe baseline=%0.8f before=%0.8f after=%0.8f", probe.Outer[0], probe.Prediction.Probability, nextPrediction.Probability)
	}
	if err = worker.Close(); err != nil {
		t.Fatal(err)
	}
	if mode == "fit" {
		if _, err = s.store.Delete(ctx, "tenant-a", "seed"); err != nil {
			t.Fatal(err)
		}
	} else {
		putTemporalFixture(t, s, "backfill-exit", at.Add(-30*time.Second))
	}
	plan.Target = s.store.Snapshot(ctx)
	if _, err = s.PrepareResearchMotionBoundWorker(ctx, source, plan, "motion-stream", path); err == nil {
		t.Fatal("post-exit source mutation reopened worker")
	}
}
