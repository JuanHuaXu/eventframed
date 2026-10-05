package researchpublicationstore

import (
	"context"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// AsOfSnapshot is one journal's independent availability-time commitment.
type AsOfSnapshot struct {
	Captured model.Snapshot
	AsOf     time.Time
}

// WithResearchAsOfSnapshotsWait validates every commitment under one owned
// writer permit. The callback must not mutate or close this store and must
// commit its own durable batch before returning success.
func (s *Store) WithResearchAsOfSnapshotsWait(ctx context.Context, checks []AsOfSnapshot, work func() error) error {
	if ctx == nil || work == nil || len(checks) == 0 || len(checks) > 8 {
		return errors.New("invalid research as-of batch")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, ok := ctx.Deadline(); !ok {
		return errors.New("research as-of batch requires a deadline")
	}
	commitments := append([]AsOfSnapshot(nil), checks...)
	for _, check := range commitments {
		if check.AsOf.IsZero() {
			return errors.New("research as-of batch has a zero horizon")
		}
	}
	start := s.timingStart()
	if err := s.writer.Acquire(ctx, 1); err != nil {
		s.failedGate("asof-batch", start)
		return err
	}
	defer s.finishGate("asof-batch", start, s.timingStart())
	if err := ctx.Err(); err != nil {
		return err
	}
	current := s.EventStore.Snapshot(ctx)
	if s.lineageFault.Load() || !s.publication.CommittedSnapshotMatches(current) {
		return errors.New("stale or quarantined research as-of batch")
	}
	for _, check := range commitments {
		if !s.publication.Compatible(check.Captured, check.AsOf) {
			return errors.New("stale or quarantined research as-of batch")
		}
	}
	return work()
}
