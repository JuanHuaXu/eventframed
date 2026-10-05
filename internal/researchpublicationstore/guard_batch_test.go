package researchpublicationstore

import (
	"context"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchpublication"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestResearchAsOfBatchGuardMixedHorizons(t *testing.T) {
	for _, kind := range []string{"valid", "later-horizon-backfill", "policy", "unknown-history"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			s, err := New(memorystore.New())
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			at := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
			first := s.Snapshot(ctx)
			future := testutil.Event("future-first", "public batch fixture", at.Add(time.Hour))
			if _, err := s.Put(ctx, future, make([]float32, 8), "digest"); err != nil {
				t.Fatal(err)
			}
			second := s.Snapshot(ctx)
			if kind == "later-horizon-backfill" {
				middle := testutil.Event("later-backfill", "public batch fixture", at.Add(90*time.Minute))
				if _, err := s.Put(ctx, middle, make([]float32, 8), "digest"); err != nil {
					t.Fatal(err)
				}
			} else if kind == "policy" {
				if _, err := s.BindBayesianPolicy(ctx, "new-policy"); err != nil {
					t.Fatal(err)
				}
			} else if kind == "unknown-history" {
				s.publication = researchpublication.New(s.Snapshot(ctx))
			}
			checks := []AsOfSnapshot{{Captured: first, AsOf: at}, {Captured: second, AsOf: at.Add(2 * time.Hour)}}
			called := false
			err = s.WithResearchAsOfSnapshotsWait(ctx, checks, func() error {
				called = true
				if s.writer.TryAcquire(1) {
					s.writer.Release(1)
					t.Error("batch callback lacks writer exclusion")
				}
				return nil
			})
			if kind == "valid" {
				if err != nil || !called {
					t.Fatalf("valid batch rejected: called=%v err=%v", called, err)
				}
			} else if err == nil || called {
				t.Fatalf("invalid batch accepted: called=%v err=%v", called, err)
			}
		})
	}
}

func TestResearchAsOfBatchGuardInputs(t *testing.T) {
	s, err := New(memorystore.New())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	check := AsOfSnapshot{Captured: s.Snapshot(ctx), AsOf: time.Now()}
	if err := s.WithResearchAsOfSnapshotsWait(context.Background(), []AsOfSnapshot{check}, func() error { return nil }); err == nil {
		t.Fatal("missing deadline accepted")
	}
	if err := s.WithResearchAsOfSnapshotsWait(ctx, make([]AsOfSnapshot, 9), func() error { return nil }); err == nil {
		t.Fatal("unbounded batch accepted")
	}
	if err := s.WithResearchAsOfSnapshotsWait(ctx, []AsOfSnapshot{{Captured: check.Captured}}, func() error { return nil }); err == nil {
		t.Fatal("zero horizon accepted")
	}
}
