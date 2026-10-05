package researchindex

import (
	"context"
	"sync/atomic"
)

type DurableCompaction struct {
	owner     *DurableGenerations
	candidate *Compaction
	finished  atomic.Bool
}

// PrepareCompaction checks health, then builds outside the durable admission
// gate. A concurrent transaction may change health or revision during the build;
// Publish must recheck health under the same gate used by Apply.
func (d *DurableGenerations) PrepareCompaction(ctx context.Context) (*DurableCompaction, error) {
	if err := d.lock(ctx); err != nil {
		return nil, err
	}
	failed := d.failed
	d.gate <- struct{}{}
	if failed {
		return nil, ErrRecoveryRequired
	}
	c, err := d.core.PrepareCompaction(ctx)
	if err != nil {
		return nil, err
	}
	return &DurableCompaction{owner: d, candidate: c}, nil
}

func (c *DurableCompaction) Publish(ctx context.Context) error {
	if !c.finished.CompareAndSwap(false, true) {
		return ErrFinished
	}
	d := c.owner
	if err := d.lock(ctx); err != nil {
		_ = c.candidate.Abort()
		return err
	}
	defer func() { d.gate <- struct{}{} }()
	if d.failed {
		_ = c.candidate.Abort()
		return ErrRecoveryRequired
	}
	return c.candidate.Publish(ctx)
}

func (c *DurableCompaction) Abort() error {
	if !c.finished.CompareAndSwap(false, true) {
		return ErrFinished
	}
	return c.candidate.Abort()
}
