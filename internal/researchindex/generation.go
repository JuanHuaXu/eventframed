package researchindex

import (
	"context"
	"errors"
	"math"
	"sync/atomic"
)

var (
	ErrCapacity       = errors.New("research index delta full")
	ErrFinished       = errors.New("research index operation already finished")
	ErrCompactionBusy = errors.New("research index compaction already active")
)

type Mutation struct {
	ID     string
	Vector []float32
	Delete bool
}
type record struct {
	vector   []float32
	deleted  bool
	revision uint64
}
type baseGeneration struct{ records map[string]record }
type generation struct {
	base     *baseGeneration
	delta    map[string]record
	revision uint64
}

// Generations models publication ownership only. It has no WAL or ANN engine.
// Vectors are copied at ingress and never mutated after publication. Old views
// retain old generations through Go GC, not an off-heap reclamation protocol.
type Generations struct {
	current             atomic.Pointer[generation]
	writer              chan struct{}
	compacting          atomic.Bool
	dimension, capacity int
}

func NewGenerations(dimension, capacity int) (*Generations, error) {
	if dimension < 1 || capacity < 1 || capacity > 128 {
		return nil, errors.New("invalid generation bounds")
	}
	m := &Generations{writer: make(chan struct{}, 1), dimension: dimension, capacity: capacity}
	m.writer <- struct{}{}
	m.current.Store(&generation{base: &baseGeneration{records: map[string]record{}}, delta: map[string]record{}})
	return m, nil
}

func (m *Generations) lock(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-m.writer:
	}
	if err := ctx.Err(); err != nil {
		m.writer <- struct{}{}
		return err
	}
	return nil
}

type View struct{ g *generation }

func (m *Generations) View() View { return View{m.current.Load()} }
func (v View) Revision() uint64   { return v.g.revision }
func (v View) DeltaCount() int    { return len(v.g.delta) }
func (v View) Lookup(id string) ([]float32, bool) {
	r, ok := v.g.delta[id]
	if !ok {
		r, ok = v.g.base.records[id]
	}
	if !ok || r.deleted {
		return nil, false
	}
	return append([]float32(nil), r.vector...), true
}

type PreparedGeneration struct {
	manager  *Generations
	next     *generation
	finished atomic.Bool
}

// Prepare owns writer exclusion until Commit or Abort. The caller must settle
// every successful preparation. A durable failure must abort only if rollback
// is known; uncertain durability requires reconciliation outside this model.
func (m *Generations) Prepare(ctx context.Context, mutations []Mutation) (*PreparedGeneration, error) {
	if len(mutations) < 1 || len(mutations) > m.capacity {
		return nil, ErrCapacity
	}
	if err := m.lock(ctx); err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			m.writer <- struct{}{}
		}
	}()
	before := m.current.Load()
	if before.revision == math.MaxUint64 {
		return nil, errors.New("revision overflow")
	}
	delta := make(map[string]record, len(before.delta)+len(mutations))
	for id, r := range before.delta {
		delta[id] = r
	}
	seen := make(map[string]bool, len(mutations))
	for _, change := range mutations {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if change.ID == "" || seen[change.ID] {
			return nil, errors.New("missing or duplicate mutation ID")
		}
		seen[change.ID] = true
		r := record{deleted: change.Delete, revision: before.revision + 1}
		if !change.Delete {
			if len(change.Vector) != m.dimension {
				return nil, errors.New("invalid vector dimension")
			}
			var norm float64
			for _, v := range change.Vector {
				if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
					return nil, errors.New("nonfinite vector")
				}
				norm += float64(v) * float64(v)
			}
			if norm == 0 {
				return nil, errors.New("zero vector")
			}
			r.vector = append([]float32(nil), change.Vector...)
		}
		delta[change.ID] = r
	}
	if len(delta) > m.capacity {
		return nil, ErrCapacity
	}
	next := &generation{base: before.base, delta: delta, revision: before.revision + 1}
	ok = true
	return &PreparedGeneration{manager: m, next: next}, nil
}

// Commit only publishes prebuilt state and releases exclusion. It intentionally
// ignores cancellation: a matching durable commit must not be left invisible.
// This model does not itself establish that durability actually happened.
func (p *PreparedGeneration) Commit() error {
	if !p.finished.CompareAndSwap(false, true) {
		return ErrFinished
	}
	p.manager.current.Store(p.next)
	p.manager.writer <- struct{}{}
	return nil
}
func (p *PreparedGeneration) Abort() error {
	if !p.finished.CompareAndSwap(false, true) {
		return ErrFinished
	}
	p.manager.writer <- struct{}{}
	return nil
}

type Compaction struct {
	manager  *Generations
	captured *generation
	built    *baseGeneration
	finished atomic.Bool
}

// PrepareCompaction materializes an exact base as a lifecycle reference, not an
// HNSW builder. One build is allowed at a time; writes may continue during it.
func (m *Generations) PrepareCompaction(ctx context.Context) (*Compaction, error) {
	if !m.compacting.CompareAndSwap(false, true) {
		return nil, ErrCompactionBusy
	}
	g := m.current.Load()
	records := make(map[string]record, len(g.base.records)+len(g.delta))
	for id, r := range g.base.records {
		if err := ctx.Err(); err != nil {
			m.compacting.Store(false)
			return nil, err
		}
		records[id] = r
	}
	for id, r := range g.delta {
		if err := ctx.Err(); err != nil {
			m.compacting.Store(false)
			return nil, err
		}
		if r.deleted {
			delete(records, id)
		} else {
			records[id] = r
		}
	}
	if err := ctx.Err(); err != nil {
		m.compacting.Store(false)
		return nil, err
	}
	return &Compaction{manager: m, captured: g, built: &baseGeneration{records: records}}, nil
}

// Publish carries newer delta entries forward, including tombstones. It does
// not change the semantic revision, because compaction is representational.
func (c *Compaction) Publish(ctx context.Context) error {
	if !c.finished.CompareAndSwap(false, true) {
		return ErrFinished
	}
	m := c.manager
	defer m.compacting.Store(false)
	if err := m.lock(ctx); err != nil {
		return err
	}
	defer func() { m.writer <- struct{}{} }()
	current := m.current.Load()
	if current.base != c.captured.base {
		return errors.New("compaction base changed")
	}
	delta := make(map[string]record, len(current.delta))
	for id, r := range current.delta {
		if r.revision > c.captured.revision {
			delta[id] = r
		}
	}
	m.current.Store(&generation{base: c.built, delta: delta, revision: current.revision})
	return nil
}
func (c *Compaction) Abort() error {
	if !c.finished.CompareAndSwap(false, true) {
		return ErrFinished
	}
	c.manager.compacting.Store(false)
	return nil
}
