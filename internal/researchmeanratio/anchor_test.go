package researchmeanratio

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	ref "github.com/JuanHuaXu/eventframed/internal/researchmeanjointref"
)

// Reverse arrivals force the replay path; gaps and the 64th trial exercise the
// prospective clock without inventing evidence in an unknown or canceled row.
func TestAnchorGapsReverseArrivalsAndCanceledHoles(t *testing.T) {
	checks := 0
	for _, mode := range []string{"noise", "local", "individual"} {
		for _, hazard := range []float64{0, 1. / 16, 1} {
			t.Run(fmt.Sprintf("%s/%g", mode, hazard), func(t *testing.T) {
				base := []float64{.25, .925}
				m, err := New(base, 1, 256, Config{mode, "learn", "learn", hazard})
				if err != nil {
					t.Fatal(err)
				}
				r, err := ref.New(base, mode, "learn", "learn", hazard)
				if err != nil {
					t.Fatal(err)
				}
				near := func(a, b float64) { close(t, a, b); checks++ }
				compare := func() {
					for i := range base {
						q, o, e := m.Predict(i)
						if e != nil {
							t.Fatal(e)
						}
						u, v, e := r.Predict(i)
						if e != nil {
							t.Fatal(e)
						}
						near(q, u)
						near(o, v)
						p, e := m.Posterior(i)
						if e != nil {
							t.Fatal(e)
						}
						s, e := r.Posterior(i)
						if e != nil {
							t.Fatal(e)
						}
						for k := range p {
							near(p[k], s[k])
						}
					}
				}
				var tickets [128]Ticket
				for n := range tickets {
					tickets[n], err = m.Issue(n%2, int64(n))
					if err != nil {
						t.Fatal(err)
					}
					if err = r.Issue(n % 2); err != nil {
						t.Fatal(err)
					}
					compare()
				}
				if m.members[0].latest.anchor != -1 || m.members[1].latest.anchor != -1 {
					t.Fatal("unknown row became evidence")
				}
				before := copied(m)
				if err = m.Cancel(tickets[7], 128); err != nil {
					t.Fatal(err)
				}
				if before.odds != m.odds || before.members[1].latest != m.members[1].latest {
					t.Fatal("cancel changed belief")
				}
				compare()
				at := int64(129)
				for _, n := range []int{48, 0, 32, 20, 127, 21, 126} {
					value := n%3 != 0
					if _, err = m.Resolve(tickets[n], value, at); err != nil {
						t.Fatal(err)
					}
					if err = r.Observe(n%2, n/2+1, 1, value); err != nil {
						t.Fatal(err)
					}
					compare()
					at++
				}
				if m.members[0].latest.anchor != 63 || m.members[1].latest.anchor != 63 {
					t.Fatal("anchor lost latest reveal")
				}
				for _, n := range []int{0, 48, 127} {
					before := copied(m)
					for _, query := range []string{"forecast", "uncertainty", "information", "falsification", "predictive", "model_class", "noise_class"} {
						x, e := m.Query(tickets[n], query)
						if e != nil {
							t.Fatal(e)
						}
						y, e := r.QueryMode(n%2, n/2+1, query)
						if e != nil {
							t.Fatal(e)
						}
						for j, v := range []float64{x.Observed, x.Uncertainty, x.Information, x.EdgeCut, x.Value, x.ClassGain} {
							near(v, []float64{y.Observed, y.Uncertainty, y.Information, y.EdgeCut, y.Value, y.ClassGain}[j])
						}
					}
					if !reflect.DeepEqual(before, m) {
						t.Fatal("query publication")
					}
					a, e := m.RequestSecond(tickets[n], at)
					if e != nil {
						t.Fatal(e)
					}
					if e = r.Audit(n%2, n/2+1); e != nil {
						t.Fatal(e)
					}
					if _, e = m.Resolve(a, n%2 == 0, at); e != nil {
						t.Fatal(e)
					}
					if e = r.Observe(n%2, n/2+1, 2, n%2 == 0); e != nil {
						t.Fatal(e)
					}
					compare()
					at++
				}
			})
		}
	}
	t.Logf("%d reverse-arrival/gap/canceled-hole scalar comparisons", checks)
}

func TestAnchorIncrementalSupportAndAtomicClock(t *testing.T) {
	for _, mode := range []string{"noise", "local", "individual"} {
		m, e := New([]float64{.47, .7}, 1, 256, Config{mode, "learn", "learn", 1. / 16})
		if e != nil {
			t.Fatal(e)
		}
		a, e := m.Issue(0, 0)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(a, true, 0); e != nil {
			t.Fatal(e)
		}
		b, e := m.RequestSecond(a, 0)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(b, false, 0); e != nil {
			t.Fatal(e)
		}
		for n := 1; n < 64; n++ {
			a, e = m.Issue(0, int64(n))
			if e != nil {
				t.Fatal(e)
			}
			if _, e = m.Resolve(a, n%3 != 0, int64(n)); e != nil {
				t.Fatal(e)
			}
			for k, v := range m.members[0].latest.log {
				if m.active(k/9, (k/3)%3) && k%3 == 0 && !math.IsInf(v, -1) {
					t.Fatal("incremental update revived impossible class")
				}
			}
			if m.members[0].latest.anchor != n {
				t.Fatal("incremental anchor")
			}
		}
		m.members[1].latest.anchor = 0 // No issue at this member: reject before any publication.
		before := copied(m)
		if _, e = m.Issue(1, 64); e == nil {
			t.Fatal("invalid clock accepted")
		}
		if !reflect.DeepEqual(before, m) {
			t.Fatal("failed clock binding published state")
		}
	}
}
