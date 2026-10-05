package observationlearners

import (
	"math"
	"reflect"
	"testing"
)

// Independent probability-space transitions over generation IDs, not ring slots.
func snapshotReferenceBranches(id uint64, e snapshotAdviceEvent, alpha float64) map[uint64]float64 {
	out := map[uint64]float64{}
	if e.publication {
		if e.version == 0 {
			for j, p := range snapshotRolePrior {
				out[uint64(j+1)] = p
			}
			return out
		}
		rho := 1 / float64(e.version+1)
		mass := 1.
		if int((id-1)/4) == e.version-1 {
			out[id] = 1 - rho
			mass = rho
		}
		for j, p := range snapshotRolePrior {
			out[uint64(4*e.version+j+1)] = mass * p
		}
		return out
	}
	out[id] = 1 - alpha
	n := 0
	for _, key := range e.ids {
		if key != 0 {
			n++
		}
	}
	for _, key := range e.ids {
		if key != 0 {
			out[key] += alpha * snapshotRolePrior[(key-1)%4] / float64(n/4)
		}
	}
	return out
}

func snapshotReferenceEmission(id uint64, e snapshotAdviceEvent) float64 {
	if e.publication || !e.known {
		return 1
	}
	p := e.raw[(id-1)%8]
	if e.outcome {
		return p
	}
	return 1 - p
}

func snapshotDenseReference(events []snapshotAdviceEvent, alpha float64) [8]float64 {
	w := map[uint64]float64{0: 1}
	for _, e := range events {
		next := map[uint64]float64{}
		for id, mass := range w {
			for key, p := range snapshotReferenceBranches(id, e, alpha) {
				next[key] += mass * snapshotReferenceEmission(id, e) * p
			}
		}
		total := 0.
		for _, mass := range next {
			total += mass
		}
		for key, mass := range next {
			next[key] = mass / total
		}
		w = next
	}
	var out [8]float64
	for id, mass := range w {
		out[(id-1)%8] = mass
	}
	return out
}

// Enumerate complete paths without merging equivalent intermediate states.
func snapshotPathReference(events []snapshotAdviceEvent, alpha float64) [8]float64 {
	var out [8]float64
	var visit func(int, uint64, float64)
	visit = func(i int, id uint64, mass float64) {
		if i == len(events) {
			out[(id-1)%8] += mass
			return
		}
		e := events[i]
		for key, p := range snapshotReferenceBranches(id, e, alpha) {
			if p > 0 {
				visit(i+1, key, mass*snapshotReferenceEmission(id, e)*p)
			}
		}
	}
	visit(0, 0, 1)
	total := 0.
	for _, p := range out {
		total += p
	}
	for i := range out {
		out[i] /= total
	}
	return out
}

func snapshotAssertWeights(t *testing.T, s snapshotAdviceState, want [8]float64) {
	t.Helper()
	got, err := s.weights()
	if err != nil {
		t.Fatal(err)
	}
	for i := range got {
		if math.Abs(got[i]-want[i]) > 1e-11 {
			t.Fatalf("slot %d: %.17g != %.17g", i, got[i], want[i])
		}
	}
}

func snapshotRaw(ids [8]uint64, origin uint64) [8]float64 {
	var out [8]float64
	for i, id := range ids {
		if id != 0 {
			out[i] = .1 + .1*float64((id+origin)%8)
		}
	}
	return out
}

func TestSnapshotAdvicePaths(t *testing.T) {
	for _, alpha := range []float64{0, .001, .2, 1} {
		var events []snapshotAdviceEvent
		s := newSnapshotAdviceState()
		for v := 0; v < 3; v++ {
			pub := snapshotAdviceEvent{publication: true, version: v, origin: uint64(32 * v)}
			if err := snapshotAdviceStep(&s, pub, alpha); err != nil {
				t.Fatal(err)
			}
			events = append(events, pub)
			e := snapshotAdviceEvent{origin: uint64(32 * v), ids: s.ids, raw: snapshotRaw(s.ids, uint64(v))}
			if err := snapshotAdviceStep(&s, e, alpha); err != nil {
				t.Fatal(err)
			}
			events = append(events, e)
		}
		for _, order := range [][]int{{1, 3, 5}, {5, 1, 3}, {3, 5, 1}} {
			es := append([]snapshotAdviceEvent(nil), events...)
			for _, index := range order {
				es[index].known = true
				es[index].outcome = index != 3
				state := newSnapshotAdviceState()
				for _, e := range es {
					if err := snapshotAdviceStep(&state, e, alpha); err != nil {
						t.Fatal(err)
					}
				}
				snapshotAssertWeights(t, state, snapshotPathReference(es, alpha))
			}
		}
	}
}

func TestSnapshotAdviceLifecycle(t *testing.T) {
	f, _ := newSnapshotAdvice(.001)
	var events []snapshotAdviceEvent
	var issueIndex [256]int
	var arrived [256]bool
	check := func() {
		t.Helper()
		snapshotAssertWeights(t, f.current, snapshotDenseReference(events, .001))
		if f.tail-f.head > 80 || f.nextOrigin-f.baseOrigin > 64 {
			t.Fatal("capacity")
		}
	}
	for clock := 0; clock < 288; clock++ {
		for i := clock - 1; i >= 0; i-- {
			if i >= 256 || arrived[i] || i%5 == 0 || i+(i%32) > clock {
				continue
			}
			if err := f.deliver(uint64(i), i%3 == 0); err != nil {
				t.Fatal(clock, i, err)
			}
			events[issueIndex[i]].known = true
			events[issueIndex[i]].outcome = i%3 == 0
			arrived[i] = true
			check()
		}
		if clock >= 32 {
			cutoff := uint64(clock - 31)
			if cutoff > f.nextOrigin {
				cutoff = f.nextOrigin
			}
			if err := f.expireBefore(cutoff); err != nil {
				t.Fatal(err)
			}
			for i := range events {
				e := &events[i]
				if !e.publication && !e.known && e.origin < cutoff {
					e.censored = true
				}
			}
			check()
		}
		if clock >= 256 {
			continue
		}
		if clock%32 == 0 {
			if err := f.publish(clock / 32); err != nil {
				t.Fatal(err)
			}
			events = append(events, snapshotAdviceEvent{publication: true, version: clock / 32, origin: uint64(clock)})
			check()
		}
		ids := f.current.ids
		raw := snapshotRaw(ids, uint64(clock))
		if err := f.issue(uint64(clock), ids, raw); err != nil {
			t.Fatal(err)
		}
		issueIndex[clock] = len(events)
		events = append(events, snapshotAdviceEvent{origin: uint64(clock), ids: ids, raw: raw})
		check()
		if clock%32 == 0 && clock%5 != 0 {
			if err := f.deliver(uint64(clock), clock%3 == 0); err != nil {
				t.Fatal(err)
			}
			events[issueIndex[clock]].known = true
			events[issueIndex[clock]].outcome = clock%3 == 0
			arrived[clock] = true
			check()
		}
	}
	if f.baseOrigin != 256 || f.head != f.tail {
		t.Fatal("unsettled suffix")
	}
}

func TestSnapshotAdviceImmediateLimit(t *testing.T) {
	for _, alpha := range []float64{0, .001, .2, 1} {
		f, _ := newSnapshotAdvice(alpha)
		m, _ := newMarkovAdvice(alpha)
		if err := f.publish(0); err != nil {
			t.Fatal(err)
		}
		for i := uint64(0); i < 32; i++ {
			w := m.current.weights()
			var want [8]float64
			copy(want[:4], w[1:])
			snapshotAssertWeights(t, f.current, want)
			raw := snapshotRaw(f.current.ids, i)
			var old [4]float64
			copy(old[:], raw[:4])
			if err := f.issue(i, f.current.ids, raw); err != nil {
				t.Fatal(err)
			}
			if err := m.issue(i, old); err != nil {
				t.Fatal(err)
			}
			if err := f.deliver(i, i%3 == 0); err != nil {
				t.Fatal(err)
			}
			if err := m.deliver(i, i%3 == 0); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestSnapshotAdviceAtomicAndRetiredDelivery(t *testing.T) {
	f, _ := newSnapshotAdvice(.001)
	reject := func(op func() error) {
		t.Helper()
		before := f
		if err := op(); err == nil {
			t.Fatal("expected rejection")
		}
		if !reflect.DeepEqual(f, before) {
			t.Fatal("failed operation changed state")
		}
	}
	reject(func() error { return f.issue(0, [8]uint64{}, [8]float64{}) })
	if err := f.publish(0); err != nil {
		t.Fatal(err)
	}
	reject(func() error { return f.publish(0) })
	bad := f.current.ids
	bad[0] = 9
	reject(func() error { return f.issue(0, bad, snapshotRaw(bad, 0)) })
	for _, p := range []float64{math.NaN(), math.Inf(1), 0, 1} {
		raw := snapshotRaw(f.current.ids, 0)
		raw[0] = p
		reject(func() error { return f.issue(0, f.current.ids, raw) })
	}
	for i := uint64(0); i < 64; i++ {
		if i == 32 {
			if err := f.publish(1); err != nil {
				t.Fatal(err)
			}
		}
		if err := f.issue(i, f.current.ids, snapshotRaw(f.current.ids, i)); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.publish(2); err != nil {
		t.Fatal(err)
	}
	if f.current.ids[0] != 9 {
		t.Fatal("slot not reused")
	}
	reject(func() error { return f.issue(64, f.current.ids, snapshotRaw(f.current.ids, 64)) })
	if err := f.deliver(0, true); err != nil {
		t.Fatal("late retired evidence must refilter", err)
	}
	if f.current.ids[0] != 9 {
		t.Fatal("historical identity escaped")
	}
	reject(func() error { return f.deliver(0, false) })
	reject(func() error { return f.deliver(64, true) })
	reject(func() error { return f.expireBefore(65) })
	if err := f.issue(64, f.current.ids, snapshotRaw(f.current.ids, 64)); err != nil {
		t.Fatal(err)
	}
	if err := f.expireBefore(64); err != nil {
		t.Fatal(err)
	}
	reject(func() error { return f.deliver(1, true) })
	if err := f.deliver(64, true); err != nil {
		t.Fatal(err)
	}
	if f.baseOrigin != 65 || f.head != f.tail {
		t.Fatal("drain")
	}
}

func TestSnapshotAdviceRecoverableUnderflow(t *testing.T) {
	f, _ := newSnapshotAdvice(0)
	if err := f.publish(0); err != nil {
		t.Fatal(err)
	}
	for i := uint64(0); i < 4; i++ {
		raw := [8]float64{1e-250, .5, .5, .5}
		if i >= 2 {
			raw = [8]float64{.5, 1e-250, 1e-250, 1e-250}
		}
		if err := f.issue(i, f.current.ids, raw); err != nil {
			t.Fatal(err)
		}
		if err := f.deliver(i, true); err != nil {
			t.Fatal(err)
		}
	}
	var want [8]float64
	copy(want[:4], snapshotRolePrior[:])
	snapshotAssertWeights(t, f.current, want)
}
