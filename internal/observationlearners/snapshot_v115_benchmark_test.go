package observationlearners

import "testing"

func BenchmarkSnapshotV115WholeFixture(b *testing.B) {
	for _, schedule := range []int{0, 1} {
		name := "immediate"
		if schedule == 1 {
			name = "delayed_missing"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				// Consumed v113 design parity4; fits and all eleven policies included.
				if _, err := snapshotV115Run(0, 3, 0, schedule, 2150111300); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
