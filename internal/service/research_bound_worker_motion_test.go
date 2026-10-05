package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
)

func TestResearchBoundWorkerMotionProofDiagnostic(t *testing.T) {
	for _, persistent := range []bool{false, true} {
		name := "memory"
		if persistent {
			name = "persistent"
		}
		t.Run(name, func(t *testing.T) {
			s, _, at := temporalBridgeFixture(t, persistent)
			wrapped, err := researchpublicationstore.Wrap(s.store)
			if err != nil {
				t.Fatal(err)
			}
			s.store = wrapped
			proof, ok := wrapped.(interface {
				ResearchPublicationCompatible(context.Context, model.Snapshot, time.Time) bool
				ResearchEventUnchanged(context.Context, model.Snapshot, model.Snapshot, string, string) (bool, error)
			})
			if !ok {
				t.Fatal("wrapper lacks motion and source proof")
			}
			observed := takeTemporalFixture(t, s, at)
			var candidate ResearchFrontierCandidate
			for _, c := range observed.Candidates {
				if c.EventID == "seed" {
					candidate = c
				}
			}
			if candidate.EventID == "" {
				t.Fatal("source not nominated")
			}
			ctx := context.Background()
			journal, err := s.store.GetBayesianJournal(ctx, "tenant-a", observed.JournalID)
			if err != nil {
				t.Fatal(err)
			}
			events, err := s.store.GetEvents(ctx, "tenant-a", []string{candidate.EventID}, at)
			if err != nil || len(events) != 1 {
				t.Fatalf("source lookup: %v", err)
			}
			binding := researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: observed.JournalID, EventID: candidate.EventID, Snapshot: observed.Snapshot}
			key := []byte("0123456789abcdef0123456789abcdef")
			witness, err := researchmemory.NewSourceWitness("motion-v7", key, binding, journal.QueryDigest, events[0], candidate.Features, candidate.Baseline)
			if err != nil {
				t.Fatal(err)
			}
			original := researchmemory.RecordedPrediction{Contract: researchmemory.RecordContract, Seed: 42,
				Prediction: researchmemory.Prediction{ID: 1, Probability: candidate.Baseline, Features: candidate.Features, Epoch: 1},
				At:         at, Outer: [4]float64{candidate.Baseline}, Binding: &binding, Witness: &witness}
			if !proof.ResearchPublicationCompatible(ctx, binding.Snapshot, at) {
				t.Fatal("initial as-of publication incompatible")
			}
			valid, err := s.ValidateResearchTransferSource(ctx, binding.Snapshot, original, "motion-v7", key, at)
			if err != nil || !valid {
				t.Fatalf("initial witnessed source invalid: valid=%v err=%v", valid, err)
			}
			putTemporalFixture(t, s, "future-v7", at.Add(time.Hour))
			future := s.store.Snapshot(ctx)
			if future == binding.Snapshot || !proof.ResearchPublicationCompatible(ctx, binding.Snapshot, at) {
				t.Fatal("future-only motion lost valid as-of proof")
			}
			valid, err = s.ValidateResearchTransferSource(ctx, future, original, "motion-v7", key, at)
			if err != nil || !valid {
				t.Fatalf("future-only motion invalidated source: valid=%v err=%v", valid, err)
			}
			putTemporalFixture(t, s, "backfill-v7", at.Add(-30*time.Second))
			backfilled := s.store.Snapshot(ctx)
			if backfilled == future || proof.ResearchPublicationCompatible(ctx, binding.Snapshot, at) {
				t.Fatal("backfill accepted as harmless as-of motion")
			}
			continuous, err := proof.ResearchEventUnchanged(ctx, binding.Snapshot, backfilled, binding.Tenant, binding.EventID)
			if err != nil || !continuous {
				t.Fatalf("unchanged source lineage unavailable: same=%v err=%v", continuous, err)
			}
			valid, err = s.ValidateResearchTransferSource(ctx, backfilled, original, "motion-v7", key, at)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("unchanged source after backfill: source_valid=%v asof_compatible=false", valid)
		})
	}
}

func TestResearchMotionBoundWorkerRequiresProofAndRejectsLegacyLog(t *testing.T) {
	s, _, at := temporalBridgeFixture(t, false)
	ctx := context.Background()
	root := t.TempDir()
	source, err := researchmemory.OpenDurable(ctx, filepath.Join(root, "source.sqlite"), "tenant-a", "source-stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	plan := ResearchBoundPlan{Target: s.store.Snapshot(ctx), Tenant: "tenant-a", Epoch: 2, Seed: 43, Cutoff: at.Add(time.Second), KeyID: "motion-v8", Key: []byte("0123456789abcdef0123456789abcdef")}
	if _, err = s.PrepareResearchMotionBoundWorker(ctx, source, plan, "motion-stream", filepath.Join(root, "no-proof.sqlite")); err == nil {
		t.Fatal("store without as-of motion proof prepared a worker")
	}
	wrapper, err := researchpublicationstore.Wrap(s.store)
	if err != nil {
		t.Fatal(err)
	}
	s.store = wrapper
	plan.Target = s.store.Snapshot(ctx)
	path := filepath.Join(root, "legacy.sqlite")
	legacy, err := researchledger.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = legacy.BindBootstrap(ctx, make([]byte, 32)); err != nil {
		t.Fatal(err)
	}
	if err = legacy.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = s.PrepareResearchMotionBoundWorker(ctx, source, plan, "motion-stream", path); err == nil {
		t.Fatal("legacy sealed log silently entered motion mode")
	}
}

func TestResearchMotionBoundWorkerFutureOnlyRestartAndBackfill(t *testing.T) {
	for _, persistent := range []bool{false, true} {
		name := "memory"
		if persistent {
			name = "persistent"
		}
		t.Run(name, func(t *testing.T) {
			s, _, at := temporalBridgeFixture(t, persistent)
			wrapped, err := researchpublicationstore.Wrap(s.store)
			if err != nil {
				t.Fatal(err)
			}
			s.store = wrapped
			ctx := context.Background()
			request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: s.embedder.ModelKey(), AsOf: at, RecallK: 3, PackK: 2, TokenBudget: 1000}
			observed := takeTemporalFixture(t, s, at)
			var source ResearchFrontierCandidate
			for _, candidate := range observed.Candidates {
				if candidate.EventID == "seed" {
					source = candidate
				}
			}
			if source.EventID == "" {
				t.Fatal("source not nominated")
			}
			binding := researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: observed.JournalID, EventID: source.EventID, Snapshot: observed.Snapshot}
			preview := researchmemory.RecordedPrediction{Contract: researchmemory.RecordContract, Seed: 42,
				Prediction: researchmemory.Prediction{ID: 1, Probability: source.Baseline, Features: source.Features, Epoch: 1},
				At:         at, Outer: [4]float64{source.Baseline}, Binding: &binding}
			key := []byte("0123456789abcdef0123456789abcdef")
			root := t.TempDir()
			durable, err := researchmemory.OpenDurable(ctx, filepath.Join(root, "source.sqlite"), "tenant-a", "source-stream", 1, 42)
			if err != nil {
				t.Fatal(err)
			}
			defer durable.Close()
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
				witness, witnessErr := researchmemory.NewSourceWitness("motion-v8", key, binding, journal.QueryDigest, events[0], source.Features, source.Baseline)
				if witnessErr != nil {
					return witnessErr
				}
				_, _, admitErr := durable.AdmitBoundWithWitness(guardCtx, 1, source.Features, source.Baseline, at, binding, witness)
				return admitErr
			})
			cancel()
			if err != nil {
				t.Fatal(err)
			}
			original, err := durable.Admission(ctx, 1)
			if err != nil {
				t.Fatal(err)
			}
			guardCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			err = s.WithValidatedResearchAsOfAdmission(guardCtx, original, request, func() error {
				_, labelErr := durable.Feedback(guardCtx, 1, true, at.Add(time.Second))
				return labelErr
			})
			cancel()
			if err != nil {
				t.Fatal(err)
			}
			waitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			if err = durable.WaitProcessed(waitCtx, 1); err != nil {
				t.Fatal(err)
			}
			cancel()
			path := filepath.Join(root, "motion.sqlite")
			plan := ResearchBoundPlan{Target: s.store.Snapshot(ctx), Tenant: "tenant-a", Epoch: 2, Seed: 43, Cutoff: at.Add(2 * time.Second), KeyID: "motion-v8", Key: key}
			prepare := func() *PreparedResearchBoundWorker {
				prepared, prepareErr := s.PrepareResearchMotionBoundWorker(ctx, durable, plan, "motion-stream", path)
				if prepareErr != nil {
					t.Fatal(prepareErr)
				}
				return prepared
			}
			open := func(prepared *PreparedResearchBoundWorker) *ResearchBoundWorker {
				openCtx, release := context.WithTimeout(ctx, 2*time.Second)
				defer release()
				worker, openErr := s.OpenPreparedResearchBoundWorker(openCtx, prepared, path)
				if openErr != nil {
					t.Fatal(openErr)
				}
				return worker
			}
			worker := open(prepare())
			laterAt := at.Add(3 * time.Second)
			later := takeTemporalFixture(t, s, laterAt)
			laterRequest := request
			laterRequest.AsOf = laterAt
			var laterCandidate ResearchFrontierCandidate
			for _, candidate := range later.Candidates {
				if candidate.EventID == "seed" {
					laterCandidate = candidate
				}
			}
			if laterCandidate.EventID == "" {
				t.Fatal("later source not nominated")
			}
			admitCtx, release := context.WithTimeout(ctx, 2*time.Second)
			first, retry, err := worker.Admit(admitCtx, laterRequest, later, laterCandidate)
			release()
			if err != nil || retry || first.ID != 1 {
				t.Fatalf("first continuing admission: %+v retry=%v err=%v", first, retry, err)
			}
			feedbackCtx, release := context.WithTimeout(ctx, 2*time.Second)
			_, err = worker.Feedback(feedbackCtx, laterRequest, later.JournalID, laterCandidate.EventID, false, laterAt.Add(time.Second))
			release()
			if err != nil {
				t.Fatal(err)
			}
			waitCtx, release = context.WithTimeout(ctx, 2*time.Second)
			if err = worker.WaitProcessed(waitCtx, 2); err != nil {
				t.Fatal(err)
			}
			release()
			if err = worker.Close(); err != nil {
				t.Fatal(err)
			}
			putTemporalFixture(t, s, "future-motion-v8", at.Add(time.Hour))
			plan.Target = s.store.Snapshot(ctx)
			worker = open(prepare())
			if completed, failed, pending, queued := worker.Counts(); completed != 2 || failed != 0 || pending != 0 || queued != 0 {
				t.Fatalf("future-only restart counts: %d/%d/%d/%d", completed, failed, pending, queued)
			}
			admitCtx, release = context.WithTimeout(ctx, 2*time.Second)
			if same, retry, err := worker.Admit(admitCtx, laterRequest, later, laterCandidate); err != nil || !retry || same != first {
				t.Fatalf("future-only source retry: %+v retry=%v err=%v", same, retry, err)
			}
			release()
			nextAt := at.Add(5 * time.Second)
			next := takeTemporalFixture(t, s, nextAt)
			nextRequest := request
			nextRequest.AsOf = nextAt
			admitCtx, release = context.WithTimeout(ctx, 2*time.Second)
			nextPrediction, retry, err := worker.Admit(admitCtx, nextRequest, next, next.Candidates[0])
			release()
			if err != nil || retry || nextPrediction.ID != 2 {
				t.Fatalf("future-only next admission: %+v retry=%v err=%v", nextPrediction, retry, err)
			}
			if err = worker.Close(); err != nil {
				t.Fatal(err)
			}
			putTemporalFixture(t, s, "backfill-motion-v8", at.Add(-30*time.Second))
			plan.Target = s.store.Snapshot(ctx)
			if prepared, prepareErr := s.PrepareResearchMotionBoundWorker(ctx, durable, plan, "motion-stream", path); prepareErr == nil {
				_ = prepared
				t.Fatal("backfill reopened origin-bound worker")
			}
		})
	}
}
