package researchindex

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"sync"
	"time"
)

type privateGraphVersion struct {
	graph       LayeredSnapshot
	summary     EntrySummary
	revision    uint64
	count, refs int
}
type PrivateWriteTiming struct{ Wait, Prepare, Persist time.Duration }

// PrivateGraphWriter is research-only. Its callback atomically stores vectors/
// IDs and revision, not graph topology. Recovery reconstructs a graph before
// calling this constructor. A callback error/panic quarantines new operations.
// Version and lease counts are bounded; retained bytes are NOT yet bounded.
type PrivateGraphWriter struct {
	gate                            chan struct{}
	mu                              sync.Mutex
	current                         *privateGraphVersion
	retired                         []*privateGraphVersion
	maxReaders, maxRetired, readers int
	failed                          bool
	closed                          bool
	ids                             map[string]uint32 // writer-gate owned, never accessed by readers
	next                            uint64
	dimension, m, ef                int
	persist                         PersistGeneration
}
type PrivateGraphLease struct {
	mu       sync.RWMutex
	owner    *PrivateGraphWriter
	version  *privateGraphVersion
	released bool
}

func NewPrivateGraphWriter(graph LayeredSnapshot, revision uint64, dimension, m, ef, maxReaders, maxRetired int, persist PersistGeneration) (*PrivateGraphWriter, error) {
	if dimension < 1 || dimension > 4096 || m < 1 || m > 56 || ef < 1 || ef > 4096 || maxReaders < 1 || maxReaders > 128 || maxRetired < 1 || maxRetired > 128 || persist == nil || graph.tree != nil && revision == 0 {
		return nil, ErrCapacity
	}
	w := &PrivateGraphWriter{gate: make(chan struct{}, 1), current: &privateGraphVersion{graph: graph, revision: revision}, retired: make([]*privateGraphVersion, 0, maxRetired), ids: map[string]uint32{}, dimension: dimension, m: m, ef: ef, maxReaders: maxReaders, maxRetired: maxRetired, persist: persist}
	// Startup-only scan reconstructs ID ownership/count/summary; no corpus scan
	// is performed during mutation preparation or publication.
	var walk func(*layeredTrie, int, uint32) error
	walk = func(n *layeredTrie, bit int, id uint32) error {
		if n == nil {
			return nil
		}
		if bit >= 0 {
			if e := walk(n.child[0], bit-1, id); e != nil {
				return e
			}
			return walk(n.child[1], bit-1, id|uint32(1)<<bit)
		}
		r := n.record
		if r == nil || r.ID == "" || len(r.Vector) != dimension {
			return errors.New("invalid recovery graph record")
		}
		if _, ok := w.ids[r.ID]; ok {
			return errors.New("duplicate recovery ID")
		}
		w.ids[r.ID] = id
		w.next = max(w.next, uint64(id)+1)
		w.current.count++
		var e error
		w.current.summary, e = w.current.summary.With(id, r.Level)
		return e
	}
	if e := walk(graph.tree, 31, 0); e != nil {
		return nil, e
	}
	if graph.tree != nil {
		entry := layeredFind(graph.tree, uint32(graph.global>>32))
		if graph.global == 0 || entry == nil || uint32(graph.global) != uint32(entry.Level+1) {
			return nil, errors.New("invalid recovery entry")
		}
	} else if graph.global != 0 {
		return nil, errors.New("nonempty entry for empty graph")
	}
	w.gate <- struct{}{}
	return w, nil
}
func (w *PrivateGraphWriter) Acquire(ctx context.Context) (*PrivateGraphLease, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if w.closed {
		return nil, ErrFinished
	}
	if w.failed {
		return nil, ErrRecoveryRequired
	}
	if w.readers >= w.maxReaders {
		return nil, ErrCapacity
	}
	w.readers++
	w.current.refs++
	return &PrivateGraphLease{owner: w, version: w.current}, nil
}
func (l *PrivateGraphLease) Release() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.released {
		return
	}
	l.released = true
	w := l.owner
	w.mu.Lock()
	defer w.mu.Unlock()
	v := l.version
	v.refs--
	w.readers--
	if v.refs == 0 {
		for i, r := range w.retired {
			if r == v {
				copy(w.retired[i:], w.retired[i+1:])
				w.retired[len(w.retired)-1] = nil
				w.retired = w.retired[:len(w.retired)-1]
				break
			}
		}
	}
	l.version = nil
	l.owner = nil
}
func (l *PrivateGraphLease) Revision() (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.released {
		return 0, ErrFinished
	}
	return l.version.revision, nil
}
func (l *PrivateGraphLease) Lookup(id uint32) (LayeredRecord, bool, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.released {
		return LayeredRecord{}, false, ErrFinished
	}
	r, ok := l.version.graph.Lookup(id)
	return r, ok, nil
}
func (l *PrivateGraphLease) Search(ctx context.Context, q []float32, k, ef, budget int) ([]Candidate, int, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.released {
		return nil, 0, ErrFinished
	}
	return SearchLayered(ctx, l.version.graph, q, k, ef, budget)
}

func (w *PrivateGraphWriter) write(ctx context.Context, id string, vector []float32, level int, remove bool) (PrivateWriteTiming, error) {
	var timing PrivateWriteTiming
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
	failed, full, closed := w.failed, len(w.retired) >= w.maxRetired, w.closed
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
	if before.revision == math.MaxUint64 {
		return timing, errors.New("revision exhausted")
	}
	ordinal, exists := w.ids[id]
	if id == "" || remove && !exists || !remove && exists {
		return timing, errors.New("invalid mutation ID")
	}
	if !remove {
		if w.next >= math.MaxUint32 || len(vector) != w.dimension {
			return timing, ErrCapacity
		}
		ordinal = uint32(w.next)
	}
	next := &privateGraphVersion{revision: before.revision + 1, count: before.count}
	start = time.Now()
	if remove {
		p, e := PreparePrivateDeletion(ctx, before.graph, before.summary, ordinal, w.m, 4096, 256, 128, 100000)
		timing.Prepare = time.Since(start)
		if e != nil {
			return timing, e
		}
		next.graph, next.summary = p.Snapshot, p.Summary
		next.count--
	} else {
		p, e := PreparePrivateInsertion(ctx, before.graph, before.summary, ordinal, id, vector, level, before.count, w.m, w.ef, 100000, 100000, 128, 1)
		timing.Prepare = time.Since(start)
		if e != nil {
			return timing, e
		}
		next.graph, next.summary = p.Snapshot, p.Summary
		next.count++
	}
	// Callback data is independently owned: a callback may mutate it without
	// changing the graph that will be published. All allocations precede commit.
	changes := []Mutation{{ID: id, Delete: remove}}
	if !remove {
		r := layeredFind(next.graph.tree, ordinal)
		changes[0].Vector = slices.Clone(r.Vector)
	}
	if e := ctx.Err(); e != nil {
		return timing, e
	}
	if !remove {
		w.ids[id] = ordinal
	} // reserve map allocation before persistence
	var e error
	timing.Persist, e = w.persistPrepared(ctx, before, next, changes)
	if e != nil {
		return timing, e
	}
	if remove {
		delete(w.ids, id)
	} else {
		w.next++
	}
	return timing, nil
}

// Caller owns the writer gate, reserved retirement capacity and map storage,
// and checked cancellation before committing. Nil is final durable success.
func (w *PrivateGraphWriter) persistPrepared(ctx context.Context, before, next *privateGraphVersion, changes []Mutation) (time.Duration, error) {
	start := time.Now()
	e := func() (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("graph persistence panic: %v", p)
			}
		}()
		return w.persist(ctx, next.revision, changes)
	}()
	elapsed := time.Since(start)
	w.mu.Lock()
	defer w.mu.Unlock()
	if e != nil {
		w.failed = true
		return elapsed, e
	}
	// Writer serialization reserved a retirement slot before preparing. Readers
	// can arrive meanwhile, but cannot consume additional retired-version slots.
	if before.refs > 0 {
		w.retired = append(w.retired, before)
	}
	w.current = next
	return elapsed, nil
}
func (w *PrivateGraphWriter) Insert(ctx context.Context, id string, vector []float32, level int) (PrivateWriteTiming, error) {
	return w.write(ctx, id, vector, level, false)
}
func (w *PrivateGraphWriter) Delete(ctx context.Context, id string) (PrivateWriteTiming, error) {
	return w.write(ctx, id, nil, 0, true)
}

// Close waits for the current write but does not invalidate live leases. A
// capacity error leaves the owner open for a later close attempt after release.
// Successful close drops this owner's roots, not external constructor aliases.
// Persistence storage is caller-owned and must be closed separately afterward.
func (w *PrivateGraphWriter) Close(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-w.gate:
	}
	defer func() { w.gate <- struct{}{} }()
	if e := ctx.Err(); e != nil {
		return e
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	if w.readers != 0 {
		return ErrCapacity
	}
	w.closed = true
	w.current = nil
	w.retired = nil
	w.ids = nil
	w.persist = nil
	return nil
}
