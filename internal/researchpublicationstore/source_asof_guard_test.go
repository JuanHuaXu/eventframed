package researchpublicationstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func sourceAsOfGuardFixture(t *testing.T) (*Store, model.Snapshot, time.Time) {
	t.Helper()
	s, err := New(memorystore.New())
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	if _, err := s.Put(context.Background(), testutil.Event("seed", "public query fixture", at.Add(-time.Minute)), []float32{1, 0}, "seed-digest"); err != nil {
		t.Fatal(err)
	}
	return s, s.Snapshot(context.Background()), at
}

func TestResearchSourceAsOfGuardWaitsWithoutReentry(t *testing.T) {
	s, from, at := sourceAsOfGuardFixture(t)
	defer s.Close()
	if !s.writer.TryAcquire(1) {
		t.Fatal("could not hold writer")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	entered := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- s.WithResearchSourceAsOfSnapshotWait(ctx, from, at, "tenant-a", "seed", func(target model.Snapshot) error {
			if target != from {
				return errors.New("unexpected guarded target")
			}
			close(entered)
			return nil
		})
	}()
	select {
	case <-entered:
		t.Fatal("callback entered while writer held")
	case <-time.After(10 * time.Millisecond):
	}
	s.writer.Release(1)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	default:
		t.Fatal("callback did not enter after release")
	}
	if !s.writer.TryAcquire(1) {
		t.Fatal("guard did not release permit")
	}
	called := false
	expired, stop := context.WithTimeout(context.Background(), time.Millisecond)
	err := s.WithResearchSourceAsOfSnapshotWait(expired, from, at, "tenant-a", "seed", func(model.Snapshot) error {
		called = true
		return nil
	})
	stop()
	s.writer.Release(1)
	if !errors.Is(err, context.DeadlineExceeded) || called {
		t.Fatalf("expired wait ran callback: called=%v err=%v", called, err)
	}
}

func TestResearchSourceAsOfGuardSeparatesFutureFromMutation(t *testing.T) {
	for _, mutation := range []string{"delete", "backfill"} {
		t.Run(mutation, func(t *testing.T) {
			s, from, at := sourceAsOfGuardFixture(t)
			defer s.Close()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			future := testutil.Event("future", "public query fixture", at.Add(time.Hour))
			if _, err := s.Put(ctx, future, []float32{1, 0}, "future-digest"); err != nil {
				t.Fatal(err)
			}
			if err := s.WithResearchSourceAsOfSnapshotWait(ctx, from, at, "tenant-a", "seed", func(target model.Snapshot) error {
				if target == from {
					return errors.New("future-only target did not move")
				}
				return nil
			}); err != nil {
				t.Fatalf("future-only history rejected: %v", err)
			}
			if mutation == "delete" {
				if _, err := s.Delete(ctx, "tenant-a", "seed"); err != nil {
					t.Fatal(err)
				}
			} else {
				backfill := testutil.Event("backfill", "public query fixture", at.Add(-30*time.Second))
				if _, err := s.Put(ctx, backfill, []float32{1, 0}, "backfill-digest"); err != nil {
					t.Fatal(err)
				}
			}
			called := false
			if err := s.WithResearchSourceAsOfSnapshotWait(ctx, from, at, "tenant-a", "seed", func(model.Snapshot) error {
				called = true
				return nil
			}); err == nil || called {
				t.Fatalf("%s retained source authority: called=%v err=%v", mutation, called, err)
			}
		})
	}
}
