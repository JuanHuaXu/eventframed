package researchindex

import (
	"context"
	"testing"
)

func TestRealHNSWBaseShadowsUpdatesAndChecksGeneration(t *testing.T) {
	ctx := context.Background()
	d, err := RestoreDurable(2, 8, 1, []Mutation{{ID: "a", Vector: []float32{1, 0}}, {ID: "b", Vector: []float32{.8, .2}}, {ID: "c", Vector: []float32{0, 1}}}, func(context.Context, uint64, []Mutation) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	old, err := d.View(ctx)
	if err != nil {
		t.Fatal(err)
	}
	h, err := BuildHNSWBase(ctx, old, 2, t.TempDir()+"/base")
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	got, err := h.Search(ctx, old, []float32{1, 0}, 1)
	if err != nil || len(got) != 1 || got[0].ID != "a" {
		t.Fatal(got, err)
	}
	if err := d.Apply(ctx, []Mutation{{ID: "a", Delete: true}, {ID: "b", Vector: []float32{-1, 0}}, {ID: "d", Vector: []float32{1, .1}}}); err != nil {
		t.Fatal(err)
	}
	current, err := d.View(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got, err = h.Search(ctx, current, []float32{1, 0}, 2)
	if err != nil || len(got) != 2 || got[0].ID != "d" || got[1].ID != "c" {
		t.Fatal(got, err)
	}
	got, err = h.Search(ctx, old, []float32{1, 0}, 1)
	if err != nil || got[0].ID != "a" {
		t.Fatal("old view changed", got, err)
	}
	c, err := d.PrepareCompaction(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Publish(ctx); err != nil {
		t.Fatal(err)
	}
	current, err = d.View(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.Search(ctx, current, []float32{1, 0}, 2); err == nil {
		t.Fatal("mismatched base accepted")
	}
	next, err := BuildHNSWBase(ctx, current, 2, t.TempDir()+"/base")
	if err != nil {
		t.Fatal(err)
	}
	defer next.Close()
	got, err = next.Search(ctx, current, []float32{1, 0}, 2)
	if err != nil || got[0].ID != "d" || got[1].ID != "c" {
		t.Fatal(got, err)
	}
	if err := h.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Search(ctx, old, []float32{1, 0}, 1); err == nil {
		t.Fatal("closed index read")
	}
}

func TestRealHNSWEmptyBaseAndInvalidQuery(t *testing.T) {
	m, _ := NewGenerations(2, 4)
	v := m.View()
	h, err := BuildHNSWBase(context.Background(), v, 2, t.TempDir()+"/base")
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	publish(t, m, Mutation{ID: "delta", Vector: []float32{1, 0}})
	got, err := h.Search(context.Background(), m.View(), []float32{1, 0}, 1)
	if err != nil || len(got) != 1 || got[0].ID != "delta" {
		t.Fatal(got, err)
	}
	if _, err := h.Search(context.Background(), v, []float32{0, 0}, 1); err == nil {
		t.Fatal("zero query")
	}
}
