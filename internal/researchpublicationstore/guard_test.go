package researchpublicationstore

import (
	"context"
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"testing"
	"time"
)

func TestSnapshotGuardConcurrentWriter(t *testing.T) {
	ctx := context.Background()
	s, err := New(memorystore.New())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	original := s.Snapshot(ctx)
	started := make(chan struct{})
	done := make(chan error, 1)
	err = s.WithResearchSnapshot(ctx, original, func() error {
		go func() {
			close(started)
			_, err := s.BindBayesianPolicy(ctx, "concurrent-writer")
			done <- err
		}()
		<-started
		select {
		case err := <-done:
			return errors.Join(errors.New("writer completed inside snapshot guard"), err)
		case <-time.After(20 * time.Millisecond):
		}
		if s.Snapshot(ctx) != original {
			return errors.New("snapshot moved inside guard")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("writer did not resume after guard")
	}
	if s.Snapshot(ctx) == original {
		t.Fatal("writer did not change snapshot")
	}
}

func TestSnapshotGuardExcludesMutations(t *testing.T) {
	ctx := context.Background()
	s, e := New(memorystore.New())
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	original := s.Snapshot(ctx)
	called := 0
	errFixture := errors.New("callback failure")
	e = s.WithResearchSnapshot(ctx, original, func() error {
		called++
		if s.writer.TryAcquire(1) {
			s.writer.Release(1)
			t.Fatal("mutation lock not held")
		}
		if e = s.WithResearchSnapshot(ctx, original, func() error { t.Fatal("nested guard entered"); return nil }); e == nil {
			t.Fatal("busy guard accepted")
		}
		return errFixture
	})
	if !errors.Is(e, errFixture) || called != 1 {
		t.Fatal(e, called)
	}
	if _, e = s.BindBayesianPolicy(ctx, "next"); e != nil {
		t.Fatal("error failed to release guard", e)
	}
	if e = s.WithResearchSnapshot(ctx, original, func() error { called++; return nil }); e == nil || called != 1 {
		t.Fatal("stale callback entered", e, called)
	}
	current := s.Snapshot(ctx)
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("panic missing")
			}
		}()
		_ = s.WithResearchSnapshot(ctx, current, func() error { panic("fixture") })
	}()
	if e = s.WithResearchSnapshot(ctx, current, func() error { return nil }); e != nil {
		t.Fatal("panic retained guard", e)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if e = s.WithResearchSnapshot(cancelled, current, func() error { t.Fatal("cancelled callback"); return nil }); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
}
