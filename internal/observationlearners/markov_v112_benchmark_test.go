package observationlearners

import "testing"

func BenchmarkMarkovV112WholeFixture(b *testing.B) {
	for _, schedule := range []int{0, 1} {
		name := "immediate"
		if schedule == 1 {
			name = "delayed_missing"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				// Consumed design parity4; includes fits and all eight policies.
				if _, err := markovV112Run(0, 3, 0, schedule, 2130110900); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
