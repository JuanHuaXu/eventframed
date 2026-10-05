package researchpairedvalue

import (
	"fmt"
	"testing"
)

func BenchmarkPredictionValueFrontier(b *testing.B) {
	for _, rounds := range []int{1, 8, 16} {
		b.Run(fmt.Sprintf("members150_rounds%d", rounds), func(b *testing.B) {
			base, targets := make([]float64, 150), make([]float64, 150)
			for i := range base {
				base[i] = .25 + .675*float64(i)/149
				targets[i] = 1. / 150
			}
			m, e := New(base, 1, 512, Config{Strength: 2, Hazard: 1. / 16})
			if e != nil {
				b.Fatal(e)
			}
			origins := make([]Ticket, 150)
			at := int64(0)
			for round := 0; round < rounds; round++ {
				for i := range base {
					origins[i], e = m.Issue(i, at)
					if e != nil {
						b.Fatal(e)
					}
					at++
					if _, e = m.Resolve(origins[i], (i+round)%3 != 0, at); e != nil {
						b.Fatal(e)
					}
					at++
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				values, e := m.PredictionValues(origins, targets)
				if e != nil || len(values) != 150 {
					b.Fatalf("query %d: %v", n, e)
				}
			}
		})
	}
}
