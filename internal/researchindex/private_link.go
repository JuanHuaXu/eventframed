package researchindex

import (
	"context"
	"errors"
	"math"
	"slices"
)

// PrivateLinkDecision is staged data, not a published mutation. Callers must
// remove Dropped backlinks and add the new backlink only when Accepted.
type PrivateLinkDecision struct {
	Links, Dropped []uint32
	Heuristic      uint32
	Accepted       bool
	PairCalls      int
}

// DecidePrivateLink mirrors serial insertion's live-record link decision.
// Unlike the backend's stale-link compaction, all supplied IDs must be live;
// callers validate the immutable snapshot before providing the metric.
func DecidePrivateLink(ctx context.Context, target, newID uint32, original []uint32, heuristic uint32, maxM, capacity, level, maxPairs int, alpha float32, metric func(uint32, uint32) (float32, error)) (PrivateLinkDecision, error) {
	zero := PrivateLinkDecision{}
	if maxM < 1 || maxM > 128 || capacity < maxM || capacity > 256 || level < 0 || level > 31 || len(original) > capacity || maxPairs < 0 || metric == nil || alpha < 1 || math.IsNaN(float64(alpha)) || math.IsInf(float64(alpha), 0) || target == newID {
		return zero, ErrCapacity
	}
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	seen := map[uint32]bool{}
	for _, id := range original {
		if id == target || seen[id] {
			return zero, errors.New("invalid live link list")
		}
		seen[id] = true
	}
	out := PrivateLinkDecision{Links: slices.Clone(original), Heuristic: heuristic}
	if seen[newID] {
		return out, nil
	}
	distance := func(a, b uint32) (float32, error) {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if out.PairCalls >= maxPairs {
			return 0, ErrCapacity
		}
		out.PairCalls++
		d, err := metric(a, b)
		if err != nil {
			return 0, err
		}
		if d < 0 || math.IsNaN(float64(d)) || math.IsInf(float64(d), 0) {
			return 0, errors.New("invalid link distance")
		}
		return d, nil
	}
	// The backend computes this even for slack admission; count it explicitly.
	nd, err := distance(target, newID)
	if err != nil {
		return zero, err
	}
	if len(original) < capacity {
		out.Links = append(out.Links, newID)
		out.Heuristic = 0
		out.Accepted = true
		return out, nil
	}
	var selected []uint32
	if int(heuristic) >= len(original) && len(original) > 0 {
		worst, err := distance(target, original[len(original)-1])
		if err != nil {
			return zero, err
		}
		if nd >= worst {
			return out, nil
		}
		inserted, accepted := false, false
		tryNew := func() (bool, error) {
			for _, id := range selected {
				d, err := distance(newID, id)
				if err != nil {
					return false, err
				}
				if d < nd/alpha {
					inserted = true
					return false, nil
				}
			}
			selected = append(selected, newID)
			inserted = true
			accepted = true
			return true, nil
		}
		for _, id := range original {
			d, err := distance(target, id)
			if err != nil {
				return zero, err
			}
			if !inserted && (nd < d || (nd == d && newID < id)) {
				ok, err := tryNew()
				if err != nil {
					return zero, err
				}
				if !ok {
					return out, nil
				}
				if len(selected) >= maxM {
					break
				}
			}
			if accepted {
				pair, err := distance(id, newID)
				if err != nil {
					return zero, err
				}
				if pair < d/alpha {
					continue
				}
			}
			selected = append(selected, id)
			if len(selected) >= maxM {
				break
			}
		}
		if !inserted && len(selected) < maxM {
			if _, err := tryNew(); err != nil {
				return zero, err
			}
		}
		if !accepted {
			return out, nil
		}
		out.Accepted = true
	} else {
		candidates := make([]NeighborCandidate, 0, len(original)+1)
		for _, id := range original {
			d, err := distance(target, id)
			if err != nil {
				return zero, err
			}
			candidates = append(candidates, NeighborCandidate{id, d})
		}
		candidates = append(candidates, NeighborCandidate{newID, nd})
		kept, _, err := SelectPrivateNeighbors(ctx, candidates, maxM, level, maxPairs-out.PairCalls, alpha, distance)
		if err != nil {
			return zero, err
		}
		for _, c := range kept {
			selected = append(selected, c.Ordinal)
			if c.Ordinal == newID {
				out.Accepted = true
			}
		}
	}
	out.Links = selected
	out.Heuristic = uint32(len(selected))
	for _, id := range original {
		if !slices.Contains(selected, id) {
			out.Dropped = append(out.Dropped, id)
		}
	}
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	return out, nil
}
