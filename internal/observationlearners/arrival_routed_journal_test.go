package observationlearners

import (
	"reflect"
	"testing"
)

func TestArrivalRoutedImmediate(t *testing.T) {
	e, err := newObservationExperts(jointFixture(false, false))
	if err != nil {
		t.Fatal(err)
	}
	g := newArrivalRoutedJournal()
	ref := newRoleRoutedFeedbackJournal()
	for origin := uint64(0); origin < 256; origin++ {
		if origin%32 == 0 {
			if err := g.publish(e); err != nil {
				t.Fatal(err)
			}
			if err := ref.publish(e); err != nil {
				t.Fatal(err)
			}
		}
		got, err := g.predict(origin, &jointReader{x: uint16(origin), epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		want, err := ref.predict(origin, &jointReader{x: uint16(origin), epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatal("immediate forecast mismatch")
		}
		if err := g.deliver(origin, origin%3 == 0); err != nil {
			t.Fatal(err)
		}
		if err := ref.deliver(origin, origin%3 == 0); err != nil {
			t.Fatal(err)
		}
		if g.journal != *ref || g.received != origin+1 {
			t.Fatal("immediate state mismatch")
		}
	}
}

func TestArrivalRoutedOrdering(t *testing.T) {
	e, err := newObservationExperts(jointFixture(false, false))
	if err != nil {
		t.Fatal(err)
	}
	g := newArrivalRoutedJournal()
	g.publish(e)
	for origin := uint64(0); origin < 4; origin++ {
		if _, err := g.predict(origin, &jointReader{x: uint16(origin * 71), epoch: 1}, 1); err != nil {
			t.Fatal(err)
		}
	}
	entries := g.journal.entries
	refBank := g.journal.core.bank
	for _, origin := range []uint64{3, 2, 0, 1} {
		// Use the frozen journal as a gate-only reference for this transition.
		gateRef := g.journal
		if err := gateRef.deliver(origin, origin%2 == 0); err != nil {
			t.Fatal(err)
		}
		refBank.pending = entries[origin].bank
		refBank.active = true
		if err := refBank.observe(origin, origin%2 == 0); err != nil {
			t.Fatal(err)
		}
		if err := g.deliver(origin, origin%2 == 0); err != nil {
			t.Fatal(err)
		}
		if g.journal.core.bank != refBank || g.journal.core.gate != gateRef.core.gate || g.journal.stats != gateRef.stats {
			t.Fatal("selector/gate order separation")
		}
		if origin == 3 && (g.journal.stats.Applied != 0 || g.received != 1) {
			t.Fatal("arrival blocked behind missing head")
		}
	}
	before := *g
	if err := g.deliver(3, true); err == nil || *g != before {
		t.Fatal("duplicate mutation")
	}
}

func TestArrivalRoutedExpiryAndVersion(t *testing.T) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	g := newArrivalRoutedJournal()
	g.publish(e)
	for origin := uint64(0); origin < 32; origin++ {
		if _, err := g.predict(origin, &jointReader{x: uint16(origin), epoch: 1}, 1); err != nil {
			t.Fatal(err)
		}
	}
	for origin := uint64(31); origin > 0; origin-- {
		if err := g.deliver(origin, true); err != nil {
			t.Fatal(err)
		}
	}
	if g.received != 31 || g.journal.stats.Applied != 0 {
		t.Fatal("arrival availability")
	}
	beforeBank := g.journal.core.bank
	if err := g.publish(e); err != nil {
		t.Fatal(err)
	}
	gate := g.journal.core.gate
	if err := g.expireBefore(32); err != nil {
		t.Fatal(err)
	}
	if g.journal.core.bank != beforeBank || g.journal.core.gate != gate || g.journal.stats.BankOnly != 31 || g.journal.stats.Censored != 1 {
		t.Fatal("expiry replayed old losses or tests")
	}
	before := *g
	if err := g.deliver(0, false); err == nil || *g != before {
		t.Fatal("expired outcome admitted")
	}
	r := &jointReader{epoch: 1, failAt: 2}
	r.hook = func() {
		if err := g.deliver(31, true); err == nil {
			t.Fatal("reentrant delivery")
		}
		if err := g.expireBefore(32); err == nil {
			t.Fatal("reentrant expiry")
		}
	}
	if _, err := g.predict(32, r, 1); err == nil || *g != before {
		t.Fatal("failed reader mutation")
	}
}

func BenchmarkArrivalRoutedDelay8Lifetime(b *testing.B) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		g := newArrivalRoutedJournal()
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
		journalCheck(b, &g.journal)
		if g.received != 256 || g.journal.stats.Pending != 0 {
			b.Fatal("arrival accounting")
		}
	}
}
