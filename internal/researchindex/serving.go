package researchindex

import (
	"context"
	"errors"
	"sync"
)

var ErrServingBusy = errors.New("research index readers or build still active")
var ErrServingClosed = errors.New("research serving index closed")

type indexHandle struct {
	index   *HNSWBase
	readers int
}

// ServingGenerations is the sole compaction owner of its coordinator. Callers
// must not compact that coordinator separately. A lease is an explicit historical
// view; callers must Release it. Bounded retirement does not expire readers.
type ServingGenerations struct {
	mu                 sync.Mutex
	durable            *DurableGenerations
	current            *indexHandle
	retired            map[*indexHandle]bool
	maxRetired, writes int
	maxLeases, leases  int
	building, closed   bool
	closeErr           error
	build              func(context.Context, View, int, string) (*HNSWBase, error)
}

func NewServing(ctx context.Context, d *DurableGenerations, directory string, maxRetired int) (*ServingGenerations, error) {
	return NewServingBounded(ctx, d, directory, maxRetired, 64)
}

func NewServingBounded(ctx context.Context, d *DurableGenerations, directory string, maxRetired, maxLeases int) (*ServingGenerations, error) {
	if d == nil || maxRetired < 1 || maxLeases < 1 {
		return nil, errors.New("invalid serving configuration")
	}
	v, err := d.View(ctx)
	if err != nil {
		return nil, err
	}
	h, err := BuildHNSWBase(ctx, v, d.core.dimension, directory)
	if err != nil {
		return nil, err
	}
	return &ServingGenerations{durable: d, current: &indexHandle{index: h}, retired: map[*indexHandle]bool{}, maxRetired: maxRetired, maxLeases: maxLeases, build: BuildHNSWBase}, nil
}

type ReadLease struct {
	mu       sync.Mutex
	owner    *ServingGenerations
	handle   *indexHandle
	view     View
	released bool
	revision uint64
}

func (s *ServingGenerations) Acquire(ctx context.Context) (*ReadLease, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, ErrServingClosed
	}
	if s.leases >= s.maxLeases {
		return nil, ErrServingBusy
	}
	v, err := s.durable.View(ctx)
	if err != nil {
		return nil, err
	}
	if v.g.base != s.current.index.base {
		return nil, errors.New("external compaction changed serving base")
	}
	s.current.readers++
	s.leases++
	return &ReadLease{owner: s, handle: s.current, view: v, revision: v.Revision()}, nil
}

func (l *ReadLease) Search(ctx context.Context, query []float32, k int) ([]Candidate, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.released {
		return nil, ErrFinished
	}
	return l.handle.index.Search(ctx, l.view, query, k)
}

func (l *ReadLease) Revision() uint64 { return l.revision }

func (l *ReadLease) Release() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.released {
		return ErrFinished
	}
	l.released = true
	s := l.owner
	handle := l.handle
	// A caller retaining a released lease must not pin historical vectors/indexes.
	l.owner, l.handle, l.view = nil, nil, View{}
	s.mu.Lock()
	s.leases--
	handle.readers--
	closeOld := handle.readers == 0 && s.retired[handle]
	s.mu.Unlock()
	if closeOld {
		err := handle.index.Close()
		s.mu.Lock()
		// Keep the resource charged until close actually settles. On error retain
		// the charge: reclamation is unproven and more builds must not hide it.
		if err == nil {
			delete(s.retired, handle)
		}
		s.mu.Unlock()
		return err
	}
	return nil
}

func (s *ServingGenerations) Apply(ctx context.Context, changes []Mutation) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrServingClosed
	}
	s.writes++
	s.mu.Unlock()
	defer func() { s.mu.Lock(); s.writes--; s.mu.Unlock() }()
	return s.durable.Apply(ctx, changes)
}

// Compact builds before publication. The serving mutex hides the short interval
// between generation publication and handle replacement from Acquire. Newer
// writes are carried by the coordinator's compaction rebase. Close errors after
// replacement concern retiring resources, not rollback of the published pair.
func (s *ServingGenerations) Compact(ctx context.Context, directory string) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrServingClosed
	}
	if s.building || len(s.retired) >= s.maxRetired {
		s.mu.Unlock()
		return ErrServingBusy
	}
	s.building = true
	s.mu.Unlock()
	defer func() { s.mu.Lock(); s.building = false; s.mu.Unlock() }()
	c, err := s.durable.PrepareCompaction(ctx)
	if err != nil {
		return err
	}
	defer c.Abort()
	baseView := View{&generation{base: c.candidate.built, delta: map[string]record{}, revision: c.candidate.captured.revision}}
	h, err := s.build(ctx, baseView, s.durable.core.dimension, directory)
	if err != nil {
		return err
	}
	published := false
	defer func() {
		if !published {
			_ = h.Close()
		}
	}()
	next := &indexHandle{index: h}
	s.mu.Lock()
	if err := c.Publish(ctx); err != nil {
		s.mu.Unlock()
		return err
	}
	old := s.current
	s.current = next
	published = true
	s.retired[old] = true
	closeOld := old.readers == 0
	s.mu.Unlock()
	if closeOld {
		err := old.index.Close()
		s.mu.Lock()
		if err == nil {
			delete(s.retired, old)
		}
		s.mu.Unlock()
		return err
	}
	return nil
}

// Close refuses to discard active readers, writes or builds. The caller drains
// them and retries; it does not report success while resources remain in use.
func (s *ServingGenerations) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return s.closeErr
	}
	if s.building || s.writes > 0 || len(s.retired) > 0 || s.current.readers > 0 {
		return ErrServingBusy
	}
	s.closed = true
	s.closeErr = s.current.index.Close()
	return s.closeErr
}
