package researchpartition

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func TestTrainingLOOIndependentRefits(t *testing.T) {
	for _, n := range []int{2, 11, 150, 200} {
		base, coordinates := inputs(n)
		order := rand.New(rand.NewSource(937)).Perm(n)
		for _, count := range []int{1, min(16, n), min(32, n), n} {
			m, _ := New(base, coordinates)
			for j, i := range order[:count] {
				if err := m.Observe(i, j%3 != 0); err != nil {
					t.Fatal(err)
				}
			}
			before := *m
			before.base = append([]float64(nil), m.base...)
			before.bin = append([]uint8(nil), m.bin...)
			before.seen = append([]bool(nil), m.seen...)
			before.useful = append([]bool(nil), m.useful...)
			for _, heldout := range order[:count] {
				refit, _ := New(base, coordinates)
				flipped, _ := New(base, coordinates)
				for j, i := range order[:count] {
					y := j%3 != 0
					if i != heldout {
						if err := refit.Observe(i, y); err != nil {
							t.Fatal(err)
						}
					} else {
						y = !y
					}
					if err := flipped.Observe(i, y); err != nil {
						t.Fatal(err)
					}
				}
				want, err := refit.Predict(heldout)
				got, e := m.TrainingLOO(heldout)
				other, f := flipped.TrainingLOO(heldout)
				if err != nil || e != nil || f != nil || math.Abs(got-want) > 2e-13 || math.Abs(got-other) > 2e-13 {
					t.Fatalf("LOO n=%d count=%d i=%d: %.17g refit %.17g flipped %.17g", n, count, heldout, got, want, other)
				}
			}
			for _, i := range []int{-1, n} {
				if _, err := m.TrainingLOO(i); err == nil {
					t.Fatal("invalid omission accepted")
				}
			}
			if count < n {
				if _, err := m.TrainingLOO(order[count]); err == nil {
					t.Fatal("unobserved omission accepted")
				}
			}
			if !reflect.DeepEqual(before, *m) {
				t.Fatal("LOO mutated the full-data model")
			}
		}
	}
}
