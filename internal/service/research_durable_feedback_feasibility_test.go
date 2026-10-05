package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
)

func TestResearchGuardedDurableFeedbackFeasibilityV2(t *testing.T) {
	s, _, now := temporalBridgeFixture(t, true)
	wrapped, err := researchpublicationstore.Wrap(s.store)
	if err != nil {
		t.Fatal(err)
	}
	s.store = wrapped
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	path := t.TempDir() + "/guarded-feedback.sqlite"
	durable, err := researchmemory.OpenDurable(ctx, path, "tenant-a", "guarded", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer durable.Close()
	preview := researchmemory.New(1, 42)

	issue := func(index int) (model.RecallRequest, researchmemory.RecordedPrediction) {
		t.Helper()
		queryAt := now.Add(time.Duration(2*index) * time.Second)
		request := model.RecallRequest{
			ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a",
			SessionID: fmt.Sprintf("guarded-feedback-%d", index),
			Query:     "public query fixture", AsOf: queryAt,
			RecallK: 3, PackK: 2, TokenBudget: 1000,
		}
		if _, err := s.Recall(ctx, request); err != nil {
			t.Fatal(err)
		}
		in, err := s.config.ResearchFrontier.Take(ctx)
		if err != nil || len(in.Candidates) != 1 || in.Candidates[0].EventID != "seed" {
			t.Fatal("unexpected frontier", in, err)
		}
		candidate := in.Candidates[0]
		prediction, err := preview.Predict(candidate.Features, candidate.Baseline, 1, in.AsOf)
		if err != nil || prediction.ID != uint64(index+1) {
			t.Fatal("preview identity", prediction, err)
		}
		envelope, err := preview.Record(prediction.ID)
		if err != nil {
			t.Fatal(err)
		}
		binding := researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: candidate.EventID, Snapshot: in.Snapshot}
		envelope.Binding = &binding
		var original researchmemory.RecordedPrediction
		err = s.WithValidatedResearchAsOfAdmission(ctx, envelope, request, func() error {
			_, retry, err := durable.AdmitBound(ctx, uint64(index+1), candidate.Features, candidate.Baseline, in.AsOf, binding)
			if err != nil || retry {
				return fmt.Errorf("durable admission retry=%v: %w", retry, err)
			}
			original, err = durable.Admission(ctx, uint64(index+1))
			if err != nil {
				return err
			}
			return s.ValidateResearchAdmission(ctx, original, request)
		})
		if err != nil || original.Binding == nil || *original.Binding != binding {
			t.Fatal("bound durable admission", err, original)
		}
		return request, original
	}

	for index := 0; index < 32; index++ {
		request, original := issue(index)
		available := request.AsOf.Add(time.Second)
		err := s.WithValidatedResearchAsOfAdmission(ctx, original, request, func() error {
			_, err := durable.Feedback(ctx, uint64(index+1), index%3 != 0, available)
			return err
		})
		if err != nil {
			t.Fatalf("guarded feedback %d: %v", index, err)
		}
		if index == 15 {
			putTemporalFixture(t, s, "future", now.Add(time.Hour))
		}
	}
	if err := durable.Close(); err != nil {
		t.Fatal(err)
	}
	log, err := researchledger.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := log.ReadAfter(ctx, 0, 128)
	if err != nil {
		log.Close()
		t.Fatal(err)
	}
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 64 {
		t.Fatalf("expected 32 admissions and 32 terminals, got %d rows", len(rows))
	}
	journalIDs := map[string]bool{}
	for i := 0; i < 32; i++ {
		admit, terminal := rows[2*i], rows[2*i+1]
		if admit.Kind != "admit" || terminal.Kind != "feedback" || admit.Key != terminal.Key || admit.Key.Event != fmt.Sprint(i+1) {
			t.Fatalf("wrong ledger order/identity at %d", i)
		}
		var original researchmemory.RecordedPrediction
		var label researchmemory.RecordedFeedback
		if err := json.Unmarshal(admit.Payload, &original); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(terminal.Payload, &label); err != nil {
			t.Fatal(err)
		}
		if original.Binding == nil || original.Binding.Tenant != "tenant-a" || original.Binding.EventID != "seed" || original.Binding.JournalID == "" || journalIDs[original.Binding.JournalID] ||
			!original.At.Equal(now.Add(time.Duration(2*i)*time.Second)) || label.ID != uint64(i+1) || label.Useful != (i%3 != 0) || !label.Available.Equal(original.At.Add(time.Second)) {
			t.Fatalf("wrong persisted source or label at %d", i)
		}
		journalIDs[original.Binding.JournalID] = true
	}

	replay := func(wantPending int) {
		t.Helper()
		log, err := researchledger.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer log.Close()
		start := time.Now()
		restored, err := researchmemory.ResumeBackground(ctx, log, "tenant-a", "guarded", 1, 42, 64)
		if err != nil {
			t.Fatal("same-epoch durable replay", err)
		}
		replayTime := time.Since(start)
		defer restored.Close()
		if completed, failed, pending, queued := restored.Counts(); completed != 32 || failed != 0 || pending != wantPending || queued != 0 {
			t.Fatalf("replay accounting: completed=%d failed=%d pending=%d queued=%d", completed, failed, pending, queued)
		}
		for x := uint16(0); x < 512; x++ {
			p, err := restored.Snapshot().Score(x, .6, 1, now.Add(2*time.Minute))
			if err != nil || math.IsNaN(p) || p <= 0 || p >= 1 {
				t.Fatalf("invalid restored score for %d: %g, %v", x, p, err)
			}
		}
		t.Logf("replayed 32 explicit feedback records in %s; validation including 512 scores in %s", replayTime, time.Since(start))
	}
	replay(0)

	durable, err = researchmemory.OpenDurable(ctx, path, "tenant-a", "guarded", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	request, unresolved := issue(32)
	if _, err := durable.Admission(ctx, 33); err != nil {
		t.Fatal("original pending record was not durable", err)
	}
	deleted, err := s.Delete(ctx, model.DeleteRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", EventID: "seed"})
	if err != nil || !deleted.Deleted {
		t.Fatal("visible deletion failed", err)
	}
	callbackRan := false
	err = s.WithValidatedResearchAsOfAdmission(ctx, unresolved, request, func() error {
		callbackRan = true
		_, err := durable.Feedback(ctx, 33, false, request.AsOf.Add(time.Second))
		return err
	})
	if err == nil || callbackRan {
		t.Fatal("stale feedback reached durable terminal write")
	}
	if _, err := durable.Discard(ctx, 33, now.Add(2*time.Minute)); err != nil {
		t.Fatal("unresolved label was not discarded", err)
	}
	if err := durable.Close(); err != nil {
		t.Fatal(err)
	}
	replay(0)
	if _, err := researchmemory.OpenDurable(ctx, path, "tenant-a", "guarded", 2, 42); err == nil {
		t.Fatal("old-epoch log opened as a new-epoch learner")
	}
}
