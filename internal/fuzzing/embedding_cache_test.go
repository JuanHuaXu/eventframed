package fuzzing_test

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/fuzzing"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

type countingEmbedder struct {
	queries   atomic.Int64
	documents atomic.Int64
}

func (*countingEmbedder) Dimension() int { return 2 }
func (*countingEmbedder) Name() string   { return "counting" }
func (*countingEmbedder) ModelKey() string {
	return "counting:d2:repr=" + model.SemanticRepresentationVersion
}
func (e *countingEmbedder) Embed(text string) ([]float32, error) {
	return e.EmbedDocument(text)
}
func (e *countingEmbedder) EmbedQuery(string) ([]float32, error) {
	e.queries.Add(1)
	return []float32{1, 0}, nil
}
func (e *countingEmbedder) EmbedDocument(text string) ([]float32, error) {
	e.documents.Add(1)
	if len(text)%2 == 0 {
		return []float32{1, 0}, nil
	}
	return []float32{0, 1}, nil
}

func TestNominationCacheReusesQueryAndStoredEventScores(t *testing.T) {
	embedder := &countingEmbedder{}
	cache := fuzzing.NewNominationCache(4, 16)
	events := testEvents()
	for i := range events {
		events[i].EmbeddingModel = embedder.ModelKey()
		events[i].Embedding = []float32{float32(1 - i), float32(i)}
	}
	var first []float64
	for run := 0; run < 2; run++ {
		predictor, err := fuzzing.NewEmbeddingNominationPredictorWithCache(context.Background(), embedder, "same query", cache)
		if err != nil {
			t.Fatal(err)
		}
		law, err := predictor.Predict(context.Background(), events)
		if err != nil {
			t.Fatal(err)
		}
		if run == 0 {
			first = law
		} else if math.Abs(first[0]-law[0]) > 1e-15 || math.Abs(first[1]-law[1]) > 1e-15 {
			t.Fatalf("cached law changed: %v -> %v", first, law)
		}
	}
	if embedder.queries.Load() != 1 || embedder.documents.Load() != 0 {
		t.Fatalf("embedding calls: query=%d document=%d", embedder.queries.Load(), embedder.documents.Load())
	}
}

func TestNominationCacheCollapsesConcurrentMisses(t *testing.T) {
	embedder := &countingEmbedder{}
	cache := fuzzing.NewNominationCache(4, 16)
	events := testEvents()
	for i := range events {
		events[i].EmbeddingModel = embedder.ModelKey()
		events[i].Embedding = []float32{float32(1 - i), float32(i)}
	}
	start := make(chan struct{})
	var group sync.WaitGroup
	for worker := 0; worker < 32; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			predictor, err := fuzzing.NewEmbeddingNominationPredictorWithCache(context.Background(), embedder, "concurrent query", cache)
			if err != nil {
				t.Errorf("construct predictor: %v", err)
				return
			}
			if _, err := predictor.Predict(context.Background(), events); err != nil {
				t.Errorf("predict: %v", err)
			}
		}()
	}
	close(start)
	group.Wait()
	if embedder.queries.Load() != 1 || embedder.documents.Load() != 0 {
		t.Fatalf("concurrent embedding calls: query=%d document=%d", embedder.queries.Load(), embedder.documents.Load())
	}
}

func TestNominationCacheDoesNotReuseScoreAfterSemanticChange(t *testing.T) {
	embedder := &countingEmbedder{}
	cache := fuzzing.NewNominationCache(4, 16)
	events := testEvents()
	predictor, err := fuzzing.NewEmbeddingNominationPredictorWithCache(context.Background(), embedder, "same query", cache)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := predictor.Predict(context.Background(), events); err != nil {
		t.Fatal(err)
	}
	firstCalls := embedder.documents.Load()
	events[0].What.Value = "semantically changed value"
	if _, err := predictor.Predict(context.Background(), events); err != nil {
		t.Fatal(err)
	}
	if embedder.documents.Load() != firstCalls+1 {
		t.Fatalf("changed EventFrame reused stale score: before=%d after=%d", firstCalls, embedder.documents.Load())
	}
}

func TestNominationCacheSeparatesSnapshotNamespaces(t *testing.T) {
	embedder := &countingEmbedder{}
	cache := fuzzing.NewNominationCache(4, 16)
	events := testEvents()
	for i := range events {
		events[i].EmbeddingModel = embedder.ModelKey()
		events[i].Embedding = []float32{float32(1 - i), float32(i)}
	}
	first, err := fuzzing.NewEmbeddingNominationPredictorWithCacheNamespace(context.Background(), embedder, "same query", "snapshot-1", cache)
	if err != nil {
		t.Fatal(err)
	}
	firstLaw, err := first.Predict(context.Background(), events)
	if err != nil {
		t.Fatal(err)
	}
	events[0].Embedding, events[1].Embedding = events[1].Embedding, events[0].Embedding
	second, err := fuzzing.NewEmbeddingNominationPredictorWithCacheNamespace(context.Background(), embedder, "same query", "snapshot-2", cache)
	if err != nil {
		t.Fatal(err)
	}
	secondLaw, err := second.Predict(context.Background(), events)
	if err != nil {
		t.Fatal(err)
	}
	if !(firstLaw[0] > firstLaw[1] && secondLaw[0] < secondLaw[1]) {
		t.Fatalf("snapshot namespace reused stale scores: first=%v second=%v", firstLaw, secondLaw)
	}
}

func TestMismatchedStoredEmbeddingFallsBackToDocumentEmbedding(t *testing.T) {
	embedder := &countingEmbedder{}
	events := testEvents()
	events[0].EmbeddingModel = "other-model:d2"
	events[0].Embedding = []float32{1, 0}
	predictor, err := fuzzing.NewEmbeddingNominationPredictor(context.Background(), embedder, "query")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := predictor.Predict(context.Background(), events[:1]); err != nil {
		t.Fatal(err)
	}
	if embedder.documents.Load() != 1 {
		t.Fatalf("mismatched vector did not fall back: document calls=%d", embedder.documents.Load())
	}
}
