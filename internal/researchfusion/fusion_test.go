package researchfusion

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestScoresIndependentRanks(t *testing.T) {
	rng := rand.New(rand.NewSource(15485863))
	for n := 0; n <= MaxFrontier; n++ {
		x := make([]float64, n)
		for i := range x {
			x[i] = float64(rng.Intn(7)) / 6
		}
		before := append([]float64{}, x...)
		for _, prefix := range []int{0, 1, 10, 200} {
			got, err := Scores(context.Background(), x, prefix)
			if err != nil {
				t.Fatal(err)
			}
			for i := range x {
				// Independent O(n^2) rank oracle, not the production sort.
				rank := 1
				for j := range x {
					if x[j] > x[i] || x[j] == x[i] && j < i {
						rank++
					}
				}
				want := 30.5 * (1/float64(61+i) + 1/float64(60+rank))
				if prefix > 0 {
					want /= 2
					if i < prefix {
						want += .5
					}
				}
				if math.Abs(got[i]-want) > 1e-15 || got[i] <= 0 || got[i] > 1 {
					t.Fatalf("n%d i%d: %g != %g", n, i, got[i], want)
				}
			}
			if prefix > 0 && n > prefix {
				inside, outside := 1., 0.
				for i, v := range got {
					if i < prefix {
						inside = math.Min(inside, v)
					} else {
						outside = math.Max(outside, v)
					}
				}
				if inside <= outside {
					t.Fatal("prefix set violated")
				}
			}
		}
		if !reflect.DeepEqual(x, before) {
			t.Fatal("input mutated")
		}
	}
}

func TestScoresControls(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Scores(ctx, nil, 0); err == nil {
		t.Fatal("cancellation")
	}
	for _, x := range [][]float64{{math.NaN()}, {math.Inf(1)}, {-.1}, {1.1}, make([]float64, 201)} {
		if _, err := Scores(context.Background(), x, 0); err == nil {
			t.Fatal("bad input accepted")
		}
	}
	for _, p := range []int{-1, 201} {
		if _, err := Scores(context.Background(), nil, p); err == nil {
			t.Fatal("bad prefix")
		}
	}
	x := []float64{.1, .8, .2, .3, .5}
	y := make([]float64, len(x))
	for i, v := range x {
		y[i] = v * v
	}
	a, _ := Scores(context.Background(), x, 0)
	b, _ := Scores(context.Background(), y, 0)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("monotone-scale dependence")
	}
	// Prefix protection is deliberately NOT a top-1 guarantee.
	c, _ := Scores(context.Background(), []float64{0, 1, 0}, 2)
	order := []int{0, 1, 2}
	sort.SliceStable(order, func(i, j int) bool { return c[order[i]] > c[order[j]] })
	if order[0] != 0 {
		t.Fatal("two-permutation symmetric tie must retain incumbent")
	}
	d, _ := Scores(context.Background(), []float64{0, .8, 1, .2}, 4)
	if d[1] <= d[0] {
		t.Fatal("expected protected-prefix top1 counterexample missing")
	}
}

func BenchmarkScores(b *testing.B) {
	for _, n := range []int{10, 50, 200} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			x := make([]float64, n)
			for i := range x {
				x[i] = float64((i*17)%n) / float64(n)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := Scores(context.Background(), x, 10); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
