package researchindex

import (
	"context"
	"testing"
)

func TestPrivateReconnectNeededAndBudget(t *testing.T) {
	ctx := context.Background()
	lim := LayeredLimits{8, 2, 1, 32}
	var edits []LayeredEdit
	for i := uint32(0); i < 4; i++ {
		edits = append(edits, LayeredEdit{i, &LayeredRecord{ID: "x", Vector: []float32{1, 2}, Links: [][]uint32{{}}, Backlinks: [][]uint32{{}}, Heuristic: []uint32{1}}})
	}
	s, _, e := PrepareLayered(ctx, LayeredSnapshot{}, edits, 0, lim)
	if e != nil {
		t.Fatal(e)
	}
	distance := func(a, b uint32) (float32, error) { return float32(b - a), nil }
	for _, limits := range [][2]int{{4, 5}, {1, 6}} {
		p, e := DiscoverReconnect(ctx, s, []uint32{0, 1, 2, 3}, 0, 2, 4, limits[0], limits[1], distance)
		if e == nil || len(p.Edits) > 0 {
			t.Fatal("budget leaked partial work")
		}
	}
	p, e := DiscoverReconnect(ctx, s, []uint32{0, 1, 2, 3}, 0, 2, 4, 4, 6, distance)
	if e != nil {
		t.Fatal(e)
	}
	if p.PairCalls != 6 || len(p.Edits) != 4 {
		t.Fatal(p)
	}
	for _, edit := range p.Edits {
		old, _ := s.Lookup(edit.Ordinal)
		if len(old.Links[0]) != 0 {
			t.Fatal("source changed")
		}
		if len(edit.Record.Links[0]) < 2 {
			t.Fatal("repair missing")
		}
	}
	next, _, e := PrepareLayered(ctx, s, p.Edits, 0, lim)
	if e != nil {
		t.Fatal(e)
	}
	p, e = DiscoverReconnect(ctx, next, []uint32{0, 1, 2, 3}, 0, 2, 4, 4, 0, func(a, b uint32) (float32, error) { t.Fatal("unneeded metric called"); return 0, nil })
	if e != nil || p.PairCalls != 0 || len(p.Edits) != 0 {
		t.Fatal(p, e)
	}
}
