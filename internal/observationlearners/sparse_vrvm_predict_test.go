package observationlearners

import (
	"fmt"
	"math"
	"sort"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type sparseIntegralStats struct {
	Evaluations    int
	EstimatedError float64
	Method         string
}

func sparseVRVMIntegral(mu, variance float64, budget int) (float64, sparseIntegralStats, error) {
	st := sparseIntegralStats{}
	if math.IsNaN(mu) || math.IsInf(mu, 0) || math.IsNaN(variance) || math.IsInf(variance, 0) || variance < 0 || budget < 1 || budget > 16385 {
		return 0, st, fmt.Errorf("sparse predictive moments or budget")
	}
	if mu == 0 {
		st.Method = "symmetry"
		return .5, st, nil
	}
	if variance == 0 {
		st.Method = "point"
		return ridgeSigmoid(mu), st, nil
	}
	const tolerance = 1e-10
	sd := math.Sqrt(variance)
	bound := 2 * math.Log(2) / (sd * math.Sqrt(2*math.Pi))
	if bound <= tolerance {
		st.Method = "logistic-smoothing-bound"
		st.EstimatedError = bound
		return math.Erfc(-mu/sd/math.Sqrt2) / 2, st, nil
	}
	st.Method = "adaptive-Simpson-estimate"
	f := func(z float64) (float64, error) {
		if st.Evaluations >= budget {
			return 0, fmt.Errorf("sparse quadrature evaluation cap")
		}
		st.Evaluations++
		return math.Exp(-z*z/2) / math.Sqrt(2*math.Pi) * ridgeSigmoid(mu+sd*z), nil
	}
	simpson := func(a, b, fa, fm, fb float64) float64 { return (b - a) * (fa + 4*fm + fb) / 6 }
	var integrate func(float64, float64, float64, float64, float64, float64, float64, int) (float64, float64, error)
	integrate = func(a, b, fa, fm, fb, whole, tol float64, depth int) (float64, float64, error) {
		mid := (a + b) / 2
		l, e := f((a + mid) / 2)
		if e != nil {
			return 0, 0, e
		}
		r, e := f((mid + b) / 2)
		if e != nil {
			return 0, 0, e
		}
		left, right := simpson(a, mid, fa, l, fm), simpson(mid, b, fm, r, fb)
		delta := left + right - whole
		estimate := math.Abs(delta) / 15
		// Coarse near-zero integrands can fool an absolute Simpson estimate.
		// Resolve the normal density before permitting early acceptance.
		if depth >= 3 && estimate <= tol {
			return left + right + delta/15, estimate, nil
		}
		if depth == 24 {
			return 0, 0, fmt.Errorf("sparse quadrature depth cap")
		}
		x, ex, e := integrate(a, mid, fa, l, fm, left, tol/2, depth+1)
		if e != nil {
			return 0, 0, e
		}
		y, ey, e := integrate(mid, b, fm, r, fb, right, tol/2, depth+1)
		return x + y, ex + ey, e
	}
	knots := []float64{-10, 0, 10}
	center := -mu / sd
	for _, z := range []float64{center - 40/sd, center, center + 40/sd} {
		if z > -10 && z < 10 {
			knots = append(knots, z)
		}
	}
	sort.Float64s(knots)
	unique := knots[:0]
	for _, z := range knots {
		if len(unique) == 0 || z != unique[len(unique)-1] {
			unique = append(unique, z)
		}
	}
	result := 0.
	for i := 1; i < len(unique); i++ {
		a, b := unique[i-1], unique[i]
		fa, e := f(a)
		if e != nil {
			return 0, st, e
		}
		fm, e := f((a + b) / 2)
		if e != nil {
			return 0, st, e
		}
		fb, e := f(b)
		if e != nil {
			return 0, st, e
		}
		v, errEstimate, e := integrate(a, b, fa, fm, fb, simpson(a, b, fa, fm, fb), tolerance/float64(len(unique)-1), 0)
		if e != nil {
			return 0, st, e
		}
		result += v
		st.EstimatedError += errEstimate
	}
	if math.IsNaN(result) || math.IsInf(result, 0) || result < -tolerance || result > 1+tolerance {
		return 0, st, fmt.Errorf("sparse probability integration")
	}
	return math.Max(0, math.Min(1, result)), st, nil
}

func (f *sparseVRVMFit) predict(x uint16) (float64, error) {
	if f == nil || f.gaussian == nil || len(f.trace) < 1 {
		return 0, fmt.Errorf("unfitted sparse model")
	}
	mu, variance, e := f.gaussian.moments(x)
	if e != nil {
		return 0, e
	}
	p, _, e := sparseVRVMIntegral(mu, variance, 16385)
	if e != nil {
		return 0, e
	}
	return math.Max(1e-12, math.Min(1-1e-12, p)), nil
}

func TestSparseVRVMIntegral(t *testing.T) {
	maxError, maxEvaluations := 0., 0
	for _, mu := range []float64{-20, -3, .3, 3, 20} {
		for _, v := range []float64{.001, 1, 10, 100, 10000} {
			got, st, e := sparseVRVMIntegral(mu, v, 16385)
			if e != nil {
				t.Fatal(mu, v, e)
			}
			// Independent fixed-grid midpoint rule, no adaptive tree or transition splits.
			const panels = 131072
			want := 0.
			for i := 0; i < panels; i++ {
				z := -10 + (float64(i)+.5)*20/panels
				want += math.Exp(-z*z/2) / math.Sqrt(2*math.Pi) * ridgeSigmoid(mu+math.Sqrt(v)*z) * 20 / panels
			}
			error := math.Abs(want - got)
			maxError = math.Max(maxError, error)
			maxEvaluations = max(maxEvaluations, st.Evaluations)
			if error > 1e-8 || st.EstimatedError > 1e-10 {
				t.Fatal("quadrature", mu, v, error, st)
			}
			opposite, _, e := sparseVRVMIntegral(-mu, v, 16385)
			if e != nil || math.Abs(got+opposite-1) > 3e-10 {
				t.Fatal("symmetry", mu, v, got, opposite, got+opposite-1, st, e)
			}
			if v <= 10 {
				old, e := variationalIntegral(mu, v)
				if e != nil || math.Abs(old-got) > 1e-8 {
					t.Fatal("old quadrature", e)
				}
			}
		}
	}
	for _, v := range []float64{0, 10, 1e4, 1e20, math.MaxFloat64} {
		p, _, e := sparseVRVMIntegral(0, v, 16385)
		if e != nil || p != .5 {
			t.Fatal("zero mean")
		}
	}
	for _, mu := range []float64{-1e100, -3, .1, 3, 1e100} {
		p, _, e := sparseVRVMIntegral(mu, 0, 16385)
		if e != nil || p != ridgeSigmoid(mu) {
			t.Fatal("zero variance")
		}
	}
	for _, mu := range []float64{-1e10, 1, 1e10} {
		p, st, e := sparseVRVMIntegral(mu, 1e20, 16385)
		want := math.Erfc(-mu/1e10/math.Sqrt2) / 2
		if e != nil || p != want || st.Method != "logistic-smoothing-bound" || st.EstimatedError > 1e-10 {
			t.Fatal("large variance", p, st, e)
		}
	}
	if _, _, e := sparseVRVMIntegral(.3, 100, 1); e == nil {
		t.Fatal("budget bypass")
	}
	for _, v := range []float64{1e8, 1e12, 1e16, 1e19} {
		for _, ratio := range []float64{-3, .3, 3} {
			mu := ratio * math.Sqrt(v)
			p, st, e := sparseVRVMIntegral(mu, v, 16385)
			cdf := math.Erfc(-ratio/math.Sqrt2) / 2
			bound := 2 * math.Log(2) / math.Sqrt(2*math.Pi*v)
			if e != nil || math.Abs(p-cdf) > bound+1e-10 || st.EstimatedError > 1e-10 {
				t.Fatal("intermediate large variance", mu, v, p, st, e)
			}
		}
	}
	for _, bad := range []float64{math.NaN(), math.Inf(1)} {
		if _, _, e := sparseVRVMIntegral(bad, 1, 16385); e == nil {
			t.Fatal("invalid mean")
		}
		if _, _, e := sparseVRVMIntegral(1, bad, 16385); e == nil {
			t.Fatal("invalid variance")
		}
	}
	if _, _, e := sparseVRVMIntegral(1, -1, 16385); e == nil {
		t.Fatal("negative variance")
	}
	t.Logf("max independent error=%g max evaluations=%d", maxError, maxEvaluations)
}

func TestSparseVRVMFittedPrediction(t *testing.T) {
	s := make([]observation.Sample, 16)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 29), Outcome: i%3 == 0}
	}
	f, e := fitSparseVRVM(s)
	if e != nil {
		t.Fatal(e)
	}
	for x := uint16(0); x < 512; x++ {
		p, e := f.predict(x)
		if e != nil || p <= 0 || p >= 1 {
			t.Fatal("fitted law", x, p, e)
		}
	}
	if _, e := f.predict(512); e == nil {
		t.Fatal("invalid input")
	}
	var empty *sparseVRVMFit
	if _, e := empty.predict(0); e == nil {
		t.Fatal("nil model")
	}
}

func BenchmarkSparseVRVMIntegral(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, _, e := sparseVRVMIntegral(.3, 100, 16385); e != nil {
			b.Fatal(e)
		}
	}
}
