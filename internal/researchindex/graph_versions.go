package researchindex

import (
	"context"
	"errors"
	"math"
	"sync/atomic"
)

type GraphEdit struct {
	ID     uint32
	Vector []float32
	Edges  []uint32
	Delete bool
}
type graphVertex struct {
	vector []float32
	edges  []uint32
}
type graphTrie struct {
	child  [2]*graphTrie
	vertex *graphVertex
}
type graphRoot struct {
	tree     *graphTrie
	revision uint64
}

func graphFind(t *graphTrie, id uint32) *graphVertex {
	for bit := 31; bit >= 0 && t != nil; bit-- {
		t = t.child[(id>>bit)&1]
	}
	if t == nil {
		return nil
	}
	return t.vertex
}
func graphReplace(t *graphTrie, id uint32, bit int, v *graphVertex) *graphTrie {
	n := &graphTrie{}
	if t != nil {
		*n = *t
	}
	if bit < 0 {
		n.vertex = v
		return n
	}
	i := (id >> bit) & 1
	n.child[i] = graphReplace(n.child[i], id, bit-1, v)
	return n
}

// GraphVersions is a research ownership model, not HNSW. A persistent binary
// radix tree copies33 lookup nodes per edited32-bit ID, sharing untouched state.
// Deletion removes the vertex; traversal skips links to absent vertices. It does
// not repair ANN connectivity, maintain backlinks, or bound retained old roots.
type GraphVersions struct {
	current   atomic.Pointer[graphRoot]
	writer    chan struct{}
	dimension int
}
type GraphView struct{ root *graphRoot }
type PreparedGraph struct {
	owner    *GraphVersions
	next     *graphRoot
	finished atomic.Bool
}

func NewGraphVersions(dimension int) (*GraphVersions, error) {
	if dimension < 1 || dimension > 4096 {
		return nil, errors.New("invalid graph dimension")
	}
	g := &GraphVersions{writer: make(chan struct{}, 1), dimension: dimension}
	g.writer <- struct{}{}
	g.current.Store(&graphRoot{})
	return g, nil
}
func (g *GraphVersions) View() GraphView { return GraphView{g.current.Load()} }
func (v GraphView) Revision() uint64     { return v.root.revision }
func (v GraphView) Lookup(id uint32) ([]float32, []uint32, bool) {
	n := graphFind(v.root.tree, id)
	if n == nil {
		return nil, nil, false
	}
	return append([]float32(nil), n.vector...), append([]uint32(nil), n.edges...), true
}
func (g *GraphVersions) Prepare(ctx context.Context, edits []GraphEdit) (*PreparedGraph, error) {
	if len(edits) < 1 || len(edits) > 128 {
		return nil, ErrCapacity
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-g.writer:
	}
	prepared := false
	defer func() {
		if !prepared {
			g.writer <- struct{}{}
		}
	}()
	before := g.current.Load()
	if before.revision == math.MaxUint64 {
		return nil, errors.New("graph revision overflow")
	}
	tree := before.tree
	seen := map[uint32]bool{}
	for _, e := range edits {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if seen[e.ID] {
			return nil, errors.New("duplicate graph edit")
		}
		seen[e.ID] = true
		var n *graphVertex
		if !e.Delete {
			if len(e.Vector) != g.dimension || len(e.Edges) > 64 {
				return nil, errors.New("graph edit bounds")
			}
			var norm float64
			for _, x := range e.Vector {
				if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
					return nil, errors.New("nonfinite graph vector")
				}
				norm += float64(x) * float64(x)
			}
			if norm == 0 {
				return nil, errors.New("zero graph vector")
			}
			links := map[uint32]bool{}
			for _, id := range e.Edges {
				if id == e.ID || links[id] {
					return nil, errors.New("invalid graph edge")
				}
				links[id] = true
			}
			n = &graphVertex{append([]float32(nil), e.Vector...), append([]uint32(nil), e.Edges...)}
		}
		tree = graphReplace(tree, e.ID, 31, n)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p := &PreparedGraph{owner: g, next: &graphRoot{tree: tree, revision: before.revision + 1}}
	prepared = true
	return p, nil
}

// Commit publishes prepared memory only. Persistence must be resolved by an
// external owner before this call; uncertain durability is not handled here.
func (p *PreparedGraph) Commit() error {
	if !p.finished.CompareAndSwap(false, true) {
		return ErrFinished
	}
	p.owner.current.Store(p.next)
	p.owner.writer <- struct{}{}
	return nil
}
func (p *PreparedGraph) Abort() error {
	if !p.finished.CompareAndSwap(false, true) {
		return ErrFinished
	}
	p.owner.writer <- struct{}{}
	return nil
}

// Walk is bounded breadth-first traversal of one immutable root. This diagnostic
// is not nearest-neighbor search. Missing/tombstoned targets consume examination
// budget but are not returned, and cannot expose a later root's vertices.
func (v GraphView) Walk(start uint32, budget int) ([]uint32, error) {
	if budget < 1 || budget > 4096 {
		return nil, errors.New("walk budget")
	}
	queue := []uint32{start}
	seen := map[uint32]bool{start: true}
	var out []uint32
	for i := 0; i < len(queue) && i < budget; i++ {
		n := graphFind(v.root.tree, queue[i])
		if n == nil {
			continue
		}
		out = append(out, queue[i])
		for _, id := range n.edges {
			if !seen[id] && len(queue) < budget {
				seen[id] = true
				queue = append(queue, id)
			}
		}
	}
	return out, nil
}
