package store

import (
	"context"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

var ErrIdempotencyConflict = errors.New("event id already exists with different content")

type PutResult struct {
	Duplicate bool
	Snapshot  model.Snapshot
}

type DeleteResult struct {
	Deleted  bool
	Snapshot model.Snapshot
}
type RetentionResult struct {
	DeletedIDs []string
	Snapshot   model.Snapshot
}

type SearchResult struct {
	Event      model.Event
	Similarity float64
}

type Stats struct {
	Backend     string `json:"backend"`
	Tenants     int    `json:"tenants"`
	Events      int    `json:"events"`
	MemoryBytes int64  `json:"memory_bytes"`
}

type EventStore interface {
	Put(ctx context.Context, event model.Event, vector []float32, digest string) (PutResult, error)
	Delete(ctx context.Context, tenantID, eventID string) (DeleteResult, error)
	DeleteBefore(ctx context.Context, tenantID string, before time.Time, limit int) (RetentionResult, error)
	Backup(ctx context.Context, destination string) error
	Compact(ctx context.Context) error
	Search(ctx context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]SearchResult, error)
	Stats(ctx context.Context) (Stats, error)
	Snapshot(ctx context.Context) model.Snapshot
	Close() error
}
