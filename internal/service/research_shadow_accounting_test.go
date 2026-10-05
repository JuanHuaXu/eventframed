package service

import (
	"context"
	"testing"
)

// Both select branches are ready: cancellation can win before dequeue, or the
// worker can remove an accepted job and only then notice cancellation.
func TestResearchShadowCancelledAccounting(t *testing.T) {
	for i := 0; i < 128; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		q := &researchShadowQueue{ctx: ctx, cancel: cancel, jobs: make(chan researchShadowJob, 1), done: make(chan struct{})}
		q.jobs <- researchShadowJob{}
		q.status.Accepted = 1
		cancel()
		go q.run()
		q.Close()
		if q.status.Cancelled != 1 || q.status.Dropped != 0 || q.status.Accepted != q.status.Completed+q.status.Stale+q.status.Failed+q.status.Cancelled {
			t.Fatalf("accepted job disappeared: %+v", q.status)
		}
		q.Close()
		if q.status.Cancelled != 1 {
			t.Fatal("close counted twice")
		}
	}
}
