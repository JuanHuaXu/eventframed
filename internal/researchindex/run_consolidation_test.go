package researchindex

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRunConsolidationConcurrentHistory(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	w, _ := NewDurableRunWriter(ctx, nil, 0, 2, 64, 4, func(context.Context, uint64, []Mutation) error { return nil })
	apply := func(ms []Mutation) {
		t.Helper()
		if err := w.Apply(ctx, ms); err != nil {
			t.Fatal(err)
		}
	}
	apply([]Mutation{{ID: "a", Vector: []float32{1, 0}}, {ID: "b", Vector: []float32{0, 1}}, {ID: "d", Vector: []float32{1, 1}}})
	f, err := w.PrepareRunFlush(ctx, filepath.Join(root, "first"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.built.Close()
	if err = f.Publish(ctx); err != nil {
		t.Fatal(err)
	}
	apply([]Mutation{{ID: "a", Delete: true}, {ID: "b", Vector: []float32{1, 0}}})
	f2, err := w.PrepareRunFlush(ctx, filepath.Join(root, "second"))
	if err != nil {
		t.Fatal(err)
	}
	defer f2.built.Close()
	if err = f2.Publish(ctx); err != nil {
		t.Fatal(err)
	}
	old, _ := w.Snapshot(ctx)
	q := []float32{1, 0}
	oldWant, _ := old.Search(ctx, q)
	entered, release := make(chan struct{}), make(chan struct{})
	done := make(chan *RunConsolidation, 1)
	errs := make(chan error, 1)
	go func() {
		c, e := w.prepareRunConsolidation(ctx, filepath.Join(root, "merged"), func(ctx context.Context, m []Mutation, d int, p string) (*ImmutableRun, error) {
			close(entered)
			<-release
			return BuildImmutableRun(ctx, m, d, p)
		})
		done <- c
		errs <- e
	}()
	<-entered
	if _, err = w.PrepareRunConsolidation(ctx, "unused"); !errors.Is(err, ErrCompactionBusy) {
		t.Fatal(err)
	}
	apply([]Mutation{{ID: "a", Vector: []float32{-1, 0}}, {ID: "b", Delete: true}, {ID: "c", Vector: []float32{1, 0}}})
	before, _ := w.Snapshot(ctx)
	want, _ := before.Search(ctx, q)
	close(release)
	c := <-done
	if err = <-errs; err != nil {
		t.Fatal(err)
	}
	defer c.flush.built.Close()
	retired, err := c.Publish(ctx)
	if err != nil || len(retired) != 2 || len(w.runs) != 1 {
		t.Fatal(retired, err)
	}
	after, _ := w.Snapshot(ctx)
	got, e := after.Search(ctx, q)
	if e != nil || !reflect.DeepEqual(got, want) || after.Revision() != 3 {
		t.Fatal(got, want, e)
	}
	if len(w.core.current.Load().delta) != 3 {
		t.Fatal("new changes removed")
	}
	if _, exists := c.flush.built.view.g.base.records["a"]; exists {
		t.Fatal("captured tombstone resurrected")
	}
	prior, e := old.Search(ctx, q)
	if e != nil || !reflect.DeepEqual(prior, oldWant) {
		t.Fatal("retired graphs closed early", prior, e)
	}
	if _, err = c.Publish(ctx); !errors.Is(err, ErrFinished) {
		t.Fatal(err)
	}
}

func TestRunConsolidationCancelledPublication(t *testing.T) {
	ctx := context.Background()
	w, _ := NewDurableRunWriter(ctx, nil, 0, 2, 64, 1, func(context.Context, uint64, []Mutation) error { return nil })
	if err := w.Apply(ctx, []Mutation{{ID: "a", Vector: []float32{1, 0}}}); err != nil {
		t.Fatal(err)
	}
	c, err := w.PrepareRunConsolidation(ctx, filepath.Join(t.TempDir(), "candidate"))
	if err != nil {
		t.Fatal(err)
	}
	stop, cancel := context.WithCancel(ctx)
	cancel()
	retired, err := c.Publish(stop)
	if !errors.Is(err, context.Canceled) || retired != nil || w.core.compacting.Load() || !c.flush.built.index.closed {
		t.Fatal(retired, err)
	}
	s, e := w.Snapshot(ctx)
	got, e2 := s.Search(ctx, []float32{1, 0})
	if e != nil || e2 != nil || len(got) != 1 || s.Revision() != 1 {
		t.Fatal(got, e, e2)
	}
}
