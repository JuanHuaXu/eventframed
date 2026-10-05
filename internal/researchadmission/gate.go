// Package researchadmission coordinates cooperating research requests. It is
// not installed in the daemon and does not replace store snapshot validation.
package researchadmission

import (
	"context"
	"errors"

	"golang.org/x/sync/semaphore"
)

// Gate permits parallel readers and exclusive mutations within one explicitly
// shared mutation domain. Do not copy it, nest admissions, or use separate gates
// for requests whose store versions can invalidate each other.
type Gate struct {
	permits  *semaphore.Weighted
	capacity int64
}

func New(capacity int64) (*Gate, error) {
	if capacity < 1 {
		return nil, errors.New("admission capacity must be positive")
	}
	return &Gate{permits: semaphore.NewWeighted(capacity), capacity: capacity}, nil
}

func (g *Gate) Read(ctx context.Context, fn func(context.Context) error) error {
	return g.run(ctx, 1, fn)
}

func (g *Gate) Write(ctx context.Context, fn func(context.Context) error) error {
	return g.run(ctx, g.capacity, fn)
}

func (g *Gate) run(ctx context.Context, weight int64, fn func(context.Context) error) error {
	if fn == nil {
		return errors.New("admission callback is nil")
	}
	if err := g.permits.Acquire(ctx, weight); err != nil {
		return err
	}
	defer g.permits.Release(weight)
	// Cancellation after this check is cooperative: callbacks must observe ctx.
	// Never release early while a canceled callback is still accessing the store.
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn(ctx)
}
