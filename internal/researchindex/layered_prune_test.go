package researchindex

import (
	"context"
	"testing"
)

func TestLayeredPruneChurn(t *testing.T) {
	ctx := context.Background()
	limits := LayeredLimits{2, 2, 1, 8}
	r := LayeredRecord{ID: "owned", Vector: []float32{1, 2}, Links: [][]uint32{{}}, Backlinks: [][]uint32{{}}, Heuristic: []uint32{0}}
	var root LayeredSnapshot
	var history []LayeredSnapshot
	for id := uint32(0); id < 1024; id++ {
		next, _, err := PrepareLayered(ctx, root, []LayeredEdit{{id, &r}}, uint64(id), limits)
		if err != nil {
			t.Fatal(err)
		}
		if id < 8 {
			history = append(history, next)
		}
		root, _, err = PrepareLayered(ctx, next, []LayeredEdit{{id, nil}}, uint64(id), limits)
		if err != nil {
			t.Fatal(err)
		}
		if root.tree != nil {
			t.Fatalf("deleted-only root retained paths at id %d", id)
		}
	}
	if got := countLayeredStorage([]LayeredSnapshot{root}); got.LogicalBytes != 0 {
		t.Fatal(got)
	}
	for id, old := range history {
		if _, ok := old.Lookup(uint32(id)); !ok {
			t.Fatal("pruning changed old root")
		}
	}
}

func TestLayeredPruneSiblingAndMissing(t *testing.T) {
	ctx := context.Background()
	limits := LayeredLimits{3, 2, 1, 8}
	r := LayeredRecord{ID: "x", Vector: []float32{1, 2}, Links: [][]uint32{{}}, Backlinks: [][]uint32{{}}, Heuristic: []uint32{0}}
	root, _, err := PrepareLayered(ctx, LayeredSnapshot{}, []LayeredEdit{{0, &r}, {1, &r}, {1 << 31, &r}}, 0, limits)
	if err != nil {
		t.Fatal(err)
	}
	oldRight := root.tree.child[1]
	next, _, err := PrepareLayered(ctx, root, []LayeredEdit{{0, nil}}, 1, limits)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []uint32{1, 1 << 31} {
		if _, ok := next.Lookup(id); !ok {
			t.Fatal("lost sibling", id)
		}
	}
	if next.tree.child[1] != oldRight {
		t.Fatal("unrelated branch copied")
	}
	missing, stats, err := PrepareLayered(ctx, next, []LayeredEdit{{123, nil}}, 2, limits)
	if err != nil {
		t.Fatal(err)
	}
	if missing.tree != next.tree || stats.Paths != 0 {
		t.Fatal("missing deletion allocated or changed root", stats)
	}
	if missing.Global() != 2 {
		t.Fatal("global update lost on no-op tree")
	}
	if _, ok := root.Lookup(0); !ok {
		t.Fatal("history lost")
	}
}
