package researchindex

import (
	"context"
	"errors"
	"sync/atomic"
)

// RunFlush owns one unpublished graph and the sole build slot. Successful
// publication transfers the graph to the writer's retained run set; callers
// still own final graph retirement until the lease layer is implemented.
type RunFlush struct {
	w        *DurableRunWriter
	captured *generation
	built    *ImmutableRun
	finished atomic.Bool
}

func (w *DurableRunWriter) PrepareRunFlush(ctx context.Context, directory string) (*RunFlush, error) {
	return w.prepareRunFlush(ctx, directory, BuildImmutableRun)
}
func (w *DurableRunWriter) prepareRunFlush(ctx context.Context, directory string, build func(context.Context, []Mutation, int, string) (*ImmutableRun, error)) (*RunFlush, error) {
	if err := w.lock(ctx); err != nil {
		return nil, err
	}
	if w.failed {
		w.gate <- struct{}{}
		return nil, ErrRecoveryRequired
	}
	if len(w.runs) >= 15 {
		w.gate <- struct{}{}
		return nil, ErrCapacity
	}
	g := w.core.current.Load()
	if len(g.delta) == 0 {
		w.gate <- struct{}{}
		return nil, errors.New("empty flush")
	}
	if !w.core.compacting.CompareAndSwap(false, true) {
		w.gate <- struct{}{}
		return nil, ErrCompactionBusy
	}
	mutations := make([]Mutation, 0, len(g.delta))
	for id, r := range g.delta {
		mutations = append(mutations, Mutation{ID: id, Vector: append([]float32(nil), r.vector...), Delete: r.deleted})
	}
	w.gate <- struct{}{}
	built, err := build(ctx, mutations, w.dimension, directory)
	if err != nil {
		// The existing builder cannot report failed-close ownership. Keep the
		// build slot charged rather than claiming all resources were released.
		// A future lifecycle owner must reconcile this state before more builds.
		return nil, err
	}
	return &RunFlush{w: w, captured: g, built: built}, nil
}
func (f *RunFlush) discard() error {
	err := f.built.Close()
	if err == nil {
		f.w.core.compacting.Store(false)
	}
	return err
}
func (f *RunFlush) Abort() error {
	if !f.finished.CompareAndSwap(false, true) {
		return ErrFinished
	}
	return f.discard()
}
func (f *RunFlush) Publish(ctx context.Context) (err error) {
	if !f.finished.CompareAndSwap(false, true) {
		return ErrFinished
	}
	published := false
	defer func() {
		if !published {
			err = errors.Join(err, f.discard())
		}
	}()
	w := f.w
	if err = w.lock(ctx); err != nil {
		return err
	}
	defer func() { w.gate <- struct{}{} }()
	if w.failed {
		return ErrRecoveryRequired
	}
	current := w.core.current.Load()
	if current.base != f.captured.base {
		return errors.New("flush generation changed")
	}
	delta := make(map[string]record, len(current.delta))
	var mutations []Mutation
	for id, r := range current.delta {
		if r.revision > f.captured.revision {
			delta[id] = r
			mutations = append(mutations, Mutation{ID: id, Vector: r.vector, Delete: r.deleted})
		}
	}
	runs := append([]*ImmutableRun{f.built}, w.runs...)
	search, err := NewRunDeltaSearch(ctx, runs, mutations, w.dimension, w.capacity, w.k)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// No semantic write is occurring: the durable revision is unchanged.
	// Gate exclusion makes the graph list, delta and query plan one snapshot.
	w.core.current.Store(&generation{base: current.base, delta: delta, revision: current.revision})
	w.runs = runs
	w.current = RunSnapshot{current.revision, search}
	published = true
	w.core.compacting.Store(false)
	return nil
}
