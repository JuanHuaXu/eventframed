package observationlearners

import "testing"

// Entire three-arm, 256-step fixture including fitting and flush. Consumed
// compatibility seeds avoid turning timing into additional quality evidence.
func BenchmarkDelayedV104Fixture(b *testing.B) {
	for schedule, name := range []string{"immediate", "delay_missing"} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				r, err := delayedV104Run(0, 3, 0, schedule, 2090110300)
				if err != nil || r.Final[0].Pending != 0 || r.Final[1].Pending != 0 {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkRoleCarryDelay8Lifetime(b *testing.B) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		g := newRoleRoutedFeedbackJournal()
		for origin := 0; origin < 256; origin++ {
			if origin%32 == 0 {
				if err := g.publish(e); err != nil {
					b.Fatal(err)
				}
			}
			if _, err := g.predict(uint64(origin), &jointReader{x: uint16(origin), epoch: 1}, 1); err != nil {
				b.Fatal(err)
			}
			if origin >= 8 {
				if err := g.deliver(uint64(origin-8), (origin-8)%3 == 0); err != nil {
					b.Fatal(err)
				}
			}
		}
		for origin := 248; origin < 256; origin++ {
			if err := g.deliver(uint64(origin), origin%3 == 0); err != nil {
				b.Fatal(err)
			}
		}
		journalCheck(b, g)
		if g.stats.Applied+g.stats.BankOnly != 256 {
			b.Fatal("unsettled carry lifetime")
		}
	}
}
