package observationlearners

import "testing"

func TestHazardJournalDelayed(t *testing.T) {
	a, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	b, err := newObservationExperts(jointFixture(true, true))
	if err != nil {
		t.Fatal(err)
	}
	g := newHazardAdviceJournal()
	var rows [][4]float64
	var known, ys []bool
	for clock := 0; clock < 288; clock++ {
		if err := g.setClock(uint64(clock)); err != nil {
			t.Fatal(err)
		}
		if clock >= 31 && clock-31 < 256 && (clock-31)%7 != 0 {
			i := clock - 31
			ref := g.journal
			if err := ref.deliver(uint64(i), ys[i]); err != nil {
				t.Fatal(err)
			}
			if err := g.deliver(uint64(i), ys[i]); err != nil {
				t.Fatal(err)
			}
			known[i] = true
			if g.journal != ref {
				t.Fatal("gate evidence duplicated or changed")
			}
		}
		if clock >= 32 {
			if err := g.expireBefore(uint64(clock - 31)); err != nil {
				t.Fatal(err)
			}
		}
		if clock < 256 {
			if clock%32 == 0 {
				e := a
				if clock/32%2 == 1 {
					e = b
				}
				if err := g.publish(e); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := g.predict(uint64(clock), &jointReader{x: uint16(clock), epoch: 1}, 1); err != nil {
				t.Fatal(err)
			}
			rows = append(rows, g.journal.entries[clock%64].bank.Experts)
			known = append(known, false)
			ys = append(ys, clock%3 == 0)
		}
		got := g.advice.weights()
		want, rates := hazardDense(rows, known, ys)
		rateWeights := g.filter.current.rateWeights()
		for h := range rates {
			adviceNear(t, rateWeights[h], rates[h])
		}
		for j := range got {
			adviceNear(t, got[j], want[j])
		}
		journalCheck(t, &g.journal)
	}
	if g.filter.next != 256 || g.filter.base != 256 || g.journal.stats.Pending != 0 {
		t.Fatal("flush invented events or lost settlement")
	}
	before := *g
	if err := g.deliver(0, true); err == nil || *g != before {
		t.Fatal("expired delivery mutation")
	}
}

func TestHazardJournalRollback(t *testing.T) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	g := newHazardAdviceJournal()
	if err := g.publish(e); err != nil {
		t.Fatal(err)
	}
	before := *g
	reader := &jointReader{epoch: 1, failAt: 1}
	reader.hook = func() {
		if err := g.deliver(0, true); err == nil {
			t.Fatal("reentrant delivery")
		}
		if err := g.expireBefore(0); err == nil {
			t.Fatal("reentrant expiry")
		}
		if err := g.publish(e); err == nil {
			t.Fatal("reentrant publication")
		}
		if err := g.setClock(0); err == nil {
			t.Fatal("reentrant clock")
		}
	}
	if _, err := g.predict(0, reader, 1); err == nil || *g != before {
		t.Fatal("acquisition rollback")
	}
	if _, err := g.predict(0, &jointReader{epoch: 1}, 1); err != nil {
		t.Fatal(err)
	}
	// Force a downstream invariant failure to verify two-store atomicity.
	g.filter.next++
	before = *g
	if err := g.deliver(0, true); err == nil || *g != before {
		t.Fatal("filter failure committed gate")
	}
}
