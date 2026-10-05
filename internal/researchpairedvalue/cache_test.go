package researchpairedvalue

import (
	orig "github.com/JuanHuaXu/eventframed/internal/researchpaired"
	"math"
	"reflect"
	"testing"
)

func TestMemoBitwiseLocalGlobalAndEpoch(t *testing.T) {
	m, _ := New([]float64{.3, .9}, 1, 128, Config{Strength: 2, Hazard: 1. / 16})
	o, _ := orig.New([]float64{.3, .9}, 1, 128, Config{Strength: 2, Hazard: 1. / 16})
	a, _ := m.Issue(0, 1)
	oa, _ := o.Issue(0, 1)
	b, _ := m.Issue(1, 2)
	ob, _ := o.Issue(1, 2)
	m.Resolve(a, true, 3)
	o.Resolve(oa, true, 3)
	m.Resolve(b, true, 4)
	o.Resolve(ob, true, 4)
	check := func() {
		t.Helper()
		x, e := m.Options(a)
		if e != nil {
			t.Fatal(e)
		}
		y, e := o.Options(oa)
		if e != nil || x != y {
			t.Fatal("bitwise option", x, y, e)
		}
	}
	check()
	cache := m.queries[0]
	rev := m.revision[0]
	check()
	if cache != m.queries[0] {
		t.Fatal("cache repeat")
	}
	s, _ := m.RequestAudit(b, 5)
	os, _ := o.RequestAudit(ob, 5)
	m.Resolve(s, false, 6)
	o.Resolve(os, false, 6)
	check()
	if cache != m.queries[0] || rev != m.revision[0] {
		t.Fatal("global update fenced member cache")
	}
	c, _ := m.Issue(0, 7)
	oc, _ := o.Issue(0, 7)
	check()
	if m.queries[0].revision == rev {
		t.Fatal("issue fence missing")
	}
	if e := m.Cancel(c, 8); e != nil {
		t.Fatal(e)
	}
	if e := o.Cancel(oc, 8); e != nil {
		t.Fatal(e)
	}
	check()
	if m.queries[0].revision != m.revision[0] {
		t.Fatal("cancel cache stale")
	}
	sa, e := m.RequestAudit(a, 9)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Options(a); e == nil {
		t.Fatal("pending query replay")
	}
	if e = m.Cancel(sa, 10); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Options(a); e == nil {
		t.Fatal("cancelled query replay")
	}
	if e = m.BeginEpoch(2, 11); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Options(a); e == nil {
		t.Fatal("old epoch")
	}
	for _, c := range m.queries {
		if c.valid {
			t.Fatal("epoch retained cache")
		}
	}
}

func TestMemoRevivalAndAtomicFaults(t *testing.T) {
	m, _ := New([]float64{.3, .9}, 1, 128, Config{Strength: 2, Hazard: 1. / 16})
	a, _ := m.Issue(0, 1)
	m.Resolve(a, true, 2)
	w := m.weights
	m.weights[0] = 0
	for c := 1; c < Components; c++ {
		m.weights[c] /= 1 - w[0]
	}
	if _, e := m.Options(a); e != nil {
		t.Fatal(e)
	}
	cache := m.queries[0]
	m.weights = w
	if _, e := m.Options(a); e != nil {
		t.Fatal(e)
	}
	if cache != m.queries[0] {
		t.Fatal("underflow revival changed conditional cache")
	}
	b, _ := m.Issue(1, 3)
	revision := append([]uint64(nil), m.revision...)
	weights := m.weights
	m.prior[Families*Atoms] = math.NaN()
	if _, e := m.Resolve(b, true, 4); e == nil {
		t.Fatal("NaN evidence accepted")
	}
	if !reflect.DeepEqual(revision, m.revision) || cache != m.queries[0] || weights != m.weights || m.pending != 1 || m.clock != 3 {
		t.Fatal("partial authoritative/cache publication")
	}
	m.queries[0].yes[0] = math.NaN()
	if _, e := m.Options(a); e == nil {
		t.Fatal("NaN cached marginal served")
	}
	if weights != m.weights || m.pending != 1 || m.clock != 3 {
		t.Fatal("failed cached query mutated posterior")
	}
}
