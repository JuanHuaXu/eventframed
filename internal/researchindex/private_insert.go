package researchindex

import (
	"context"
	"errors"
	"math"
)

type PrivateInsertion struct {
	Snapshot               LayeredSnapshot
	Summary                EntrySummary
	Evaluations, PairCalls int
}

// PreparePrivateInsertion is a serial graph-only prototype. count, summary and
// before must describe the same live snapshot; ordinal/ID uniqueness is owned by
// the caller. Level is supplied, not sampled here. No durable state is changed.
// maxEdits bounds each level's connection stage, not distinct edits across all
// levels. Evaluation/pair budgets span the insertion; total graph reads and
// retained bytes are not certified by these counters.
func PreparePrivateInsertion(ctx context.Context, before LayeredSnapshot, summary EntrySummary, ordinal uint32, id string, vector []float32, level, count, m, ef, maxEvaluations, maxPairs, maxEdits int, alpha float32) (PrivateInsertion, error) {
	fail := func(e error) (PrivateInsertion, error) { return PrivateInsertion{}, e }
	if level < 0 || level > 31 || count < 0 || m < 1 || m > 56 || ef < 1 || ef > 4096 || maxEvaluations < 1 || maxPairs < 0 || maxEdits < 1 || maxEdits > 128 || alpha < 1 || math.IsNaN(float64(alpha)) || math.IsInf(float64(alpha), 0) {
		return fail(ErrCapacity)
	}
	if layeredFind(before.tree, ordinal) != nil {
		return fail(errors.New("ordinal already present"))
	}
	if (before.tree == nil) != (count == 0) {
		return fail(errors.New("count/root disagreement"))
	}
	n := &LayeredRecord{ID: id, Level: level, Vector: vector, Links: make([][]uint32, level+1), Backlinks: make([][]uint32, level+1), Heuristic: make([]uint32, level+1)}
	current, _, e := PrepareLayered(ctx, before, []LayeredEdit{{ordinal, n}}, before.global, LayeredLimits{1, len(vector), 32, 16384})
	if e != nil {
		return fail(e)
	}
	nextSummary, e := summary.With(ordinal, level)
	if e != nil {
		return fail(e)
	}
	out := PrivateInsertion{Snapshot: current, Summary: nextSummary}
	if count == 0 {
		out.Snapshot.global = uint64(ordinal)<<32 | uint64(level+1)
		return out, nil
	}
	entry := uint32(before.global >> 32)
	root := layeredFind(before.tree, entry)
	if before.global == 0 || root == nil {
		return fail(errors.New("invalid insertion entry"))
	}
	if count == 1 {
		// Backend's second-node shortcut creates only the level-zero pair and
		// does not mark either adjacency as diversity-checked.
		a, _ := current.Lookup(ordinal)
		b, _ := current.Lookup(entry)
		a.Links[0] = append(a.Links[0], entry)
		a.Backlinks[0] = append(a.Backlinks[0], entry)
		b.Links[0] = append(b.Links[0], ordinal)
		b.Backlinks[0] = append(b.Backlinks[0], ordinal)
		b.Heuristic[0] = 0
		current, _, e = PrepareLayered(ctx, current, []LayeredEdit{{ordinal, &a}, {entry, &b}}, current.global, LayeredLimits{maxEdits, len(vector), 32, 16384})
		if e != nil {
			return fail(e)
		}
	} else {
		search := func(l, width int) ([]NeighborCandidate, error) {
			got, calls, err := SearchConstructionLayer(ctx, current, vector, entry, l, width, maxEvaluations-out.Evaluations)
			out.Evaluations += calls
			return got, err
		}
		for l := root.Level; l > level; l-- {
			got, err := search(l, 1)
			if err != nil {
				return fail(err)
			}
			entry = got[0].Ordinal
		}
		for l := min(level, root.Level); l >= 0; l-- {
			candidates, err := search(l, ef)
			if err != nil {
				return fail(err)
			}
			maxM := m
			physical := m
			if l == 0 {
				maxM = int(float64(m) * 2.25)
				physical = 2 * m
			}
			physical += max(4, physical/4)
			metric := func(a, b uint32) (float32, error) {
				x, y := layeredFind(current.tree, a), layeredFind(current.tree, b)
				if x == nil || y == nil {
					return 0, errors.New("missing insertion metric node")
				}
				return float32(max(0, 1-cosineWithNorms(x.Vector, y.Vector, x.normSquared, y.normSquared))), nil
			}
			selected, calls, err := SelectPrivateNeighbors(ctx, candidates, maxM, l, maxPairs-out.PairCalls, alpha, metric)
			out.PairCalls += calls
			if err != nil {
				return fail(err)
			}
			ids := make([]uint32, len(selected))
			for i, c := range selected {
				ids[i] = c.Ordinal
			}
			connected, err := PreparePrivateConnections(ctx, current, ordinal, ids, l, maxM, physical, maxEdits, maxPairs-out.PairCalls, alpha)
			if err != nil {
				return fail(err)
			}
			out.PairCalls += connected.PairCalls
			current = connected.Snapshot
			if len(selected) > 0 {
				entry = selected[0].Ordinal
			}
		}
	}
	if level > root.Level {
		current.global = uint64(ordinal)<<32 | uint64(level+1)
	}
	if e := ctx.Err(); e != nil {
		return fail(e)
	}
	out.Snapshot = current
	return out, nil
}
