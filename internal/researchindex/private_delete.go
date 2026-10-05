package researchindex

import (
	"context"
	"errors"
	"slices"
)

type PrivateDeletion struct {
	Snapshot           LayeredSnapshot
	Summary            EntrySummary
	Changed, PairCalls int
}

// PreparePrivateDeletion composes graph-only stages. Summary must describe the
// same live ordinals/levels as before. It does not publish or persist anything,
// and excludes ID maps, external vector-store ownership and semantic revisions.
// maxReads bounds each incoming stage, not all graph reads; maxPairs and the
// distinct edit cap apply to the entire deletion. Metric is declared cosine.
func PreparePrivateDeletion(ctx context.Context, before LayeredSnapshot, summary EntrySummary, target uint32, m, maxReads, maxNeighbors, maxEdits, maxPairs int) (PrivateDeletion, error) {
	fail := func(e error) (PrivateDeletion, error) { return PrivateDeletion{}, e }
	if maxEdits < 1 || maxEdits > 128 || maxPairs < 0 || m < 1 || m > 128 {
		return fail(ErrCapacity)
	}
	n := layeredFind(before.tree, target)
	if n == nil {
		return fail(errors.New("delete target absent"))
	}
	current := before
	changed := map[uint32]bool{}
	pairs := 0
	apply := func(edits []LayeredEdit) error {
		for _, e := range edits {
			changed[e.Ordinal] = true
		}
		if len(changed) > maxEdits {
			return ErrCapacity
		}
		next, _, e := PrepareLayered(ctx, current, edits, current.global, LayeredLimits{maxEdits, len(n.Vector), 32, 1024})
		if e == nil {
			current = next
		}
		return e
	}
	for level := 0; level <= n.Level; level++ {
		neighbors := slices.Clone(n.Links[level])
		incoming, e := discoverIncomingLevel(ctx, current, target, maxReads, maxEdits, level)
		if e != nil {
			return fail(e)
		}
		for _, id := range incoming.Affected[level] {
			if !slices.Contains(neighbors, id) {
				neighbors = append(neighbors, id)
			}
		}
		if e = apply(incoming.Edits); e != nil {
			return fail(e)
		}
		distance := func(a, b uint32) (float32, error) {
			x, y := layeredFind(current.tree, a), layeredFind(current.tree, b)
			if x == nil || y == nil {
				return 0, errors.New("distance node absent")
			}
			return float32(max(0, 1-cosineWithNorms(x.Vector, y.Vector, x.normSquared, y.normSquared))), nil
		}
		repair, e := DiscoverReconnect(ctx, current, neighbors, level, m, maxNeighbors, maxEdits, maxPairs-pairs, distance)
		if e != nil {
			return fail(e)
		}
		pairs += repair.PairCalls
		if e = apply(repair.Edits); e != nil {
			return fail(e)
		}
	}
	if e := apply([]LayeredEdit{{Ordinal: target}}); e != nil {
		return fail(e)
	}
	nextSummary, e := summary.With(target, -1)
	if e != nil {
		return fail(e)
	}
	if before.global != 0 && uint32(before.global>>32) == target {
		id, level, ok := nextSummary.Best()
		if !ok {
			current.global = 0
		} else {
			current.global = uint64(id)<<32 | uint64(level+1)
		}
	}
	if e := ctx.Err(); e != nil {
		return fail(e)
	}
	return PrivateDeletion{current, nextSummary, len(changed), pairs}, nil
}
