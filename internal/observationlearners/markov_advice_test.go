package observationlearners

import (
	"math"
	"testing"
)

// Enumerate all latent paths, including the transition after the last issue.
// This deliberately uses a dense probability-domain transition, not the filter.
func markovPaths(rows [][4]float64, known, ys []bool, alpha float64) [5]float64 {
	prior := newAgedAdvice(false, false).prior
	var end [5]float64
	var walk func(int, int, float64)
	walk = func(state, step int, mass float64) {
		if step == len(rows) {
			end[state] += mass
			return
		}
		if known[step] {
			p := rows[step][state-1]
			if !ys[step] {
				p = 1 - p
			}
			mass *= p
		}
		for next := 1; next < 5; next++ {
			transition := alpha * prior[next]
			if next == state {
				transition += 1 - alpha
			}
			walk(next, step+1, mass*transition)
		}
	}
	for state := 1; state < 5; state++ {
		walk(state, 0, prior[state])
	}
	sum := 0.
	for _, v := range end {
		sum += v
	}
	for j := range end {
		end[j] /= sum
	}
	return end
}

func TestMarkovAdvicePaths(t *testing.T) {
	rows := [][4]float64{{.1, .8, .3, .6}, {.6, .2, .8, .4}, {.2, .7, .3, .8}, {.9, .1, .5, .4}}
	ys := []bool{true, false, true, false}
	for _, alpha := range []float64{0, .001, .2, 1} {
		var finals [][5]float64
		for _, order := range [][]int{{0, 1, 2, 3}, {3, 2, 1, 0}, {1, 3, 0, 2}} {
			f, err := newMarkovAdvice(alpha)
			if err != nil {
				t.Fatal(err)
			}
			known := make([]bool, len(rows))
			for i, row := range rows {
				if err := f.issue(uint64(i), row); err != nil {
					t.Fatal(err)
				}
			}
			for _, i := range order {
				if err := f.deliver(uint64(i), ys[i]); err != nil {
					t.Fatal(err)
				}
				known[i] = true
				want, got := markovPaths(rows, known, ys, alpha), f.current.weights()
				for j := range got {
					adviceNear(t, got[j], want[j])
				}
			}
			if f.base != 4 || f.next != 4 {
				t.Fatal("settled frontier")
			}
			finals = append(finals, f.current.weights())
		}
		for _, w := range finals {
			for j := range w {
				adviceNear(t, w[j], finals[0][j])
			}
		}
	}
}

// Dense full-history reference verifies checkpointing independently of the ring.
func markovDense(rows [][4]float64, known, ys []bool, alpha float64) [5]float64 {
	prior := newAgedAdvice(false, false).prior
	w := prior
	for i, row := range rows {
		sum := 0.
		for j := 1; j < 5; j++ {
			if known[i] {
				p := row[j-1]
				if !ys[i] {
					p = 1 - p
				}
				w[j] *= p
			}
			sum += w[j]
		}
		var next [5]float64
		for j := 1; j < 5; j++ {
			for k := 1; k < 5; k++ {
				transition := alpha * prior[k]
				if j == k {
					transition += 1 - alpha
				}
				next[k] += transition * w[j] / sum
			}
		}
		w = next
	}
	return w
}

func TestMarkovAdviceCheckpoint(t *testing.T) {
	f, err := newMarkovAdvice(.001)
	if err != nil {
		t.Fatal(err)
	}
	var rows [][4]float64
	var known, ys []bool
	check := func() {
		t.Helper()
		got, want := f.current.weights(), markovDense(rows, known, ys, .001)
		for j := range got {
			adviceNear(t, got[j], want[j])
		}
	}
	for block := 0; block < 4; block++ {
		for i := block * 64; i < (block+1)*64; i++ {
			row := [4]float64{.1 + .1*float64(i%7), .2, .8, .7}
			if err := f.issue(uint64(i), row); err != nil {
				t.Fatal(err)
			}
			rows = append(rows, row)
			known = append(known, false)
			ys = append(ys, i%3 == 0)
			check()
		}
		before := f
		if err := f.issue(f.next, [4]float64{.5, .5, .5, .5}); err == nil || f != before {
			t.Fatal("backpressure mutation")
		}
		for i := (block+1)*64 - 1; i >= block*64; i-- {
			if i%7 == 0 {
				continue
			}
			if err := f.deliver(uint64(i), ys[i]); err != nil {
				t.Fatal(err)
			}
			known[i] = true
			check()
			before = f
			if err := f.deliver(uint64(i), ys[i]); err == nil || f != before {
				t.Fatal("duplicate mutation")
			}
		}
		if err := f.expireBefore(f.next); err != nil {
			t.Fatal(err)
		}
		check()
		if f.base != f.next {
			t.Fatal("checkpoint not advanced")
		}
		before = f
		if err := f.deliver(uint64(block*64), true); err == nil || f != before {
			t.Fatal("committed label mutated")
		}
	}
}

func TestMarkovAdviceAtomicAndUnderflow(t *testing.T) {
	for _, alpha := range []float64{-1, 1.1, math.NaN(), math.Inf(1)} {
		if _, err := newMarkovAdvice(alpha); err == nil {
			t.Fatal("invalid transition")
		}
	}
	f, _ := newMarkovAdvice(0)
	for _, p := range []float64{0, 1, -1, math.NaN(), math.Inf(1)} {
		before := f
		if err := f.issue(0, [4]float64{.5, .5, .5, p}); err == nil || f != before {
			t.Fatal("invalid emission mutation")
		}
	}
	for i := uint64(0); i < 240; i++ {
		if err := f.issue(i, [4]float64{.001, .999, .5, .5}); err != nil {
			t.Fatal(err)
		}
		if err := f.deliver(i, i < 120); err != nil {
			t.Fatal(err)
		}
		if i == 119 && f.current.weights()[1] != 0 {
			t.Fatal("underflow not reached")
		}
	}
	if math.Abs(f.current.logs[1]-f.current.logs[2]-math.Log(57)) > 1e-9 {
		t.Fatal("log evidence lost")
	}
	before := f
	if err := f.deliver(250, true); err == nil || f != before {
		t.Fatal("future label mutation")
	}
	if err := f.expireBefore(250); err == nil || f != before {
		t.Fatal("future expiry mutation")
	}
}

func TestMarkovJournalImmediateIdentity(t *testing.T) {
	a, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	b, err := newObservationExperts(jointFixture(true, true))
	if err != nil {
		t.Fatal(err)
	}
	g, ref := newMarkovAdviceJournal(), newLogAdviceJournal(false)
	for i := uint64(0); i < 256; i++ {
		if err := g.setClock(i); err != nil {
			t.Fatal(err)
		}
		if err := ref.setClock(i); err != nil {
			t.Fatal(err)
		}
		if i%32 == 0 {
			e := a
			if i/32%2 == 1 {
				e = b
			}
			if err := g.publish(e); err != nil {
				t.Fatal(err)
			}
			if err := ref.publish(e); err != nil {
				t.Fatal(err)
			}
		}
		x, err := g.predict(i, &jointReader{x: uint16(i), epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		y, err := ref.predict(i, &jointReader{x: uint16(i), epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		compareJointTrace(t, x, y)
		if err := g.deliver(i, i%3 == 0); err != nil {
			t.Fatal(err)
		}
		if err := ref.deliver(i, i%3 == 0); err != nil {
			t.Fatal(err)
		}
		if g.agedAdviceJournal != *ref.agedAdviceJournal {
			t.Fatal("immediate state mismatch", i)
		}
	}
}

func TestMarkovJournalDelayed(t *testing.T) {
	a, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	b, err := newObservationExperts(jointFixture(true, true))
	if err != nil {
		t.Fatal(err)
	}
	g := newMarkovAdviceJournal()
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
		got, want := g.advice.weights(), markovDense(rows, known, ys, .001)
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

func TestMarkovJournalRollback(t *testing.T) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	g := newMarkovAdviceJournal()
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
