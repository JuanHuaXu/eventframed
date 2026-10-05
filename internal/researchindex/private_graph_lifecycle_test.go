package researchindex

import (
	"context"
	"errors"
	"testing"
)

func TestPrivateGraphClose(t *testing.T) {
	ctx := context.Background()
	w, e := NewPrivateGraphWriter(LayeredSnapshot{}, 0, 2, 4, 32, 2, 1, func(context.Context, uint64, []Mutation) error { return nil })
	if e != nil {
		t.Fatal(e)
	}
	lease, e := w.Acquire(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if e = w.Close(ctx); !errors.Is(e, ErrCapacity) {
		t.Fatal("live lease lost", e)
	}
	if _, e = w.Insert(ctx, "a", []float32{1, 0}, 0); e != nil {
		t.Fatal("failed close changed admission", e)
	}
	lease.Release()
	if e = w.Close(ctx); e != nil {
		t.Fatal(e)
	}
	if e = w.Close(ctx); e != nil {
		t.Fatal("non-idempotent", e)
	}
	if _, e = w.Acquire(ctx); !errors.Is(e, ErrFinished) {
		t.Fatal(e)
	}
	if _, e = w.Insert(ctx, "b", []float32{0, 1}, 0); !errors.Is(e, ErrFinished) {
		t.Fatal(e)
	}
	if _, e = w.Delete(ctx, "a"); !errors.Is(e, ErrFinished) {
		t.Fatal(e)
	}
	if w.current != nil || w.ids != nil || w.retired != nil || w.persist != nil {
		t.Fatal("owner retained closed state")
	}
}

func TestPrivateGraphClosePendingWrite(t *testing.T) {
	ctx := context.Background()
	entered, proceed := make(chan struct{}), make(chan struct{})
	w, e := NewPrivateGraphWriter(LayeredSnapshot{}, 0, 2, 4, 32, 2, 1, func(context.Context, uint64, []Mutation) error { close(entered); <-proceed; return nil })
	if e != nil {
		t.Fatal(e)
	}
	written := make(chan error, 1)
	go func() { _, e := w.Insert(ctx, "a", []float32{1, 0}, 0); written <- e }()
	<-entered
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if e = w.Close(canceled); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	closed := make(chan error, 1)
	go func() { closed <- w.Close(ctx) }()
	select {
	case e := <-closed:
		t.Fatal("close bypassed persistence", e)
	default:
	}
	close(proceed)
	if e = <-written; e != nil {
		t.Fatal(e)
	}
	if e = <-closed; e != nil {
		t.Fatal(e)
	}
	if _, e = w.Acquire(ctx); !errors.Is(e, ErrFinished) {
		t.Fatal(e)
	}
}
