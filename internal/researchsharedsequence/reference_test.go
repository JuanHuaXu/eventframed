package researchsharedsequence_test

import (
	legacy "github.com/JuanHuaXu/eventframed/internal/researchjointsequence"
	candidate "github.com/JuanHuaXu/eventframed/internal/researchsharedsequence"
	reference "github.com/JuanHuaXu/eventframed/internal/researchsharedsequenceref"
	"math"
	"testing"
)

func near(t *testing.T, a, b float64) {
	t.Helper()
	if math.IsNaN(a) || math.IsNaN(b) || math.Abs(a-b) > 3e-10 {
		t.Fatalf("independent mismatch %.17g %.17g", a, b)
	}
}
func TestIndependentDelayedJoint(t *testing.T) {
	checks := 0
	for _, mode := range []string{"local", "noise", "shared", "hybrid"} {
		for _, hazard := range []float64{0, 1. / 16, 1} {
			for _, delay := range []int{0, 3, 11} {
				base := []float64{.25, .47, .7, .925}
				m, e := candidate.New(base, 1, 128, candidate.Config{Mode: mode, Hazard: hazard})
				if e != nil {
					t.Fatal(e)
				}
				r, e := reference.New(base, mode, hazard)
				if e != nil {
					t.Fatal(e)
				}
				tickets := make([]candidate.Ticket, 32)
				type event struct {
					at, n  int
					second bool
				}
				queue := []event{}
				verify := func() {
					for i := range base {
						a, b, e := m.Predict(i)
						if e != nil {
							t.Fatal(e)
						}
						c, d, e := r.Predict(i)
						if e != nil {
							t.Fatal(e)
						}
						near(t, a, c)
						near(t, b, d)
						checks += 2
					}
					mw, e := m.ModelWeights()
					if e != nil {
						t.Fatal(e)
					}
					rw := r.ModelWeights()
					for j := range mw {
						near(t, mw[j], rw[j])
						checks++
					}
				}
				deliver := func(at int) {
					for j := 0; j < len(queue); {
						x := queue[j]
						if x.at > at {
							j++
							continue
						}
						queue = append(queue[:j], queue[j+1:]...)
						i, n := x.n%4, x.n/4+1
						measurement := 1
						value := x.n%3 != 0
						if x.second {
							measurement = 2
							value = x.n%5 == 0
						}
						if _, e = m.Resolve(tickets[x.n], value, int64(at)); e != nil {
							t.Fatal(mode, hazard, delay, "candidate receipt", e)
						}
						if e = r.Observe(i, n, measurement, value); e != nil {
							t.Fatal(mode, hazard, delay, "reference receipt", e)
						}
						verify()
						if !x.second && x.n%5 == 0 {
							a, e := m.Query(tickets[x.n], "predictive")
							if e != nil {
								t.Fatal(mode, hazard, delay, "candidate query", e)
							}
							b, e := r.Query(i, n)
							if e != nil {
								t.Fatal(mode, hazard, delay, "reference query", e)
							}
							near(t, a.Observed, b.Observed)
							near(t, a.Value, b.Value)
							for _, policy := range []string{"information", "falsification"} {
								o, e := m.Query(tickets[x.n], policy)
								if e != nil {
									t.Fatal(e)
								}
								near(t, o.Information, b.Information)
								near(t, o.EdgeCut, b.EdgeCut)
							}
							checks += 6
							tickets[x.n], e = m.RequestSecond(tickets[x.n], int64(at))
							if e != nil {
								t.Fatal(e)
							}
							if e = r.Audit(i, n); e != nil {
								t.Fatal(e)
							}
							queue = append(queue, event{at + 13, x.n, true})
						}
					}
				}
				for n := 0; n < 32; n++ {
					at := n * 10
					deliver(at)
					x, e := m.Issue(n%4, int64(at))
					if e != nil {
						t.Fatal(e)
					}
					tickets[n] = x
					if e = r.Issue(n % 4); e != nil {
						t.Fatal(e)
					}
					queue = append(queue, event{at + delay, n, false})
					deliver(at)
					verify()
				}
				deliver(10000)
				deliver(20000)
				if m.Pending() != 0 {
					t.Fatal("pending")
				}
			}
		}
	}
	t.Log("independent forecast,model-weight and acquisition comparisons", checks)
}
func TestLocalLegacyEquivalence(t *testing.T) {
	base := []float64{.25, .47, .7, .925}
	a, _ := candidate.New(base, 1, 128, candidate.Config{Mode: "local", Hazard: 1. / 16})
	b, _ := legacy.New(base, 1, 128, legacy.Config{Hazard: 1. / 16})
	checks := 0
	for n := 0; n < 160; n++ {
		i := n % 4
		x, e := a.Issue(i, int64(n))
		if e != nil {
			t.Fatal(e)
		}
		y, e := b.Issue(i, int64(n))
		if e != nil {
			t.Fatal(e)
		}
		near(t, x.Forecast(), y.Forecast())
		value := n%3 == 0
		if _, e = a.Resolve(x, value, int64(n)); e != nil {
			t.Fatal(e)
		}
		if _, e = b.Resolve(y, value, int64(n)); e != nil {
			t.Fatal(e)
		}
		if n%7 == 0 {
			one, e := a.Query(x, "predictive")
			if e != nil {
				t.Fatal(e)
			}
			two, e := b.Query(y, "predictive")
			if e != nil {
				t.Fatal(e)
			}
			near(t, one.Observed, two.Observed)
			near(t, one.Value, two.Value)
			ax, _ := a.RequestSecond(x, int64(n))
			by, _ := b.RequestSecond(y, int64(n))
			if _, e = a.Resolve(ax, !value, int64(n)); e != nil {
				t.Fatal(e)
			}
			if _, e = b.Resolve(by, !value, int64(n)); e != nil {
				t.Fatal(e)
			}
			checks += 2
		}
		for j := range base {
			q, o, e := a.Predict(j)
			if e != nil {
				t.Fatal(e)
			}
			r, s, e := b.Predict(j)
			if e != nil {
				t.Fatal(e)
			}
			near(t, q, r)
			near(t, o, s)
			checks += 2
		}
		checks++
	}
	t.Log("legacy public law comparisons", checks)
}
func TestFutureForkAndNonlocalInfluence(t *testing.T) {
	for _, mode := range []string{"local", "noise", "shared", "hybrid"} {
		run := func(fork bool) []float64 {
			m, _ := candidate.New([]float64{.3, .7}, 1, 128, candidate.Config{Mode: mode, Hazard: 1. / 16})
			var pending candidate.Ticket
			out := []float64{}
			for n := 0; n < 48; n++ {
				if n > 0 {
					value := (n-1)%3 == 0
					if fork && n-1 >= 24 {
						value = !value
					}
					if _, e := m.Resolve(pending, value, int64(n)); e != nil {
						t.Fatal(e)
					}
				}
				x, e := m.Issue(n%2, int64(n))
				if e != nil {
					t.Fatal(e)
				}
				pending = x
				out = append(out, x.Forecast())
			}
			return out
		}
		a, b := run(false), run(true)
		changed := false
		for n := range a {
			if n <= 24 && a[n] != b[n] {
				t.Fatal("future leak", mode, n)
			}
			if n > 24 && a[n] != b[n] {
				changed = true
			}
		}
		if !changed {
			t.Fatal("vacuous fork", mode)
		}
	}
	m, _ := candidate.New([]float64{.3, .7}, 1, 128, candidate.Config{Mode: "shared", Hazard: 1. / 16})
	x, _ := m.Issue(0, 0)
	m.Resolve(x, true, 0)
	before, _, _ := m.Predict(1)
	a, _ := m.RequestSecond(x, 1)
	m.Resolve(a, false, 2)
	after, _, _ := m.Predict(1)
	if math.Abs(before-after) < 1e-7 {
		t.Fatal("vacuous nonlocal influence")
	}
}
