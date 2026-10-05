package researchindex

import (
	"context"
	"errors"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

// HNSWBase is an immutable, research-only ANN base. Its private derived database
// is not the authoritative event store. Close must follow retirement of readers;
// subsequent queries fail rather than touching a closed index.
type HNSWBase struct {
	mu         sync.RWMutex
	db         *libra.Database
	collection *libra.Collection
	base       *baseGeneration
	dimension  int
	closed     bool
	closeErr   error
}

// BuildHNSWBase creates a NEW directory and builds only the captured base, not
// its delta. The directory is retained even on failure for inspection. Building
// occurs off the serving path; no production collection is changed.
func BuildHNSWBase(ctx context.Context, view View, dimension int, directory string) (*HNSWBase, error) {
	if view.g == nil || dimension < 1 {
		return nil, errors.New("invalid base view")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := os.Mkdir(directory, 0700); err != nil {
		return nil, err
	}
	db, err := libra.Open(libra.WithStoragePath(filepath.Join(directory, "base.libravdb")))
	if err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			_ = db.Close()
		}
	}()
	c, err := db.EnsureCollection(ctx, "base", dimension, libra.WithMetric(libra.CosineDistance), libra.WithHNSW(16, 200, 100))
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(view.g.base.records))
	for id := range view.g.base.records {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		r := view.g.base.records[id]
		if len(r.vector) != dimension {
			return nil, errors.New("base dimension mismatch")
		}
		if err := c.Insert(ctx, id, r.vector, nil); err != nil {
			return nil, err
		}
	}
	success = true
	return &HNSWBase{db: db, collection: c, base: view.g.base, dimension: dimension}, nil
}

func cosine(a, b []float32) float64 {
	var dot, aa, bb float64
	for i, v := range a {
		dot += float64(v) * float64(b[i])
		aa += float64(v) * float64(v)
		bb += float64(b[i]) * float64(b[i])
	}
	return dot / math.Sqrt(aa*bb)
}

func (h *HNSWBase) Search(ctx context.Context, view View, query []float32, k int) ([]Candidate, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return nil, errors.New("research HNSW base closed")
	}
	if view.g == nil || view.g.base != h.base {
		return nil, errors.New("base generation mismatch")
	}
	if len(query) != h.dimension || k < 1 || k > 200 {
		return nil, errors.New("invalid query")
	}
	var norm float64
	for _, v := range query {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return nil, errors.New("invalid query vector")
		}
		norm += float64(v) * float64(v)
	}
	if norm == 0 {
		return nil, errors.New("zero query")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var base []Candidate
	limit := min(k+len(view.g.delta), len(h.base.records))
	if limit > 0 {
		results, err := h.collection.Search(ctx, query, limit)
		if err != nil {
			return nil, err
		}
		for _, r := range results.Results {
			owned, ok := h.base.records[r.ID]
			if !ok {
				return nil, errors.New("foreign base ID")
			}
			// Re-score returned IDs with the same precision and formula as delta.
			base = append(base, Candidate{r.ID, cosine(query, owned.vector)})
		}
	}
	delta := make([]Change, 0, len(view.g.delta))
	for id, r := range view.g.delta {
		score := 0.0
		if !r.deleted {
			score = cosine(query, r.vector)
		}
		delta = append(delta, Change{Candidate{id, score}, r.deleted})
	}
	return Merge(base, delta, k)
}

func (h *HNSWBase) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return h.closeErr
	}
	h.closed = true
	h.closeErr = h.db.Close()
	return h.closeErr
}
