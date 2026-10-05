package observationlearners

import (
	"math"
	"reflect"
	"testing"
)

func TestRoleSnapshotOriginalBudget(t *testing.T) {
	e0, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	e1, err := newObservationExperts(jointFixture(true, true))
	if err != nil {
		t.Fatal(err)
	}
	for _, delay := range []int{0, 31} {
		g, err := newRoleSnapshotJournal(6400)
		if err != nil {
			t.Fatal(err)
		}
		old := newMarkovAdviceJournal()
		var arrived [256]bool
		for clock := 0; clock < 288; clock++ {
			if err := g.setClock(uint64(clock)); err != nil {
				t.Fatal(err)
			}
			if err := old.setClock(uint64(clock)); err != nil {
				t.Fatal(err)
			}
			for origin := clock - 1; origin >= 0; origin-- {
				if origin >= 256 || arrived[origin] || delay > 0 && origin%5 == 0 || origin+delay > clock {
					continue
				}
				if err := g.deliver(uint64(origin), origin%3 == 0); err != nil {
					t.Fatal(err)
				}
				if err := old.deliver(uint64(origin), origin%3 == 0); err != nil {
					t.Fatal(err)
				}
				arrived[origin] = true
			}
			if clock >= 32 {
				cutoff := clock - 31
				if cutoff > 256 {
					cutoff = 256
				}
				if err := g.expireBefore(uint64(cutoff)); err != nil {
					t.Fatal(err)
				}
				if err := old.expireBefore(uint64(cutoff)); err != nil {
					t.Fatal(err)
				}
			}
			if g.stats != old.journal.stats || g.received != old.received {
				t.Fatal("control accounting", clock)
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
					t.Fatal(err)
				}
				if err := old.publish(e); err != nil {
					t.Fatal(err)
				}
			}
			x := uint16(clock * 17 % 512)
			got, err := g.predict(uint64(clock), &jointReader{x: x, epoch: 1}, 1)
			if err != nil {
				t.Fatal(err)
			}
			want, err := old.predict(uint64(clock), &jointReader{x: x, epoch: 1}, 1)
			if err != nil {
				t.Fatal(err)
			}
			if got.Cost != want.Cost || got.Stop != want.Stop || len(got.Trace) != len(want.Trace) || math.Abs(got.Probability-want.Probability) > 1e-12 {
				t.Fatal("control prediction", delay, clock, got.Probability, want.Probability)
			}
			for i, a := range got.Trace {
				b := want.Trace[i]
				if a.View != b.View || a.Observed != b.Observed || a.Values != b.Values || math.Abs(a.Probability-b.Probability) > 1e-12 {
					t.Fatal("control acquisition", delay, clock, i)
				}
			}
			if delay == 0 {
				if err := g.deliver(uint64(clock), clock%3 == 0); err != nil {
					t.Fatal(err)
				}
				if err := old.deliver(uint64(clock), clock%3 == 0); err != nil {
					t.Fatal(err)
				}
				arrived[clock] = true
			}
		}
	}
}

func TestRoleSnapshotBoundaryAndRollback(t *testing.T) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	g := newMatchedSnapshotJournal()
	if err := g.publish(e); err != nil {
		t.Fatal(err)
	}
	for i := uint64(0); i < 32; i++ {
		if err := g.setClock(i); err != nil {
			t.Fatal(err)
		}
		if _, err := g.predict(i, &jointReader{x: 0, epoch: 1}, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.deliver(i, true); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.setClock(32); err != nil {
		t.Fatal(err)
	}
	before := *g
	r := &snapshotReentrantReader{inner: jointReader{x: 0, epoch: 1}}
	if _, err := g.predict(32, r, 1); err == nil || r.calls != 0 || !reflect.DeepEqual(*g, before) {
		t.Fatal("publication bypass")
	}
	if err := g.publish(e); err != nil {
		t.Fatal(err)
	}
	before = *g
	r = &snapshotReentrantReader{inner: jointReader{x: 0, epoch: 1}, fail: true}
	r.callback = func() {
		if err := g.expireBefore(32); err == nil {
			t.Fatal("reentrancy")
		}
	}
	if _, err := g.predict(32, r, 1); err == nil || !reflect.DeepEqual(*g, before) {
		t.Fatal("read failure rollback")
	}
	if g.gate.boundary != snapshotGateBoundary {
		t.Fatal("unmatched budget")
	}
}
