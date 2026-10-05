package service

import (
	"context"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
)

func TestResearchFeedbackMixedMutationRebind(t *testing.T) {
	for _, persistent := range []bool{false, true} {
		name := "memory"
		if persistent {
			name = "persistent"
		}
		t.Run(name, func(t *testing.T) {
			s, old, now := temporalBridgeFixture(t, persistent)
			wrapped, err := researchpublicationstore.Wrap(s.store)
			if err != nil {
				t.Fatal(err)
			}
			s.store = wrapped
			ctx := context.Background()

			first := takeTemporalFixture(t, s, now)
			if len(first.Candidates) != 1 || first.Candidates[0].EventID != "seed" {
				t.Fatalf("unexpected original frontier: %+v", first.Candidates)
			}
			if _, err := old.Admit(ctx, first); err != nil {
				t.Fatal(err)
			}
			putTemporalFixture(t, s, "future", now.Add(time.Hour))
			if err := old.Feedback(ctx, first.JournalID, "seed", true, now.Add(time.Second)); err != nil {
				t.Fatal("future-only insertion invalidated old as-of feedback", err)
			}
			wait, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			if err := old.worker.WaitProcessed(wait, 1); err != nil {
				t.Fatal("first label did not complete", err)
			}
			if _, err := old.Score(ctx, 1, .6, now.Add(2*time.Hour)); err == nil {
				t.Fatal("future-visible score accepted")
			}

			second := takeTemporalFixture(t, s, now.Add(2*time.Second))
			if second.JournalID == first.JournalID || len(second.Candidates) != 1 || second.Candidates[0].EventID != "seed" {
				t.Fatal("second frontier did not bind a fresh journal")
			}
			if _, err := old.Admit(ctx, second); err != nil {
				t.Fatal(err)
			}
			deleted, err := s.Delete(ctx, model.DeleteRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", EventID: "seed"})
			if err != nil || !deleted.Deleted {
				t.Fatal("visible delete failed", err)
			}
			if err := old.Feedback(ctx, second.JournalID, "seed", false, now.Add(3*time.Second)); err == nil {
				t.Fatal("old bridge accepted feedback after visible deletion")
			}
			if _, err := old.Score(ctx, 1, .6, now.Add(3*time.Second)); err == nil {
				t.Fatal("old bridge scored after visible deletion")
			}
			if completed, failed, _, _ := old.worker.Counts(); completed != 1 || failed != 0 {
				t.Fatalf("rejected label changed worker state: %d completed, %d failed", completed, failed)
			}

			putTemporalFixture(t, s, "reseed", now.Add(-time.Minute))
			fresh, err := NewResearchTemporalFeedbackBridge(s, "tenant-a", 43)
			if err != nil {
				t.Fatal(err)
			}
			defer fresh.Close()
			third := takeTemporalFixture(t, s, now.Add(4*time.Second))
			if len(third.Candidates) != 1 || third.Candidates[0].EventID != "reseed" {
				t.Fatalf("fresh frontier contains stale/deleted support: %+v", third.Candidates)
			}
			if _, err := old.Admit(ctx, third); err == nil {
				t.Fatal("old bridge accepted a new-epoch journal")
			}
			if _, err := fresh.Admit(ctx, third); err != nil {
				t.Fatal("fresh bridge rejected current journal", err)
			}
			if err := fresh.Feedback(ctx, third.JournalID, "reseed", true, now.Add(5*time.Second)); err != nil {
				t.Fatal("fresh feedback rejected", err)
			}
			freshWait, freshCancel := context.WithTimeout(ctx, time.Second)
			defer freshCancel()
			if err := fresh.worker.WaitProcessed(freshWait, 1); err != nil {
				t.Fatal("fresh label did not complete", err)
			}
			if completed, failed, _, _ := fresh.worker.Counts(); completed != 1 || failed != 0 {
				t.Fatalf("fresh worker accounting: %d completed, %d failed", completed, failed)
			}
			if err := old.Feedback(ctx, second.JournalID, "seed", false, now.Add(6*time.Second)); err == nil {
				t.Fatal("old pending feedback revived after bridge rebind")
			}
		})
	}
}
