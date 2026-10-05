package researchdispersion

import (
	"math"
	"reflect"
	"testing"
)

func continuingForTest(t testing.TB, h float64) *ContinuingObserver {
	t.Helper()
	m, err := NewContinuingObserver([]float64{.4, .8}, 1, 2*MaxTrials, h)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func BenchmarkContinuingNew150(b *testing.B) {
	base := make([]float64, 150)
	for i := range base {
		base[i] = .25 + .675*float64(i)/149
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		m, err := NewContinuingObserver(base, 1, 2400, 1./32)
		if err != nil || m.Pending() != 0 {
			b.Fatal(err)
		}
	}
}
func nearContinuing(t testing.TB, a, b float64) {
	t.Helper()
	if math.Abs(a-b) > 2e-12 {
		t.Fatalf("%.17g != %.17g", a, b)
	}
}

// Exhaustive path enumeration is independent of forward filtering. Three
// nominations use21^3latent paths; only the externally supplied arrived mask
// contributes likelihood. The final result predicts the next latent state.
func continuingPathOracle(prior [RateAtoms]float64, h float64, arrived [3]bool, labels [3]bool) float64 {
	var mass [RateAtoms]float64
	for a := 0; a < RateAtoms; a++ {
		for b := 0; b < RateAtoms; b++ {
			for c := 0; c < RateAtoms; c++ {
				ab, bc := h*prior[b], h*prior[c]
				if a == b {
					ab += 1 - h
				}
				if b == c {
					bc += 1 - h
				}
				w := prior[a] * ab * bc
				for j, z := range []int{a, b, c} {
					if arrived[j] {
						q := (float64(z) + .5) / 21
						if labels[j] {
							w *= q
						} else {
							w *= 1 - q
						}
					}
				}
				mass[c] += w
			}
		}
	}
	sum, q, priorMean := 0., 0., 0.
	for z := 0; z < RateAtoms; z++ {
		sum += mass[z]
		q += mass[z] * (float64(z) + .5) / 21
		priorMean += prior[z] * (float64(z) + .5) / 21
	}
	return (1-h)*q/sum + h*priorMean
}
func TestContinuingPathAndDelayedOrder(t *testing.T) {
	for _, h := range []float64{0, 1. / 32, 1. / 16, .8} {
		for _, order := range [][3]int{{0, 1, 2}, {2, 0, 1}, {1, 2, 0}} {
			m := continuingForTest(t, h)
			labels := [3]bool{true, false, true}
			var tickets [3]DelayedTicket
			var arrived [3]bool
			for j := range tickets {
				tickets[j], _ = m.Issue(0, int64(j))
			}
			q, _ := m.Predict(0)
			nearContinuing(t, q, continuingPathOracle(m.prior[0], h, arrived, labels))
			for n, j := range order {
				receipt, err := m.Resolve(tickets[j], labels[j], int64(100+n))
				if err != nil {
					t.Fatal(err)
				}
				nearContinuing(t, receipt.Forecast, tickets[j].Forecast())
				arrived[j] = true
				q, _ = m.Predict(0)
				nearContinuing(t, q, continuingPathOracle(m.prior[0], h, arrived, labels))
			}
		}
	}
}
func TestContinuingStationaryBatchAndMemberIsolation(t *testing.T) {
	m := continuingForTest(t, 0)
	before, _ := m.Predict(1)
	success, fail := 0, 0
	for j := 0; j < 16; j++ {
		ticket, err := m.Issue(0, int64(2*j))
		if err != nil {
			t.Fatal(err)
		}
		y := j%3 != 0
		if _, err = m.Resolve(ticket, y, int64(2*j+1)); err != nil {
			t.Fatal(err)
		}
		if y {
			success++
		} else {
			fail++
		}
		sum, q := 0., 0.
		for z, p := range m.prior[0] {
			v := rateAtom(z)
			w := p * math.Pow(v, float64(success)) * math.Pow(1-v, float64(fail))
			sum += w
			q += v * w
		}
		actual, _ := m.Predict(0)
		nearContinuing(t, actual, q/sum)
		other, _ := m.Predict(1)
		nearContinuing(t, other, before)
	}
}
func TestContinuingFuturePrefixIdentityAndAtomicErrors(t *testing.T) {
	a, b := continuingForTest(t, 1./16), continuingForTest(t, 1./16)
	var ta, tb []DelayedTicket
	for j := 0; j < 4; j++ {
		x, e := a.Issue(j%2, int64(j))
		if e != nil {
			t.Fatal(e)
		}
		y, e := b.Issue(j%2, int64(j))
		if e != nil {
			t.Fatal(e)
		}
		nearContinuing(t, x.Forecast(), y.Forecast())
		ta = append(ta, x)
		tb = append(tb, y)
	}
	if _, e := a.Resolve(tb[0], true, 100); e == nil {
		t.Fatal("foreign ticket accepted")
	}
	state := append([][RateAtoms]float64(nil), a.forward...)
	if _, e := a.Resolve(ta[0], true, 2); e == nil || !reflect.DeepEqual(state, a.forward) {
		t.Fatal("premature ticket or partial state")
	}
	modified := ta[0]
	modified.q = .99999
	r, e := a.Resolve(modified, true, 100)
	if e != nil {
		t.Fatal(e)
	}
	nearContinuing(t, r.Forecast, ta[0].Forecast())
	if _, e = a.Resolve(ta[0], false, 101); e == nil {
		t.Fatal("replay accepted")
	}
	if _, e = b.Resolve(tb[0], false, 100); e != nil {
		t.Fatal(e)
	}
	if reflect.DeepEqual(a.forward, b.forward) {
		t.Fatal("different available labels not incorporated")
	}
	// Prefix forecasts remain identical even though the now-arrived future label
	// differs. The private issued law never becomes the newly fitted forecast.
	for j := range ta {
		nearContinuing(t, ta[j].Forecast(), tb[j].Forecast())
		nearContinuing(t, a.ledger.trials[ta[j].slot].forecast, b.ledger.trials[tb[j].slot].forecast)
	}
	if e = a.Cancel(ta[1], 101); e != nil {
		t.Fatal(e)
	}
	if _, e = a.Resolve(ta[1], true, 102); e == nil {
		t.Fatal("cancelled label accepted")
	}
	if e = a.BeginEpoch(2, 102); e != nil {
		t.Fatal(e)
	}
	if _, e = a.Resolve(ta[2], true, 103); e == nil {
		t.Fatal("old epoch accepted")
	}
	fresh := continuingForTest(t, 1./16)
	q, _ := a.Predict(0)
	p, _ := fresh.Predict(0)
	nearContinuing(t, q, p)
}
func TestContinuingCapsAndContract(t *testing.T) {
	for _, h := range []float64{-1, 1, math.NaN(), math.Inf(1)} {
		if _, e := NewContinuingObserver([]float64{.4, .8}, 1, 1, h); e == nil {
			t.Fatal("bad hazard")
		}
	}
	m, e := NewContinuingObserver([]float64{.4, .8}, 1, 1, 1./16)
	if e != nil {
		t.Fatal(e)
	}
	x, e := m.Issue(0, 0)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Issue(1, 1); e == nil {
		t.Fatal("pending cap")
	}
	if e = m.Cancel(x, 1); e != nil {
		t.Fatal(e)
	}
	for j := 1; j < MaxTrials; j++ {
		x, e = m.Issue(0, int64(2*j))
		if e != nil {
			t.Fatal(e)
		}
		if e = m.Cancel(x, int64(2*j+1)); e != nil {
			t.Fatal(e)
		}
	}
	if _, e = m.Issue(0, 1000); e == nil {
		t.Fatal("history cap")
	}
	if m.Pending() != 0 {
		t.Fatal("lost accounting")
	}
}

func BenchmarkContinuingPredict150(b *testing.B) {
	base := make([]float64, 150)
	for i := range base {
		base[i] = .25 + .675*float64(i)/149
	}
	m, e := NewContinuingObserver(base, 1, 150*MaxTrials, 1./16)
	if e != nil {
		b.Fatal(e)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for i := range base {
			if _, e = m.Predict(i); e != nil {
				b.Fatal(e)
			}
		}
	}
}
func BenchmarkContinuingLateReplay64(b *testing.B) {
	m := continuingForTest(b, 1./16)
	var tickets [MaxTrials]DelayedTicket
	for j := range tickets {
		tickets[j], _ = m.Issue(0, int64(j))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if _, e := m.Resolve(tickets[0], n%2 == 0, 100); e != nil {
			b.Fatal(e)
		}
		// Benchmark-only restoration of the one resolved slot; original unknown
		// emissions are restored by the next complete suffix replay.
		m.ledger.trials[0].status = 1
		m.ledger.pending++
	}
}
