package researchmoment

import (
	"math"
	"reflect"
	"testing"
)

func TestMeanProjectionAndKKT(t *testing.T) {
	for _, mu := range []float64{.025, .075, .25, .5, .925, .975} {
		for _, s := range []float64{.25, 2, 4, 32} {
			p, e := DiscretePrior(mu, s, "moment")
			if e != nil {
				t.Fatal(mu, s, e)
			}
			nu, e := DiscretePrior(mu, s, "density")
			if e != nil {
				t.Fatal(e)
			}
			sum, mean := 0., 0.
			for z, v := range p {
				sum += v
				mean += v * atom(z)
			}
			if math.Abs(mean-mu) > 2e-12 || math.Abs(sum-1) > 2e-12 {
				t.Fatal("mean contract", mu, s, mean, sum)
			}
			// The log ratio must be affine in the rate, the KL stationarity equation.
			a, b := math.Log(p[0]/nu[0]), math.Log(p[20]/nu[20])
			for z := range p {
				want := a + (b-a)*float64(z)/20
				if math.Abs(math.Log(p[z]/nu[z])-want) > 2e-10 {
					t.Fatal("nonaffine projection", mu, s, z)
				}
			}
			// A nonzero feasible perturbation keeps both total mass and first moment.
			q := p
			eps := math.Min(p[0], math.Min(p[10], p[20])) * .2
			q[0] += eps
			q[10] -= 2 * eps
			q[20] += eps
			diff, excess := 0., 0.
			for z := range q {
				diff += q[z]*math.Log(q[z]/nu[z]) - p[z]*math.Log(p[z]/nu[z])
				excess += q[z] * math.Log(q[z]/p[z])
			}
			if diff < -2e-11 || math.Abs(diff-excess) > 2e-11 {
				t.Fatal("KL identity", mu, s, diff, excess)
			}
		}
	}
	for _, mu := range []float64{.5 / 21, 20.5 / 21, 0, 1, math.NaN(), math.Inf(1)} {
		if _, e := DiscretePrior(mu, 2, "moment"); e == nil {
			t.Fatal("unsupported mean", mu)
		}
	}
}
func TestMeanInvariantButHigherMomentsChange(t *testing.T) {
	base := []float64{.25, .57, .925}
	for _, family := range []string{"narrow", "rich"} {
		a, e := New(base, 1, 192, Config{family, "moment", 2, 1. / 16, true})
		if e != nil {
			t.Fatal(e)
		}
		b, e := New(base, 1, 192, Config{family, "moment", 4, 1. / 16, true})
		if e != nil {
			t.Fatal(e)
		}
		changed := false
		for i := range base {
			x, _ := a.Predict(i)
			y, _ := b.Predict(i)
			if math.Abs(x-y) > 2e-12 {
				t.Fatal("strength changed initial mean")
			}
			for h := 0; h < a.h; h++ {
				for z := 0; z < 21; z++ {
					changed = changed || math.Abs(a.prior[(i*a.h+h)*21+z]-b.prior[(i*b.h+h)*21+z]) > 1e-5
				}
			}
		}
		if !changed {
			t.Fatal("null strength contrast")
		}
	}
	// Density control deliberately retains the V40 first-moment confound.
	a, _ := DiscretePrior(.925, 2, "density")
	b, _ := DiscretePrior(.925, 4, "density")
	x, y := 0., 0.
	for z := range a {
		x += a[z] * atom(z)
		y += b[z] * atom(z)
	}
	if math.Abs(x-y) < .04 {
		t.Fatal("density control lost confound")
	}
}
func TestLifecycleAndPrivateIssuedLaw(t *testing.T) {
	c := Config{"rich", "moment", 2, .1, true}
	m, e := New([]float64{.3, .9}, 1, 2, c)
	if e != nil {
		t.Fatal(e)
	}
	other, _ := New([]float64{.3, .9}, 1, 2, c)
	a, _ := m.Issue(0, 1)
	b, _ := m.Issue(1, 2)
	before := append([]trial(nil), m.trials...)
	if _, e = m.Issue(0, 3); e == nil || !reflect.DeepEqual(before, m.trials) {
		t.Fatal("cap mutation")
	}
	if _, e = other.Resolve(a, true, 2); e == nil {
		t.Fatal("owner")
	}
	if _, e = m.Resolve(a, true, 1); e == nil {
		t.Fatal("time")
	}
	want := a.Forecast()
	a.q = math.NaN()
	r, e := m.Resolve(a, true, 3)
	if e != nil || r.Forecast != want || r.Member != 0 || r.TrialOrdinal != 1 || r.IssuedAt != 1 || r.ArrivedAt != 3 {
		t.Fatal("original receipt", e)
	}
	if _, e = m.Resolve(a, false, 4); e == nil {
		t.Fatal("replay")
	}
	if e = m.Cancel(b, 4); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Resolve(b, true, 4); e == nil {
		t.Fatal("cancel")
	}
	old, _ := m.Issue(0, 5)
	if e = m.BeginEpoch(2, 6); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Resolve(old, true, 6); e == nil {
		t.Fatal("epoch")
	}
	if m.pending != 0 || m.issued[0] != 0 {
		t.Fatal("epoch retains labels")
	}
	if e = m.BeginEpoch(2, 6); e == nil {
		t.Fatal("repeat epoch")
	}
}
func TestFailureDoesNotPublish(t *testing.T) {
	m, _ := New([]float64{.3, .9}, 1, 128, Config{"rich", "moment", 4, .1, true})
	a, _ := m.Issue(0, 0)
	trials := append([]trial(nil), m.trials...)
	latest := append([]float64(nil), m.latest...)
	logs := append([]float64(nil), m.logs...)
	w, g := m.weights, m.global
	m.prior[0] = math.NaN()
	if _, e := m.Resolve(a, true, 1); e == nil {
		t.Fatal("invalid normalization accepted")
	}
	if !reflect.DeepEqual(trials, m.trials) || !reflect.DeepEqual(latest, m.latest) || !reflect.DeepEqual(logs, m.logs) || w != m.weights || g != m.global || m.pending != 1 || m.clock != 0 {
		t.Fatal("partial publication")
	}
}
func TestInvalidContractsAndCaps(t *testing.T) {
	valid := Config{"narrow", "density", 2, .1, true}
	for _, c := range []Config{{"unknown", "moment", 2, .1, true}, {"rich", "unknown", 2, .1, true}, {"rich", "moment", math.NaN(), .1, true}, {"rich", "moment", 2, math.NaN(), true}, {"rich", "moment", 2, 1, true}, {"rich", "moment", 33, .1, true}} {
		if _, e := New([]float64{.3, .9}, 1, 128, c); e == nil {
			t.Fatal("config")
		}
	}
	for _, base := range [][]float64{{.3}, {.1, .9}, {.3, math.NaN()}, make([]float64, 201)} {
		if _, e := New(base, 1, 1, valid); e == nil {
			t.Fatal("baseline")
		}
	}
	for _, x := range [][2]int{{0, 128}, {1, 0}, {1, 129}} {
		if _, e := New([]float64{.3, .9}, uint64(x[0]), x[1], valid); e == nil {
			t.Fatal("epoch/cap")
		}
	}
	m, _ := New([]float64{.3, .9}, 1, 128, valid)
	for j := 0; j < 64; j++ {
		if _, e := m.Issue(0, int64(j)); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := m.Issue(0, 64); e == nil {
		t.Fatal("history")
	}
	if _, e := m.Predict(2); e == nil {
		t.Fatal("member")
	}
}
func benchmarkBase() []float64 {
	base := make([]float64, 150)
	for i := range base {
		base[i] = .25 + .675*float64(i)/149
	}
	return base
}
func BenchmarkMomentConstructor150(b *testing.B) {
	base := benchmarkBase()
	b.ReportAllocs()
	for j := 0; j < b.N; j++ {
		m, e := New(base, 1, 2400, Config{"rich", "moment", 4, 1. / 16, true})
		if e != nil || m.h != 28 {
			b.Fatal(e)
		}
	}
}
func BenchmarkMomentPredict150(b *testing.B) {
	m, e := New(benchmarkBase(), 1, 2400, Config{"rich", "moment", 4, 1. / 16, true})
	if e != nil {
		b.Fatal(e)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		for i := 0; i < 150; i++ {
			if _, e := m.Predict(i); e != nil {
				b.Fatal(e)
			}
		}
	}
}

// Restore the private checkpoint outside the timed replay. All 63 later
// emissions are already known; this is not a one-label unit-suffix benchmark.
func BenchmarkMomentReplay64FullyObserved(b *testing.B) {
	m, e := New([]float64{.3, .8}, 1, 128, Config{"rich", "moment", 4, 1. / 16, true})
	if e != nil {
		b.Fatal(e)
	}
	tickets := make([]Ticket, 64)
	for j := range tickets {
		tickets[j], e = m.Issue(0, int64(j))
		if e != nil {
			b.Fatal(e)
		}
	}
	for j := 1; j < 64; j++ {
		if _, e = m.Resolve(tickets[j], j%2 == 0, 64); e != nil {
			b.Fatal(e)
		}
	}
	latest, logs := append([]float64(nil), m.latest...), append([]float64(nil), m.logs...)
	weights, global := m.weights, m.global
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		b.StopTimer()
		copy(m.latest, latest)
		copy(m.logs, logs)
		m.weights, m.global = weights, global
		m.trials[0].status = 1
		m.trials[0].useful = false
		m.pending = 1
		m.clock = 64
		b.StartTimer()
		if _, e = m.Resolve(tickets[0], true, 65); e != nil {
			b.Fatal(e)
		}
	}
}
