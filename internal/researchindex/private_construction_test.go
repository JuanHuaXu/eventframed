package researchindex

import (
	"context"
	"testing"
)

func TestConstructionLayerIsolation(t *testing.T) {
	ctx := context.Background()
	record := func(id string, v []float32, links [][]uint32) *LayeredRecord {
		return &LayeredRecord{ID: id, Vector: v, Level: len(links) - 1, Links: links, Backlinks: make([][]uint32, len(links)), Heuristic: make([]uint32, len(links))}
	}
	s, _, err := PrepareLayered(ctx, LayeredSnapshot{}, []LayeredEdit{
		{0, record("entry", []float32{1, 0}, [][]uint32{{1}, {2}})},
		{1, record("lower", []float32{0, 1}, [][]uint32{{}})},
		{2, record("upper", []float32{1, 1}, [][]uint32{{}, {}})},
	}, 2, LayeredLimits{3, 2, 2, 8})
	if err != nil {
		t.Fatal(err)
	}
	back := record("entry", []float32{1, 0}, [][]uint32{{}, {}})
	back.Backlinks[0] = []uint32{1}
	backOnly, _, err := PrepareLayered(ctx, s, []LayeredEdit{{0, back}}, 2, LayeredLimits{1, 2, 2, 8})
	if err != nil {
		t.Fatal(err)
	}
	if got, _, err := SearchConstructionLayer(ctx, backOnly, []float32{0, 1}, 0, 0, 2, 8); err != nil || len(got) != 2 || got[0].Ordinal != 1 {
		t.Fatal("backlink unreachable", got, err)
	}
	if got, _, err := SearchLayered(ctx, backOnly, []float32{0, 1}, 1, 2, 8); err != nil || len(got) != 1 || got[0].ID != "lower" {
		t.Fatal("query backlink unreachable", got, err)
	}
	for _, tc := range []struct {
		level, ef, count int
		first            uint32
	}{
		{0, 1, 1, 1}, {0, 2, 2, 1}, {1, 1, 1, 2}, {1, 2, 2, 2},
	} {
		got, calls, err := SearchConstructionLayer(ctx, s, []float32{0, 1}, 0, tc.level, tc.ef, 8)
		if err != nil || len(got) != tc.count || got[0].Ordinal != tc.first || calls != 2 {
			t.Fatalf("%+v: %v %d %v", tc, got, calls, err)
		}
	}
	// A non-global entry does not descend or jump to the global entry.
	got, calls, err := SearchConstructionLayer(ctx, s, []float32{0, 1}, 2, 0, 8, 8)
	if err != nil || len(got) != 1 || got[0].Ordinal != 2 || calls != 1 {
		t.Fatal(got, calls, err)
	}
	for _, tc := range []struct {
		start         uint32
		level, budget int
	}{{0, 0, 1}, {1, 1, 8}, {99, 0, 8}} {
		got, _, err := SearchConstructionLayer(ctx, s, []float32{0, 1}, tc.start, tc.level, 2, tc.budget)
		if err == nil || got != nil {
			t.Fatal("accepted invalid/partial result", tc, got, err)
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if got, _, err := SearchConstructionLayer(canceled, s, []float32{0, 1}, 0, 0, 2, 8); err == nil || got != nil {
		t.Fatal(got, err)
	}
	if got, _, err := SearchConstructionLayer(ctx, LayeredSnapshot{}, []float32{1, 0}, 0, 0, 2, 8); err == nil || got != nil {
		t.Fatal(got, err)
	}
}
