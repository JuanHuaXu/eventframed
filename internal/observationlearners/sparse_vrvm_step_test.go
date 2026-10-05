package observationlearners

import (
	"fmt"
	"math"
	"math/bits"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type sparseVRVMStep struct {
	n, p                          int
	masks                         [256]uint16
	mean, diagonal, priorVariance [256]float64
	factor                        [64][256]float64
	residual, logdet              float64
}

func sparseVRVMPhi(x, mask uint16) float64 {
	if bits.OnesCount16(mask & ^x)%2 != 0 {
		return -1
	}
	return 1
}

func sparseVRVMVariance(prior, removed float64) (float64, error) {
	v := prior - removed
	if math.IsNaN(v) || math.IsInf(v, 0) || v < -1e-10*math.Max(1, prior) {
		return 0, fmt.Errorf("sparse variance cancellation")
	}
	return math.Max(0, v), nil
}

func (m *sparseVRVMStep) moments(x uint16) (float64, float64, error) {
	if m == nil || m.n < 1 || m.p < 1 || x >= 512 {
		return 0, 0, fmt.Errorf("invalid sparse moments")
	}
	mu, prior, removed := 0., 0., 0.
	for j := 0; j < m.p; j++ {
		mu += m.mean[j] * sparseVRVMPhi(x, m.masks[j])
		prior += m.priorVariance[j]
	}
	for i := 0; i < m.n; i++ {
		v := 0.
		for j := 0; j < m.p; j++ {
			v += m.factor[i][j] * sparseVRVMPhi(x, m.masks[j])
		}
		removed += v * v
	}
	variance, e := sparseVRVMVariance(prior, removed)
	if math.IsNaN(mu) || math.IsInf(mu, 0) {
		return 0, 0, fmt.Errorf("sparse mean overflow")
	}
	return mu, variance, e
}

// Woodbury update keeps the solve in sample space. No feature is pruned and
// no Gamma hyperparameter is fitted here; callers supply expected precisions.
func sparseVRVMGaussian(s []observation.Sample, masks []uint16, precision, xi []float64) (*sparseVRVMStep, error) {
	n, p := len(s), len(masks)
	if n < 1 || n > 64 || p < 1 || p > 256 || len(precision) != p || len(xi) != n {
		return nil, fmt.Errorf("sparse dimensions")
	}
	m := &sparseVRVMStep{n: n, p: p}
	var seen [512]bool
	for j, mask := range masks {
		if mask >= 512 || seen[mask] || precision[j] <= 0 || math.IsNaN(precision[j]) || math.IsInf(precision[j], 0) {
			return nil, fmt.Errorf("sparse basis or precision")
		}
		seen[mask] = true
		m.masks[j] = mask
		m.priorVariance[j] = 1 / precision[j]
		m.logdet -= math.Log(precision[j])
		if math.IsInf(m.priorVariance[j], 0) {
			return nil, fmt.Errorf("sparse prior overflow")
		}
	}
	var u [64][256]float64
	var b [256]float64
	for i, r := range s {
		if r.Bits >= 512 || xi[i] < 0 || math.IsNaN(xi[i]) || math.IsInf(xi[i], 0) {
			return nil, fmt.Errorf("sparse input")
		}
		root := math.Sqrt(2 * variationalLambda(xi[i]))
		y := -.5
		if r.Outcome {
			y = .5
		}
		for j, mask := range masks {
			v := sparseVRVMPhi(r.Bits, mask)
			u[i][j] = root * v
			b[j] += v * y
		}
	}
	var l [64][64]float64
	for i := 0; i < n; i++ {
		for j := 0; j <= i; j++ {
			v := 0.
			if i == j {
				v = 1
			}
			for k := 0; k < p; k++ {
				v += u[i][k] * m.priorVariance[k] * u[j][k]
			}
			for k := 0; k < j; k++ {
				v -= l[i][k] * l[j][k]
			}
			if i == j {
				if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
					return nil, fmt.Errorf("sparse Cholesky")
				}
				l[i][j] = math.Sqrt(v)
				m.logdet -= 2 * math.Log(l[i][j])
			} else {
				l[i][j] = v / l[j][j]
			}
		}
	}
	for j := 0; j < p; j++ {
		removed := 0.
		for i := 0; i < n; i++ {
			v := u[i][j] * m.priorVariance[j]
			for k := 0; k < i; k++ {
				v -= l[i][k] * m.factor[k][j]
			}
			m.factor[i][j] = v / l[i][i]
			removed += m.factor[i][j] * m.factor[i][j]
		}
		v, e := sparseVRVMVariance(m.priorVariance[j], removed)
		if e != nil {
			return nil, e
		}
		m.diagonal[j] = v
	}
	var fb [64]float64
	for i := 0; i < n; i++ {
		for j := 0; j < p; j++ {
			fb[i] += m.factor[i][j] * b[j]
		}
	}
	for j := 0; j < p; j++ {
		m.mean[j] = m.priorVariance[j] * b[j]
		for i := 0; i < n; i++ {
			m.mean[j] -= m.factor[i][j] * fb[i]
		}
	}
	var logits [64]float64
	for i, r := range s {
		for j, mask := range masks {
			logits[i] += m.mean[j] * sparseVRVMPhi(r.Bits, mask)
		}
	}
	for j, mask := range masks {
		v := precision[j]*m.mean[j] - b[j]
		for i, r := range s {
			v += 2 * variationalLambda(xi[i]) * sparseVRVMPhi(r.Bits, mask) * logits[i]
		}
		m.residual = math.Max(m.residual, math.Abs(v)/math.Max(1, math.Abs(b[j])))
	}
	if math.IsNaN(m.residual) || math.IsInf(m.residual, 0) || m.residual > 1e-8 {
		return nil, fmt.Errorf("sparse normal residual %g", m.residual)
	}
	return m, nil
}

// Independent coefficient-space reference, intentionally not used by fitter.
func sparseVRVMDense(s []observation.Sample, masks []uint16, precision, xi []float64) ([]float64, [][]float64) {
	p := len(masks)
	a := make([][]float64, p)
	l := make([][]float64, p)
	b := make([]float64, p)
	phi := func(x, mask uint16) float64 {
		v := 1.
		for bit := 0; bit < 9; bit++ {
			if mask&(1<<bit) != 0 {
				v *= 2*float64((x>>bit)&1) - 1
			}
		}
		return v
	}
	for j := 0; j < p; j++ {
		a[j] = make([]float64, p)
		l[j] = make([]float64, p)
		a[j][j] = precision[j]
	}
	for i, r := range s {
		y := -.5
		if r.Outcome {
			y = .5
		}
		for j := 0; j < p; j++ {
			v := phi(r.Bits, masks[j])
			b[j] += v * y
			for k := 0; k < p; k++ {
				a[j][k] += 2 * variationalLambda(xi[i]) * v * phi(r.Bits, masks[k])
			}
		}
	}
	for i := 0; i < p; i++ {
		for j := 0; j <= i; j++ {
			v := a[i][j]
			for k := 0; k < j; k++ {
				v -= l[i][k] * l[j][k]
			}
			if i == j {
				l[i][j] = math.Sqrt(v)
			} else {
				l[i][j] = v / l[j][j]
			}
		}
	}
	solve := func(rhs []float64) []float64 {
		z := make([]float64, p)
		x := make([]float64, p)
		for i := 0; i < p; i++ {
			v := rhs[i]
			for j := 0; j < i; j++ {
				v -= l[i][j] * z[j]
			}
			z[i] = v / l[i][i]
		}
		for i := p - 1; i >= 0; i-- {
			v := z[i]
			for j := i + 1; j < p; j++ {
				v -= l[j][i] * x[j]
			}
			x[i] = v / l[i][i]
		}
		return x
	}
	cov := make([][]float64, p)
	for i := 0; i < p; i++ {
		cov[i] = make([]float64, p)
	}
	for j := 0; j < p; j++ {
		unit := make([]float64, p)
		unit[j] = 1
		x := solve(unit)
		for i := 0; i < p; i++ {
			cov[i][j] = x[i]
		}
	}
	return solve(b), cov
}

func TestSparseVRVMGaussian(t *testing.T) {
	for _, p := range []int{10, 256} {
		masks := []uint16{0}
		for d := 1; d <= 4; d++ {
			for m := uint16(1); m < 512; m++ {
				if bits.OnesCount16(m) == d && len(masks) < p {
					masks = append(masks, m)
				}
			}
		}
		for _, n := range []int{1, 7, 64} {
			s := make([]observation.Sample, n)
			xi := make([]float64, n)
			precision := make([]float64, p)
			for i := range s {
				s[i] = observation.Sample{Bits: uint16(i % 13 * 37), Outcome: i%3 == 0}
				xi[i] = float64(i % 9)
			}
			for j := range precision {
				precision[j] = []float64{.03, 1, 12, 1000}[j%4]
			}
			beforeS := append([]observation.Sample(nil), s...)
			beforeXi := append([]float64(nil), xi...)
			beforePrecision := append([]float64(nil), precision...)
			beforeMasks := append([]uint16(nil), masks...)
			m, e := sparseVRVMGaussian(s, masks, precision, xi)
			if e != nil {
				t.Fatal(e)
			}
			mean, cov := sparseVRVMDense(s, masks, precision, xi)
			for j := 0; j < p; j++ {
				if math.Abs(mean[j]-m.mean[j]) > 1e-8 {
					t.Fatal("mean", n, p, j)
				}
				if math.Abs(cov[j][j]-m.diagonal[j]) > 1e-8 {
					t.Fatal("stored diagonal", n, p, j)
				}
				for k := 0; k < p; k++ {
					v := 0.
					if j == k {
						v = m.priorVariance[j]
					}
					for i := 0; i < n; i++ {
						v -= m.factor[i][j] * m.factor[i][k]
					}
					if math.Abs(v-cov[j][k]) > 1e-8 {
						t.Fatal("covariance", n, p, j, k)
					}
				}
			}
			for _, x := range []uint16{0, 7, 55, 255, 511} {
				wantM, wantV := 0., 0.
				for j, mask := range masks {
					v := sparseVRVMPhi(x, mask)
					wantM += mean[j] * v
					for k, mk := range masks {
						wantV += v * cov[j][k] * sparseVRVMPhi(x, mk)
					}
				}
				gotM, gotV, e := m.moments(x)
				if e != nil || math.Abs(gotM-wantM) > 1e-8 || math.Abs(gotV-wantV) > 1e-7 {
					t.Fatal("moments", n, p, x, e)
				}
			}
			if !reflect.DeepEqual(s, beforeS) || !reflect.DeepEqual(xi, beforeXi) || !reflect.DeepEqual(precision, beforePrecision) || !reflect.DeepEqual(masks, beforeMasks) {
				t.Fatal("input mutation")
			}
		}
	}
}

func TestSparseVRVMOriginalControl(t *testing.T) {
	masks := []uint16{0, 1, 2, 4, 8, 16, 32, 64, 128, 256}
	precision := make([]float64, 10)
	for i := range precision {
		precision[i] = 1
	}
	s := make([]observation.Sample, 32)
	xi := make([]float64, 32)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 13), Outcome: i%3 == 0}
		xi[i] = float64(i % 7)
	}
	want, e := variationalStep(s, xi)
	if e != nil {
		t.Fatal(e)
	}
	got, e := sparseVRVMGaussian(s, masks, precision, xi)
	if e != nil {
		t.Fatal(e)
	}
	for j := 0; j < 10; j++ {
		if math.Abs(want.mean[j]-got.mean[j]) > 1e-11 {
			t.Fatal("original mean")
		}
	}
	for x := uint16(0); x < 512; x++ {
		a, b := want.moments(x)
		c, d, e := got.moments(x)
		if e != nil || math.Abs(a-c) > 1e-11 || math.Abs(b-d) > 1e-11 {
			t.Fatal("original moments", x, e)
		}
	}
}

func TestSparseVRVMInvalid(t *testing.T) {
	s := []observation.Sample{{}}
	for _, p := range []float64{0, -1, math.NaN(), math.Inf(1), math.SmallestNonzeroFloat64} {
		if _, e := sparseVRVMGaussian(s, []uint16{0}, []float64{p}, []float64{0}); e == nil {
			t.Fatal("precision")
		}
	}
	for _, x := range []float64{-1, math.NaN(), math.Inf(1)} {
		if _, e := sparseVRVMGaussian(s, []uint16{0}, []float64{1}, []float64{x}); e == nil {
			t.Fatal("xi")
		}
	}
	for _, masks := range [][]uint16{nil, {512}, {0, 0}} {
		precision := make([]float64, len(masks))
		for j := range precision {
			precision[j] = 1
		}
		if _, e := sparseVRVMGaussian(s, masks, precision, []float64{0}); e == nil {
			t.Fatal("mask")
		}
	}
	for _, samples := range [][]observation.Sample{nil, make([]observation.Sample, 65), {{Bits: 512}}} {
		if _, e := sparseVRVMGaussian(samples, []uint16{0}, []float64{1}, make([]float64, len(samples))); e == nil {
			t.Fatal("samples")
		}
	}
	if _, e := sparseVRVMVariance(1, 2); e == nil {
		t.Fatal("negative variance")
	}
	if v, e := sparseVRVMVariance(1, 1+1e-12); e != nil || v != 0 {
		t.Fatal("roundoff")
	}
}

func BenchmarkSparseVRVMGaussian(b *testing.B) {
	masks := []uint16{}
	for m := uint16(0); m < 512; m++ {
		if bits.OnesCount16(m) <= 4 {
			masks = append(masks, m)
		}
	}
	s := make([]observation.Sample, 64)
	xi := make([]float64, 64)
	precision := make([]float64, 256)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 7), Outcome: i%3 == 0}
		xi[i] = 1
	}
	for j := range precision {
		precision[j] = 1
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, e := sparseVRVMGaussian(s, masks, precision, xi); e != nil {
			b.Fatal(e)
		}
	}
}
