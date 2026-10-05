package researchmoment_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/researchmoment"
	"github.com/JuanHuaXu/eventframed/internal/researchmomentref"
	"github.com/JuanHuaXu/eventframed/internal/researchprior"
)

func near(t *testing.T, a, b float64) {
	t.Helper()
	if math.Abs(a-b) > 2e-12 {
		t.Fatalf("%.17g != %.17g", a, b)
	}
}
func configs() []researchmoment.Config {
	x := []researchmoment.Config{}
	for _, family := range []string{"narrow", "rich"} {
		for _, prior := range []string{"density", "moment"} {
			for _, s := range []float64{2, 4} {
				for _, shared := range []bool{false, true} {
					x = append(x, researchmoment.Config{Family: family, Prior: prior, Strength: s, Hazard: 1. / 16, Shared: shared})
				}
			}
		}
	}
	return x
}
func TestIndependentPriorAndWholeHistory(t *testing.T) {
	for _, c := range configs() {
		for _, mu := range []float64{.025, .075, .5, .925, .975} {
			a, e := researchmoment.DiscretePrior(mu, c.Strength, c.Prior)
			if e != nil {
				t.Fatal(e)
			}
			b, e := researchmomentref.Prior(mu, c.Strength, c.Prior)
			if e != nil {
				t.Fatal(e)
			}
			for z := range a {
				near(t, a[z], b[z])
			}
		}
		base := []float64{.25, .57, .925}
		m, e := researchmoment.New(base, 1, 192, c)
		if e != nil {
			t.Fatal(e)
		}
		r, e := researchmomentref.New(base, c)
		if e != nil {
			t.Fatal(e)
		}
		tickets := make([]researchmoment.Ticket, 192)
		for k := range tickets {
			q, e := r.Predict(k % 3)
			if e != nil {
				t.Fatal(e)
			}
			tickets[k], e = m.Issue(k%3, int64(k))
			if e != nil {
				t.Fatal(e)
			}
			near(t, q, tickets[k].Forecast())
			if e = r.Issue(k % 3); e != nil {
				t.Fatal(e)
			}
		}
		rng := rand.New(rand.NewSource(740031))
		for _, k := range rng.Perm(192) {
			if k%11 == 0 {
				if e = m.Cancel(tickets[k], 200); e != nil {
					t.Fatal(e)
				}
			} else {
				y := rng.Intn(2) == 1
				receipt, e := m.Resolve(tickets[k], y, 200)
				if e != nil || receipt.Member != k%3 || receipt.TrialOrdinal != k/3+1 || receipt.Forecast != tickets[k].Forecast() {
					t.Fatal("delayed receipt", e)
				}
				if e = r.Resolve(k%3, k/3+1, y); e != nil {
					t.Fatal(e)
				}
			}
			for i := range base {
				a, e := m.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				b, e := r.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				near(t, a, b)
			}
		}
		if m.Pending() != 0 {
			t.Fatal("drain")
		}
	}
}
func TestInterleavedAsOfAndCensoring(t *testing.T) {
	for _, c := range configs() {
		base := []float64{.29, .59, .91}
		m, e := researchmoment.New(base, 1, 192, c)
		if e != nil {
			t.Fatal(e)
		}
		r, e := researchmomentref.New(base, c)
		if e != nil {
			t.Fatal(e)
		}
		type queued struct {
			ticket researchmoment.Ticket
			i, j   int
		}
		queue := []queued{}
		n := [3]int{}
		rng := rand.New(rand.NewSource(740033))
		for at := 0; at < 300; at++ {
			if at < 100 && len(queue) < 10 {
				i := rng.Intn(3)
				q, e := r.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				ticket, e := m.Issue(i, int64(at))
				if e != nil {
					t.Fatal(e)
				}
				near(t, q, ticket.Forecast())
				if e = r.Issue(i); e != nil {
					t.Fatal(e)
				}
				n[i]++
				queue = append(queue, queued{ticket, i, n[i]})
			}
			if len(queue) > 0 && (at >= 100 || rng.Intn(2) == 1) {
				k := rng.Intn(len(queue))
				x := queue[k]
				queue = append(queue[:k], queue[k+1:]...)
				if rng.Intn(5) == 0 {
					if e = m.Cancel(x.ticket, int64(at)); e != nil {
						t.Fatal(e)
					}
				} else {
					y := rng.Intn(2) == 1
					if _, e = m.Resolve(x.ticket, y, int64(at)); e != nil {
						t.Fatal(e)
					}
					if e = r.Resolve(x.i, x.j, y); e != nil {
						t.Fatal(e)
					}
				}
			}
			for i := range base {
				a, e := m.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				b, e := r.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				near(t, a, b)
			}
		}
		if m.Pending() != 0 {
			t.Fatal("pending")
		}
	}
}
func TestDensityNarrowEqualsFrozenV40(t *testing.T) {
	for _, s := range []float64{2, 4} {
		base := []float64{.25, .57, .925}
		m, e := researchmoment.New(base, 1, 192, researchmoment.Config{Family: "narrow", Prior: "density", Strength: s, Hazard: 1. / 16, Shared: true})
		if e != nil {
			t.Fatal(e)
		}
		r, e := researchprior.New(base, 1, 192, researchprior.Config{Center: "raw", Strength: s, Hazard: 1. / 16, Shared: true})
		if e != nil {
			t.Fatal(e)
		}
		a := make([]researchmoment.Ticket, 96)
		b := make([]researchprior.Ticket, 96)
		for k := range a {
			a[k], e = m.Issue(k%3, int64(k))
			if e != nil {
				t.Fatal(e)
			}
			b[k], e = r.Issue(k%3, int64(k))
			if e != nil {
				t.Fatal(e)
			}
			near(t, a[k].Forecast(), b[k].Forecast())
		}
		for _, k := range rand.New(rand.NewSource(740039)).Perm(96) {
			if k%7 == 0 {
				if e = m.Cancel(a[k], 100); e != nil {
					t.Fatal(e)
				}
				if e = r.Cancel(b[k], 100); e != nil {
					t.Fatal(e)
				}
			} else {
				if _, e = m.Resolve(a[k], k%3 == 0, 100); e != nil {
					t.Fatal(e)
				}
				if _, e = r.Resolve(b[k], k%3 == 0, 100); e != nil {
					t.Fatal(e)
				}
			}
			for i := range base {
				x, e := m.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				y, e := r.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				near(t, x, y)
			}
		}
	}
}
func TestExplicitThreePositionPaths(t *testing.T) {
	for _, c := range []researchmoment.Config{{Family: "narrow", Prior: "moment", Strength: 2, Hazard: 0, Shared: true}, {Family: "narrow", Prior: "moment", Strength: 4, Hazard: .2, Shared: true}} {
		m, e := researchmoment.New([]float64{.32, .85}, 1, 128, c)
		if e != nil {
			t.Fatal(e)
		}
		tickets := make([]researchmoment.Ticket, 3)
		for j := range tickets {
			tickets[j], e = m.Issue(0, int64(j))
			if e != nil {
				t.Fatal(e)
			}
		}
		if _, e = m.Resolve(tickets[2], true, 3); e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(tickets[0], false, 4); e != nil {
			t.Fatal(e)
		}
		num, den := 0., 0.
		for h, mu := range []float64{.32, .68, .5} {
			p, e := researchmomentref.Prior(mu, c.Strength, c.Prior)
			if e != nil {
				t.Fatal(e)
			}
			mean := 0.
			for z, v := range p {
				mean += v * (float64(z) + .5) / 21
			}
			for a := 0; a < 21; a++ {
				for b := 0; b < 21; b++ {
					for d := 0; d < 21; d++ {
						ab, bd := c.Hazard*p[b], c.Hazard*p[d]
						if a == b {
							ab += 1 - c.Hazard
						}
						if b == d {
							bd += 1 - c.Hazard
						}
						q0, q2 := (float64(a)+.5)/21, (float64(d)+.5)/21
						mass := []float64{.8, .1, .1}[h] * p[a] * ab * bd * (1 - q0) * q2
						den += mass
						num += mass * ((1-c.Hazard)*q2 + c.Hazard*mean)
					}
				}
			}
		}
		q, e := m.Predict(0)
		if e != nil {
			t.Fatal(e)
		}
		near(t, q, num/den)
	}
}
