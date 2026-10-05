// Package researchpublicationstore is an unconfigured research adapter. It must
// exclusively own its backend: mutations through another handle bypass its proof.
// No daemon constructor or production service installs it.
package researchpublicationstore

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchlineage"
	"github.com/JuanHuaXu/eventframed/internal/researchpublication"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type Store struct {
	store.EventStore
	writer         *semaphore.Weighted
	publication    *researchpublication.Publisher
	timing         *GateRecorder
	initial        model.Snapshot
	eventTouches   map[eventTouchKey]uint64
	lineage        *researchlineage.Ledger
	lineageFault   atomic.Bool
	pendingTouches []researchlineage.EventKey
}

type eventTouchKey struct{ tenant, event string }

func New(backend store.EventStore) (*Store, error) {
	if backend == nil {
		return nil, errors.New("nil research backend")
	}
	initial := backend.Snapshot(context.Background())
	return &Store{EventStore: backend, writer: semaphore.NewWeighted(1), publication: researchpublication.New(initial), initial: initial, eventTouches: make(map[eventTouchKey]uint64)}, nil
}

// NewWithDurableLineage is opt-in research state. create must be false for a
// restart: silently creating a missing sidecar would erase deletion history.
func NewWithDurableLineage(backend store.EventStore, path string, create bool) (*Store, error) {
	s, err := New(backend)
	if err != nil {
		return nil, err
	}
	if create {
		s.lineage, err = researchlineage.Create(path, s.initial)
	} else {
		s.lineage, err = researchlineage.Open(path, s.initial)
	}
	if err != nil {
		return nil, err
	}
	motion, err := s.lineage.Motion(context.Background())
	if err != nil {
		_ = s.lineage.Close()
		return nil, err
	}
	s.publication, err = researchpublication.NewWithMotion(s.initial, motion)
	if err != nil {
		_ = s.lineage.Close()
		return nil, err
	}
	return s, nil
}

func mutate[T any](s *Store, kind researchpublication.Kind, at time.Time, call func() (T, error)) (result T, err error) {
	return mutateWithTouches(s, kind, at, call, nil)
}

func mutateWithTouches[T any](s *Store, kind researchpublication.Kind, at time.Time, call func() (T, error), touched func(T, model.Snapshot)) (result T, err error) {
	// Existing mutators retain their unconditional serialization; only the new
	// research guard may cancel while waiting for this shared one-token gate.
	start := s.timingStart()
	_ = s.writer.Acquire(context.Background(), 1)
	gateKind := "general"
	if kind == researchpublication.Ingestion {
		gateKind = "ingestion"
	}
	defer s.finishGate(gateKind, start, s.timingStart())
	if s.lineageFault.Load() {
		return result, errors.New("research lineage faulted")
	}
	s.pendingTouches = nil
	before := s.EventStore.Snapshot(context.Background())
	ticket, err := s.publication.Begin(kind, at)
	if err != nil {
		return result, err
	}
	finished := false
	// Errors and panics can follow a durable commit. Never assume rollback.
	defer func() {
		if !finished {
			_ = s.publication.Abort(ticket, true)
		}
	}()
	result, err = call()
	if err != nil {
		return result, err
	}
	after := s.EventStore.Snapshot(context.Background())
	if after == before {
		if kind == researchpublication.Ingestion {
			return result, errors.New("research ingestion needs authoritative duplicate resolution")
		}
		err = s.publication.Abort(ticket, false)
	} else {
		err = s.publication.Commit(ticket, after)
	}
	if err == nil && after != before {
		if touched != nil {
			touched(result, after)
		}
		finished = true
		if s.lineage != nil {
			ingestionAt := time.Time{}
			if kind == researchpublication.Ingestion {
				ingestionAt = at
			}
			if err = s.lineage.RecordWithMotion(context.Background(), before, after, s.pendingTouches, ingestionAt); err != nil {
				s.lineageFault.Store(true)
				return result, err
			}
		}
	}
	finished = err == nil
	return result, err
}

func (s *Store) touchEvent(tenant, event string, after model.Snapshot) {
	if s.lineage != nil {
		s.pendingTouches = append(s.pendingTouches, researchlineage.EventKey{Tenant: tenant, Event: event})
		return
	}
	s.eventTouches[eventTouchKey{tenant, event}] = after.RuntimeVersion
}

func (s *Store) ResearchPublicationCompatible(ctx context.Context, captured model.Snapshot, at time.Time) bool {
	return ctx.Err() == nil && !s.lineageFault.Load() && s.publication.Compatible(captured, at)
}

// Journals do not change forecast-state versions. Their own durable identity and
// temporal validation remains in the backend; this is not a skipped write.
func (s *Store) PutBayesianJournal(ctx context.Context, j model.BayesianJournalEntry) error {
	return s.EventStore.PutBayesianJournal(ctx, j)
}

func (s *Store) Close() error {
	start := s.timingStart()
	_ = s.writer.Acquire(context.Background(), 1)
	defer s.finishGate("close", start, s.timingStart())
	ticket, err := s.publication.Begin(researchpublication.General, time.Time{})
	// Closing after quarantine must still release the backend's resources.
	defer func() {
		if err == nil {
			_ = s.publication.Abort(ticket, true)
		}
	}()
	backendErr := s.EventStore.Close()
	if s.lineage != nil {
		return errors.Join(backendErr, s.lineage.Close())
	}
	return backendErr
}
