package researchwindowbank

import (
	"math"
	"reflect"
	"testing"
)

func near(t *testing.T, a, b float64) {
	t.Helper()
	if !finite(a) || !finite(b) || math.Abs(a-b) > 2e-10 {
		t.Fatalf("not near %.17g %.17g", a, b)
	}
}
func testModel(t *testing.T, window int) *Model {
	t.Helper()
	m, e := New([]float64{.25, .47, .78, .925}, 1, 128, Config{Depth: 2, Window: window})
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
	c.individuals = append([]individual(nil), m.individuals...)
	return c
}
func TestBaselinePreservedAcrossTreeDepths(t *testing.T) {
	for _, depth := range []int{0, 1, 2, 7} {
		base := []float64{.25, .3, .47, .5, .6, .78, .8, .9, .925}
		m, e := New(base, 1, 128, Config{Depth: depth, Window: 32})
		if e != nil {
			t.Fatal(e)
		}
		for i, b := range base {
			q, o, e := m.Predict(i)
			if e != nil {
				t.Fatal(e)
			}
			near(t, q, b)
			near(t, o, .03+.94*b)
			near(t, m.rates[i][0], b)
		}
		for _, w := range m.weights {
			if w <= 0 {
				t.Fatal("prior source support")
			}
		}
	}
}
func TestLifecycleAndAtomicFences(t *testing.T) {
	m, other := testModel(t, 6), testModel(t, 6)
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
			t.Fatal("invalid transition mutated")
		}
	}
	r, e := m.Resolve(a, true, 4)
	if e != nil || r.Forecast != a.Forecast() || r.Measurement != 1 || r.IssuedAt != 2 || r.ArrivedAt != 4 {
		t.Fatal(r, e)
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
	if _, e = m.Resolve(b, true, 8); e == nil || !reflect.DeepEqual(before, snapshot(m)) {
		t.Fatal("second replay")
	}
	c, e := m.Issue(2, 8)
	if e != nil {
		t.Fatal(e)
	}
	if e = m.Cancel(c, 9); e != nil {
		t.Fatal(e)
	}
	before = snapshot(m)
	if _, e = m.Resolve(c, false, 10); e == nil || !reflect.DeepEqual(before, snapshot(m)) {
		t.Fatal("canceled response")
	}
	if e = m.BeginEpoch(2, 10); e != nil || m.Pending() != 0 || m.seq != 0 {
		t.Fatal(e)
	}
	if _, e = m.Resolve(a, false, 10); e == nil {
		t.Fatal("old epoch")
	}
}
func TestExpiryAndZeroSupportRevival(t *testing.T) {
	m := testModel(t, 2)
	a, _ := m.Issue(0, 0)
	if _, e := m.Resolve(a, true, 0); e != nil {
		t.Fatal(e)
	}
	s, e := m.RequestAudit(a, 0)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Resolve(s, false, 0); e != nil {
		t.Fatal(e)
	}
	if m.weights[0] != 0 {
		t.Fatal("eta0 disagreement")
	}
	c, _ := m.Issue(1, 1)
	d, _ := m.Issue(2, 2)
	if m.weights[0] <= 0 {
		t.Fatal("negative infinity incorrectly retained after expiry")
	}
	if _, e = m.Query(a, "forecast"); e == nil {
		t.Fatal("expired nomination")
	}
	m.Resolve(c, false, 3)
	second, e := m.RequestAudit(c, 3)
	if e != nil {
		t.Fatal(e)
	}
	f, _ := m.Issue(3, 4)
	m.Cancel(f, 4)
	nodes, w := m.nodes, m.weights
	r, e := m.Resolve(second, true, 5)
	if e != nil || r.Forecast != second.Forecast() || m.nodes != nodes || m.weights != w {
		t.Fatal("expired second changed suffix", e)
	}
	m.Cancel(d, 5)
	for _, n := range m.nodes {
		for h := 0; h < 3; h++ {
			for z, v := range n.zeros[h] {
				if v < 0 {
					t.Fatal("negative support count")
				}
				if !finite(n.finite[h][z]) {
					t.Fatal("infinite finite ledger")
				}
			}
		}
	}
}
func TestCapsShapesAndFaultAtomicity(t *testing.T) {
	for _, cfg := range []Config{{Depth: 8, Window: 2}, {Depth: 1, Window: 0}, {Depth: -1, Window: 2}} {
		if _, e := New([]float64{.3, .7}, 1, 2, cfg); e == nil {
			t.Fatal("bad config")
		}
	}
	if _, e := New([]float64{math.NaN(), .7}, 1, 2, Config{Depth: 1, Window: 2}); e == nil {
		t.Fatal("NaN input")
	}
	m, e := New([]float64{.3, .7}, 1, 1, Config{Depth: 1, Window: 2})
	if e != nil {
		t.Fatal(e)
	}
	a, _ := m.Issue(0, 0)
	before := snapshot(m)
	if _, e = m.Issue(1, 0); e == nil || !reflect.DeepEqual(before, snapshot(m)) {
		t.Fatal("pending cap")
	}
	leaf := (1 << m.cfg.Depth) - 1 + m.ranks[0]
	m.nodes[leaf].finite[1][0] = math.Inf(1)
	before = snapshot(m)
	if _, e = m.Resolve(a, true, 1); e == nil || !reflect.DeepEqual(before, snapshot(m)) {
		t.Fatal("failed normalization mutated")
	}
	m.nodes[leaf].finite[1][0] = 0
	if _, e = m.Resolve(a, true, 1); e != nil {
		t.Fatal(e)
	}
	for _, w := range [][]float64{{1}, {.3, .3}, {-1, 2}, {math.NaN(), 0}} {
		before = snapshot(m)
		if _, e = m.PredictionValues([]Ticket{a}, w); e == nil || !reflect.DeepEqual(before, snapshot(m)) {
			t.Fatal("value shape")
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
		t.Fatal("member cap")
	}
}
func TestFeaturePermutationAndTies(t *testing.T) {
	base := []float64{.7, .3, .7, .5}
	perm := []int{2, 0, 3, 1}
	other := make([]float64, 4)
	for j, i := range perm {
		other[j] = base[i]
	}
	a, e := New(base, 1, 128, Config{Depth: 3, Window: 8})
	if e != nil {
		t.Fatal(e)
	}
	b, e := New(other, 1, 128, a.cfg)
	if e != nil {
		t.Fatal(e)
	}
	if a.ranks[0] != a.ranks[2] {
		t.Fatal("tied feature ID split")
	}
	for k := 0; k < 30; k++ {
		i, j := k%4, 0
		for perm[j] != i {
			j++
		}
		qa, oa, _ := a.Predict(i)
		qb, ob, _ := b.Predict(j)
		near(t, qa, qb)
		near(t, oa, ob)
		ta, _ := a.Issue(i, int64(k))
		tb, _ := b.Issue(j, int64(k))
		if _, e = a.Resolve(ta, k%3 == 0, int64(k)); e != nil {
			t.Fatal(e)
		}
		if _, e = b.Resolve(tb, k%3 == 0, int64(k)); e != nil {
			t.Fatal(e)
		}
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

func TestIndividualAndProductFaultAtomicity(t *testing.T) {
	for _, fault := range []string{"member", "factor", "product"} {
		m := testModel(t, 6)
		a, e := m.Issue(0, 0)
		if e != nil {
			t.Fatal(e)
		}
		leaf := (1 << m.cfg.Depth) - 1 + m.ranks[0]
		switch fault {
		case "member":
			m.individuals[0].finite[1][0] = math.Inf(1)
		case "factor":
			m.localFactors[0][1][0][1] = math.Inf(1)
		case "product":
			m.nodes[leaf].localFinite[1] = math.Inf(1)
		}
		before := snapshot(m)
		if _, e = m.Resolve(a, true, 1); e == nil || !reflect.DeepEqual(before, snapshot(m)) {
			t.Fatal("non-atomic local fault", fault, e)
		}
	}
}

func TestLateEvidenceDoesNotRestoreIndividualFactors(t *testing.T) {
	m := testModel(t, 2)
	a, e := m.Issue(0, 0)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Resolve(a, false, 0); e != nil {
		t.Fatal(e)
	}
	b, e := m.RequestAudit(a, 0)
	if e != nil {
		t.Fatal(e)
	}
	for k := 1; k <= 3; k++ {
		x, e := m.Issue(k%4, int64(k))
		if e != nil {
			t.Fatal(e)
		}
		if e = m.Cancel(x, int64(k)); e != nil {
			t.Fatal(e)
		}
	}
	old := append([]individual(nil), m.individuals...)
	nodes, w := m.nodes, m.weights
	if _, e = m.Resolve(b, true, 4); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(old, m.individuals) || nodes != m.nodes || w != m.weights {
		t.Fatal("expired paired evidence restored local branch")
	}
	for i, v := range m.individuals {
		if v.counts != ([6]int{}) {
			t.Fatal("expired counts", i, v.counts)
		}
	}
}
