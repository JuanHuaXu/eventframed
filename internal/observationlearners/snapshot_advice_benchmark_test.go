package observationlearners

import "testing"

func BenchmarkSnapshotAdviceLifecycle(b *testing.B) {
	for _, delay := range []int{0, 31} {
		name := "immediate"
		if delay == 31 {
			name = "delay31"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				f, _ := newSnapshotAdvice(.001)
				for clock := 0; clock < 256+delay; clock++ {
					if delay > 0 && clock >= delay {
						if err := f.deliver(uint64(clock-delay), (clock-delay)%3 == 0); err != nil {
							b.Fatal(err)
						}
					}
					if clock >= 256 {
						continue
					}
					if clock%32 == 0 {
						if err := f.publish(clock / 32); err != nil {
							b.Fatal(err)
						}
					}
					if err := f.issue(uint64(clock), f.current.ids, snapshotRaw(f.current.ids, uint64(clock))); err != nil {
						b.Fatal(err)
					}
					if delay == 0 {
						if err := f.deliver(uint64(clock), clock%3 == 0); err != nil {
							b.Fatal(err)
						}
					}
				}
				if f.head != f.tail {
					b.Fatal("pending tail")
				}
			}
		})
	}
}

func BenchmarkSnapshotBankPublication(b *testing.B) {
	models := jointFixture(true, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bank := newSnapshotBank()
		if err := bank.publish(0, models); err != nil {
			b.Fatal(err)
		}
	}
}
