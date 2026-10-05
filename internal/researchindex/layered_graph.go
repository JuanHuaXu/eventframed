package researchindex

import (
	"context"
	"errors"
	"math"
	"slices"
)

// LayeredRecord is an owned input/output value. Internal records are immutable;
// no internal slice is returned to callers. This is not an HNSW update algorithm.
type LayeredRecord struct {
	ID               string
	Level            int
	Vector           []float32
	Links, Backlinks [][]uint32
	Heuristic        []uint32
	normSquared      float64 // derived by PrepareLayered, never trusted from input
}
type LayeredEdit struct {
	Ordinal uint32
	Record  *LayeredRecord
}
type layeredTrie struct {
	child  [2]*layeredTrie
	record *LayeredRecord
}
type LayeredSnapshot struct {
	tree   *layeredTrie
	global uint64
}
type LayeredLimits struct{ Edits, Dimension, Levels, Links int }
type LayeredStats struct{ Paths, Records, CopiedLinks, SharedLinks, CopiedVectorValues, SharedVectorValues int }

func layeredFind(t *layeredTrie, id uint32) *LayeredRecord {
	for bit := 31; bit >= 0 && t != nil; bit-- {
		t = t.child[(id>>bit)&1]
	}
	if t == nil {
		return nil
	}
	return t.record
}
func layeredReplace(t *layeredTrie, id uint32, bit int, r *LayeredRecord, stats *LayeredStats) *layeredTrie {
	if t == nil && r == nil {
		return nil
	}
	if bit < 0 {
		if r == nil {
			return nil
		}
		stats.Paths++
		return &layeredTrie{record: r}
	}
	i := (id >> bit) & 1
	var child [2]*layeredTrie
	if t != nil {
		child = t.child
	}
	replacement := layeredReplace(child[i], id, bit-1, r, stats)
	if replacement == child[i] {
		return t
	}
	child[i] = replacement
	// Prune only the new path. Historical roots still own their original nodes.
	if child[0] == nil && child[1] == nil {
		return nil
	}
	n := &layeredTrie{child: child}
	stats.Paths++
	return n
}
func copyLayers(in [][]uint32) [][]uint32 {
	if in == nil {
		return nil
	}
	out := make([][]uint32, len(in))
	for i := range in {
		out[i] = slices.Clone(in[i])
	}
	return out
}
func (s LayeredSnapshot) Global() uint64 { return s.global }
func (s LayeredSnapshot) Lookup(id uint32) (LayeredRecord, bool) {
	r := layeredFind(s.tree, id)
	if r == nil {
		return LayeredRecord{}, false
	}
	c := *r
	c.normSquared = 0 // keep derived internal state out of editable values
	c.Vector = slices.Clone(r.Vector)
	c.Links = copyLayers(r.Links)
	c.Backlinks = copyLayers(r.Backlinks)
	c.Heuristic = slices.Clone(r.Heuristic)
	return c, true
}
func shareLayers(in, old [][]uint32, stats *LayeredStats) [][]uint32 {
	if in == nil {
		return nil
	}
	out := make([][]uint32, len(in))
	for i := range in {
		if i < len(old) && (in[i] == nil) == (old[i] == nil) && slices.Equal(in[i], old[i]) {
			out[i] = old[i]
			stats.SharedLinks += len(in[i])
		} else {
			out[i] = slices.Clone(in[i])
			stats.CopiedLinks += len(in[i])
		}
	}
	return out
}
func sameVector(a, b []float32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Float32bits(a[i]) != math.Float32bits(b[i]) {
			return false
		}
	}
	return true
}

// PrepareLayered privately path-copies supplied logical edits, sharing unchanged
// adjacency levels and vectors. It does not discover edits, publish a root,
// serialize writers, persist data, repair edges, or bound retained history.
// Limits reject the entire unpublished result; they never truncate adjacency.
func PrepareLayered(ctx context.Context, before LayeredSnapshot, edits []LayeredEdit, global uint64, limits LayeredLimits) (LayeredSnapshot, LayeredStats, error) {
	var stats LayeredStats
	fail := func(err error) (LayeredSnapshot, LayeredStats, error) { return LayeredSnapshot{}, stats, err }
	if limits.Edits < 1 || limits.Dimension < 1 || limits.Dimension > 4096 || limits.Levels < 1 || limits.Levels > 32 || limits.Links < 1 || len(edits) > limits.Edits {
		return fail(ErrCapacity)
	}
	tree := before.tree
	seen := map[uint32]bool{}
	for _, e := range edits {
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		if seen[e.Ordinal] {
			return fail(errors.New("duplicate layered edit"))
		}
		seen[e.Ordinal] = true
		var record *LayeredRecord
		if r := e.Record; r != nil {
			if r.Level < 0 || r.Level >= limits.Levels || len(r.Links) != r.Level+1 || len(r.Backlinks) != r.Level+1 || len(r.Heuristic) != r.Level+1 || len(r.Vector) != limits.Dimension {
				return fail(errors.New("layered record shape"))
			}
			count := 0
			for i := range r.Links {
				// Subtraction avoids overflow in a caller-supplied aggregate length.
				if len(r.Links[i]) > limits.Links-count {
					return fail(ErrCapacity)
				}
				count += len(r.Links[i])
				if len(r.Backlinks[i]) > limits.Links-count {
					return fail(ErrCapacity)
				}
				count += len(r.Backlinks[i])
			}
			var norm float64
			for _, x := range r.Vector {
				if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
					return fail(errors.New("nonfinite layered vector"))
				}
				norm += float64(x) * float64(x)
			}
			if norm == 0 {
				return fail(errors.New("zero layered vector"))
			}
			old := layeredFind(before.tree, e.Ordinal)
			if old == nil {
				old = &LayeredRecord{}
			}
			c := *r
			c.normSquared = norm
			c.Links = shareLayers(r.Links, old.Links, &stats)
			c.Backlinks = shareLayers(r.Backlinks, old.Backlinks, &stats)
			c.Heuristic = slices.Clone(r.Heuristic)
			if sameVector(r.Vector, old.Vector) {
				c.Vector = old.Vector
				stats.SharedVectorValues += len(r.Vector)
			} else {
				c.Vector = slices.Clone(r.Vector)
				stats.CopiedVectorValues += len(r.Vector)
			}
			record = &c
			stats.Records++
		}
		tree = layeredReplace(tree, e.Ordinal, 31, record, &stats)
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	return LayeredSnapshot{tree: tree, global: global}, stats, nil
}

// Same dot-product and denominator arithmetic as cosine, with norms reused
// from immutable validated vectors. Callers supply matching nonzero norms.
func cosineWithNorms(a, b []float32, aa, bb float64) float64 {
	var dot float64
	for i, v := range a {
		dot += float64(v) * float64(b[i])
	}
	return dot / math.Sqrt(aa*bb)
}
