package researchclasssequence

import (
	"math"
	"testing"
)

// Enumerate full static two-member latent states, rather than replaying either
// inference implementation. Rates are the declared atoms; separate dense
// reference tests independently reconstruct the priors, rates and transitions.
func TestExhaustiveStaticMixture(t *testing.T) {
	checks := 0
	base := []float64{.3, .7}
	for _, mode := range []string{"local", "noise", "shared", "hybrid"} {
		for flags := 0; flags < 4; flags++ {
			m, e := New(base, 1, 128, Config{Mode: mode, Hazard: 0})
			if e != nil {
				t.Fatal(e)
			}
			a, b := flags&1 != 0, flags&2 != 0
			x, _ := m.Issue(0, 0)
			if _, e = m.Resolve(x, a, 0); e != nil {
				t.Fatal(e)
			}
			y, _ := m.Issue(1, 1)
			if _, e = m.Resolve(y, b, 1); e != nil {
				t.Fatal(e)
			}
			var evidence, yes, no [3]float64
			var before, one, zero [3][2]float64
			measure := func(p, e float64, a, b bool, paired bool) float64 {
				sum := 0.
				for v := 0; v < 2; v++ {
					w := p
					if v == 0 {
						w = 1 - p
					}
					s := e
					if (v == 1) == a {
						s = 1 - e
					}
					w *= s
					if paired {
						s = e
						if (v == 1) == b {
							s = 1 - e
						}
						w *= s
					}
					sum += w
				}
				return sum
			}
			add := func(model int, prior, p0, p1, e0, e1 float64) {
				first := prior * measure(p0, e0, a, false, false) * measure(p1, e1, b, false, false)
				positive := prior * measure(p0, e0, a, true, true) * measure(p1, e1, b, false, false)
				negative := prior * measure(p0, e0, a, false, true) * measure(p1, e1, b, false, false)
				evidence[model] += first
				yes[model] += positive
				no[model] += negative
				for i, p := range []float64{p0, p1} {
					before[model][i] += first * p
					one[model][i] += positive * p
					zero[model][i] += negative * p
				}
			}
			for h, e0 := range noise {
				for g, e1 := range noise {
					for z, p0 := range m.priors[0] {
						for v, p1 := range m.priors[1] {
							add(0, noisePrior[h]*noisePrior[g]*p0*p1, rate(base[0], z), rate(base[1], v), e0, e1)
						}
					}
				}
			}
			for h, eta := range noise {
				for z, p0 := range m.priors[0] {
					for v, p1 := range m.priors[1] {
						add(1, noisePrior[h]*p0*p1, rate(base[0], z), rate(base[1], v), eta, eta)
					}
				}
				for z, w := range fieldPrior {
					add(2, noisePrior[h]*w, m.fields[0][z], m.fields[1][z], eta, eta)
				}
			}
			prior := [3]float64{}
			switch mode {
			case "local":
				prior[0] = 1
			case "noise":
				prior[1] = 1
			case "shared":
				prior[2] = 1
			case "hybrid":
				prior = [3]float64{1. / 3, 1. / 3, 1. / 3}
			}
			den, denYes, denNo := 0., 0., 0.
			var mean, meanYes, meanNo [2]float64
			for h, w := range prior {
				den += w * evidence[h]
				denYes += w * yes[h]
				denNo += w * no[h]
				for i := 0; i < 2; i++ {
					mean[i] += w * before[h][i]
					meanYes[i] += w * one[h][i]
					meanNo[i] += w * zero[h][i]
				}
			}
			near := func(a, b float64) {
				if math.Abs(a-b) > 2e-12 || !finite(a) || !finite(b) {
					t.Fatal(mode, flags, a, b)
				}
				checks++
			}
			near(denYes+denNo, den)
			mw, _ := m.ModelWeights()
			for h, w := range prior {
				near(mw[h], w*evidence[h]/den)
			}
			q := denYes / den
			value := 0.
			for i := 0; i < 2; i++ {
				p, _, e := m.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				near(p, mean[i]/den)
				before, one, zero := mean[i]/den, meanYes[i]/denYes, meanNo[i]/denNo
				near(q*one+(1-q)*zero, before)
				value += (q*(one-before)*(one-before) + (1-q)*(zero-before)*(zero-before)) / 2
			}
			option, e := m.Query(x, "predictive")
			if e != nil {
				t.Fatal(e)
			}
			near(option.Observed, q)
			near(option.Value, value)
		}
	}
	t.Log("exhaustive static-mixture scalar checks", checks, "latent states per case", 3*3*22*22+3*22*22+3*16)
}
