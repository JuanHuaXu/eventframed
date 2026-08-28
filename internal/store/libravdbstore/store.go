package libravdbstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

type Config struct {
	Path           string
	Dimension      int
	Quantization   string
	MemoryMapping  bool
	EmbeddingModel string
}

type Store struct {
	db          *libra.Database
	config      Config
	mu          sync.Mutex
	writeMu     sync.Mutex
	collections map[string]*libra.Collection
	snapshot    model.Snapshot
}

const systemCollection = "_eventframe_system"

type persistentState struct {
	SchemaVersion  uint64         `json:"schema_version"`
	EmbeddingModel string         `json:"embedding_model"`
	Dimension      int            `json:"dimension"`
	Quantization   string         `json:"quantization"`
	Snapshot       model.Snapshot `json:"snapshot"`
}

func Open(config Config) (*Store, error) {
	if config.Path == "" {
		return nil, errors.New("database path is required")
	}
	if config.Dimension <= 0 {
		return nil, errors.New("dimension must be positive")
	}
	if config.EmbeddingModel == "" {
		return nil, errors.New("embedding model key is required")
	}
	db, err := libra.Open(
		libra.WithStoragePath(config.Path),
		libra.WithMaxConcurrentWrites(2),
		libra.WithMaxWriteQueueDepth(64),
	)
	if err != nil {
		return nil, fmt.Errorf("open libravdb: %w", err)
	}
	s := &Store{db: db, config: config, collections: make(map[string]*libra.Collection)}
	if err := s.loadOrInitializeState(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Put(ctx context.Context, event model.Event, vector []float32, digest string) (store.PutResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	collectionKey := collectionName(event.TenantID, s.config.EmbeddingModel)
	collection, err := s.collection(ctx, event.TenantID)
	if err != nil {
		return store.PutResult{}, err
	}
	record, err := collection.Get(ctx, event.ID)
	if err == nil {
		if current, _ := record.Metadata["content_digest"].(string); current != digest {
			return store.PutResult{}, store.ErrIdempotencyConflict
		}
		return store.PutResult{Duplicate: true, Snapshot: s.snapshot}, nil
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
	next := s.snapshot
	next.RuntimeVersion++
	next.EvidenceEpoch++
	stateMetadata, err := s.stateMetadata(next)
	if err != nil {
		return store.PutResult{}, err
	}
	err = s.db.WithTx(ctx, func(tx libra.Tx) error {
		if err := tx.Insert(ctx, collectionKey, event.ID, vector, metadata); err != nil {
			return err
		}
		return tx.Upsert(ctx, systemCollection, "runtime", nil, stateMetadata)
	})
	if err != nil {
		// Resolve the timeout/racing-insert ambiguity through the durable record.
		if current, getErr := collection.Get(ctx, event.ID); getErr == nil {
			if currentDigest, _ := current.Metadata["content_digest"].(string); currentDigest == digest {
				return store.PutResult{Duplicate: true, Snapshot: s.snapshot}, nil
			}
			return store.PutResult{}, store.ErrIdempotencyConflict
		}
		return store.PutResult{}, fmt.Errorf("insert event: %w", err)
	}
	s.snapshot = next
	return store.PutResult{Snapshot: next}, nil
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
	stats := store.Stats{Backend: "libravdb", MemoryBytes: databaseStats.MemoryUsage}
	for name, collection := range databaseStats.Collections {
		if !strings.HasPrefix(name, "events_v2_") {
			continue
		}
		stats.Tenants++
		stats.Events += collection.LiveRecordCount
	}
	return stats, nil
}

func (s *Store) Delete(ctx context.Context, tenantID, eventID string) (store.DeleteResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	collection, err := s.collection(ctx, tenantID)
	if err != nil {
		return store.DeleteResult{}, err
	}
	if _, err := collection.Get(ctx, eventID); errors.Is(err, libra.ErrRecordNotFound) {
		return store.DeleteResult{Snapshot: s.snapshot}, nil
	} else if err != nil {
		return store.DeleteResult{}, err
	}
	next := invalidatedSnapshot(s.snapshot)
	metadata, err := s.stateMetadata(next)
	if err != nil {
		return store.DeleteResult{}, err
	}
	key := collectionName(tenantID, s.config.EmbeddingModel)
	err = s.db.WithTx(ctx, func(tx libra.Tx) error {
		if err := tx.Delete(ctx, key, eventID); err != nil {
			return err
		}
		return tx.Upsert(ctx, systemCollection, "runtime", nil, metadata)
	})
	if err != nil {
		return store.DeleteResult{}, fmt.Errorf("delete event: %w", err)
	}
	s.snapshot = next
	return store.DeleteResult{Deleted: true, Snapshot: next}, nil
}

func (s *Store) DeleteBefore(ctx context.Context, tenantID string, before time.Time, limit int) (store.RetentionResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	collection, err := s.collection(ctx, tenantID)
	if err != nil {
		return store.RetentionResult{}, err
	}
	records, err := collection.ListAll(ctx)
	if err != nil {
		return store.RetentionResult{}, err
	}
	type datedID struct {
		id string
		at time.Time
	}
	eligible := make([]datedID, 0)
	for _, record := range records {
		payload, ok := record.Metadata["event_json"].(string)
		if !ok {
			continue
		}
		var event model.Event
		if json.Unmarshal([]byte(payload), &event) != nil {
			continue
		}
		if event.AvailableAt.Before(before) {
			eligible = append(eligible, datedID{id: event.ID, at: event.AvailableAt})
		}
	}
	sort.Slice(eligible, func(i, j int) bool {
		if eligible[i].at.Equal(eligible[j].at) {
			return eligible[i].id < eligible[j].id
		}
		return eligible[i].at.Before(eligible[j].at)
	})
	if len(eligible) > limit {
		eligible = eligible[:limit]
	}
	ids := make([]string, len(eligible))
	for i := range eligible {
		ids[i] = eligible[i].id
	}
	if len(ids) == 0 {
		return store.RetentionResult{Snapshot: s.snapshot}, nil
	}
	next := invalidatedSnapshot(s.snapshot)
	metadata, err := s.stateMetadata(next)
	if err != nil {
		return store.RetentionResult{}, err
	}
	key := collectionName(tenantID, s.config.EmbeddingModel)
	err = s.db.WithTx(ctx, func(tx libra.Tx) error {
		if err := tx.DeleteBatch(ctx, key, ids); err != nil {
			return err
		}
		return tx.Upsert(ctx, systemCollection, "runtime", nil, metadata)
	})
	if err != nil {
		return store.RetentionResult{}, fmt.Errorf("retention delete: %w", err)
	}
	s.snapshot = next
	return store.RetentionResult{DeletedIDs: ids, Snapshot: next}, nil
}

func (s *Store) Backup(ctx context.Context, destination string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.db.Backup(ctx, destination)
}

func (s *Store) Compact(ctx context.Context) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.db.Vacuum(ctx)
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Snapshot(_ context.Context) model.Snapshot {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.snapshot
}

func (s *Store) collection(ctx context.Context, tenantID string) (*libra.Collection, error) {
	name := collectionName(tenantID, s.config.EmbeddingModel)
	s.mu.Lock()
	defer s.mu.Unlock()
	if collection := s.collections[name]; collection != nil {
		return collection, nil
	}
	names, err := s.db.ListCollectionsWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	for _, existingName := range names {
		if existingName != name {
			continue
		}
		collection, err := s.db.GetCollection(name)
		if err != nil {
			return nil, fmt.Errorf("load tenant collection: %w", err)
		}
		if collection.Dimension() != s.config.Dimension {
			return nil, fmt.Errorf("tenant collection dimension %d does not match %d", collection.Dimension(), s.config.Dimension)
		}
		s.collections[name] = collection
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

func collectionName(tenantID, embeddingModel string) string {
	digest := sha256.Sum256([]byte(tenantID + "\x00" + embeddingModel))
	return "events_v2_" + hex.EncodeToString(digest[:12])
}

func (s *Store) loadOrInitializeState(ctx context.Context) error {
	collection, err := s.db.EnsureCollection(ctx, systemCollection, 0, libra.WithMetadataOnly())
	if err != nil {
		return fmt.Errorf("ensure system collection: %w", err)
	}
	record, err := collection.Get(ctx, "runtime")
	if err == nil {
		encoded, ok := record.Metadata["state_json"].(string)
		if !ok {
			return errors.New("durable runtime state is malformed")
		}
		var state persistentState
		if err := json.Unmarshal([]byte(encoded), &state); err != nil {
			return fmt.Errorf("decode durable runtime state: %w", err)
		}
		if state.SchemaVersion != 2 {
			return fmt.Errorf("database schema %d requires migration", state.SchemaVersion)
		}
		if state.EmbeddingModel != s.config.EmbeddingModel || state.Dimension != s.config.Dimension || state.Quantization != s.config.Quantization {
			return fmt.Errorf("database embedding contract %q/d%d does not match active %q/d%d", state.EmbeddingModel, state.Dimension, s.config.EmbeddingModel, s.config.Dimension)
		}
		s.snapshot = state.Snapshot
		return nil
	}
	if !errors.Is(err, libra.ErrRecordNotFound) {
		return fmt.Errorf("read durable runtime state: %w", err)
	}
	names, listErr := s.db.ListCollectionsWithContext(ctx)
	if listErr != nil {
		return fmt.Errorf("inspect database schema: %w", listErr)
	}
	for _, name := range names {
		if isLegacyCollection(name) {
			return errors.New("legacy Phase 1 collections detected; run with -migrate-v1 and -migration-backup before starting")
		}
	}
	s.snapshot = model.Snapshot{PolicyVersion: 1, ContractVersion: 2, GraphVersion: 1, PosteriorVersion: 1, ResidualVersion: 1, AbstractionVersion: 1}
	metadata, err := s.stateMetadata(s.snapshot)
	if err != nil {
		return err
	}
	if err := collection.Insert(ctx, "runtime", nil, metadata); err != nil {
		return fmt.Errorf("initialize durable runtime state: %w", err)
	}
	return nil
}

func isLegacyCollection(name string) bool {
	return strings.HasPrefix(name, "events_") && !strings.HasPrefix(name, "events_v2_")
}

func (s *Store) stateMetadata(snapshot model.Snapshot) (map[string]interface{}, error) {
	encoded, err := json.Marshal(persistentState{SchemaVersion: 2, EmbeddingModel: s.config.EmbeddingModel, Dimension: s.config.Dimension, Quantization: s.config.Quantization, Snapshot: snapshot})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"state_json": string(encoded)}, nil
}

func invalidatedSnapshot(snapshot model.Snapshot) model.Snapshot {
	snapshot.RuntimeVersion++
	snapshot.EvidenceEpoch++
	snapshot.GraphVersion++
	snapshot.PosteriorVersion++
	snapshot.ResidualVersion++
	snapshot.AbstractionVersion++
	return snapshot
}
