package observationlearners

import (
	"context"
	"math"
	"reflect"
	"testing"
)

func TestBatchFullPair(t *testing.T) {
	maxP, maxZ, maxW := 0., 0., 0.
	for _, n := range []int{16, 64, 272} {
		for _, cap := range []int{1, 16, 64} {
			for _, mass := range []float64{0, 1e-12, .5, .95, 1 - 1e-12, 1} {
				h := transformHistory(n)
				a, e := fitSegmentPosterior(-16, n-16, h, cap, .01, mass)
				if e != nil {
					t.Fatal(e)
				}
				b, e := fitSegmentPosteriorBatch(context.Background(), -16, n-16, h, cap, .01, mass)
				if e != nil {
					t.Fatal(e)
				}
				if !reflect.DeepEqual(a.origins, b.origins) {
					t.Fatal("origin mismatch")
				}
				maxZ = math.Max(maxZ, math.Abs(a.logEvidence-b.logEvidence))
				for i, p := range a.lastStart {
					maxW = math.Max(maxW, math.Abs(p-b.lastStart[i]))
				}
				for i, p := range a.predictions {
					maxP = math.Max(maxP, math.Abs(p-b.predictions[i]))
				}
			}
		}
	}
	if math.IsNaN(maxP+maxZ+maxW) || maxP > 1e-12 || maxZ > 1e-10 || maxW > 1e-12 {
		t.Fatal("parity", maxP, maxZ, maxW)
	}
	t.Logf("max forecast %.17g evidence %.17g boundary weight %.17g", maxP, maxZ, maxW)
}

func BenchmarkBatchFullPair(b *testing.B) {
	h := transformHistory(272)
	for _, batch := range []bool{false, true} {
		name := "pairwise"
		if batch {
			name = "batch"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var m *segmentPosterior
				var e error
				if batch {
					m, e = fitSegmentPosteriorBatch(context.Background(), -16, 256, h, 64, .01, .95)
				} else {
					m, e = fitSegmentPosteriorContext(context.Background(), -16, 256, h, 64, .01, .95)
				}
				if e != nil {
					b.Fatal(e)
				}
				segmentBenchmarkModel = m
			}
		})
	}
}
