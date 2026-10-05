// Package researchadmission is an isolated, non-preemptive admission experiment.
// It changes when work starts, never its evidence or durability requirements.
package researchadmission

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type Kind uint8

const (
	Outcome Kind = iota
	Recall
	Ingest
)

var (
	ErrClosed   = errors.New("admission closed")
	ErrCapacity = errors.New("admission pending cap exceeded")
	ErrKind     = errors.New("invalid admission kind")
)

type waiter struct {
	kind    Kind
	ready   chan struct{}
	granted bool
	err     error
}

type Stats struct {
	Queued        [3]int
	Granted       [3]uint64
	Cancelled     [3]uint64
	ActiveReaders int
	ActiveWriter  bool
	MaxReaders    int
	PeakQueued    int
	Closed        bool
}

type Scheduler struct {
	mu                     sync.Mutex
	queues                 [3][]*waiter
	stats                  Stats
	maxReaders, maxPending int
	cursor, cohortGrants   int
}

// NewScheduler uses a fixed cycle: outcome, outcome, bounded Recall cohort, ingest.
// Empty slots are skipped; no prediction score changes the service order.
// This is slot-weighted rotation, not cost-normalized deficit round robin.
func NewScheduler(maxReaders, maxPending int) (*Scheduler, error) {
	if maxReaders < 1 || maxReaders > 256 || maxPending < 1 || maxPending > 4096 {
		return nil, errors.New("invalid admission bounds")
	}
	return &Scheduler{maxReaders: maxReaders, maxPending: maxPending}, nil
}

type Lease struct {
	gate     *Scheduler
	kind     Kind
	released atomic.Bool
}

func (l *Lease) Release() {
	if l == nil || l.released.Swap(true) {
		return
	}
	l.gate.mu.Lock()
	l.gate.releaseLocked(l.kind)
	l.gate.mu.Unlock()
}

func (g *Scheduler) Acquire(ctx context.Context, kind Kind) (*Lease, error) {
	if kind > Ingest {
		return nil, ErrKind
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	g.mu.Lock()
	if g.stats.Closed {
		g.mu.Unlock()
		return nil, ErrClosed
	}
	queued := g.stats.Queued[0] + g.stats.Queued[1] + g.stats.Queued[2]
	if queued >= g.maxPending {
		g.mu.Unlock()
		return nil, ErrCapacity
	}
	w := &waiter{kind: kind, ready: make(chan struct{})}
	g.queues[kind] = append(g.queues[kind], w)
	g.stats.Queued[kind]++
	queued++
	if queued > g.stats.PeakQueued {
		g.stats.PeakQueued = queued
	}
	g.dispatchLocked()
	g.mu.Unlock()
	select {
	case <-w.ready:
		if w.err != nil {
			return nil, w.err
		}
		lease := &Lease{gate: g, kind: kind}
		if err := ctx.Err(); err != nil {
			lease.Release()
			return nil, err
		}
		return lease, nil
	case <-ctx.Done():
		g.mu.Lock()
		if w.granted {
			g.releaseLocked(kind)
		} else {
			for i, pending := range g.queues[kind] {
				if pending == w {
					copy(g.queues[kind][i:], g.queues[kind][i+1:])
					last := len(g.queues[kind]) - 1
					g.queues[kind][last] = nil
					g.queues[kind] = g.queues[kind][:last]
					g.stats.Queued[kind]--
					g.stats.Cancelled[kind]++
					break
				}
			}
			g.dispatchLocked()
		}
		g.mu.Unlock()
		return nil, ctx.Err()
	}
}

func (g *Scheduler) grantLocked(kind Kind) {
	w := g.queues[kind][0]
	g.queues[kind][0] = nil
	g.queues[kind] = g.queues[kind][1:]
	g.stats.Queued[kind]--
	g.stats.Granted[kind]++
	w.granted = true
	if kind == Recall {
		g.stats.ActiveReaders++
		g.cohortGrants++
		if g.stats.ActiveReaders > g.stats.MaxReaders {
			g.stats.MaxReaders = g.stats.ActiveReaders
		}
	} else {
		g.stats.ActiveWriter = true
	}
	close(w.ready)
}

func (g *Scheduler) dispatchLocked() {
	if g.stats.Closed || g.stats.ActiveWriter {
		return
	}
	if g.stats.ActiveReaders > 0 {
		// New arrivals may fill the original cohort, never extend it forever or
		// join once a writer is pending. This bounds reader-induced writer delay.
		if g.stats.Queued[Outcome]+g.stats.Queued[Ingest] == 0 {
			for g.stats.Queued[Recall] > 0 && g.cohortGrants < g.maxReaders {
				g.grantLocked(Recall)
			}
		}
		return
	}
	g.cohortGrants = 0
	slots := [4]Kind{Outcome, Outcome, Recall, Ingest}
	for scanned := 0; scanned < len(slots); scanned++ {
		index := (g.cursor + scanned) % len(slots)
		kind := slots[index]
		if g.stats.Queued[kind] == 0 {
			continue
		}
		g.cursor = (index + 1) % len(slots)
		if kind == Recall {
			for g.stats.Queued[Recall] > 0 && g.cohortGrants < g.maxReaders {
				g.grantLocked(Recall)
			}
		} else {
			g.grantLocked(kind)
		}
		return
	}
}

func (g *Scheduler) releaseLocked(kind Kind) {
	if kind == Recall {
		g.stats.ActiveReaders--
	} else {
		g.stats.ActiveWriter = false
	}
	g.dispatchLocked()
}

func (g *Scheduler) Snapshot() Stats {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.stats
}

// Close rejects queued/new work; existing lease owners retain permission to
// finish. It does not claim that the associated storage worker has drained.
func (g *Scheduler) Close() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.stats.Closed {
		return
	}
	g.stats.Closed = true
	for k := range g.queues {
		for _, w := range g.queues[k] {
			w.err = ErrClosed
			close(w.ready)
		}
		g.queues[k] = nil
		g.stats.Queued[k] = 0
	}
}
