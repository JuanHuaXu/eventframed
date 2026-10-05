package researchpairedvalue

import (
	"math"
	"reflect"
	"testing"
)

func valueState(m *Model) Model {
	s := *m
	s.base = append([]float64(nil), m.base...)
	s.issued = append([]int(nil), m.issued...)
	s.trials = append([]trial(nil), m.trials...)
	s.prior = append([]float64(nil), m.prior...)
	s.latest = append([]float64(nil), m.latest...)
	s.logs = append([]float64(nil), m.logs...)
	s.revision = append([]uint64(nil), m.revision...)
	s.queries = append([]queryCache(nil), m.queries...)
	return s
}

func TestPredictionValueSnapshotAndContracts(t *testing.T) {
	m, e := New([]float64{.3, .8}, 1, 128, Config{Strength: 2, Hazard: 1. / 16})
	if e != nil {
		t.Fatal(e)
	}
	origin, _ := m.Issue(0, 0)
	if _, e = m.Resolve(origin, true, 1); e != nil {
		t.Fatal(e)
	}
	prior := valueState(m)
	a, e := m.PredictionValues([]Ticket{origin}, []float64{.5, .5})
	if e != nil || len(a) != 1 || a[0].Value <= 0 {
		t.Fatalf("valid nonvacuous query: %v %v", a, e)
	}
	if !reflect.DeepEqual(prior, valueState(m)) {
		t.Fatal("query committed state or cache")
	}
	b, e := m.PredictionValues([]Ticket{origin}, []float64{.5, .5})
	if e != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("repeat snapshot changed")
	}
	foreign, _ := New([]float64{.3, .8}, 1, 128, m.cfg)
	ft, _ := foreign.Issue(0, 0)
	foreign.Resolve(ft, true, 1)
	pending, _ := m.Issue(1, 2)
	canceled, _ := m.Issue(1, 3)
	m.Cancel(canceled, 4)
	for _, tc := range []struct {
		origins []Ticket
		weights []float64
	}{
		{nil, []float64{.5, .5}},
		{[]Ticket{origin, origin}, []float64{.5, .5}},
		{[]Ticket{origin}, nil},
		{[]Ticket{origin}, []float64{0, 0}},
		{[]Ticket{origin}, []float64{-.5, 1.5}},
		{[]Ticket{origin}, []float64{math.NaN(), .5}},
		{[]Ticket{origin}, []float64{math.Inf(1), .5}},
		{[]Ticket{{}}, []float64{.5, .5}},
		{[]Ticket{ft}, []float64{.5, .5}},
		{[]Ticket{pending}, []float64{.5, .5}},
		{[]Ticket{canceled}, []float64{.5, .5}},
	} {
		s := valueState(m)
		if got, e := m.PredictionValues(tc.origins, tc.weights); e == nil || got != nil {
			t.Fatal("invalid query accepted")
		}
		if !reflect.DeepEqual(s, valueState(m)) {
			t.Fatal("failed query changed state")
		}
	}
	second, e := m.RequestAudit(origin, 5)
	if e != nil {
		t.Fatal(e)
	}
	for _, ticket := range []Ticket{origin, second} {
		if _, e = m.PredictionValues([]Ticket{ticket}, []float64{.5, .5}); e == nil {
			t.Fatal("requested origin or second ticket accepted")
		}
	}
	if e = m.BeginEpoch(2, 6); e != nil {
		t.Fatal(e)
	}
	if _, e = m.PredictionValues([]Ticket{origin}, []float64{.5, .5}); e == nil {
		t.Fatal("stale epoch accepted")
	}
}

func TestPredictionValueGlobalLocalAndNoHiddenOutcome(t *testing.T) {
	a, _ := New([]float64{.3, .8}, 1, 128, Config{Strength: 2, Hazard: 1. / 16})
	b, _ := New([]float64{.3, .8}, 1, 128, a.cfg)
	ta, _ := a.Issue(0, 0)
	tb, _ := b.Issue(0, 0)
	a.Resolve(ta, false, 1)
	b.Resolve(tb, false, 1)
	// A pending outcome differs in private scratch but is not yet evidence.
	pa, _ := a.Issue(1, 2)
	pb, _ := b.Issue(1, 2)
	a.trials[pa.slot].w1, b.trials[pb.slot].w1 = true, false
	x, ea := a.PredictionValues([]Ticket{ta}, []float64{0, 1})
	y, eb := b.PredictionValues([]Ticket{tb}, []float64{0, 1})
	if ea != nil || eb != nil || !reflect.DeepEqual(x, y) || x[0].Value <= 0 {
		t.Fatal("unrevealed outcome leak or vacuous global effect")
	}
	a.Resolve(pa, true, 3)
	b.Resolve(pb, false, 3)
	x, ea = a.PredictionValues([]Ticket{ta}, []float64{0, 1})
	y, eb = b.PredictionValues([]Ticket{tb}, []float64{0, 1})
	if ea != nil || eb != nil || math.Abs(x[0].Value-y[0].Value) < 1e-8 {
		t.Fatal("global revealed evidence ignored")
	}
	local, _ := a.Issue(0, 4)
	a.Resolve(local, true, 5)
	z, e := a.PredictionValues([]Ticket{ta}, []float64{0, 1})
	if e != nil || math.Abs(z[0].Value-x[0].Value) < 1e-8 {
		t.Fatal("later local evidence ignored")
	}
}

func TestPredictionValueSettledClassStillLearns(t *testing.T) {
	for etaIndex := 0; etaIndex < 3; etaIndex++ {
		m, _ := New([]float64{.3, .8}, 1, 128, Config{Strength: 2, Hazard: 0})
		origin, _ := m.Issue(0, 0)
		m.Resolve(origin, false, 1)
		// Condition on one model class, leaving its member rate uncertain.
		m.weights = [Components]float64{}
		component := etaIndex * Families
		m.weights[component] = 1
		// Remove the other components from the independently normalized law too.
		for c := 0; c < Components; c++ {
			if c != component {
				m.logs[Components+c] = math.Inf(-1)
			}
		}
		v, e := m.PredictionValues([]Ticket{origin}, []float64{1, 0})
		if e != nil {
			t.Fatal(e)
		}
		class, e := m.Query(origin, "falsification")
		if e != nil || math.Abs(class.EdgeCut) > 2e-10 {
			t.Fatal("class-concentration control not settled")
		}
		if etaIndex == 0 && v[0].Value > 2e-10 {
			t.Fatal("noiseless duplicate invented evidence")
		}
		if etaIndex > 0 && v[0].Value <= 1e-6 {
			t.Fatal("predictive value missed uncertain local rate")
		}
	}
}

func TestPredictionValueNumericFaultIsReadOnly(t *testing.T) {
	m, _ := New([]float64{.3, .8}, 1, 128, Config{Strength: 2, Hazard: 1. / 16})
	first, _ := m.Issue(0, 0)
	m.Resolve(first, false, 1)
	before := valueState(m)
	m.logs[0] = math.NaN()
	if values, e := m.PredictionValues([]Ticket{first}, []float64{.5, .5}); e == nil || values != nil {
		t.Fatal("numeric fault accepted")
	}
	if !math.IsNaN(m.logs[0]) {
		t.Fatal("query hid corrupt evidence")
	}
	m.logs[0] = before.logs[0]
	if !reflect.DeepEqual(before, valueState(m)) {
		t.Fatal("fault changed unrelated state")
	}
}
