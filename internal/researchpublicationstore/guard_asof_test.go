package researchpublicationstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchpublication"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestAsOfGuardMotionBoundary(t *testing.T) {
	for _, kind := range []string{"future", "equal", "backfill", "policy", "unknown-history", "bypass", "quarantine"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			backend := memorystore.New()
			s, err := New(backend)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			asOf := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
			captured := s.Snapshot(ctx)
			if kind == "policy" {
				_, err = s.BindBayesianPolicy(ctx, "changed")
			} else {
				at := asOf.Add(time.Hour)
				if kind == "equal" {
					at = asOf
				}
				if kind == "backfill" {
					at = asOf.Add(-time.Hour)
				}
				event := testutil.Event("public-asof", "public test fixture", at)
				if kind == "bypass" {
					_, err = backend.Put(ctx, event, make([]float32, 8), "digest")
				} else {
					_, err = s.Put(ctx, event, make([]float32, 8), "digest")
				}
				if err != nil {
					t.Fatal(err)
				}
				if kind == "unknown-history" {
					s.publication = researchpublication.New(s.Snapshot(ctx))
				}
				if kind == "quarantine" {
					if _, duplicateErr := s.Put(ctx, event, make([]float32, 8), "digest"); duplicateErr == nil {
						t.Fatal("duplicate accepted")
					}
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			called := false
			err = s.WithResearchAsOfSnapshotWait(ctx, captured, asOf, func() error {
				called = true
				if s.writer.TryAcquire(1) {
					s.writer.Release(1)
					t.Error("callback lacks mutation exclusion")
				}
				return nil
			})
			if kind == "future" {
				if err != nil || !called {
					t.Fatal("future-only history rejected", err)
				}
			} else if err == nil || called {
				t.Fatal("invalid history accepted", kind, err)
			}
			if err = s.WithResearchSnapshotWait(ctx, captured, func() error { t.Error("exact guard relaxed"); return nil }); err == nil {
				t.Fatal("exact guard accepted motion")
			}
		})
	}
}

func TestAsOfGuardInputs(t *testing.T) {
	s, err := New(memorystore.New())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	snapshot := s.Snapshot(ctx)
	work := func() error { t.Error("invalid input entered callback"); return nil }
	if s.WithResearchAsOfSnapshotWait(context.Background(), snapshot, time.Now(), work) == nil {
		t.Fatal("missing deadline accepted")
	}
	if s.WithResearchAsOfSnapshotWait(ctx, snapshot, time.Time{}, work) == nil {
		t.Fatal("zero as-of accepted")
	}
	if s.WithResearchAsOfSnapshotWait(ctx, snapshot, time.Now(), nil) == nil {
		t.Fatal("nil callback accepted")
	}
	cancel()
	if s.WithResearchAsOfSnapshotWait(ctx, snapshot, time.Now(), work) != context.Canceled {
		t.Fatal("cancellation ignored")
	}
}

func TestAsOfGuardWaitAndRelease(t *testing.T) {
	s, err := New(memorystore.New())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	bg := context.Background()
	snapshot := s.Snapshot(bg)
	if err = s.writer.Acquire(bg, 1); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(bg, 20*time.Millisecond)
	err = s.WithResearchAsOfSnapshotWait(ctx, snapshot, time.Now(), func() error { t.Error("expired waiter entered"); return nil })
	cancel()
	s.writer.Release(1)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(bg, time.Second)
	defer cancel()
	want := errors.New("callback failure")
	if err = s.WithResearchAsOfSnapshotWait(ctx, snapshot, time.Now(), func() error { return want }); !errors.Is(err, want) {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("missing panic")
			}
		}()
		_ = s.WithResearchAsOfSnapshotWait(ctx, snapshot, time.Now(), func() error { panic("fixture") })
	}()
	if err = s.WithResearchAsOfSnapshotWait(ctx, snapshot, time.Now(), func() error { return nil }); err != nil {
		t.Fatal("callback retained gate", err)
	}
}
