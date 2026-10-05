package observationlearners

import "testing"

func BenchmarkHazardV113WholeFixture(b *testing.B) {
	for _, schedule := range []int{0, 1} {
		name := "immediate"
		if schedule == 1 {
			name = "delayed_missing"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				// Consumed v112 design parity4; fits and all nine policies included.
				if _, err := hazardV113Run(0, 3, 0, schedule, 2144111200); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
