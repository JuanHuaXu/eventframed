package researchindex

import (
	"context"
	"errors"
	"math"
)

// ResearchNomination diagnoses ANN nomination on one unchanged graph. It omits
// delta merging and must not be used as a serving substitute. ef=0 selects the
// direct collection default; positive values use the public query override.
func (h *HNSWBase) ResearchNomination(ctx context.Context, query []float32, k, ef int) ([]Candidate, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return nil, errors.New("research HNSW base closed")
	}
	if len(query) != h.dimension || k < 1 || k > 200 || ef < 0 || ef > 6400 {
		return nil, errors.New("invalid diagnostic query")
	}
	var norm float64
	for _, x := range query {
		if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
			return nil, errors.New("invalid vector")
		}
		norm += float64(x) * float64(x)
	}
	if norm == 0 {
		return nil, errors.New("zero vector")
	}
	results, err := h.collection.Query(ctx).WithVector(query).Limit(k).WithEfSearch(ef).Execute()
	if err != nil {
		return nil, err
	}
	candidates := make([]Candidate, 0, len(results.Results))
	for _, r := range results.Results {
		owned, ok := h.base.records[r.ID]
		if !ok {
			return nil, errors.New("foreign base ID")
		}
		candidates = append(candidates, Candidate{r.ID, cosine(query, owned.vector)})
	}
	return candidates, nil
}
