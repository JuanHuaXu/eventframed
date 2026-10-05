package researchpublicationstore

import (
	"context"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// WithResearchSnapshot excludes owned backend mutations while a bounded external
// ledger operation runs. The callback may read the backend, but must not call a
// mutator/Close on this adapter or retain the guard. Busy acquisition rejects;
// callers decide when to retry. This is not a cross-database atomic transaction.
func (s *Store) WithResearchSnapshot(ctx context.Context, captured model.Snapshot, work func() error) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	if work == nil {
		return errors.New("missing research operation")
	}
	start := s.timingStart()
	if !s.writer.TryAcquire(1) {
		s.failedGate("try-snapshot", start)
		return errors.New("research snapshot guard busy")
	}
	defer s.finishGate("try-snapshot", start, s.timingStart())
	return s.withHeldResearchSnapshot(ctx, captured, work)
}

// WithResearchSnapshotWait queues behind owned mutations until its required
// context deadline. It does not extend snapshot validity or bound a callback
// that ignores cancellation. Callers must bound their own work and never mutate
// or close this adapter from the callback.
func (s *Store) WithResearchSnapshotWait(ctx context.Context, captured model.Snapshot, work func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, ok := ctx.Deadline(); !ok {
		return errors.New("research guard wait requires a deadline")
	}
	if work == nil {
		return errors.New("missing research operation")
	}
	start := s.timingStart()
	if err := s.writer.Acquire(ctx, 1); err != nil {
		s.failedGate("snapshot", start)
		return err
	}
	defer s.finishGate("snapshot", start, s.timingStart())
	return s.withHeldResearchSnapshot(ctx, captured, work)
}

func (s *Store) withHeldResearchSnapshot(ctx context.Context, captured model.Snapshot, work func() error) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	current := s.EventStore.Snapshot(ctx)
	if s.lineageFault.Load() || current != captured || !s.publication.Compatible(captured, time.Now()) {
		return errors.New("stale or quarantined research snapshot")
	}
	return work()
}

// WithResearchAsOfSnapshotWait accepts only complete, owned future-ingestion
// history relative to asOf while holding the mutation gate through work.
// Callers must bind asOf to the operation being validated. This does not authorize
// labels, restore model-history authority, or tolerate unknown version motion.
func (s *Store) WithResearchAsOfSnapshotWait(ctx context.Context, captured model.Snapshot, asOf time.Time, work func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, ok := ctx.Deadline(); !ok {
		return errors.New("research guard wait requires a deadline")
	}
	if asOf.IsZero() || work == nil {
		return errors.New("missing as-of research operation")
	}
	start := s.timingStart()
	if err := s.writer.Acquire(ctx, 1); err != nil {
		s.failedGate("asof", start)
		return err
	}
	defer s.finishGate("asof", start, s.timingStart())
	if err := ctx.Err(); err != nil {
		return err
	}
	current := s.EventStore.Snapshot(ctx)
	if s.lineageFault.Load() || !s.publication.CommittedSnapshotMatches(current) || !s.publication.Compatible(captured, asOf) {
		return errors.New("stale or quarantined as-of research snapshot")
	}
	return work()
}
