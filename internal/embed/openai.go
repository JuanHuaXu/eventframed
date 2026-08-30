package embed

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

type OpenAICompatibleConfig struct {
	URL, Model, APIKey string
	Dimension          int
	Timeout            time.Duration
	DocumentPrefix     string
	QueryPrefix        string
}

type OpenAICompatible struct {
	config OpenAICompatibleConfig
	client *http.Client
}

func NewOpenAICompatible(config OpenAICompatibleConfig) (*OpenAICompatible, error) {
	if config.URL == "" || config.Model == "" || config.Dimension <= 0 {
		return nil, errors.New("embedding URL, model, and positive dimension are required")
	}
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	return &OpenAICompatible{config: config, client: &http.Client{Timeout: config.Timeout}}, nil
}

func (e *OpenAICompatible) Dimension() int { return e.config.Dimension }
func (e *OpenAICompatible) Name() string   { return "openai-compatible" }
func (e *OpenAICompatible) ModelKey() string {
	base := fmt.Sprintf("openai-compatible:%s:d%d", e.config.Model, e.config.Dimension)
	if e.config.DocumentPrefix == "" && e.config.QueryPrefix == "" {
		return BindRepresentation(base)
	}
	digest := sha256.Sum256([]byte(e.config.DocumentPrefix + "\x00" + e.config.QueryPrefix))
	return BindRepresentation(fmt.Sprintf("%s:rp%x", base, digest[:6]))
}

func (e *OpenAICompatible) Embed(text string) ([]float32, error) {
	return e.EmbedDocument(text)
}

func (e *OpenAICompatible) EmbedDocument(text string) ([]float32, error) {
	return e.embed(e.config.DocumentPrefix + text)
}

func (e *OpenAICompatible) EmbedQuery(text string) ([]float32, error) {
	return e.embed(e.config.QueryPrefix + text)
}

func (e *OpenAICompatible) embed(text string) ([]float32, error) {
	payload, _ := json.Marshal(map[string]any{"model": e.config.Model, "input": text, "encoding_format": "float"})
	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, e.config.URL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	if e.config.APIKey != "" {
		request.Header.Set("Authorization", "Bearer "+e.config.APIKey)
	}
	response, err := e.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("embedding request: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message := strings.TrimSpace(string(body))
		if len(message) > 4096 {
			message = message[:4096]
		}
		return nil, fmt.Errorf("embedding service status %d: %s", response.StatusCode, message)
	}
	var decoded struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("decode embedding: %w", err)
	}
	if len(decoded.Data) != 1 || len(decoded.Data[0].Embedding) != e.config.Dimension {
		return nil, fmt.Errorf("embedding service returned an unexpected vector shape")
	}
	return normalize(decoded.Data[0].Embedding), nil
}

func normalize(vector []float32) []float32 {
	var norm float64
	for _, value := range vector {
		norm += float64(value * value)
	}
	if norm == 0 {
		return vector
	}
	scale := float32(1 / math.Sqrt(norm))
	for index := range vector {
		vector[index] *= scale
	}
	return vector
}
