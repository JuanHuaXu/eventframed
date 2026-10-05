package researchindex

import (
	"context"
	"sync"
)

type layeredVersion struct {
	snapshot LayeredSnapshot
	readers  int
}

// LayeredOwner bounds reader handles, historical versions, and one candidate.
// It is not a byte budget or durable transaction owner. Initial snapshot aliases
// outside this owner remain the caller's responsibility. Serving must use leases;
// no internal root is exported by a lease or prepared candidate.
type LayeredOwner struct {
	mu                              sync.Mutex
	current                         *layeredVersion
	retired                         map[*layeredVersion]bool
	readers, maxReaders, maxRetired int
	preparing, closed               bool
	limits                          LayeredLimits
}
type LayeredLease struct {
	owner   *LayeredOwner
	version *layeredVersion
}
type LayeredCandidate struct {
	owner    *LayeredOwner
	snapshot LayeredSnapshot
	finished bool
}

func NewLayeredOwner(initial LayeredSnapshot, limits LayeredLimits, maxReaders, maxRetired int) (*LayeredOwner, error) {
	if maxReaders < 1 || maxReaders > 128 || maxRetired < 1 || maxRetired > 128 {
		return nil, ErrCapacity
	}
	return &LayeredOwner{current: &layeredVersion{snapshot: initial}, retired: map[*layeredVersion]bool{}, limits: limits, maxReaders: maxReaders, maxRetired: maxRetired}, nil
}
func (o *LayeredOwner) Acquire() (*LayeredLease, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.closed {
		return nil, ErrFinished
	}
	if o.readers >= o.maxReaders {
		return nil, ErrServingBusy
	}
	o.readers++
	o.current.readers++
	return &LayeredLease{owner: o, version: o.current}, nil
}
func (l *LayeredLease) Lookup(id uint32) (LayeredRecord, bool, error) {
	o := l.owner
	o.mu.Lock()
	defer o.mu.Unlock()
	if l.version == nil {
		return LayeredRecord{}, false, ErrFinished
	}
	r, ok := l.version.snapshot.Lookup(id)
	return r, ok, nil
}
func (l *LayeredLease) Release() error {
	o := l.owner
	o.mu.Lock()
	defer o.mu.Unlock()
	if l.version == nil {
		return ErrFinished
	}
	v := l.version
	v.readers--
	o.readers--
	l.version = nil
	if v.readers == 0 {
		delete(o.retired, v)
	}
	return nil
}
func (o *LayeredOwner) Prepare(ctx context.Context, edits []LayeredEdit, global uint64) (*LayeredCandidate, error) {
	o.mu.Lock()
	if o.closed {
		o.mu.Unlock()
		return nil, ErrFinished
	}
	// Reserve a historical slot even if the current root has no reader yet:
	// readers may acquire it while preparation runs outside the mutex.
	if o.preparing || len(o.retired) >= o.maxRetired {
		o.mu.Unlock()
		return nil, ErrServingBusy
	}
	o.preparing = true
	before := o.current.snapshot
	o.mu.Unlock()
	next, _, err := PrepareLayered(ctx, before, edits, global, o.limits)
	if err != nil {
		o.mu.Lock()
		o.preparing = false
		o.mu.Unlock()
		return nil, err
	}
	return &LayeredCandidate{owner: o, snapshot: next}, nil
}
func (p *LayeredCandidate) Commit() error {
	o := p.owner
	o.mu.Lock()
	defer o.mu.Unlock()
	if p.finished {
		return ErrFinished
	}
	if o.current.readers > 0 {
		o.retired[o.current] = true
	}
	o.current = &layeredVersion{snapshot: p.snapshot}
	p.snapshot = LayeredSnapshot{}
	p.finished = true
	o.preparing = false
	return nil
}
func (p *LayeredCandidate) Abort() error {
	o := p.owner
	o.mu.Lock()
	defer o.mu.Unlock()
	if p.finished {
		return ErrFinished
	}
	p.finished = true
	p.snapshot = LayeredSnapshot{}
	o.preparing = false
	return nil
}

// Close refuses active leases/candidates; it never forcibly invalidates a reader.
func (o *LayeredOwner) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.closed {
		return ErrFinished
	}
	if o.readers > 0 || o.preparing {
		return ErrServingBusy
	}
	o.closed = true
	o.current = nil
	clear(o.retired)
	return nil
}
