package researchindex

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestGraphVersionsOldTraversalAndAbort(t *testing.T) {
	ctx := context.Background()
	g, _ := NewGraphVersions(2)
	put := func(es []GraphEdit) {
		t.Helper()
		p, e := g.Prepare(ctx, es)
		if e != nil {
			t.Fatal(e)
		}
		if e = p.Commit(); e != nil {
			t.Fatal(e)
		}
	}
	put([]GraphEdit{{ID: 1, Vector: []float32{1, 0}, Edges: []uint32{2}}, {ID: 2, Vector: []float32{0, 1}}})
	old := g.View()
	v := []float32{1, 1}
	edges := []uint32{3}
	p, e := g.Prepare(ctx, []GraphEdit{{ID: 1, Vector: v, Edges: edges}, {ID: 3, Vector: []float32{-1, 0}}, {ID: 2, Delete: true}})
	if e != nil {
		t.Fatal(e)
	}
	v[0] = 99
	edges[0] = 99
	if got, _ := g.View().Walk(1, 10); !reflect.DeepEqual(got, []uint32{1, 2}) {
		t.Fatal("prepared state visible", got)
	}
	if e = p.Commit(); e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		view GraphView
		want []uint32
	}{{old, []uint32{1, 2}}, {g.View(), []uint32{1, 3}}} {
		got, e := tc.view.Walk(1, 10)
		if e != nil || !reflect.DeepEqual(got, tc.want) {
			t.Fatal(got, e)
		}
	}
	vec, links, ok := g.View().Lookup(1)
	if !ok || vec[0] != 1 || links[0] != 3 {
		t.Fatal(vec, links)
	}
	vec[0] = 42
	links[0] = 42
	vec, _, _ = g.View().Lookup(1)
	if vec[0] != 1 {
		t.Fatal("lookup aliases root")
	}
	p, e = g.Prepare(ctx, []GraphEdit{{ID: 1, Delete: true}})
	if e != nil {
		t.Fatal(e)
	}
	if e = p.Abort(); e != nil {
		t.Fatal(e)
	}
	if _, _, ok = g.View().Lookup(1); !ok {
		t.Fatal("aborted deletion visible")
	}
	if e = p.Commit(); !errors.Is(e, ErrFinished) {
		t.Fatal(e)
	}
}

func TestGraphVersionsSharingAndCancellation(t *testing.T) {
	ctx := context.Background()
	g, _ := NewGraphVersions(2)
	p, e := g.Prepare(ctx, []GraphEdit{{ID: 0, Vector: []float32{1, 0}}, {ID: 1 << 31, Vector: []float32{0, 1}}})
	if e != nil {
		t.Fatal(e)
	}
	p.Commit()
	old := g.View()
	p, e = g.Prepare(ctx, []GraphEdit{{ID: 0, Vector: []float32{1, 1}}})
	if e != nil {
		t.Fatal(e)
	}
	if p.next.tree.child[1] != old.root.tree.child[1] {
		t.Fatal("untouched half copied")
	}
	c, cancel := context.WithCancel(ctx)
	cancel()
	if _, e = g.Prepare(c, []GraphEdit{{ID: 8, Delete: true}}); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if e = p.Commit(); e != nil {
		t.Fatal(e)
	}
	if old.Revision() != 1 || g.View().Revision() != 2 {
		t.Fatal("revision")
	}
	if _, e = g.Prepare(ctx, []GraphEdit{{ID: 0, Vector: []float32{0, 0}}}); e == nil {
		t.Fatal("zero accepted")
	}
	p, e = g.Prepare(ctx, []GraphEdit{{ID: 9, Delete: true}})
	if e != nil {
		t.Fatal("failed prepare retained writer", e)
	}
	p.Abort()
}

func TestGraphVersionsConcurrentHistoricalWalk(t *testing.T) {
	ctx := context.Background()
	g, _ := NewGraphVersions(2)
	p, e := g.Prepare(ctx, []GraphEdit{{ID: 1, Vector: []float32{1, 0}, Edges: []uint32{2}}, {ID: 2, Vector: []float32{0, 1}}})
	if e != nil {
		t.Fatal(e)
	}
	p.Commit()
	old := g.View()
	done := make(chan error, 1)
	go func() {
		for i := 0; i < 100; i++ {
			p, e := g.Prepare(ctx, []GraphEdit{{ID: 1, Vector: []float32{1, 0}, Edges: []uint32{3}}, {ID: 3, Vector: []float32{float32(i + 1), 1}}, {ID: 2, Delete: true}})
			if e != nil {
				done <- e
				return
			}
			if e = p.Commit(); e != nil {
				done <- e
				return
			}
		}
		done <- nil
	}()
	for i := 0; i < 1000; i++ {
		got, e := old.Walk(1, 10)
		if e != nil || !reflect.DeepEqual(got, []uint32{1, 2}) {
			t.Error("historical traversal changed", got, e)
		}
	}
	if e = <-done; e != nil {
		t.Fatal(e)
	}
}
