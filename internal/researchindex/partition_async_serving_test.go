package researchindex

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newAsyncPartitionTest(t *testing.T) *AsyncPartitionServing {
	t.Helper()
	d, err := RestorePartitioned(2, 64, 2, 1, []Mutation{{ID: partitionID(0), Vector: []float32{1, 0}}, {ID: partitionID(1), Vector: []float32{0, 1}}}, func(context.Context, uint64, []Mutation) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewAsyncPartitionServing(context.Background(), d, t.TempDir()+"/initial", 1, 8)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAsyncPartitionReleaseDoesNotRunCloseInline(t *testing.T) {
	ctx := context.Background()
	s := newAsyncPartitionTest(t)
	defer s.Close()
	l, err := s.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	old := l.handles[0]
	if err := s.Compact(ctx, 0, t.TempDir()+"/next"); err != nil {
		t.Fatal(err)
	}
	old.index.mu.Lock()
	locked := true
	defer func() {
		if locked {
			old.index.mu.Unlock()
		}
	}()
	released := make(chan error, 1)
	go func() { released <- l.Release() }()
	select {
	case err := <-released:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("release waited for blocked close")
	}
	if l.owner != nil || l.handles != nil || l.view.Count() != 0 {
		t.Fatal("released lease retains state")
	}
	if err := s.Compact(ctx, 1, t.TempDir()+"/blocked"); !errors.Is(err, ErrServingBusy) {
		t.Fatal("pending close escaped cap", err)
	}
	current, err := s.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got, err := current.Search(ctx, []float32{1, 0}, 1)
	if err != nil || len(got) != 1 || got[0].ID != partitionID(0) {
		t.Fatal(got, err)
	}
	if err := current.Release(); err != nil {
		t.Fatal(err)
	}
	closed := make(chan error, 1)
	go func() { closed <- s.Close() }()
	select {
	case err := <-closed:
		t.Fatal("shutdown did not join blocked cleanup", err)
	case <-time.After(10 * time.Millisecond):
	}
	old.index.mu.Unlock()
	locked = false
	if err := <-closed; err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestAsyncPartitionCleanupFailureSurfacedAtShutdown(t *testing.T) {
	ctx := context.Background()
	s := newAsyncPartitionTest(t)
	l, err := s.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	old := l.handles[0]
	if err := s.Compact(ctx, 0, t.TempDir()+"/next"); err != nil {
		t.Fatal(err)
	}
	if err := old.index.Close(); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("unproven close")
	old.index.closeErr = failure
	if err := l.Release(); err != nil {
		t.Fatal("release should report enqueue only", err)
	}
	if err := s.Close(); !errors.Is(err, failure) {
		t.Fatal("async cleanup failure lost", err)
	}
	if len(s.retired) != 1 {
		t.Fatal("failed resource uncharged")
	}
}

func TestAsyncPartitionCancelledCandidateCleanup(t *testing.T) {
	s := newAsyncPartitionTest(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	failure := errors.New("unpublished cleanup error")
	s.build = func(c context.Context, v View, dim int, path string) (*HNSWBase, error) {
		h, err := BuildHNSWBase(c, v, dim, path)
		if err != nil {
			return nil, err
		}
		if err := h.Close(); err != nil {
			return nil, err
		}
		h.closeErr = failure
		cancel()
		return h, nil
	}
	if err := s.Compact(ctx, 0, t.TempDir()+"/cancelled"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := s.Close(); !errors.Is(err, failure) {
		t.Fatal("unpublished cleanup hidden", err)
	}
}
