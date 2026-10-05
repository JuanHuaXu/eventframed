package observationlearners

import (
	"math"
	"testing"
)

// Literal path enumeration is deliberately independent of the recursive
// selector. The terminal state is after the final share transition.
func logAdvicePaths(prior [5]float64, rows [][4]float64, ys []bool, share float64) [5]float64 {
	var end [5]float64
	var walk func(int, int, float64)
	walk = func(state, step int, mass float64) {
		if step == len(rows) {
			end[state] += mass
			return
		}
		p := .5
		if state > 0 {
			p = rows[step][state-1]
		}
		if !ys[step] {
			p = 1 - p
		}
		mass *= p
		for next := 0; next < 5; next++ {
			transition := share * prior[next]
			if next == state {
				transition += 1 - share
			}
			walk(next, step+1, mass*transition)
		}
	}
	for j, p := range prior {
		walk(j, 0, p)
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

func TestLogAdviceHiddenPaths(t *testing.T) {
	rows := [][4]float64{{.95, .1, .4, .8}, {.1, .8, .2, .5}, {.2, .3, .9, .1}, {.7, .1, .6, .8}, {.4, .9, .5, .2}}
	ys := []bool{false, true, true, false, true}
	for _, neutral := range []bool{false, true} {
		for _, share := range []float64{0, .001, .2} {
			s := newAgedAdvice(neutral, false)
			for i, row := range rows {
				if err := s.advance(uint64(i)); err != nil {
					t.Fatal(err)
				}
				if err := observeLogAdvice(&s, uint64(i), row, ys[i], share); err != nil {
					t.Fatal(err)
				}
				want, got := logAdvicePaths(s.prior, rows[:i+1], ys[:i+1], share), s.weights()
				for j := range want {
					adviceNear(t, got[j], want[j])
				}
			}
		}
	}
}

func TestLogAdviceCommutativityAndUnderflow(t *testing.T) {
	rows := [][4]float64{{.01, .9, .3, .8}, {.8, .1, .7, .2}, {.2, .6, .9, .1}}
	ys := []bool{true, false, true}
	a, b := newAgedAdvice(true, false), newAgedAdvice(true, false)
	a.advance(3)
	b.advance(3)
	for i := 0; i < 3; i++ {
		if err := observeLogAdvice(&a, uint64(i), rows[i], ys[i], 0); err != nil {
			t.Fatal(err)
		}
		j := 2 - i
		if err := observeLogAdvice(&b, uint64(j), rows[j], ys[j], 0); err != nil {
			t.Fatal(err)
		}
	}
	wa, wb := a.weights(), b.weights()
	for j := range wa {
		adviceNear(t, wa[j], wb[j])
	}
	s := newAgedAdvice(false, false)
	p := [4]float64{.001, .999, .5, .5}
	for i := uint64(0); i < 240; i++ {
		s.advance(i)
		if err := observeLogAdvice(&s, i, p, i < 120, 0); err != nil {
			t.Fatal(err)
		}
		if i == 119 && s.weights()[1] != 0 {
			t.Fatal("underflow fixture did not exercise probability underflow")
		}
	}
	// Their final likelihoods are equal despite an intermediate exp underflow.
	if math.IsInf(s.logs[1], 0) || math.IsInf(s.logs[2], 0) || math.Abs((s.logs[1]-s.logs[2])-math.Log(57)) > 1e-9 {
		t.Fatal("recoverable log evidence lost")
	}
}

func TestLogAdviceInputAtomicity(t *testing.T) {
	s := newAgedAdvice(true, false)
	before := s
	for _, p := range []float64{0, 1, -.1, 1.1, math.NaN(), math.Inf(1)} {
		if err := observeLogAdvice(&s, 0, [4]float64{.1, .2, .3, p}, true, .001); err == nil || s != before {
			t.Fatal("invalid probability mutation")
		}
	}
	for _, share := range []float64{-1, 1, math.NaN(), math.Inf(1)} {
		if err := observeLogAdvice(&s, 0, [4]float64{.1, .2, .3, .4}, true, share); err == nil || s != before {
			t.Fatal("invalid share mutation")
		}
	}
	if err := observeLogAdvice(&s, 1, [4]float64{.1, .2, .3, .4}, true, .001); err == nil || s != before {
		t.Fatal("future origin mutation")
	}
	aged := newAgedAdvice(true, true)
	if err := observeLogAdvice(&aged, 0, [4]float64{.1, .2, .3, .4}, true, .001); err == nil {
		t.Fatal("mixed score/age contract accepted")
	}
}

func TestLogAdviceJournal(t *testing.T) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	for _, neutral := range []bool{false, true} {
		g := newLogAdviceJournal(neutral)
		if err := g.publish(e); err != nil {
			t.Fatal(err)
		}
		for i := uint64(0); i < 32; i++ {
			if err := g.setClock(i); err != nil {
				t.Fatal(err)
			}
			if _, err := g.predict(i, &jointReader{x: uint16(i*17) & 511, epoch: 1}, 1); err != nil {
				t.Fatal(err)
			}
		}
		for i := uint64(31); i > 0; i-- {
			ref := g.journal
			advice := g.advice
			entry := ref.entries[i]
			if err := ref.deliver(i, true); err != nil {
				t.Fatal(err)
			}
			if err := observeLogAdvice(&advice, i, entry.bank.Experts, true, .001); err != nil {
				t.Fatal(err)
			}
			if err := g.deliver(i, true); err != nil {
				t.Fatal(err)
			}
			if g.journal != ref || g.advice != advice {
				t.Fatal("delivery routing mismatch")
			}
		}
		if g.received != 31 || g.journal.stats.Applied != 0 {
			t.Fatal("blocked advice update")
		}
		g.setClock(32)
		if err := g.publish(e); err != nil {
			t.Fatal(err)
		}
		advice, gate := g.advice, g.journal.core.gate
		if err := g.expireBefore(32); err != nil {
			t.Fatal(err)
		}
		if advice != g.advice || gate != g.journal.core.gate || g.journal.stats.BankOnly != 31 {
			t.Fatal("expiry doubled evidence or updated new tests")
		}
		before := *g.agedAdviceJournal
		for _, i := range []uint64{0, 31, 32} {
			if err := g.deliver(i, true); err == nil || *g.agedAdviceJournal != before {
				t.Fatal("invalid/duplicate delivery mutation")
			}
		}
		reader := &jointReader{epoch: 1, failAt: 2}
		reader.hook = func() {
			if err := g.deliver(31, true); err == nil {
				t.Fatal("reentrant delivery accepted")
			}
		}
		if _, err := g.predict(32, reader, 1); err == nil || *g.agedAdviceJournal != before {
			t.Fatal("failed acquisition mutation")
		}
		if _, err := g.predict(32, &jointReader{x: 4, epoch: 1}, 1); err != nil {
			t.Fatal(err)
		}
		journalCheck(t, &g.journal)
	}
}
