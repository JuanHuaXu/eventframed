package observationlearners

import (
	"reflect"
	"testing"
)

func journalCheck(t testing.TB, g *routedFeedbackJournal) {
	t.Helper()
	s := g.stats
	if s.Issued != s.Pending+s.Applied+s.BankOnly+s.Stale+s.Censored || s.Issued != g.core.bank.issued || s.Issued != g.core.gate.gate.issued {
		t.Fatal("journal accounting", s)
	}
	active := uint64(0)
	for _, e := range g.entries {
		if e.active {
			active++
		}
	}
	if active != s.Pending || g.core.bank.active || g.core.gate.gate.active {
		t.Fatal("pending accounting")
	}
}

func TestRoutedJournalImmediateParity(t *testing.T) {
	e, err := newObservationExperts(jointFixture(false, false))
	if err != nil {
		t.Fatal(err)
	}
	g := newRoutedFeedbackJournal()
	control := newRoutedObservationState()
	for step := uint64(0); step < 256; step++ {
		if step%32 == 0 {
			if err := g.publish(e); err != nil {
				t.Fatal(err)
			}
		}
		x := uint16(step*11) & 511
		got, err := g.predict(step, &jointReader{x: x, epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		want, err := control.predict(step, e, &jointReader{x: x, epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatal("immediate forecast mismatch", step)
		}
		if err := g.deliver(step, step%3 == 0); err != nil {
			t.Fatal(err)
		}
		if err := control.observe(step, step%3 == 0); err != nil {
			t.Fatal(err)
		}
		if g.core != *control {
			t.Fatal("immediate state mismatch", step)
		}
		journalCheck(t, g)
	}
	if g.stats.Applied != 256 || g.stats.Pending != 0 {
		t.Fatal(g.stats)
	}
}

func TestRoutedJournalOrdering(t *testing.T) {
	e, _ := newObservationExperts(jointFixture(false, false))
	g := newRoutedFeedbackJournal()
	if err := g.publish(e); err != nil {
		t.Fatal(err)
	}
	for step := uint64(0); step < 3; step++ {
		if _, err := g.predict(step, &jointReader{x: uint16(step * 41), epoch: 1}, 1); err != nil {
			t.Fatal(err)
		}
	}
	entries := g.entries
	ref := g.core
	// Deliberately feed the reference the original captured raw forecasts, not
	// current predictions. The journal must defer the early-arriving third label.
	for _, origin := range []uint64{2, 0, 1} {
		before := g.core
		if err := g.deliver(origin, origin != 1); err != nil {
			t.Fatal(err)
		}
		if origin == 2 && g.core != before {
			t.Fatal("out-of-order outcome trained early")
		}
	}
	for origin := uint64(0); origin < 3; origin++ {
		entry := entries[origin]
		ref.bank.pending = entry.bank
		ref.bank.active = true
		ref.gate.gate.pending = entry.gate
		ref.gate.gate.active = true
		if err := ref.observe(origin, origin != 1); err != nil {
			t.Fatal(err)
		}
	}
	if g.core != ref {
		t.Fatal("original forecast update mismatch")
	}
	journalCheck(t, g)
	before := *g
	if err := g.deliver(2, false); err == nil || *g != before {
		t.Fatal("duplicate changed state")
	}
}

func TestRoutedJournalVersionExpiry(t *testing.T) {
	e, _ := newObservationExperts(jointFixture(true, true))
	g := newRoutedFeedbackJournal()
	if err := g.publish(e); err != nil {
		t.Fatal(err)
	}
	for step := uint64(0); step < 32; step++ {
		if _, err := g.predict(step, &jointReader{epoch: 1}, 1); err != nil {
			t.Fatal(err)
		}
	}
	before := *g
	if _, err := g.predict(32, &jointReader{epoch: 1}, 1); err == nil || *g != before {
		t.Fatal("missing publication accepted")
	}
	if err := g.publish(e); err != nil {
		t.Fatal(err)
	}
	core := g.core
	for origin := uint64(0); origin < 32; origin++ {
		if err := g.deliver(origin, true); err != nil {
			t.Fatal(err)
		}
	}
	if g.core != core || g.stats.Stale != 32 {
		t.Fatal("old forecasts changed new version")
	}
	for origin := uint64(32); origin < 35; origin++ {
		if _, err := g.predict(origin, &jointReader{epoch: 1}, 1); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.deliver(34, true); err != nil {
		t.Fatal(err)
	}
	if err := g.expireBefore(35); err != nil {
		t.Fatal(err)
	}
	if g.stats.Applied != 1 || g.stats.Censored != 2 || g.stats.Pending != 0 {
		t.Fatal("expiry lost delivered evidence", g.stats)
	}
	journalCheck(t, g)
}

func TestRoutedJournalBackpressureAtomicity(t *testing.T) {
	e, _ := newObservationExperts(jointFixture(true, true))
	g := newRoutedFeedbackJournal()
	for origin := uint64(0); origin < 64; origin++ {
		if origin%32 == 0 {
			if err := g.publish(e); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := g.predict(origin, &jointReader{epoch: 1}, 1); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.publish(e); err != nil {
		t.Fatal(err)
	}
	before := *g
	r := &jointReader{epoch: 1}
	if _, err := g.predict(64, r, 1); err == nil || r.calls != 0 || *g != before {
		t.Fatal("backpressure read/mutation")
	}
	if err := g.expireBefore(32); err != nil {
		t.Fatal(err)
	}
	before = *g
	r = &jointReader{epoch: 1, failAt: 2}
	r.hook = func() {
		if err := g.publish(e); err == nil {
			t.Fatal("reentrant publication")
		}
		if err := g.deliver(32, true); err == nil {
			t.Fatal("reentrant feedback")
		}
		if err := g.expireBefore(64); err == nil {
			t.Fatal("reentrant expiry")
		}
	}
	if _, err := g.predict(64, r, 1); err == nil || *g != before {
		t.Fatal("failed read mutation")
	}
	if _, err := g.predict(64, &jointReader{epoch: 1}, 1); err != nil {
		t.Fatal(err)
	}
	before = *g
	if err := g.publish(e); err == nil || *g != before {
		t.Fatal("off-cadence publication")
	}
	if err := g.expireBefore(66); err == nil || *g != before {
		t.Fatal("future expiry")
	}
	if err := g.deliver(0, true); err == nil || *g != before {
		t.Fatal("overwritten origin accepted")
	}
	journalCheck(t, g)
}

func TestRoutedJournalBufferedRelease(t *testing.T) {
	e, _ := newObservationExperts(jointFixture(true, false))
	g := newRoutedFeedbackJournal()
	g.publish(e)
	for origin := uint64(0); origin < 8; origin++ {
		if _, err := g.predict(origin, &jointReader{x: uint16(origin), epoch: 1}, 1); err != nil {
			t.Fatal(err)
		}
	}
	for origin := uint64(7); origin > 0; origin-- {
		if err := g.deliver(origin, true); err != nil {
			t.Fatal(err)
		}
	}
	if g.stats.Applied != 0 {
		t.Fatal("missing head bypassed")
	}
	if err := g.expireBefore(1); err != nil {
		t.Fatal(err)
	}
	if g.stats.Applied != 7 || g.stats.Censored != 1 {
		t.Fatal("buffered release", g.stats)
	}
	journalCheck(t, g)
}

func TestRoutedJournalRoleCarry(t *testing.T) {
	e, err := newObservationExperts(jointFixture(false, false))
	if err != nil {
		t.Fatal(err)
	}
	g := newRoleRoutedFeedbackJournal()
	if err := g.publish(e); err != nil {
		t.Fatal(err)
	}
	for origin := uint64(0); origin < 32; origin++ {
		if _, err := g.predict(origin, &jointReader{x: uint16(origin*7) & 511, epoch: 1}, 1); err != nil {
			t.Fatal(err)
		}
	}
	entries := g.entries
	if err := g.publish(e); err != nil {
		t.Fatal(err)
	}
	frozenGate := g.core.gate
	ref := g.core.bank
	for origin := uint64(32); origin > 0; origin-- {
		if err := g.deliver(origin-1, origin%3 == 0); err != nil {
			t.Fatal(err)
		}
		if g.core.gate != frozenGate {
			t.Fatal("stale evidence changed current tests/credits")
		}
	}
	for origin := uint64(0); origin < 32; origin++ {
		ref.pending = entries[origin].bank
		ref.active = true
		if err := ref.observe(origin, (origin+1)%3 == 0); err != nil {
			t.Fatal(err)
		}
	}
	if ref != g.core.bank || g.stats.BankOnly != 32 || g.stats.Stale != 0 {
		t.Fatal("role loss carry", g.stats)
	}
	journalCheck(t, g)
	// The next issued forecast uses the updated selector, but no stale test is
	// allowed to reject an expert in the new publication.
	got, err := g.predict(32, &jointReader{x: 19, epoch: 1}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Probability != g.entries[32].served {
		t.Fatal("issued law binding")
	}
	for _, test := range g.core.gate.gate.tests {
		if test.Count != 0 || test.Rejected {
			t.Fatal("stale certificate")
		}
	}
}

func TestRoutedJournalDelayGrid(t *testing.T) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	for delay := 0; delay < 32; delay++ {
		for _, missing := range []bool{false, true} {
			control, rescue := newRoutedFeedbackJournal(), newRoleRoutedFeedbackJournal()
			for clock := 0; clock < 288; clock++ {
				for _, g := range []*routedFeedbackJournal{control, rescue} {
					if clock < 256 {
						if clock%32 == 0 {
							if err := g.publish(e); err != nil {
								t.Fatal(err)
							}
						}
						if _, err := g.predict(uint64(clock), &jointReader{x: uint16(clock), epoch: 1}, 1); err != nil {
							t.Fatal(err)
						}
					}
					origin := clock - delay
					if origin >= 0 && origin < 256 && !(missing && origin%5 == 0) {
						if err := g.deliver(uint64(origin), origin%3 == 0); err != nil {
							t.Fatal(err)
						}
					}
					if clock >= 32 {
						if err := g.expireBefore(uint64(clock - 31)); err != nil {
							t.Fatal(err)
						}
					}
					journalCheck(t, g)
				}
			}
			wantDelivered, wantCensored := uint64(256), uint64(0)
			if missing {
				wantDelivered, wantCensored = 204, 52
			}
			if rescue.stats.Applied+rescue.stats.BankOnly != wantDelivered || rescue.stats.Censored != wantCensored || rescue.stats.Pending != 0 || control.stats.Applied+control.stats.Stale != wantDelivered || control.stats.Censored != wantCensored || control.stats.Pending != 0 {
				t.Fatal("delay accounting", delay, missing, control.stats, rescue.stats)
			}
			if !missing && (control.stats.Stale != uint64(7*delay) || rescue.stats.BankOnly != uint64(7*delay)) {
				t.Fatal("publication exposure", delay, control.stats, rescue.stats)
			}
			if delay == 0 || delay == 8 || delay == 16 || delay == 31 {
				t.Logf("delay=%d missingEvery5=%t control=%+v roleCarry=%+v", delay, missing, control.stats, rescue.stats)
			}
		}
	}
}

func BenchmarkRoutedJournalLifetime(b *testing.B) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		b.Fatal(err)
	}
	for _, delay := range []int{0, 8} {
		name := "immediate"
		if delay > 0 {
			name = "delay8"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				g := newRoutedFeedbackJournal()
				for origin := 0; origin < 256; origin++ {
					if origin%32 == 0 {
						if err := g.publish(e); err != nil {
							b.Fatal(err)
						}
					}
					if _, err := g.predict(uint64(origin), &jointReader{x: uint16(origin), epoch: 1}, 1); err != nil {
						b.Fatal(err)
					}
					if origin >= delay {
						if err := g.deliver(uint64(origin-delay), (origin-delay)%3 == 0); err != nil {
							b.Fatal(err)
						}
					}
				}
				for origin := 256 - delay; origin < 256; origin++ {
					if err := g.deliver(uint64(origin), origin%3 == 0); err != nil {
						b.Fatal(err)
					}
				}
				journalCheck(b, g)
				if g.stats.Pending != 0 {
					b.Fatal("unsettled lifetime")
				}
			}
		})
	}
}

func BenchmarkRoutedImmediateLifetime(b *testing.B) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s := newRoutedObservationState()
		for origin := uint64(0); origin < 256; origin++ {
			if _, err := s.predict(origin, e, &jointReader{x: uint16(origin), epoch: 1}, 1); err != nil {
				b.Fatal(err)
			}
			if err := s.observe(origin, origin%3 == 0); err != nil {
				b.Fatal(err)
			}
		}
	}
}
