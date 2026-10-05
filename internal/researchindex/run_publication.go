package researchindex

import (
	"context"
	"errors"
)

// AcquireCurrent and PublishConsolidation share the pool lock. Serving callers
// must use this pair exclusively, never raw snapshots or direct consolidation
// publication. Lock order is pool then writer, never writer then pool.
func (p *RunLeasePool) AcquireCurrent(ctx context.Context, w *DurableRunWriter) (*RunLease, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, ErrServingClosed
	}
	if p.leases >= p.maxLeases {
		return nil, ErrServingBusy
	}
	s, err := w.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	return p.acquireLocked(s)
}

// Capacity rejection leaves the candidate unsettled so its owner can abort or
// retry it. After successful publication the replaced graphs belong to the pool.
func (p *RunLeasePool) PublishConsolidation(ctx context.Context, c *RunConsolidation) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return ErrServingClosed
	}
	if c == nil {
		return errors.New("nil consolidation")
	}
	if len(c.runs) > p.maxRetired-len(p.retired) {
		return ErrCapacity
	}
	seen := map[*ImmutableRun]bool{}
	for _, r := range c.runs {
		if r == nil || seen[r] {
			return errors.New("invalid replaced run set")
		}
		seen[r] = true
		if _, exists := p.retired[r]; exists {
			return errors.New("replaced run already retired")
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Holding pool.mu is the reservation: no competing publication/retirement
	// can consume capacity or acquire the old snapshot during this transition.
	retired, err := c.Publish(ctx)
	if err != nil {
		return err
	}
	for _, r := range retired {
		p.retired[r] = false
	}
	for _, r := range retired {
		if p.refs[r] == 0 {
			err = errors.Join(err, p.schedule(r))
		}
	}
	return err
}
