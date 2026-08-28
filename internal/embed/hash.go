package embed

import (
	"errors"
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

type Embedder interface {
	Embed(text string) ([]float32, error)
	Dimension() int
	Name() string
}

// HashEmbedder is a deterministic, dependency-free development fallback. It
// preserves exact-token signal but is not a semantic embedding model.
type HashEmbedder struct {
	dimension int
}

func NewHashEmbedder(dimension int) (*HashEmbedder, error) {
	if dimension <= 0 {
		return nil, errors.New("embedding dimension must be positive")
	}
	return &HashEmbedder{dimension: dimension}, nil
}

func (h *HashEmbedder) Dimension() int { return h.dimension }
func (h *HashEmbedder) Name() string   { return "feature-hash-v1" }

func (h *HashEmbedder) Embed(text string) ([]float32, error) {
	vector := make([]float32, h.dimension)
	for _, token := range tokenize(text) {
		hash := fnv.New64a()
		_, _ = hash.Write([]byte(token))
		value := hash.Sum64()
		index := int(value % uint64(h.dimension))
		sign := float32(1)
		if value&(1<<63) != 0 {
			sign = -1
		}
		vector[index] += sign
	}
	var norm float64
	for _, value := range vector {
		norm += float64(value * value)
	}
	if norm == 0 {
		return vector, nil
	}
	scale := float32(1 / math.Sqrt(norm))
	for index := range vector {
		vector[index] *= scale
	}
	return vector, nil
}

func tokenize(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != '-' && r != '.' && r != '/'
	})
}
