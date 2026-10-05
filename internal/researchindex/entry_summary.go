package researchindex

import "errors"

type entryChoice struct {
	id    uint32
	level int
}
type entrySummaryNode struct {
	child [2]*entrySummaryNode
	best  entryChoice
}

// EntrySummary is a persistent selection index, not the graph's actual entry
// point. Query only when replacement is needed: insertion may retain an existing
// tied entry point. Every nonnil subtree contains at least one live candidate.
type EntrySummary struct{ root *entrySummaryNode }

func preferredEntry(a, b entryChoice) entryChoice {
	if b.level > a.level || b.level == a.level && b.id < a.id {
		return b
	}
	return a
}
func editEntry(n *entrySummaryNode, id uint32, level, bit int) *entrySummaryNode {
	if n == nil && level < 0 {
		return nil
	}
	if bit < 0 {
		if level < 0 {
			return nil
		}
		return &entrySummaryNode{best: entryChoice{id, level}}
	}
	var children [2]*entrySummaryNode
	if n != nil {
		children = n.child
	}
	i := (id >> bit) & 1
	next := editEntry(children[i], id, level, bit-1)
	if next == children[i] {
		return n
	}
	children[i] = next
	if children[0] == nil && children[1] == nil {
		return nil
	}
	var best entryChoice
	if children[0] == nil {
		best = children[1].best
	} else if children[1] == nil {
		best = children[0].best
	} else {
		best = preferredEntry(children[0].best, children[1].best)
	}
	return &entrySummaryNode{child: children, best: best}
}

// With replaces one candidate's level or deletes it with level=-1. Preparation
// copies at most33 nodes; Best is O(1). Caller must publish this root atomically
// with the matching graph snapshot. No concurrency/durability owner is included.
func (s EntrySummary) With(id uint32, level int) (EntrySummary, error) {
	if level < -1 || level >= 32 {
		return EntrySummary{}, errors.New("entry level out of bounds")
	}
	return EntrySummary{editEntry(s.root, id, level, 31)}, nil
}
func (s EntrySummary) Best() (uint32, int, bool) {
	if s.root == nil {
		return 0, 0, false
	}
	return s.root.best.id, s.root.best.level, true
}
