package researchjointsequence_test

import (
	candidate "github.com/JuanHuaXu/eventframed/internal/researchjointsequence"
	ref "github.com/JuanHuaXu/eventframed/internal/researchjointsequenceref"
	"math"
	"testing"
)

func near(t *testing.T, a, b float64) {
	t.Helper()
	if math.IsNaN(a) || math.IsNaN(b) || math.Abs(a-b) > 3e-10 {
		t.Fatalf("independent mismatch %.17g %.17g", a, b)
	}
}
func TestIndependentDelayedJointAndTower(t *testing.T) {
	checks := 0
	for _, hazard := range []float64{0, 1. / 16, 1} {
		for _, delay := range []int{0, 3, 11} {
			base := []float64{.25, .47, .7, .925}
			m, e := candidate.New(base, 1, 128, candidate.Config{Hazard: hazard})
			if e != nil {
				t.Fatal(e)
			}
			r, e := ref.New(base, hazard)
			if e != nil {
				t.Fatal(e)
			}
			tickets := make([]candidate.Ticket, 32)
			type event struct {
				at, n  int
				second bool
			}
			queue := []event{}
			issued := 0
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
						t.Fatal(e)
					}
					if e = r.Observe(i, n, measurement, value); e != nil {
						t.Fatal(e)
					}
					verify()
					if !x.second && x.n%5 == 0 {
						a, e := m.Query(tickets[x.n], "predictive")
						if e != nil {
							t.Fatal(e)
						}
						b, e := r.Query(i, n)
						if e != nil {
							t.Fatal(e)
						}
						near(t, a.Observed, b.Observed)
						near(t, a.Value, b.Value)
						for _, mode := range []string{"information", "falsification"} {
							o, e := m.Query(tickets[x.n], mode)
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
				issued++
				queue = append(queue, event{at + delay, n, false})
				deliver(at)
				verify()
			}
			deliver(10000)
			deliver(20000)
			if m.Pending() != 0 || issued != 32 {
				t.Fatal("drain")
			}
		}
	}
	t.Log("independent joint/tower scalar comparisons", checks)
}
func TestFutureForkNonvacuous(t *testing.T) {
	run := func(change bool) []float64 {
		m, _ := candidate.New([]float64{.3, .7}, 1, 64, candidate.Config{Hazard: 1. / 16})
		var pending candidate.Ticket
		out := []float64{}
		for n := 0; n < 24; n++ {
			if n > 0 {
				value := (n-1)%3 == 0
				if change && n-1 >= 12 {
					value = !value
				}
				if _, e := m.Resolve(pending, value, int64(n)); e != nil {
					t.Fatal(e)
				}
			}
			var e error
			pending, e = m.Issue(n%2, int64(n))
			if e != nil {
				t.Fatal(e)
			}
			out = append(out, pending.Forecast())
		}
		return out
	}
	a, b := run(false), run(true)
	changed := false
	for n := range a {
		if n <= 12 && a[n] != b[n] {
			t.Fatal("future data leak")
		}
		if n > 12 && a[n] != b[n] {
			changed = true
		}
	}
	if !changed {
		t.Fatal("vacuous fork")
	}
}
