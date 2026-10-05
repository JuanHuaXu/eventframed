package researchfusion

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
)

type MemoStats struct {
	Hits          int   `json:"hits"`
	Misses        int   `json:"misses"`
	AcquisitionNS int64 `json:"acquisition_ns"`
	Entries       int   `json:"entries"`
}

// Memo is an isolated experiment cache, not a runtime feature. It retains exact
// role-separated embeddings to make arms use the same vectors. Recall timing
// on hits excludes network/model acquisition, which is reported separately.
type Memo struct {
	mu     sync.Mutex
	base   embed.Embedder
	key    string
	dim    int
	cap    int
	values map[string][]float32
	stats  MemoStats
}

func NewMemo(base embed.Embedder, cap int) (*Memo, error) {
	if base == nil || cap < 1 || cap > 512 || base.Dimension() < 1 || base.Dimension() > 4096 {
		return nil, errors.New("invalid memo configuration")
	}
	return &Memo{base: base, key: base.ModelKey(), dim: base.Dimension(), cap: cap, values: make(map[string][]float32)}, nil
}
func (m *Memo) Dimension() int                    { return m.dim }
func (m *Memo) Name() string                      { return m.base.Name() }
func (m *Memo) ModelKey() string                  { return m.key }
func (m *Memo) Embed(s string) ([]float32, error) { return m.EmbedDocument(s) }
func (m *Memo) EmbedDocument(s string) ([]float32, error) {
	return m.EmbedDocumentContext(context.Background(), s)
}
func (m *Memo) EmbedQuery(s string) ([]float32, error) {
	return m.EmbedQueryContext(context.Background(), s)
}
func (m *Memo) EmbedDocumentContext(ctx context.Context, s string) ([]float32, error) {
	return m.get(ctx, "d", s)
}
func (m *Memo) EmbedQueryContext(ctx context.Context, s string) ([]float32, error) {
	return m.get(ctx, "q", s)
}
func (m *Memo) Stats() MemoStats { m.mu.Lock(); defer m.mu.Unlock(); return m.stats }
func (m *Memo) get(ctx context.Context, role, s string) ([]float32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(s) > 16384 || m.base.ModelKey() != m.key || m.base.Dimension() != m.dim {
		return nil, errors.New("embedding identity changed")
	}
	key := role + "\x00" + s
	if v, ok := m.values[key]; ok {
		m.stats.Hits++
		return append([]float32(nil), v...), nil
	}
	if len(m.values) >= m.cap {
		return nil, errors.New("embedding memo capacity exhausted")
	}
	start := time.Now()
	var v []float32
	var err error
	if role == "d" {
		v, err = embed.DocumentContext(ctx, m.base, s)
	} else {
		v, err = embed.QueryContext(ctx, m.base, s)
	}
	m.stats.AcquisitionNS += time.Since(start).Nanoseconds()
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if m.base.ModelKey() != m.key || m.base.Dimension() != m.dim || len(v) != m.dim {
		return nil, errors.New("embedding dimension changed")
	}
	for _, value := range v {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return nil, errors.New("nonfinite embedding")
		}
	}
	m.values[key] = append([]float32(nil), v...)
	m.stats.Misses++
	m.stats.Entries = len(m.values)
	return append([]float32(nil), v...), nil
}
