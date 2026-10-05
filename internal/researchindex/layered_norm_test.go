package researchindex

import (
	"context"
	"math"
	"math/rand/v2"
	"testing"
)

func TestLayeredNormExactness(t *testing.T) {
	rng := rand.New(rand.NewPCG(17, 31))
	for _, dim := range []int{1, 2, 17, 768, 4096} {
		for trial := 0; trial < 32; trial++ {
			a, b := make([]float32, dim), make([]float32, dim)
			var aa, bb float64
			for i := range a {
				a[i] = float32(math.Ldexp(rng.Float64()*2-1, trial*7-100))
				b[i] = float32(math.Ldexp(rng.Float64()*2-1, 100-trial*7))
				aa += float64(a[i]) * float64(a[i])
				bb += float64(b[i]) * float64(b[i])
			}
			got, want := cosineWithNorms(a, b, aa, bb), cosine(a, b)
			if math.Float64bits(got) != math.Float64bits(want) {
				t.Fatal("distance differs", dim, trial, got, want)
			}
		}
	}
	ctx := context.Background()
	r := &LayeredRecord{ID: "a", Vector: []float32{3, 4}, Links: [][]uint32{{}}, Backlinks: [][]uint32{{}}, Heuristic: []uint32{0}, normSquared: 999}
	s, _, e := PrepareLayered(ctx, LayeredSnapshot{}, []LayeredEdit{{0, r}}, 1, LayeredLimits{1, 2, 1, 8})
	if e != nil {
		t.Fatal(e)
	}
	if layeredFind(s.tree, 0).normSquared != 25 {
		t.Fatal("trusted caller norm")
	}
	edited, _ := s.Lookup(0)
	if edited.normSquared != 0 {
		t.Fatal("derived state exported")
	}
	edited.Vector[0] = 0
	edited.normSquared = 777
	next, _, e := PrepareLayered(ctx, s, []LayeredEdit{{0, &edited}}, 1, LayeredLimits{1, 2, 1, 8})
	if e != nil {
		t.Fatal(e)
	}
	if layeredFind(next.tree, 0).normSquared != 16 || layeredFind(s.tree, 0).normSquared != 25 {
		t.Fatal("stale or mutated norm")
	}
}
