package researchindex

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPartialRunMergeRetainsTombstonesAndWholeDelta(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	var runs []*ImmutableRun
	for i, ms := range [][]Mutation{
		{{ID: "a", Delete: true}, {ID: "b", Vector: []float32{0, 1}}},
		{{ID: "a", Vector: []float32{1, 0}}, {ID: "b", Vector: []float32{1, 0}}},
		{{ID: "a", Vector: []float32{1, 0}}, {ID: "c", Vector: []float32{1, 1}}},
	} {
		r, err := BuildImmutableRun(ctx, ms, 2, filepath.Join(root, string(rune('a'+i))))
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()
		runs = append(runs, r)
	}
	w, err := NewDurableRunWriter(ctx, runs, 1, 2, 64, 3, func(context.Context, uint64, []Mutation) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if err = w.Apply(ctx, []Mutation{{ID: "x", Vector: []float32{-1, 0}}}); err != nil {
		t.Fatal(err)
	}
	p, _ := NewRunLeasePool(2, 2)
	old, err := p.AcquireCurrent(ctx, w)
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	done := make(chan *PartialRunMerge, 1)
	errs := make(chan error, 1)
	go func() {
		c, e := w.preparePartialRunMerge(ctx, filepath.Join(root, "merged"), func(ctx context.Context, m []Mutation, d int, s string) (*ImmutableRun, error) {
			close(entered)
			<-release
			return BuildImmutableRun(ctx, m, d, s)
		})
		done <- c
		errs <- e
	}()
	<-entered
	if err = w.Apply(ctx, []Mutation{{ID: "b", Delete: true}}); err != nil {
		t.Fatal(err)
	}
	before, _ := w.Snapshot(ctx)
	want, e := before.Search(ctx, []float32{1, 0})
	if e != nil {
		t.Fatal(e)
	}
	close(release)
	c := <-done
	if err = <-errs; err != nil {
		t.Fatal(err)
	}
	defer c.flush.built.Close()
	if err = p.PublishPartialRunMerge(ctx, c); err != nil {
		t.Fatal(err)
	}
	after, _ := w.Snapshot(ctx)
	got, e := after.Search(ctx, []float32{1, 0})
	if e != nil || !reflect.DeepEqual(got, want) || after.Revision() != 3 || len(w.core.current.Load().delta) != 2 {
		t.Fatal(got, want, e)
	}
	if len(w.runs) != 2 || w.runs[1] != runs[2] {
		t.Fatal("older graph replaced")
	}
	found := false
	for _, entry := range c.flush.built.manifest {
		if entry.ID == "a" && entry.Deleted {
			found = true
		}
	}
	if !found {
		t.Fatal("partial merge dropped tombstone")
	}
	if err = old.Release(); err != nil {
		t.Fatal(err)
	}
	if err = p.Close(); err != nil {
		t.Fatal(err)
	}
	if runs[2].index.closed {
		t.Fatal("unmerged history retired")
	}
}
