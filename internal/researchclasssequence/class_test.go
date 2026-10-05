package researchclasssequence

import (
	"math"
	"reflect"
	"testing"
)

func TestClassRefinementAndBoundary(t *testing.T) {
	makeProbe := func(n int, response [4]float64) (p probe) {
		p.mass[0], p.mass[1] = .25, .25
		p.yes[0], p.yes[1] = response[0], response[1]
		for j := 0; j < n; j++ {
			p.mass[66+2*j], p.mass[67+2*j] = .25/float64(n), .25/float64(n)
			p.yes[66+2*j], p.yes[67+2*j] = response[2], response[3]
		}
		return
	}
	latent := func(p probe) float64 {
		q, g := 0., 0.
		for z, w := range p.mass {
			q += w * p.yes[z]
		}
		for z, w := range p.mass {
			y := p.yes[z]
			g += w * w * (y*y/q + (1-y)*(1-y)/(1-q) - 1)
		}
		return g
	}
	x, y := makeProbe(1, [4]float64{.9, .9, .1, .1}), makeProbe(1, [4]float64{1, 0, .5, .5})
	rx, ry := makeProbe(10, [4]float64{.9, .9, .1, .1}), makeProbe(10, [4]float64{1, 0, .5, .5})
	if !(latent(x) > latent(y) && latent(rx) < latent(ry)) {
		t.Fatal("missing representation counterexample")
	}
	for _, mode := range []string{"model_class", "noise_class"} {
		for _, p := range []probe{x, rx} {
			g, e := classGain(p, mode)
			if e != nil || math.Abs(g-.32) > 2e-12 {
				t.Fatal("class refinement", mode, g, e)
			}
		}
		for _, p := range []probe{y, ry} {
			g, e := classGain(p, mode)
			if e != nil || math.Abs(g) > 2e-12 {
				t.Fatal("nuisance reward", mode, g, e)
			}
		}
		for _, q := range []float64{0, 1e-12, .5, 1 - 1e-12, 1} {
			p := x
			for z := range p.yes {
				p.yes[z] = q
			}
			g, e := classGain(p, mode)
			if e != nil || math.Abs(g) > 2e-12 {
				t.Fatal("independence boundary", mode, q, g, e)
			}
		}
	}
	bad := x
	bad.mass[0] = math.NaN()
	if _, e := classGain(bad, "model_class"); e == nil {
		t.Fatal("invalid mass")
	}
	bad = x
	bad.yes[0] = 1.1
	if _, e := classGain(bad, "model_class"); e == nil {
		t.Fatal("invalid likelihood")
	}
}

func TestClassActualPosteriorBranches(t *testing.T) {
	checks := 0
	for _, mode := range []string{"local", "noise", "shared", "hybrid"} {
		m, e := New([]float64{.3, .5, .8}, 1, 128, Config{Mode: mode, Hazard: 1. / 16})
		if e != nil {
			t.Fatal(e)
		}
		var targets []Ticket
		for n := 0; n < 24; n++ {
			x, e := m.Issue(n%3, int64(n))
			if e != nil {
				t.Fatal(e)
			}
			if _, e = m.Resolve(x, n%4 == 0, int64(n)); e != nil {
				t.Fatal(e)
			}
			if n < 3 {
				targets = append(targets, x)
			}
		}
		classes := func(m *Model, i int, kind string) [9]float64 {
			v, e := m.currentView(-1, nil, nil)
			if e != nil {
				t.Fatal(e)
			}
			if kind == "model_class" {
				return [9]float64{v.model[0], v.model[1], v.model[2]}
			}
			lh, _, e := weights3(m.members[i].latest.log, noisePrior)
			if e != nil {
				t.Fatal(e)
			}
			sh, _, e := weights3(m.latest.log, noisePrior)
			if e != nil {
				t.Fatal(e)
			}
			var out [9]float64
			for h := 0; h < 3; h++ {
				out[h] = v.model[0] * lh[h]
				out[3+h] = v.model[1] * v.noise[h]
				out[6+h] = v.model[2] * sh[h]
			}
			return out
		}
		concentration := func(p [9]float64) float64 {
			v := 0.
			for _, w := range p {
				v += w * w
			}
			return v
		}
		for i, x := range targets {
			for _, kind := range []string{"model_class", "noise_class"} {
				before := snapshot(m)
				o, e := m.Query(x, kind)
				if e != nil {
					t.Fatal(e)
				}
				gain := -concentration(classes(m, i, kind))
				for _, value := range []bool{false, true} {
					prob := 1 - o.Observed
					if value {
						prob = o.Observed
					}
					if prob == 0 {
						continue
					}
					b := snapshot(m)
					ticket := x
					ticket.owner = &b
					audit, e := b.RequestSecond(ticket, b.clock)
					if e != nil {
						t.Fatal(e)
					}
					if _, e = b.Resolve(audit, value, b.clock); e != nil {
						t.Fatal(e)
					}
					gain += prob * concentration(classes(&b, i, kind))
				}
				if math.Abs(gain-o.ClassGain) > 2e-10 {
					t.Fatal("actual class posterior", mode, kind, gain, o.ClassGain)
				}
				if !reflect.DeepEqual(before, snapshot(m)) {
					t.Fatal("branch publication")
				}
				checks++
			}
		}
	}
	t.Log("actual posterior branch class comparisons", checks)
}
