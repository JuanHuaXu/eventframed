// Matched public-API timing. Neither model nor its private caches are modified.
package researchmeanperfcheck

import (
	"math"
	"testing"

	v75 "github.com/JuanHuaXu/eventframed/internal/researchmeananchor"
	v74 "github.com/JuanHuaXu/eventframed/internal/researchmeanjoint"
)

var value float64

func setup74(t testing.TB) (*v74.Model, [2]v74.Ticket) {
	t.Helper()
	base := make([]float64, 150)
	for i := range base {
		base[i] = .25 + .675*float64(i)/149
	}
	m, e := v74.New(base, 1, 19200, v74.Config{Mode: "noise", Family: "learn", Means: "learn", Hazard: 1. / 16})
	if e != nil {
		t.Fatal(e)
	}
	var tickets [2]v74.Ticket
	for n := 0; n < 16; n++ {
		for i := range base {
			x, e := m.Issue(i, int64(n))
			if e != nil {
				t.Fatal(e)
			}
			if i == 0 {
				tickets[1] = x
				if n == 0 {
					tickets[0] = x
				}
			}
			if _, e = m.Resolve(x, (i+n)%3 != 0, int64(n)); e != nil {
				t.Fatal(e)
			}
		}
	}
	return m, tickets
}
func setup75(t testing.TB) (*v75.Model, [2]v75.Ticket) {
	t.Helper()
	base := make([]float64, 150)
	for i := range base {
		base[i] = .25 + .675*float64(i)/149
	}
	m, e := v75.New(base, 1, 19200, v75.Config{Mode: "noise", Family: "learn", Means: "learn", Hazard: 1. / 16})
	if e != nil {
		t.Fatal(e)
	}
	var tickets [2]v75.Ticket
	for n := 0; n < 16; n++ {
		for i := range base {
			x, e := m.Issue(i, int64(n))
			if e != nil {
				t.Fatal(e)
			}
			if i == 0 {
				tickets[1] = x
				if n == 0 {
					tickets[0] = x
				}
			}
			if _, e = m.Resolve(x, (i+n)%3 != 0, int64(n)); e != nil {
				t.Fatal(e)
			}
		}
	}
	return m, tickets
}
func TestMatchedPublicAPIFixture(t *testing.T) {
	m, a := setup74(t)
	r, b := setup75(t)
	checks := 0
	maximum := 0.
	near := func(x, y float64) {
		t.Helper()
		if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) || math.Abs(x-y) > 2e-10 {
			t.Fatalf("%.17g != %.17g", x, y)
		}
		checks++
		maximum = math.Max(maximum, math.Abs(x-y))
	}
	for i := 0; i < 150; i++ {
		q, o, e := m.Predict(i)
		if e != nil {
			t.Fatal(e)
		}
		s, u, e := r.Predict(i)
		if e != nil {
			t.Fatal(e)
		}
		near(q, s)
		near(o, u)
		p, e := m.Posterior(i)
		if e != nil {
			t.Fatal(e)
		}
		v, e := r.Posterior(i)
		if e != nil {
			t.Fatal(e)
		}
		for k := range p {
			near(p[k], v[k])
		}
	}
	for j := 0; j < 2; j++ {
		for _, mode := range []string{"forecast", "uncertainty", "information", "falsification", "predictive", "model_class", "noise_class"} {
			x, e := m.Query(a[j], mode)
			if e != nil {
				t.Fatal(e)
			}
			y, e := r.Query(b[j], mode)
			if e != nil {
				t.Fatal(e)
			}
			for k, v := range []float64{x.Observed, x.Uncertainty, x.Information, x.EdgeCut, x.Value, x.ClassGain} {
				near(v, []float64{y.Observed, y.Uncertainty, y.Information, y.EdgeCut, y.Value, y.ClassGain}[k])
			}
		}
	}
	t.Logf("%d final public law/posterior/query scalar comparisons, max defect %.17g; same complete 150-member/16-round fixture, not whole-cohort equivalence", checks, maximum)
}
func BenchmarkMatchedQuery(b *testing.B) {
	for _, position := range []string{"old", "latest"} {
		for _, mode := range []string{"model_class", "noise_class", "predictive"} {
			j := 0
			if position == "latest" {
				j = 1
			}
			b.Run("v74/"+position+"/"+mode, func(b *testing.B) {
				m, x := setup74(b)
				b.ReportAllocs()
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					o, e := m.Query(x[j], mode)
					if e != nil {
						b.Fatal(e)
					}
					value = o.ClassGain + o.Value
				}
			})
			b.Run("v75/"+position+"/"+mode, func(b *testing.B) {
				m, x := setup75(b)
				b.ReportAllocs()
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					o, e := m.Query(x[j], mode)
					if e != nil {
						b.Fatal(e)
					}
					value = o.ClassGain + o.Value
				}
			})
		}
	}
}
