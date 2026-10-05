package observationlearners

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func transformHistory(n int) []segmentPacket {
	h := make([]segmentPacket, n)
	for i := range h {
		h[i] = segmentPacket{Bits: uint16(i * 137 % 512), Outcome: i%3 == 0, Arrives: max(0, i-16) + i%17}
		if i%7 == 0 {
			h[i].Arrives = -1
		}
	}
	return h
}

func TestTransformFullPair(t *testing.T) {
	count := 0
	maxError := 0.
	for _, n := range []int{16, 64, 272} {
		history := transformHistory(n)
		for _, cap := range []int{1, 16, 64} {
			for _, mass := range []float64{0, .95, 1} {
				for _, hazard := range []float64{.01, .5} {
					a, e := fitSegmentPosterior(-16, n-16, history, cap, hazard, mass)
					if e != nil {
						t.Fatal(e)
					}
					b, e := fitSegmentPosteriorTransform(-16, n-16, history, cap, hazard, mass)
					if e != nil {
						t.Fatal(e)
					}
					if a.logEvidence != b.logEvidence || !reflect.DeepEqual(a.lastStart, b.lastStart) || !reflect.DeepEqual(a.origins, b.origins) {
						t.Fatal("posterior state mismatch")
					}
					for x, p := range a.predictions {
						delta := math.Abs(p - b.predictions[x])
						maxError = math.Max(maxError, delta)
						count++
						if math.IsNaN(delta) || delta > 1e-12 {
							t.Fatal("forecast mismatch", n, cap, mass, hazard, x, delta)
						}
					}
				}
			}
		}
	}
	t.Logf("%d full forecast comparisons; maximum absolute error %.17g", count, maxError)
}

func BenchmarkTransformFullPair(b *testing.B) {
	for _, n := range []int{32, 80, 272} {
		history := transformHistory(n)
		for _, fast := range []bool{false, true} {
			b.Run(fmt.Sprintf("history%d/transform%t", n, fast), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					var m *segmentPosterior
					var e error
					if fast {
						m, e = fitSegmentPosteriorTransform(-16, n-16, history, 64, .01, .95)
					} else {
						m, e = fitSegmentPosterior(-16, n-16, history, 64, .01, .95)
					}
					if e != nil {
						b.Fatal(e)
					}
					segmentBenchmarkModel = m
				}
			})
		}
	}
}
