package observationlearners

import "testing"

type compatibilityBenchAdapter struct{ *compatibilityAdviceJournal }

func (g compatibilityBenchAdapter) statsSnapshot() routedJournalStats { return g.journal.stats }
func (g compatibilityBenchAdapter) selectorCount() uint64             { return g.received }

func BenchmarkCompatibilityJournal(b *testing.B) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		b.Fatal(err)
	}
	other, err := newObservationExperts(jointFixture(true, true))
	if err != nil {
		b.Fatal(err)
	}
	for _, name := range []string{"log_identity", "handoff_identity", "pending_identity", "log_moving", "handoff_moving", "pending_moving"} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				var g logV109Journal
				switch name {
				case "log_identity", "log_moving":
					g = likelihoodV109Adapter{newLogAdviceJournal(false)}
				case "handoff_identity", "handoff_moving":
					g = compatibilityBenchAdapter{newCompatibilityAdviceJournal(false)}
				case "pending_identity", "pending_moving":
					g = compatibilityBenchAdapter{newCompatibilityAdviceJournal(true)}
				}
				// Same 256 forecasts, release-before-issue delay8 schedule, expiry
				// scans, eight publications and complete flush for every variant.
				for clock := 0; clock < 264; clock++ {
					if timed, ok := g.(interface{ setClock(uint64) error }); ok {
						if err := timed.setClock(uint64(clock)); err != nil {
							b.Fatal(err)
						}
					}
					if clock >= 8 {
						origin := clock - 8
						if err := g.deliver(uint64(origin), origin%3 == 0); err != nil {
							b.Fatal(err)
						}
					}
					if clock >= 32 {
						if err := g.expireBefore(uint64(clock - 31)); err != nil {
							b.Fatal(err)
						}
					}
					if clock < 256 {
						if clock%32 == 0 {
							model := e
							if (name == "log_moving" || name == "handoff_moving" || name == "pending_moving") && clock/32%2 == 1 {
								model = other
							}
							if err := g.publish(model); err != nil {
								b.Fatal(err)
							}
						}
						if _, err := g.predict(uint64(clock), &jointReader{x: uint16(clock), epoch: 1}, 1); err != nil {
							b.Fatal(err)
						}
					}
				}
				if g.statsSnapshot().Pending != 0 || g.selectorCount() != 256 {
					b.Fatal("benchmark settlement")
				}
			}
		})
	}
}
