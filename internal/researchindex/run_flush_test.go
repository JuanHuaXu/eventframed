package researchindex

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRunFlushConcurrentOverwriteAndDelete(t *testing.T) {
	ctx := context.Background()
	w, err := NewDurableRunWriter(ctx, nil, 0, 2, 64, 3, func(context.Context, uint64, []Mutation) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if err = w.Apply(ctx, []Mutation{{ID: "a", Vector: []float32{1, 0}}, {ID: "b", Vector: []float32{0, 1}}}); err != nil {
		t.Fatal(err)
	}
	old, _ := w.Snapshot(ctx)
	entered, release := make(chan struct{}), make(chan struct{})
	done := make(chan *RunFlush, 1)
	errs := make(chan error, 1)
	dir := filepath.Join(t.TempDir(), "flush")
	go func() {
		f, e := w.prepareRunFlush(ctx, dir, func(c context.Context, m []Mutation, d int, p string) (*ImmutableRun, error) {
			close(entered)
			<-release
			return BuildImmutableRun(c, m, d, p)
		})
		done <- f
		errs <- e
	}()
	<-entered
	if _, err = w.PrepareRunFlush(ctx, dir+"-other"); !errors.Is(err, ErrCompactionBusy) {
		t.Fatal(err)
	}
	if err = w.Apply(ctx, []Mutation{{ID: "a", Delete: true}, {ID: "b", Vector: []float32{1, 0}}, {ID: "c", Vector: []float32{0, 1}}}); err != nil {
		t.Fatal(err)
	}
	before, _ := w.Snapshot(ctx)
	close(release)
	f := <-done
	if err = <-errs; err != nil {
		t.Fatal(err)
	}
	defer f.built.Close()
	if err = f.Publish(ctx); err != nil {
		t.Fatal(err)
	}
	after, _ := w.Snapshot(ctx)
	x, e := before.Search(ctx, []float32{1, 0})
	y, g := after.Search(ctx, []float32{1, 0})
	if e != nil || g != nil || !reflect.DeepEqual(x, y) || after.Revision() != 2 {
		t.Fatal(x, y, e, g)
	}
	if len(w.core.current.Load().delta) != 3 {
		t.Fatal("newer versions drained")
	}
	prior, e := old.Search(ctx, []float32{1, 0})
	if e != nil || len(prior) != 2 || prior[0].ID != "a" {
		t.Fatal("old view changed", prior, e)
	}
	if err = f.Publish(ctx); !errors.Is(err, ErrFinished) {
		t.Fatal(err)
	}
	second, err := w.PrepareRunFlush(ctx, dir+"-second")
	if err != nil {
		t.Fatal(err)
	}
	defer second.built.Close()
	if err = second.Publish(ctx); err != nil {
		t.Fatal(err)
	}
	if len(w.core.current.Load().delta) != 0 {
		t.Fatal("captured delta not drained")
	}
	final, _ := w.Snapshot(ctx)
	z, e := final.Search(ctx, []float32{1, 0})
	if e != nil || !reflect.DeepEqual(z, y) || final.Revision() != 2 {
		t.Fatal(z, e)
	}
}

func TestRunFlushCancellationAndQuarantine(t *testing.T) {
	ctx := context.Background()
	fail := false
	w, err := NewDurableRunWriter(ctx, nil, 0, 2, 64, 1, func(context.Context, uint64, []Mutation) error {
		if fail {
			return errors.New("lost ack")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = w.Apply(ctx, []Mutation{{ID: "a", Vector: []float32{1, 0}}}); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	f, err := w.PrepareRunFlush(ctx, filepath.Join(root, "cancel"))
	if err != nil {
		t.Fatal(err)
	}
	c, cancel := context.WithCancel(ctx)
	cancel()
	if err = f.Publish(c); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if w.core.compacting.Load() || !f.built.index.closed || len(w.core.current.Load().delta) != 1 {
		t.Fatal("cancel cleanup")
	}
	f, err = w.PrepareRunFlush(ctx, filepath.Join(root, "quarantine"))
	if err != nil {
		t.Fatal(err)
	}
	fail = true
	if err = w.Apply(ctx, []Mutation{{ID: "b", Delete: true}}); !errors.Is(err, ErrRecoveryRequired) {
		t.Fatal(err)
	}
	if err = f.Publish(ctx); !errors.Is(err, ErrRecoveryRequired) {
		t.Fatal(err)
	}
	if !f.built.index.closed {
		t.Fatal("quarantined candidate leaked")
	}
}

func TestRunFlushFailedBuildRemainsCharged(t *testing.T) {
	ctx := context.Background()
	w, _ := NewDurableRunWriter(ctx, nil, 0, 2, 64, 1, func(context.Context, uint64, []Mutation) error { return nil })
	if err := w.Apply(ctx, []Mutation{{ID: "a", Delete: true}}); err != nil {
		t.Fatal(err)
	}
	_, err := w.prepareRunFlush(ctx, "unused", func(context.Context, []Mutation, int, string) (*ImmutableRun, error) {
		return nil, errors.New("unknown build cleanup")
	})
	if err == nil || !w.core.compacting.Load() {
		t.Fatal("unknown resources uncharged")
	}
	if _, err = w.PrepareRunFlush(ctx, "unused"); !errors.Is(err, ErrCompactionBusy) {
		t.Fatal(err)
	}
}
