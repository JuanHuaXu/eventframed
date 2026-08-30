package rankdelta

import (
	"container/list"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	_ "modernc.org/sqlite"
)

type Record struct {
	TenantID           string
	Key                string
	EventID            string
	Delta              float64
	Reliability        float64
	PolicyVersion      uint64
	EvidenceEpoch      uint64
	GraphVersion       uint64
	PosteriorVersion   uint64
	ResidualVersion    uint64
	AbstractionVersion uint64
	UpdatedAt          time.Time
	ExpiresAt          time.Time
}

func (r Record) ValidFor(snapshot model.Snapshot, now time.Time) bool {
	return r.TenantID != "" && r.Key != "" && r.EventID != "" &&
		!math.IsNaN(r.Delta) && !math.IsInf(r.Delta, 0) && r.Delta >= -1 && r.Delta <= 1 &&
		r.Reliability >= 0 && r.Reliability <= 1 &&
		r.PolicyVersion == snapshot.PolicyVersion && r.EvidenceEpoch == snapshot.EvidenceEpoch &&
		r.GraphVersion == snapshot.GraphVersion && r.PosteriorVersion == snapshot.PosteriorVersion &&
		r.ResidualVersion == snapshot.ResidualVersion && r.AbstractionVersion == snapshot.AbstractionVersion &&
		!r.UpdatedAt.After(now) && (r.ExpiresAt.IsZero() || now.Before(r.ExpiresAt))
}

type Store interface {
	GetBatch(context.Context, string, []string, model.Snapshot, time.Time) (map[string]Record, error)
	PutBatch(context.Context, []Record) error
	Close() error
}

type SQLiteStore struct {
	db    *sql.DB
	cache *lruCache
}

func Open(path string, cacheEntries int) (*SQLiteStore, error) {
	if !filepath.IsAbs(path) {
		return nil, errors.New("rank-delta SQLite path must be absolute")
	}
	if cacheEntries <= 0 {
		cacheEntries = 100_000
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create rank-delta directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open rank-delta SQLite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	for _, statement := range []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA synchronous=NORMAL`,
		`PRAGMA busy_timeout=5000`,
		`CREATE TABLE IF NOT EXISTS rank_deltas (
			tenant_id TEXT NOT NULL,
			delta_key TEXT NOT NULL,
			event_id TEXT NOT NULL,
			delta REAL NOT NULL,
			confidence REAL NOT NULL,
			policy_version INTEGER NOT NULL,
			evidence_epoch INTEGER NOT NULL,
			graph_version INTEGER NOT NULL,
			posterior_version INTEGER NOT NULL,
			residual_version INTEGER NOT NULL,
			abstraction_version INTEGER NOT NULL,
			updated_at_ns INTEGER NOT NULL,
			expires_at_ns INTEGER NOT NULL,
			PRIMARY KEY (tenant_id, delta_key)
		) WITHOUT ROWID`,
		`CREATE INDEX IF NOT EXISTS rank_deltas_event ON rank_deltas (tenant_id, event_id)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("initialize rank-delta SQLite: %w", err)
		}
	}
	if err := os.Chmod(path, 0o600); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("secure rank-delta SQLite: %w", err)
	}
	return &SQLiteStore{db: db, cache: newLRU(cacheEntries)}, nil
}

func (s *SQLiteStore) GetBatch(ctx context.Context, tenantID string, keys []string, snapshot model.Snapshot, now time.Time) (map[string]Record, error) {
	result := make(map[string]Record, len(keys))
	missing := make([]string, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		if record, ok := s.cache.get(cacheKey(tenantID, key)); ok && record.ValidFor(snapshot, now) {
			result[key] = record
			continue
		}
		missing = append(missing, key)
	}
	if len(missing) == 0 {
		return result, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(missing)), ",")
	arguments := make([]any, 0, len(missing)+1)
	arguments = append(arguments, tenantID)
	for _, key := range missing {
		arguments = append(arguments, key)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT tenant_id, delta_key, event_id, delta, confidence,
		policy_version, evidence_epoch, graph_version, posterior_version, residual_version,
		abstraction_version, updated_at_ns, expires_at_ns
		FROM rank_deltas WHERE tenant_id = ? AND delta_key IN (`+placeholders+`)`, arguments...)
	if err != nil {
		return nil, fmt.Errorf("read rank deltas: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		record, scanErr := scanRecord(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		s.cache.put(cacheKey(record.TenantID, record.Key), record)
		if record.ValidFor(snapshot, now) {
			result[record.Key] = record
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rank deltas: %w", err)
	}
	return result, nil
}

func (s *SQLiteStore) PutBatch(ctx context.Context, records []Record) error {
	if len(records) == 0 {
		return nil
	}
	pending := make([]Record, 0, len(records))
	for _, record := range records {
		if err := validateRecord(record); err != nil {
			return err
		}
		if cached, ok := s.cache.get(cacheKey(record.TenantID, record.Key)); ok && materiallyEqual(cached, record) {
			continue
		}
		pending = append(pending, record)
	}
	if len(pending) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin rank-delta transaction: %w", err)
	}
	statement, err := tx.PrepareContext(ctx, `INSERT INTO rank_deltas (
		tenant_id, delta_key, event_id, delta, confidence, policy_version, evidence_epoch,
		graph_version, posterior_version, residual_version, abstraction_version, updated_at_ns, expires_at_ns
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(tenant_id, delta_key) DO UPDATE SET
		event_id=excluded.event_id, delta=excluded.delta, confidence=excluded.confidence,
		policy_version=excluded.policy_version, evidence_epoch=excluded.evidence_epoch,
		graph_version=excluded.graph_version, posterior_version=excluded.posterior_version,
		residual_version=excluded.residual_version, abstraction_version=excluded.abstraction_version,
		updated_at_ns=excluded.updated_at_ns, expires_at_ns=excluded.expires_at_ns`)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("prepare rank-delta upsert: %w", err)
	}
	defer statement.Close()
	for _, record := range pending {
		if _, err := statement.ExecContext(ctx, record.TenantID, record.Key, record.EventID, record.Delta, record.Reliability,
			record.PolicyVersion, record.EvidenceEpoch, record.GraphVersion, record.PosteriorVersion,
			record.ResidualVersion, record.AbstractionVersion, record.UpdatedAt.UnixNano(), timeToUnixNano(record.ExpiresAt)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("write rank delta: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit rank deltas: %w", err)
	}
	for _, record := range pending {
		s.cache.put(cacheKey(record.TenantID, record.Key), record)
	}
	return nil
}

func (s *SQLiteStore) Close() error { return s.db.Close() }

type scanner interface{ Scan(...any) error }

func scanRecord(row scanner) (Record, error) {
	var record Record
	var updatedAt, expiresAt int64
	err := row.Scan(&record.TenantID, &record.Key, &record.EventID, &record.Delta, &record.Reliability,
		&record.PolicyVersion, &record.EvidenceEpoch, &record.GraphVersion, &record.PosteriorVersion,
		&record.ResidualVersion, &record.AbstractionVersion, &updatedAt, &expiresAt)
	if err != nil {
		return Record{}, fmt.Errorf("decode rank delta: %w", err)
	}
	record.UpdatedAt = time.Unix(0, updatedAt).UTC()
	if expiresAt != 0 {
		record.ExpiresAt = time.Unix(0, expiresAt).UTC()
	}
	return record, nil
}

func validateRecord(record Record) error {
	if record.TenantID == "" || record.Key == "" || record.EventID == "" || record.UpdatedAt.IsZero() {
		return errors.New("rank delta requires tenant, key, event, and update time")
	}
	if math.IsNaN(record.Delta) || math.IsInf(record.Delta, 0) || record.Delta < -1 || record.Delta > 1 || record.Reliability < 0 || record.Reliability > 1 {
		return errors.New("rank delta or correction reliability is outside its bounded range")
	}
	return nil
}

func materiallyEqual(left, right Record) bool {
	left.UpdatedAt, right.UpdatedAt = time.Time{}, time.Time{}
	return left == right
}

func timeToUnixNano(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.UnixNano()
}

func cacheKey(tenantID, key string) string { return tenantID + "\x00" + key }

type cacheEntry struct {
	key    string
	record Record
}

type lruCache struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

func newLRU(capacity int) *lruCache {
	return &lruCache{capacity: capacity, items: make(map[string]*list.Element, capacity), order: list.New()}
}

func (c *lruCache) get(key string) (Record, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	element, ok := c.items[key]
	if !ok {
		return Record{}, false
	}
	c.order.MoveToFront(element)
	return element.Value.(cacheEntry).record, true
}

func (c *lruCache) put(key string, record Record) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if element, ok := c.items[key]; ok {
		element.Value = cacheEntry{key: key, record: record}
		c.order.MoveToFront(element)
		return
	}
	element := c.order.PushFront(cacheEntry{key: key, record: record})
	c.items[key] = element
	if c.order.Len() > c.capacity {
		oldest := c.order.Back()
		delete(c.items, oldest.Value.(cacheEntry).key)
		c.order.Remove(oldest)
	}
}
