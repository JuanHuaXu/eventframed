package observationlearners

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func TestSnapshotJournalFullLifecycle(t *testing.T) {
	g := newSnapshotJournal()
	pub0, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	pub1, err := newObservationExperts(jointFixture(true, true))
	if err != nil {
		t.Fatal(err)
	}
	var events []snapshotAdviceEvent
	var index [256]int
	var state [256]int
	var want routedJournalStats
	var counts [8]int
	drain, version := 0, -1
	check := func() {
		t.Helper()
		for drain < int(want.Issued) && state[drain] >= 2 {
			if state[drain] == 3 {
				want.Censored++
			} else if drain/32 == version {
				want.Applied++
				for i, id := range g.gate.ids {
					if id != 0 {
						counts[i]++
					}
				}
			} else {
				want.BankOnly++
			}
			want.Pending--
			drain++
		}
		if want != g.stats || uint64(drain) != g.drainAt {
			t.Fatal("independent journal accounting", want, g.stats, drain, g.drainAt)
		}
		for i, c := range counts {
			if c != g.gate.tests[i].count {
				t.Fatal("gate window count", i, c, g.gate.tests[i].count)
			}
		}
		if len(events) > 0 {
			snapshotAssertWeights(t, g.bank.filter.current, snapshotDenseReference(events, .001))
		}
	}
	for clock := 0; clock < 288; clock++ {
		if err := g.setClock(uint64(clock)); err != nil {
			t.Fatal(err)
		}
		for origin := clock - 1; origin >= 0; origin-- {
			if origin >= 256 || state[origin] != 1 || origin%5 == 0 || origin+origin%32 > clock {
				continue
			}
			if err := g.deliver(uint64(origin), origin%3 == 0); err != nil {
				t.Fatal(clock, origin, err)
			}
			state[origin] = 2
			events[index[origin]].known = true
			events[index[origin]].outcome = origin%3 == 0
			check()
		}
		if clock >= 32 {
			cutoff := clock - 31
			if cutoff > 256 {
				cutoff = 256
			}
			if err := g.expireBefore(uint64(cutoff)); err != nil {
				t.Fatal(err)
			}
			for origin := 0; origin < cutoff; origin++ {
				if state[origin] == 1 {
					state[origin] = 3
					events[index[origin]].censored = true
				}
			}
			check()
		}
		if clock >= 256 {
			continue
		}
		if clock%32 == 0 {
			e := pub0
			if clock%64 == 32 {
				e = pub1
			}
			if err := g.publish(e); err != nil {
				t.Fatal(err)
			}
			version = clock / 32
			counts = [8]int{}
			events = append(events, snapshotAdviceEvent{publication: true, origin: uint64(clock), version: version})
			check()
		}
		base, err := g.bank.snapshot()
		if err != nil {
			t.Fatal(err)
		}
		weights, err := g.gate.route(base.ids, base.weights)
		if err != nil {
			t.Fatal(err)
		}
		x := uint16(clock * 17 % 512)
		r, err := g.predict(uint64(clock), &jointReader{x: x, epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		last := r.Trace[len(r.Trace)-1]
		raw, err := base.raw(last.Observed, last.Values)
		if err != nil {
			t.Fatal(err)
		}
		p := .5 * weights[0]
		for i, v := range raw {
			p += weights[i+1] * v
			if g.gate.tests[i].rejected && weights[i+1] != 0 {
				t.Fatal("rejected model in served law")
			}
		}
		if math.Abs(p-r.Probability) > 1e-12 {
			t.Fatal("issued law mismatch")
		}
		index[clock] = len(events)
		events = append(events, snapshotAdviceEvent{origin: uint64(clock), ids: base.ids, raw: raw})
		state[clock] = 1
		want.Issued++
		want.Pending++
		check()
		if clock%32 == 0 && clock%5 != 0 {
			if err := g.deliver(uint64(clock), clock%3 == 0); err != nil {
				t.Fatal(err)
			}
			state[clock] = 2
			events[index[clock]].known = true
			events[index[clock]].outcome = clock%3 == 0
			check()
		}
	}
	if g.stats.Pending != 0 || g.stats.BankOnly == 0 || g.stats.Censored == 0 || g.received != g.stats.Applied+g.stats.BankOnly {
		t.Fatal("coverage/accounting")
	}
}

func TestSnapshotObserverFirstBankCompatibility(t *testing.T) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	g := newSnapshotJournal()
	if err := g.publish(e); err != nil {
		t.Fatal(err)
	}
	for t0 := uint64(0); t0 < 32; t0++ {
		if err := g.setClock(t0); err != nil {
			t.Fatal(err)
		}
		base, _ := g.bank.snapshot()
		w, err := g.gate.route(base.ids, base.weights)
		if err != nil {
			t.Fatal(err)
		}
		var oldWeights [5]float64
		copy(oldWeights[:], w[:5])
		old, err := e.snapshot(oldWeights)
		if err != nil {
			t.Fatal(err)
		}
		x := uint16(t0 * 13 % 512)
		want, err := runJointObserver(old, &jointReader{x: x, epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		got, err := g.predict(t0, &jointReader{x: x, epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		if got.Cost != want.Cost || got.Stop != want.Stop || len(got.Trace) != len(want.Trace) {
			t.Fatal("observer shape changed")
		}
		for i, a := range got.Trace {
			b := want.Trace[i]
			if a.View != b.View || a.Observed != b.Observed || a.Values != b.Values || math.Abs(a.Probability-b.Probability) > 1e-12 {
				t.Fatal("observer trace changed")
			}
		}
		if err := g.deliver(t0, t0%3 == 0); err != nil {
			t.Fatal(err)
		}
	}
}

type snapshotReentrantReader struct {
	inner    jointReader
	callback func()
	fail     bool
	calls    int
}

func (r *snapshotReentrantReader) Epoch() uint64 { return r.inner.Epoch() }
func (r *snapshotReentrantReader) Read(v observation.View) (uint16, uint16, error) {
	r.calls++
	if r.callback != nil {
		r.callback()
	}
	if r.fail {
		return 0, 0, errors.New("injected read failure")
	}
	return r.inner.Read(v)
}

func TestSnapshotJournalRollbackAndReentrancy(t *testing.T) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	g := newSnapshotJournal()
	if err := g.publish(e); err != nil {
		t.Fatal(err)
	}
	before := *g
	calls := 0
	r := &snapshotReentrantReader{inner: jointReader{x: 0, epoch: 1}, fail: true}
	r.callback = func() {
		calls++
		for _, op := range []func() error{func() error { return g.publish(e) }, func() error { return g.setClock(0) }, func() error { return g.deliver(0, true) }, func() error { return g.expireBefore(0) }, func() error { _, err := g.predict(0, &jointReader{x: 0, epoch: 1}, 1); return err }} {
			if err := op(); err == nil {
				t.Fatal("reentrant mutation admitted")
			}
		}
	}
	if _, err := g.predict(0, r, 1); err == nil || calls == 0 || !reflect.DeepEqual(*g, before) {
		t.Fatal("reader rollback")
	}
	// Failure after observation must roll back the journal as well as the filter.
	g.bank.filter.nextOrigin = 1
	before = *g
	r = &snapshotReentrantReader{inner: jointReader{x: 0, epoch: 1}}
	if _, err := g.predict(0, r, 1); err == nil || r.calls == 0 || !reflect.DeepEqual(*g, before) {
		t.Fatal("post-read filter failure rollback")
	}
	g.bank.filter.nextOrigin = 0
	if _, err := g.predict(0, &jointReader{x: 0, epoch: 1}, 1); err != nil {
		t.Fatal(err)
	}
	// A bad gate owner is detected after advice delivery, without committing it.
	g.entries[0].ids[0] = 9
	before = *g
	if err := g.deliver(0, true); err == nil || !reflect.DeepEqual(*g, before) {
		t.Fatal("downstream gate failure rollback")
	}
	g.entries[0].ids[0] = 1
	if err := g.deliver(0, true); err != nil {
		t.Fatal(err)
	}
}

func TestSnapshotJournalLateGateOwnership(t *testing.T) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	g := newSnapshotJournal()
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
	}
	if err := g.setClock(32); err != nil {
		t.Fatal(err)
	}
	if err := g.publish(e); err != nil {
		t.Fatal(err)
	}
	gate := g.gate
	if err := g.deliver(0, true); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gate, g.gate) || g.stats.BankOnly != 1 {
		t.Fatal("old evidence changed new gate")
	}
	if err := g.expireBefore(32); err != nil {
		t.Fatal(err)
	}
	if _, err := g.predict(32, &jointReader{x: 0, epoch: 1}, 1); err != nil {
		t.Fatal(err)
	}
	if err := g.deliver(32, true); err != nil {
		t.Fatal(err)
	}
	for _, test := range g.gate.tests {
		if test.count != 1 {
			t.Fatal("current evidence not admitted")
		}
	}
}
