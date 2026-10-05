package researchindex

import (
	"errors"
	"sync"
)

// BuildPairCache belongs to ONE immutable build. IDs must identify fixed vectors
// and distance must be deterministic and safe for concurrent calls. Never reuse
// this cache across graphs, ID reuse, vector updates or metric changes. Directed
// keys preserve even an asymmetric callback's exact float32 result.
type BuildPairCache struct {
	mu           sync.Mutex
	values       map[uint64]float32
	limit        int
	hits, misses uint64
}
type BuildPairStats struct {
	Hits, Misses uint64
	Entries      int
}

func NewBuildPairCache(limit int) (*BuildPairCache, error) {
	if limit < 0 || limit > 65536 {
		return nil, errors.New("invalid build cache contract")
	}
	return &BuildPairCache{values: make(map[uint64]float32), limit: limit}, nil
}

// compute must return the same metric value for this pair throughout the build.
// The closure keeps the caller's already-resolved vectors on the miss path.
func (c *BuildPairCache) Distance(a, b uint32, compute func() float32) float32 {
	key := uint64(a)<<32 | uint64(b)
	c.mu.Lock()
	if value, ok := c.values[key]; ok {
		c.hits++
		c.mu.Unlock()
		return value
	}
	c.misses++
	c.mu.Unlock()
	// Concurrent misses may compute twice. Computing outside the lock avoids
	// serializing the expensive metric; deterministic immutable inputs agree.
	value := compute()
	c.mu.Lock()
	if len(c.values) < c.limit {
		c.values[key] = value
	}
	c.mu.Unlock()
	return value
}
func (c *BuildPairCache) Stats() BuildPairStats {
	c.mu.Lock()
	defer c.mu.Unlock()
	return BuildPairStats{c.hits, c.misses, len(c.values)}
}
