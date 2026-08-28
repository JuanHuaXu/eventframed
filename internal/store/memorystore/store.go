package memorystore

import (
	"context"
	"errors"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type entry struct {
	event  model.Event
	vector []float32
	digest string
}

type Store struct {
	mu       sync.RWMutex
	entries  map[string]map[string]entry
	snapshot model.Snapshot
}

func New() *Store {
	return &Store{entries: make(map[string]map[string]entry), snapshot: initialSnapshot()}
}

func (s *Store) Put(_ context.Context, event model.Event, vector []float32, digest string) (store.PutResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tenant := s.entries[event.TenantID]
	if tenant == nil {
		tenant = make(map[string]entry)
		s.entries[event.TenantID] = tenant
	}
	if current, ok := tenant[event.ID]; ok {
		if current.digest != digest {
			return store.PutResult{}, store.ErrIdempotencyConflict
		}
		return store.PutResult{Duplicate: true, Snapshot: s.snapshot}, nil
	}
	tenant[event.ID] = entry{event: event, vector: append([]float32(nil), vector...), digest: digest}
	s.snapshot.RuntimeVersion++
	s.snapshot.EvidenceEpoch++
	return store.PutResult{Snapshot: s.snapshot}, nil
}

func (s *Store) Search(_ context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]store.SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	results := make([]store.SearchResult, 0, len(s.entries[tenantID]))
	for _, candidate := range s.entries[tenantID] {
		if candidate.event.AvailableAt.After(availableBy) {
			continue
		}
		results = append(results, store.SearchResult{Event: candidate.event, Similarity: cosine(vector, candidate.vector)})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Similarity > results[j].Similarity })
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func (s *Store) Stats(_ context.Context) (store.Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stats := store.Stats{Backend: "memory", Tenants: len(s.entries)}
	for _, tenant := range s.entries {
		stats.Events += len(tenant)
	}
	return stats, nil
}

func (s *Store) Delete(_ context.Context, tenantID, eventID string) (store.DeleteResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.entries[tenantID][eventID]; !ok {
		return store.DeleteResult{Snapshot: s.snapshot}, nil
	}
	delete(s.entries[tenantID], eventID)
	s.invalidate()
	return store.DeleteResult{Deleted: true, Snapshot: s.snapshot}, nil
}

func (s *Store) DeleteBefore(_ context.Context, tenantID string, before time.Time, limit int) (store.RetentionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0)
	for id, item := range s.entries[tenantID] {
		if item.event.AvailableAt.Before(before) && len(ids) < limit {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		delete(s.entries[tenantID], id)
	}
	if len(ids) > 0 {
		s.invalidate()
	}
	return store.RetentionResult{DeletedIDs: ids, Snapshot: s.snapshot}, nil
}

func (s *Store) Backup(context.Context, string) error {
	return errors.New("memory store does not support backup")
}
func (s *Store) Compact(context.Context) error { return nil }

func (s *Store) Close() error { return nil }

func (s *Store) Snapshot(_ context.Context) model.Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot
}

func initialSnapshot() model.Snapshot {
	return model.Snapshot{PolicyVersion: 1, ContractVersion: 2, GraphVersion: 1, PosteriorVersion: 1, ResidualVersion: 1, AbstractionVersion: 1}
}

func (s *Store) invalidate() {
	s.snapshot.RuntimeVersion++
	s.snapshot.EvidenceEpoch++
	s.snapshot.GraphVersion++
	s.snapshot.PosteriorVersion++
	s.snapshot.ResidualVersion++
	s.snapshot.AbstractionVersion++
}

func cosine(left, right []float32) float64 {
	if len(left) != len(right) || len(left) == 0 {
		return 0
	}
	var dot, leftNorm, rightNorm float64
	for index := range left {
		l, r := float64(left[index]), float64(right[index])
		dot += l * r
		leftNorm += l * l
		rightNorm += r * r
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return dot / math.Sqrt(leftNorm*rightNorm)
}
