package observationlearners

import (
	"fmt"
	"math"
	"math/bits"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

const sparsePrior = 1e-6
const sparseShape = sparsePrior + .5

type sparseVRVMFit struct {
	gaussian   *sparseVRVMStep
	gaussianXi []float64
	rates      []float64
	xi         []float64
	trace      [][3]float64
	stop       string
}

func sparseVRVMBound(s []observation.Sample, m *sparseVRVMStep, rates, xi []float64) (float64, error) {
	if m == nil || len(s) != m.n || len(rates) != m.p || len(xi) != m.n || m.masks[0] != 0 {
		return 0, fmt.Errorf("sparse bound dimensions")
	}
	v := float64(m.p)/2 + m.logdet/2 - (m.diagonal[0]+m.mean[0]*m.mean[0])/2
	lga, _ := math.Lgamma(sparsePrior)
	lgA, _ := math.Lgamma(sparseShape)
	for j := 1; j < m.p; j++ {
		b := rates[j]
		if b <= 0 || math.IsNaN(b) || math.IsInf(b, 0) {
			return 0, fmt.Errorf("sparse Gamma rate")
		}
		second := m.diagonal[j] + m.mean[j]*m.mean[j]
		v += sparsePrior*math.Log(sparsePrior) - lga - sparseShape*math.Log(b) + lgA + (b-sparsePrior-second/2)*sparseShape/b
	}
	for i, r := range s {
		if xi[i] < 0 || math.IsNaN(xi[i]) || math.IsInf(xi[i], 0) {
			return 0, fmt.Errorf("sparse bound xi")
		}
		mu, variance, e := m.moments(r.Bits)
		if e != nil {
			return 0, e
		}
		k := -.5
		if r.Outcome {
			k = .5
		}
		v += -ridgeNLL(xi[i], true) - xi[i]/2 + variationalLambda(xi[i])*(xi[i]*xi[i]-variance-mu*mu) + k*mu
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("sparse bound nonfinite")
	}
	return v, nil
}

func sparseVRVMNondecrease(before, after float64) error {
	if math.IsNaN(before) || math.IsNaN(after) || math.IsInf(before, 0) || math.IsInf(after, 0) || after < before-1e-8*math.Max(1, math.Abs(before)) {
		return fmt.Errorf("sparse coordinate bound decreased: %g -> %g", before, after)
	}
	return nil
}

func fitSparseVRVM(s []observation.Sample) (*sparseVRVMFit, error) {
	return fitSparseVRVMBudget(s, 64)
}

// Budget changes isolate convergence; the default model and update order stay fixed.
func fitSparseVRVMBudget(s []observation.Sample, limit int) (*sparseVRVMFit, error) {
	if limit < 1 || limit > 1024 {
		return nil, fmt.Errorf("sparse iteration budget")
	}
	if len(s) < 1 || len(s) > 64 {
		return nil, fmt.Errorf("sparse fit samples")
	}
	masks := []uint16{}
	for mask := uint16(0); mask < 512; mask++ {
		if bits.OnesCount16(mask) <= 4 {
			masks = append(masks, mask)
		}
	}
	f := &sparseVRVMFit{rates: make([]float64, len(masks)), xi: make([]float64, len(s))}
	for j := range f.rates {
		f.rates[j] = sparseShape
	}
	for i := range f.xi {
		f.xi[i] = 1
	}
	previous := math.Inf(-1)
	for iteration := 0; iteration < limit; iteration++ {
		precision := make([]float64, len(masks))
		precision[0] = 1
		for j := 1; j < len(masks); j++ {
			precision[j] = sparseShape / f.rates[j]
		}
		m, e := sparseVRVMGaussian(s, masks, precision, f.xi)
		if e != nil {
			return nil, e
		}
		// Record the xi that produced this covariance, before the xi update.
		f.gaussianXi = append(f.gaussianXi[:0], f.xi...)
		before, e := sparseVRVMBound(s, m, f.rates, f.xi)
		if e != nil {
			return nil, e
		}
		if iteration > 0 {
			if e = sparseVRVMNondecrease(previous, before); e != nil {
				return nil, e
			}
		}
		for j := 1; j < len(masks); j++ {
			f.rates[j] = sparsePrior + (m.diagonal[j]+m.mean[j]*m.mean[j])/2
		}
		afterGamma, e := sparseVRVMBound(s, m, f.rates, f.xi)
		if e != nil {
			return nil, e
		}
		if e = sparseVRVMNondecrease(before, afterGamma); e != nil {
			return nil, e
		}
		for i, r := range s {
			mu, variance, e := m.moments(r.Bits)
			if e != nil {
				return nil, e
			}
			f.xi[i] = math.Sqrt(variance + mu*mu)
		}
		afterXi, e := sparseVRVMBound(s, m, f.rates, f.xi)
		if e != nil {
			return nil, e
		}
		if e = sparseVRVMNondecrease(afterGamma, afterXi); e != nil {
			return nil, e
		}
		f.trace = append(f.trace, [3]float64{before, afterGamma, afterXi})
		f.gaussian = m
		if iteration > 0 && math.Abs(afterXi-previous) <= 1e-6 {
			f.stop = "bound-change"
			return f, nil
		}
		previous = afterXi
	}
	f.stop = "iteration-cap"
	return f, nil
}

func sparseVRVMUncollapsed(s []observation.Sample, m *sparseVRVMStep, rates, xi []float64) float64 {
	// Independent entropy/prior bookkeeping; numerical digamma is test-only.
	lg := func(x float64) float64 { v, _ := math.Lgamma(x); return v }
	psi := (lg(sparseShape+1e-5) - lg(sparseShape-1e-5)) / 2e-5
	v := float64(m.p)*(1+math.Log(2*math.Pi))/2 + m.logdet/2
	v -= math.Log(2*math.Pi)/2 + (m.diagonal[0]+m.mean[0]*m.mean[0])/2
	for j := 1; j < m.p; j++ {
		b := rates[j]
		elog := psi - math.Log(b)
		ea := sparseShape / b
		v += .5 * (elog - math.Log(2*math.Pi) - ea*(m.diagonal[j]+m.mean[j]*m.mean[j]))
		v += sparsePrior*math.Log(sparsePrior) - lg(sparsePrior) + (sparsePrior-1)*elog - sparsePrior*ea
		v += sparseShape - math.Log(b) + lg(sparseShape) + (1-sparseShape)*psi
	}
	for i, r := range s {
		mu, varr, e := m.moments(r.Bits)
		if e != nil {
			panic(e)
		}
		signed := -1.
		if r.Outcome {
			signed = 1
		}
		v += -math.Log1p(math.Exp(-xi[i])) + (signed*mu-xi[i])/2 - variationalLambda(xi[i])*(varr+mu*mu-xi[i]*xi[i])
	}
	return v
}

func TestSparseVRVMBound(t *testing.T) {
	for _, p := range []int{10, 256} {
		masks := []uint16{0}
		for d := 1; d <= 4; d++ {
			for m := uint16(1); m < 512; m++ {
				if bits.OnesCount16(m) == d && len(masks) < p {
					masks = append(masks, m)
				}
			}
		}
		s := make([]observation.Sample, 16)
		xi := make([]float64, 16)
		rates := make([]float64, p)
		precision := make([]float64, p)
		for i := range s {
			s[i] = observation.Sample{Bits: uint16(i * 29), Outcome: i%3 == 0}
			xi[i] = float64(i % 4)
		}
		precision[0] = 1
		rates[0] = sparseShape
		for j := 1; j < p; j++ {
			rates[j] = .1 + float64(j%3)
			precision[j] = sparseShape / rates[j]
		}
		m, e := sparseVRVMGaussian(s, masks, precision, xi)
		if e != nil {
			t.Fatal(e)
		}
		_, cov := sparseVRVMDense(s, masks, precision, xi)
		logdet := 0.
		for i := 0; i < p; i++ {
			for j := 0; j <= i; j++ {
				v := cov[i][j]
				for k := 0; k < j; k++ {
					v -= cov[i][k] * cov[j][k]
				}
				if i == j {
					cov[i][j] = math.Sqrt(v)
					logdet += 2 * math.Log(cov[i][j])
				} else {
					cov[i][j] = v / cov[j][j]
				}
			}
		}
		if math.Abs(logdet-m.logdet) > 1e-8 {
			t.Fatal("determinant", p, logdet, m.logdet)
		}
		for _, optimal := range []bool{false, true} {
			if optimal {
				for j := 1; j < p; j++ {
					rates[j] = sparsePrior + (m.diagonal[j]+m.mean[j]*m.mean[j])/2
				}
			}
			got, e := sparseVRVMBound(s, m, rates, xi)
			if e != nil {
				t.Fatal(e)
			}
			want := sparseVRVMUncollapsed(s, m, rates, xi)
			if math.Abs(got-want) > 1e-8 {
				t.Fatal("complete bound", p, optimal, got, want)
			}
		}
	}
}

func TestSparseVRVMFit(t *testing.T) {
	for _, n := range []int{1, 16, 64} {
		s := make([]observation.Sample, n)
		for i := range s {
			s[i] = observation.Sample{Bits: uint16(i % 13 * 37), Outcome: i%3 == 0}
		}
		copy := append([]observation.Sample(nil), s...)
		f, e := fitSparseVRVM(s)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(s, copy) || len(f.trace) > 64 || f.gaussian.priorVariance[0] != 1 {
			t.Fatal("fit state")
		}
		for i, tr := range f.trace {
			if tr[1] < tr[0]-1e-7 || tr[2] < tr[1]-1e-7 || (i > 0 && tr[0] < f.trace[i-1][2]-1e-7) {
				t.Fatal("coordinate descent", i)
			}
		}
		for j := 1; j < len(f.rates); j++ {
			want := sparsePrior + (f.gaussian.diagonal[j]+f.gaussian.mean[j]*f.gaussian.mean[j])/2
			if f.rates[j] != want {
				t.Fatal("Gamma rate")
			}
		}
		for i, r := range s {
			mu, v, e := f.gaussian.moments(r.Bits)
			if e != nil || math.Abs(f.xi[i]*f.xi[i]-v-mu*mu) > 1e-8 {
				t.Fatal("xi update", e)
			}
		}
		for x := uint16(0); x < 512; x++ {
			if _, _, e := f.gaussian.moments(x); e != nil {
				t.Fatal(e)
			}
		}
		t.Logf("n=%d iterations=%d stop=%s bound=%g->%g", n, len(f.trace), f.stop, f.trace[0][0], f.trace[len(f.trace)-1][2])
	}
	if sparseVRVMNondecrease(0, -1) == nil || sparseVRVMNondecrease(math.NaN(), 0) == nil {
		t.Fatal("bound guard")
	}
}

func BenchmarkSparseVRVMFit(b *testing.B) {
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 7), Outcome: i%3 == 0}
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, e := fitSparseVRVM(s); e != nil {
			b.Fatal(e)
		}
	}
}
