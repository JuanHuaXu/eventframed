package researchdynvarcache

import (
	"math"
	"testing"
)

// A separate static joint summation: no candidate transitions, emission helper,
// normalized beliefs or reference-package inference are used here.
func urnPrior(b, strength, spike float64) (out [22]float64) {
	w := [21]float64{1}
	for n := 0; n < 20; n++ {
		var next [21]float64
		for z := 0; z <= n; z++ {
			next[z+1] += w[z] * (strength*b + float64(z)) / (strength + float64(n))
			next[z] += w[z] * (strength*(1-b) + float64(n-z)) / (strength + float64(n))
		}
		w = next
	}
	out[0] = spike
	for z, x := range w {
		out[z+1] = (1 - spike) * x
	}
	return
}
func latentMeasurement(p, e float64, a bool, paired bool, b bool) float64 {
	s := 0.
	for _, y := range []bool{false, true} {
		q := p
		if !y {
			q = 1 - p
		}
		v := e
		if y == a {
			v = 1 - e
		}
		q *= v
		if paired {
			v = e
			if y == b {
				v = 1 - e
			}
			q *= v
		}
		s += q
	}
	return s
}

type jointResult struct {
	family   [3]float64
	classes  [9]float64
	forecast [2]float64
	observed [2]float64
	prob     float64
}

func enumerate(mode string, base [2]float64, values [4]bool, paired bool, value bool) (out jointResult) {
	var priors [2][3][22]float64
	for i, b := range base {
		priors[i][0][0] = 1
		priors[i][1] = urnPrior(b, 1, .8)
		priors[i][2] = urnPrior(b, 2, 0)
	}
	hp := [3]float64{.8, .1, .1}
	noise := [3]float64{0, .1, .2}
	for a := 0; a < 3; a++ {
		for h0 := 0; h0 < 3; h0++ {
			for h1 := 0; h1 < 3; h1++ {
				if mode == "noise" && h0 != h1 {
					continue
				}
				w := hp[h0] / 3
				if mode == "local" {
					w *= hp[h1]
				}
				for z0, p0 := range priors[0][a] {
					for z1, p1 := range priors[1][a] {
						p := [2]float64{float64(z0-1) / 20, float64(z1-1) / 20}
						if z0 == 0 {
							p[0] = base[0]
						}
						if z1 == 0 {
							p[1] = base[1]
						}
						x := w * p0 * p1
						x *= latentMeasurement(p[0], noise[h0], values[0], paired, value)
						x *= latentMeasurement(p[0], noise[h0], values[1], false, false)
						x *= latentMeasurement(p[1], noise[h1], values[2], false, false)
						x *= latentMeasurement(p[1], noise[h1], values[3], false, false)
						out.prob += x
						out.family[a] += x
						out.classes[3*a+h0] += x
						out.forecast[0] += x * p[0]
						out.forecast[1] += x * p[1]
						out.observed[0] += x * (noise[h0] + (1-2*noise[h0])*p[0])
						out.observed[1] += x * (noise[h1] + (1-2*noise[h1])*p[1])
					}
				}
			}
		}
	}
	for a := range out.family {
		out.family[a] /= out.prob
	}
	for k := range out.classes {
		out.classes[k] /= out.prob
	}
	for i := range out.forecast {
		out.forecast[i] /= out.prob
		out.observed[i] /= out.prob
	}
	return
}
func staticHistory(t *testing.T, mode string, base [2]float64, values [4]bool) (*Model, Ticket) {
	t.Helper()
	m := create(t, base[:], Config{mode, "learn", 0})
	var tickets [4]Ticket
	// Resolve in reverse issue order to test journal position, not arrival order.
	for n := 0; n < 4; n++ {
		x, e := m.Issue(n/2, int64(n))
		if e != nil {
			t.Fatal(e)
		}
		tickets[n] = x
	}
	for n := 3; n >= 0; n-- {
		if _, e := m.Resolve(tickets[n], values[n], int64(7-n)); e != nil {
			t.Fatal(e)
		}
	}
	return m, tickets[0]
}
func concentration(x []float64) float64 {
	s := 0.
	for _, p := range x {
		s += p * p
	}
	return s
}
func TestExhaustiveJointAndActualQueryBranches(t *testing.T) {
	checks := 0
	for _, mode := range []string{"noise", "local"} {
		for _, base := range [][2]float64{{.25, .925}, {.47, .7}} {
			for mask := 0; mask < 16; mask++ {
				values := [4]bool{}
				for k := range values {
					values[k] = mask&(1<<k) != 0
				}
				m, ticket := staticHistory(t, mode, base, values)
				before := enumerate(mode, base, values, false, false)
				branches := [2]jointResult{enumerate(mode, base, values, true, false), enumerate(mode, base, values, true, true)}
				p := branches[1].prob / before.prob
				closeValue(t, "joint branches sum", (branches[0].prob+branches[1].prob)/before.prob, 1)
				w, e := m.ModelWeights()
				if e != nil {
					t.Fatal(e)
				}
				for a := range w {
					closeValue(t, "joint family", w[a], before.family[a])
					checks++
				}
				for i := 0; i < 2; i++ {
					q, o, e := m.Predict(i)
					if e != nil {
						t.Fatal(e)
					}
					closeValue(t, "joint prediction", q, before.forecast[i])
					closeValue(t, "joint measured", o, before.observed[i])
					checks += 2
				}
				for _, query := range []string{"model_class", "noise_class", "predictive"} {
					o, e := m.Query(ticket, query)
					if e != nil {
						t.Fatal(e)
					}
					closeValue(t, "joint query", o.Observed, p)
					want := 0.
					if query == "model_class" {
						want = (1-p)*concentration(branches[0].family[:]) + p*concentration(branches[1].family[:]) - concentration(before.family[:])
					}
					if query == "noise_class" {
						want = (1-p)*concentration(branches[0].classes[:]) + p*concentration(branches[1].classes[:]) - concentration(before.classes[:])
					}
					if query == "predictive" {
						for i := 0; i < 2; i++ {
							q := before.forecast[i]
							want += (q*(1-q) - (1-p)*branches[0].forecast[i]*(1-branches[0].forecast[i]) - p*branches[1].forecast[i]*(1-branches[1].forecast[i])) / 2
						}
						closeValue(t, "joint predictive utility", o.Value, want)
					} else {
						closeValue(t, "joint class utility", o.ClassGain, want)
					}
					checks += 2
				}
				for y := 0; y < 2; y++ {
					branch, first := staticHistory(t, mode, base, values)
					a, e := branch.RequestSecond(first, 10)
					if e != nil {
						t.Fatal(e)
					}
					closeValue(t, "actual branch request", a.Forecast(), p)
					if _, e = branch.Resolve(a, y == 1, 11); e != nil {
						t.Fatal(e)
					}
					w, e = branch.ModelWeights()
					if e != nil {
						t.Fatal(e)
					}
					for k := range w {
						closeValue(t, "actual family", w[k], branches[y].family[k])
						checks++
					}
					for i := 0; i < 2; i++ {
						q, o, e := branch.Predict(i)
						if e != nil {
							t.Fatal(e)
						}
						closeValue(t, "actual branch clean", q, branches[y].forecast[i])
						closeValue(t, "actual branch measured", o, branches[y].observed[i])
						checks += 2
					}
				}
				for i := 0; i < 2; i++ {
					closeValue(t, "all-target tower", before.forecast[i], (1-p)*branches[0].forecast[i]+p*branches[1].forecast[i])
					checks++
				}
			}
		}
	}
	t.Logf("%d joint/actual-branch scalar checks across 64 histories", checks)
}

func TestNonlocalDispersionInfluence(t *testing.T) {
	m := create(t, []float64{.25, .7}, Config{"local", "learn", 0})
	// Member 1 has local evidence; learning family from member 0 can then move it.
	x, e := m.Issue(1, 0)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Resolve(x, true, 0); e != nil {
		t.Fatal(e)
	}
	before, _, _ := m.Predict(1)
	for n := 0; n < 12; n++ {
		x, e = m.Issue(0, int64(n+1))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(x, true, int64(n+1)); e != nil {
			t.Fatal(e)
		}
	}
	after, _, _ := m.Predict(1)
	if math.Abs(after-before) < 1e-4 {
		t.Fatal("no hyperparameter borrowing", before, after)
	}
	// Without evidence about its own rate, the mean of an untouched member stays b.
	m = create(t, []float64{.25, .7}, Config{"local", "learn", 0})
	for n := 0; n < 12; n++ {
		x, e = m.Issue(0, int64(n))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(x, true, int64(n)); e != nil {
			t.Fatal(e)
		}
	}
	q, _, e := m.Predict(1)
	if e != nil {
		t.Fatal(e)
	}
	closeValue(t, "unobserved independent rate mean", q, .7)
}

func TestIndividualHasNoBorrowing(t *testing.T) {
	m := create(t, []float64{.25, .7}, Config{"individual", "learn", 0})
	x, e := m.Issue(1, 0)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Resolve(x, true, 0); e != nil {
		t.Fatal(e)
	}
	before, observed, e := m.Predict(1)
	if e != nil {
		t.Fatal(e)
	}
	v, e := m.currentView(-1, nil)
	if e != nil {
		t.Fatal(e)
	}
	weights, e := m.memberWeights(1, v)
	if e != nil {
		t.Fatal(e)
	}
	for n := 0; n < 12; n++ {
		x, e = m.Issue(0, int64(n+1))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(x, true, int64(n+1)); e != nil {
			t.Fatal(e)
		}
	}
	after, o, e := m.Predict(1)
	if e != nil {
		t.Fatal(e)
	}
	if before != after || observed != o {
		t.Fatal("independent member law moved")
	}
	v, e = m.currentView(-1, nil)
	if e != nil {
		t.Fatal(e)
	}
	w, e := m.memberWeights(1, v)
	if e != nil {
		t.Fatal(e)
	}
	if weights != w {
		t.Fatal("independent member hyperstate moved")
	}
}
