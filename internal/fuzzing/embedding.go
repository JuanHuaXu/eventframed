package fuzzing

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"golang.org/x/sync/singleflight"
)

type EmbeddingNominationPredictor struct {
	embedder    embed.Embedder
	queryVector []float32
	queryKey    string
	namespace   string
	cache       *NominationCache
}

func NewEmbeddingNominationPredictor(ctx context.Context, embedder embed.Embedder, query string) (*EmbeddingNominationPredictor, error) {
	return NewEmbeddingNominationPredictorWithCache(ctx, embedder, query, nil)
}

// NominationCache is a bounded, concurrency-safe cache for immutable query
// vectors and raw event nomination scores. FIFO eviction is sufficient because
// cache membership affects cost only, never prediction semantics.
type NominationCache struct {
	mu         sync.Mutex
	maxQueries int
	maxScores  int
	queries    map[string][]float32
	queryOrder []string
	scores     map[nominationScoreKey]float64
	scoreOrder []nominationScoreKey
	queryGroup singleflight.Group
	scoreGroup singleflight.Group
}

type nominationScoreKey struct {
	queryKey  string
	namespace string
	tenantID  string
	identity  string
	stored    bool
}

func NewNominationCache(maxQueries, maxScores int) *NominationCache {
	if maxQueries <= 0 || maxScores <= 0 {
		return nil
	}
	return &NominationCache{maxQueries: maxQueries, maxScores: maxScores, queries: make(map[string][]float32), scores: make(map[nominationScoreKey]float64)}
}

func NewEmbeddingNominationPredictorWithCache(ctx context.Context, embedder embed.Embedder, query string, cache *NominationCache) (*EmbeddingNominationPredictor, error) {
	return NewEmbeddingNominationPredictorWithCacheNamespace(ctx, embedder, query, "", cache)
}

func NewEmbeddingNominationPredictorWithCacheNamespace(ctx context.Context, embedder embed.Embedder, query, namespace string, cache *NominationCache) (*EmbeddingNominationPredictor, error) {
	if embedder == nil {
		return nil, errors.New("embedder is required")
	}
	queryText := frame.QueryText(query)
	queryKey := digestKey(embedder.ModelKey(), queryText)
	queryVector, err := loadQueryVector(ctx, embedder, cache, queryKey, queryText)
	if err != nil {
		return nil, fmt.Errorf("embed fuzz query: %w", err)
	}
	return &EmbeddingNominationPredictor{embedder: embedder, queryVector: queryVector, queryKey: queryKey, namespace: namespace, cache: cache}, nil
}

func NewEmbeddingNominationPredictorFromVector(embedder embed.Embedder, queryVector []float32, queryKey, namespace string, cache *NominationCache) (*EmbeddingNominationPredictor, error) {
	// Background curiosity reuses the serving-time query vector. This avoids
	// retaining raw query text or paying a second embedding cost; the snapshot
	// namespace keeps cached nomination semantics tied to the captured corpus.
	if embedder == nil {
		return nil, errors.New("embedder is required")
	}
	if len(queryVector) != embedder.Dimension() || queryKey == "" {
		return nil, errors.New("fuzz query vector dimension and key must match the active embedder")
	}
	for _, value := range queryVector {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return nil, errors.New("fuzz query vector must be finite")
		}
	}
	return &EmbeddingNominationPredictor{embedder: embedder, queryVector: append([]float32(nil), queryVector...), queryKey: queryKey, namespace: namespace, cache: cache}, nil
}

func (predictor *EmbeddingNominationPredictor) Kind() string {
	return "embedding-nomination/" + predictor.embedder.ModelKey()
}

func (predictor *EmbeddingNominationPredictor) Predict(ctx context.Context, events []model.Event) ([]float64, error) {
	law := make([]float64, len(events))
	total := 0.0
	for index, event := range events {
		mass, err := predictor.rawMass(ctx, event)
		if err != nil {
			return nil, err
		}
		law[index] = mass
		total += mass
	}
	if total <= 0 || math.IsNaN(total) || math.IsInf(total, 0) {
		return nil, errors.New("embedding nomination produced invalid mass")
	}
	for index := range law {
		law[index] /= total
	}
	return law, nil
}

func loadQueryVector(ctx context.Context, embedder embed.Embedder, cache *NominationCache, key, text string) ([]float32, error) {
	if cache == nil {
		return embed.QueryContext(ctx, embedder, text)
	}
	if vector, ok := cacheQuery(cache, key); ok {
		return vector, nil
	}
	value, err, _ := cache.queryGroup.Do(key, func() (any, error) {
		if vector, ok := cacheQuery(cache, key); ok {
			return vector, nil
		}
		vector, embedErr := embed.QueryContext(ctx, embedder, text)
		if embedErr != nil {
			return nil, embedErr
		}
		storeQuery(cache, key, vector)
		return vector, nil
	})
	if err != nil {
		return nil, err
	}
	return value.([]float32), nil
}

func (predictor *EmbeddingNominationPredictor) rawMass(ctx context.Context, event model.Event) (float64, error) {
	stored := event.EmbeddingModel == predictor.embedder.ModelKey() && len(event.Embedding) == predictor.embedder.Dimension()
	if predictor.cache == nil {
		vector, err := predictor.eventVector(ctx, event, stored, "")
		if err != nil {
			return 0, err
		}
		return math.Max(1e-12, (cosine(predictor.queryVector, vector)+1)/2), nil
	}
	text := ""
	scoreKey := nominationScoreKey{queryKey: predictor.queryKey, namespace: predictor.namespace, tenantID: event.TenantID, identity: event.ID, stored: true}
	if !stored {
		text = event.FrameText()
		scoreKey = nominationScoreKey{queryKey: predictor.queryKey, namespace: predictor.namespace, identity: digestKey("text", text)}
	}
	if mass, ok := cacheScore(predictor.cache, scoreKey); ok {
		return mass, nil
	}
	flightKey := digestKey(scoreKey.queryKey, scoreKey.namespace, scoreKey.tenantID, scoreKey.identity)
	value, err, _ := predictor.cache.scoreGroup.Do(flightKey, func() (any, error) {
		if mass, ok := cacheScore(predictor.cache, scoreKey); ok {
			return mass, nil
		}
		vector, vectorErr := predictor.eventVector(ctx, event, stored, text)
		if vectorErr != nil {
			return nil, vectorErr
		}
		mass := math.Max(1e-12, (cosine(predictor.queryVector, vector)+1)/2)
		storeScore(predictor.cache, scoreKey, mass)
		return mass, nil
	})
	if err != nil {
		return 0, err
	}
	return value.(float64), nil
}

func digestKey(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(part))
	}
	return string(hash.Sum(nil))
}

func (predictor *EmbeddingNominationPredictor) eventVector(ctx context.Context, event model.Event, stored bool, text string) ([]float32, error) {
	if stored {
		return event.Embedding, nil
	}
	if text == "" {
		text = event.FrameText()
	}
	vector, err := embed.DocumentContext(ctx, predictor.embedder, text)
	if err != nil {
		return nil, fmt.Errorf("embed event %q: %w", event.ID, err)
	}
	return vector, nil
}

func cacheQuery(cache *NominationCache, key string) ([]float32, bool) {
	if cache == nil {
		return nil, false
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	vector, ok := cache.queries[key]
	return vector, ok
}

func storeQuery(cache *NominationCache, key string, vector []float32) {
	if cache == nil {
		return
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if _, exists := cache.queries[key]; exists {
		return
	}
	if len(cache.queryOrder) >= cache.maxQueries {
		delete(cache.queries, cache.queryOrder[0])
		cache.queryOrder = cache.queryOrder[1:]
	}
	cache.queries[key] = append([]float32(nil), vector...)
	cache.queryOrder = append(cache.queryOrder, key)
}

func cacheScore(cache *NominationCache, key nominationScoreKey) (float64, bool) {
	if cache == nil {
		return 0, false
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	score, ok := cache.scores[key]
	return score, ok
}

func storeScore(cache *NominationCache, key nominationScoreKey, score float64) {
	if cache == nil {
		return
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if _, exists := cache.scores[key]; exists {
		return
	}
	if len(cache.scoreOrder) >= cache.maxScores {
		delete(cache.scores, cache.scoreOrder[0])
		cache.scoreOrder = cache.scoreOrder[1:]
	}
	cache.scores[key] = score
	cache.scoreOrder = append(cache.scoreOrder, key)
}

func cosine(left, right []float32) float64 {
	if len(left) != len(right) || len(left) == 0 {
		return 0
	}
	dot, leftNorm, rightNorm := 0.0, 0.0, 0.0
	for index := range left {
		l, r := float64(left[index]), float64(right[index])
		dot += l * r
		leftNorm += l * l
		rightNorm += r * r
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return math.Max(-1, math.Min(1, dot/math.Sqrt(leftNorm*rightNorm)))
}
