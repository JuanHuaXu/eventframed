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
	Search(ctx context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]SearchResult, error)
	Stats(ctx context.Context) (Stats, error)
	Close() error
}
