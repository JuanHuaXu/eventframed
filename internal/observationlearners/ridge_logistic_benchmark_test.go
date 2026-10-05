package observationlearners

import (
	"fmt"
	"testing"
	"unsafe"
)

var ridgeBenchmarkModel *ridgeLogistic
var ridgeBenchmarkLaw *ConditionalForest
var ridgeBenchmarkSubset *subsetModel
var ridgeBenchmarkBoolean *booleanSpecialist
var ridgeBenchmarkTree *contextTree
var ridgeBenchmarkProbability float64

func BenchmarkRidgeComponent(b *testing.B) {
	for _, n := range []int{32, 64} {
		samples := ridgeTestSamples(n)
		b.Run(fmt.Sprintf("Fit%d/Ridge", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m, err := fitRidgeLogistic(samples)
				if err != nil {
					b.Fatal(err)
				}
				ridgeBenchmarkModel = m
			}
			b.ReportMetric(float64(ridgeBenchmarkModel.stats.Iterations), "iterations/fit")
			b.ReportMetric(float64(unsafe.Sizeof(ridgeLogistic{})), "model-bytes")
		})
		b.Run(fmt.Sprintf("Fit%d/Subset", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m, err := fitSubset(samples)
				if err != nil {
					b.Fatal(err)
				}
				ridgeBenchmarkSubset = m
			}
		})
		b.Run(fmt.Sprintf("Fit%d/Boolean", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m, err := fitBooleanSpecialist(samples)
				if err != nil {
					b.Fatal(err)
				}
				ridgeBenchmarkBoolean = m
			}
		})
		b.Run(fmt.Sprintf("Fit%d/ContextTree", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m, err := fitContextTree(samples)
				if err != nil {
					b.Fatal(err)
				}
				ridgeBenchmarkTree = m
			}
		})
	}
	m, err := fitRidgeLogistic(ridgeTestSamples(64))
	if err != nil {
		b.Fatal(err)
	}
	var weights [512]float64
	for i := range weights {
		weights[i] = 1
	}
	b.Run("CompileOnly", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			law, err := m.conditional(weights)
			if err != nil {
				b.Fatal(err)
			}
			ridgeBenchmarkLaw = law
		}
		b.ReportMetric(float64(unsafe.Sizeof(ConditionalForest{})), "table-bytes")
	})
	b.Run("FullPrediction", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p, err := m.predict(uint16(i & 511))
			if err != nil {
				b.Fatal(err)
			}
			ridgeBenchmarkProbability = p
		}
	})
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
