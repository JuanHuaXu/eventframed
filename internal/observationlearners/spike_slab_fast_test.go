package observationlearners

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// No p-by-p Gram matrix: a sweep maintains the n fitted linear predictors.
type fastSpikeProfile struct {
	x                        [][]float64
	k, xi, w, t, b, diagonal []float64
	v0, ksum, constant       float64
}

func fastSpikeSetup(samples []observation.Sample, masks []uint16, xi []float64) (*fastSpikeProfile, error) {
	n, p := len(samples), len(masks)
	if n < 1 || n > 64 || p < 1 || p > 255 || len(xi) != n {
		return nil, fmt.Errorf("fast spike dimensions")
	}
	var seen [512]bool
	for _, mask := range masks {
		if mask == 0 || mask >= 512 || seen[mask] {
			return nil, fmt.Errorf("fast spike masks")
		}
		seen[mask] = true
	}
	f := &fastSpikeProfile{x: make([][]float64, n), k: make([]float64, n), xi: append([]float64(nil), xi...), w: make([]float64, n), t: make([]float64, p), b: make([]float64, p), diagonal: make([]float64, p)}
	storage := make([]float64, n*p)
	sw := 0.
	for i, s := range samples {
		if s.Bits >= 512 || xi[i] < 0 || !spikeFinite(xi[i]) {
			return nil, fmt.Errorf("fast spike evidence")
		}
		f.x[i] = storage[i*p : (i+1)*p]
		f.k[i] = -.5
		if s.Outcome {
			f.k[i] = .5
		}
		f.ksum += f.k[i]
		f.w[i] = 2 * variationalLambda(xi[i])
		sw += f.w[i]
		f.constant += -ridgeNLL(xi[i], true) - xi[i]/2 + variationalLambda(xi[i])*xi[i]*xi[i]
		for j, mask := range masks {
			x := sparseVRVMPhi(s.Bits, mask)
			f.x[i][j] = x
			f.t[j] += f.w[i] * x
			f.b[j] += f.k[i] * x
		}
	}
	f.v0 = 1 / (1 + sw)
	f.constant += (math.Log(f.v0) + f.v0*f.ksum*f.ksum) / 2
	for j := range masks {
		f.b[j] -= f.v0 * f.t[j] * f.ksum
		f.diagonal[j] = sw - f.v0*f.t[j]*f.t[j]
		if f.diagonal[j] < 0 || !spikeFinite(f.diagonal[j]) {
			return nil, fmt.Errorf("fast spike diagonal")
		}
	}
	if !spikeFinite(f.constant) {
		return nil, fmt.Errorf("fast spike profile overflow")
	}
	return f, nil
}

func (p *fastSpikeProfile) predictors(q spikeFactors) []float64 {
	v := make([]float64, len(p.x))
	var inclusion [255]float64
	for j := range p.t {
		inclusion[j] = q.probability(j)
	}
	for i, row := range p.x {
		for j, x := range row {
			v[i] += x * inclusion[j] * q.mean[j]
		}
	}
	return v
}
func (p *fastSpikeProfile) bound(q spikeFactors, pi, c2 float64) (float64, error) {
	if err := q.check(len(p.t), pi, c2); err != nil {
		return 0, err
	}
	v := p.constant - q.kl(pi, c2)
	for j := range p.t {
		e, variance := q.moments(j)
		v += p.b[j]*e - p.diagonal[j]*variance/2
	}
	predictors := p.predictors(q)
	quadratic, weighted := 0., 0.
	for i, x := range predictors {
		quadratic += p.w[i] * x * x
		weighted += p.w[i] * x
	}
	v -= (quadratic - p.v0*weighted*weighted) / 2
	if !spikeFinite(v) {
		return 0, fmt.Errorf("fast spike bound overflow")
	}
	return v, nil
}

func (p *fastSpikeProfile) sweep(q spikeFactors, pi, c2 float64) (spikeFactors, error) {
	if err := q.check(len(p.t), pi, c2); err != nil {
		return q, err
	}
	out := q.clone()
	predictors := p.predictors(out)
	for j := range p.t {
		old := out.probability(j) * out.mean[j]
		score, rsum := 0., 0.
		for i, row := range p.x {
			r := p.k[i] - p.w[i]*(predictors[i]-row[j]*old)
			score += row[j] * r
			rsum += r
		}
		score -= p.v0 * p.t[j] * rsum
		variance := 1 / (1/c2 + p.diagonal[j])
		mean := variance * score
		logOdds := spikeLogOdds(pi) + (math.Log(variance/c2)+mean*mean/variance)/2
		inclusion, _, _, _ := spikeBernoulli(logOdds)
		if !spikeFinite(logOdds) || !(variance > 0) || !spikeFinite(variance) || !spikeFinite(mean) {
			return q, fmt.Errorf("fast spike coordinate overflow: coordinate=%d inclusion=%.17g logOdds=%.17g mean=%.17g variance=%.17g", j, inclusion, logOdds, mean, variance)
		}
		out.logOdds[j], out.mean[j], out.variance[j] = logOdds, mean, variance
		change := inclusion*mean - old
		for i, row := range p.x {
			predictors[i] += row[j] * change
		}
	}
	return out, nil
}

func (p *fastSpikeProfile) moments(q spikeFactors, x []float64) (float64, float64) {
	mu, v := p.v0*p.ksum, p.v0
	for j := range p.t {
		e, variance := q.moments(j)
		a := x[j] - p.v0*p.t[j]
		mu += a * e
		v += a * a * variance
	}
	return mu, v
}

func (p *fastSpikeProfile) momentsPrepared(x []float64, means, variances *[255]float64) (float64, float64) {
	mu, v := p.v0*p.ksum, p.v0
	for j := range p.t {
		a := x[j] - p.v0*p.t[j]
		mu += a * means[j]
		v += a * a * variances[j]
	}
	return mu, v
}

func TestSpikeSlabFast(t *testing.T) {
	all, masks, _, initial := spikeLargeFixture()
	maxFactor, maxBound := 0., 0.
	for _, n := range []int{1, 7, 64} {
		for _, d := range []int{1, 9, 255} {
			s := append([]observation.Sample(nil), all[:n]...)
			xi := make([]float64, n)
			for i := range xi {
				xi[i] = float64(i % 7)
			}
			for _, prior := range [][2]float64{{.01, .25}, {.2, 1}, {.8, 4}} {
				pi, c2 := prior[0], prior[1]
				q := spikeFactors{append([]float64(nil), initial.logOdds[:d]...), append([]float64(nil), initial.mean[:d]...), append([]float64(nil), initial.variance[:d]...)}
				for cycle := 0; cycle < 4; cycle++ {
					reference, err := spikeSetup(s, masks[:d], xi)
					if err != nil {
						t.Fatal(err)
					}
					fast, err := fastSpikeSetup(s, masks[:d], xi)
					if err != nil {
						t.Fatal(err)
					}
					before, err := fast.bound(q, pi, c2)
					if err != nil {
						t.Fatal(err)
					}
					saved := q.clone()
					got, err := fast.sweep(q, pi, c2)
					if err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(q, saved) {
						t.Fatal("fast mutates input")
					}
					want := q.clone()
					for j := 0; j < d; j++ {
						want, err = reference.coordinate(want, j, pi, c2)
						if err != nil {
							t.Fatal(err)
						}
					}
					for j := 0; j < d; j++ {
						for _, pair := range [][2]float64{{got.probability(j), want.probability(j)}, {got.logOdds[j], want.logOdds[j]}, {got.mean[j], want.mean[j]}, {got.variance[j], want.variance[j]}} {
							e := math.Abs(pair[0] - pair[1])
							maxFactor = math.Max(maxFactor, e)
							if e > 1e-9 {
								t.Fatal("fast factor mismatch", e)
							}
						}
					}
					after, err := fast.bound(got, pi, c2)
					if err != nil || after < before-1e-9 {
						t.Fatal("fast bound decrease", err)
					}
					rebuilt, err := reference.bound(got, pi, c2)
					if err != nil {
						t.Fatal(err)
					}
					e := math.Abs(after - rebuilt)
					maxBound = math.Max(maxBound, e)
					if e > 1e-9 {
						t.Fatal("fast bound mismatch", e)
					}
					newXi := make([]float64, n)
					for i := range s {
						a, b := fast.moments(got, fast.x[i])
						c, d := reference.moments(got, reference.x[i])
						if math.Abs(a-c) > 1e-10 || math.Abs(b-d) > 1e-10 {
							t.Fatal("fast moments")
						}
						newXi[i] = math.Sqrt(a*a + b)
					}
					next, err := fastSpikeSetup(s, masks[:d], newXi)
					if err != nil {
						t.Fatal(err)
					}
					v, err := next.bound(got, pi, c2)
					if err != nil || v < after-1e-9 {
						t.Fatal("fast xi decrease", err)
					}
					q, xi = got, newXi
				}
			}
		}
	}
	t.Logf("max factor error=%g max bound error=%g", maxFactor, maxBound)
}

func TestSpikeSlabFastInvalid(t *testing.T) {
	s, masks, xi, q := spikeLargeFixture()
	for _, bad := range [][]uint16{nil, {0}, {1, 1}, {512}} {
		if _, err := fastSpikeSetup(s, bad, xi); err == nil {
			t.Fatal("bad masks")
		}
	}
	badXi := append([]float64(nil), xi...)
	badXi[0] = math.NaN()
	if _, err := fastSpikeSetup(s, masks, badXi); err == nil {
		t.Fatal("bad xi")
	}
	p, err := fastSpikeSetup(s, masks, xi)
	if err != nil {
		t.Fatal(err)
	}
	for _, prior := range [][2]float64{{0, 1}, {1, 1}, {.1, 0}, {math.NaN(), 1}, {.1, math.Inf(1)}} {
		if _, err := p.sweep(q, prior[0], prior[1]); err == nil {
			t.Fatal("bad prior")
		}
	}
	saved := q.clone()
	q.mean[0] = math.Inf(1)
	if _, err := p.sweep(q, .1, 1); err == nil {
		t.Fatal("bad factor")
	}
	q = saved
	original := append([]observation.Sample(nil), s...)
	for i := range s {
		s[i].Bits = 511
		s[i].Outcome = !s[i].Outcome
	}
	for i := range xi {
		xi[i] = 99
	}
	ref, err := fastSpikeSetup(original, masks, makeOnes(64))
	if err != nil {
		t.Fatal(err)
	}
	a, err := p.sweep(q, .1, 1)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ref.sweep(q, .1, 1)
	if err != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("profile aliased input", err)
	}
}

func makeOnes(n int) []float64 {
	v := make([]float64, n)
	for i := range v {
		v[i] = 1
	}
	return v
}

func BenchmarkSpikeSlabFastSweep(b *testing.B) {
	s, masks, xi, q := spikeLargeFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, err := fastSpikeSetup(s, masks, xi)
		if err != nil {
			b.Fatal(err)
		}
		state, err := p.sweep(q, .1, 1)
		if err != nil {
			b.Fatal(err)
		}
		if _, err = p.bound(state, .1, 1); err != nil {
			b.Fatal(err)
		}
	}
}
