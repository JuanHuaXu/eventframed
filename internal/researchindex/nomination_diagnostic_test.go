package researchindex

import (
	"context"
	"math"
	"testing"
)

func TestNominationDiagnosticGuards(t *testing.T) {
	ctx := context.Background()
	d, err := RestoreDurable(2, 8, 1, []Mutation{{ID: "a", Vector: []float32{1, 0}}}, func(context.Context, uint64, []Mutation) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	v, err := d.View(ctx)
	if err != nil {
		t.Fatal(err)
	}
	h, err := BuildHNSWBase(ctx, v, 2, t.TempDir()+"/base")
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	for _, ef := range []int{0, 100, 800, 3200} {
		r, e := h.ResearchNomination(ctx, []float32{1, 0}, 1, ef)
		if e != nil || len(r) != 1 || r[0].ID != "a" || r[0].Score != 1 {
			t.Fatal(r, e)
		}
	}
	for _, q := range [][]float32{nil, {0, 0}, {float32(math.NaN()), 1}, {float32(math.Inf(1)), 0}} {
		if _, e := h.ResearchNomination(ctx, q, 1, 0); e == nil {
			t.Fatal("invalid query accepted", q)
		}
	}
	for _, ef := range []int{-1, 6401} {
		if _, e := h.ResearchNomination(ctx, []float32{1, 0}, 1, ef); e == nil {
			t.Fatal("bad effort accepted")
		}
	}
	if err := h.Close(); err != nil {
		t.Fatal(err)
	}
	if _, e := h.ResearchNomination(ctx, []float32{1, 0}, 1, 0); e == nil {
		t.Fatal("closed base accepted")
	}
}
