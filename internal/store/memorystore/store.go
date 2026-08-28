package memorystore

import (
	"context"
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
	mu      sync.RWMutex
	entries map[string]map[string]entry
}

func New() *Store { return &Store{entries: make(map[string]map[string]entry)} }

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
		return store.PutResult{Duplicate: true}, nil
	}
	tenant[event.ID] = entry{event: event, vector: append([]float32(nil), vector...), digest: digest}
	return store.PutResult{}, nil
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

func (s *Store) Close() error { return nil }

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
