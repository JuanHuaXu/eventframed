package researchindex

import (
	"context"
	"testing"
)

func TestStoredDiagnosticIsIndependentAndOwned(t *testing.T) {
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
	got, err := h.ResearchStoredVector(ctx, "a")
	if err != nil || len(got) != 2 || got[0] != 1 {
		t.Fatal(got, err)
	}
	got[0] = 99
	again, err := h.ResearchStoredVector(ctx, "a")
	if err != nil || again[0] != 1 {
		t.Fatal("returned storage alias", again, err)
	}
	if _, err := h.ResearchStoredVector(ctx, "missing"); err == nil {
		t.Fatal("missing record accepted")
	}
	if err := h.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := h.ResearchStoredVector(ctx, "a"); err == nil {
		t.Fatal("closed database accepted")
	}
}
