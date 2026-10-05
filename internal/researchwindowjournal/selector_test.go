package researchwindowjournal

import (
	"math"
	"reflect"
	"testing"
)

// The fixture explicitly sums over one shared Y, rather than multiplying two
// marginal predictions. This constructor is not used by the selector.
func joints(qs [3]float64, etas [3]float64) Forecasts {
	var f Forecasts
	for k, q := range qs {
		f[k].Clean = q
		for a := 0; a < 2; a++ {
			for b := 0; b < 2; b++ {
				for y := 0; y < 2; y++ {
					mass := q
					if y == 0 {
						mass = 1 - q
					}
					one, two := etas[k], etas[k]
					if a == y {
						one = 1 - etas[k]
					}
					if b == y {
						two = 1 - etas[k]
					}
					f[k].Joint[2*a+b] += mass * one * two
				}
			}
		}
	}
	return f
}
func selectorCloseValue(t *testing.T, a, b float64) {
	t.Helper()
	if !selectorFinite(a) || !selectorFinite(b) || math.Abs(a-b) > 2e-12 {
		t.Fatalf("%.17g != %.17g", a, b)
	}
}
func selectorSnapshot(s *Selector) Selector {
	out := *s
	out.members = append([]selectorMember(nil), s.members...)
	return out
}
func selectorUnchanged(t *testing.T, s *Selector, b Selector) {
	t.Helper()
	if !reflect.DeepEqual(*s, b) {
		t.Fatal("failed operation changed state")
	}
}

// Reference sums every latent expert path. It does not call transition,
// factor, prepare, normalize or the production backward recurrence.
func enumerate(m selectorMember, origin int) (Weights, Weights) {
	var terminal, at Weights
	var visit func(int, int, float64, int)
	visit = func(n, previous int, mass float64, selected int) {
		if n == m.count {
			terminal[previous] += mass
			at[selected] += mass
			return
		}
		r := m.rows[n]
		for k := 0; k < 3; k++ {
			p := 1. / 3
			if n > 0 {
				p = 1 / (3 * float64(n+1))
				if k == previous {
					p += 1 - 1/float64(n+1)
				}
			}
			likelihood := 1.
			if r.first == 2 {
				a := 0
				if r.w1 {
					a = 2
				}
				likelihood = r.forecasts[k].Joint[a] + r.forecasts[k].Joint[a+1]
				if r.second == 2 {
					b := 0
					if r.w2 {
						b = 1
					}
					likelihood = r.forecasts[k].Joint[a+b]
				}
			}
			label := selected
			if n == origin {
				label = k
			}
			visit(n+1, k, mass*p*likelihood, label)
		}
	}
	visit(0, 0, 1, 0)
	sum := terminal[0] + terminal[1] + terminal[2]
	for k := 0; k < 3; k++ {
		terminal[k] /= sum
		at[k] /= sum
	}
	return terminal, at
}
func auditSmall(t *testing.T, s *Selector, tickets []SelectorTicket) int {
	t.Helper()
	m := s.members[0]
	last, _ := enumerate(m, 0)
	for k := range last {
		selectorCloseValue(t, m.rows[m.count-1].posterior[k], last[k])
	}
	tests := 1
	for _, ticket := range tickets {
		r := m.rows[ticket.ordinal]
		if r.first != 2 || r.second != 0 {
			continue
		}
		_, want := enumerate(m, ticket.ordinal)
		got, e := s.OriginWeights(ticket)
		if e != nil {
			t.Fatal(e)
		}
		q := 0.
		a := 2 * bit(r.w1)
		for k := range want {
			selectorCloseValue(t, got[k], want[k])
			q += want[k] * r.forecasts[k].Joint[a+1] / (r.forecasts[k].Joint[a] + r.forecasts[k].Joint[a+1])
		}
		actual, e := s.QuerySecond(ticket)
		if e != nil {
			t.Fatal(e)
		}
		selectorCloseValue(t, q, actual)
		tests++
	}
	return tests
}
func TestEnumeratedDelayedJoint(t *testing.T) {
	checks := 0
	for length := 1; length <= 6; length++ {
		for pattern := 0; pattern < 16; pattern++ {
			s, _ := newSelector(1, 1)
			tickets := make([]SelectorTicket, length)
			for n := range tickets {
				f := joints([3]float64{.2 + .03*float64(n), .5, .8 - .02*float64(n)}, [3]float64{.05, .1, .2})
				var e error
				tickets[n], e = s.Issue(0, int64(n), f)
				if e != nil {
					t.Fatal(e)
				}
			}
			checks += auditSmall(t, s, tickets)
			at := int64(length)
			for _, n := range []int{4, 1, 5, 0, 3, 2} {
				if n >= length {
					continue
				}
				at++
				if _, e := s.Resolve(tickets[n], pattern&(1<<uint(n%4)) != 0, at); e != nil {
					t.Fatal(e)
				}
				checks += auditSmall(t, s, tickets)
			}
			for n := range tickets {
				if n%2 != 0 {
					continue
				}
				at++
				second, e := s.RequestSecond(tickets[n], at)
				if e != nil {
					t.Fatal(e)
				}
				at++
				if _, e = s.Resolve(second, pattern&(1<<uint((n+1)%4)) != 0, at); e != nil {
					t.Fatal(e)
				}
				checks += auditSmall(t, s, tickets)
			}
		}
	}
	t.Log("exact path/reference checks", checks)
}
func TestPairReplacementNotIndependentTrials(t *testing.T) {
	s, _ := newSelector(1, 1)
	f := joints([3]float64{.15, .5, .85}, [3]float64{.1, .1, .1})
	one, _ := s.Issue(0, 0, f)
	receipt, e := s.Resolve(one, true, 1)
	if e != nil {
		t.Fatal(e)
	}
	selectorCloseValue(t, receipt.Forecast, .5)
	query, e := s.QuerySecond(one)
	if e != nil {
		t.Fatal(e)
	}
	selectorCloseValue(t, query, .82)
	second, e := s.RequestSecond(one, 2)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Resolve(second, false, 3); e != nil {
		t.Fatal(e)
	}
	// A discordant pair has probability eta*(1-eta), independent of the clean
	// rate. Independent marginal trials instead favor the central expert.
	for _, p := range s.members[0].rows[0].posterior {
		selectorCloseValue(t, p, 1./3)
	}
	wrong := Weights{}
	sum := 0.
	for k, v := range f {
		p := v.Joint[2] + v.Joint[3]
		wrong[k] = p * (1 - p)
		sum += wrong[k]
	}
	if math.Abs(wrong[1]/sum-1./3) < .05 {
		t.Fatal("vacuous paired negative control")
	}
}
func TestOriginReplayNotNaiveArrival(t *testing.T) {
	s, _ := newSelector(1, 1)
	f := joints([3]float64{.05, .5, .95}, [3]float64{.1, .1, .1})
	ts := make([]SelectorTicket, 5)
	for n := range ts {
		ts[n], _ = s.Issue(0, int64(n), f)
	}
	_, _ = s.Resolve(ts[4], false, 5)
	before := s.members[0].rows[4].posterior
	_, _ = s.Resolve(ts[0], true, 6)
	actual := s.members[0].rows[4].posterior
	naive := before
	sum := 0.
	for k := range naive {
		naive[k] *= f[k].Joint[2] + f[k].Joint[3]
		sum += naive[k]
	}
	discrepancy := 0.
	for k := range naive {
		discrepancy += math.Abs(naive[k]/sum - actual[k])
	}
	if discrepancy < .01 {
		t.Fatal("vacuous origin/arrival negative control", discrepancy)
	}
	auditSmall(t, s, ts)
}
func TestLifecycleMissingCapsAndEpochs(t *testing.T) {
	s, _ := newSelector(2, 1)
	other, _ := newSelector(2, 1)
	f := joints([3]float64{.3, .5, .7}, [3]float64{.1, .1, .1})
	one, _ := s.Issue(0, 0, f)
	foreign, _ := other.Issue(0, 0, f)
	b := selectorSnapshot(s)
	if _, e := s.RequestSecond(one, 1); e == nil {
		t.Fatal("unobserved first allowed")
	}
	selectorUnchanged(t, s, b)
	if _, e := s.Resolve(foreign, true, 1); e == nil {
		t.Fatal("foreign ticket")
	}
	selectorUnchanged(t, s, b)
	if e := s.Cancel(one, -1); e == nil {
		t.Fatal("old time")
	}
	selectorUnchanged(t, s, b)
	_, e := s.Resolve(one, true, 1)
	if e != nil {
		t.Fatal(e)
	}
	b = selectorSnapshot(s)
	if _, e = s.Resolve(one, true, 2); e == nil {
		t.Fatal("duplicate first")
	}
	selectorUnchanged(t, s, b)
	second, e := s.RequestSecond(one, 2)
	if e != nil {
		t.Fatal(e)
	}
	b = selectorSnapshot(s)
	if _, e = s.RequestSecond(one, 3); e == nil {
		t.Fatal("duplicate request")
	}
	selectorUnchanged(t, s, b)
	w, _ := s.Weights(0)
	if e = s.Cancel(second, 3); e != nil {
		t.Fatal(e)
	}
	after, _ := s.Weights(0)
	if w != after {
		t.Fatal("missing second changed posterior")
	}
	b = selectorSnapshot(s)
	if _, e = s.Resolve(second, true, 4); e == nil {
		t.Fatal("cancelled second replay")
	}
	selectorUnchanged(t, s, b)
	missing, _ := s.Issue(1, 4, f)
	before, _ := s.Weights(1)
	if e = s.Cancel(missing, 5); e != nil {
		t.Fatal(e)
	}
	after, _ = s.Weights(1)
	if before != after {
		t.Fatal("missing first learned")
	}
	for n := 1; n < SelectorMaxTrials; n++ {
		if _, e = s.Issue(0, int64(6+n), f); e != nil {
			t.Fatal(e)
		}
	}
	b = selectorSnapshot(s)
	if _, e = s.Issue(0, 100, f); e == nil {
		t.Fatal("trial cap")
	}
	selectorUnchanged(t, s, b)
	if e = s.BeginEpoch(1, 100); e == nil {
		t.Fatal("reused epoch")
	}
	selectorUnchanged(t, s, b)
	if e = s.BeginEpoch(2, 100); e != nil {
		t.Fatal(e)
	}
	b = selectorSnapshot(s)
	if _, e = s.Resolve(one, false, 101); e == nil {
		t.Fatal("stale epoch")
	}
	selectorUnchanged(t, s, b)
	if s.members[0].count != 0 {
		t.Fatal("epoch retained evidence")
	}
	for _, args := range [][2]int{{0, 1}, {201, 1}, {1, 0}} {
		if _, e := newSelector(args[0], uint64(args[1])); e == nil {
			t.Fatal("invalid constructor")
		}
	}
}
func TestInvalidInputsAndZeroSupportAtomicity(t *testing.T) {
	s, _ := newSelector(1, 1)
	f := joints([3]float64{.2, .5, .8}, [3]float64{.1, .1, .1})
	for fault := 0; fault < 6; fault++ {
		bad := f
		switch fault {
		case 0:
			bad[0].Clean = math.NaN()
		case 1:
			bad[1].Clean = 2
		case 2:
			bad[0].Joint[0] = math.Inf(1)
		case 3:
			bad[1].Joint[0] = -.1
		case 4:
			bad[2].Joint[0] += .01
		case 5:
			bad[2].Joint = [4]float64{}
		}
		b := selectorSnapshot(s)
		if _, e := s.Issue(0, 0, bad); e == nil {
			t.Fatal("invalid forecast", fault)
		}
		selectorUnchanged(t, s, b)
	}
	var impossible Forecasts
	for k := range impossible {
		impossible[k] = Expert{.5, [4]float64{1, 0, 0, 0}}
	}
	one, _ := s.Issue(0, 0, impossible)
	b := selectorSnapshot(s)
	if _, e := s.Resolve(one, true, 1); e == nil {
		t.Fatal("zero-supported evidence")
	}
	selectorUnchanged(t, s, b)
	if _, e := s.Resolve(one, false, 1); e != nil {
		t.Fatal(e)
	}
	second, e := s.RequestSecond(one, 2)
	if e != nil {
		t.Fatal(e)
	}
	selectorCloseValue(t, second.Forecast(), 0)
	b = selectorSnapshot(s)
	if _, e = s.Resolve(second, true, 3); e == nil {
		t.Fatal("zero-supported pair")
	}
	selectorUnchanged(t, s, b)
	if _, e = s.Resolve(second, false, 3); e != nil {
		t.Fatal(e)
	}
}
func TestFutureForksPermutationAndImmutableIssue(t *testing.T) {
	f := joints([3]float64{.2, .5, .8}, [3]float64{.05, .1, .2})
	a, _ := newSelector(2, 1)
	b, _ := newSelector(2, 1)
	ta, _ := a.Issue(0, 0, f)
	tb, _ := b.Issue(0, 0, f)
	// No outcome can enter any forecast before receipt. Alternate hidden future
	// outcomes remain outside both states; the split occurs only on actual resolve.
	for k := 0; k < 6; k++ {
		wa, _ := a.Weights(0)
		wb, _ := b.Weights(0)
		if wa != wb {
			t.Fatal("future leak")
		}
	}
	_, _ = a.Resolve(ta, true, 1)
	_, _ = b.Resolve(tb, false, 1)
	if a.members[0].rows[0].posterior == b.members[0].rows[0].posterior {
		t.Fatal("vacuous future fork")
	}
	selectorCloseValue(t, ta.Forecast(), .5)
	selectorCloseValue(t, a.members[0].rows[0].issuedClean, .5)
	u, _ := a.Weights(1)
	if u != uniform() {
		t.Fatal("other selectorMember polluted")
	}
	f[0].Clean = 0
	selectorCloseValue(t, a.members[0].rows[0].forecasts[0].Clean, .2)
	c, _ := newSelector(1, 1)
	d, _ := newSelector(1, 1)
	original := joints([3]float64{.2, .5, .8}, [3]float64{.05, .1, .2})
	permuted := Forecasts{original[2], original[0], original[1]}
	tc, _ := c.Issue(0, 0, original)
	td, _ := d.Issue(0, 0, permuted)
	_, _ = c.Resolve(tc, true, 1)
	_, _ = d.Resolve(td, true, 1)
	qc, _ := c.QuerySecond(tc)
	qd, _ := d.QuerySecond(td)
	selectorCloseValue(t, qc, qd)
	wc, _ := c.Weights(0)
	wd, _ := d.Weights(0)
	for k, p := range []int{2, 0, 1} {
		selectorCloseValue(t, wd[k], wc[p])
	}
}
func TestArrivalOrderAndSmoothedRequest(t *testing.T) {
	f := joints([3]float64{.2, .5, .8}, [3]float64{.1, .1, .2})
	s, _ := newSelector(1, 1)
	r, _ := newSelector(1, 1)
	ts := make([]SelectorTicket, 4)
	tr := make([]SelectorTicket, 4)
	for n := range ts {
		ts[n], _ = s.Issue(0, int64(n), f)
		tr[n], _ = r.Issue(0, int64(n), f)
	}
	for j, n := range []int{3, 0, 2, 1} {
		_, _ = s.Resolve(ts[n], n%2 == 0, int64(4+j))
	}
	for n := range tr {
		_, _ = r.Resolve(tr[n], n%2 == 0, int64(4+n))
	}
	ws, _ := s.Weights(0)
	wr, _ := r.Weights(0)
	for k := range ws {
		selectorCloseValue(t, ws[k], wr[k])
	}
	qs, _ := s.QuerySecond(ts[0])
	qr, _ := r.QuerySecond(tr[0])
	selectorCloseValue(t, qs, qr)
	a, _ := s.RequestSecond(ts[0], 8)
	b, _ := r.RequestSecond(tr[0], 8)
	_, _ = s.Resolve(a, false, 9)
	_, _ = r.Resolve(b, false, 9)
	ws, _ = s.Weights(0)
	wr, _ = r.Weights(0)
	for k := range ws {
		selectorCloseValue(t, ws[k], wr[k])
	}
	auditSmall(t, s, ts)
	auditSmall(t, r, tr)
}
