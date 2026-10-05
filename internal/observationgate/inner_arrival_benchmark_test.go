package observationgate

import "testing"

func BenchmarkInnerArrivalFixture(b *testing.B) {
	cfg := forestDelayCases()[0]
	base, err := forestDependenceBase(cfg, 2026092191)
	if err != nil {
		b.Fatal(err)
	}
	for _, delay := range []bool{false, true} {
		name := "immediate"
		if delay {
			name = "delayed"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := innerArrivalRun(base, cfg, 0, 0, 2026092192, delay); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
