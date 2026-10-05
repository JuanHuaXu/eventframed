package researchindex

import (
	"context"
	"errors"
	"sync"
)

// RunLeasePool bounds admitted readers and retired GRAPH handles, not batches.
// Publication is still the coordinator's job: retire only graphs removed from
// its current layout. Do not expose unleased raw snapshots to serving clients.
type RunLeasePool struct {
	mu                            sync.Mutex
	refs                          map[*ImmutableRun]int
	retired                       map[*ImmutableRun]bool // value means cleanup has been scheduled
	queue                         *retirementQueue
	maxLeases, maxRetired, leases int
	closed                        bool
	closeRun                      func(*ImmutableRun) error
}
type RunLease struct {
	mu       sync.Mutex
	pool     *RunLeasePool
	snapshot RunSnapshot
	released bool
}

func NewRunLeasePool(maxLeases, maxRetired int) (*RunLeasePool, error) {
	if maxLeases < 1 || maxLeases > 128 || maxRetired < 1 || maxRetired > 30 {
		return nil, errors.New("invalid run lease bounds")
	}
	return &RunLeasePool{refs: map[*ImmutableRun]int{}, retired: map[*ImmutableRun]bool{}, queue: newRetirementQueue(maxRetired), maxLeases: maxLeases, maxRetired: maxRetired, closeRun: func(r *ImmutableRun) error { return r.Close() }}, nil
}
func (p *RunLeasePool) Acquire(s RunSnapshot) (*RunLease, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.acquireLocked(s)
}
func (p *RunLeasePool) acquireLocked(s RunSnapshot) (*RunLease, error) {
	if p.closed {
		return nil, ErrServingClosed
	}
	if p.leases >= p.maxLeases {
		return nil, ErrServingBusy
	}
	if s.search == nil {
		return nil, errors.New("invalid run snapshot")
	}
	for _, r := range s.search.runs {
		if _, retired := p.retired[r]; retired {
			return nil, ErrServingBusy
		}
		r.index.mu.RLock()
		closed := r.index.closed
		r.index.mu.RUnlock()
		if closed {
			return nil, ErrServingClosed
		}
	}
	for _, r := range s.search.runs {
		p.refs[r]++
	}
	p.leases++
	return &RunLease{pool: p, snapshot: s}, nil
}
func (l *RunLease) Search(ctx context.Context, q []float32) ([]Candidate, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.released {
		return nil, ErrServingClosed
	}
	return l.snapshot.Search(ctx, q)
}
func (l *RunLease) Release() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.released {
		return ErrFinished
	}
	p := l.pool
	p.mu.Lock()
	defer p.mu.Unlock()
	var err error
	for _, r := range l.snapshot.search.runs {
		p.refs[r]--
		if p.refs[r] == 0 {
			delete(p.refs, r)
			if _, retired := p.retired[r]; retired {
				err = errors.Join(err, p.schedule(r))
			}
		}
	}
	p.leases--
	l.released = true
	l.snapshot = RunSnapshot{}
	l.pool = nil
	return err
}

// schedule holds pool.mu. Completion keeps failed handles charged, so a failed
// close cannot silently restore capacity. The queue never waits on this mutex.
func (p *RunLeasePool) schedule(r *ImmutableRun) error {
	if p.retired[r] {
		return nil
	}
	err := p.queue.submit(func() error { return p.closeRun(r) }, func(err error) {
		p.mu.Lock()
		defer p.mu.Unlock()
		if err == nil {
			delete(p.retired, r)
		}
	})
	if err == nil {
		p.retired[r] = true
	}
	return err
}
func (p *RunLeasePool) Retire(runs []*ImmutableRun) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return ErrServingClosed
	}
	if len(runs) > p.maxRetired-len(p.retired) {
		return ErrCapacity
	}
	seen := map[*ImmutableRun]bool{}
	for _, r := range runs {
		if r == nil || seen[r] {
			return errors.New("invalid retirement set")
		}
		seen[r] = true
		if _, ok := p.retired[r]; ok {
			return errors.New("already retired")
		}
	}
	for _, r := range runs {
		p.retired[r] = false
	}
	var err error
	for _, r := range runs {
		if p.refs[r] == 0 {
			err = errors.Join(err, p.schedule(r))
		}
	}
	return err
}

// Close stops admission and joins admitted cleanup; it does not close graphs
// still in the coordinator's current layout. Active leases prevent shutdown.
func (p *RunLeasePool) Close() error {
	p.mu.Lock()
	if p.leases > 0 {
		p.mu.Unlock()
		return ErrServingBusy
	}
	p.closed = true
	p.mu.Unlock()
	return p.queue.close()
}
