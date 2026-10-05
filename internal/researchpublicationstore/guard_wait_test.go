package researchpublicationstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

func TestQueuedSnapshotGuard(t *testing.T) {
	s, err := New(memorystore.New())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	bg := context.Background()
	snapshot := s.Snapshot(bg)
	if err = s.WithResearchSnapshotWait(bg, snapshot, func() error { return nil }); err == nil {
		t.Fatal("unbounded wait accepted")
	}
	if err = s.writer.Acquire(bg, 1); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(bg, 20*time.Millisecond)
	err = s.WithResearchSnapshotWait(ctx, snapshot, func() error { t.Error("expired callback entered"); return nil })
	cancel()
	s.writer.Release(1)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}

	if err = s.writer.Acquire(bg, 1); err != nil {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(bg, time.Second)
	defer cancel()
	started, entered, done := make(chan struct{}), make(chan struct{}, 1), make(chan error, 1)
	go func() {
		close(started)
		done <- s.WithResearchSnapshotWait(ctx, snapshot, func() error {
			entered <- struct{}{}
			if s.writer.TryAcquire(1) {
				s.writer.Release(1)
				return errors.New("queued callback lacks exclusion")
			}
			return nil
		})
	}()
	<-started
	select {
	case <-entered:
		t.Error("entered while writer owns gate")
	case <-time.After(20 * time.Millisecond):
	}
	s.writer.Release(1)
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	default:
		t.Fatal("queued callback never entered")
	}
	if _, err = s.BindBayesianPolicy(bg, "new-policy"); err != nil {
		t.Fatal(err)
	}
	if err = s.WithResearchSnapshotWait(ctx, snapshot, func() error { t.Error("stale callback entered"); return nil }); err == nil {
		t.Fatal("stale guard accepted")
	}
	current := s.Snapshot(bg)
	want := errors.New("callback error")
	if err = s.WithResearchSnapshotWait(ctx, current, func() error { return want }); !errors.Is(err, want) {
		t.Fatal(err)
	}
	if err = s.WithResearchSnapshotWait(ctx, current, func() error { return nil }); err != nil {
		t.Fatal("callback error retained token", err)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("missing panic")
			}
		}()
		_ = s.WithResearchSnapshotWait(ctx, current, func() error { panic("fixture") })
	}()
	if err = s.WithResearchSnapshotWait(ctx, current, func() error { return nil }); err != nil {
		t.Fatal("panic retained token", err)
	}
}
