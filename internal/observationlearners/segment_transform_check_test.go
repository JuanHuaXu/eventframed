package observationlearners

import (
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"testing"
)

func TestSegmentTransformParity(t *testing.T) {
	maxError := 0.
	checks := 0
	for _, n := range []int{0, 1, 16, 64} {
		for fixture := 0; fixture < 3; fixture++ {
			samples := ridgeTestSamples(n)
			for i := range samples {
				if fixture == 1 {
					samples[i] = observation.Sample{Bits: 511, Outcome: true}
				}
				if fixture == 2 {
					samples[i] = observation.Sample{Bits: uint16(i * 137 % 512), Outcome: i%2 == 0}
				}
			}
			for _, mass := range []float64{0, 1e-12, .5, .95, 1 - 1e-12, 1} {
				direct, e := buildSegmentLikelihoods(samples, mass)
				if e != nil {
					t.Fatal(e)
				}
				got, e := buildSegmentLikelihoodsTransform(samples, mass)
				if e != nil {
					t.Fatal(e)
				}
				if direct.logM != got.logM {
					t.Fatal("likelihood changed")
				}
				for start := 0; start <= n; start++ {
					for x, p := range direct.tail[start] {
						delta := math.Abs(p - got.tail[start][x])
						maxError = math.Max(maxError, delta)
						checks++
						if math.IsNaN(delta) || delta > 1e-12 || got.tail[start][x] < 0 || got.tail[start][x] > 1 {
							t.Fatalf("n=%d fixture=%d mass=%g start=%d x=%d delta=%g", n, fixture, mass, start, x, delta)
						}
					}
				}
			}
		}
	}
	t.Logf("%d tail comparisons; maximum absolute error %.17g", checks, maxError)
}

func BenchmarkSegmentTransformPair(b *testing.B) {
	for _, n := range []int{16, 64} {
		samples := ridgeTestSamples(n)
		for _, fast := range []bool{false, true} {
			b.Run(fmt.Sprintf("n%d/transform%t", n, fast), func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					var m *segmentLikelihoods
					var e error
					if fast {
						m, e = buildSegmentLikelihoodsTransform(samples, .95)
					} else {
						m, e = buildSegmentLikelihoods(samples, .95)
					}
					if e != nil {
						b.Fatal(e)
					}
					segmentBenchmarkLikelihoods = m
				}
			})
		}
	}
}
