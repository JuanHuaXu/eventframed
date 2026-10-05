package researchindex

import (
	"container/heap"
	"context"
	"errors"
	"math"
	"slices"
)

type layerSearchItem struct {
	id       uint32
	distance float64
}
type layerSearchQueue []layerSearchItem

func (q layerSearchQueue) Len() int           { return len(q) }
func (q layerSearchQueue) Less(i, j int) bool { return compareLayerItem(q[i], q[j]) < 0 }
func (q layerSearchQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *layerSearchQueue) Push(x any)        { *q = append(*q, x.(layerSearchItem)) }
func (q *layerSearchQueue) Pop() any          { a := *q; x := a[len(a)-1]; *q = a[:len(a)-1]; return x }
func compareLayerItem(a, b layerSearchItem) int {
	if a.distance < b.distance {
		return -1
	}
	if a.distance > b.distance {
		return 1
	}
	if a.id < b.id {
		return -1
	}
	if a.id > b.id {
		return 1
	}
	return 0
}

// SearchLayered is a research HNSW-style traversal over one immutable root, not
// an exact search or a byte-for-byte port of the backend search implementation.
// The evaluation budget spans all levels. Exhaustion returns no partial answer.
func SearchLayered(ctx context.Context, s LayeredSnapshot, query []float32, k, ef, maxEvaluations int) ([]Candidate, int, error) {
	if k < 1 || k > 200 {
		return nil, 0, ErrCapacity
	}
	best, calls, e := searchLayered(ctx, s, query, k, ef, maxEvaluations, nil, 0)
	if e != nil {
		return nil, calls, e
	}
	if best == nil {
		return nil, calls, nil
	}
	out := make([]Candidate, 0, len(best))
	for _, x := range best {
		r := layeredFind(s.tree, x.id)
		out = append(out, Candidate{ID: r.ID, Score: 1 - x.distance})
	}
	return out, calls, nil
}

// SearchConstructionLayer searches only the declared level from start. Its ef
// is literal: no public-query quality floor or descent is applied. Candidates
// are sorted by distance/ordinal for subsequent private neighbor selection.
// This serial research traversal does not include concurrent in-flight nodes.
func SearchConstructionLayer(ctx context.Context, s LayeredSnapshot, query []float32, start uint32, level, ef, maxEvaluations int) ([]NeighborCandidate, int, error) {
	if level < 0 || level > 31 {
		return nil, 0, ErrCapacity
	}
	best, calls, e := searchLayered(ctx, s, query, ef, ef, maxEvaluations, &start, level)
	if e != nil {
		return nil, calls, e
	}
	out := make([]NeighborCandidate, len(best))
	for i, x := range best {
		out[i] = NeighborCandidate{Ordinal: x.id, Distance: float32(x.distance)}
	}
	return out, calls, nil
}

func searchLayered(ctx context.Context, s LayeredSnapshot, query []float32, k, ef, maxEvaluations int, start *uint32, onlyLevel int) ([]layerSearchItem, int, error) {
	if k < 1 || ef < k || ef > 4096 || maxEvaluations < 1 || maxEvaluations > 1000000 {
		return nil, 0, ErrCapacity
	}
	if e := ctx.Err(); e != nil {
		return nil, 0, e
	}
	var norm float64
	for _, x := range query {
		if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
			return nil, 0, errors.New("nonfinite query")
		}
		norm += float64(x) * float64(x)
	}
	if norm == 0 {
		return nil, 0, errors.New("zero query")
	}
	if s.tree == nil {
		if start != nil {
			return nil, 0, errors.New("missing construction entry")
		}
		return nil, 0, nil
	}
	entry := uint32(s.global >> 32)
	if start != nil {
		entry = *start
	}
	n := layeredFind(s.tree, entry)
	if (start == nil && s.global == 0) || n == nil || len(n.Vector) != len(query) || n.Level < onlyLevel {
		return nil, 0, errors.New("invalid search entry or dimension")
	}
	calls := 0
	distance := func(id uint32) (layerSearchItem, error) {
		if e := ctx.Err(); e != nil {
			return layerSearchItem{}, e
		}
		if calls >= maxEvaluations {
			return layerSearchItem{}, ErrCapacity
		}
		r := layeredFind(s.tree, id)
		if r == nil || len(r.Vector) != len(query) {
			return layerSearchItem{}, errors.New("invalid search record")
		}
		calls++
		return layerSearchItem{id, max(0, 1-cosineWithNorms(query, r.Vector, norm, r.normSquared))}, nil
	}
	search := func(start uint32, level, width int) ([]layerSearchItem, error) {
		first, e := distance(start)
		if e != nil {
			return nil, e
		}
		queue := &layerSearchQueue{first}
		best := []layerSearchItem{first}
		seen := map[uint32]bool{start: true}
		for queue.Len() > 0 {
			if e := ctx.Err(); e != nil {
				return nil, e
			}
			next := heap.Pop(queue).(layerSearchItem)
			if len(best) >= width && compareLayerItem(next, best[len(best)-1]) > 0 {
				break
			}
			r := layeredFind(s.tree, next.id)
			if r == nil || r.Level < level {
				continue
			}
			// Backend traversal treats backlinks as navigable too. Deduplicate
			// through seen across both lists without allocating a combined list.
			for _, neighbors := range [][]uint32{r.Links[level], r.Backlinks[level]} {
				for _, id := range neighbors {
					if seen[id] {
						continue
					}
					seen[id] = true
					node := layeredFind(s.tree, id)
					if node == nil || node.Level < level {
						continue
					}
					candidate, e := distance(id)
					if e != nil {
						return nil, e
					}
					if len(best) < width || compareLayerItem(candidate, best[len(best)-1]) < 0 {
						heap.Push(queue, candidate)
						at, _ := slices.BinarySearchFunc(best, candidate, compareLayerItem)
						best = append(best, layerSearchItem{})
						copy(best[at+1:], best[at:])
						best[at] = candidate
						if len(best) > width {
							best = best[:width]
						}
					}
				}
			}
		}
		return best, nil
	}
	top := n.Level
	if start != nil {
		top = 0
	}
	for level := top; level > 0; level-- {
		best, e := search(entry, level, 1)
		if e != nil {
			return nil, calls, e
		}
		entry = best[0].id
	}
	best, e := search(entry, onlyLevel, ef)
	if e != nil {
		return nil, calls, e
	}
	if len(best) > k {
		best = best[:k]
	}
	return best, calls, nil
}
