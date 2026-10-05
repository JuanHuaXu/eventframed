package researchindex

import (
	"context"
	"errors"
)

type RunSnapshot struct {
	revision uint64
	search   *RunDeltaSearch
}

func (s RunSnapshot) Revision() uint64 { return s.revision }
func (s RunSnapshot) Search(ctx context.Context, q []float32) ([]Candidate, error) {
	return s.search.Search(ctx, q)
}

// DurableRunWriter coordinates a caller-owned coherent run set and a bounded
// pending delta. It does not authenticate recovery state or own graph leases.
// The caller must keep every supplied graph alive for all historical readers.
type DurableRunWriter struct {
	gate                   chan struct{}
	core                   *Generations
	runs                   []*ImmutableRun
	current                RunSnapshot
	persist                PersistGeneration
	failed                 bool
	dimension, capacity, k int
}

// NewDurableRunWriter requires an exclusively recovered, coherent run set and
// semantic revision. Pending durable changes must be included in those runs.
func NewDurableRunWriter(ctx context.Context, runs []*ImmutableRun, revision uint64, dimension, capacity, k int, persist PersistGeneration) (*DurableRunWriter, error) {
	if persist == nil {
		return nil, errors.New("missing persistence callback")
	}
	if revision == 0 && len(runs) > 0 {
		return nil, errors.New("run state with zero revision")
	}
	core, err := NewGenerations(dimension, capacity)
	if err != nil {
		return nil, err
	}
	search, err := NewRunDeltaSearch(ctx, runs, nil, dimension, capacity, k)
	if err != nil {
		return nil, err
	}
	core.current.Load().revision = revision // unpublished constructor state
	w := &DurableRunWriter{gate: make(chan struct{}, 1), core: core, runs: append([]*ImmutableRun(nil), runs...), current: RunSnapshot{revision, search}, persist: persist, dimension: dimension, capacity: capacity, k: k}
	w.gate <- struct{}{}
	return w, nil
}
func (w *DurableRunWriter) lock(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-w.gate:
	}
	if err := ctx.Err(); err != nil {
		w.gate <- struct{}{}
		return err
	}
	return nil
}
func (w *DurableRunWriter) Snapshot(ctx context.Context) (RunSnapshot, error) {
	if err := w.lock(ctx); err != nil {
		return RunSnapshot{}, err
	}
	defer func() { w.gate <- struct{}{} }()
	if w.failed {
		return RunSnapshot{}, ErrRecoveryRequired
	}
	return w.current, nil
}
func (w *DurableRunWriter) Apply(ctx context.Context, mutations []Mutation) error {
	if err := w.lock(ctx); err != nil {
		return err
	}
	defer func() { w.gate <- struct{}{} }()
	if w.failed {
		return ErrRecoveryRequired
	}
	p, err := w.core.Prepare(ctx, mutations)
	if err != nil {
		return err
	}
	settled, entered := false, false
	defer func() {
		if !settled {
			if entered {
				w.failed = true
			}
			_ = p.Abort()
		}
	}()
	all := make([]Mutation, 0, len(p.next.delta))
	for id, r := range p.next.delta {
		all = append(all, Mutation{ID: id, Vector: r.vector, Delete: r.deleted})
	}
	search, err := NewRunDeltaSearch(ctx, w.runs, all, w.dimension, w.capacity, w.k)
	if err != nil {
		return err
	}
	next := RunSnapshot{p.next.revision, search}
	owned := make([]Mutation, 0, len(mutations))
	for _, m := range mutations {
		r := p.next.delta[m.ID]
		owned = append(owned, Mutation{ID: m.ID, Vector: append([]float32(nil), r.vector...), Delete: r.deleted})
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	entered = true
	if err := w.persist(ctx, next.revision, owned); err != nil {
		return errors.Join(ErrRecoveryRequired, err)
	}
	// Cancellation after durable success cannot cancel visibility. The private
	// prepared state has one owner, so Commit only publishes and unlocks.
	if err := p.Commit(); err != nil {
		return errors.Join(ErrRecoveryRequired, err)
	}
	w.current = next
	settled = true
	return nil
}
