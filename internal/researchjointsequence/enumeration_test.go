package researchjointsequence_test

import (
	model "github.com/JuanHuaXu/eventframed/internal/researchjointsequence"
	"testing"
)

func TestExhaustiveTwoStageJoint(t *testing.T) {
	base, hazard := .47, 1./16
	weights := [21]float64{1}
	for n := 0; n < 20; n++ {
		var next [21]float64
		for z := 0; z <= n; z++ {
			next[z+1] += weights[z] * (base + float64(z)) / (1 + float64(n))
			next[z] += weights[z] * (1 - base + float64(n-z)) / (1 + float64(n))
		}
		weights = next
	}
	prior := [22]float64{.8}
	rates := [22]float64{base}
	for z, w := range weights {
		prior[z+1] = .2 * w
		rates[z+1] = float64(z) / 20
	}
	measurement := func(p, eta float64, a, c bool, paired bool) float64 {
		sum := 0.
		for y := 0; y < 2; y++ {
			mass := p
			if y == 0 {
				mass = 1 - p
			}
			one := eta
			if (y == 1) == a {
				one = 1 - eta
			}
			mass *= one
			if paired {
				two := eta
				if (y == 1) == c {
					two = 1 - eta
				}
				mass *= two
			}
			sum += mass
		}
		return sum
	}
	checks := 0
	for flags := 0; flags < 8; flags++ {
		a, b, c := flags&1 != 0, flags&2 != 0, flags&4 != 0
		den, numBefore, numYes, denYes, numNo, denNo := 0., 0., 0., 0., 0., 0.
		for h, eta := range []float64{0, .1, .2} {
			for z, p0 := range prior {
				for v, p1 := range prior {
					transition := hazard * p1
					if z == v {
						transition += 1 - hazard
					}
					mass := []float64{.8, .1, .1}[h] * p0 * transition * measurement(rates[v], eta, b, false, false)
					clean := (1-hazard)*rates[v] + hazard*base
					first := mass * measurement(rates[z], eta, a, false, false)
					yes := mass * measurement(rates[z], eta, a, true, true)
					no := mass * measurement(rates[z], eta, a, false, true)
					den += first
					numBefore += first * clean
					denYes += yes
					numYes += yes * clean
					denNo += no
					numNo += no * clean
				}
			}
		}
		q, before := denYes/den, numBefore/den
		one, zero := numYes/denYes, numNo/denNo
		near(t, denYes+denNo, den)
		near(t, q*one+(1-q)*zero, before)
		m, e := model.New([]float64{base, .7}, 1, 16, model.Config{Hazard: hazard})
		if e != nil {
			t.Fatal(e)
		}
		x, e := m.Issue(0, 0)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(x, a, 0); e != nil {
			t.Fatal(e)
		}
		y, e := m.Issue(0, 1)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(y, b, 1); e != nil {
			t.Fatal(e)
		}
		actual, _, e := m.Predict(0)
		if e != nil {
			t.Fatal(e)
		}
		near(t, actual, before)
		o, e := m.Query(x, "predictive")
		if e != nil {
			t.Fatal(e)
		}
		near(t, o.Observed, q)
		near(t, o.Value, (q*(one-before)*(one-before)+(1-q)*(zero-before)*(zero-before))/2)
		x, e = m.RequestSecond(x, 2)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(x, c, 2); e != nil {
			t.Fatal(e)
		}
		actual, _, e = m.Predict(0)
		if e != nil {
			t.Fatal(e)
		}
		expected := zero
		if c {
			expected = one
		}
		near(t, actual, expected)
		checks += 6
	}
	t.Log("exhaustive latent path checks", checks, "paths per case", 3*22*22)
}
