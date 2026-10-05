package researchpublicationstore

import (
	"context"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

// Wrap preserves the backend's optional vector capability without advertising
// it for backends that cannot provide vectors. New constructs only the core
// adapter; research service integration should use this capability-aware factory.
func Wrap(backend store.EventStore) (store.EventStore, error) {
	return wrapWithRecorder(backend, nil)
}

// WrapWithDurableLineage preserves optional vector capabilities while opting
// into a separately owned research sidecar. The default Wrap remains unchanged.
func WrapWithDurableLineage(backend store.EventStore, path string, create bool) (store.EventStore, error) {
	base, err := NewWithDurableLineage(backend, path, create)
	if err != nil {
		return nil, err
	}
	if vectors, ok := backend.(store.VectorEventStore); ok {
		return &vectorStore{Store: base, vectors: vectors}, nil
	}
	return base, nil
}

// WrapMeasured preserves the same capabilities and admission semantics as
// Wrap. The recorder must be installed before any concurrent use.
func WrapMeasured(backend store.EventStore, recorder *GateRecorder) (store.EventStore, error) {
	if recorder == nil {
		return nil, errors.New("missing research gate recorder")
	}
	return wrapWithRecorder(backend, recorder)
}

func wrapWithRecorder(backend store.EventStore, recorder *GateRecorder) (store.EventStore, error) {
	base, err := New(backend)
	if err != nil {
		return nil, err
	}
	base.timing = recorder
	if vectors, ok := backend.(store.VectorEventStore); ok {
		return &vectorStore{Store: base, vectors: vectors}, nil
	}
	return base, nil
}

type vectorStore struct {
	*Store
	vectors store.VectorEventStore
}

func (s *vectorStore) GetEventsWithVectors(ctx context.Context, tenant string, ids []string, at time.Time) ([]model.Event, error) {
	return s.vectors.GetEventsWithVectors(ctx, tenant, ids, at)
}

// The older research shadow API can use the same owned publication history.
// Neither method grants authority for publishing a forecast into serving.
func (s *Store) ResearchSnapshotCompatible(ctx context.Context, captured model.Snapshot, at time.Time) bool {
	return s.ResearchPublicationCompatible(ctx, captured, at)
}
