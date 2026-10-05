package researchindex

import (
	"context"
	"errors"
	"fmt"
	"math"
)

type deleteBundle struct {
	graph    LayeredSnapshot
	summary  EntrySummary
	revision uint64
}
type PrivateDeleteView struct{ bundle *deleteBundle }

func (v PrivateDeleteView) Revision() uint64                       { return v.bundle.revision }
func (v PrivateDeleteView) Lookup(id uint32) (LayeredRecord, bool) { return v.bundle.graph.Lookup(id) }

// PrivateDeleteWriter is a deletion-only research coordinator. Persistence covers
// authoritative ID deletion and revision, not exact graph topology. Recovery must
// rebuild or restore a coherent graph/summary from authoritative state. This owner
// does not enumerate recovery data, enforce reader retention, or support inserts.
type PrivateDeleteWriter struct {
	gate    chan struct{}
	current *deleteBundle
	failed  bool
	persist PersistGeneration
	m       int
}

func NewPrivateDeleteWriter(graph LayeredSnapshot, summary EntrySummary, revision uint64, m int, persist PersistGeneration) (*PrivateDeleteWriter, error) {
	if persist == nil || m < 1 || m > 128 || graph.tree != nil && revision == 0 {
		return nil, errors.New("invalid private deletion writer")
	}
	w := &PrivateDeleteWriter{gate: make(chan struct{}, 1), current: &deleteBundle{graph, summary, revision}, persist: persist, m: m}
	w.gate <- struct{}{}
	return w, nil
}
func (w *PrivateDeleteWriter) lock(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-w.gate:
	}
	if e := ctx.Err(); e != nil {
		w.gate <- struct{}{}
		return e
	}
	return nil
}
func (w *PrivateDeleteWriter) View(ctx context.Context) (PrivateDeleteView, error) {
	if e := w.lock(ctx); e != nil {
		return PrivateDeleteView{}, e
	}
	defer func() { w.gate <- struct{}{} }()
	if w.failed {
		return PrivateDeleteView{}, ErrRecoveryRequired
	}
	return PrivateDeleteView{w.current}, nil
}
func (w *PrivateDeleteWriter) Delete(ctx context.Context, ordinal uint32) error {
	if e := w.lock(ctx); e != nil {
		return e
	}
	defer func() { w.gate <- struct{}{} }()
	if w.failed {
		return ErrRecoveryRequired
	}
	before := w.current
	if before.revision == math.MaxUint64 {
		return errors.New("private deletion revision overflow")
	}
	r := layeredFind(before.graph.tree, ordinal)
	if r == nil || r.ID == "" {
		return errors.New("private deletion ID absent")
	}
	p, e := PreparePrivateDeletion(ctx, before.graph, before.summary, ordinal, w.m, 4096, 256, 128, 65536)
	if e != nil {
		return e
	}
	// Finish allocations before entering the persistence callback. Caller mutation
	// of the callback's owned mutation slice cannot affect the prepared bundle.
	next := &deleteBundle{p.Snapshot, p.Summary, before.revision + 1}
	mutations := []Mutation{{ID: r.ID, Delete: true}}
	if e = ctx.Err(); e != nil {
		return e
	}
	e = func() (err error) {
		defer func() {
			if x := recover(); x != nil {
				err = fmt.Errorf("private deletion persistence panic: %v", x)
			}
		}()
		return w.persist(ctx, next.revision, mutations)
	}()
	if e != nil {
		w.failed = true
		return e
	}
	// Nil means durable success. A late context cancellation cannot undo it.
	w.current = next
	return nil
}
