package observationlearners

import "testing"

// Both builders receive the same fixture. These timings exclude evidence
// acquisition, queueing, publication and selector integration.
func BenchmarkBooleanComparison(b *testing.B) {
	for _, n := range []int{16, 64, 256} {
		for _, arm := range []string{"subset", "boolean"} {
			name := map[int]string{16: "n16", 64: "n64", 256: "n256"}[n] + "/" + arm
			b.Run(name, func(b *testing.B) {
				samples := booleanFixture(n)
				var weights [512]float64
				for i := range weights {
					weights[i] = 1
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					var m *ConditionalForest
					var err error
					if arm == "subset" {
						m, err = NewSubsetConditional(samples, weights)
					} else {
						m, err = NewBooleanConditional(samples)
					}
					if err != nil {
						b.Fatal(err)
					}
					booleanBenchmarkSink = m
				}
			})
		}
	}
}
