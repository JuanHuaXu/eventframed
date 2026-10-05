package researchmeanratio

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	ref "github.com/JuanHuaXu/eventframed/internal/researchmeanjointref"
)

func TestAtAnchorBothBranchesAndUnknownSuffix(t *testing.T) {
	checks := 0
	maximum := 0.
	for _, mode := range []string{"noise", "local", "individual"} {
		for _, hazard := range []float64{0, 1. / 16, 1} {
			for _, family := range []string{"learn", "baseline", "current", "free"} {
				for y := 0; y < 2; y++ {
					t.Run(fmt.Sprintf("%s/%g/%s/%d", mode, hazard, family, y), func(t *testing.T) {
						base := []float64{.25, .925}
						m, e := New(base, 1, 256, Config{mode, family, "learn", hazard})
						if e != nil {
							t.Fatal(e)
						}
						r, e := ref.New(base, mode, family, "learn", hazard)
						if e != nil {
							t.Fatal(e)
						}
						near := func(x, y float64) { close(t, x, y); checks++; maximum = math.Max(maximum, math.Abs(x-y)) }
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
						for n := 0; n < 128; n++ {
							tickets[n], e = m.Issue(n%2, int64(n))
							if e != nil {
								t.Fatal(e)
							}
							if e = r.Issue(n % 2); e != nil {
								t.Fatal(e)
							}
							if n < 100 {
								value := n%3 != 0
								if _, e = m.Resolve(tickets[n], value, int64(n)); e != nil {
									t.Fatal(e)
								}
								if e = r.Observe(n%2, n/2+1, 1, value); e != nil {
									t.Fatal(e)
								}
							}
						}
						compare()
						before := copied(m)
						// Row98 is the last revealed row of member0. Fourteen later member
						// issues are unknown; they must not advance the original pair outcome.
						o, e := m.Query(tickets[98], "predictive")
						if e != nil {
							t.Fatal(e)
						}
						v, e := r.QueryMode(0, 50, "predictive")
						if e != nil {
							t.Fatal(e)
						}
						near(o.Observed, v.Observed)
						near(o.Value, v.Value)
						if !reflect.DeepEqual(before, m) {
							t.Fatal("hypothetical ratio published")
						}
						var after [2][2]float64
						for z := 0; z < 2; z++ {
							x := copied(m)
							ticket := tickets[98]
							ticket.owner = x
							b, e := x.RequestSecond(ticket, 128)
							if e != nil {
								t.Fatal(e)
							}
							if _, e = x.Resolve(b, z == 1, 128); e != nil {
								t.Fatal(e)
							}
							for i := range base {
								after[z][i], _, e = x.Predict(i)
								if e != nil {
									t.Fatal(e)
								}
							}
						}
						gain := 0.
						for i := range base {
							q, _, e := m.Predict(i)
							if e != nil {
								t.Fatal(e)
							}
							near(q, o.Observed*after[1][i]+(1-o.Observed)*after[0][i])
							gain += (o.Observed*math.Pow(after[1][i]-q, 2) + (1-o.Observed)*math.Pow(after[0][i]-q, 2)) / 2
						}
						near(gain, o.Value)
						a, e := m.RequestSecond(tickets[98], 128)
						if e != nil {
							t.Fatal(e)
						}
						if e = r.Audit(0, 50); e != nil {
							t.Fatal(e)
						}
						if _, e = m.Resolve(a, y == 1, 128); e != nil {
							t.Fatal(e)
						}
						if e = r.Observe(0, 50, 2, y == 1); e != nil {
							t.Fatal(e)
						}
						compare()
						if m.members[0].latest.anchor != 49 {
							t.Fatal("pair advanced anchor")
						}
						// An older pair still takes the full replay path.
						b, e := m.RequestSecond(tickets[0], 129)
						if e != nil {
							t.Fatal(e)
						}
						if e = r.Audit(0, 1); e != nil {
							t.Fatal(e)
						}
						if _, e = m.Resolve(b, y == 0, 129); e != nil {
							t.Fatal(e)
						}
						if e = r.Observe(0, 1, 2, y == 0); e != nil {
							t.Fatal(e)
						}
						compare()
					})
				}
			}
		}
	}
	t.Logf("%d anchor-branch/unknown-suffix/old-pair scalar checks, max defect %.17g,72 configurations", checks, maximum)
}

func TestAtAnchorPositiveMassOnZeroFactorFailsAtomically(t *testing.T) {
	m, e := New([]float64{.47, .7}, 1, 256, Config{"noise", "free", "learn", 1. / 16})
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
	// The rate-zero/zero-noise atom cannot emit the first observed one.
	if m.members[0].latest.free[0][0][0] != 0 {
		t.Fatal("fixture lacks zero support")
	}
	m.members[0].latest.free[0][0][0] = .001
	before := copied(m)
	if _, e = m.Resolve(b, false, 1); e == nil {
		t.Fatal("inconsistent factor support accepted")
	}
	if !reflect.DeepEqual(before, m) {
		t.Fatal("support failure published partial state")
	}
}
