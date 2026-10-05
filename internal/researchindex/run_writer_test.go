package researchindex

import (
	"context"
	"errors"
	"math"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunWriterCommitBoundary(t *testing.T) {
	ctx := context.Background()
	entered, release := make(chan struct{}), make(chan struct{})
	var rev uint64
	w, err := NewDurableRunWriter(ctx, nil, 0, 2, 64, 1, func(_ context.Context, r uint64, m []Mutation) error {
		rev = r
		close(entered)
		<-release
		m[0].Vector[0] = 99
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	old, err := w.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- w.Apply(ctx, []Mutation{{ID: "a", Vector: []float32{1, 1}}}) }()
	<-entered
	short, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	if _, err = w.Snapshot(short); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("uncommitted snapshot exposed", err)
	}
	close(release)
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	s, err := w.Snapshot(ctx)
	if err != nil || s.Revision() != 1 || rev != 1 {
		t.Fatal(s, rev, err)
	}
	got, err := s.Search(ctx, []float32{1, 0})
	if err != nil || len(got) != 1 || math.Abs(got[0].Score-1/math.Sqrt(2)) > 1e-15 {
		t.Fatal("callback changed published state", got, err)
	}
	prior, err := old.Search(ctx, []float32{1, 0})
	if err != nil || len(prior) != 0 || old.Revision() != 0 {
		t.Fatal("historical view changed", prior, err)
	}
}

func TestRunWriterUnknownCommitQuarantine(t *testing.T) {
	ctx := context.Background()
	for _, panicMode := range []bool{false, true} {
		w, err := NewDurableRunWriter(ctx, nil, 0, 2, 64, 1, func(context.Context, uint64, []Mutation) error {
			if panicMode {
				panic("lost commit result")
			}
			return errors.New("lost ack")
		})
		if err != nil {
			t.Fatal(err)
		}
		old, _ := w.Snapshot(ctx)
		func() {
			defer func() {
				r := recover()
				if panicMode && r == nil {
					t.Error("expected panic")
				}
				if !panicMode && r != nil {
					t.Error(r)
				}
			}()
			err = w.Apply(ctx, []Mutation{{ID: "a", Vector: []float32{1, 0}}})
			if !errors.Is(err, ErrRecoveryRequired) {
				t.Error(err)
			}
		}()
		if _, err = w.Snapshot(ctx); !errors.Is(err, ErrRecoveryRequired) {
			t.Fatal(err)
		}
		if err = w.Apply(ctx, []Mutation{{ID: "b", Delete: true}}); !errors.Is(err, ErrRecoveryRequired) {
			t.Fatal(err)
		}
		got, err := old.Search(ctx, []float32{1, 0})
		if err != nil || len(got) != 0 {
			t.Fatal(got, err)
		}
	}
}

func TestRunWriterPreflightAndLateCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int64
	w, err := NewDurableRunWriter(ctx, nil, 0, 2, 1, 1, func(context.Context, uint64, []Mutation) error { calls.Add(1); cancel(); return nil })
	if err != nil {
		t.Fatal(err)
	}
	if err = w.Apply(ctx, []Mutation{{ID: "a", Vector: []float32{0, 0}}}); err == nil {
		t.Fatal("invalid vector accepted")
	}
	if calls.Load() != 0 {
		t.Fatal("invalid write persisted")
	}
	if err = w.Apply(ctx, []Mutation{{ID: "a", Vector: []float32{1, 0}}}); err != nil {
		t.Fatal("late cancellation hid commit", err)
	}
	s, err := w.Snapshot(context.Background())
	if err != nil || s.Revision() != 1 {
		t.Fatal(s, err)
	}
	if err = w.Apply(context.Background(), []Mutation{{ID: "b", Delete: true}}); !errors.Is(err, ErrCapacity) {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatal("capacity rejection persisted")
	}
}
