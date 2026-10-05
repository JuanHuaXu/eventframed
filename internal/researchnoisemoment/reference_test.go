package researchnoisemoment_test

import (
	old "github.com/JuanHuaXu/eventframed/internal/researchmoment"
	n "github.com/JuanHuaXu/eventframed/internal/researchnoisemoment"
	ref "github.com/JuanHuaXu/eventframed/internal/researchnoisemomentref"
	"math"
	"testing"
)

func near(t *testing.T, a, b float64) {
	t.Helper()
	if math.Abs(a-b) > 2e-10 || math.IsNaN(a) || math.IsNaN(b) {
		t.Fatalf("reference mismatch %.17g %.17g", a, b)
	}
}
func cfg(eta float64) n.Config {
	return n.Config{Family: "rich", Prior: "moment", Strength: 2, Hazard: 1. / 16, Shared: true, Noise: eta}
}

func TestDelayedReferenceAndZeroParity(t *testing.T) {
	base := []float64{.25, .47, .78, .925}
	for _, eta := range []float64{0, .1, .2} {
		c := cfg(eta)
		m, e := n.NewMemoV49(base, 1, 256, c)
		if e != nil {
			t.Fatal(e)
		}
		r, e := ref.New(base, c)
		if e != nil {
			t.Fatal(e)
		}
		legacy, e := old.NewMemoV49(base, 1, 256, old.Config{Family: "rich", Prior: "moment", Strength: 2, Hazard: 1. / 16, Shared: true})
		if e != nil {
			t.Fatal(e)
		}
		tickets := make([]n.Ticket, 256)
		ot := make([]old.Ticket, 256)
		for j := range tickets {
			i := j % 4
			q, e := m.Predict(i)
			if e != nil {
				t.Fatal(e)
			}
			v, e := r.Predict(i)
			if e != nil {
				t.Fatal(e)
			}
			near(t, q, v)
			tickets[j], e = m.Issue(i, int64(j))
			if e != nil {
				t.Fatal(e)
			}
			if e = r.Issue(i); e != nil {
				t.Fatal(e)
			}
			ot[j], e = legacy.Issue(i, int64(j))
			if e != nil {
				t.Fatal(e)
			}
			if eta == 0 && ot[j].Forecast() != tickets[j].Forecast() {
				t.Fatal("zero initial parity")
			}
		}
		// Permutation covers every original position once, latest to oldest rather
		// than multiplying a receipt into the most recent prediction.
		for step := 0; step < 256; step++ {
			j := (step*73 + 251) % 256
			y := j%7 < 3
			at := int64(256 + step)
			a, e := m.Resolve(tickets[j], y, at)
			if e != nil {
				t.Fatal(e)
			}
			if e = r.Resolve(j%4, j/4+1, y); e != nil {
				t.Fatal(e)
			}
			b, e := legacy.Resolve(ot[j], y, at)
			if e != nil {
				t.Fatal(e)
			}
			if eta == 0 && (a.Member != b.Member || a.Forecast != b.Forecast || a.TrialOrdinal != b.TrialOrdinal) {
				t.Fatal("zero receipt parity")
			}
			for i := range base {
				q, e := m.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				v, e := r.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				near(t, q, v)
				o, _ := m.ObservedPredict(i)
				near(t, o, eta+(1-2*eta)*v)
				if eta == 0 {
					l, _ := legacy.Predict(i)
					if q != l {
						t.Fatal("zero posterior parity")
					}
				}
			}
			aev, e := m.LogEvidence()
			if e != nil {
				t.Fatal(e)
			}
			bev, e := r.LogEvidence()
			if e != nil {
				t.Fatal(e)
			}
			near(t, aev, bev)
		}
	}
}

func TestMixtureJointEvidenceAndFutureFork(t *testing.T) {
	base := []float64{.3, .9}
	m, e := n.NewMixture(base, 1, 128, cfg(0))
	if e != nil {
		t.Fatal(e)
	}
	refs := [3]*ref.Reference{}
	for j, eta := range []float64{0, .1, .2} {
		refs[j], e = ref.New(base, cfg(eta))
		if e != nil {
			t.Fatal(e)
		}
	}
	tickets := make([]n.MixtureTicket, 24)
	for k := range tickets {
		i := k % 2
		tickets[k], e = m.Issue(i, int64(k))
		if e != nil {
			t.Fatal(e)
		}
		for _, r := range refs {
			if e = r.Issue(i); e != nil {
				t.Fatal(e)
			}
		}
	}
	distinct := false
	for step := 0; step < 24; step++ {
		k := (step*7 + 23) % 24
		y := k%5 < 2
		at := int64(24 + step)
		if step == 3 {
			if e = m.Cancel(tickets[k], at); e != nil {
				t.Fatal(e)
			}
		} else {
			if _, e = m.Resolve(tickets[k], y, at); e != nil {
				t.Fatal(e)
			}
			for _, r := range refs {
				if e = r.Resolve(k%2, k/2+1, y); e != nil {
					t.Fatal(e)
				}
			}
		}
		logs := [3]float64{}
		mx := math.Inf(-1)
		for j, r := range refs {
			v, e := r.LogEvidence()
			if e != nil {
				t.Fatal(e)
			}
			logs[j] = math.Log([]float64{.8, .1, .1}[j]) + v
			mx = math.Max(mx, logs[j])
		}
		sum := 0.
		for j := range logs {
			logs[j] = math.Exp(logs[j] - mx)
			sum += logs[j]
		}
		weights, e := m.Weights()
		if e != nil {
			t.Fatal(e)
		}
		for j := range logs {
			logs[j] /= sum
			near(t, logs[j], weights[j])
		}
		for i := range base {
			clean, observed, etaMean := 0., 0., 0.
			for j, r := range refs {
				q, _ := r.Predict(i)
				eta := []float64{0, .1, .2}[j]
				clean += logs[j] * q
				observed += logs[j] * ((1-eta)*q + eta*(1-q))
				etaMean += logs[j] * eta
			}
			q, e := m.Predict(i)
			if e != nil {
				t.Fatal(e)
			}
			near(t, q, clean)
			w, e := m.ObservedPredict(i)
			if e != nil {
				t.Fatal(e)
			}
			near(t, w, observed)
			distinct = distinct || math.Abs(observed-(etaMean+(1-2*etaMean)*clean)) > 1e-6
		}
	}
	if !distinct {
		t.Fatal("joint forecast control vacuous")
	}
	a, _ := n.NewMixture(base, 1, 128, cfg(0))
	b, _ := n.NewMixture(base, 1, 128, cfg(0))
	ta, _ := a.Issue(0, 1)
	tb, _ := b.Issue(0, 1)
	qa, _ := a.Predict(0)
	qb, _ := b.Predict(0)
	if qa != qb {
		t.Fatal("future before reveal")
	}
	a.Resolve(ta, true, 2)
	b.Resolve(tb, false, 2)
	qa, _ = a.Predict(0)
	qb, _ = b.Predict(0)
	if math.Abs(qa-qb) < .01 {
		t.Fatal("revealed fork vacuous")
	}
}

// Enumerate all 21^3 rate-state paths, independently of filtering/replay. One
// pending middle observation still consumes its state transition.
func TestExplicitLatentPaths(t *testing.T) {
	for _, eta := range []float64{0, .1, .2} {
		for _, hazard := range []float64{0, .125} {
			c := cfg(eta)
			c.Family = "narrow"
			c.Hazard = hazard
			base := []float64{.3, .8}
			m, e := n.New(base, 1, 128, c)
			if e != nil {
				t.Fatal(e)
			}
			ts := make([]n.Ticket, 3)
			for j := range ts {
				ts[j], e = m.Issue(0, int64(j))
				if e != nil {
					t.Fatal(e)
				}
			}
			m.Resolve(ts[2], false, 3)
			m.Resolve(ts[0], true, 4)
			den, num := 0., 0.
			for h, hw := range []float64{.8, .1, .1} {
				mu := base[0]
				if h == 1 {
					mu = 1 - mu
				}
				if h == 2 {
					mu = .5
				}
				p, e := ref.Prior(mu, 2, "moment")
				if e != nil {
					t.Fatal(e)
				}
				mean := 0.
				for z, v := range p {
					mean += v * (float64(z) + .5) / 21
				}
				for z0 := 0; z0 < 21; z0++ {
					a0 := (float64(z0) + .5) / 21
					for z1 := 0; z1 < 21; z1++ {
						tr1 := hazard * p[z1]
						if z0 == z1 {
							tr1 += 1 - hazard
						}
						for z2 := 0; z2 < 21; z2++ {
							tr2 := hazard * p[z2]
							if z1 == z2 {
								tr2 += 1 - hazard
							}
							a2 := (float64(z2) + .5) / 21
							mass := hw * p[z0] * tr1 * tr2 * ((1-eta)*a0 + eta*(1-a0)) * ((1-eta)*(1-a2) + eta*a2)
							den += mass
							num += mass * ((1-hazard)*a2 + hazard*mean)
						}
					}
				}
			}
			q, e := m.Predict(0)
			if e != nil {
				t.Fatal(e)
			}
			near(t, q, num/den)
			le, e := m.LogEvidence()
			if e != nil {
				t.Fatal(e)
			}
			near(t, le, math.Log(den))
		}
	}
}
