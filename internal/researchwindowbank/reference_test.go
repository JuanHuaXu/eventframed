package researchwindowbank_test

import (
	ref "github.com/JuanHuaXu/eventframed/internal/researchlocalref"
	tree "github.com/JuanHuaXu/eventframed/internal/researchwindowbank"
	"math"
	"testing"
)

func close(t *testing.T, a, b float64) {
	t.Helper()
	if math.IsNaN(a) || math.IsNaN(b) || math.Abs(a-b) > 2e-10 {
		t.Fatalf("reference mismatch %.17g %.17g", a, b)
	}
}

type measured struct {
	member       int
	a, b, paired bool
}
type pruning struct {
	leaves      []int
	independent map[int]bool
	probability float64
}

func enumerate(n, depth int) []pruning {
	if depth == 0 {
		return []pruning{{leaves: []int{n}, probability: .5}, {leaves: []int{n}, independent: map[int]bool{n: true}, probability: .5}}
	}
	out := []pruning{{leaves: []int{n}, probability: .5}}
	for _, l := range enumerate(2*n+1, depth-1) {
		for _, r := range enumerate(2*n+2, depth-1) {
			leaves := append(append([]int(nil), l.leaves...), r.leaves...)
			individual := map[int]bool{}
			for k, v := range l.independent {
				individual[k] = v
			}
			for k, v := range r.independent {
				individual[k] = v
			}
			out = append(out, pruning{leaves: leaves, independent: individual, probability: .5 * l.probability * r.probability})
		}
	}
	return out
}
func contains(n, member int) bool {
	for p := 3 + member; ; p = (p - 1) / 2 {
		if p == n {
			return true
		}
		if p == 0 {
			return false
		}
	}
}
func conditionalRates(b float64) [21]float64 {
	var rates [21]float64
	rates[0] = b
	lo, hi := -40., 40.
	offset := func(z int) float64 {
		if z <= 10 {
			return -.6 * float64(z)
		}
		return .6 * float64(z-10)
	}
	for k := 0; k < 120; k++ {
		x := (lo + hi) / 2
		mean := 0.
		for j := 1; j <= 20; j++ {
			mean += 1 / (20 * (1 + math.Exp(-x-offset(j))))
		}
		if mean < b {
			lo = x
		} else {
			hi = x
		}
	}
	for j := 1; j <= 20; j++ {
		rates[j] = 1 / (1 + math.Exp(-(lo+hi)/2-offset(j)))
	}
	return rates
}
func TestExplicitAllPruningsAndLatentY(t *testing.T) {
	base := []float64{.25, .47, .78, .925}
	cfg := tree.Config{Depth: 2, Window: 32}
	m, e := tree.New(base, 1, 128, cfg)
	if e != nil {
		t.Fatal(e)
	}
	history := []measured{{0, true, false, true}, {1, false, false, false}, {2, true, true, true}, {3, false, true, true}, {0, false, false, false}, {2, true, false, false}}
	for k, x := range history {
		a, _ := m.Issue(x.member, int64(k))
		if _, e = m.Resolve(a, x.a, int64(k)); e != nil {
			t.Fatal(e)
		}
		if x.paired {
			b, e := m.RequestAudit(a, int64(k))
			if e != nil {
				t.Fatal(e)
			}
			if _, e = m.Resolve(b, x.b, int64(k)); e != nil {
				t.Fatal(e)
			}
		}
	}
	models := enumerate(0, 2)
	if len(models) != 26 {
		t.Fatal("enum count")
	}
	z := [3]float64{}
	means := [3][4]float64{}
	rates := make([][21]float64, 4)
	for i, b := range base {
		rates[i] = conditionalRates(b)
	}
	for h, eta := range []float64{0, .1, .2} {
		for _, model := range models {
			mass := model.probability
			local := map[int][4]float64{}
			for _, n := range model.leaves {
				evidence := 0.
				moments := [4]float64{}
				if model.independent[n] {
					evidence = 1
					for i, b := range base {
						if contains(n, i) {
							rates, priors := independentPrior(b)
							one, moment := 0., 0.
							for z, p := range rates {
								lik := priors[z]
								for _, x := range history {
									if x.member == i {
										lik *= measurement(p, eta, x)
									}
								}
								one += lik
								moment += lik * p
							}
							evidence *= one
							if one > 0 {
								moments[i] = moment / one
							}
						}
					}
					mass *= evidence
					local[n] = moments
					continue
				}
				for atom := 0; atom < 21; atom++ {
					weight := .01
					if atom == 0 {
						weight = .8
					}
					lik := weight
					for _, x := range history {
						if !contains(n, x.member) {
							continue
						}
						rate := rates[x.member][atom]
						sum := 0.
						for latent := 0; latent < 2; latent++ {
							one := 1 - rate
							if latent == 1 {
								one = rate
							}
							if x.a == (latent == 1) {
								one *= 1 - eta
							} else {
								one *= eta
							}
							if x.paired {
								if x.b == (latent == 1) {
									one *= 1 - eta
								} else {
									one *= eta
								}
							}
							sum += one
						}
						lik *= sum
					}
					evidence += lik
					for i := range base {
						moments[i] += lik * rates[i][atom]
					}
				}
				mass *= evidence
				if evidence > 0 {
					for i := range base {
						moments[i] /= evidence
					}
				}
				local[n] = moments
			}
			z[h] += mass
			for i := range base {
				for _, n := range model.leaves {
					if contains(n, i) {
						means[h][i] += mass * local[n][i]
					}
				}
			}
		}
	}
	den := .8*z[0] + .1*z[1] + .1*z[2]
	w := m.NoiseWeights()
	for h, pp := range []float64{.8, .1, .1} {
		close(t, w[h], pp*z[h]/den)
	}
	for i := range base {
		q, o, _ := m.Predict(i)
		expected, observed := 0., 0.
		for h, eta := range []float64{0, .1, .2} {
			if z[h] > 0 {
				p := means[h][i] / z[h]
				expected += w[h] * p
				observed += w[h] * (eta + (1-2*eta)*p)
			}
		}
		close(t, q, expected)
		close(t, o, observed)
	}
}

// Closed gamma-function prior is independent of both candidate recurrence and
// reference Polya dynamic programming; only used in small exhaustive tests.
func independentPrior(b float64) ([22]float64, [22]float64) {
	var rates, priors [22]float64
	rates[0], priors[0] = b, .8
	lg := func(x float64) float64 { v, _ := math.Lgamma(x); return v }
	for z := 0; z <= 20; z++ {
		rates[z+1] = float64(z) / 20
		logMass := lg(float64(z)+b) + lg(21-float64(z)-b) - lg(float64(z)+1) - lg(21-float64(z)) - lg(b) - lg(1-b)
		priors[z+1] = .2 * math.Exp(logMass)
	}
	return rates, priors
}
func measurement(p, eta float64, x measured) float64 {
	sum := 0.
	for y := 0; y < 2; y++ {
		v := 1 - p
		if y == 1 {
			v = p
		}
		if x.a == (y == 1) {
			v *= 1 - eta
		} else {
			v *= eta
		}
		if x.paired {
			if x.b == (y == 1) {
				v *= 1 - eta
			} else {
				v *= eta
			}
		}
		sum += v
	}
	return sum
}
func TestExplicitTerminalJointAndMemberDisagreement(t *testing.T) {
	for _, count := range []int{0, 1, 4, 20} {
		base := []float64{.925, .925}
		m, e := tree.New(base, 1, 128, tree.Config{Depth: 0, Window: 128})
		if e != nil {
			t.Fatal(e)
		}
		history := []measured{}
		for k := 0; k < count; k++ {
			for i := 0; i < 2; i++ {
				x := measured{member: i, a: i == 1, b: i == 1, paired: true}
				history = append(history, x)
				ticket, e := m.Issue(i, int64(k))
				if e != nil {
					t.Fatal(e)
				}
				if _, e = m.Resolve(ticket, x.a, int64(k)); e != nil {
					t.Fatal(e)
				}
				second, e := m.RequestAudit(ticket, int64(k))
				if e != nil {
					t.Fatal(e)
				}
				if _, e = m.Resolve(second, x.b, int64(k)); e != nil {
					t.Fatal(e)
				}
			}
		}
		den := 0.
		moments := [2]float64{}
		weights := [3]float64{}
		states := 0
		rates, prior := independentPrior(.925)
		pooled := conditionalRates(.925)
		for h, eta := range []float64{0, .1, .2} {
			noiseMass := []float64{.8, .1, .1}[h]
			accumulate := func(p1, p2, mass float64) {
				states++
				for _, x := range history {
					p := p1
					if x.member == 1 {
						p = p2
					}
					mass *= measurement(p, eta, x)
				}
				den += mass
				weights[h] += mass
				moments[0] += mass * p1
				moments[1] += mass * p2
			}
			for z, p := range pooled {
				w := .01
				if z == 0 {
					w = .8
				}
				accumulate(p, p, .5*noiseMass*w)
			}
			for z, a := range rates {
				for k, b := range rates {
					accumulate(a, b, .5*noiseMass*prior[z]*prior[k])
				}
			}
		}
		if states != 1515 {
			t.Fatal("latent enumeration count", states)
		}
		for i := 0; i < 2; i++ {
			q, _, e := m.Predict(i)
			if e != nil {
				t.Fatal(e)
			}
			close(t, q, moments[i]/den)
		}
		for h, w := range m.NoiseWeights() {
			close(t, w, weights[h]/den)
		}
		if count == 4 {
			q, _, _ := m.Predict(0)
			r, _, _ := m.Predict(1)
			close(t, q, .18461235673454438)
			close(t, r, .9307465783613622)
		}
		if count == 20 {
			q, _, _ := m.Predict(0)
			r, _, _ := m.Predict(1)
			if q >= .25 || r <= .75 {
				t.Fatal("member-local escape failed", q, r)
			}
		}
	}
}
func TestDelayedSuffixIndependentRebuild(t *testing.T) {
	base := []float64{.25, .47, .78, .925}
	comparisons, values := 0, 0
	for _, depth := range []int{0, 2, 3, 7} {
		for _, window := range []int{2, 9, 64} {
			cfg := tree.Config{Depth: depth, Window: window}
			m, e := tree.New(base, 1, 128, cfg)
			if e != nil {
				t.Fatal(e)
			}
			r, e := ref.New(base, cfg)
			if e != nil {
				t.Fatal(e)
			}
			tickets := make([]tree.Ticket, 48)
			check := func() {
				t.Helper()
				if e := r.RebuildAll(); e != nil {
					t.Fatal(e)
				}
				rw, _ := r.NoiseWeights()
				for h, w := range m.NoiseWeights() {
					close(t, w, rw[h])
				}
				for i := range base {
					q, o, e := m.Predict(i)
					if e != nil {
						t.Fatal(e)
					}
					a, b, e := r.Predict(i)
					if e != nil {
						t.Fatal(e)
					}
					close(t, q, a)
					close(t, o, b)
					comparisons++
				}
			}
			for k := range tickets {
				check()
				tickets[k], e = m.Issue(k%4, int64(k))
				if e != nil {
					t.Fatal(e)
				}
				if e = r.Issue(k % 4); e != nil {
					t.Fatal(e)
				}
				if k >= 3 {
					j := k - 3
					y := j%5 < 2
					if _, e = m.Resolve(tickets[j], y, int64(k)); e != nil {
						t.Fatal(e)
					}
					if e = r.Observe(j%4, j/4+1, 1, y); e != nil {
						t.Fatal(e)
					}
					if window > 3 && j%3 == 0 {
						before := m.NoiseWeights()
						for _, mode := range []string{"forecast", "uncertainty", "information", "falsification"} {
							o, e := m.Query(tickets[j], mode)
							if e != nil {
								t.Fatal(e)
							}
							gold, e := r.Options(j%4, j/4+1)
							if e != nil {
								t.Fatal(e)
							}
							close(t, o.Observed, gold.Observed)
							switch mode {
							case "uncertainty":
								close(t, o.Uncertainty, gold.Uncertainty)
							case "information":
								close(t, o.Information, gold.Information)
							case "falsification":
								close(t, o.EdgeCut, gold.EdgeCut)
							}
						}
						v, e := m.PredictionValues([]tree.Ticket{tickets[j]}, []float64{.1, .2, .3, .4})
						if e != nil {
							t.Fatal(e)
						}
						q, gold, e := r.PredictionValue(j%4, j/4+1, []float64{.1, .2, .3, .4})
						if e != nil {
							t.Fatal(e)
						}
						close(t, v[0].Observed, q)
						close(t, v[0].Value, gold)
						values++
						if before != m.NoiseWeights() {
							t.Fatal("query mutation")
						}
						s, e := m.RequestAudit(tickets[j], int64(k))
						if e != nil {
							t.Fatal(e)
						}
						if e = r.Audit(j%4, j/4+1); e != nil {
							t.Fatal(e)
						}
						if _, e = m.Resolve(s, !y, int64(k)); e != nil {
							t.Fatal(e)
						}
						if e = r.Observe(j%4, j/4+1, 2, !y); e != nil {
							t.Fatal(e)
						}
					}
				}
				check()
			}
			for j := 45; j < 48; j++ {
				if _, e = m.Resolve(tickets[j], false, 48); e != nil {
					t.Fatal(e)
				}
				if e = r.Observe(j%4, j/4+1, 1, false); e != nil {
					t.Fatal(e)
				}
			}
			check()
		}
	}
	t.Log("independent forecasts", comparisons, "prediction values", values)
}
