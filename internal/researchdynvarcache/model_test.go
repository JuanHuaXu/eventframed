package researchdynvarcache

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	old "github.com/JuanHuaXu/eventframed/internal/researchclasssequence"
	ref "github.com/JuanHuaXu/eventframed/internal/researchdynvarianceref"
)

func closeValue(t *testing.T, label string, got, want float64) {
	t.Helper()
	if !finite(got) || !finite(want) || math.Abs(got-want) > 2e-11 {
		t.Fatalf("%s got %.17g want %.17g", label, got, want)
	}
}
func snapshot(m *Model) Model {
	x := *m
	x.base = append([]float64(nil), m.base...)
	x.priors = append([][3]vector(nil), m.priors...)
	x.factors = append([][3][RateAtoms][6]float64(nil), m.factors...)
	x.members = append([]member(nil), m.members...)
	x.rows = append([]row(nil), m.rows...)
	return x
}
func create(t *testing.T, base []float64, cfg Config) *Model {
	t.Helper()
	m, e := New(base, 1, 2*len(base)*MaxTrials, cfg)
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func checkReference(t *testing.T, m *Model, r *ref.Reference) {
	t.Helper()
	w, e := m.ModelWeights()
	if e != nil {
		t.Fatal(e)
	}
	v := r.ModelWeights()
	for a := range w {
		closeValue(t, "family", w[a], v[a])
	}
	for i := range m.base {
		q, o, e := m.Predict(i)
		if e != nil {
			t.Fatal(e)
		}
		u, v, e := r.Predict(i)
		if e != nil {
			t.Fatal(e)
		}
		closeValue(t, "clean", q, u)
		closeValue(t, "observed", o, v)
	}
}

func TestPriorMomentsAndEmissions(t *testing.T) {
	base := []float64{.25, .35, .47, .5, .7, .85, .925}
	m := create(t, base, Config{"noise", "learn", 1. / 16})
	checks := 0
	for i, b := range base {
		q, o, e := m.Predict(i)
		if e != nil {
			t.Fatal(e)
		}
		closeValue(t, "mean", q, b)
		closeValue(t, "measurement", o, .03+.94*b)
		for a := 0; a < 3; a++ {
			mean, variance, sum := 0., 0., 0.
			for z, w := range m.priors[i][a] {
				p := rate(b, z)
				sum += w
				mean += w * p
				variance += w * (p - b) * (p - b)
			}
			closeValue(t, "sum", sum, 1)
			closeValue(t, "rate mean", mean, b)
			want := 0.
			if a == 1 {
				want = .2 * b * (1 - b) * 21 / 40
			}
			if a == 2 {
				want = b * (1 - b) * 22 / 60
			}
			closeValue(t, "variance", variance, want)
			checks += 3
		}
		for h, noise := range eta {
			for z := 0; z < RateAtoms; z++ {
				p := rate(b, z)
				f := m.factors[i][h][z]
				closeValue(t, "pair sum", f[2]+f[3]+f[4]+f[5], 1)
				closeValue(t, "marginal zero", f[2]+f[3], f[0])
				closeValue(t, "marginal one", f[4]+f[5], f[1])
				// Independent latent-Y enumeration, not the emission helper.
				for a := 0; a < 2; a++ {
					for b := 0; b < 2; b++ {
						s := 0.
						for y := 0; y < 2; y++ {
							w := p
							if y == 0 {
								w = 1 - p
							}
							wa, wb := noise, noise
							if a == y {
								wa = 1 - noise
							}
							if b == y {
								wb = 1 - noise
							}
							s += w * wa * wb
						}
						closeValue(t, "same Y", f[2+2*a+b], s)
						checks++
					}
				}
			}
		}
	}
	base[0] = .9
	q, _, _ := m.Predict(0)
	closeValue(t, "copied baseline", q, .25)
	t.Logf("%d prior/emission checks", checks)
}

func TestDispersionNeedsDistinctOutcomes(t *testing.T) {
	for _, mode := range []string{"noise", "local", "individual"} {
		m := create(t, []float64{.3, .47, .7, .925}, Config{mode, "learn", 1. / 16})
		var tickets []Ticket
		for i := range m.base {
			x, e := m.Issue(i, int64(i))
			if e != nil {
				t.Fatal(e)
			}
			if _, e = m.Resolve(x, i%2 == 0, int64(i)); e != nil {
				t.Fatal(e)
			}
			tickets = append(tickets, x)
		}
		w, e := m.ModelWeights()
		if e != nil {
			t.Fatal(e)
		}
		for _, v := range w {
			closeValue(t, "one Y no dispersion", v, 1./3)
		}
		at := int64(10)
		for i, x := range tickets {
			o, e := m.Query(x, "model_class")
			if e != nil {
				t.Fatal(e)
			}
			closeValue(t, "same-Y pair cannot identify dispersion", o.ClassGain, 0)
			a, e := m.RequestSecond(x, at)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = m.Resolve(a, i%3 == 0, at); e != nil {
				t.Fatal(e)
			}
			at++
		}
		w, e = m.ModelWeights()
		if e != nil {
			t.Fatal(e)
		}
		for _, v := range w {
			closeValue(t, "paired still one Y", v, 1./3)
		}
		for n := 0; n < 8; n++ {
			x, e := m.Issue(0, at)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = m.Resolve(x, true, at); e != nil {
				t.Fatal(e)
			}
			at++
		}
		w, e = m.ModelWeights()
		if e != nil {
			t.Fatal(e)
		}
		if math.Abs(w[0]-1./3) < 1e-4 {
			t.Fatal("distinct outcomes supplied no dispersion evidence", mode, w)
		}
	}
}

func TestIndependentDenseDelayedReplay(t *testing.T) {
	comparisons := 0
	for _, mode := range []string{"noise", "local", "individual"} {
		for _, family := range []string{"learn", "baseline", "current", "free"} {
			for _, hazard := range []float64{0, 1. / 16, 1} {
				for _, delay := range []int{0, 3, 11} {
					t.Run(fmt.Sprintf("%s/%s/%g/%d", mode, family, hazard, delay), func(t *testing.T) {
						base := []float64{.25, .47, .7, .925}
						m := create(t, base, Config{mode, family, hazard})
						r, e := ref.New(base, mode, family, hazard)
						if e != nil {
							t.Fatal(e)
						}
						var tickets []Ticket
						for n := 0; n < 24; n++ {
							checkReference(t, m, r)
							comparisons += 11
							i := n % 4
							x, e := m.Issue(i, int64(n))
							if e != nil {
								t.Fatal(e)
							}
							if e = r.Issue(i); e != nil {
								t.Fatal(e)
							}
							tickets = append(tickets, x)
							if n >= delay {
								j := n - delay
								if _, e = m.Resolve(tickets[j], j%5 < 3, int64(n)); e != nil {
									t.Fatal(e)
								}
								if e = r.Observe(j%4, j/4+1, 1, j%5 < 3); e != nil {
									t.Fatal(e)
								}
							}
						}
						at := int64(24)
						for j := 24 - delay; j < 24; j++ {
							if _, e = m.Resolve(tickets[j], j%5 < 3, at); e != nil {
								t.Fatal(e)
							}
							if e = r.Observe(j%4, j/4+1, 1, j%5 < 3); e != nil {
								t.Fatal(e)
							}
							at++
						}
						checkReference(t, m, r)
						for j := 0; j < 8; j++ {
							before := snapshot(m)
							for _, query := range []string{"forecast", "uncertainty", "information", "falsification", "predictive", "model_class", "noise_class"} {
								a, e := m.Query(tickets[j], query)
								if e != nil {
									t.Fatal(e)
								}
								b, e := r.QueryMode(j%4, j/4+1, query)
								if e != nil {
									t.Fatal(e)
								}
								for k, p := range []float64{a.Observed, a.Uncertainty, a.Information, a.EdgeCut, a.Value, a.ClassGain} {
									q := []float64{b.Observed, b.Uncertainty, b.Information, b.EdgeCut, b.Value, b.ClassGain}[k]
									closeValue(t, query, p, q)
									comparisons++
								}
							}
							if !reflect.DeepEqual(before, snapshot(m)) {
								t.Fatal("query publishes")
							}
							a, e := m.RequestSecond(tickets[j], at)
							if e != nil {
								t.Fatal(e)
							}
							if e = r.Audit(j%4, j/4+1); e != nil {
								t.Fatal(e)
							}
							checkReference(t, m, r)
							if j%4 == 3 {
								if e = m.Cancel(a, at); e != nil {
									t.Fatal(e)
								}
							} else {
								receipt, e := m.Resolve(a, j%3 != 0, at)
								if e != nil {
									t.Fatal(e)
								}
								if receipt.Ordinal != j/4+1 || receipt.Measurement != 2 || receipt.ArrivedAt != at || receipt.Forecast != a.Forecast() {
									t.Fatal("receipt", receipt)
								}
								if e = r.Observe(j%4, j/4+1, 2, j%3 != 0); e != nil {
									t.Fatal(e)
								}
							}
							checkReference(t, m, r)
							comparisons += 22
							at++
						}
					})
				}
			}
		}
	}
	t.Logf("%d scalar dense-reference comparisons", comparisons)
}

func TestCurrentFamilyMatchesPreviousLaw(t *testing.T) {
	for _, mode := range []string{"noise", "local"} {
		for _, hazard := range []float64{0, 1. / 16, 1} {
			base := []float64{.25, .47, .7, .925}
			m := create(t, base, Config{mode, "current", hazard})
			r, e := old.New(base, 1, 512, old.Config{Mode: mode, Hazard: hazard})
			if e != nil {
				t.Fatal(e)
			}
			var ts []Ticket
			var os []old.Ticket
			for n := 0; n < 36; n++ {
				for i := range base {
					q, o, e := m.Predict(i)
					if e != nil {
						t.Fatal(e)
					}
					u, v, e := r.Predict(i)
					if e != nil {
						t.Fatal(e)
					}
					closeValue(t, "old clean", q, u)
					closeValue(t, "old measured", o, v)
				}
				x, e := m.Issue(n%4, int64(n))
				if e != nil {
					t.Fatal(e)
				}
				y, e := r.Issue(n%4, int64(n))
				if e != nil {
					t.Fatal(e)
				}
				ts = append(ts, x)
				os = append(os, y)
				if n >= 3 {
					j := n - 3
					if _, e = m.Resolve(ts[j], j%5 < 3, int64(n)); e != nil {
						t.Fatal(e)
					}
					if _, e = r.Resolve(os[j], j%5 < 3, int64(n)); e != nil {
						t.Fatal(e)
					}
				}
			}
			for j := 0; j < 12; j++ {
				for _, query := range []string{"forecast", "uncertainty", "information", "falsification", "predictive"} {
					a, e := m.Query(ts[j], query)
					if e != nil {
						t.Fatal(e)
					}
					b, e := r.Query(os[j], query)
					if e != nil {
						t.Fatal(e)
					}
					closeValue(t, "old query probability", a.Observed, b.Observed)
					closeValue(t, "old query information", a.Information, b.Information)
					closeValue(t, "old query edge", a.EdgeCut, b.EdgeCut)
					closeValue(t, "old query value", a.Value, b.Value)
				}
			}
		}
	}
}
