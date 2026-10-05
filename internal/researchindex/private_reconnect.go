package researchindex

import (
	"context"
	"errors"
	"math"
	"slices"
)

type ReconnectWork struct {
	Edits     []LayeredEdit
	PairCalls int
}

// DiscoverReconnect privately applies one level's reconnection stage. Distance
// must implement the declared backend metric for these immutable ordinal values;
// its implementation cost is not bounded by the call-count limit. This function
// neither retires a target nor publishes a complete deletion transaction.
func DiscoverReconnect(ctx context.Context, before LayeredSnapshot, neighbors []uint32, level, m, maxNeighbors, maxEdits, maxPairs int, distance func(uint32, uint32) (float32, error)) (ReconnectWork, error) {
	fail := func(err error) (ReconnectWork, error) { return ReconnectWork{}, err }
	if level < 0 || level >= 32 || m < 1 || m > 128 || maxNeighbors < 1 || maxNeighbors > 1024 || len(neighbors) > maxNeighbors || maxEdits < 1 || maxEdits > 128 || maxPairs < 0 || distance == nil {
		return fail(ErrCapacity)
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	var valid []uint32
	seen := map[uint32]bool{}
	for _, id := range neighbors {
		if seen[id] {
			return fail(errors.New("duplicate reconnect neighbor"))
		}
		seen[id] = true
		if layeredFind(before.tree, id) != nil {
			valid = append(valid, id)
		}
	}
	d := len(valid)
	if d < 2 {
		return ReconnectWork{}, nil
	}
	maxM := m
	if level == 0 {
		maxM *= 2
	}
	capacity := maxM + max(4, maxM/4)
	edited := map[uint32]*LayeredRecord{}
	read := func(id uint32) *LayeredRecord {
		if r := edited[id]; r != nil {
			return r
		}
		return layeredFind(before.tree, id)
	}
	edit := func(id uint32) (*LayeredRecord, error) {
		if r := edited[id]; r != nil {
			return r, nil
		}
		if len(edited) >= maxEdits {
			return nil, ErrCapacity
		}
		r, ok := before.Lookup(id)
		if !ok {
			return nil, errors.New("reconnect node absent")
		}
		edited[id] = &r
		return &r, nil
	}
	appendEdge := func(from, to uint32) error {
		a, b := read(from), read(to)
		if a == nil || a.Level < level || slices.Contains(a.Links[level], to) || len(a.Links[level]) >= capacity {
			return nil
		}
		a, err := edit(from)
		if err != nil {
			return err
		}
		a.Links[level] = append(a.Links[level], to)
		a.Heuristic[level] = 0
		if b != nil && b.Level >= level && !slices.Contains(b.Backlinks[level], from) && len(b.Backlinks[level]) < capacity {
			b, err = edit(to)
			if err != nil {
				return err
			}
			b.Backlinks[level] = append(b.Backlinks[level], from)
		}
		return nil
	}
	var matrix []float32
	calls := 0
	type candidate struct {
		id   uint32
		dist float32
	}
	for i, id := range valid {
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		r := read(id)
		if r.Level < level || len(r.Links[level]) >= max(1, maxM/2) {
			continue
		}
		if matrix == nil {
			if d*(d-1)/2 > maxPairs {
				return fail(ErrCapacity)
			}
			matrix = make([]float32, d*d)
			for a := 0; a < d; a++ {
				for b := a + 1; b < d; b++ {
					if err := ctx.Err(); err != nil {
						return fail(err)
					}
					x, err := distance(valid[a], valid[b])
					calls++
					if err != nil {
						return fail(err)
					}
					if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
						return fail(errors.New("invalid reconnect distance"))
					}
					matrix[a*d+b] = x
					matrix[b*d+a] = x
				}
			}
		}
		var candidates []candidate
		for j, other := range valid {
			if i == j || slices.Contains(r.Links[level], other) {
				continue
			}
			candidates = append(candidates, candidate{other, matrix[i*d+j]})
		}
		selectN := min(len(candidates), maxM-len(r.Links[level]))
		// Match the source: preserve input order when every candidate is selected.
		if len(candidates) > selectN {
			slices.SortFunc(candidates, func(a, b candidate) int {
				if a.dist < b.dist {
					return -1
				}
				if a.dist > b.dist {
					return 1
				}
				return 0
			})
		}
		for _, c := range candidates[:selectN] {
			if err := appendEdge(id, c.id); err != nil {
				return fail(err)
			}
			if err := appendEdge(c.id, id); err != nil {
				return fail(err)
			}
		}
	}
	out := ReconnectWork{PairCalls: calls}
	for id, r := range edited {
		out.Edits = append(out.Edits, LayeredEdit{id, r})
	}
	slices.SortFunc(out.Edits, func(a, b LayeredEdit) int {
		if a.Ordinal < b.Ordinal {
			return -1
		}
		if a.Ordinal > b.Ordinal {
			return 1
		}
		return 0
	})
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	return out, nil
}
