package researchtree_test

import (
	"math"
	"testing"

	prior "github.com/JuanHuaXu/eventframed/internal/researchnoisemomentref"
	tree "github.com/JuanHuaXu/eventframed/internal/researchtree"
	ref "github.com/JuanHuaXu/eventframed/internal/researchtreeref"
)

func close(t *testing.T, a, b float64) {
	t.Helper()
	if math.IsNaN(a) || math.IsNaN(b) || math.Abs(a-b) > 2e-10 {
		t.Fatalf("independent mismatch %.17g %.17g", a, b)
	}
}

type measured struct {
	member       int
	a, b, paired bool
}
type pruning struct {
	leaves      []int
	probability float64
}

func enumerate(n, depth int) []pruning {
	if depth == 0 {
		return []pruning{{[]int{n}, 1}}
	}
	out := []pruning{{[]int{n}, .5}}
	for _, l := range enumerate(2*n+1, depth-1) {
		for _, r := range enumerate(2*n+2, depth-1) {
			leaves := append(append([]int(nil), l.leaves...), r.leaves...)
			out = append(out, pruning{leaves, .5 * l.probability * r.probability})
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
func TestExplicitTreePruningsAndLatentOutcomes(t *testing.T) {
	base := []float64{.25, .47, .78, .925}
	cfg := tree.Config{Depth: 2, Window: 32, Strength: 2}
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
	if len(models) != 5 {
		t.Fatal("wrong enumerator")
	}
	z := [3]float64{}
	means := [3][4]float64{}
	for h, eta := range []float64{0, .1, .2} {
		for _, model := range models {
			mass := model.probability
			local := map[int]float64{}
			for _, n := range model.leaves {
				mu, count := 0., 0
				for i, b := range base {
					if contains(n, i) {
						mu += b
						count++
					}
				}
				mu /= float64(count)
				p, e := prior.Prior(mu, 2, "moment")
				if e != nil {
					t.Fatal(e)
				}
				evidence, moment := 0., 0.
				for atom, weight := range p {
					rate := (float64(atom) + .5) / 21
					lik := weight
					for _, x := range history {
						if !contains(n, x.member) {
							continue
						}
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
					moment += lik * rate
				}
				mass *= evidence
				if evidence > 0 {
					local[n] = moment / evidence
				}
			}
			z[h] += mass
			for i := range base {
				for _, n := range model.leaves {
					if contains(n, i) {
						means[h][i] += mass * local[n]
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
func TestDelayedSuffixIndependentReconstruction(t *testing.T) {
	base := []float64{.25, .47, .78, .925}
	comparisons, values := 0, 0
	for _, depth := range []int{0, 2, 3, 7} {
		for _, window := range []int{2, 9, 64} {
			cfg := tree.Config{Depth: depth, Window: window, Strength: 2}
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
				m.Resolve(tickets[j], false, 48)
				r.Observe(j%4, j/4+1, 1, false)
			}
			check()
		}
	}
	t.Log("independent forecast comparisons", comparisons, "predictive values", values)
}
