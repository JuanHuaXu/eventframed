package researchindex

import (
	"context"
	"errors"
	"math"
	"slices"
	"time"
)

type PrivateInsertRequest struct {
	ID     string
	Vector []float32
	Level  int
}

// InsertBatch atomically prepares/persists/publishes one to four insertions.
// Revision advances by the number of inserted events; intermediate revisions
// are not externally visible. The caller supplies a deadline valid for ALL
// members. This method does not collect requests or acknowledge queued work.
func (w *PrivateGraphWriter) InsertBatch(ctx context.Context, requests []PrivateInsertRequest) (PrivateWriteTiming, error) {
	var timing PrivateWriteTiming
	if len(requests) < 1 || len(requests) > 4 {
		return timing, ErrCapacity
	}
	start := time.Now()
	select {
	case <-ctx.Done():
		return timing, ctx.Err()
	case <-w.gate:
	}
	timing.Wait = time.Since(start)
	defer func() { w.gate <- struct{}{} }()
	if e := ctx.Err(); e != nil {
		return timing, e
	}
	w.mu.Lock()
	before := w.current
	closed, failed, full := w.closed, w.failed, len(w.retired) >= w.maxRetired
	w.mu.Unlock()
	if closed {
		return timing, ErrFinished
	}
	if failed {
		return timing, ErrRecoveryRequired
	}
	if full {
		return timing, ErrCapacity
	}
	size := uint64(len(requests))
	if before.revision > math.MaxUint64-size || w.next > uint64(math.MaxUint32)-size {
		return timing, ErrCapacity
	}
	ids := map[string]bool{}
	for _, r := range requests {
		if r.ID == "" || len(r.Vector) != w.dimension || ids[r.ID] {
			return timing, errors.New("invalid batch request")
		}
		if _, ok := w.ids[r.ID]; ok {
			return timing, errors.New("existing batch ID")
		}
		ids[r.ID] = true
	}
	next := &privateGraphVersion{graph: before.graph, summary: before.summary, revision: before.revision + size, count: before.count}
	changes := make([]Mutation, len(requests))
	start = time.Now()
	for i, r := range requests {
		ordinal := uint32(w.next + uint64(i))
		p, e := PreparePrivateInsertion(ctx, next.graph, next.summary, ordinal, r.ID, r.Vector, r.Level, next.count, w.m, w.ef, 100000, 100000, 128, 1)
		if e != nil {
			timing.Prepare = time.Since(start)
			return timing, e
		}
		next.graph, next.summary = p.Snapshot, p.Summary
		next.count++
		record := layeredFind(next.graph.tree, ordinal)
		changes[i] = Mutation{ID: r.ID, Vector: slices.Clone(record.Vector)}
	}
	timing.Prepare = time.Since(start)
	if e := ctx.Err(); e != nil {
		return timing, e
	}
	// These IDs remain private to the held writer gate. Any uncertain callback
	// quarantines the owner; no rolled-forward map can be used without recovery.
	for i, r := range requests {
		w.ids[r.ID] = uint32(w.next + uint64(i))
	}
	var e error
	timing.Persist, e = w.persistPrepared(ctx, before, next, changes)
	if e != nil {
		return timing, e
	}
	w.next += size
	return timing, nil
}
