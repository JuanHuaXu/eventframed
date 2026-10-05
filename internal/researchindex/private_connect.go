package researchindex

import (
	"context"
	"errors"
	"math"
	"slices"
)

type PrivateConnection struct {
	Snapshot           LayeredSnapshot
	Changed, PairCalls int
}

// PreparePrivateConnections composes one insertion level on an immutable root.
// The new node must already exist in before. All neighbors/links must be live;
// stale-link repair is deliberately not inferred here. No state is published.
func PreparePrivateConnections(ctx context.Context, before LayeredSnapshot, nodeID uint32, neighbors []uint32, level, maxM, capacity, maxEdits, maxPairs int, alpha float32) (PrivateConnection, error) {
	fail := func(err error) (PrivateConnection, error) { return PrivateConnection{}, err }
	if level < 0 || level > 31 || maxM < 1 || maxM > 128 || capacity < maxM || capacity > 256 || maxEdits < 1 || maxEdits > 128 || maxPairs < 0 || len(neighbors) > capacity || alpha < 1 || math.IsNaN(float64(alpha)) || math.IsInf(float64(alpha), 0) {
		return fail(ErrCapacity)
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	edits := map[uint32]*LayeredRecord{}
	edit := func(id uint32) (*LayeredRecord, error) {
		if r := edits[id]; r != nil {
			return r, nil
		}
		if len(edits) >= maxEdits {
			return nil, ErrCapacity
		}
		r, ok := before.Lookup(id)
		if !ok || r.Level < level {
			return nil, errors.New("connection node absent at level")
		}
		if len(r.Links[level]) > capacity || len(r.Backlinks[level]) > capacity {
			return nil, errors.New("connection capacity mismatch")
		}
		edits[id] = &r
		return &r, nil
	}
	node, err := edit(nodeID)
	if err != nil {
		return fail(err)
	}
	seen := map[uint32]bool{}
	for _, id := range neighbors {
		if id == nodeID || seen[id] {
			return fail(errors.New("invalid selected neighbors"))
		}
		seen[id] = true
		if _, err := edit(id); err != nil {
			return fail(err)
		}
	}
	appendBounded := func(list *[]uint32, id uint32) bool {
		if len(*list) >= capacity || slices.Contains(*list, id) {
			return false
		}
		*list = append(*list, id)
		return true
	}
	// Follow backend ordering: all forward links first, then reverse decisions.
	for _, id := range neighbors {
		appendBounded(&node.Links[level], id)
		appendBounded(&edits[id].Backlinks[level], nodeID)
	}
	node.Heuristic[level] = uint32(len(node.Links[level]))
	pairs := 0
	for _, id := range neighbors {
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		target := edits[id]
		for _, link := range target.Links[level] {
			r := layeredFind(before.tree, link)
			if r == nil || r.Level < level {
				return fail(errors.New("stale connection link"))
			}
		}
		metric := func(a, b uint32) (float32, error) {
			x, y := layeredFind(before.tree, a), layeredFind(before.tree, b)
			if x == nil || y == nil || len(x.Vector) != len(y.Vector) {
				return 0, errors.New("invalid connection metric record")
			}
			return float32(max(0, 1-cosineWithNorms(x.Vector, y.Vector, x.normSquared, y.normSquared))), nil
		}
		decision, err := DecidePrivateLink(ctx, id, nodeID, target.Links[level], target.Heuristic[level], maxM, capacity, level, maxPairs-pairs, alpha, metric)
		if err != nil {
			return fail(err)
		}
		pairs += decision.PairCalls
		target.Links[level] = decision.Links
		target.Heuristic[level] = decision.Heuristic
		for _, dropped := range decision.Dropped {
			r, err := edit(dropped)
			if err != nil {
				return fail(err)
			}
			for i, v := range r.Backlinks[level] {
				if v == id {
					r.Backlinks[level][i] = r.Backlinks[level][len(r.Backlinks[level])-1]
					r.Backlinks[level] = r.Backlinks[level][:len(r.Backlinks[level])-1]
					break
				}
			}
		}
		if decision.Accepted {
			appendBounded(&node.Backlinks[level], id)
		}
	}
	list := make([]LayeredEdit, 0, len(edits))
	for id, r := range edits {
		list = append(list, LayeredEdit{id, r})
	}
	slices.SortFunc(list, func(a, b LayeredEdit) int {
		if a.Ordinal < b.Ordinal {
			return -1
		}
		if a.Ordinal > b.Ordinal {
			return 1
		}
		return 0
	})
	after, _, err := PrepareLayered(ctx, before, list, before.global, LayeredLimits{maxEdits, len(node.Vector), 32, 16384})
	if err != nil {
		return fail(err)
	}
	return PrivateConnection{after, len(edits), pairs}, nil
}
