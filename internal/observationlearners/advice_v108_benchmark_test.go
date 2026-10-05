package observationlearners

import "testing"

func BenchmarkAdviceV108Journal(b *testing.B) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		b.Fatal(err)
	}
	for _, name := range []string{"arrival", "neutral", "age", "combined"} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				var g adviceV108Journal
				switch name {
				case "arrival":
					g = adviceV108Adapter{newArrivalRoutedJournal()}
				case "neutral":
					g = agedV108Adapter{newAgedAdviceJournal(true, false)}
				case "age":
					g = agedV108Adapter{newAgedAdviceJournal(false, true)}
				case "combined":
					g = agedV108Adapter{newAgedAdviceJournal(true, true)}
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
							if err := g.publish(e); err != nil {
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

func BenchmarkAdviceV108WholeFixture(b *testing.B) {
	for _, schedule := range []int{0, 1} {
		name := "immediate"
		if schedule == 1 {
			name = "delayed_missing"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				// Consumed v106 design parity4 trajectory, not a new quality sample.
				if _, err := adviceV108Run(0, 3, 0, schedule, 2110110600); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
