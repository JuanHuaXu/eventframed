package researchjointsequence

import (
	"math"
	"reflect"
	"testing"
)

func snap(m *Model) Model {
	x := *m
	x.base = append([]float64(nil), m.base...)
	x.priors = append([][Atoms]float64(nil), m.priors...)
	x.factors = append([][States][6]float64(nil), m.factors...)
	x.members = append([]member(nil), m.members...)
	return x
}
func TestAtomicLifecycleAndEpochs(t *testing.T) {
	m, e := New([]float64{.3, .7}, 1, 4, Config{Hazard: 1. / 16})
	if e != nil {
		t.Fatal(e)
	}
	other, _ := New([]float64{.3, .7}, 1, 4, Config{Hazard: 1. / 16})
	x, _ := m.Issue(0, 0)
	foreign, _ := other.Issue(0, 0)
	before := snap(m)
	if _, e = m.Resolve(foreign, true, 1); e == nil {
		t.Fatal("foreign owner")
	}
	if !reflect.DeepEqual(before, snap(m)) {
		t.Fatal("foreign mutation")
	}
	if _, e = m.RequestSecond(x, 1); e == nil {
		t.Fatal("unreceived")
	}
	if _, e = m.Resolve(x, true, 1); e != nil {
		t.Fatal(e)
	}
	before = snap(m)
	for _, mode := range []string{"forecast", "uncertainty", "information", "falsification", "predictive"} {
		if _, e = m.Query(x, mode); e != nil {
			t.Fatal(e)
		}
	}
	if !reflect.DeepEqual(before, snap(m)) {
		t.Fatal("hypothetical mutation")
	}
	y, e := m.RequestSecond(x, 2)
	if e != nil {
		t.Fatal(e)
	}
	if e = m.Cancel(y, 3); e != nil {
		t.Fatal(e)
	}
	before = snap(m)
	if _, e = m.Resolve(y, false, 4); e == nil {
		t.Fatal("missing replay")
	}
	if !reflect.DeepEqual(before, snap(m)) {
		t.Fatal("replay mutation")
	}
	if e = m.BeginEpoch(2, 4); e != nil {
		t.Fatal(e)
	}
	if m.Pending() != 0 {
		t.Fatal("pending reset")
	}
	if _, e = m.Resolve(x, false, 5); e == nil {
		t.Fatal("old epoch")
	}
	for i, b := range []float64{.3, .7} {
		q, _, e := m.Predict(i)
		if e != nil || math.Abs(q-b) > 2e-14 {
			t.Fatal("epoch prior", q, e)
		}
	}
}
func TestFaultsCapsAndPriorSupport(t *testing.T) {
	for _, cfg := range []Config{{Hazard: -1}, {Hazard: math.NaN()}, {Hazard: 2}} {
		if _, e := New([]float64{.3, .7}, 1, 4, cfg); e == nil {
			t.Fatal("bad hazard")
		}
	}
	for _, hazard := range []float64{0, 1. / 16, 1} {
		base := []float64{.25, .47, .7, .925}
		m, e := New(base, 1, 4, Config{Hazard: hazard})
		if e != nil {
			t.Fatal(e)
		}
		base[0] = .9
		for i, b := range []float64{.25, .47, .7, .925} {
			q, _, e := m.Predict(i)
			if e != nil || math.Abs(q-b) > 2e-14 {
				t.Fatal("prior mean", q, e)
			}
		}
	}
	m, _ := New([]float64{.3, .7}, 1, 4, Config{Hazard: 1. / 16})
	x, _ := m.Issue(0, 0)
	m.factors[0][0][1] = math.Inf(1)
	before := snap(m)
	if _, e := m.Resolve(x, true, 1); e == nil {
		t.Fatal("nonfinite factor accepted")
	}
	if !reflect.DeepEqual(before, snap(m)) {
		t.Fatal("partial failed replay")
	}
	m, _ = New([]float64{.3, .7}, 1, 1, Config{Hazard: 1. / 16})
	_, _ = m.Issue(0, 0)
	before = snap(m)
	if _, e := m.Issue(1, 1); e == nil {
		t.Fatal("pending cap")
	}
	if !reflect.DeepEqual(before, snap(m)) {
		t.Fatal("cap mutation")
	}
}
