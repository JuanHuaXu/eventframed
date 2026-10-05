package researchindex

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// PartitionServing owns all compaction for its coordinator. Every query visits
// every fixed partition; the layout is not a semantic nomination shortcut.
type PartitionServing struct {
	mu                                    sync.Mutex
	durable                               *PartitionedDurable
	current                               []*indexHandle
	retired                               map[*indexHandle]bool
	maxRetired, maxLeases, leases, writes int
	building, closed                      bool
	closeErr                              error
	build                                 func(context.Context, View, int, string) (*HNSWBase, error)
}

func NewPartitionServing(ctx context.Context, d *PartitionedDurable, directory string, maxRetired, maxLeases int) (*PartitionServing, error) {
	if d == nil || maxRetired < 1 || maxLeases < 1 {
		return nil, errors.New("invalid partition serving configuration")
	}
	v, err := d.View(ctx)
	if err != nil {
		return nil, err
	}
	if err := os.Mkdir(directory, 0700); err != nil {
		return nil, err
	}
	s := &PartitionServing{durable: d, retired: map[*indexHandle]bool{}, maxRetired: maxRetired, maxLeases: maxLeases, build: BuildHNSWBase}
	for i, pv := range v.views {
		h, err := s.build(ctx, pv, d.cores[i].dimension, filepath.Join(directory, fmt.Sprintf("partition-%d", i)))
		if err != nil {
			for _, old := range s.current {
				err = errors.Join(err, old.index.Close())
			}
			return nil, err
		}
		s.current = append(s.current, &indexHandle{index: h})
	}
	return s, nil
}

type PartitionLease struct {
	mu       sync.Mutex
	owner    *PartitionServing
	handles  []*indexHandle
	view     PartitionView
	revision uint64
	released bool
}

func (s *PartitionServing) Acquire(ctx context.Context) (*PartitionLease, error) {
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
	for i, h := range s.current {
		if v.views[i].g.base != h.index.base {
			return nil, errors.New("external compaction changed partition base")
		}
	}
	l := &PartitionLease{owner: s, handles: append([]*indexHandle(nil), s.current...), view: v, revision: v.Revision()}
	for _, h := range l.handles {
		h.readers++
	}
	s.leases++
	return l, nil
}
func (l *PartitionLease) Revision() uint64 { return l.revision }
func (l *PartitionLease) Search(ctx context.Context, q []float32, k int) ([]Candidate, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.released {
		return nil, ErrFinished
	}
	parts := make([][]Candidate, len(l.handles))
	for i, h := range l.handles {
		cs, err := h.index.Search(ctx, l.view.views[i], q, k)
		if err != nil {
			return nil, err
		}
		parts[i] = cs
	}
	return MergeResearchPartitions(parts, k)
}

// closeRetired keeps a handle charged until close succeeds, including while a
// slow close is in flight. Callers select each zero-reader handle exactly once.
func (s *PartitionServing) closeRetired(h *indexHandle) error {
	err := h.index.Close()
	s.mu.Lock()
	if err == nil {
		delete(s.retired, h)
	}
	s.mu.Unlock()
	return err
}
func (l *PartitionLease) Release() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.released {
		return ErrFinished
	}
	l.released = true
	s := l.owner
	handles := l.handles
	l.owner = nil
	l.handles = nil
	l.view = PartitionView{}
	s.mu.Lock()
	s.leases--
	var closing []*indexHandle
	for _, h := range handles {
		h.readers--
		if h.readers == 0 && s.retired[h] {
			closing = append(closing, h)
		}
	}
	s.mu.Unlock()
	var err error
	for _, h := range closing {
		err = errors.Join(err, s.closeRetired(h))
	}
	return err
}
func (s *PartitionServing) Apply(ctx context.Context, changes []Mutation) error {
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

// Compact builds just one new graph. Publication hides the view/handle swap
// from Acquire. Apply can advance deltas while building, but cannot change bases.
func (s *PartitionServing) Compact(ctx context.Context, partition int, directory string) (err error) {
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
	c, err := s.durable.PrepareCompaction(ctx, partition)
	if err != nil {
		return err
	}
	defer c.Abort()
	baseView := View{&generation{base: c.candidate.built, delta: map[string]record{}, revision: c.candidate.captured.revision}}
	h, err := s.build(ctx, baseView, s.durable.cores[partition].dimension, directory)
	if err != nil {
		return err
	}
	published := false
	next := &indexHandle{index: h}
	defer func() {
		if !published {
			// Unpublished does not mean reclaimed. Charge failed candidates too,
			// so repeated cancellation cannot hide failed off-heap cleanup.
			s.mu.Lock()
			s.retired[next] = true
			s.mu.Unlock()
			err = errors.Join(err, s.closeRetired(next))
		}
	}()
	s.mu.Lock()
	if s.current[partition].index.base != c.candidate.captured.base {
		s.mu.Unlock()
		return errors.New("external compaction changed partition base")
	}
	if err := c.Publish(ctx); err != nil {
		s.mu.Unlock()
		return err
	}
	old := s.current[partition]
	s.current[partition] = next
	published = true
	s.retired[old] = true
	closeOld := old.readers == 0
	s.mu.Unlock()
	if closeOld {
		return s.closeRetired(old)
	}
	return nil
}
func (s *PartitionServing) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return s.closeErr
	}
	if s.building || s.writes > 0 || s.leases > 0 || len(s.retired) > 0 {
		return ErrServingBusy
	}
	s.closed = true
	for _, h := range s.current {
		s.closeErr = errors.Join(s.closeErr, h.index.Close())
	}
	return s.closeErr
}
