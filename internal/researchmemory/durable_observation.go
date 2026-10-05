package researchmemory

import "context"

// WaitProcessed observes completion of an absolute number of durable labels.
// It does not establish successful fitting or authorize the label's source;
// callers must inspect Counts and keep their service dependency guard.
func (d *Durable) WaitProcessed(ctx context.Context, target uint64) error {
	return d.worker.WaitProcessed(ctx, target)
}

// Counts reports a worker snapshot without changing admission or fitting.
func (d *Durable) Counts() (completed, failed uint64, pending, queued int) {
	return d.worker.Counts()
}
