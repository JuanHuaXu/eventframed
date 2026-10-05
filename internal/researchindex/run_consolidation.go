package researchindex

import (
	"context"
	"errors"
)

// RunConsolidation captures the entire run history plus delta. Tombstones may
// be removed only because every older version is included. Retired handles
// returned at publication must remain alive until all old readers are released.
type RunConsolidation struct {
	flush *RunFlush
	runs  []*ImmutableRun
}

func (w *DurableRunWriter) PrepareRunConsolidation(ctx context.Context, dir string) (*RunConsolidation, error) {
	return w.prepareRunConsolidation(ctx, dir, BuildImmutableRun)
}
func (w *DurableRunWriter) prepareRunConsolidation(ctx context.Context, dir string, build func(context.Context, []Mutation, int, string) (*ImmutableRun, error)) (*RunConsolidation, error) {
	if err := w.lock(ctx); err != nil {
		return nil, err
	}
	if w.failed {
		w.gate <- struct{}{}
		return nil, ErrRecoveryRequired
	}
	if !w.core.compacting.CompareAndSwap(false, true) {
		w.gate <- struct{}{}
		return nil, ErrCompactionBusy
	}
	g := w.core.current.Load()
	runs := append([]*ImmutableRun(nil), w.runs...)
	w.gate <- struct{}{}
	entered := false
	defer func() {
		if !entered {
			w.core.compacting.Store(false)
		}
	}()
	seen := map[string]bool{}
	var mutations []Mutation
	for id, r := range g.delta {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		seen[id] = true
		if !r.deleted {
			mutations = append(mutations, Mutation{ID: id, Vector: append([]float32(nil), r.vector...)})
		}
	}
	for _, run := range runs {
		for _, entry := range run.manifest {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if seen[entry.ID] {
				continue
			}
			seen[entry.ID] = true
			if !entry.Deleted {
				r := run.view.g.base.records[entry.ID]
				mutations = append(mutations, Mutation{ID: entry.ID, Vector: append([]float32(nil), r.vector...)})
			}
		}
	}
	entered = true
	built, err := build(ctx, mutations, w.dimension, dir)
	if err != nil {
		return nil, err
	}
	return &RunConsolidation{flush: &RunFlush{w: w, captured: g, built: built}, runs: runs}, nil
}
func (c *RunConsolidation) Abort() error { return c.flush.Abort() }
func (c *RunConsolidation) Publish(ctx context.Context) (retired []*ImmutableRun, err error) {
	f := c.flush
	if !f.finished.CompareAndSwap(false, true) {
		return nil, ErrFinished
	}
	published := false
	defer func() {
		if !published {
			err = errors.Join(err, f.discard())
		}
	}()
	w := f.w
	if err = w.lock(ctx); err != nil {
		return nil, err
	}
	defer func() { w.gate <- struct{}{} }()
	if w.failed {
		return nil, ErrRecoveryRequired
	}
	if len(w.runs) != len(c.runs) {
		return nil, errors.New("consolidation run set changed")
	}
	for i, r := range c.runs {
		if w.runs[i] != r {
			return nil, errors.New("consolidation run identity changed")
		}
	}
	current := w.core.current.Load()
	if current.base != f.captured.base {
		return nil, errors.New("consolidation generation changed")
	}
	delta := map[string]record{}
	var mutations []Mutation
	for id, r := range current.delta {
		if r.revision > f.captured.revision {
			delta[id] = r
			mutations = append(mutations, Mutation{ID: id, Vector: r.vector, Delete: r.deleted})
		}
	}
	runs := []*ImmutableRun{f.built}
	search, err := NewRunDeltaSearch(ctx, runs, mutations, w.dimension, w.capacity, w.k)
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	retired = append([]*ImmutableRun(nil), c.runs...)
	w.core.current.Store(&generation{base: current.base, delta: delta, revision: current.revision})
	w.runs = runs
	w.current = RunSnapshot{current.revision, search}
	published = true
	w.core.compacting.Store(false)
	return retired, nil
}
