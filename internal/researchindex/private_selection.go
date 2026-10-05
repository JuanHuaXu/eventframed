package researchindex

import (
	"context"
	"errors"
	"math"
	"slices"
)

type NeighborCandidate struct {
	Ordinal  uint32
	Distance float32
}

// SelectPrivateNeighbors mirrors the valid-record branch of the backend's
// selection rule. Pair distance must use the same immutable metric as candidate
// distance. Missing/invalid records must be rejected by the caller, not invented.
func SelectPrivateNeighbors(ctx context.Context, input []NeighborCandidate, maxM, level, maxPairs int, alphaSquared float32, pair func(uint32, uint32) (float32, error)) ([]NeighborCandidate, int, error) {
	if maxM < 1 || maxM > 128 || level < 0 || level >= 32 || len(input) > 4096 || maxPairs < 0 || pair == nil || alphaSquared < 1 || math.IsNaN(float64(alphaSquared)) || math.IsInf(float64(alphaSquared), 0) {
		return nil, 0, ErrCapacity
	}
	if e := ctx.Err(); e != nil {
		return nil, 0, e
	}
	seen := map[uint32]bool{}
	for _, c := range input {
		if seen[c.Ordinal] || c.Distance < 0 || math.IsNaN(float64(c.Distance)) || math.IsInf(float64(c.Distance), 0) {
			return nil, 0, errors.New("invalid selection candidate")
		}
		seen[c.Ordinal] = true
	}
	candidates := slices.Clone(input)
	if len(candidates) <= maxM {
		return candidates, 0, nil
	}
	slices.SortFunc(candidates, func(a, b NeighborCandidate) int {
		if a.Distance < b.Distance {
			return -1
		}
		if a.Distance > b.Distance {
			return 1
		}
		if a.Ordinal < b.Ordinal {
			return -1
		}
		if a.Ordinal > b.Ordinal {
			return 1
		}
		return 0
	})
	limit := maxM * 4
	if level > 0 {
		limit = maxM * 2
	}
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	selected := []NeighborCandidate{candidates[0]}
	picked := make([]bool, len(candidates))
	picked[0] = true
	calls := 0
	for i := 1; i < len(candidates) && len(selected) < maxM; i++ {
		if e := ctx.Err(); e != nil {
			return nil, calls, e
		}
		c := candidates[i]
		reject := false
		for _, s := range selected {
			if calls >= maxPairs {
				return nil, calls, ErrCapacity
			}
			d, e := pair(c.Ordinal, s.Ordinal)
			calls++
			if e != nil {
				return nil, calls, e
			}
			if d < 0 || math.IsNaN(float64(d)) || math.IsInf(float64(d), 0) {
				return nil, calls, errors.New("invalid pair distance")
			}
			if d < c.Distance/alphaSquared {
				reject = true
				break
			}
		}
		if !reject {
			selected = append(selected, c)
			picked[i] = true
		}
	}
	for i := 1; i < len(candidates) && len(selected) < maxM; i++ {
		if !picked[i] {
			selected = append(selected, candidates[i])
			picked[i] = true
		}
	}
	if e := ctx.Err(); e != nil {
		return nil, calls, e
	}
	return selected, calls, nil
}
