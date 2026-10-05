package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestResearchSourceWitnessSurvivesUnrelatedDeletion(t *testing.T) {
	for _, persistent := range []bool{false, true} {
		name := "memory"
		if persistent {
			name = "persistent"
		}
		t.Run(name, func(t *testing.T) {
			s, _, now := temporalBridgeFixture(t, persistent)
			backend := s.store
			wrapped, err := researchpublicationstore.Wrap(backend)
			if err != nil {
				t.Fatal(err)
			}
			s.store = wrapped
			putTemporalFixture(t, s, "survivor", now.Add(-time.Minute))
			in := takeTemporalFixture(t, s, now)
			if len(in.Candidates) != 2 {
				t.Fatalf("expected two source candidates, got %d", len(in.Candidates))
			}
			ctx := context.Background()
			request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: s.embedder.ModelKey(), AsOf: now, RecallK: 3, PackK: 2, TokenBudget: 1000}
			journal, err := s.store.GetBayesianJournal(ctx, "tenant-a", in.JournalID)
			if err != nil {
				t.Fatal(err)
			}
			key := []byte("0123456789abcdef0123456789abcdef")
			path := filepath.Join(t.TempDir(), "witness.sqlite")
			d, err := researchmemory.OpenDurable(ctx, path, "tenant-a", "witness-stream", 1, 42)
			if err != nil {
				t.Fatal(err)
			}
			originals := make(map[string]researchmemory.RecordedPrediction)
			for i, candidate := range in.Candidates {
				id := uint64(i + 1)
				binding := researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: candidate.EventID, Snapshot: in.Snapshot}
				preview := researchmemory.RecordedPrediction{
					Contract: researchmemory.RecordContract, Seed: 42,
					Prediction: researchmemory.Prediction{ID: id, Probability: candidate.Baseline, Features: candidate.Features, Epoch: 1},
					At:         in.AsOf, Outer: [4]float64{candidate.Baseline}, Binding: &binding,
				}
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
				if err != nil || original.Witness == nil {
					t.Fatalf("durable witness missing: %v", err)
				}
				originals[candidate.EventID] = original
			}
			if err := d.Close(); err != nil {
				t.Fatal(err)
			}
			d, err = researchmemory.OpenDurable(ctx, path, "tenant-a", "witness-stream", 1, 42)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = d.Close() }()
			for _, original := range originals {
				replayed, err := d.Admission(ctx, original.Prediction.ID)
				if err != nil || replayed.Witness == nil || *replayed.Witness != *original.Witness {
					t.Fatalf("witness changed across durable reopen: %v", err)
				}
			}
			before := s.store.Snapshot(ctx)
			missing, err := s.Delete(ctx, model.DeleteRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", EventID: "absent"})
			if err != nil || missing.Deleted || s.store.Snapshot(ctx) != before {
				t.Fatalf("no-op deletion changed source history: %v", err)
			}
			for _, original := range originals {
				valid, err := s.ValidateResearchTransferSource(ctx, before, original, "lab-v1", key, now.Add(time.Second))
				if err != nil || !valid {
					t.Fatalf("original source did not validate: valid=%v err=%v", valid, err)
				}
			}
			labels := make([]researchmemory.BoundLabel, 0, len(in.Candidates))
			for i, candidate := range in.Candidates {
				original := originals[candidate.EventID]
				feedback := researchmemory.RecordedFeedback{ID: original.Prediction.ID, Useful: candidate.EventID == "survivor", Available: now.Add(time.Duration(i+1) * time.Second)}
				guardCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				err := s.WithValidatedResearchAsOfAdmission(guardCtx, original, request, func() error {
					_, writeErr := d.Feedback(guardCtx, feedback.ID, feedback.Useful, feedback.Available)
					return writeErr
				})
				cancel()
				if err != nil {
					t.Fatalf("guarded label failed: %v", err)
				}
				labels = append(labels, researchmemory.BoundLabel{Prediction: original, Feedback: feedback})
			}
			waitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			if err := d.WaitProcessed(waitCtx, 2); err != nil {
				t.Fatalf("bound labels did not fit: %v", err)
			}
			cancel()
			if err := d.Close(); err != nil {
				t.Fatal(err)
			}
			d, err = researchmemory.OpenDurable(ctx, path, "tenant-a", "witness-stream", 1, 42)
			if err != nil {
				t.Fatalf("guarded labels did not replay: %v", err)
			}
			if completed, failed, pending, queued := d.Counts(); completed != 2 || failed != 0 || pending != 0 || queued != 0 {
				t.Fatalf("replayed label accounting: completed=%d failed=%d pending=%d queued=%d", completed, failed, pending, queued)
			}
			deleted, err := s.Delete(ctx, model.DeleteRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", EventID: "seed"})
			if err != nil || !deleted.Deleted {
				t.Fatalf("source deletion failed: %v", err)
			}
			after := s.store.Snapshot(ctx)
			if after == before {
				t.Fatal("deletion did not move snapshot")
			}
			for event, original := range originals {
				valid, err := s.ValidateResearchTransferSource(ctx, after, original, "lab-v1", key, now.Add(2*time.Second))
				if err != nil || valid != (event == "survivor") {
					t.Fatalf("post-delete %s: valid=%v err=%v", event, valid, err)
				}
			}
			cutoff := now.Add(3 * time.Second)
			validate := func(ctx context.Context, target model.Snapshot, label researchmemory.BoundLabel) (bool, error) {
				return s.ValidateResearchTransferSource(ctx, target, label.Prediction, "lab-v1", key, cutoff)
			}
			rebuilt, retained, err := researchmemory.RebuildFromBoundLabels(ctx, after, "tenant-a", 2, 43, cutoff, labels, validate)
			if err != nil || retained != 1 || rebuilt == nil {
				t.Fatalf("post-delete rebuild: retained=%d err=%v", retained, err)
			}
			if _, err := rebuilt.Freeze().Score(1, .6, 1, cutoff); err == nil {
				t.Fatal("rebuild accepted old epoch")
			}
			if valid, err := s.ValidateResearchTransferSource(ctx, before, originals["survivor"], "lab-v1", key, now.Add(2*time.Second)); err == nil || valid {
				t.Fatal("old target snapshot remained usable")
			}
			slot := &ResearchBoundSlot{}
			plan := ResearchBoundPlan{Target: after, Tenant: "tenant-a", Epoch: 2, Seed: 43, Cutoff: cutoff, KeyID: "lab-v1", Key: key}
			prepared, err := s.PrepareResearchBound(ctx, slot, d, plan)
			if err != nil || prepared.candidate.TerminalCount != 2 || prepared.candidate.Retained != 1 {
				t.Fatalf("complete durable preparation: prepared=%v err=%v", prepared, err)
			}
			publishCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			if err := s.PublishPreparedResearchBound(publishCtx, prepared); err != nil {
				t.Fatalf("guarded research publication failed: %v", err)
			}
			cancel()
			state, ok := slot.Current()
			if !ok || state.Epoch != 2 || state.Retained != 1 || state.TerminalCount != 2 {
				t.Fatal("published slot did not retain complete-stream accounting")
			}
			if _, err := s.PrepareResearchBound(ctx, slot, d, plan); err == nil {
				t.Fatal("equal-epoch replacement was prepared")
			}
			scoreCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			if _, err := slot.Score(scoreCtx, s, originals["survivor"].Prediction.Features, originals["survivor"].Outer[0], cutoff); err != nil {
				t.Fatalf("published model could not score: %v", err)
			}
			cancel()
			workerPath := filepath.Join(t.TempDir(), "bound-worker.sqlite")
			preparedWorker, err := s.PrepareResearchBoundWorker(ctx, d, plan, "bound-worker")
			if err != nil {
				t.Fatalf("worker preparation failed: %v", err)
			}
			noDeadlinePath := filepath.Join(t.TempDir(), "no-deadline.sqlite")
			if invalid, err := s.OpenPreparedResearchBoundWorker(ctx, preparedWorker, noDeadlinePath); err == nil {
				invalid.Close()
				t.Fatal("worker open without deadline")
			}
			if _, err := os.Stat(noDeadlinePath); !os.IsNotExist(err) {
				t.Fatalf("deadline rejection created ledger: %v", err)
			}
			openCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			boundWorker, err := s.OpenPreparedResearchBoundWorker(openCtx, preparedWorker, workerPath)
			cancel()
			if err != nil {
				t.Fatalf("guarded worker open failed: %v", err)
			}
			workerAt := cutoff.Add(time.Second)
			later := takeTemporalFixture(t, s, workerAt)
			laterRequest := request
			laterRequest.AsOf = workerAt
			candidate := later.Candidates[0]
			admitCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			prediction, retry, err := boundWorker.Admit(admitCtx, laterRequest, later, candidate)
			cancel()
			if err != nil || retry || prediction.ID != 1 {
				t.Fatalf("guarded worker admission: %+v retry=%v err=%v", prediction, retry, err)
			}
			admitCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if same, retry, err := boundWorker.Admit(admitCtx, laterRequest, later, candidate); err != nil || !retry || same != prediction {
				t.Fatalf("guarded worker retry: %+v retry=%v err=%v", same, retry, err)
			}
			cancel()
			changed := candidate
			changed.Features ^= 1
			admitCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if _, _, err := boundWorker.Admit(admitCtx, laterRequest, later, changed); err == nil {
				t.Fatal("changed candidate admitted as retry")
			}
			cancel()
			feedbackCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			if _, err := boundWorker.Feedback(feedbackCtx, laterRequest, later.JournalID, candidate.EventID, true, workerAt.Add(time.Second)); err != nil {
				t.Fatal(err)
			}
			cancel()
			waitCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if err := boundWorker.WaitProcessed(waitCtx, 2); err != nil {
				t.Fatal(err)
			}
			cancel()
			if err := boundWorker.Close(); err != nil {
				t.Fatal(err)
			}
			preparedWorker, err = s.PrepareResearchBoundWorker(ctx, d, plan, "bound-worker")
			if err != nil {
				t.Fatal(err)
			}
			openCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			boundWorker, err = s.OpenPreparedResearchBoundWorker(openCtx, preparedWorker, workerPath)
			cancel()
			if err != nil {
				t.Fatalf("guarded worker restart failed: %v", err)
			}
			if completed, failed, pending, queued := boundWorker.Counts(); completed != 2 || failed != 0 || pending != 0 || queued != 0 {
				t.Fatalf("bound worker restart counts: %d/%d/%d/%d", completed, failed, pending, queued)
			}
			admitCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if same, retry, err := boundWorker.Admit(admitCtx, laterRequest, later, candidate); err != nil || !retry || same != prediction {
				t.Fatalf("restarted worker retry: %+v retry=%v err=%v", same, retry, err)
			}
			cancel()
			nextAt := workerAt.Add(2 * time.Second)
			nextObserved := takeTemporalFixture(t, s, nextAt)
			nextRequest := request
			nextRequest.AsOf = nextAt
			var nextCandidate ResearchFrontierCandidate
			for _, c := range nextObserved.Candidates {
				if c.EventID == "survivor" {
					nextCandidate = c
				}
			}
			if nextCandidate.EventID == "" {
				t.Fatal("survivor absent from later research frontier")
			}
			admitCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			nextPrediction, retry, err := boundWorker.Admit(admitCtx, nextRequest, nextObserved, nextCandidate)
			cancel()
			if err != nil || retry || nextPrediction.ID != 2 {
				t.Fatalf("restarted worker did not continue IDs: %+v retry=%v err=%v", nextPrediction, retry, err)
			}
			if err := boundWorker.Close(); err != nil {
				t.Fatal(err)
			}
			preparedWorker, err = s.PrepareResearchBoundWorker(ctx, d, plan, "bound-worker")
			if err != nil {
				t.Fatal(err)
			}
			openCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			boundWorker, err = s.OpenPreparedResearchBoundWorker(openCtx, preparedWorker, workerPath)
			cancel()
			if err != nil {
				t.Fatalf("pending worker restart failed: %v", err)
			}
			if completed, failed, pending, queued := boundWorker.Counts(); completed != 2 || failed != 0 || pending != 1 || queued != 0 {
				t.Fatalf("missing feedback became evidence: %d/%d/%d/%d", completed, failed, pending, queued)
			}
			wrongRequest := nextRequest
			wrongRequest.Query = "wrong query"
			feedbackCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if _, err := boundWorker.Feedback(feedbackCtx, wrongRequest, nextObserved.JournalID, nextCandidate.EventID, true, nextAt.Add(time.Second)); err == nil {
				t.Fatal("wrong query labeled pending source")
			}
			cancel()
			plan.Epoch = 3
			firstPrepared, err := s.PrepareResearchBound(ctx, slot, d, plan)
			if err != nil {
				t.Fatal(err)
			}
			secondPrepared, err := s.PrepareResearchBound(ctx, slot, d, plan)
			if err != nil {
				t.Fatal(err)
			}
			publishCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if err := s.PublishPreparedResearchBound(publishCtx, firstPrepared); err != nil {
				t.Fatal(err)
			}
			cancel()
			publishCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if err := s.PublishPreparedResearchBound(publishCtx, secondPrepared); err == nil {
				t.Fatal("competing prepared publication replaced the winner")
			}
			cancel()
			plan.Epoch = 4
			logStale, err := s.PrepareResearchBound(ctx, slot, d, plan)
			if err != nil {
				t.Fatal(err)
			}
			logStaleWorker, err := s.PrepareResearchBoundWorker(ctx, d, plan, "stale-log-worker")
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := d.Admit(ctx, 3, 7, .6, cutoff); err != nil {
				t.Fatal(err)
			}
			publishCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if err := s.PublishPreparedResearchBound(publishCtx, logStale); err == nil {
				t.Fatal("durable log changed without invalidating preparation")
			}
			cancel()
			openCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if stale, err := s.OpenPreparedResearchBoundWorker(openCtx, logStaleWorker, filepath.Join(t.TempDir(), "stale-log.sqlite")); err == nil {
				stale.Close()
				t.Fatal("changed source log opened prepared worker")
			}
			cancel()
			storeStale, err := s.PrepareResearchBound(ctx, slot, d, plan)
			if err != nil {
				t.Fatal(err)
			}
			storeStaleWorker, err := s.PrepareResearchBoundWorker(ctx, d, plan, "stale-store-worker")
			if err != nil {
				t.Fatal(err)
			}
			deleted, err = s.Delete(ctx, model.DeleteRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", EventID: "survivor"})
			if err != nil || !deleted.Deleted {
				t.Fatalf("survivor deletion failed: %v", err)
			}
			feedbackCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if _, err := boundWorker.Feedback(feedbackCtx, nextRequest, nextObserved.JournalID, nextCandidate.EventID, true, nextAt.Add(time.Second)); err == nil {
				t.Fatal("deleted source labeled pending worker record")
			}
			cancel()
			admitCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if _, _, err := boundWorker.Admit(admitCtx, nextRequest, nextObserved, nextCandidate); err == nil {
				t.Fatal("deleted source retried as a current admission")
			}
			cancel()
			if completed, failed, pending, queued := boundWorker.Counts(); completed != 2 || failed != 0 || pending != 1 || queued != 0 {
				t.Fatalf("rejected source mutated learner: %d/%d/%d/%d", completed, failed, pending, queued)
			}
			if err := boundWorker.Close(); err != nil {
				t.Fatal(err)
			}
			publishCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if err := s.PublishPreparedResearchBound(publishCtx, storeStale); err == nil {
				t.Fatal("store mutation between preparation and publication was accepted")
			}
			cancel()
			openCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if stale, err := s.OpenPreparedResearchBoundWorker(openCtx, storeStaleWorker, filepath.Join(t.TempDir(), "stale-store.sqlite")); err == nil {
				stale.Close()
				t.Fatal("changed event store opened prepared worker")
			}
			cancel()
			state, ok = slot.Current()
			if !ok || state.Epoch != 3 {
				t.Fatal("failed publication changed the visible model")
			}
			scoreCtx, cancel = context.WithTimeout(ctx, 2*time.Second)
			if _, err := slot.Score(scoreCtx, s, originals["survivor"].Prediction.Features, originals["survivor"].Outer[0], cutoff); err == nil {
				t.Fatal("deleted source remained scoreable")
			}
			cancel()
			last := s.store.Snapshot(ctx)
			if valid, err := s.ValidateResearchTransferSource(ctx, last, originals["survivor"], "lab-v1", key, now.Add(3*time.Second)); err != nil || valid {
				t.Fatalf("deleted survivor retained: valid=%v err=%v", valid, err)
			}
			rebuilt, retained, err = researchmemory.RebuildFromBoundLabels(ctx, last, "tenant-a", 3, 44, cutoff, labels, validate)
			if err != nil || retained != 0 || rebuilt == nil {
				t.Fatalf("all-deleted rebuild: retained=%d err=%v", retained, err)
			}
			copy := testutil.Event("survivor", "public query fixture", now.Add(-time.Minute))
			recreated, err := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: "survivor", Event: copy})
			if err != nil {
				t.Logf("same-ID recreation unavailable in this store: %v", err)
			} else if recreated.Duplicate {
				t.Log("same-ID recreation returned duplicate")
			} else {
				reborn := s.store.Snapshot(ctx)
				valid, err := s.ValidateResearchTransferSource(ctx, reborn, originals["survivor"], "lab-v1", key, now.Add(4*time.Second))
				if err != nil {
					t.Fatal(err)
				}
				if valid {
					t.Fatal("same-ID recreation revived the old source")
				}
				rebuilt, retained, err = researchmemory.RebuildFromBoundLabels(ctx, reborn, "tenant-a", 4, 45, cutoff, labels, validate)
				if err != nil || retained != 0 || rebuilt == nil {
					t.Fatalf("resurrected-source rebuild: retained=%d err=%v", retained, err)
				}
			}
			restarted, err := researchpublicationstore.Wrap(backend)
			if err != nil {
				t.Fatal(err)
			}
			s.store = restarted
			if valid, err := s.ValidateResearchTransferSource(ctx, restarted.Snapshot(ctx), originals["survivor"], "lab-v1", key, now.Add(5*time.Second)); err == nil || valid {
				t.Fatal("new wrapper certified pre-wrapper source history")
			}
			if rebuilt, retained, err := researchmemory.RebuildFromBoundLabels(ctx, restarted.Snapshot(ctx), "tenant-a", 5, 46, cutoff, labels, validate); err == nil || rebuilt != nil || retained != 0 {
				t.Fatal("restart exposed transferred model without lineage history")
			}
		})
	}
}
