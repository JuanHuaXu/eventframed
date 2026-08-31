package embed_test

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
)

func TestOpenAICompatibleValidatesAndNormalizesVector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Error("missing API key")
		}
		_ = json.NewEncoder(writer).Encode(map[string]any{"data": []any{map[string]any{"embedding": []float32{3, 4}}}})
	}))
	defer server.Close()
	embedder, err := embed.NewOpenAICompatible(embed.OpenAICompatibleConfig{URL: server.URL, Model: "model-a", APIKey: "secret", Dimension: 2})
	if err != nil {
		t.Fatal(err)
	}
	vector, err := embedder.Embed("hello")
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(float64(vector[0]-0.6)) > 1e-6 || math.Abs(float64(vector[1]-0.8)) > 1e-6 {
		t.Fatalf("vector = %v", vector)
	}
	if embedder.ModelKey() != "openai-compatible:model-a:d2:repr=eventframe-5w1h-v1" {
		t.Fatalf("model key = %s", embedder.ModelKey())
	}
}

func TestOpenAICompatibleContextEmbeddingCancelsRequest(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
	}))
	defer server.Close()
	defer close(release)
	embedder, _ := embed.NewOpenAICompatible(embed.OpenAICompatibleConfig{URL: server.URL, Model: "model-a", Dimension: 2, Timeout: time.Minute})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := embed.QueryContext(ctx, embedder, "question")
		done <- err
	}()
	<-started
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled embedding request returned no error")
		}
	case <-time.After(time.Second):
		t.Fatal("canceled embedding request did not stop")
	}
}

func TestOpenAICompatibleRejectsWrongDimension(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"data":[{"embedding":[1]}]}`))
	}))
	defer server.Close()
	embedder, _ := embed.NewOpenAICompatible(embed.OpenAICompatibleConfig{URL: server.URL, Model: "model-a", Dimension: 2})
	if _, err := embedder.Embed("hello"); err == nil {
		t.Fatal("expected dimension error")
	}
}

func TestOpenAICompatibleAppliesRolePrefixesAndPinsThemInModelKey(t *testing.T) {
	var inputs []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var body struct {
			Input string `json:"input"`
		}
		_ = json.NewDecoder(request.Body).Decode(&body)
		inputs = append(inputs, body.Input)
		_ = json.NewEncoder(writer).Encode(map[string]any{"data": []any{map[string]any{"embedding": []float32{1, 0}}}})
	}))
	defer server.Close()
	embedder, err := embed.NewOpenAICompatible(embed.OpenAICompatibleConfig{
		URL: server.URL, Model: "model-a", Dimension: 2,
		DocumentPrefix: "search_document: ", QueryPrefix: "search_query: ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := embed.Document(embedder, "stored"); err != nil {
		t.Fatal(err)
	}
	if _, err := embed.Query(embedder, "question"); err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 2 || inputs[0] != "search_document: stored" || inputs[1] != "search_query: question" {
		t.Fatalf("role inputs = %#v", inputs)
	}
	if embedder.ModelKey() == "openai-compatible:model-a:d2:repr=eventframe-5w1h-v1" {
		t.Fatal("role prefixes were absent from the embedding contract key")
	}
}
