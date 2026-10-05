package researchindex

import (
	"context"
	"errors"
	"math"
	"sync/atomic"
)

// PartitionView is a coherent global revision. Partition-local revisions are
// compaction sequence counters, not externally comparable event revisions.
type PartitionView struct {
	revision uint64
	views    []View
}

func (v PartitionView) Revision() uint64 { return v.revision }
func (v PartitionView) Count() int       { return len(v.views) }
func (v PartitionView) DeltaCount() int {
	n := 0
	for _, s := range v.views {
		n += s.DeltaCount()
	}
	return n
}
func (v PartitionView) Partition(i int) (View, error) {
	if i < 0 || i >= len(v.views) {
		return View{}, errors.New("invalid partition")
	}
	return v.views[i], nil
}
func (v PartitionView) Lookup(id string) ([]float32, bool) {
	p, err := ResearchPartition(id, len(v.views))
	if err != nil {
		return nil, false
	}
	return v.views[p].Lookup(id)
}

// PartitionedDurable keeps ONE total delta budget and ONE durable revision across
// all shards. It owns each core exclusively. This is research infrastructure;
// index-pair publication and off-heap reader retirement are separate concerns.
type PartitionedDurable struct {
	gate       chan struct{}
	cores      []*Generations
	snapshot   PartitionView
	capacity   int
	persist    PersistGeneration
	failed     bool
	compacting atomic.Bool
}

// RestorePartitioned has the same trusted coherent-snapshot requirement as
// RestoreDurable. The layout count is fixed for this coordinator's lifetime.
func RestorePartitioned(dimension, capacity, count int, revision uint64, records []Mutation, persist PersistGeneration) (*PartitionedDurable, error) {
	if count < 1 || count > 16 || persist == nil {
		return nil, errors.New("invalid partitioned contract")
	}
	groups := make([][]Mutation, count)
	for _, r := range records {
		p, err := ResearchPartition(r.ID, count)
		if err != nil {
			return nil, err
		}
		groups[p] = append(groups[p], r)
	}
	d := &PartitionedDurable{gate: make(chan struct{}, 1), capacity: capacity, persist: persist, snapshot: PartitionView{revision: revision, views: make([]View, count)}}
	for i, rs := range groups {
		single, err := RestoreDurable(dimension, capacity, revision, rs, persist)
		if err != nil {
			return nil, err
		}
		d.cores = append(d.cores, single.core)
		d.snapshot.views[i] = single.core.View()
	}
	d.gate <- struct{}{}
	return d, nil
}
func (d *PartitionedDurable) lock(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-d.gate:
	}
	if err := ctx.Err(); err != nil {
		d.gate <- struct{}{}
		return err
	}
	return nil
}
func (d *PartitionedDurable) View(ctx context.Context) (PartitionView, error) {
	if err := d.lock(ctx); err != nil {
		return PartitionView{}, err
	}
	defer func() { d.gate <- struct{}{} }()
	if d.failed {
		return PartitionView{}, ErrRecoveryRequired
	}
	return d.snapshot, nil
}

func (d *PartitionedDurable) Apply(ctx context.Context, changes []Mutation) error {
	if err := d.lock(ctx); err != nil {
		return err
	}
	defer func() { d.gate <- struct{}{} }()
	if d.failed {
		return ErrRecoveryRequired
	}
	if len(changes) < 1 || len(changes) > d.capacity {
		return ErrCapacity
	}
	if d.snapshot.revision == math.MaxUint64 {
		return errors.New("revision overflow")
	}
	groups := make([][]Mutation, len(d.cores))
	for _, r := range changes {
		p, err := ResearchPartition(r.ID, len(groups))
		if err != nil {
			return err
		}
		groups[p] = append(groups[p], r)
	}
	prepared := make([]*PreparedGeneration, len(groups))
	settled, entered := false, false
	defer func() {
		if !settled {
			if entered {
				d.failed = true
			}
			for _, p := range prepared {
				if p != nil {
					_ = p.Abort()
				}
			}
		}
	}()
	next := PartitionView{revision: d.snapshot.revision + 1, views: append([]View(nil), d.snapshot.views...)}
	for i, rs := range groups {
		if len(rs) == 0 {
			continue
		}
		p, err := d.cores[i].Prepare(ctx, rs)
		if err != nil {
			return err
		}
		prepared[i] = p
		next.views[i] = View{p.next}
	}
	if next.DeltaCount() > d.capacity {
		return ErrCapacity
	}
	owned := make([]Mutation, 0, len(changes))
	for _, r := range changes {
		p, _ := ResearchPartition(r.ID, len(groups))
		pr := prepared[p].next.delta[r.ID]
		owned = append(owned, Mutation{ID: r.ID, Vector: append([]float32(nil), pr.vector...), Delete: pr.deleted})
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// A callback error never proves rollback. Once dispatched, uncertainty quarantines
	// the whole layout; preexisting views remain explicitly historical.
	entered = true
	if err := d.persist(ctx, next.revision, owned); err != nil {
		return errors.Join(ErrRecoveryRequired, err)
	}
	for _, p := range prepared {
		if p != nil {
			if err := p.Commit(); err != nil {
				return errors.Join(ErrRecoveryRequired, err)
			}
		}
	}
	d.snapshot = next
	settled = true
	return nil
}

type PartitionCompaction struct {
	owner     *PartitionedDurable
	partition int
	candidate *Compaction
	finished  atomic.Bool
}

// PrepareCompaction reserves one global builder but materializes only the
// selected partition. Other writes may continue; publication rebases its delta.
func (d *PartitionedDurable) PrepareCompaction(ctx context.Context, partition int) (*PartitionCompaction, error) {
	if partition < 0 || partition >= len(d.cores) {
		return nil, errors.New("invalid partition")
	}
	if !d.compacting.CompareAndSwap(false, true) {
		return nil, ErrCompactionBusy
	}
	ok := false
	defer func() {
		if !ok {
			d.compacting.Store(false)
		}
	}()
	if err := d.lock(ctx); err != nil {
		return nil, err
	}
	failed := d.failed
	d.gate <- struct{}{}
	if failed {
		return nil, ErrRecoveryRequired
	}
	c, err := d.cores[partition].PrepareCompaction(ctx)
	if err != nil {
		return nil, err
	}
	ok = true
	return &PartitionCompaction{owner: d, partition: partition, candidate: c}, nil
}
func (c *PartitionCompaction) Publish(ctx context.Context) error {
	if !c.finished.CompareAndSwap(false, true) {
		return ErrFinished
	}
	d := c.owner
	defer d.compacting.Store(false)
	if err := d.lock(ctx); err != nil {
		_ = c.candidate.Abort()
		return err
	}
	defer func() { d.gate <- struct{}{} }()
	if d.failed {
		_ = c.candidate.Abort()
		return ErrRecoveryRequired
	}
	next := PartitionView{revision: d.snapshot.revision, views: append([]View(nil), d.snapshot.views...)}
	if err := c.candidate.Publish(ctx); err != nil {
		return err
	}
	next.views[c.partition] = d.cores[c.partition].View()
	d.snapshot = next
	return nil
}
func (c *PartitionCompaction) Abort() error {
	if !c.finished.CompareAndSwap(false, true) {
		return ErrFinished
	}
	defer c.owner.compacting.Store(false)
	return c.candidate.Abort()
}
