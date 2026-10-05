package observationlearners

import "testing"

func BenchmarkCompatV111WholeFixture(b *testing.B) {
	for _, schedule := range []int{0, 1} {
		name := "immediate"
		if schedule == 1 {
			name = "delayed_missing"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				// Consumed design parity4, not a fresh quality sample. Includes
				// fitted tables and all nine policies, not a single serving call.
				if _, err := compatV111Run(0, 3, 0, schedule, 2130110900); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
