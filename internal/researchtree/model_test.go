package researchtree

import (
	"math"
	"reflect"
	"testing"
)

func testModel(t *testing.T, window int) *Model {
	t.Helper()
	m, e := New([]float64{.25, .47, .78, .925}, 1, 128, Config{Depth: 2, Window: window, Strength: 2})
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func snapshot(m *Model) Model {
	c := *m
	c.trials = append([]trial(nil), m.trials...)
	c.issued = append([]int(nil), m.issued...)
	c.seqSlots = append([]int(nil), m.seqSlots...)
	return c
}
func near(t *testing.T, a, b float64) {
	t.Helper()
	if !finite(a) || !finite(b) || math.Abs(a-b) > 2e-10 {
		t.Fatalf("not near %.17g %.17g", a, b)
	}
}
func TestLifecycleAndAtomicFences(t *testing.T) {
	m := testModel(t, 6)
	other := testModel(t, 6)
	a, e := m.Issue(0, 2)
	if e != nil {
		t.Fatal(e)
	}
	foreign, _ := other.Issue(0, 2)
	bad := []func() error{func() error { _, e := m.Issue(-1, 2); return e }, func() error { _, e := m.Issue(0, 1); return e }, func() error { _, e := m.Resolve(foreign, true, 3); return e }, func() error { _, e := m.Resolve(Ticket{}, true, 3); return e }, func() error { _, e := m.RequestAudit(a, 3); return e }, func() error { _, e := m.Query(a, "forecast"); return e }, func() error { return m.BeginEpoch(1, 3) }}
	for _, f := range bad {
		before := snapshot(m)
		if f() == nil {
			t.Fatal("invalid operation accepted")
		}
		if !reflect.DeepEqual(before, snapshot(m)) {
			t.Fatal("failed transition mutated")
		}
	}
	r, e := m.Resolve(a, true, 4)
	if e != nil || r.Forecast != a.Forecast() || r.IssuedAt != 2 || r.ArrivedAt != 4 || r.Measurement != 1 {
		t.Fatal("original receipt", r, e)
	}
	before := snapshot(m)
	for _, mode := range []string{"forecast", "uncertainty", "information", "falsification"} {
		if _, e = m.Query(a, mode); e != nil {
			t.Fatal(e)
		}
	}
	if _, e = m.PredictionValues([]Ticket{a}, []float64{.25, .25, .25, .25}); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(before, snapshot(m)) {
		t.Fatal("hypothetical evidence committed")
	}
	b, e := m.RequestAudit(a, 5)
	if e != nil {
		t.Fatal(e)
	}
	r, e = m.Resolve(b, false, 7)
	if e != nil || r.Forecast != b.Forecast() || r.Measurement != 2 || r.IssuedAt != 5 {
		t.Fatal(r, e)
	}
	before = snapshot(m)
	if _, e = m.Resolve(b, true, 8); e == nil {
		t.Fatal("replayed second")
	}
	if !reflect.DeepEqual(before, snapshot(m)) {
		t.Fatal("replay mutation")
	}
	c, _ := m.Issue(2, 8)
	if e = m.Cancel(c, 9); e != nil {
		t.Fatal(e)
	}
	before = snapshot(m)
	if _, e = m.Resolve(c, false, 10); e == nil {
		t.Fatal("canceled response")
	}
	if !reflect.DeepEqual(before, snapshot(m)) {
		t.Fatal("cancel replay mutation")
	}
	if e = m.BeginEpoch(2, 10); e != nil || m.Pending() != 0 || m.seq != 0 {
		t.Fatal("epoch", e)
	}
	if _, e = m.Resolve(a, false, 10); e == nil {
		t.Fatal("old epoch")
	}
}
func TestExpiryPairingAndSupportRevival(t *testing.T) {
	m := testModel(t, 2)
	a, _ := m.Issue(0, 0)
	m.Resolve(a, true, 0)
	b, e := m.RequestAudit(a, 0)
	if e != nil {
		t.Fatal(e)
	}
	m.Resolve(b, false, 0)
	if m.weights[0] != 0 {
		t.Fatal("eta zero supports disagreement")
	}
	old := m.trials[a.slot]
	c, _ := m.Issue(1, 1)
	d, _ := m.Issue(2, 2)
	if m.weights[0] <= 0 {
		t.Fatal("expired contradiction failed to revive noise model")
	}
	if _, e = m.Query(a, "forecast"); e == nil {
		t.Fatal("expired origin nominated")
	}
	nodes, weights := m.nodes, m.weights
	r, e := m.Resolve(c, false, 3)
	if e != nil {
		t.Fatal(e)
	}
	_ = r
	second, _ := m.RequestAudit(c, 3)
	f, _ := m.Issue(3, 4)
	m.Cancel(f, 4)
	nodes, weights = m.nodes, m.weights
	r, e = m.Resolve(second, true, 5)
	if e != nil || r.Forecast != second.Forecast() || m.nodes != nodes || m.weights != weights {
		t.Fatal("late expired second changed suffix", r, e)
	}
	if m.trials[a.slot] != old {
		t.Fatal("expiry altered retained receipt history")
	}
	m.Cancel(d, 5)
	for n := 0; n < (1<<(m.cfg.Depth+1))-1; n++ {
		sum := 0
		for _, v := range m.nodes[n].counts {
			sum += v
		}
		if sum > m.cfg.Window {
			t.Fatal("window cap")
		}
	}
}
func TestCapsFaultsAndValueShape(t *testing.T) {
	for _, cfg := range []Config{{Depth: 8, Window: 2, Strength: 2}, {Depth: 2, Window: 0, Strength: 2}, {Depth: 2, Window: 2, Strength: math.NaN()}} {
		if _, e := New([]float64{.3, .7}, 1, 2, cfg); e == nil {
			t.Fatal("bad config")
		}
	}
	m, e := New([]float64{.3, .7}, 1, 1, Config{Depth: 1, Window: 2, Strength: 2})
	if e != nil {
		t.Fatal(e)
	}
	a, _ := m.Issue(0, 0)
	before := snapshot(m)
	if _, e = m.Issue(1, 0); e == nil || !reflect.DeepEqual(before, snapshot(m)) {
		t.Fatal("pending cap")
	}
	leaf := (1 << m.cfg.Depth) - 1 + m.ranks[0]
	m.logPrior[leaf][0] = math.Inf(1)
	before = snapshot(m)
	if _, e = m.Resolve(a, true, 1); e == nil || !reflect.DeepEqual(before, snapshot(m)) {
		t.Fatal("arithmetic fault not atomic")
	}
	m.logPrior[leaf][0] = math.Log(.01)
	m.Resolve(a, true, 1)
	for _, w := range [][]float64{{1}, {.3, .3}, {-1, 2}, {math.NaN(), 0}} {
		before = snapshot(m)
		if _, e = m.PredictionValues([]Ticket{a}, w); e == nil || !reflect.DeepEqual(before, snapshot(m)) {
			t.Fatal("value weights")
		}
	}
	if _, e = m.PredictionValues([]Ticket{a, a}, []float64{.5, .5}); e == nil {
		t.Fatal("duplicate origins")
	}
	m = testModel(t, 1)
	for k := 0; k < 64; k++ {
		x, e := m.Issue(0, int64(k))
		if e != nil {
			t.Fatal(e)
		}
		if e = m.Cancel(x, int64(k)); e != nil {
			t.Fatal(e)
		}
	}
	before = snapshot(m)
	if _, e = m.Issue(0, 64); e == nil || !reflect.DeepEqual(before, snapshot(m)) {
		t.Fatal("per member cap")
	}
}
func TestPublicFeaturePermutationAndTies(t *testing.T) {
	base := []float64{.7, .3, .7, .5}
	perm := []int{2, 0, 3, 1}
	other := make([]float64, 4)
	for j, i := range perm {
		other[j] = base[i]
	}
	a, e := New(base, 1, 128, Config{Depth: 3, Window: 8, Strength: 2})
	if e != nil {
		t.Fatal(e)
	}
	b, e := New(other, 1, 128, a.cfg)
	if e != nil {
		t.Fatal(e)
	}
	if a.ranks[0] != a.ranks[2] {
		t.Fatal("equal features split by ID")
	}
	for k := 0; k < 30; k++ {
		i := k % 4
		j := 0
		for perm[j] != i {
			j++
		}
		qa, oa, _ := a.Predict(i)
		qb, ob, _ := b.Predict(j)
		near(t, qa, qb)
		near(t, oa, ob)
		ta, _ := a.Issue(i, int64(k))
		tb, _ := b.Issue(j, int64(k))
		a.Resolve(ta, k%3 == 0, int64(k))
		b.Resolve(tb, k%3 == 0, int64(k))
		if k%4 == 0 {
			sa, _ := a.RequestAudit(ta, int64(k))
			sb, _ := b.RequestAudit(tb, int64(k))
			a.Resolve(sa, k%2 == 0, int64(k))
			b.Resolve(sb, k%2 == 0, int64(k))
		}
		for h, w := range a.weights {
			near(t, w, b.weights[h])
		}
	}
}
