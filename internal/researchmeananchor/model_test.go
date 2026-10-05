package researchmeananchor

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	old "github.com/JuanHuaXu/eventframed/internal/researchdynvarcache"
	ref "github.com/JuanHuaXu/eventframed/internal/researchmeanjointref"
)

func copied(m *Model) *Model {
	x := *m
	x.base = append([]float64(nil), m.base...)
	x.members = append([]member(nil), m.members...)
	x.rows = append([]row(nil), m.rows...)
	return &x
}
func close(t *testing.T, a, b float64) {
	t.Helper()
	if !finite(a) || !finite(b) || math.Abs(a-b) > 2e-10 {
		t.Fatalf("%.17g vs %.17g", a, b)
	}
}

func TestIndependentDelayedJointReplay(t *testing.T) {
	checks := 0
	maxDefect := 0.
	near := func(t *testing.T, a, b float64) {
		t.Helper()
		close(t, a, b)
		checks++
		maxDefect = math.Max(maxDefect, math.Abs(a-b))
	}
	for _, mode := range []string{"noise", "local", "individual"} {
		for _, means := range []string{"baseline", "learn"} {
			for _, family := range []string{"baseline", "current", "free", "learn"} {
				for _, hazard := range []float64{0, 1. / 16, 1} {
					for _, delay := range []int{0, 3, 11} {
						t.Run(fmt.Sprintf("%s/%s/%s/%g/%d", mode, means, family, hazard, delay), func(t *testing.T) {
							base := []float64{.25, .47, .7, .925}
							m, e := New(base, 1, 512, Config{mode, family, means, hazard})
							if e != nil {
								t.Fatal(e)
							}
							r, e := ref.New(base, mode, family, means, hazard)
							if e != nil {
								t.Fatal(e)
							}
							check := func() {
								w, e := m.ModelWeights()
								if e != nil {
									t.Fatal(e)
								}
								v := r.ModelWeights()
								for a := range w {
									near(t, w[a], v[a])
								}
								for i := range base {
									q, o, e := m.Predict(i)
									if e != nil {
										t.Fatal(e)
									}
									u, v, e := r.Predict(i)
									if e != nil {
										t.Fatal(e)
									}
									near(t, q, u)
									near(t, o, v)
									p, e := m.Posterior(i)
									if e != nil {
										t.Fatal(e)
									}
									s, e := r.Posterior(i)
									if e != nil {
										t.Fatal(e)
									}
									for k := range p {
										near(t, p[k], s[k])
									}
								}
							}
							check()
							var tickets [16]Ticket
							observe := func(j int, at int64) {
								x, e := m.Resolve(tickets[j], j%7 < 3, at)
								if e != nil {
									t.Fatal(e)
								}
								if x.Ordinal != j/4+1 || x.Member != j%4 || x.IssuedAt != int64(j) || x.ArrivedAt != at || x.Measurement != 1 {
									t.Fatal("receipt binding")
								}
								if e = r.Observe(j%4, j/4+1, 1, j%7 < 3); e != nil {
									t.Fatal(e)
								}
								check()
							}
							for n := 0; n < 16; n++ {
								tickets[n], e = m.Issue(n%4, int64(n))
								if e != nil {
									t.Fatal(e)
								}
								if e = r.Issue(n % 4); e != nil {
									t.Fatal(e)
								}
								check()
								if n >= delay {
									observe(n-delay, int64(n))
								}
							}
							at := int64(16)
							for j := 16 - delay; j < 16; j++ {
								observe(j, at)
								at++
							}
							for _, j := range []int{0, 6} {
								before := copied(m)
								for _, query := range []string{"forecast", "uncertainty", "information", "falsification", "predictive", "model_class", "noise_class"} {
									o, e := m.Query(tickets[j], query)
									if e != nil {
										t.Fatal(e)
									}
									v, e := r.QueryMode(j%4, j/4+1, query)
									if e != nil {
										t.Fatal(e)
									}
									for k, x := range []float64{o.Observed, o.Uncertainty, o.Information, o.EdgeCut, o.Value, o.ClassGain} {
										near(t, x, []float64{v.Observed, v.Uncertainty, v.Information, v.EdgeCut, v.Value, v.ClassGain}[k])
									}
								}
								if !reflect.DeepEqual(before, m) {
									t.Fatal("query published state")
								}
								a, e := m.RequestSecond(tickets[j], at)
								if e != nil {
									t.Fatal(e)
								}
								o, e := r.QueryMode(j%4, j/4+1, "forecast")
								if e != nil {
									t.Fatal(e)
								}
								near(t, a.Forecast(), o.Observed)
								if e = r.Audit(j%4, j/4+1); e != nil {
									t.Fatal(e)
								}
								v := !(j%7 < 3)
								x, e := m.Resolve(a, v, at)
								if e != nil {
									t.Fatal(e)
								}
								if x.Forecast != a.Forecast() || x.Measurement != 2 {
									t.Fatal("paired receipt")
								}
								if e = r.Observe(j%4, j/4+1, 2, v); e != nil {
									t.Fatal(e)
								}
								check()
								at++
							}
						})
					}
				}
			}
		}
	}
	t.Logf("%d counted scalar comparisons, max defect %.17g,216 configurations", checks, maxDefect)
}

func TestFixedMeanMatchesUncachedFamily(t *testing.T) {
	checks := 0
	for _, mode := range []string{"noise", "local", "individual"} {
		for _, family := range []string{"baseline", "current", "free", "learn"} {
			for _, hazard := range []float64{0, 1. / 16, 1} {
				base := []float64{.25, .47, .7, .925}
				m, e := New(base, 1, 512, Config{mode, family, "baseline", hazard})
				if e != nil {
					t.Fatal(e)
				}
				r, e := old.New(base, 1, 512, old.Config{Mode: mode, Family: family, Hazard: hazard})
				if e != nil {
					t.Fatal(e)
				}
				for n := 0; n < 32; n++ {
					i := n % 4
					q, o, e := m.Predict(i)
					if e != nil {
						t.Fatal(e)
					}
					u, v, e := r.Predict(i)
					if e != nil {
						t.Fatal(e)
					}
					close(t, q, u)
					close(t, o, v)
					checks += 2
					a, e := m.Issue(i, int64(n))
					if e != nil {
						t.Fatal(e)
					}
					b, e := r.Issue(i, int64(n))
					if e != nil {
						t.Fatal(e)
					}
					y := n%7 < 3
					if _, e = m.Resolve(a, y, int64(n)); e != nil {
						t.Fatal(e)
					}
					if _, e = r.Resolve(b, y, int64(n)); e != nil {
						t.Fatal(e)
					}
					if n%9 == 0 {
						for _, mode := range []string{"forecast", "uncertainty", "information", "falsification", "predictive", "model_class", "noise_class"} {
							x, e := m.Query(a, mode)
							if e != nil {
								t.Fatal(e)
							}
							z, e := r.Query(b, mode)
							if e != nil {
								t.Fatal(e)
							}
							for k, x := range []float64{x.Observed, x.Uncertainty, x.Information, x.EdgeCut, x.Value, x.ClassGain} {
								close(t, x, []float64{z.Observed, z.Uncertainty, z.Information, z.EdgeCut, z.Value, z.ClassGain}[k])
								checks++
							}
						}
					}
				}
			}
		}
	}
	t.Logf("%d fixed-mean limiting comparisons; tolerance, not bitwise", checks)
}

func TestJointActualBranches(t *testing.T) {
	for _, mode := range []string{"noise", "local", "individual"} {
		for _, hazard := range []float64{0, 1. / 16, 1} {
			for mask := 0; mask < 16; mask++ {
				m, e := New([]float64{.47, .7}, 1, 256, Config{mode, "learn", "learn", hazard})
				if e != nil {
					t.Fatal(e)
				}
				var first Ticket
				for n := 0; n < 4; n++ {
					x, e := m.Issue(n%2, int64(n))
					if e != nil {
						t.Fatal(e)
					}
					if n == 0 {
						first = x
					}
					if _, e = m.Resolve(x, mask&(1<<n) != 0, int64(n)); e != nil {
						t.Fatal(e)
					}
				}
				before := copied(m)
				o, e := m.Query(first, "predictive")
				if e != nil {
					t.Fatal(e)
				}
				if !reflect.DeepEqual(before, m) {
					t.Fatal("query mutation")
				}
				var after [2][2]float64
				var posterior [2][States]float64
				for y := 0; y < 2; y++ {
					b := copied(m)
					x := first
					x.owner = b
					a, e := b.RequestSecond(x, 4)
					if e != nil {
						t.Fatal(e)
					}
					if _, e = b.Resolve(a, y == 1, 4); e != nil {
						t.Fatal(e)
					}
					for i := 0; i < 2; i++ {
						q, _, e := b.Predict(i)
						if e != nil {
							t.Fatal(e)
						}
						after[y][i] = q
					}
					posterior[y], e = b.Posterior(0)
					if e != nil {
						t.Fatal(e)
					}
				}
				value := 0.
				for i := 0; i < 2; i++ {
					q, _, e := m.Predict(i)
					if e != nil {
						t.Fatal(e)
					}
					p := o.Observed
					close(t, q, p*after[1][i]+(1-p)*after[0][i])
					value += (p*math.Pow(after[1][i]-q, 2) + (1-p)*math.Pow(after[0][i]-q, 2)) / 2
				}
				close(t, o.Value, value)
				p0, e := m.Posterior(0)
				if e != nil {
					t.Fatal(e)
				}
				for _, query := range []string{"model_class", "noise_class"} {
					x, e := m.Query(first, query)
					if e != nil {
						t.Fatal(e)
					}
					var prior [States]float64
					var branch [2][States]float64
					for k, q := range p0 {
						c := k
						if query == "model_class" {
							c = k / 3
						}
						prior[c] += q
						for y := 0; y < 2; y++ {
							branch[y][c] += posterior[y][k]
						}
					}
					gain := 0.
					for k, q := range prior {
						gain += o.Observed*branch[1][k]*branch[1][k] + (1-o.Observed)*branch[0][k]*branch[0][k] - q*q
					}
					close(t, x.ClassGain, gain)
				}
			}
		}
	}
}

func TestLifecycleFutureAndFaultBoundaries(t *testing.T) {
	cfg := Config{"local", "learn", "learn", 1. / 16}
	m, e := New([]float64{.47, .7}, 1, 256, cfg)
	if e != nil {
		t.Fatal(e)
	}
	x, e := m.Issue(0, 0)
	if e != nil {
		t.Fatal(e)
	}
	fork := copied(m)
	if _, e = m.Resolve(x, true, 1); e != nil {
		t.Fatal(e)
	}
	y := x
	y.owner = fork
	if _, e = fork.Resolve(y, false, 1); e != nil {
		t.Fatal(e)
	}
	a, _, _ := m.Predict(0)
	b, _, _ := fork.Predict(0)
	if math.Abs(a-b) < 1e-6 {
		t.Fatal("vacuous outcome fork")
	}
	before := copied(m)
	if _, e = m.Resolve(x, true, 1); e == nil {
		t.Fatal("duplicate first")
	}
	if !reflect.DeepEqual(before, m) {
		t.Fatal("failed duplicate mutation")
	}
	q, e := m.RequestSecond(x, 2)
	if e != nil {
		t.Fatal(e)
	}
	before = copied(m)
	if e = m.Cancel(q, 1); e == nil {
		t.Fatal("past cancellation")
	}
	if !reflect.DeepEqual(before, m) {
		t.Fatal("failed cancellation mutation")
	}
	if e = m.Cancel(q, 2); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Query(x, "forecast"); e == nil {
		t.Fatal("cancelled pair reusable")
	}
	if e = m.BeginEpoch(2, 3); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Query(x, "forecast"); e == nil {
		t.Fatal("old epoch query")
	}
	if _, e = m.RequestSecond(x, 3); e == nil {
		t.Fatal("old epoch request")
	}
	for i := 0; i < 2; i++ {
		q, _, e := m.Predict(i)
		if e != nil {
			t.Fatal(e)
		}
		close(t, q, []float64{.47, .7}[i])
	}
	x, e = m.Issue(0, 4)
	if e != nil {
		t.Fatal(e)
	}
	m.members[1].latest.log[0] = math.NaN()
	before = copied(m)
	if _, e = m.Resolve(x, true, 4); e == nil {
		t.Fatal("numeric source fault accepted")
	}
	if math.Float64bits(m.members[1].latest.log[0]) != math.Float64bits(before.members[1].latest.log[0]) {
		t.Fatal("fault bits changed")
	}
	// DeepEqual considers NaN unequal even to itself; compare its bits above.
	m.members[1].latest.log[0], before.members[1].latest.log[0] = 0, 0
	if !reflect.DeepEqual(before, m) {
		t.Fatal("fault partially published")
	}
	for _, cfg := range []Config{{"bad", "learn", "learn", 0}, {"noise", "bad", "learn", 0}, {"noise", "learn", "bad", 0}, {"noise", "learn", "learn", math.NaN()}, {"noise", "learn", "learn", -1}, {"noise", "learn", "learn", 2}} {
		if _, e = New([]float64{.47, .7}, 1, 256, cfg); e == nil {
			t.Fatal("invalid configuration")
		}
	}
	for _, base := range [][]float64{nil, {.5}, make([]float64, 201), {math.NaN(), .5}, {.2, .5}, {.5, .94}} {
		if _, e = New(base, 1, 1, Config{"noise", "learn", "learn", 0}); e == nil {
			t.Fatal("invalid base")
		}
	}
}

func TestIndependentMemberNoBorrowing(t *testing.T) {
	m, e := New([]float64{.47, .7}, 1, 256, Config{"individual", "learn", "learn", 1. / 16})
	if e != nil {
		t.Fatal(e)
	}
	p, e := m.Posterior(1)
	if e != nil {
		t.Fatal(e)
	}
	x, e := m.Issue(0, 0)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Resolve(x, true, 1); e != nil {
		t.Fatal(e)
	}
	q, e := m.Posterior(1)
	if e != nil {
		t.Fatal(e)
	}
	if p != q {
		t.Fatal("independent borrowed evidence")
	}
}
