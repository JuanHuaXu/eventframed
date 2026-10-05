package observationlearners

import (
	"math"
	"testing"
)

func TestCompatibilityGeometry(t *testing.T) {
	for _, uniform := range []bool{false, true} {
		a, err := newObservationExperts(jointFixture(uniform, false))
		if err != nil {
			t.Fatal(err)
		}
		b, err := newObservationExperts(jointFixture(uniform, true))
		if err != nil {
			t.Fatal(err)
		}
		g, err := forecastCompatibility(a, a)
		if err != nil || !compatibilityIdentity(g) {
			t.Fatal("identity", g, err)
		}
		g, err = forecastCompatibility(a, b)
		if err != nil {
			t.Fatal(err)
		}
		want := math.Pow(math.Sqrt(.05*.5)+math.Sqrt(.95*.5), 32)
		for j := 1; j < 5; j++ {
			adviceNear(t, g[j], want)
		}
		reverse, err := forecastCompatibility(b, a)
		if err != nil {
			t.Fatal(err)
		}
		for j := range g {
			adviceNear(t, g[j], reverse[j])
		}
		other, err := newObservationExperts(jointFixture(!uniform, true))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := forecastCompatibility(a, other); err == nil {
			t.Fatal("changed input law accepted")
		}
	}
}

func TestCompatibilityTransferLimits(t *testing.T) {
	s := newAgedAdvice(false, false)
	for i := 0; i < 120; i++ {
		s.advance(uint64(i))
		if err := observeLogAdvice(&s, uint64(i), [4]float64{.001, .999, .3, .7}, true, 0); err != nil {
			t.Fatal(err)
		}
	}
	if s.weights()[1] != 0 {
		t.Fatal("underflow not exercised")
	}
	before := s
	if err := transferCompatibility(&s, [5]float64{1, 1, 1, 1, 1}); err != nil || s != before {
		t.Fatal("identity changed state")
	}
	g := [5]float64{1, .2, .6, .8, 0}
	a, b := s, s
	for j := range b.logs {
		b.logs[j] += 1024
	}
	if err := transferCompatibility(&a, g); err != nil {
		t.Fatal(err)
	}
	if err := transferCompatibility(&b, g); err != nil {
		t.Fatal(err)
	}
	wa, wb := a.weights(), b.weights()
	for j := range wa {
		adviceNear(t, wa[j], wb[j])
	}
	if math.IsInf(a.logs[1], 0) || wa[0] != 0 {
		t.Fatal("support or log evidence lost")
	}
	if err := transferCompatibility(&s, [5]float64{}); err != nil {
		t.Fatal(err)
	}
	w := s.weights()
	for j := range w {
		adviceNear(t, w[j], s.prior[j])
	}
	for _, v := range []float64{-1, 1.1, math.NaN(), math.Inf(1)} {
		before = s
		if err := transferCompatibility(&s, [5]float64{1, 1, 1, 1, v}); err == nil || s != before {
			t.Fatal("invalid transfer mutation")
		}
	}
}

func TestCompatibilityPendingReference(t *testing.T) {
	for _, y := range []bool{false, true} {
		s := newAgedAdvice(false, false)
		row := [4]float64{.1, .8, .3, .7}
		a, b := s, s
		if err := observeLogAdvice(&a, 0, row, y, .001); err != nil {
			t.Fatal(err)
		}
		if err := observeCompatibility(&b, 0, row, y, [5]float64{1, 1, 1, 1, 1}); err != nil || a != b {
			t.Fatal("identity update differs")
		}
		g := [5]float64{1, 0, .25, .5, 1}
		var want [5]float64
		sum := 0.
		for j, p := range row {
			if !y {
				p = 1 - p
			}
			want[j+1] = s.prior[j+1] * math.Pow(p/.5, g[j+1])
			sum += want[j+1]
		}
		for j := range want {
			want[j] = .999*want[j]/sum + .001*s.prior[j]
		}
		if err := observeCompatibility(&s, 0, row, y, g); err != nil {
			t.Fatal(err)
		}
		got := s.weights()
		for j := range got {
			adviceNear(t, got[j], want[j])
		}
		before := s
		if err := observeCompatibility(&s, 0, [4]float64{0, .5, .5, .5}, y, g); err == nil || s != before {
			t.Fatal("invalid pending mutation")
		}
		before = s
		if err := observeCompatibility(&s, 1, row, y, g); err == nil || s != before {
			t.Fatal("future pending mutation")
		}
	}
}

func TestCompatibilityTransferExplicitFormula(t *testing.T) {
	s := newAgedAdvice(false, false)
	if err := observeLogAdvice(&s, 0, [4]float64{.1, .9, .4, .7}, true, .001); err != nil {
		t.Fatal(err)
	}
	w := s.weights()
	gamma := [5]float64{1, .25, .5, 0, 1}
	var want [5]float64
	total := 0.
	for j, p := range s.prior {
		if p > 0 {
			want[j] = p * math.Pow(w[j]/p, gamma[j])
			total += want[j]
		}
	}
	if err := transferCompatibility(&s, gamma); err != nil {
		t.Fatal(err)
	}
	got := s.weights()
	for j := range got {
		adviceNear(t, got[j], want[j]/total)
	}
	// Completely ignored pending evidence still incurs the declared share step.
	before := got
	if err := observeCompatibility(&s, 0, [4]float64{.01, .99, .2, .8}, true, [5]float64{}); err != nil {
		t.Fatal(err)
	}
	got = s.weights()
	for j := range got {
		adviceNear(t, got[j], .999*before[j]+.001*s.prior[j])
	}
}

func TestCompatibilityJournalIdentity(t *testing.T) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	for _, pending := range []bool{false, true} {
		for _, delay := range []int{0, 8, 31} {
			g, ref := newCompatibilityAdviceJournal(pending), newLogAdviceJournal(false)
			for clock := 0; clock < 256+delay; clock++ {
				if err := g.setClock(uint64(clock)); err != nil {
					t.Fatal(err)
				}
				if err := ref.setClock(uint64(clock)); err != nil {
					t.Fatal(err)
				}
				if delay > 0 && clock >= delay {
					i := uint64(clock - delay)
					if err := g.deliver(i, i%3 == 0); err != nil {
						t.Fatal(err)
					}
					if err := ref.deliver(i, i%3 == 0); err != nil {
						t.Fatal(err)
					}
				}
				if clock < 256 {
					if clock%32 == 0 {
						if err := g.publish(e); err != nil {
							t.Fatal(err)
						}
						if err := ref.publish(e); err != nil {
							t.Fatal(err)
						}
					}
					a, err := g.predict(uint64(clock), &jointReader{x: uint16(clock), epoch: 1}, 1)
					if err != nil {
						t.Fatal(err)
					}
					b, err := ref.predict(uint64(clock), &jointReader{x: uint16(clock), epoch: 1}, 1)
					if err != nil || a.Probability != b.Probability {
						t.Fatal("issued law differs", err)
					}
					if delay == 0 {
						if err := g.deliver(uint64(clock), clock%3 == 0); err != nil {
							t.Fatal(err)
						}
						if err := ref.deliver(uint64(clock), clock%3 == 0); err != nil {
							t.Fatal(err)
						}
					}
				}
				if g.agedAdviceJournal != *ref.agedAdviceJournal {
					t.Fatal("identity journal mismatch", pending, delay, clock)
				}
			}
		}
	}
}

func TestCompatibilityJournalVersions(t *testing.T) {
	a, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	b, err := newObservationExperts(jointFixture(true, true))
	if err != nil {
		t.Fatal(err)
	}
	bad, err := newObservationExperts(jointFixture(false, true))
	if err != nil {
		t.Fatal(err)
	}
	factor, err := forecastCompatibility(a, b)
	if err != nil {
		t.Fatal(err)
	}
	for _, pending := range []bool{false, true} {
		g := newCompatibilityAdviceJournal(pending)
		if err := g.publish(a); err != nil {
			t.Fatal(err)
		}
		for i := uint64(0); i < 64; i++ {
			if err := g.setClock(i); err != nil {
				t.Fatal(err)
			}
			if i == 32 {
				before := *g
				if err := g.publish(bad); err == nil || *g != before {
					t.Fatal("failed publication mutation")
				}
				if err := g.publish(b); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := g.predict(i, &jointReader{x: uint16(i), epoch: 1}, 1); err != nil {
				t.Fatal(err)
			}
		}
		if err := g.setClock(64); err != nil {
			t.Fatal(err)
		}
		if err := g.publish(a); err != nil {
			t.Fatal(err)
		}
		for j, f := range factor {
			adviceNear(t, math.Exp(g.logTransfer[1][j]), f*f)
		}
		entry, advice, gate := g.journal.entries[0], g.advice, g.journal.core.gate
		want := [5]float64{1, 1, 1, 1, 1}
		if pending {
			for j, f := range factor {
				want[j] = f * f
			}
		}
		if err := observeCompatibility(&advice, 0, entry.bank.Experts, true, want); err != nil {
			t.Fatal(err)
		}
		if err := g.deliver(0, true); err != nil {
			t.Fatal(err)
		}
		wa, wb := advice.weights(), g.advice.weights()
		for j := range wa {
			adviceNear(t, wa[j], wb[j])
		}
		if gate != g.journal.core.gate || g.received != 1 {
			t.Fatal("old evidence reached new gate")
		}
		before := *g
		if err := g.deliver(0, true); err == nil || *g != before {
			t.Fatal("duplicate mutation")
		}
		if err := g.expireBefore(64); err != nil {
			t.Fatal(err)
		}
		if g.advice != before.advice || g.received != 1 {
			t.Fatal("expiry duplicated evidence")
		}
		before = *g
		if err := g.deliver(1, true); err == nil || *g != before {
			t.Fatal("expired evidence was reused")
		}
		// This fixture may stop after its first view; fail that guaranteed read.
		reader := &jointReader{epoch: 1, failAt: 1}
		reader.hook = func() {
			if err := g.publish(a); err == nil {
				t.Fatal("reentrant publication")
			}
			if err := g.deliver(1, true); err == nil {
				t.Fatal("reentrant delivery")
			}
		}
		if _, err := g.predict(64, reader, 1); err == nil || *g != before {
			t.Fatalf("acquisition rollback: err=%v calls=%d stateChanged=%v", err, reader.calls, *g != before)
		}
		if _, err := g.predict(64, &jointReader{epoch: 1}, 1); err != nil {
			t.Fatal("valid issuance after rollback", err)
		}
		journalCheck(t, &g.journal)
	}
}
