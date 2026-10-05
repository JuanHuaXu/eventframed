package researchsharedsequence

import (
	"math"
	"reflect"
	"testing"
)

func snapshot(m *Model) Model {
	x := *m
	x.base = append([]float64(nil), m.base...)
	x.priors = append([]localVector(nil), m.priors...)
	x.localFactors = append([][66][6]float64(nil), m.localFactors...)
	x.fields = append([][FieldAtoms]float64(nil), m.fields...)
	x.sharedFactors = append([][SharedStates][6]float64(nil), m.sharedFactors...)
	x.members = append([]member(nil), m.members...)
	x.rows = append([]row(nil), m.rows...)
	x.checkpoints = append([]checkpoint(nil), m.checkpoints...)
	return x
}
func TestPriorAndLifecycle(t *testing.T) {
	for _, mode := range []string{"local", "noise", "shared", "hybrid"} {
		base := []float64{.25, .47, .7, .925}
		m, e := New(base, 1, 128, Config{Mode: mode, Hazard: 1. / 16})
		if e != nil {
			t.Fatal(e)
		}
		for i, b := range base {
			q, _, e := m.Predict(i)
			if e != nil || math.Abs(q-b) > 3e-14 {
				t.Fatal("prior", mode, i, q, e)
			}
		}
		base[0] = .9
		q, _, _ := m.Predict(0)
		if math.Abs(q-.25) > 3e-14 {
			t.Fatal("alias")
		}
		var tickets []Ticket
		for n := 0; n < 132; n++ {
			x, e := m.Issue(n%4, int64(n))
			if e != nil {
				t.Fatal(e)
			}
			tickets = append(tickets, x)
			if _, e = m.Resolve(x, n%3 != 0, int64(n)); e != nil {
				t.Fatal(mode, n, e)
			}
		}
		x := tickets[0]
		before := snapshot(m)
		for _, query := range []string{"forecast", "uncertainty", "information", "falsification", "predictive"} {
			if _, e = m.Query(x, query); e != nil {
				t.Fatal(mode, query, e)
			}
		}
		if !reflect.DeepEqual(before, snapshot(m)) {
			t.Fatal("hypothetical publication")
		}
		audit, e := m.RequestSecond(x, 133)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(audit, true, 134); e != nil {
			t.Fatal(e)
		}
		before = snapshot(m)
		if _, e = m.Resolve(audit, false, 135); e == nil {
			t.Fatal("replay")
		}
		if !reflect.DeepEqual(before, snapshot(m)) {
			t.Fatal("replay publication")
		}
		if e = m.BeginEpoch(2, 136); e != nil {
			t.Fatal(e)
		}
		if m.Pending() != 0 {
			t.Fatal("pending epoch")
		}
		if _, e = m.Resolve(x, false, 137); e == nil {
			t.Fatal("old epoch")
		}
	}
}
func TestFaultsAndCaps(t *testing.T) {
	for _, mode := range []string{"local", "noise", "shared", "hybrid"} {
		m, _ := New([]float64{.3, .7}, 1, 128, Config{Mode: mode, Hazard: 1. / 16})
		x, _ := m.Issue(0, 0)
		if m.needsShared() {
			for z := range m.sharedFactors[0] {
				m.sharedFactors[0][z][1] = math.Inf(1)
			}
		} else {
			m.localFactors[0][0][1] = math.Inf(1)
		}
		before := snapshot(m)
		if _, e := m.Resolve(x, true, 1); e == nil {
			t.Fatal("fault", mode)
		}
		if !reflect.DeepEqual(before, snapshot(m)) {
			t.Fatal("partial fault", mode)
		}
	}
	m, _ := New([]float64{.3, .7}, 1, 1, Config{Mode: "hybrid", Hazard: 1. / 16})
	x, _ := m.Issue(0, 0)
	before := snapshot(m)
	if _, e := m.Issue(1, 1); e == nil {
		t.Fatal("pending cap")
	}
	if !reflect.DeepEqual(before, snapshot(m)) {
		t.Fatal("cap publication")
	}
	other, _ := New([]float64{.3, .7}, 1, 1, Config{Mode: "hybrid", Hazard: 1. / 16})
	foreign, _ := other.Issue(0, 0)
	if _, e := m.Resolve(foreign, true, 1); e == nil {
		t.Fatal("owner")
	}
	if e := m.Cancel(x, 1); e != nil {
		t.Fatal(e)
	}
	if _, e := m.Resolve(x, true, 2); e == nil {
		t.Fatal("cancel replay")
	}
}
func TestZeroNoiseElimination(t *testing.T) {
	m, _ := New([]float64{.3, .7}, 1, 128, Config{Mode: "hybrid", Hazard: 1. / 16})
	x, _ := m.Issue(0, 0)
	if _, e := m.Resolve(x, true, 0); e != nil {
		t.Fatal(e)
	}
	a, e := m.RequestSecond(x, 1)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Resolve(a, false, 2); e != nil {
		t.Fatal(e)
	}
	if !math.IsInf(m.members[0].latest.log[0], -1) {
		t.Fatal("zero-noise support retained")
	}
	for n := 3; n < 20; n++ {
		x, e := m.Issue(n%2, int64(n))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(x, n%3 == 0, int64(n)); e != nil {
			t.Fatal(e)
		}
	}
	weights, e := m.ModelWeights()
	if e != nil {
		t.Fatal(e)
	}
	sum := 0.
	for _, w := range weights {
		if !finite(w) || w < 0 {
			t.Fatal("model evidence")
		}
		sum += w
	}
	if math.Abs(sum-1) > 1e-12 {
		t.Fatal("model normalization")
	}
}
