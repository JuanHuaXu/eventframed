package researchprior

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

// Unnormalized whole-history integration is intentionally not the learner's
// normalized cached suffix recursion. Up to64labels cannot underflow its
// bounded21-atom Bernoulli masses; products across members stay in log space.
func reference(base []float64, cfg Config, n []int, known, yes []uint64) []float64 {
	logs := make([][3]float64, len(base))
	next := make([][3]float64, len(base))
	for i, b := range base {
		mu := b
		if cfg.Center == "inverse" {
			mu = (.9*b - .05) / .8
		}
		for h, p := range []float64{mu, 1 - mu, .5} {
			var prior, mass [21]float64
			sum := 0.
			for z := 0; z < 21; z++ {
				q := (float64(z) + .5) / 21
				prior[z] = math.Exp((cfg.Strength*p-1)*math.Log(q) + (cfg.Strength*(1-p)-1)*math.Log1p(-q))
				sum += prior[z]
			}
			for z := range prior {
				prior[z] /= sum
				mass[z] = prior[z]
			}
			for j := 0; j < n[i]; j++ {
				if j > 0 {
					total := 0.
					for _, v := range mass {
						total += v
					}
					for z := range mass {
						mass[z] = (1-cfg.Hazard)*mass[z] + cfg.Hazard*prior[z]*total
					}
				}
				if known[i]&(1<<j) != 0 {
					for z := range mass {
						q := (float64(z) + .5) / 21
						if yes[i]&(1<<j) == 0 {
							q = 1 - q
						}
						mass[z] *= q
					}
				}
			}
			total, mean, initial := 0., 0., 0.
			for z, v := range mass {
				total += v
				mean += v * (float64(z) + .5) / 21
				initial += prior[z] * (float64(z) + .5) / 21
			}
			logs[i][h] = math.Log(total)
			next[i][h] = mean / total
			if n[i] > 0 {
				next[i][h] = (1-cfg.Hazard)*next[i][h] + cfg.Hazard*initial
			}
		}
	}
	out := make([]float64, len(base))
	for i := range base {
		l := logs[i]
		if cfg.Shared {
			l = [3]float64{}
			for _, v := range logs {
				for h := range l {
					l[h] += v[h]
				}
			}
		}
		weights := [3]float64{.8, .1, .1}
		maximum := math.Inf(-1)
		for h := range weights {
			weights[h] = math.Log(weights[h]) + l[h]
			maximum = math.Max(maximum, weights[h])
		}
		sum := 0.
		for h := range weights {
			weights[h] = math.Exp(weights[h] - maximum)
			sum += weights[h]
		}
		for h := range weights {
			out[i] += weights[h] / sum * next[i][h]
		}
	}
	return out
}
func near(t *testing.T, a, b float64) {
	t.Helper()
	if math.Abs(a-b) > 2e-12 {
		t.Fatalf("%.17g != %.17g", a, b)
	}
}
func TestWholeHistoryDelayedReference(t *testing.T) {
	for _, center := range []string{"raw", "inverse"} {
		for _, s := range []float64{2, 4} {
			for _, shared := range []bool{false, true} {
				for _, hazard := range []float64{0, 1. / 16} {
					cfg := Config{center, s, hazard, shared}
					base := []float64{.27, .57, .89}
					m, e := New(base, 1, 96, cfg)
					if e != nil {
						t.Fatal(e)
					}
					n := make([]int, 3)
					known, yes := make([]uint64, 3), make([]uint64, 3)
					tickets := make([]Ticket, 96)
					for k := range tickets {
						i := k % 3
						tickets[k], e = m.Issue(i, int64(k))
						if e != nil {
							t.Fatal(e)
						}
						n[i]++
					}
					rng := rand.New(rand.NewSource(720031))
					for _, k := range rng.Perm(96) {
						i, j := k%3, k/3
						y := rng.Intn(2) == 1
						receipt, e := m.Resolve(tickets[k], y, 100)
						if e != nil {
							t.Fatal(e)
						}
						if receipt.Forecast != tickets[k].Forecast() || receipt.Member != i || receipt.TrialOrdinal != j+1 {
							t.Fatal("issued identity")
						}
						known[i] |= 1 << j
						if y {
							yes[i] |= 1 << j
						}
						want := reference(base, cfg, n, known, yes)
						for i, v := range want {
							got, e := m.Predict(i)
							if e != nil {
								t.Fatal(e)
							}
							near(t, got, v)
						}
					}
					if m.Pending() != 0 {
						t.Fatal("drain")
					}
				}
			}
		}
	}
}
func TestEqualPriorAndCrossMemberBorrowing(t *testing.T) {
	base := []float64{.8, .8}
	a, _ := New(base, 1, 128, Config{"raw", 4, 0, false})
	b, _ := New(base, 1, 128, Config{"raw", 4, 0, true})
	for i := range base {
		x, _ := a.Predict(i)
		y, _ := b.Predict(i)
		near(t, x, y)
	}
	before, _ := a.Predict(1)
	for j := 0; j < 8; j++ {
		x, _ := a.Issue(0, int64(j))
		y, _ := b.Issue(0, int64(j))
		if _, e := a.Resolve(x, false, int64(j)); e != nil {
			t.Fatal(e)
		}
		if _, e := b.Resolve(y, false, int64(j)); e != nil {
			t.Fatal(e)
		}
	}
	private, _ := a.Predict(1)
	shared, _ := b.Predict(1)
	near(t, private, before)
	if shared >= before-.05 {
		t.Fatal("no informative cross-member borrowing")
	}
}
func TestOpaqueOriginalForecastAndAtomicLifecycle(t *testing.T) {
	cfg := Config{"raw", 4, 1. / 16, true}
	m, _ := New([]float64{.5, .8}, 1, 2, cfg)
	other, _ := New([]float64{.5, .8}, 1, 2, cfg)
	a, _ := m.Issue(0, 1)
	b, _ := m.Issue(1, 2)
	before := append([]trial(nil), m.trials...)
	if _, e := m.Issue(0, 3); e == nil || !reflect.DeepEqual(before, m.trials) {
		t.Fatal("pending cap mutation")
	}
	if _, e := other.Resolve(a, true, 2); e == nil {
		t.Fatal("owner")
	}
	if _, e := m.Resolve(a, true, 1); e == nil {
		t.Fatal("backward")
	}
	want := a.Forecast()
	a.q = math.NaN()
	r, e := m.Resolve(a, true, 3)
	if e != nil || r.Forecast != want {
		t.Fatal("forged scoring law", e)
	}
	if _, e = m.Resolve(a, true, 3); e == nil {
		t.Fatal("replay")
	}
	if e = m.Cancel(b, 4); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Resolve(b, false, 4); e == nil {
		t.Fatal("cancelled")
	}
	old, _ := m.Issue(0, 5)
	if e = m.BeginEpoch(2, 6); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Resolve(old, true, 6); e == nil {
		t.Fatal("stale epoch")
	}
	if e = m.BeginEpoch(2, 6); e == nil {
		t.Fatal("repeated epoch")
	}
	if m.Pending() != 0 || m.issued[0] != 0 {
		t.Fatal("epoch retains labels")
	}
}
func TestNormalizationFailureIsAtomic(t *testing.T) {
	m, _ := New([]float64{.3, .9}, 1, 128, Config{"raw", 4, .1, true})
	ticket, _ := m.Issue(0, 0)
	beforeTrials := append([]trial(nil), m.trials...)
	beforeForward := append([]row(nil), m.forward...)
	beforeEvidence := append([][3]float64(nil), m.evidence...)
	weights, global := m.weights, m.global
	m.prior[0][0][0] = math.NaN()
	// The first issue row already exists; poison that row to exercise failed
	// replay normalization without turning a source mutation into a production bug.
	m.prior[0][0][1] = math.NaN()
	if _, e := m.Resolve(ticket, true, 1); e == nil {
		t.Fatal("bad normalization accepted")
	}
	if !reflect.DeepEqual(beforeTrials, m.trials) || !reflect.DeepEqual(beforeForward, m.forward) || !reflect.DeepEqual(beforeEvidence, m.evidence) || weights != m.weights || global != m.global || m.pending != 1 || m.clock != 0 {
		t.Fatal("partial publication")
	}
}
func TestCapsAndInvalidContracts(t *testing.T) {
	base := []float64{.3, .9}
	valid := Config{"raw", 4, 1. / 16, true}
	for _, c := range []Config{{"unknown", 4, 0, true}, {"raw", math.NaN(), 0, false}, {"raw", 4, math.Inf(1), false}, {"raw", 4, -.1, false}, {"raw", 33, 0, false}} {
		if _, e := New(base, 1, 128, c); e == nil {
			t.Fatal("bad config")
		}
	}
	for _, b := range [][]float64{{.3}, {.1, .9}, {.3, math.NaN()}, make([]float64, 201)} {
		if _, e := New(b, 1, 1, valid); e == nil {
			t.Fatal("bad baseline")
		}
	}
	m, _ := New(base, 1, 128, valid)
	for j := 0; j < 64; j++ {
		if _, e := m.Issue(0, int64(j)); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := m.Issue(0, 64); e == nil {
		t.Fatal("history cap")
	}
	if _, e := m.Predict(2); e == nil {
		t.Fatal("member cap")
	}
}
func BenchmarkPriorConstructor150(b *testing.B) {
	base := make([]float64, 150)
	for i := range base {
		base[i] = .3 + .5*float64(i)/149
	}
	b.ReportAllocs()
	for j := 0; j < b.N; j++ {
		m, e := New(base, 1, 2400, Config{"raw", 4, 1. / 16, true})
		if e != nil || len(m.forward) != 9600 {
			b.Fatal(e)
		}
	}
}
func BenchmarkPriorPredict150(b *testing.B) {
	m, _ := New(makeBase(150), 1, 2400, Config{"raw", 4, 1. / 16, true})
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
func makeBase(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = .3 + .5*float64(i)/float64(n-1)
	}
	return out
}
