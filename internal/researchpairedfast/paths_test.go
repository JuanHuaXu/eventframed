package researchpairedfast_test

import (
	prior "github.com/JuanHuaXu/eventframed/internal/researchnoisemomentref"
	p "github.com/JuanHuaXu/eventframed/internal/researchpairedfast"
	"math"
	"testing"
)

// Enumerate every 21^3 state path and both outcomes of Y at the paired
// positions. The middle unobserved nomination still consumes a transition.
func TestPairedExplicitLatentPaths(t *testing.T) {
	for _, hazard := range []float64{0, .125} {
		cfg := p.Config{Strength: 2, Hazard: hazard}
		m, e := p.New([]float64{.3, .8}, 1, 128, cfg)
		if e != nil {
			t.Fatal(e)
		}
		tickets := make([]p.Ticket, 3)
		for j := range tickets {
			tickets[j], e = m.Issue(0, int64(j))
			if e != nil {
				t.Fatal(e)
			}
		}
		if _, e = m.Resolve(tickets[2], false, 3); e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(tickets[0], true, 4); e != nil {
			t.Fatal(e)
		}
		audit0, e := m.RequestAudit(tickets[0], 5)
		if e != nil {
			t.Fatal(e)
		}
		audit2, e := m.RequestAudit(tickets[2], 5)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(audit0, false, 6); e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(audit2, false, 7); e != nil {
			t.Fatal(e)
		}
		den, num, observed := 0., 0., 0.
		for h := 0; h < 28; h++ {
			mu := .3
			hw := .004
			switch h {
			case 0:
				hw = .72
			case 1:
				mu = .7
				hw = .09
			case 2:
				mu = .5
				hw = .09
			default:
				mu = .1 + .2*float64((h-3)/5)
			}
			mass, e := prior.Prior(mu, 2, "moment")
			if e != nil {
				t.Fatal(e)
			}
			mean := 0.
			for z, v := range mass {
				mean += v * (float64(z) + .5) / 21
			}
			for j, eta := range []float64{0, .1, .2} {
				ew := []float64{.8, .1, .1}[j]
				for z0 := 0; z0 < 21; z0++ {
					for z1 := 0; z1 < 21; z1++ {
						one := hazard * mass[z1]
						if z1 == z0 {
							one += 1 - hazard
						}
						for z2 := 0; z2 < 21; z2++ {
							two := hazard * mass[z2]
							if z2 == z1 {
								two += 1 - hazard
							}
							rate := (float64(z2) + .5) / 21
							// true/false disagreement at position0 has likelihood eta(1-eta).
							// false/false at position2 has p eta^2+(1-p)(1-eta)^2.
							joint := eta * (1 - eta) * (rate*eta*eta + (1-rate)*(1-eta)*(1-eta))
							weight := hw * ew * mass[z0] * one * two * joint
							next := (1-hazard)*rate + hazard*mean
							den += weight
							num += weight * next
							observed += weight * (eta + (1-2*eta)*next)
						}
					}
				}
			}
		}
		q, o, e := m.Predict(0)
		if e != nil {
			t.Fatal(e)
		}
		near(t, q, num/den)
		near(t, o, observed/den)
		if m.NoiseWeights()[0] != 0 || math.IsNaN(den) || den <= 0 {
			t.Fatal("path support")
		}
	}
}
