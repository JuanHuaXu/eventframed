package observationlearners

import (
	"context"
	"reflect"
	"testing"
)

func TestTableBitwiseBatchParity(t *testing.T) {
	for _, n := range []int{0, 1, 16, 64, 272} {
		for _, mass := range []float64{0, 1e-12, .5, .95, 1} {
			h := transformHistory(n)
			left, clock := -16, n-16
			if n < 16 {
				left, clock = 0, n
				for i := range h {
					h[i].Arrives = i
				}
			}
			a, e := fitSegmentPosteriorBatch(context.Background(), left, clock, h, 64, .01, mass)
			if e != nil {
				t.Fatal(e)
			}
			b, e := fitSegmentPosteriorTable(context.Background(), left, clock, h, 64, .01, mass)
			if e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(a, b) {
				t.Fatal("table changed batch output", n, mass)
			}
		}
	}
}

func BenchmarkTablePair(b *testing.B) {
	h := transformHistory(272)
	for _, table := range []bool{false, true} {
		name := "batch"
		if table {
			name = "table"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var m *segmentPosterior
				var e error
				if table {
					m, e = fitSegmentPosteriorTable(context.Background(), -16, 256, h, 64, .01, .95)
				} else {
					m, e = fitSegmentPosteriorBatch(context.Background(), -16, 256, h, 64, .01, .95)
				}
				if e != nil {
					b.Fatal(e)
				}
				segmentBenchmarkModel = m
			}
		})
	}
}
