package observationlearners

import (
	"math"
	"testing"
)

func TestAgedAdviceJournalCompatibility(t *testing.T) {
	e, err := newObservationExperts(jointFixture(false, false))
	if err != nil {
		t.Fatal(err)
	}
	for _, delay := range []int{0, 8, 31} {
		g := newAgedAdviceJournal(false, false)
		ref := newArrivalRoutedJournal()
		release := func(clock int) {
			for i := 0; i < clock && i < 256; i++ {
				if i+delay == clock && i%5 != 0 {
					if err := g.deliver(uint64(i), i%3 == 0); err != nil {
						t.Fatal(err)
					}
					if err := ref.deliver(uint64(i), i%3 == 0); err != nil {
						t.Fatal(err)
					}
				}
			}
			if clock >= 32 {
				if err := g.expireBefore(uint64(clock - 31)); err != nil {
					t.Fatal(err)
				}
				if err := ref.expireBefore(uint64(clock - 31)); err != nil {
					t.Fatal(err)
				}
			}
		}
		for clock := 0; clock < 288; clock++ {
			if err := g.setClock(uint64(clock)); err != nil {
				t.Fatal(err)
			}
			release(clock)
			if clock < 256 {
				if clock%32 == 0 {
					if err := g.publish(e); err != nil {
						t.Fatal(err)
					}
					if err := ref.publish(e); err != nil {
						t.Fatal(err)
					}
				}
				x := uint16(clock*71) & 511
				got, err := g.predict(uint64(clock), &jointReader{x: x, epoch: 1}, 1)
				if err != nil {
					t.Fatal(err)
				}
				want, err := ref.predict(uint64(clock), &jointReader{x: x, epoch: 1}, 1)
				if err != nil {
					t.Fatal(err)
				}
				adviceNear(t, got.Probability, want.Probability)
				if got.Cost != want.Cost || got.Trace[len(got.Trace)-1].Observed != want.Trace[len(want.Trace)-1].Observed {
					t.Fatal("disabled ablation changed acquisition")
				}
				if delay == 0 && clock%5 != 0 {
					if err := g.deliver(uint64(clock), clock%3 == 0); err != nil {
						t.Fatal(err)
					}
					if err := ref.deliver(uint64(clock), clock%3 == 0); err != nil {
						t.Fatal(err)
					}
				}
			}
			if g.journal.stats != ref.journal.stats || g.received != ref.received || g.journal.core.gate.gate.tests != ref.journal.core.gate.gate.tests || g.journal.core.gate.credits != ref.journal.core.gate.credits {
				t.Fatal("gate or lifecycle changed", clock)
			}
			journalCheck(t, &g.journal)
		}
		if g.journal.stats.Pending != 0 || g.received != 204 {
			t.Fatal("final coverage")
		}
	}
}

func TestAgedAdviceJournalLifecycle(t *testing.T) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	for _, neutral := range []bool{false, true} {
		for _, aged := range []bool{false, true} {
			g := newAgedAdviceJournal(neutral, aged)
			if err := g.publish(e); err != nil {
				t.Fatal(err)
			}
			for origin := uint64(0); origin < 32; origin++ {
				if err := g.setClock(origin); err != nil {
					t.Fatal(err)
				}
				prior := g.advice.weights()
				gate := g.journal.core.gate
				preview, err := adviceRoutedWeights(&gate, origin, [4]float64{.5, .5, .5, .5}, prior)
				if err != nil {
					t.Fatal(err)
				}
				got, err := g.predict(origin, &jointReader{x: uint16(origin*71) & 511, epoch: 1}, 1)
				if err != nil {
					t.Fatal(err)
				}
				last := got.Trace[len(got.Trace)-1]
				want := .5 * preview.Weights[0]
				for j := 0; j < 4; j++ {
					p, err := e.models[j].Forecast(last.Observed, last.Values)
					if err != nil {
						t.Fatal(err)
					}
					want += p * preview.Weights[j+1]
				}
				adviceNear(t, want, got.Probability)
			}
			for origin := uint64(31); origin > 0; origin-- {
				ref := g.journal
				if err := ref.deliver(origin, true); err != nil {
					t.Fatal(err)
				}
				if err := g.deliver(origin, true); err != nil {
					t.Fatal(err)
				}
				if g.journal.core.gate != ref.core.gate || g.journal.stats != ref.stats {
					t.Fatal("gate reference mismatch")
				}
			}
			if g.received != 31 || g.journal.stats.Applied != 0 {
				t.Fatal("arrival blocked")
			}
			if err := g.setClock(32); err != nil {
				t.Fatal(err)
			}
			if err := g.publish(e); err != nil {
				t.Fatal(err)
			}
			advice, gate := g.advice, g.journal.core.gate
			if err := g.expireBefore(32); err != nil {
				t.Fatal(err)
			}
			if g.advice != advice || g.journal.core.gate != gate || g.journal.stats.BankOnly != 31 || g.journal.stats.Censored != 1 {
				t.Fatal("expiry double update/version leak")
			}
			before := *g
			for _, op := range []func() error{func() error { return g.deliver(0, false) }, func() error { return g.deliver(31, false) }, func() error { return g.setClock(33) }, func() error { return g.expireBefore(33) }} {
				if err := op(); err == nil || *g != before {
					t.Fatal("invalid operation mutated journal")
				}
			}
			reader := &jointReader{epoch: 1, failAt: 2}
			reader.hook = func() {
				for _, op := range []func() error{func() error { return g.setClock(32) }, func() error { return g.deliver(31, true) }, func() error { return g.expireBefore(32) }, func() error { return g.publish(e) }} {
					if err := op(); err == nil {
						t.Fatal("reentrant operation accepted")
					}
				}
			}
			if _, err := g.predict(32, reader, 1); err == nil || *g != before {
				t.Fatal("failed acquisition changed state")
			}
			if _, err := g.predict(32, &jointReader{x: 4, epoch: 1}, 1); err != nil {
				t.Fatal(err)
			}
			if math.IsNaN(g.journal.entries[32].served) {
				t.Fatal("NaN served")
			}
			journalCheck(t, &g.journal)
		}
	}
}

func BenchmarkAgedAdviceDelay8Lifetime(b *testing.B) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		g := newAgedAdviceJournal(true, true)
		for clock := 0; clock < 264; clock++ {
			if err := g.setClock(uint64(clock)); err != nil {
				b.Fatal(err)
			}
			if clock >= 8 {
				origin := clock - 8
				if err := g.deliver(uint64(origin), origin%3 == 0); err != nil {
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
		if g.received != 256 || g.journal.stats.Pending != 0 {
			b.Fatal("benchmark accounting")
		}
		journalCheck(b, &g.journal)
	}
}
