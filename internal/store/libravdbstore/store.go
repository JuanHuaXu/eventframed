package libravdbstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

type Config struct {
	Path          string
	Dimension     int
	Quantization  string
	MemoryMapping bool
}

type Store struct {
	db          *libra.Database
	config      Config
	mu          sync.Mutex
	collections map[string]*libra.Collection
}

func Open(config Config) (*Store, error) {
	if config.Path == "" {
		return nil, errors.New("database path is required")
	}
	if config.Dimension <= 0 {
		return nil, errors.New("dimension must be positive")
	}
	db, err := libra.Open(
		libra.WithStoragePath(config.Path),
		libra.WithMaxConcurrentWrites(2),
		libra.WithMaxWriteQueueDepth(64),
	)
	if err != nil {
		return nil, fmt.Errorf("open libravdb: %w", err)
	}
	return &Store{db: db, config: config, collections: make(map[string]*libra.Collection)}, nil
}

func (s *Store) Put(ctx context.Context, event model.Event, vector []float32, digest string) (store.PutResult, error) {
	collection, err := s.collection(ctx, event.TenantID)
	if err != nil {
		return store.PutResult{}, err
	}
	record, err := collection.Get(ctx, event.ID)
	if err == nil {
		if current, _ := record.Metadata["content_digest"].(string); current != digest {
			return store.PutResult{}, store.ErrIdempotencyConflict
		}
		return store.PutResult{Duplicate: true}, nil
	}
	if !errors.Is(err, libra.ErrRecordNotFound) {
		return store.PutResult{}, fmt.Errorf("check existing event: %w", err)
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return store.PutResult{}, fmt.Errorf("encode event: %w", err)
	}
	metadata := map[string]interface{}{
		"event_json":     string(payload),
		"content_digest": digest,
		"session_id":     event.SessionID,
		"kind":           event.Kind,
		"available_at":   event.AvailableAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		"priority":       event.Priority,
	}
	if err := collection.Insert(ctx, event.ID, vector, metadata); err != nil {
		// Resolve the timeout/racing-insert ambiguity through the durable record.
		if current, getErr := collection.Get(ctx, event.ID); getErr == nil {
			if currentDigest, _ := current.Metadata["content_digest"].(string); currentDigest == digest {
				return store.PutResult{Duplicate: true}, nil
			}
			return store.PutResult{}, store.ErrIdempotencyConflict
		}
		return store.PutResult{}, fmt.Errorf("insert event: %w", err)
	}
	return store.PutResult{}, nil
}

func (s *Store) Search(ctx context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]store.SearchResult, error) {
	collection, err := s.collection(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	// LibraVDB's metadata filter is applied after ANN limiting in this release.
	// Expand deterministically until enough availability-eligible candidates are
	// found, or every live record has been examined. This prevents a dense set
	// of future records from crowding valid history out before reranking.
	live := collection.Stats(ctx).LiveRecordCount
	probe := min(max(limit, 1), max(live, 1))
	for {
		results, searchErr := collection.Search(ctx, vector, probe)
		if searchErr != nil {
			return nil, fmt.Errorf("search events: %w", searchErr)
		}
		out := decodeEligible(results.Results, availableBy, limit)
		if len(out) >= limit || probe >= live {
			return out, nil
		}
		probe = min(live, probe*2)
	}
}

func decodeEligible(results []*libra.SearchResult, availableBy time.Time, limit int) []store.SearchResult {
	out := make([]store.SearchResult, 0, min(limit, len(results)))
	for _, result := range results {
		payload, ok := result.Metadata["event_json"].(string)
		if !ok {
			continue
		}
		var event model.Event
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			continue
		}
		event.Embedding = nil
		if event.AvailableAt.After(availableBy) {
			continue
		}
		out = append(out, store.SearchResult{Event: event, Similarity: float64(result.Score)})
		if len(out) >= limit {
			break
		}
	}
	return out
}

func (s *Store) Stats(ctx context.Context) (store.Stats, error) {
	databaseStats := s.db.Stats(ctx)
	stats := store.Stats{Backend: "libravdb", Tenants: databaseStats.CollectionCount, MemoryBytes: databaseStats.MemoryUsage}
	for _, collection := range databaseStats.Collections {
		stats.Events += collection.LiveRecordCount
	}
	return stats, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) collection(ctx context.Context, tenantID string) (*libra.Collection, error) {
	name := collectionName(tenantID)
	s.mu.Lock()
	defer s.mu.Unlock()
	if collection := s.collections[name]; collection != nil {
		return collection, nil
	}
	options := []libra.CollectionOption{
		libra.WithMetric(libra.CosineDistance),
		libra.WithHNSW(16, 200, 100),
		libra.WithMemoryMapping(s.config.MemoryMapping),
	}
	switch s.config.Quantization {
	case "", "none":
	case "sq8":
		options = append(options, libra.WithScalarQuantization(8, 0.10))
	case "fsq6":
		options = append(options, libra.WithFSQQuantization(6, 0.10))
	case "pq8":
		options = append(options, libra.WithProductQuantization(8, 8, 0.10))
	default:
		return nil, fmt.Errorf("unsupported quantization %q", s.config.Quantization)
	}
	collection, err := s.db.EnsureCollection(ctx, name, s.config.Dimension, options...)
	if err != nil {
		return nil, fmt.Errorf("ensure tenant collection: %w", err)
	}
	s.collections[name] = collection
	return collection, nil
}

func collectionName(tenantID string) string {
	digest := sha256.Sum256([]byte(tenantID))
	return "events_" + hex.EncodeToString(digest[:12])
}
