package researchpaired

import (
	"math"
	"reflect"
	"testing"
)

func testConfig() Config { return Config{Strength: 2, Hazard: 1. / 16} }
func TestPairedLifecycleAndAtomicFailure(t *testing.T) {
	m, e := New([]float64{.3, .9}, 1, 2, testConfig())
	if e != nil {
		t.Fatal(e)
	}
	other, _ := New([]float64{.3, .9}, 1, 2, testConfig())
	a, _ := m.Issue(0, 1)
	b, _ := m.Issue(1, 2)
	if _, e = m.RequestAudit(a, 2); e == nil {
		t.Fatal("unknown first")
	}
	if _, e = other.Resolve(a, true, 2); e == nil {
		t.Fatal("owner")
	}
	if _, e = m.Resolve(a, true, 1); e == nil {
		t.Fatal("clock")
	}
	if _, e = m.Issue(0, 3); e == nil {
		t.Fatal("cap")
	}
	served := a.Forecast()
	a.q = math.NaN()
	r, e := m.Resolve(a, true, 3)
	if e != nil || r.Forecast != served || r.Measurement != 1 || r.IssuedAt != 1 {
		t.Fatal("private first receipt", r, e)
	}
	second, e := m.RequestAudit(a, 4)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.RequestAudit(a, 4); e == nil {
		t.Fatal("duplicate audit")
	}
	option := second.Forecast()
	second.q = math.NaN()
	r, e = m.Resolve(second, false, 5)
	if e != nil || r.Forecast != option || r.Measurement != 2 {
		t.Fatal("private second receipt", r, e)
	}
	if m.NoiseWeights()[0] != 0 {
		t.Fatal("eta0 impossible model not eliminated")
	}
	if _, _, e = m.Predict(0); e != nil {
		t.Fatal("elimination poisoned usable model", e)
	}
	if _, e = m.Resolve(second, false, 5); e == nil {
		t.Fatal("second replay")
	}
	trials := append([]trial(nil), m.trials...)
	latest := append([]float64(nil), m.latest...)
	logs := append([]float64(nil), m.logs...)
	weights := m.weights
	m.prior[Families*Atoms] = math.NaN()
	if _, e = m.Resolve(b, true, 6); e == nil {
		t.Fatal("NaN publication")
	}
	if !reflect.DeepEqual(trials, m.trials) || !reflect.DeepEqual(latest, m.latest) || !reflect.DeepEqual(logs, m.logs) || weights != m.weights || m.pending != 1 || m.clock != 5 {
		t.Fatal("partial numerical publication")
	}
	if e = m.BeginEpoch(2, 7); e != nil || m.pending != 0 {
		t.Fatal("epoch repair", e)
	}
	if _, e = m.Resolve(b, true, 7); e == nil {
		t.Fatal("old epoch")
	}
}
func TestPairedCancellationUnknownAndCaps(t *testing.T) {
	m, _ := New([]float64{.3, .9}, 1, 128, testConfig())
	a, _ := m.Issue(0, 1)
	m.Resolve(a, true, 2)
	s, _ := m.RequestAudit(a, 3)
	w := m.weights
	if e := m.Cancel(s, 4); e != nil || w != m.weights {
		t.Fatal("cancel changes evidence", e)
	}
	if _, e := m.Resolve(s, false, 4); e == nil {
		t.Fatal("cancel replay")
	}
	b, _ := m.Issue(1, 5)
	if e := m.Cancel(b, 6); e != nil {
		t.Fatal(e)
	}
	if _, e := m.RequestAudit(b, 7); e == nil {
		t.Fatal("cancelled first")
	}
	for j := 1; j < 64; j++ {
		if _, e := m.Issue(0, int64(j+7)); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := m.Issue(0, 100); e == nil {
		t.Fatal("history")
	}
	if e := m.BeginEpoch(1, 100); e == nil {
		t.Fatal("same epoch")
	}
}
func TestPairedContracts(t *testing.T) {
	for _, c := range []Config{{math.NaN(), .1}, {0, .1}, {33, .1}, {2, math.NaN()}, {2, 1}, {2, -.1}} {
		if _, e := New([]float64{.3, .9}, 1, 128, c); e == nil {
			t.Fatal("config")
		}
	}
	for _, b := range [][]float64{{.3}, {.1, .9}, {.3, math.NaN()}, make([]float64, 201)} {
		if _, e := New(b, 1, 1, testConfig()); e == nil {
			t.Fatal("base")
		}
	}
	for _, x := range [][2]int{{0, 1}, {1, 0}, {1, 257}} {
		if _, e := New([]float64{.3, .9}, uint64(x[0]), x[1], testConfig()); e == nil {
			t.Fatal("epoch/cap")
		}
	}
}
