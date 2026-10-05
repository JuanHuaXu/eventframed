package researchindex

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestRunPublicationCapacityPrecedesVisibility(t *testing.T) {
	ctx := context.Background()
	a, _ := leaseFixture(t)
	b, _ := leaseFixture(t)
	w, err := NewDurableRunWriter(ctx, []*ImmutableRun{a, b}, 1, 2, 64, 1, func(context.Context, uint64, []Mutation) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	p, _ := NewRunLeasePool(2, 1)
	defer p.Close()
	c, err := w.PrepareRunConsolidation(ctx, filepath.Join(t.TempDir(), "candidate"))
	if err != nil {
		t.Fatal(err)
	}
	if err = p.PublishConsolidation(ctx, c); !errors.Is(err, ErrCapacity) {
		t.Fatal(err)
	}
	if c.flush.finished.Load() || len(w.runs) != 2 || w.runs[0] != a {
		t.Fatal("capacity checked after publication")
	}
	if err = c.Abort(); err != nil {
		t.Fatal(err)
	}
	l, err := p.AcquireCurrent(ctx, w)
	if err != nil {
		t.Fatal(err)
	}
	if err = l.Release(); err != nil {
		t.Fatal(err)
	}
}

func TestRunPublicationKeepsOldLeaseAndAdmitsNew(t *testing.T) {
	ctx := context.Background()
	r, _ := leaseFixture(t)
	w, err := NewDurableRunWriter(ctx, []*ImmutableRun{r}, 1, 2, 64, 1, func(context.Context, uint64, []Mutation) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	p, _ := NewRunLeasePool(2, 1)
	old, err := p.AcquireCurrent(ctx, w)
	if err != nil {
		t.Fatal(err)
	}
	if err = w.Apply(ctx, []Mutation{{ID: "a", Delete: true}, {ID: "b", Vector: []float32{1, 0}}}); err != nil {
		t.Fatal(err)
	}
	c, err := w.PrepareRunConsolidation(ctx, filepath.Join(t.TempDir(), "replacement"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.flush.built.Close()
	if err = p.PublishConsolidation(ctx, c); err != nil {
		t.Fatal(err)
	}
	fresh, err := p.AcquireCurrent(ctx, w)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		lease *RunLease
		want  string
	}{{old, "a"}, {fresh, "b"}} {
		got, e := test.lease.Search(ctx, []float32{1, 0})
		if e != nil || len(got) != 1 || got[0].ID != test.want {
			t.Fatal(got, e)
		}
	}
	if _, err = p.Acquire(old.snapshot); !errors.Is(err, ErrServingBusy) {
		t.Fatal(err)
	}
	if err = old.Release(); err != nil {
		t.Fatal(err)
	}
	if err = fresh.Release(); err != nil {
		t.Fatal(err)
	}
	if err = p.Close(); err != nil {
		t.Fatal(err)
	}
	if !r.index.closed || c.flush.built.index.closed {
		t.Fatal("wrong graph retired")
	}
}
