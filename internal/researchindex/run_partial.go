package researchindex

import (
	"context"
	"errors"
)

// PartialRunMerge merges only the newest two graph runs, excluding the delta.
// Tombstones must survive because older unmerged graphs may still contain IDs.
type PartialRunMerge struct {
	flush *RunFlush
	runs  []*ImmutableRun
}

func (w *DurableRunWriter) PreparePartialRunMerge(ctx context.Context, dir string) (*PartialRunMerge, error) {
	return w.preparePartialRunMerge(ctx, dir, BuildImmutableRun)
}
func (w *DurableRunWriter) preparePartialRunMerge(ctx context.Context, dir string, build func(context.Context, []Mutation, int, string) (*ImmutableRun, error)) (*PartialRunMerge, error) {
	if err := w.lock(ctx); err != nil {
		return nil, err
	}
	if w.failed {
		w.gate <- struct{}{}
		return nil, ErrRecoveryRequired
	}
	if len(w.runs) < 3 {
		w.gate <- struct{}{}
		return nil, errors.New("partial merge needs two runs and older history")
	}
	if !w.core.compacting.CompareAndSwap(false, true) {
		w.gate <- struct{}{}
		return nil, ErrCompactionBusy
	}
	runs := append([]*ImmutableRun(nil), w.runs...)
	captured := w.core.current.Load()
	w.gate <- struct{}{}
	entered := false
	defer func() {
		if !entered {
			w.core.compacting.Store(false)
		}
	}()
	seen := map[string]bool{}
	var mutations []Mutation
	for _, r := range runs[:2] {
		for _, e := range r.manifest {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if seen[e.ID] {
				continue
			}
			seen[e.ID] = true
			m := Mutation{ID: e.ID, Delete: e.Deleted}
			if !e.Deleted {
				m.Vector = append([]float32(nil), r.view.g.base.records[e.ID].vector...)
			}
			mutations = append(mutations, m)
		}
	}
	entered = true
	built, err := build(ctx, mutations, w.dimension, dir)
	if err != nil {
		return nil, err
	}
	return &PartialRunMerge{flush: &RunFlush{w: w, captured: captured, built: built}, runs: runs}, nil
}
func (c *PartialRunMerge) Abort() error { return c.flush.Abort() }
func (c *PartialRunMerge) publish(ctx context.Context) (retired []*ImmutableRun, err error) {
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
		return nil, errors.New("partial run count changed")
	}
	for i, r := range c.runs {
		if w.runs[i] != r {
			return nil, errors.New("partial run identity changed")
		}
	}
	current := w.core.current.Load()
	var mutations []Mutation
	for id, r := range current.delta {
		mutations = append(mutations, Mutation{ID: id, Vector: r.vector, Delete: r.deleted})
	}
	runs := append([]*ImmutableRun{f.built}, c.runs[2:]...)
	search, err := NewRunDeltaSearch(ctx, runs, mutations, w.dimension, w.capacity, w.k)
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	retired = append([]*ImmutableRun(nil), c.runs[:2]...)
	// Delta was not part of this merge: preserve it in full, including entries
	// that predate capture. Only graph representation changes.
	w.runs = runs
	w.current = RunSnapshot{current.revision, search}
	published = true
	w.core.compacting.Store(false)
	return retired, nil
}
func (p *RunLeasePool) PublishPartialRunMerge(ctx context.Context, c *PartialRunMerge) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return ErrServingClosed
	}
	if c == nil {
		return errors.New("nil partial merge")
	}
	if p.maxRetired-len(p.retired) < 2 {
		return ErrCapacity
	}
	for _, r := range c.runs[:2] {
		if _, exists := p.retired[r]; exists {
			return errors.New("run already retired")
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	retired, err := c.publish(ctx)
	if err != nil {
		return err
	}
	for _, r := range retired {
		p.retired[r] = false
	}
	for _, r := range retired {
		if p.refs[r] == 0 {
			err = errors.Join(err, p.schedule(r))
		}
	}
	return err
}
