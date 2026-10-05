package packing

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func diversityFixture(n int, seed int64) ([]model.Candidate, map[string]string) {
	r := rand.New(rand.NewSource(seed))
	c := make([]model.Candidate, n)
	keys := make(map[string]string)
	for i := range c {
		id := fmt.Sprintf("e%d", i)
		c[i] = model.Candidate{Score: float64(r.Intn(8)) / 8, Event: model.Event{ID: id, Priority: float64(r.Intn(5)) / 4}}
		// The existing FrameText method defines normalization, including empty
		// and overlapping frames; parity must preserve it rather than substitute.
		c[i].Event.What.Value = fmt.Sprintf("mission %d arrived at destination %d", r.Intn(7), r.Intn(3))
		if i%7 == 0 {
			c[i].Event.What.Value = ""
		}
		if i%3 == 0 {
			keys[id] = fmt.Sprintf("ap:%d", r.Intn(4))
		}
	}
	return c, keys
}

func TestResearchIncrementalDiversityParity(t *testing.T) {
	for seed := int64(0); seed < 100; seed++ {
		for _, n := range []int{0, 1, 10, 50} {
			c, keys := diversityFixture(n, seed)
			before := append([]model.Candidate(nil), c...)
			for _, limit := range []int{0, 1, 10, 20, n + 2} {
				p := DefaultPolicy()
				if seed%2 == 0 {
					p.DiversityPenalty = 0
				}
				if seed%5 == 0 {
					p.PriorityPenalty = .2
				}
				want := diversify(c, keys, limit, p)
				got := researchDiversifyIncremental(c, keys, limit, p)
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("seed%d n%d limit%d", seed, n, limit)
				}
				if len(c) > 0 && !reflect.DeepEqual(c, before) {
					t.Fatal("mutated input")
				}
			}
		}
	}
}

func BenchmarkResearchIncrementalDiversity(b *testing.B) {
	for _, n := range []int{50, 200} {
		c, keys := diversityFixture(n, 42)
		for _, fast := range []bool{false, true} {
			b.Run(fmt.Sprintf("n%d/fast%t", n, fast), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if fast {
						researchDiversifyIncremental(c, keys, 20, DefaultPolicy())
					} else {
						diversify(c, keys, 20, DefaultPolicy())
					}
				}
			})
		}
	}
}

func TestResearchIncrementalDiversityFullFrontier(t *testing.T) {
	for seed := int64(0); seed < 20; seed++ {
		c, keys := diversityFixture(200, seed)
		if seed%2 == 0 {
			// Preserve even the reference's duplicate-ID token-map behavior.
			c[199].Event.ID = c[0].Event.ID
		}
		want := diversify(c, keys, 20, DefaultPolicy())
		got := researchDiversifyIncremental(c, keys, 20, DefaultPolicy())
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("full frontier seed%d", seed)
		}
	}
}
