package observationlearners

import (
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type snapshotBenchJournal interface {
	setClock(uint64) error
	publish(*observationExperts) error
	predict(uint64, observation.Reader, uint64) (observation.Result, error)
	deliver(uint64, bool) error
	expireBefore(uint64) error
}

func BenchmarkSnapshotJournalLifecycle(b *testing.B) {
	e0, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		b.Fatal(err)
	}
	e1, err := newObservationExperts(jointFixture(true, true))
	if err != nil {
		b.Fatal(err)
	}
	for _, delay := range []int{0, 31} {
		for _, variant := range []string{"arrival_log", "fixed_markov", "snapshot"} {
			name := "immediate/" + variant
			if delay > 0 {
				name = "delay31/" + variant
			}
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for n := 0; n < b.N; n++ {
					var g snapshotBenchJournal
					if variant == "arrival_log" {
						g = newLogAdviceJournal(false)
					} else if variant == "fixed_markov" {
						g = newMarkovAdviceJournal()
					} else {
						g = newSnapshotJournal()
					}
					for clock := 0; clock < 256+delay; clock++ {
						if err := g.setClock(uint64(clock)); err != nil {
							b.Fatal(err)
						}
						if delay > 0 && clock >= delay {
							if err := g.deliver(uint64(clock-delay), (clock-delay)%3 == 0); err != nil {
								b.Fatal(err)
							}
						}
						if clock >= 32 {
							if err := g.expireBefore(uint64(clock - 31)); err != nil {
								b.Fatal(err)
							}
						}
						if clock >= 256 {
							continue
						}
						if clock%32 == 0 {
							e := e0
							if clock%64 == 32 {
								e = e1
							}
							if err := g.publish(e); err != nil {
								b.Fatal(err)
							}
						}
						if _, err := g.predict(uint64(clock), &jointReader{x: uint16(clock * 17 % 512), epoch: 1}, 1); err != nil {
							b.Fatal(err)
						}
						if delay == 0 {
							if err := g.deliver(uint64(clock), clock%3 == 0); err != nil {
								b.Fatal(err)
							}
						}
					}
				}
			})
		}
	}
}
