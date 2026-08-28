package embed_test

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

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
	if embedder.ModelKey() != "openai-compatible:model-a:d2" {
		t.Fatalf("model key = %s", embedder.ModelKey())
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
