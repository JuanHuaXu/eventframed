package researchhybridref

import (
	"math"
	"testing"
)

func TestReferenceExplicitPaths(t *testing.T) {
	for _, prior := range [][]float64{{.9, .1}, {.8, .1, .1}} {
		for _, alpha := range []float64{0, .02, .7, 1} {
			q := make([][]float64, 8)
			for j := range q {
				q[j] = make([]float64, len(prior))
				for h := range prior {
					q[j][h] = .1 + .1*float64((j+h)%8)
				}
			}
			known := map[int]bool{0: true, 3: false, 7: true}
			got, err := End(prior, alpha, q, known)
			if err != nil {
				t.Fatal(err)
			}
			want := make([]float64, len(prior))
			var visit func(int, int, float64)
			visit = func(j, previous int, mass float64) {
				for h := range prior {
					p := prior[h]
					if j > 0 {
						p *= alpha
						if h == previous {
							p += 1 - alpha
						}
					}
					if y, ok := known[j]; ok {
						v := q[j][h]
						if !y {
							v = 1 - v
						}
						p *= v
					}
					if j == len(q)-1 {
						want[h] += mass * p
					} else {
						visit(j+1, h, mass*p)
					}
				}
			}
			visit(0, -1, 1)
			sum := 0.
			for _, v := range want {
				sum += v
			}
			for h := range want {
				if math.Abs(got[h]-want[h]/sum) > 1e-12 {
					t.Fatal("literal paths disagree", alpha, got, want)
				}
			}
		}
	}
}

func TestReferenceRejectsFutureObservation(t *testing.T) {
	if _, err := End([]float64{.9, .1}, 0, nil, map[int]bool{0: true}); err == nil {
		t.Fatal("future label accepted")
	}
}
