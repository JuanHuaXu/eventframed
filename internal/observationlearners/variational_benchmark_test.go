package observationlearners

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

var variationalBenchmarkModel *variationalLogistic

func BenchmarkVariationalComponent(b *testing.B) {
	for _, n := range []int{32, 64, 256} {
		samples := ridgeTestSamples(n)
		if n == 256 {
			for i := range samples {
				samples[i] = observation.Sample{Bits: 0, Outcome: true}
			}
		}
		b.Run(fmt.Sprintf("Fit%d/Variational", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m, err := fitVariationalLogistic(samples)
				if err != nil {
					b.Fatal(err)
				}
				variationalBenchmarkModel = m
			}
			b.ReportMetric(float64(variationalBenchmarkModel.iterations), "iterations/fit")
			b.ReportMetric(float64(unsafe.Sizeof(variationalLogistic{})), "model-bytes")
		})
		b.Run(fmt.Sprintf("Fit%d/MAP", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m, err := fitRidgeLogistic(samples)
				if err != nil {
					b.Fatal(err)
				}
				ridgeBenchmarkModel = m
			}
			b.ReportMetric(float64(ridgeBenchmarkModel.stats.Iterations), "iterations/fit")
		})
	}
	m, err := fitVariationalLogistic(ridgeTestSamples(64))
	if err != nil {
		b.Fatal(err)
	}
	r, err := fitRidgeLogistic(ridgeTestSamples(64))
	if err != nil {
		b.Fatal(err)
	}
	var weights [512]float64
	for i := range weights {
		weights[i] = 1
	}
	for _, variant := range []string{"Variational", "MAP"} {
		b.Run("CompileOnly/"+variant, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var law *ConditionalForest
				var err error
				if variant == "MAP" {
					law, err = r.conditional(weights)
				} else {
					law, err = m.conditional(weights)
				}
				if err != nil {
					b.Fatal(err)
				}
				ridgeBenchmarkLaw = law
			}
			b.ReportMetric(float64(unsafe.Sizeof(ConditionalForest{})), "table-bytes")
		})
		b.Run("FullPrediction/"+variant, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var p float64
				var err error
				if variant == "MAP" {
					p, err = r.predict(uint16(i & 511))
				} else {
					p, err = m.predict(uint16(i & 511))
				}
				if err != nil {
					b.Fatal(err)
				}
				ridgeBenchmarkProbability = p
			}
		})
	}
	law, err := m.conditional(weights)
	if err != nil {
		b.Fatal(err)
	}
	b.Run("PartialLookup", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p, err := law.Forecast(63, uint16(i&63))
			if err != nil {
				b.Fatal(err)
			}
			ridgeBenchmarkProbability = p
		}
	})
}
