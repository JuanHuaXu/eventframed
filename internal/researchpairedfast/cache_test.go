package researchpairedfast

import (
	orig "github.com/JuanHuaXu/eventframed/internal/researchpaired"
	"math"
	"reflect"
	"testing"
)

func assertOption(t *testing.T, a Option, b orig.Option) {
	t.Helper()
	for j, x := range []float64{a.Observed, a.Uncertainty, a.Information, a.EdgeCut} {
		y := []float64{b.Observed, b.Uncertainty, b.Information, b.EdgeCut}[j]
		if !finite(x) || math.Abs(x-y) > 2e-10 {
			t.Fatalf("option %d %.17g %.17g", j, x, y)
		}
	}
}

func TestFastCacheLocalAndGlobalFences(t *testing.T) {
	m, _ := New([]float64{.3, .9}, 1, 128, testConfig())
	o, _ := orig.New([]float64{.3, .9}, 1, 128, testConfig())
	a, _ := m.Issue(0, 1)
	oa, _ := o.Issue(0, 1)
	b, _ := m.Issue(1, 2)
	ob, _ := o.Issue(1, 2)
	m.Resolve(a, true, 3)
	o.Resolve(oa, true, 3)
	m.Resolve(b, true, 4)
	o.Resolve(ob, true, 4)
	before, e := m.Options(a)
	if e != nil {
		t.Fatal(e)
	}
	old, e := o.Options(oa)
	if e != nil {
		t.Fatal(e)
	}
	assertOption(t, before, old)
	cache := m.queries[0]
	rev := m.revision[0]
	for _, mode := range []string{"forecast", "uncertainty", "information", "falsification"} {
		if _, e = m.Query(a, mode); e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(cache, m.queries[0]) {
			t.Fatal("cache did not reuse component conditionals")
		}
	}
	s, _ := m.RequestAudit(b, 5)
	os, _ := o.RequestAudit(ob, 5)
	if _, e = m.Resolve(s, false, 6); e != nil {
		t.Fatal(e)
	}
	if _, e = o.Resolve(os, false, 6); e != nil {
		t.Fatal(e)
	}
	after, e := m.Options(a)
	if e != nil {
		t.Fatal(e)
	}
	old, e = o.Options(oa)
	if e != nil {
		t.Fatal(e)
	}
	assertOption(t, after, old)
	if m.revision[0] != rev || !reflect.DeepEqual(cache, m.queries[0]) {
		t.Fatal("global evidence invalidated member-local conditional")
	}
	if math.Abs(after.Observed-before.Observed) < 1e-5 {
		t.Fatal("global recombination test vacuous")
	}
	c, _ := m.Issue(0, 7)
	oc, _ := o.Issue(0, 7)
	if m.revision[0] == rev {
		t.Fatal("issue did not fence cache")
	}
	if e = m.Cancel(c, 8); e != nil {
		t.Fatal(e)
	}
	if e = o.Cancel(oc, 8); e != nil {
		t.Fatal(e)
	}
	after, e = m.Options(a)
	if e != nil {
		t.Fatal(e)
	}
	old, e = o.Options(oa)
	if e != nil {
		t.Fatal(e)
	}
	assertOption(t, after, old)
	if m.queries[0].revision != m.revision[0] {
		t.Fatal("cancel fence")
	}
	s0, e := m.RequestAudit(a, 9)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Options(a); e == nil {
		t.Fatal("pending audit replay")
	}
	if e = m.Cancel(s0, 10); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Options(a); e == nil {
		t.Fatal("cancelled audit replay")
	}
	if e = m.BeginEpoch(2, 11); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Options(a); e == nil {
		t.Fatal("old epoch cache")
	}
	for _, cache := range m.queries {
		for _, valid := range cache.valid {
			if valid {
				t.Fatal("epoch retained cache")
			}
		}
	}
}

func TestFastCacheRevivalAndNumericFault(t *testing.T) {
	m, _ := New([]float64{.3, .9}, 1, 128, testConfig())
	a, _ := m.Issue(0, 1)
	m.Resolve(a, true, 2)
	w := m.weights
	// Simulate a numerically underflowed global component. A later global
	// normalization may revive it without changing this member's local history.
	m.weights[0] = 0
	for c := 1; c < Components; c++ {
		m.weights[c] /= 1 - w[0]
	}
	if _, e := m.Options(a); e != nil {
		t.Fatal(e)
	}
	if m.queries[0].valid[0] {
		t.Fatal("zero-weight conditional computed")
	}
	m.weights = w
	if _, e := m.Options(a); e != nil {
		t.Fatal(e)
	}
	if !m.queries[0].valid[0] {
		t.Fatal("revived conditional missing")
	}
	m.queries[0].probabilities[0] = math.NaN()
	if _, e := m.Options(a); e == nil {
		t.Fatal("NaN cached query served")
	}
	if m.weights != w || m.pending != 0 || m.clock != 2 {
		t.Fatal("failed query changed authoritative state")
	}
}
