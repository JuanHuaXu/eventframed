package researchindex

import (
	"context"
	"errors"
)

// ResearchStoredVector reads the derived database by ID, independently of ANN
// nomination and the owned map. Presence here does not prove graph membership.
func (h *HNSWBase) ResearchStoredVector(ctx context.Context, id string) ([]float32, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return nil, errors.New("research HNSW base closed")
	}
	r, err := h.collection.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if r.ID != id {
		return nil, errors.New("stored identity mismatch")
	}
	return append([]float32(nil), r.Vector...), nil
}
