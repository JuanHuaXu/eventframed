package researchindex

import (
	"context"
	"errors"
	"testing"
)

func TestPrivateGraphWriterRetention(t *testing.T) {
	ctx := context.Background()
	calls := 0
	w, e := NewPrivateGraphWriter(LayeredSnapshot{}, 0, 2, 4, 32, 2, 1, func(_ context.Context, rev uint64, m []Mutation) error {
		calls++
		if rev != uint64(calls) {
			t.Fatal("revision")
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = w.Insert(ctx, "a", []float32{1, 0}, 0); e != nil {
		t.Fatal(e)
	}
	old, e := w.Acquire(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = w.Insert(ctx, "b", []float32{0, 1}, 0); e != nil {
		t.Fatal(e)
	}
	current, e := w.Acquire(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = w.Acquire(ctx); !errors.Is(e, ErrCapacity) {
		t.Fatal("reader cap", e)
	}
	if _, e = w.Insert(ctx, "c", []float32{1, 1}, 0); !errors.Is(e, ErrCapacity) || calls != 2 {
		t.Fatal("retirement preflight", e, calls)
	}
	if _, ok, e := old.Lookup(1); e != nil || ok {
		t.Fatal("old lease changed")
	}
	if rev, e := old.Revision(); e != nil || rev != 1 {
		t.Fatal(rev, e)
	}
	old.Release()
	old.Release()
	if _, _, e = old.Lookup(0); !errors.Is(e, ErrFinished) {
		t.Fatal(e)
	}
	if _, e = w.Delete(ctx, "a"); e != nil {
		t.Fatal(e)
	}
	if _, ok, e := current.Lookup(0); e != nil || !ok {
		t.Fatal("old deleted node missing")
	}
	current.Release()
	now, e := w.Acquire(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer now.Release()
	if _, ok, e := now.Lookup(0); e != nil || ok {
		t.Fatal("deletion not published")
	}
	if _, e = w.Insert(ctx, "b", []float32{1, 1}, 0); e == nil || calls != 3 {
		t.Fatal("duplicate persisted")
	}
}

func TestPrivateGraphWriterCommitBoundary(t *testing.T) {
	for _, mode := range []string{"success", "error", "panic"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			entered := make(chan struct{})
			proceed := make(chan struct{})
			w, e := NewPrivateGraphWriter(LayeredSnapshot{}, 0, 2, 4, 32, 4, 2, func(_ context.Context, _ uint64, m []Mutation) error {
				close(entered)
				<-proceed
				m[0].Vector[0] = 99
				cancel()
				if mode == "error" {
					return errors.New("lost acknowledgement")
				}
				if mode == "panic" {
					panic("after commit")
				}
				return nil
			})
			if e != nil {
				t.Fatal(e)
			}
			done := make(chan error, 1)
			go func() { _, err := w.Insert(ctx, "a", []float32{1, 0}, 0); done <- err }()
			<-entered
			old, e := w.Acquire(context.Background())
			if e != nil {
				t.Fatal(e)
			}
			if rev, _ := old.Revision(); rev != 0 {
				t.Fatal("published before callback")
			}
			close(proceed)
			e = <-done
			if mode == "success" {
				if e != nil {
					t.Fatal(e)
				}
				fresh, e := w.Acquire(context.Background())
				if e != nil {
					t.Fatal(e)
				}
				r, ok, e := fresh.Lookup(0)
				if e != nil || !ok || r.Vector[0] != 1 {
					t.Fatal("callback alias", r, e)
				}
				fresh.Release()
			} else {
				if e == nil {
					t.Fatal("failure ignored")
				}
				if _, e = w.Acquire(context.Background()); !errors.Is(e, ErrRecoveryRequired) {
					t.Fatal(e)
				}
				if _, e = w.Insert(context.Background(), "b", []float32{1, 1}, 0); !errors.Is(e, ErrRecoveryRequired) {
					t.Fatal(e)
				}
			}
			if rev, _ := old.Revision(); rev != 0 {
				t.Fatal("old version changed")
			}
			old.Release()
		})
	}
}

func TestPrivateGraphWriterPrepareFailure(t *testing.T) {
	calls := 0
	w, e := NewPrivateGraphWriter(LayeredSnapshot{}, 0, 2, 4, 32, 4, 2, func(context.Context, uint64, []Mutation) error { calls++; return nil })
	if e != nil {
		t.Fatal(e)
	}
	if _, e = w.Insert(context.Background(), "bad", []float32{0, 0}, 0); e == nil || calls != 0 {
		t.Fatal("invalid vector persisted")
	}
	if _, e = w.Insert(context.Background(), "good", []float32{1, 0}, 0); e != nil || calls != 1 {
		t.Fatal("prepare failure poisoned owner", e)
	}
}
