package researchclasssequence_test

import (
	candidate "github.com/JuanHuaXu/eventframed/internal/researchclasssequence"
	reference "github.com/JuanHuaXu/eventframed/internal/researchclasssequenceref"
	"testing"
)

func TestCheckpointCrossingAndMissingSecond(t *testing.T) {
	checks := 0
	for _, mode := range []string{"local", "noise", "shared", "hybrid"} {
		base := []float64{.25, .47, .7, .925}
		m, _ := candidate.New(base, 1, 256, candidate.Config{Mode: mode, Hazard: 1. / 16})
		r, _ := reference.New(base, mode, 1./16)
		tickets := make([]candidate.Ticket, 224)
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
				if x.second && x.n%11 == 0 {
					if e := m.Cancel(tickets[x.n], int64(at)); e != nil {
						t.Fatal(e)
					}
					verify()
					continue
				}
				measurement, value := 1, x.n%3 == 0
				if x.second {
					measurement, value = 2, x.n%5 == 0
				}
				if _, e := m.Resolve(tickets[x.n], value, int64(at)); e != nil {
					t.Fatal(mode, x, e)
				}
				if e := r.Observe(i, n, measurement, value); e != nil {
					t.Fatal(e)
				}
				verify()
				if !x.second && x.n%17 == 0 {
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
					checks += 2
					tickets[x.n], e = m.RequestSecond(tickets[x.n], int64(at))
					if e != nil {
						t.Fatal(e)
					}
					if e = r.Audit(i, n); e != nil {
						t.Fatal(e)
					}
					queue = append(queue, event{at + 180, x.n, true})
				}
			}
		}
		for n := 0; n < 224; n++ {
			at := n * 5
			deliver(at)
			x, e := m.Issue(n%4, int64(at))
			if e != nil {
				t.Fatal(e)
			}
			tickets[n] = x
			if e = r.Issue(n % 4); e != nil {
				t.Fatal(e)
			}
			delay := 10 + (n%3)*150
			queue = append(queue, event{at + delay, n, false})
			deliver(at)
			verify()
		}
		deliver(10000)
		deliver(20000)
		if m.Pending() != 0 {
			t.Fatal("drain")
		}
	}
	t.Log("cross-checkpoint independent forecast/value comparisons", checks)
}
