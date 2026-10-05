package researchmemory

import (
	"context"
	"errors"
)

// WaitProcessed waits for an absolute completed+failed count, not successful
// learning. Callers must still inspect failures. It creates no goroutine and
// does not change feedback ordering, publication, or cancellation of worker jobs.
func (b *Background) WaitProcessed(ctx context.Context, target uint64) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		b.mu.Lock()
		if b.completed+b.failed >= target {
			b.mu.Unlock()
			return nil
		}
		if b.closed {
			b.mu.Unlock()
			return errors.New("background closed before target")
		}
		// Register and check the counter under the publisher's lock: a completion
		// cannot fall between observing the counter and subscribing to its change.
		if b.changed == nil {
			b.changed = make(chan struct{})
		}
		changed := b.changed
		b.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		}
	}
}

// Allocation is lazy: polling-only clients do not allocate notification channels.
// Closing broadcasts to every waiter; each rechecks its own target under mu.
func (b *Background) notifyLocked() {
	if b.changed != nil {
		close(b.changed)
		b.changed = nil
	}
}
