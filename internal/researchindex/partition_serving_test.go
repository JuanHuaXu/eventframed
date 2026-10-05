package researchindex

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"testing"
	"time"
)

func newPartitionTestServing(t *testing.T, retired, leases int) (*PartitionServing, string, string) {
	t.Helper()
	a, b := partitionID(0), partitionID(1)
	d, err := RestorePartitioned(2, 64, 2, 1, []Mutation{{ID: a, Vector: []float32{1, 0}}, {ID: b, Vector: []float32{0, 1}}}, func(context.Context, uint64, []Mutation) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewPartitionServing(context.Background(), d, t.TempDir()+"/initial", retired, leases)
	if err != nil {
		t.Fatal(err)
	}
	return s, a, b
}
func mustPartitionLease(t *testing.T, s *PartitionServing) *PartitionLease {
	t.Helper()
	l, e := s.Acquire(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	return l
}
func requireTop(t *testing.T, l *PartitionLease, q []float32, id string) {
	t.Helper()
	got, e := l.Search(context.Background(), q, 1)
	if e != nil || len(got) != 1 || got[0].ID != id {
		t.Fatal(got, e, id)
	}
}

func TestPartitionServingRetainsOldGraphAndBounds(t *testing.T) {
	ctx := context.Background()
	s, a, b := newPartitionTestServing(t, 1, 2)
	defer s.Close()
	old := mustPartitionLease(t, s)
	if e := s.Apply(ctx, []Mutation{{ID: a, Vector: []float32{-1, 0}}}); e != nil {
		t.Fatal(e)
	}
	if e := s.Compact(ctx, 0, t.TempDir()+"/next"); e != nil {
		t.Fatal(e)
	}
	now := mustPartitionLease(t, s)
	if now.Revision() != 2 || old.Revision() != 1 {
		t.Fatal("revision mismatch")
	}
	requireTop(t, old, []float32{1, 0}, a)
	requireTop(t, now, []float32{1, 0}, b)
	if old.handles[0] == now.handles[0] || old.handles[1] != now.handles[1] {
		t.Fatal("wrong handles replaced")
	}
	if _, e := s.Acquire(ctx); !errors.Is(e, ErrServingBusy) {
		t.Fatal("lease limit", e)
	}
	if e := s.Compact(ctx, 1, t.TempDir()+"/blocked"); !errors.Is(e, ErrServingBusy) {
		t.Fatal("retirement bound", e)
	}
	if e := s.Close(); !errors.Is(e, ErrServingBusy) {
		t.Fatal("closed active readers", e)
	}
	if e := old.Release(); e != nil {
		t.Fatal(e)
	}
	if old.owner != nil || old.handles != nil || old.view.Count() != 0 || old.Revision() != 1 {
		t.Fatal("released lease retains resources")
	}
	if e := old.Release(); !errors.Is(e, ErrFinished) {
		t.Fatal(e)
	}
	if e := s.Compact(ctx, 1, t.TempDir()+"/allowed"); e != nil {
		t.Fatal(e)
	}
	requireTop(t, now, []float32{1, 0}, b)
	if e := now.Release(); e != nil {
		t.Fatal(e)
	}
	if e := s.Close(); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Acquire(ctx); !errors.Is(e, ErrServingClosed) {
		t.Fatal(e)
	}
}

func TestPartitionServingRebasesDuringBuild(t *testing.T) {
	ctx := context.Background()
	s, a, b := newPartitionTestServing(t, 2, 8)
	defer s.Close()
	if e := s.Apply(ctx, []Mutation{{ID: a, Vector: []float32{.5, .5}}}); e != nil {
		t.Fatal(e)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	s.build = func(c context.Context, v View, dim int, path string) (*HNSWBase, error) {
		close(entered)
		<-release
		return BuildHNSWBase(c, v, dim, path)
	}
	done := make(chan error, 1)
	path := t.TempDir() + "/next"
	go func() { done <- s.Compact(ctx, 0, path) }()
	<-entered
	if e := s.Apply(ctx, []Mutation{{ID: a, Vector: []float32{-1, 0}}, {ID: b, Delete: true}}); e != nil {
		t.Fatal(e)
	}
	close(release)
	if e := <-done; e != nil {
		t.Fatal(e)
	}
	l := mustPartitionLease(t, s)
	defer l.Release()
	if l.Revision() != 3 {
		t.Fatal(l.Revision())
	}
	requireTop(t, l, []float32{-1, 0}, a)
	got, e := l.Search(ctx, []float32{-1, 0}, 10)
	if e != nil || len(got) != 1 {
		t.Fatal("deleted member returned", got, e)
	}
}

func TestPartitionServingFailedBuildDoesNotSwap(t *testing.T) {
	ctx := context.Background()
	s, a, _ := newPartitionTestServing(t, 2, 8)
	defer s.Close()
	old := s.current[0]
	s.build = func(context.Context, View, int, string) (*HNSWBase, error) { return nil, errors.New("build failed") }
	if e := s.Compact(ctx, 0, t.TempDir()+"/bad"); e == nil {
		t.Fatal("expected failure")
	}
	if s.current[0] != old || len(s.retired) != 0 {
		t.Fatal("failed build published")
	}
	l := mustPartitionLease(t, s)
	requireTop(t, l, []float32{1, 0}, a)
	if e := l.Release(); e != nil {
		t.Fatal(e)
	}
	s.build = BuildHNSWBase
	if e := s.Compact(ctx, 0, t.TempDir()+"/good"); e != nil {
		t.Fatal("build slot leaked", e)
	}
}

func TestPartitionServingClosingRemainsCharged(t *testing.T) {
	ctx := context.Background()
	s, _, _ := newPartitionTestServing(t, 1, 8)
	defer s.Close()
	l := mustPartitionLease(t, s)
	old := l.handles[0]
	if e := s.Compact(ctx, 0, t.TempDir()+"/next"); e != nil {
		t.Fatal(e)
	}
	old.index.mu.Lock()
	done := make(chan error, 1)
	go func() { done <- l.Release() }()
	deadline := time.Now().Add(time.Second)
	for {
		s.mu.Lock()
		closing := old.readers == 0
		s.mu.Unlock()
		if closing {
			break
		}
		if time.Now().After(deadline) {
			old.index.mu.Unlock()
			<-done
			t.Fatal("release stalled")
		}
		runtime.Gosched()
	}
	e := s.Compact(ctx, 1, t.TempDir()+"/blocked")
	old.index.mu.Unlock()
	if releaseErr := <-done; releaseErr != nil {
		t.Fatal(releaseErr)
	}
	if !errors.Is(e, ErrServingBusy) {
		t.Fatal("closing handle escaped retirement bound", e)
	}
	if e := s.Compact(ctx, 1, t.TempDir()+"/allowed"); e != nil {
		t.Fatal(e)
	}
}

func TestPartitionServingConcurrentQueriesAndCompaction(t *testing.T) {
	ctx := context.Background()
	s, a, b := newPartitionTestServing(t, 2, 8)
	defer s.Close()
	done := make(chan error, 1)
	root := t.TempDir()
	go func() {
		for i := 0; i < 20; i++ {
			if e := s.Apply(ctx, []Mutation{{ID: a, Vector: []float32{1, 0}}, {ID: b, Vector: []float32{0, 1}}}); e != nil {
				done <- e
				return
			}
			if e := s.Compact(ctx, i%2, fmt.Sprintf("%s/build-%d", root, i)); e != nil {
				done <- e
				return
			}
		}
		done <- nil
	}()
	for i := 0; i < 100; i++ {
		l, e := s.Acquire(ctx)
		if e != nil {
			t.Error(e)
			break
		}
		got, e := l.Search(ctx, []float32{1, 0}, 1)
		releaseErr := l.Release()
		if e != nil || releaseErr != nil || len(got) != 1 || got[0].ID != a {
			t.Error(got, e, releaseErr)
			break
		}
	}
	if e := <-done; e != nil {
		t.Fatal(e)
	}
}

func TestPartitionServingUnpublishedCloseFailureRemainsCharged(t *testing.T) {
	s, _, _ := newPartitionTestServing(t, 1, 8)
	defer func() {
		for _, h := range s.current {
			_ = h.index.Close()
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cleanupErr := errors.New("simulated unproven cleanup")
	s.build = func(c context.Context, v View, dim int, path string) (*HNSWBase, error) {
		h, err := BuildHNSWBase(c, v, dim, path)
		if err != nil {
			return nil, err
		}
		// Close real fixture storage, then inject the API's persistent close-error
		// state. This tests accounting without intentionally leaking a real DB.
		if err := h.Close(); err != nil {
			return nil, err
		}
		h.closeErr = cleanupErr
		cancel()
		return h, nil
	}
	err := s.Compact(ctx, 0, t.TempDir()+"/cancelled")
	if !errors.Is(err, context.Canceled) || !errors.Is(err, cleanupErr) {
		t.Fatalf("cleanup error hidden: %v", err)
	}
	if len(s.retired) != 1 {
		t.Fatal("unpublished resource no longer charged")
	}
	if err := s.Compact(context.Background(), 1, t.TempDir()+"/blocked"); !errors.Is(err, ErrServingBusy) {
		t.Fatal("cleanup uncertainty did not block new build", err)
	}
}
