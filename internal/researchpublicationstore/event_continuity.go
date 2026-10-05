package researchpublicationstore

import (
	"context"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchlineage"
)

// ResearchEventUnchanged checks source lineage under the wrapper's writer
// gate. The default wrapper knows only its current lifetime; the opt-in
// durable sidecar can resume when its checkpoint matches the owned backend.
// Neither mode grants permission to publish a transferred model.
func (s *Store) ResearchEventUnchanged(ctx context.Context, from, target model.Snapshot, tenant, event string) (bool, error) {
	if ctx == nil || tenant == "" || event == "" {
		return false, errors.New("invalid research event continuity request")
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if s.lineageFault.Load() {
		return false, errors.New("research lineage faulted")
	}
	if !s.writer.TryAcquire(1) {
		return false, errors.New("research event continuity guard busy")
	}
	defer s.writer.Release(1)
	return s.researchEventUnchangedHeld(ctx, from, target, tenant, event)
}

// WithResearchSourceAsOfSnapshotWait holds one writer permit across as-of
// validity, source continuity, and the caller's guarded durable operation.
// The callback must not reenter the writer gate or mutate this store.
func (s *Store) WithResearchSourceAsOfSnapshotWait(ctx context.Context, from model.Snapshot, asOf time.Time, tenant, event string, work func(model.Snapshot) error) error {
	if ctx == nil || tenant == "" || event == "" || asOf.IsZero() || work == nil {
		return errors.New("invalid research source guard")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, ok := ctx.Deadline(); !ok {
		return errors.New("research source guard wait requires a deadline")
	}
	start := s.timingStart()
	if err := s.writer.Acquire(ctx, 1); err != nil {
		s.failedGate("source-asof", start)
		return err
	}
	defer s.finishGate("source-asof", start, s.timingStart())
	if err := ctx.Err(); err != nil {
		return err
	}
	target := s.EventStore.Snapshot(ctx)
	if s.lineageFault.Load() || !s.publication.CommittedSnapshotMatches(target) || !s.publication.Compatible(from, asOf) {
		return errors.New("stale or quarantined research source")
	}
	continuous, err := s.researchEventUnchangedHeld(ctx, from, target, tenant, event)
	if err != nil {
		return err
	}
	if !continuous {
		return errors.New("research source continuity changed")
	}
	return work(target)
}

// researchEventUnchangedHeld requires the caller to own the writer permit.
func (s *Store) researchEventUnchangedHeld(ctx context.Context, from, target model.Snapshot, tenant, event string) (bool, error) {
	current := s.EventStore.Snapshot(ctx)
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if s.lineageFault.Load() {
		return false, errors.New("research lineage faulted")
	}
	if current != target || !s.publication.CommittedSnapshotMatches(current) {
		return false, errors.New("research event continuity history unavailable")
	}
	if s.lineage != nil {
		return s.lineage.Unchanged(ctx, from, target, researchlineage.EventKey{Tenant: tenant, Event: event})
	}
	if from.RuntimeVersion < s.initial.RuntimeVersion || from.RuntimeVersion > current.RuntimeVersion || (from.RuntimeVersion == s.initial.RuntimeVersion && from != s.initial) {
		return false, errors.New("research event continuity history unavailable")
	}
	return s.eventTouches[eventTouchKey{tenant, event}] <= from.RuntimeVersion, nil
}
