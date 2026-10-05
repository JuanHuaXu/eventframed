package observationlearners

import (
	"fmt"
	"testing"
	"unsafe"
)

var segmentBenchmarkModel *segmentPosterior
var segmentBenchmarkLikelihoods *segmentLikelihoods

func BenchmarkSegmentComponent(b *testing.B) {
	for _, n := range []int{16, 32, 64} {
		samples := ridgeTestSamples(n)
		history := make([]segmentPacket, n)
		for i, s := range samples {
			history[i] = segmentPacket{Bits: s.Bits, Outcome: s.Outcome, Arrives: i}
		}
		b.Run(fmt.Sprintf("Likelihoods%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m, e := buildSegmentLikelihoods(samples, .95)
				if e != nil {
					b.Fatal(e)
				}
				segmentBenchmarkLikelihoods = m
			}
			b.ReportMetric(float64(unsafe.Sizeof(segmentLikelihoods{})), "workspace-bytes")
		})
		b.Run(fmt.Sprintf("FullFit%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m, e := fitSegmentPosterior(0, n, history, 64, .01, .95)
				if e != nil {
					b.Fatal(e)
				}
				segmentBenchmarkModel = m
			}
		})
		b.Run(fmt.Sprintf("Generic%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m, e := fitSubset(samples)
				if e != nil {
					b.Fatal(e)
				}
				ridgeBenchmarkSubset = m
			}
		})
		b.Run(fmt.Sprintf("Boolean%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m, e := fitBooleanSpecialist(samples)
				if e != nil {
					b.Fatal(e)
				}
				ridgeBenchmarkBoolean = m
			}
		})
	}
	history := make([]segmentPacket, 272)
	for i := range history {
		history[i] = segmentPacket{Bits: uint16(i * 17 % 512), Outcome: i%3 == 0, Arrives: max(0, i-16)}
	}
	b.Run("FullHistory272Cap64", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			m, e := fitSegmentPosterior(-16, 256, history, 64, .01, .95)
			if e != nil {
				b.Fatal(e)
			}
			segmentBenchmarkModel = m
		}
	})
}
