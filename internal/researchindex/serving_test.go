package researchindex

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"
)

func TestServingPublishesPairAndRetiresAfterReader(t *testing.T) {
	ctx := context.Background()
	d, _ := RestoreDurable(2, 8, 1, []Mutation{{ID: "a", Vector: []float32{1, 0}}}, func(context.Context, uint64, []Mutation) error { return nil })
	s, err := NewServing(ctx, d, t.TempDir()+"/initial", 1)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	old, err := s.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Apply(ctx, []Mutation{{ID: "a", Delete: true}, {ID: "b", Vector: []float32{1, 0}}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Compact(ctx, t.TempDir()+"/next"); err != nil {
		t.Fatal(err)
	}
	newer, err := s.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got, err := newer.Search(ctx, []float32{1, 0}, 1)
	if err != nil || len(got) != 1 || got[0].ID != "b" {
		t.Fatal(got, err)
	}
	got, err = old.Search(ctx, []float32{1, 0}, 1)
	if err != nil || len(got) != 1 || got[0].ID != "a" {
		t.Fatal(got, err)
	}
	if err := s.Compact(ctx, t.TempDir()+"/blocked"); !errors.Is(err, ErrServingBusy) {
		t.Fatal("unbounded retired generations", err)
	}
	if err := s.Close(); !errors.Is(err, ErrServingBusy) {
		t.Fatal(err)
	}
	oldIndex := old.handle.index
	if err := old.Release(); err != nil {
		t.Fatal(err)
	}
	if !oldIndex.closed {
		t.Fatal("retired index not closed")
	}
	if _, err := old.Search(ctx, []float32{1, 0}, 1); !errors.Is(err, ErrFinished) {
		t.Fatal(err)
	}
	if err := newer.Release(); err != nil {
		t.Fatal(err)
	}
	if err := s.Compact(ctx, t.TempDir()+"/final"); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Acquire(ctx); !errors.Is(err, ErrServingClosed) {
		t.Fatal(err)
	}
}

func TestServingRetirementBoundIncludesClosingIndex(t *testing.T) {
	ctx := context.Background()
	d, _ := RestoreDurable(2, 8, 1, []Mutation{{ID: "a", Vector: []float32{1, 0}}}, func(context.Context, uint64, []Mutation) error { return nil })
	s, err := NewServing(ctx, d, t.TempDir()+"/initial", 1)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	old, err := s.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Compact(ctx, t.TempDir()+"/replacement"); err != nil {
		t.Fatal(err)
	}
	// Keep Close blocked after the final reader detaches, without delaying the
	// serving mutex. The retirement budget must still account for this graph.
	handle := old.handle
	handle.index.mu.Lock()
	done := make(chan error, 1)
	go func() { done <- old.Release() }()
	detached := false
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		detached = handle.readers == 0
		s.mu.Unlock()
		if detached {
			break
		}
		runtime.Gosched()
	}
	err = s.Compact(ctx, t.TempDir()+"/must-not-build")
	handle.index.mu.Unlock()
	if releaseErr := <-done; releaseErr != nil {
		t.Fatal(releaseErr)
	}
	if !detached {
		t.Fatal("release did not reach close")
	}
	if !errors.Is(err, ErrServingBusy) {
		t.Fatal("closing graph escaped retirement bound", err)
	}
}

func TestServingBuildAllowsWritesAndQuarantinesFailedPublication(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "unknown"}[fail], func(t *testing.T) {
			ctx := context.Background()
			d, _ := RestoreDurable(2, 8, 1, []Mutation{{ID: "a", Vector: []float32{1, 0}}}, func(context.Context, uint64, []Mutation) error {
				if fail {
					return errors.New("uncertain storage")
				}
				return nil
			})
			s, err := NewServing(ctx, d, t.TempDir()+"/initial", 2)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			old, err := s.Acquire(ctx)
			if err != nil {
				t.Fatal(err)
			}
			entered, release := make(chan struct{}), make(chan struct{})
			s.build = func(c context.Context, v View, dim int, path string) (*HNSWBase, error) {
				close(entered)
				<-release
				return BuildHNSWBase(c, v, dim, path)
			}
			done := make(chan error, 1)
			path := t.TempDir() + "/next"
			go func() { done <- s.Compact(ctx, path) }()
			<-entered
			if err := s.Close(); !errors.Is(err, ErrServingBusy) {
				t.Fatal(err)
			}
			err = s.Apply(ctx, []Mutation{{ID: "a", Delete: true}, {ID: "b", Vector: []float32{1, 0}}})
			if fail && !errors.Is(err, ErrRecoveryRequired) || !fail && err != nil {
				t.Fatal(err)
			}
			close(release)
			err = <-done
			if fail {
				if !errors.Is(err, ErrRecoveryRequired) {
					t.Fatal(err)
				}
				if _, err := s.Acquire(ctx); !errors.Is(err, ErrRecoveryRequired) {
					t.Fatal(err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				fresh, err := s.Acquire(ctx)
				if err != nil {
					t.Fatal(err)
				}
				got, err := fresh.Search(ctx, []float32{1, 0}, 1)
				if err != nil || len(got) != 1 || got[0].ID != "b" {
					t.Fatal("newer delta lost", got, err)
				}
				if err := fresh.Release(); err != nil {
					t.Fatal(err)
				}
			}
			got, err := old.Search(ctx, []float32{1, 0}, 1)
			if err != nil || len(got) != 1 || got[0].ID != "a" {
				t.Fatal(got, err)
			}
			if err := old.Release(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
