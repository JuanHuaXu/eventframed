package researchindex

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestLayeredOwnerBounds(t *testing.T) {
	ctx := context.Background()
	lim := LayeredLimits{2, 2, 1, 8}
	r := LayeredRecord{ID: "old", Vector: []float32{1, 2}, Links: [][]uint32{{}}, Backlinks: [][]uint32{{}}, Heuristic: []uint32{0}}
	initial, _, e := PrepareLayered(ctx, LayeredSnapshot{}, []LayeredEdit{{0, &r}}, 0, lim)
	if e != nil {
		t.Fatal(e)
	}
	o, e := NewLayeredOwner(initial, lim, 2, 1)
	if e != nil {
		t.Fatal(e)
	}
	a, _ := o.Acquire()
	b, _ := o.Acquire()
	if _, e = o.Acquire(); !errors.Is(e, ErrServingBusy) {
		t.Fatal(e)
	}
	r.ID = "new"
	p, e := o.Prepare(ctx, []LayeredEdit{{0, &r}}, 1)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = o.Prepare(ctx, nil, 2); !errors.Is(e, ErrServingBusy) {
		t.Fatal(e)
	}
	if e = o.Close(); !errors.Is(e, ErrServingBusy) {
		t.Fatal(e)
	}
	if e = p.Commit(); e != nil {
		t.Fatal(e)
	}
	if e = p.Abort(); !errors.Is(e, ErrFinished) {
		t.Fatal(e)
	}
	if p.snapshot.tree != nil {
		t.Fatal("finished candidate retains root")
	}
	if _, e = o.Prepare(ctx, nil, 2); !errors.Is(e, ErrServingBusy) {
		t.Fatal("retirement cap", e)
	}
	got, _, _ := a.Lookup(0)
	if got.ID != "old" {
		t.Fatal("history")
	}
	a.Release()
	c, _ := o.Acquire()
	got, _, _ = c.Lookup(0)
	if got.ID != "new" {
		t.Fatal("publication")
	}
	if _, e = o.Prepare(ctx, nil, 2); !errors.Is(e, ErrServingBusy) {
		t.Fatal("second old reader still holds slot", e)
	}
	b.Release()
	p, e = o.Prepare(ctx, nil, 2)
	if e != nil {
		t.Fatal(e)
	}
	if e = p.Abort(); e != nil {
		t.Fatal(e)
	}
	if _, _, e = b.Lookup(0); !errors.Is(e, ErrFinished) {
		t.Fatal(e)
	}
	c.Release()
	if e = o.Close(); e != nil {
		t.Fatal(e)
	}
	if _, e = o.Acquire(); !errors.Is(e, ErrFinished) {
		t.Fatal(e)
	}
}
func TestLayeredOwnerCancellationAndConcurrentRelease(t *testing.T) {
	ctx := context.Background()
	o, _ := NewLayeredOwner(LayeredSnapshot{}, LayeredLimits{2, 2, 1, 8}, 8, 2)
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, e := o.Prepare(canceled, nil, 0); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	p, e := o.Prepare(ctx, nil, 0)
	if e != nil {
		t.Fatal("slot leaked", e)
	}
	l, _ := o.Acquire()
	if e = p.Commit(); e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_, _, e := l.Lookup(0)
			if e != nil && !errors.Is(e, ErrFinished) {
				t.Error(e)
			}
		}
	}()
	go func() {
		defer wg.Done()
		if e := l.Release(); e != nil {
			t.Error(e)
		}
	}()
	wg.Wait()
	if len(o.retired) != 0 {
		t.Fatal("retired root leaked")
	}
	if e = o.Close(); e != nil {
		t.Fatal(e)
	}
}
