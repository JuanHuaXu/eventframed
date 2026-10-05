package observationlearners

import (
	"fmt"
	"math"
	"math/bits"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type spikeFactors struct{ logOdds, mean, variance []float64 }

func spikeLogOdds(p float64) float64 { return math.Log(p) - math.Log1p(-p) }

// Retain both tails without subtracting a rounded probability from one.
func spikeBernoulli(odds float64) (g, h, logG, logH float64) {
	a := math.Abs(odds)
	u := math.Exp(-a)
	small, large := u/(1+u), 1/(1+u)
	ls, ll := -a-math.Log1p(u), -math.Log1p(u)
	if odds >= 0 {
		return large, small, ll, ls
	}
	return small, large, ls, ll
}

func (q spikeFactors) probability(j int) float64 {
	g, _, _, _ := spikeBernoulli(q.logOdds[j])
	return g
}

type spikeProfile struct {
	x                  [][]float64
	k, xi, w, t, b     []float64
	gram               [][]float64
	v0, ksum, constant float64
}

func spikeSetup(samples []observation.Sample, masks []uint16, xi []float64) (*spikeProfile, error) {
	if len(samples) < 1 || len(samples) > 64 || len(masks) < 1 || len(masks) > 255 || len(xi) != len(samples) {
		return nil, fmt.Errorf("spike dimensions")
	}
	seen := map[uint16]bool{}
	for _, m := range masks {
		if m == 0 || m >= 512 || seen[m] {
			return nil, fmt.Errorf("spike masks")
		}
		seen[m] = true
	}
	p := &spikeProfile{x: make([][]float64, len(samples)), k: make([]float64, len(samples)), xi: append([]float64(nil), xi...), w: make([]float64, len(samples)), t: make([]float64, len(masks)), b: make([]float64, len(masks)), gram: make([][]float64, len(masks))}
	sw := 0.
	for i, s := range samples {
		if s.Bits >= 512 || xi[i] < 0 || !spikeFinite(xi[i]) {
			return nil, fmt.Errorf("spike sample")
		}
		p.x[i] = make([]float64, len(masks))
		p.k[i] = -.5
		if s.Outcome {
			p.k[i] = .5
		}
		p.ksum += p.k[i]
		p.w[i] = 2 * variationalLambda(xi[i])
		sw += p.w[i]
		p.constant += -ridgeNLL(xi[i], true) - xi[i]/2 + variationalLambda(xi[i])*xi[i]*xi[i]
		for j, m := range masks {
			p.x[i][j] = sparseVRVMPhi(s.Bits, m)
			p.t[j] += p.w[i] * p.x[i][j]
			p.b[j] += p.x[i][j] * p.k[i]
		}
	}
	p.v0 = 1 / (1 + sw)
	p.constant += (math.Log(p.v0) + p.v0*p.ksum*p.ksum) / 2
	for j := range masks {
		p.b[j] -= p.v0 * p.t[j] * p.ksum
		p.gram[j] = make([]float64, len(masks))
		for h := range masks {
			p.gram[j][h] = -p.v0 * p.t[j] * p.t[h]
			for i := range samples {
				p.gram[j][h] += p.w[i] * p.x[i][j] * p.x[i][h]
			}
		}
	}
	if !spikeFinite(p.constant) {
		return nil, fmt.Errorf("spike profile overflow")
	}
	return p, nil
}

func spikeFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func (q spikeFactors) check(p int, pi, c2 float64) error {
	if len(q.mean) != p || len(q.logOdds) != p || len(q.variance) != p || !(pi > 0 && pi < 1) || !(c2 > 0) || !spikeFinite(c2) {
		return fmt.Errorf("spike factor dimensions or prior")
	}
	for j := 0; j < p; j++ {
		if !spikeFinite(q.logOdds[j]) || !(q.variance[j] > 0) || !spikeFinite(q.variance[j]) || !spikeFinite(q.mean[j]) {
			return fmt.Errorf("spike factor value")
		}
	}
	return nil
}
func (q spikeFactors) clone() spikeFactors {
	return spikeFactors{append([]float64(nil), q.logOdds...), append([]float64(nil), q.mean...), append([]float64(nil), q.variance...)}
}
func (q spikeFactors) moments(j int) (float64, float64) {
	g, h, _, _ := spikeBernoulli(q.logOdds[j])
	m := q.mean[j]
	return g * m, g*q.variance[j] + g*h*m*m
}
func (q spikeFactors) kl(pi, c2 float64) float64 {
	v := 0.
	for j, odds := range q.logOdds {
		g, h, logG, logH := spikeBernoulli(odds)
		v += g*(logG-math.Log(pi)) + h*(logH-math.Log1p(-pi))
		v += g / 2 * ((q.variance[j]+q.mean[j]*q.mean[j])/c2 - 1 + math.Log(c2/q.variance[j]))
	}
	return v
}
func (p *spikeProfile) bound(q spikeFactors, pi, c2 float64) (float64, error) {
	if err := q.check(len(p.t), pi, c2); err != nil {
		return 0, err
	}
	v := p.constant - q.kl(pi, c2)
	for j := range p.t {
		e, vj := q.moments(j)
		v += p.b[j]*e - p.gram[j][j]*vj/2
		for h := range p.t {
			eh, _ := q.moments(h)
			v -= p.gram[j][h] * e * eh / 2
		}
	}
	if !spikeFinite(v) {
		return 0, fmt.Errorf("spike bound overflow")
	}
	return v, nil
}
func (p *spikeProfile) moments(q spikeFactors, x []float64) (float64, float64) {
	mu, v := p.v0*p.ksum, p.v0
	for j := range p.t {
		e, vj := q.moments(j)
		a := x[j] - p.v0*p.t[j]
		mu += a * e
		v += a * a * vj
	}
	return mu, v
}

// Compare with the expected likelihood and conditional-intercept KL rather
// than reusing the collapsed Gram expression.
func (p *spikeProfile) uncollapsed(q spikeFactors, pi, c2 float64) float64 {
	mean0, varConditionalMean := p.v0*p.ksum, 0.
	for j := range p.t {
		e, v := q.moments(j)
		mean0 -= p.v0 * p.t[j] * e
		varConditionalMean += p.v0 * p.v0 * p.t[j] * p.t[j] * v
	}
	v := -q.kl(pi, c2) - (p.v0+mean0*mean0+varConditionalMean-1-math.Log(p.v0))/2
	for i := range p.x {
		mu, variance := mean0, p.v0+varConditionalMean
		for j, x := range p.x[i] {
			e, vj := q.moments(j)
			mu += x * e
			variance += x*x*vj - 2*x*p.v0*p.t[j]*vj
		}
		v += -ridgeNLL(p.xi[i], true) - p.xi[i]/2 + variationalLambda(p.xi[i])*(p.xi[i]*p.xi[i]-mu*mu-variance) + p.k[i]*mu
	}
	return v
}
func (p *spikeProfile) coordinate(q spikeFactors, j int, pi, c2 float64) (spikeFactors, error) {
	if err := q.check(len(p.t), pi, c2); err != nil {
		return q, err
	}
	if j < 0 || j >= len(p.t) {
		return q, fmt.Errorf("spike coordinate")
	}
	out := q.clone()
	rhs := p.b[j]
	for h := range p.t {
		if h != j {
			e, _ := q.moments(h)
			rhs -= p.gram[j][h] * e
		}
	}
	out.variance[j] = 1 / (1/c2 + p.gram[j][j])
	out.mean[j] = out.variance[j] * rhs
	out.logOdds[j] = spikeLogOdds(pi) + (math.Log(out.variance[j]/c2)+out.mean[j]*out.mean[j]/out.variance[j])/2
	if err := out.check(len(p.t), pi, c2); err != nil {
		return q, err
	}
	return out, nil
}

func TestSpikeSlabComponent(t *testing.T) {
	for _, n := range []int{1, 7, 64} {
		for _, masks := range [][]uint16{{1}, {1, 2, 3}, {1, 2, 3, 7, 15, 31, 63, 127, 255}} {
			s := make([]observation.Sample, n)
			xi := make([]float64, n)
			for i := range s {
				s[i] = observation.Sample{Bits: uint16(i * 37 % 512), Outcome: i%3 == 0}
				xi[i] = float64(i % 7)
			}
			original := append([]observation.Sample(nil), s...)
			oldXi := append([]float64(nil), xi...)
			p, err := spikeSetup(s, masks, xi)
			if err != nil {
				t.Fatal(err)
			}
			q := spikeFactors{make([]float64, len(masks)), make([]float64, len(masks)), make([]float64, len(masks))}
			for j := range masks {
				q.logOdds[j] = spikeLogOdds(.1 + float64(j%3)*.2)
				q.mean[j] = float64(j%5) - 2
				q.variance[j] = .2 + float64(j%4)*.1
			}
			for _, pi := range []float64{.01, .2, .8} {
				for _, c2 := range []float64{.25, 1, 4} {
					state := q.clone()
					before, err := p.bound(state, pi, c2)
					if err != nil {
						t.Fatal(err)
					}
					if math.Abs(before-p.uncollapsed(state, pi, c2)) > 1e-9 {
						t.Fatal("collapsed bound mismatch")
					}
					for j := range masks {
						old := state.clone()
						next, err := p.coordinate(state, j, pi, c2)
						if err != nil {
							t.Fatal(err)
						}
						if !reflect.DeepEqual(old, state) {
							t.Fatal("mutated factor input")
						}
						after, err := p.bound(next, pi, c2)
						if err != nil || after < before-1e-9 {
							t.Fatal("coordinate decrease", before, after, err)
						}
						if math.Abs(after-p.uncollapsed(next, pi, c2)) > 1e-9 {
							t.Fatal("updated collapsed mismatch")
						}
						for _, sign := range []float64{-1, 1} {
							for _, field := range []int{0, 1, 2} {
								probe := next.clone()
								switch field {
								case 0:
									probe.logOdds[j] += sign * .01
								case 1:
									probe.mean[j] += sign * .01
								case 2:
									probe.variance[j] *= math.Exp(sign * .01)
								}
								v, err := p.bound(probe, pi, c2)
								if err != nil || v > after+1e-9 {
									t.Fatal("coordinate not local maximum", err)
								}
							}
						}
						state = next
						before = after
					}
					newXi := make([]float64, n)
					for i := range s {
						mu, v := p.moments(state, p.x[i])
						newXi[i] = math.Sqrt(mu*mu + v)
					}
					nextProfile, err := spikeSetup(s, masks, newXi)
					if err != nil {
						t.Fatal(err)
					}
					after, err := nextProfile.bound(state, pi, c2)
					if err != nil || after < before-1e-9 {
						t.Fatal("xi/reprofile decrease", err)
					}
				}
			}
			if !reflect.DeepEqual(original, s) || !reflect.DeepEqual(oldXi, xi) {
				t.Fatal("mutated evidence")
			}
		}
	}
}

func TestSpikeSlabOneFeature(t *testing.T) {
	s := []observation.Sample{{Bits: 0, Outcome: false}, {Bits: 1, Outcome: true}, {Bits: 1, Outcome: false}}
	p, err := spikeSetup(s, []uint16{1}, []float64{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	const pi, c2 = .2, 1.
	q, err := p.coordinate(spikeFactors{[]float64{spikeLogOdds(.3)}, []float64{.4}, []float64{.6}}, 0, pi, c2)
	if err != nil {
		t.Fatal(err)
	}
	z, m, v := 0., 0., 0.
	const panels = 131072
	dx := 20. / panels
	for i := 0; i < panels; i++ {
		x := -10 + (float64(i)+.5)*dx
		weight := math.Exp(p.b[0]*x-.5*p.gram[0][0]*x*x-x*x/(2*c2)) * dx / math.Sqrt(2*math.Pi*c2)
		z += weight
		m += weight * x
		v += weight * x * x
	}
	if math.Abs(q.probability(0)-pi*z/(1-pi+pi*z)) > 1e-10 || math.Abs(q.mean[0]-m/z) > 1e-10 || math.Abs(q.variance[0]-(v/z-m*m/(z*z))) > 1e-10 {
		t.Fatal("surrogate quadrature mismatch")
	}
	for i := range s {
		s[i].Outcome = !s[i].Outcome
	}
	r, err := spikeSetup(s, []uint16{1}, []float64{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	other, err := r.coordinate(spikeFactors{[]float64{spikeLogOdds(.3)}, []float64{-.4}, []float64{.6}}, 0, pi, c2)
	if err != nil {
		t.Fatal(err)
	}
	if q.logOdds[0] != other.logOdds[0] || q.mean[0] != -other.mean[0] || q.variance[0] != other.variance[0] {
		t.Fatal("complement symmetry")
	}
}

func TestSpikeSlabInvalid(t *testing.T) {
	s := []observation.Sample{{Bits: 0}}
	for _, masks := range [][]uint16{nil, {0}, {512}, {1, 1}} {
		if _, err := spikeSetup(s, masks, []float64{1}); err == nil {
			t.Fatal("bad masks")
		}
	}
	for _, xi := range [][]float64{nil, {-1}, {math.NaN()}, {math.Inf(1)}} {
		if _, err := spikeSetup(s, []uint16{1}, xi); err == nil {
			t.Fatal("bad xi")
		}
	}
	p, _ := spikeSetup(s, []uint16{1}, []float64{1})
	q := spikeFactors{[]float64{spikeLogOdds(.2)}, []float64{0}, []float64{1}}
	for _, pi := range []float64{0, 1, math.NaN()} {
		if _, err := p.bound(q, pi, 1); err == nil {
			t.Fatal("bad prior")
		}
	}
	for _, c2 := range []float64{0, -1, math.Inf(1), math.NaN()} {
		if _, err := p.bound(q, .2, c2); err == nil {
			t.Fatal("bad slab")
		}
	}
	bad := q.clone()
	bad.variance[0] = math.Inf(1)
	if _, err := p.coordinate(bad, 0, .2, 1); err == nil {
		t.Fatal("bad variance")
	}
}

func spikeLargeFixture() ([]observation.Sample, []uint16, []float64, spikeFactors) {
	s := make([]observation.Sample, 64)
	xi := make([]float64, len(s))
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 7), Outcome: i%3 == 0}
		xi[i] = 1
	}
	var masks []uint16
	for m := uint16(1); m < 512; m++ {
		if bits.OnesCount16(m) <= 4 {
			masks = append(masks, m)
		}
	}
	q := spikeFactors{make([]float64, len(masks)), make([]float64, len(masks)), make([]float64, len(masks))}
	for j := range masks {
		q.logOdds[j], q.variance[j] = spikeLogOdds(.05), .25
		q.mean[j] = .001 * float64(j%3-1)
	}
	return s, masks, xi, q
}

func TestSpikeSlabFullBasis(t *testing.T) {
	s, masks, xi, q := spikeLargeFixture()
	if len(masks) != 255 {
		t.Fatal("basis size")
	}
	p, err := spikeSetup(s, masks, xi)
	if err != nil {
		t.Fatal(err)
	}
	before, err := p.bound(q, .1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(before-p.uncollapsed(q, .1, 1)) > 1e-9 {
		t.Fatal("full initial bound")
	}
	for j := range masks {
		q, err = p.coordinate(q, j, .1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if j%32 == 0 || j == 254 {
			after, err := p.bound(q, .1, 1)
			if err != nil || after < before-1e-9 {
				t.Fatal("full update bound", err)
			}
			if math.Abs(after-p.uncollapsed(q, .1, 1)) > 1e-9 {
				t.Fatal("full updated bound")
			}
			before = after
		}
	}
}

func TestSpikeSlabMixtureMoments(t *testing.T) {
	s := []observation.Sample{{Bits: 0, Outcome: true}, {Bits: 3, Outcome: false}, {Bits: 7, Outcome: true}}
	p, err := spikeSetup(s, []uint16{1, 2, 3}, []float64{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	q := spikeFactors{[]float64{spikeLogOdds(.1), spikeLogOdds(.4), spikeLogOdds(.8)}, []float64{-.3, .7, 1.2}, []float64{.2, .3, .6}}
	for x := uint16(0); x < 8; x++ {
		features := []float64{sparseVRVMPhi(x, 1), sparseVRVMPhi(x, 2), sparseVRVMPhi(x, 3)}
		mean, second, mass := 0., 0., 0.
		for active := 0; active < 8; active++ {
			weight, mu, v := 1., p.v0*p.ksum, p.v0
			for j, odds := range q.logOdds {
				g, h, _, _ := spikeBernoulli(odds)
				if active&(1<<j) == 0 {
					weight *= h
					continue
				}
				weight *= g
				a := features[j] - p.v0*p.t[j]
				mu += a * q.mean[j]
				v += a * a * q.variance[j]
			}
			mass += weight
			mean += weight * mu
			second += weight * (v + mu*mu)
		}
		gotMean, gotVariance := p.moments(q, features)
		if math.Abs(mass-1) > 1e-14 || math.Abs(gotMean-mean) > 1e-12 || math.Abs(gotVariance-(second-mean*mean)) > 1e-12 {
			t.Fatal("enumerated mixture moments")
		}
	}
}

func BenchmarkSpikeSlabReferenceSweep(b *testing.B) {
	s, masks, xi, q := spikeLargeFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, err := spikeSetup(s, masks, xi)
		if err != nil {
			b.Fatal(err)
		}
		state := q.clone()
		for j := range masks {
			state, err = p.coordinate(state, j, .1, 1)
			if err != nil {
				b.Fatal(err)
			}
		}
		if _, err = p.bound(state, .1, 1); err != nil {
			b.Fatal(err)
		}
	}
}
