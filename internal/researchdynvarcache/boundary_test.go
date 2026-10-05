package researchdynvarcache

import (
	"math"
	"reflect"
	"testing"
)

func TestLifecycleCapsEpochsAndAtomicFailures(t *testing.T) {
	for _, mode := range []string{"noise", "local", "individual"} {
		m, e := New([]float64{.3, .7}, 1, 1, Config{mode, "learn", 1. / 16})
		if e != nil {
			t.Fatal(e)
		}
		x, e := m.Issue(0, 0)
		if e != nil {
			t.Fatal(e)
		}
		other := create(t, []float64{.3, .7}, Config{mode, "learn", 1. / 16})
		foreign, e := other.Issue(0, 0)
		if e != nil {
			t.Fatal(e)
		}
		before := snapshot(m)
		for _, f := range []func() error{
			func() error { _, e := m.Issue(1, 1); return e },
			func() error { _, e := m.Resolve(foreign, true, 1); return e },
			func() error { _, e := m.Resolve(x, true, -1); return e },
			func() error { _, e := m.RequestSecond(x, 1); return e },
			func() error { return m.BeginEpoch(1, 1) },
		} {
			if f() == nil {
				t.Fatal("invalid transition accepted")
			}
			if !reflect.DeepEqual(before, snapshot(m)) {
				t.Fatal("invalid transition published")
			}
		}
		if e = m.Cancel(x, 1); e != nil {
			t.Fatal(e)
		}
		if m.Pending() != 0 {
			t.Fatal("cancel pending")
		}
		if _, e = m.Resolve(x, true, 2); e == nil {
			t.Fatal("cancel replay")
		}
		x, e = m.Issue(0, 2)
		if e != nil {
			t.Fatal(e)
		}
		r, e := m.Resolve(x, true, 3)
		if e != nil {
			t.Fatal(e)
		}
		if r.Ordinal != 2 || r.IssuedAt != 2 || r.ArrivedAt != 3 || r.Measurement != 1 || r.Epoch != 1 {
			t.Fatal(r)
		}
		a, e := m.RequestSecond(x, 4)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.RequestSecond(x, 4); e == nil {
			t.Fatal("double request")
		}
		if e = m.Cancel(a, 5); e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(a, true, 6); e == nil {
			t.Fatal("cancel pair replay")
		}
		if e = m.BeginEpoch(2, 6); e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(x, true, 7); e == nil {
			t.Fatal("stale epoch")
		}
		if m.Pending() != 0 || m.count != 0 {
			t.Fatal("epoch state")
		}
		w, e := m.ModelWeights()
		if e != nil {
			t.Fatal(e)
		}
		for _, p := range w {
			closeValue(t, "epoch prior", p, 1./3)
		}
		for n := 0; n < MaxTrials; n++ {
			x, e = m.Issue(0, int64(8+n))
			if e != nil {
				t.Fatal(e)
			}
			if _, e = m.Resolve(x, n%2 == 0, int64(8+n)); e != nil {
				t.Fatal(e)
			}
		}
		before = snapshot(m)
		if _, e = m.Issue(0, 100); e == nil {
			t.Fatal("ordinal cap")
		}
		if !reflect.DeepEqual(before, snapshot(m)) {
			t.Fatal("cap publication")
		}
	}
}

func TestFaultAtomicity(t *testing.T) {
	for _, mode := range []string{"noise", "local", "individual"} {
		for _, bad := range []float64{math.Inf(1), -1} {
			m := create(t, []float64{.3, .7}, Config{mode, "learn", 1. / 16})
			x, e := m.Issue(0, 0)
			if e != nil {
				t.Fatal(e)
			}
			for h := range eta {
				for z := 0; z < RateAtoms; z++ {
					m.factors[0][h][z][1] = bad
				}
			}
			before := snapshot(m)
			if _, e = m.Resolve(x, true, 1); e == nil {
				t.Fatal("fault accepted")
			}
			if !reflect.DeepEqual(before, snapshot(m)) {
				t.Fatal("partial factor commit")
			}
		}
		m := create(t, []float64{.3, .7}, Config{mode, "learn", 1. / 16})
		x, e := m.Issue(0, 0)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(x, true, 0); e != nil {
			t.Fatal(e)
		}
		m.members[1].latest.log[0] = math.Inf(1)
		before := snapshot(m)
		if _, e = m.RequestSecond(x, 1); e == nil {
			t.Fatal("invalid dependent evidence accepted")
		}
		if !reflect.DeepEqual(before, snapshot(m)) {
			t.Fatal("partial request")
		}
	}
}

func TestConstructorAndQueryRejections(t *testing.T) {
	base := []float64{.3, .7}
	cfg := Config{"noise", "learn", 1. / 16}
	for _, bad := range []Config{{"unknown", "learn", 0}, {"noise", "unknown", 0}, {"noise", "learn", math.NaN()}, {"noise", "learn", -1}, {"noise", "learn", 1.1}} {
		if _, e := New(base, 1, 128, bad); e == nil {
			t.Fatal("config", bad)
		}
	}
	for _, b := range [][]float64{nil, {.3}, {.1, .7}, {.3, 1}, {math.NaN(), .7}} {
		if _, e := New(b, 1, 128, cfg); e == nil {
			t.Fatal("base", b)
		}
	}
	for _, cap := range []int{0, 257} {
		if _, e := New(base, 1, cap, cfg); e == nil {
			t.Fatal("cap", cap)
		}
	}
	if _, e := New(base, 0, 128, cfg); e == nil {
		t.Fatal("zero epoch")
	}
	m := create(t, base, cfg)
	x, e := m.Issue(0, 0)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Query(x, "forecast"); e == nil {
		t.Fatal("pending query")
	}
	if _, e = m.Resolve(x, true, 0); e != nil {
		t.Fatal(e)
	}
	before := snapshot(m)
	if _, e = m.Query(x, "bad"); e == nil {
		t.Fatal("query mode")
	}
	if !reflect.DeepEqual(before, snapshot(m)) {
		t.Fatal("invalid query publication")
	}
}

func TestFutureForkAndArrivalClock(t *testing.T) {
	for _, mode := range []string{"noise", "local", "individual"} {
		cfg := Config{mode, "learn", 1. / 16}
		a := create(t, []float64{.3, .7}, cfg)
		b := create(t, []float64{.3, .7}, cfg)
		var ta, tb []Ticket
		for n := 0; n < 16; n++ {
			x, e := a.Issue(n%2, int64(n))
			if e != nil {
				t.Fatal(e)
			}
			y, e := b.Issue(n%2, int64(n*100))
			if e != nil {
				t.Fatal(e)
			}
			ta = append(ta, x)
			tb = append(tb, y)
			// Pending rows do not read these future values; the forks diverge only on reveal.
			if n >= 5 {
				j := n - 5
				if _, e = a.Resolve(ta[j], j%3 == 0, int64(n)); e != nil {
					t.Fatal(e)
				}
				if _, e = b.Resolve(tb[j], j%3 == 0, int64(n*100)); e != nil {
					t.Fatal(e)
				}
			}
			for i := 0; i < 2; i++ {
				p, o, e := a.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				q, v, e := b.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				closeValue(t, "arrival clean", p, q)
				closeValue(t, "arrival noise", o, v)
			}
		}
		for _, query := range []string{"forecast", "information", "falsification", "predictive", "model_class", "noise_class"} {
			x, e := a.Query(ta[0], query)
			if e != nil {
				t.Fatal(e)
			}
			y, e := b.Query(tb[0], query)
			if e != nil {
				t.Fatal(e)
			}
			if x != y {
				t.Fatal("arrival-time dependence", query, x, y)
			}
		}
		if _, e := a.Resolve(ta[15], true, 20); e != nil {
			t.Fatal(e)
		}
		if _, e := b.Resolve(tb[15], false, 2000); e != nil {
			t.Fatal(e)
		}
		p, _, _ := a.Predict(1)
		q, _, _ := b.Predict(1)
		if math.Abs(p-q) < 1e-5 {
			t.Fatal("opposite revealed future has no effect")
		}
	}
}

func TestImpossibleNoiseCannotRevive(t *testing.T) {
	for _, mode := range []string{"noise", "local", "individual"} {
		m := create(t, []float64{.3, .7}, Config{mode, "learn", 1})
		x, e := m.Issue(0, 0)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(x, true, 0); e != nil {
			t.Fatal(e)
		}
		a, e := m.RequestSecond(x, 1)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(a, false, 2); e != nil {
			t.Fatal(e)
		}
		for n := 3; n < 67; n++ {
			if n == 66 {
				break
			}
			x, e = m.Issue(0, int64(n))
			if e != nil {
				t.Fatal(e)
			}
			if _, e = m.Resolve(x, n%3 == 0, int64(n)); e != nil {
				t.Fatal(e)
			}
		}
		for a := 0; a < 3; a++ {
			if !math.IsInf(m.members[0].latest.log[3*a], -1) {
				t.Fatal("zero-noise hypothesis revived")
			}
		}
		v, e := m.currentView(-1, nil)
		if e != nil {
			t.Fatal(e)
		}
		w, e := m.memberWeights(0, v)
		if e != nil {
			t.Fatal(e)
		}
		for a := 0; a < 3; a++ {
			if w[3*a] != 0 {
				t.Fatal("eliminated noise mass", w)
			}
		}
	}
}
