package researchpairedfast_test

import (
	old "github.com/JuanHuaXu/eventframed/internal/researchnoisemoment"
	p "github.com/JuanHuaXu/eventframed/internal/researchpairedfast"
	ref "github.com/JuanHuaXu/eventframed/internal/researchpairedref"
	"math"
	"testing"
)

func near(t *testing.T, a, b float64) {
	t.Helper()
	if math.IsNaN(a) || math.IsNaN(b) || math.Abs(a-b) > 2e-10 {
		t.Fatalf("reference %.17g %.17g", a, b)
	}
}
func TestPairedDelayedJointReference(t *testing.T) {
	base := []float64{.25, .47, .78, .925}
	cfg := p.Config{Strength: 2, Hazard: 1. / 16}
	m, e := p.New(base, 1, 512, cfg)
	if e != nil {
		t.Fatal(e)
	}
	r, e := ref.New(base, cfg)
	if e != nil {
		t.Fatal(e)
	}
	first := make([]p.Ticket, 256)
	seconds := make([]p.Ticket, 256)
	for k := range first {
		first[k], e = m.Issue(k%4, int64(k))
		if e != nil {
			t.Fatal(e)
		}
		if e = r.Issue(k % 4); e != nil {
			t.Fatal(e)
		}
	}
	for step := 0; step < 256; step++ {
		k := (step*73 + 251) % 256
		at := int64(256 + step)
		value := k%7 < 3
		if _, e = m.Resolve(first[k], value, at); e != nil {
			t.Fatal(e)
		}
		if e = r.Observe(k%4, k/4+1, 1, value); e != nil {
			t.Fatal(e)
		}
		a, e := m.Options(first[k])
		if e != nil {
			t.Fatal(e)
		}
		b, e := r.Options(k%4, k/4+1)
		if e != nil {
			t.Fatal(e)
		}
		near(t, a.Observed, b.Observed)
		near(t, a.Information, b.Information)
		near(t, a.EdgeCut, b.EdgeCut)
		near(t, a.Uncertainty, b.Uncertainty)
		seconds[k], e = m.RequestAudit(first[k], at)
		if e != nil {
			t.Fatal(e)
		}
		if e = r.Audit(k%4, k/4+1); e != nil {
			t.Fatal(e)
		}
	}
	for step := 0; step < 256; step++ {
		k := (step*41 + 17) % 256
		y := k%11 < 6
		if _, e = m.Resolve(seconds[k], y, int64(512+step)); e != nil {
			t.Fatal(e)
		}
		if e = r.Observe(k%4, k/4+1, 2, y); e != nil {
			t.Fatal(e)
		}
		w, e := r.NoiseWeights()
		if e != nil {
			t.Fatal(e)
		}
		for j, v := range m.NoiseWeights() {
			near(t, v, w[j])
		}
		for i := range base {
			q, o, e := m.Predict(i)
			if e != nil {
				t.Fatal(e)
			}
			v, u, e := r.Predict(i)
			if e != nil {
				t.Fatal(e)
			}
			near(t, q, v)
			near(t, o, u)
		}
	}
}
func TestPairedSingleMarginalAndFutureFork(t *testing.T) {
	base := []float64{.3, .9}
	cfg := p.Config{Strength: 2, Hazard: 1. / 16}
	m, _ := p.New(base, 1, 128, cfg)
	legacy, _ := old.NewMixture(base, 1, 128, old.Config{Family: "rich", Prior: "moment", Strength: 2, Hazard: 1. / 16, Shared: true})
	ts := make([]p.Ticket, 32)
	os := make([]old.MixtureTicket, 32)
	for k := range ts {
		var e error
		ts[k], e = m.Issue(k%2, int64(k))
		if e != nil {
			t.Fatal(e)
		}
		os[k], e = legacy.Issue(k%2, int64(k))
		if e != nil {
			t.Fatal(e)
		}
		near(t, ts[k].Forecast(), os[k].Forecast())
	}
	for step := 0; step < 32; step++ {
		k := (step*7 + 31) % 32
		y := k%3 == 0
		at := int64(32 + step)
		if _, e := m.Resolve(ts[k], y, at); e != nil {
			t.Fatal(e)
		}
		if _, e := legacy.Resolve(os[k], y, at); e != nil {
			t.Fatal(e)
		}
		for i := range base {
			q, o, e := m.Predict(i)
			if e != nil {
				t.Fatal(e)
			}
			v, e := legacy.Predict(i)
			if e != nil {
				t.Fatal(e)
			}
			u, e := legacy.ObservedPredict(i)
			if e != nil {
				t.Fatal(e)
			}
			near(t, q, v)
			near(t, o, u)
		}
	}
	a, _ := p.New(base, 1, 128, cfg)
	b, _ := p.New(base, 1, 128, cfg)
	fa, _ := a.Issue(0, 1)
	fb, _ := b.Issue(0, 1)
	a.Resolve(fa, true, 2)
	b.Resolve(fb, true, 2)
	sa, _ := a.RequestAudit(fa, 3)
	sb, _ := b.RequestAudit(fb, 3)
	if sa.Forecast() != sb.Forecast() {
		t.Fatal("unrevealed second leak")
	}
	a.Resolve(sa, true, 4)
	b.Resolve(sb, false, 4)
	qa, _, _ := a.Predict(0)
	qb, _, _ := b.Predict(0)
	if math.Abs(qa-qb) < 1e-4 {
		t.Fatal("second fork vacuous")
	}
}
