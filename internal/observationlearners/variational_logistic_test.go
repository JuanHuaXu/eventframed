package observationlearners

import (
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func TestVariationalPrior(t *testing.T) {
	m, err := fitVariationalLogistic(nil)
	if err != nil {
		t.Fatal(err)
	}
	if m.bound != 0 || m.mean != [10]float64{} {
		t.Fatal("incorrect prior", m)
	}
	for bits := uint16(0); bits < 512; bits++ {
		mean, v := m.moments(bits)
		if mean != 0 || v != 10 {
			t.Fatal(bits, mean, v)
		}
	}
}

func TestVariationalRepeatedRow(t *testing.T) {
	for _, n := range []int{1, 16, 64, 256} {
		for _, bits := range []uint16{0, 17, 511} {
			for _, label := range []bool{false, true} {
				s := make([]observation.Sample, n)
				for i := range s {
					s[i] = observation.Sample{Bits: bits, Outcome: label}
				}
				m, err := fitVariationalLogistic(s)
				if err != nil {
					t.Fatalf("n=%d bits=%d: %v", n, bits, err)
				}
				// Independent scalar root plus Sherman-Morrison covariance.
				lo, hi := 0., float64(n)*5+10
				for j := 0; j < 100; j++ {
					x := (lo + hi) / 2
					l := math.Tanh(x/2) / (4 * x)
					den := 1 + 20*float64(n)*l
					mu := 5 * float64(n) / den
					if x*x > 10/den+mu*mu {
						hi = x
					} else {
						lo = x
					}
				}
				xi := (lo + hi) / 2
				l := math.Tanh(xi/2) / (4 * xi)
				den := 1 + 20*float64(n)*l
				x := ridgeFeatures(bits)
				sign := 1.
				if !label {
					sign = -1
				}
				for j := range x {
					want := sign * float64(n) * x[j] / (2 * den)
					if math.Abs(m.mean[j]-want) > 2e-7 {
						t.Fatal("mean", n, m.mean[j], want)
					}
					for k := range x {
						want := -2 * float64(n) * l * x[j] * x[k] / den
						if j == k {
							want++
						}
						if math.Abs(m.cov[j][k]-want) > 2e-8 {
							t.Fatal("covariance", n, j, k)
						}
					}
				}
			}
		}
	}
}

func TestVariationalSymmetryAndBound(t *testing.T) {
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16((i * 137) % 512), Outcome: i%3 == 0}
	}
	before := append([]observation.Sample(nil), s...)
	m, err := fitVariationalLogistic(s)
	if err != nil {
		t.Fatal(err)
	}
	other := make([]observation.Sample, len(s))
	for i, v := range s {
		if v != before[i] {
			t.Fatal("mutated input")
		}
		v.Outcome = !v.Outcome
		v.Bits = (v.Bits &^ 3) | ((v.Bits & 1) << 1) | ((v.Bits & 2) >> 1)
		other[len(s)-1-i] = v
	}
	u, err := fitVariationalLogistic(other)
	if err != nil {
		t.Fatal(err)
	}
	for j := range m.mean {
		k := j
		if j == 1 {
			k = 2
		}
		if j == 2 {
			k = 1
		}
		if math.Abs(m.mean[j]+u.mean[k]) > 1e-10 {
			t.Fatal("symmetry")
		}
	}
	xi := make([]float64, len(s))
	for i := range xi {
		xi[i] = 1
	}
	prev := math.Inf(-1)
	for step := 0; step < 100; step++ {
		q, err := variationalStep(s, xi)
		if err != nil {
			t.Fatal(err)
		}
		if q.bound < prev-1e-10 {
			t.Fatal("decreasing ELBO", step, prev, q.bound)
		}
		prev = q.bound
		for i, v := range s {
			a, b := q.moments(v.Bits)
			xi[i] = math.Sqrt(a*a + b)
		}
	}
	if math.Abs(prev-m.bound) > 1e-8 {
		t.Fatal("bound convergence")
	}
}

func TestVariationalInvalidAndCap(t *testing.T) {
	for _, s := range [][]observation.Sample{make([]observation.Sample, 257), {{Bits: 512}}} {
		if m, err := fitVariationalLogistic(s); err == nil || m != nil {
			t.Fatal("invalid fit accepted")
		}
	}
	if m, err := fitVariationalLogisticLimit([]observation.Sample{{Bits: 0, Outcome: true}}, 1); err == nil || m != nil {
		t.Fatal("unconverged fit accepted")
	}
	for _, x := range []float64{-1, math.NaN(), math.Inf(1)} {
		if _, err := variationalStep([]observation.Sample{{}}, []float64{x}); err == nil {
			t.Fatal("invalid xi")
		}
	}
	if variationalLambda(0) != .125 {
		t.Fatal("zero curvature")
	}
}
