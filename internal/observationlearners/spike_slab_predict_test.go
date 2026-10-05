package observationlearners

import (
	"fmt"
	"math"
	"math/cmplx"
	"testing"
)

type spikeLaw struct {
	offset, variance              float64
	inclusion, mean, slabVariance []float64
	exclusion                     []float64
}

func (l spikeLaw) excluded(j int) float64 {
	if l.exclusion != nil {
		return l.exclusion[j]
	}
	return 1 - l.inclusion[j]
}

type spikePredictionStats struct {
	Evaluations, Intervals    int
	EstimatedError, TailBound float64
	RangeClamped              bool
}

func (l spikeLaw) validate() (float64, float64, float64, error) {
	if len(l.mean) > 255 || len(l.inclusion) != len(l.mean) || len(l.slabVariance) != len(l.mean) || (l.exclusion != nil && len(l.exclusion) != len(l.mean)) || !spikeFinite(l.offset) || !spikeFinite(l.variance) || l.variance < 0 {
		return 0, 0, 0, fmt.Errorf("spike law dimensions")
	}
	m, v, expected := math.Abs(l.offset), l.variance, l.offset
	for j, g := range l.inclusion {
		h := l.excluded(j)
		if !(g >= 0 && g <= 1) || !(h >= 0 && h <= 1) || math.Abs(g+h-1) > 4e-16 || !spikeFinite(l.mean[j]) || !spikeFinite(l.slabVariance[j]) || l.slabVariance[j] < 0 {
			return 0, 0, 0, fmt.Errorf("spike law factor")
		}
		if g > 0 {
			m += math.Abs(l.mean[j])
			v += l.slabVariance[j]
			expected += g * l.mean[j]
		}
	}
	if !spikeFinite(m) || !spikeFinite(v) || !spikeFinite(expected) {
		return 0, 0, 0, fmt.Errorf("spike law overflow")
	}
	return m, v, expected, nil
}
func (l spikeLaw) characteristic(u float64) complex128 {
	s, c := math.Sincos(u * l.offset)
	scale := math.Exp(-l.variance * u * u / 2)
	z := complex(scale*c, scale*s)
	for j, g := range l.inclusion {
		if g == 0 {
			continue
		}
		s, c = math.Sincos(u * l.mean[j])
		scale = g * math.Exp(-l.slabVariance[j]*u*u/2)
		z *= complex(l.excluded(j)+scale*c, scale*s)
	}
	return z
}
func (l spikeLaw) probability(budget int) (float64, spikePredictionStats, error) {
	st := spikePredictionStats{TailBound: 2 / math.Pi * math.Atanh(math.Exp(-10*math.Pi))}
	m, v, expected, err := l.validate()
	if err != nil {
		return 0, st, err
	}
	if budget < 1 || budget > 65537 {
		return 0, st, fmt.Errorf("spike predictive budget")
	}
	count := math.Max(8, math.Ceil(10*(1+m+math.Sqrt(v))/math.Pi))
	if !spikeFinite(count) || count > float64(budget/9) {
		return 0, st, fmt.Errorf("spike predictive resolution cap")
	}
	st.Intervals = int(count)
	f := func(u float64) (float64, error) {
		if st.Evaluations >= budget {
			return 0, fmt.Errorf("spike predictive evaluation cap")
		}
		st.Evaluations++
		value := expected / math.Pi
		if u != 0 {
			value = imag(l.characteristic(u)) / math.Sinh(math.Pi*u)
		}
		if !spikeFinite(value) {
			return 0, fmt.Errorf("spike predictive integrand")
		}
		return value, nil
	}
	simpson := func(a, b, fa, fm, fb float64) float64 { return (b - a) * (fa + 4*fm + fb) / 6 }
	var integrate func(float64, float64, float64, float64, float64, float64, float64, int) (float64, float64, error)
	integrate = func(a, b, fa, fm, fb, whole, tol float64, depth int) (float64, float64, error) {
		mid := (a + b) / 2
		fl, err := f((a + mid) / 2)
		if err != nil {
			return 0, 0, err
		}
		fr, err := f((mid + b) / 2)
		if err != nil {
			return 0, 0, err
		}
		left, right := simpson(a, mid, fa, fl, fm), simpson(mid, b, fm, fr, fb)
		delta := left + right - whole
		estimate := math.Abs(delta) / 15
		if depth >= 1 && estimate <= tol {
			return left + right + delta/15, estimate, nil
		}
		if depth == 20 {
			return 0, 0, fmt.Errorf("spike predictive depth cap")
		}
		x, ex, err := integrate(a, mid, fa, fl, fm, left, tol/2, depth+1)
		if err != nil {
			return 0, 0, err
		}
		y, ey, err := integrate(mid, b, fm, fr, fb, right, tol/2, depth+1)
		return x + y, ex + ey, err
	}
	total := .5
	for i := 0; i < st.Intervals; i++ {
		a, b := 10*float64(i)/count, 10*float64(i+1)/count
		fa, err := f(a)
		if err != nil {
			return 0, st, err
		}
		fm, err := f((a + b) / 2)
		if err != nil {
			return 0, st, err
		}
		fb, err := f(b)
		if err != nil {
			return 0, st, err
		}
		value, estimate, err := integrate(a, b, fa, fm, fb, simpson(a, b, fa, fm, fb), 1e-10/count, 0)
		if err != nil {
			return 0, st, err
		}
		total += value
		st.EstimatedError += estimate
	}
	if !spikeFinite(total) || total < -1e-8 || total > 1+1e-8 {
		return 0, st, fmt.Errorf("spike predictive range")
	}
	st.RangeClamped = total < 0 || total > 1
	return math.Max(0, math.Min(1, total)), st, nil
}

func (f *spikeFit) law(x uint16) (spikeLaw, error) {
	if f == nil || f.profile == nil || x >= 512 || len(f.masks) != len(f.profile.t) {
		return spikeLaw{}, fmt.Errorf("spike query")
	}
	if err := f.q.check(len(f.masks), .5, 1); err != nil {
		return spikeLaw{}, err
	}
	p := f.profile
	l := spikeLaw{offset: p.v0 * p.ksum, variance: p.v0, inclusion: make([]float64, len(f.masks)), exclusion: make([]float64, len(f.masks)), mean: make([]float64, len(f.masks)), slabVariance: make([]float64, len(f.masks))}
	for j, mask := range f.masks {
		l.inclusion[j], l.exclusion[j], _, _ = spikeBernoulli(f.q.logOdds[j])
		a := sparseVRVMPhi(x, mask) - p.v0*p.t[j]
		l.mean[j] = a * f.q.mean[j]
		l.slabVariance[j] = a * a * f.q.variance[j]
	}
	_, _, _, err := l.validate()
	return l, err
}
func (f *spikeFit) predict(x uint16) (float64, spikePredictionStats, error) {
	l, err := f.law(x)
	if err != nil {
		return 0, spikePredictionStats{}, err
	}
	p, st, err := l.probability(65537)
	if err != nil {
		return 0, st, err
	}
	return math.Max(1e-12, math.Min(1-1e-12, p)), st, nil
}

func enumerateSpikeProbability(l spikeLaw) (float64, error) {
	if len(l.mean) > 9 {
		return 0, fmt.Errorf("enumeration test size")
	}
	result := 0.
	for state := 0; state < 1<<len(l.mean); state++ {
		weight, mu, v := 1., l.offset, l.variance
		for j, g := range l.inclusion {
			if state&(1<<j) == 0 {
				weight *= l.excluded(j)
			} else {
				weight *= g
				mu += l.mean[j]
				v += l.slabVariance[j]
			}
		}
		p, _, err := sparseVRVMIntegral(mu, v, 16385)
		if err != nil {
			return 0, err
		}
		result += weight * p
	}
	return result, nil
}

func TestSpikePredictiveIntegral(t *testing.T) {
	maxError, maxEval := 0., 0
	for _, mu := range []float64{-20, -3, 0, .3, 3, 20} {
		for _, v := range []float64{0, .001, 1, 10, 100, 10000} {
			l := spikeLaw{offset: mu, variance: v}
			got, st, err := l.probability(65537)
			if err != nil {
				t.Fatal(err)
			}
			want, _, err := sparseVRVMIntegral(mu, v, 16385)
			if err != nil {
				t.Fatal(err)
			}
			e := math.Abs(got - want)
			maxError = math.Max(maxError, e)
			maxEval = max(maxEval, st.Evaluations)
			if e > 2e-9 {
				t.Fatal("Gaussian inversion mismatch", mu, v, got, want)
			}
			if st.TailBound > 1.45e-14 || st.EstimatedError > 1e-10 || st.Evaluations > 65537 {
				t.Fatal("integration accounting")
			}
		}
	}
	for _, d := range []int{1, 3, 9} {
		l := spikeLaw{offset: -.7, variance: .3, inclusion: make([]float64, d), mean: make([]float64, d), slabVariance: make([]float64, d)}
		for j := 0; j < d; j++ {
			l.inclusion[j] = .1 + .2*float64(j%4)
			l.mean[j] = float64(j%5) - 2
			l.slabVariance[j] = .1 + .3*float64(j%3)
		}
		got, st, err := l.probability(65537)
		if err != nil {
			t.Fatal(err)
		}
		want, err := enumerateSpikeProbability(l)
		if err != nil {
			t.Fatal(err)
		}
		e := math.Abs(got - want)
		maxError = math.Max(maxError, e)
		maxEval = max(maxEval, st.Evaluations)
		if e > 2e-9 {
			t.Fatal("mixture inversion mismatch", got, want)
		}
		l.offset = -l.offset
		for j := range l.mean {
			l.mean[j] = -l.mean[j]
		}
		opposite, _, err := l.probability(65537)
		if err != nil || math.Abs(got+opposite-1) > 1e-12 {
			t.Fatal("mixture complement", err)
		}
	}
	// Zero mean is not symmetry. Moment-matched Gaussian prediction would be .5.
	skew := spikeLaw{offset: -1, inclusion: []float64{.1}, mean: []float64{10}, slabVariance: []float64{0}}
	got, _, err := skew.probability(65537)
	want := .9*ridgeSigmoid(-1) + .1*ridgeSigmoid(9)
	if err != nil || math.Abs(got-want) > 2e-9 || math.Abs(got-.5) < .1 {
		t.Fatal("zero-mean asymmetric mixture", got, want, err)
	}
	t.Logf("max independent difference=%g max evaluations=%d skew probability=%g", maxError, maxEval, got)
}

func TestSpikePredictiveCharacteristic(t *testing.T) {
	l := spikeLaw{offset: .3, variance: .2, inclusion: []float64{.1, .4, .8}, mean: []float64{-2, .3, 1}, slabVariance: []float64{.1, .2, .3}}
	for _, u := range []float64{0, .01, .1, 1, 5, 10} {
		var want complex128
		for state := 0; state < 8; state++ {
			weight, mu, v := 1., l.offset, l.variance
			for j, g := range l.inclusion {
				if state&(1<<j) == 0 {
					weight *= 1 - g
				} else {
					weight *= g
					mu += l.mean[j]
					v += l.slabVariance[j]
				}
			}
			want += complex(weight, 0) * cmplx.Exp(complex(-v*u*u/2, mu*u))
		}
		if cmplx.Abs(want-l.characteristic(u)) > 1e-14 {
			t.Fatal("mixture characteristic")
		}
	}
}

func TestSpikePredictiveFit(t *testing.T) {
	s, masks, _, _ := spikeLargeFixture()
	f, err := fitSpikeFixed(s, masks, .1, 1, 1024)
	if err != nil {
		t.Fatal(err)
	}
	maxEvaluations := 0
	for _, x := range []uint16{0, 1, 63, 127, 255, 511} {
		p, st, err := f.predict(x)
		if err != nil || p <= 0 || p >= 1 {
			t.Fatal("fitted prediction", err)
		}
		maxEvaluations = max(maxEvaluations, st.Evaluations)
		l, err := f.law(x)
		if err != nil {
			t.Fatal(err)
		}
		_, _, mean, err := l.validate()
		if err != nil {
			t.Fatal(err)
		}
		features := make([]float64, len(masks))
		for j, m := range masks {
			features[j] = sparseVRVMPhi(x, m)
		}
		want, wantVariance := f.profile.moments(f.q, features)
		variance := l.variance
		for j, g := range l.inclusion {
			variance += g*l.slabVariance[j] + g*l.excluded(j)*l.mean[j]*l.mean[j]
		}
		if math.Abs(mean-want) > 1e-10 || math.Abs(variance-wantVariance) > 1e-10 {
			t.Fatal("conditional intercept query")
		}
		// Query laws own their factors; callers cannot mutate the fitted state.
		l.inclusion[0], l.mean[0], l.slabVariance[0] = 0, 999, 999
		again, _, err := f.predict(x)
		if err != nil || again != p {
			t.Fatal("query law aliases fit", err)
		}
	}
	if _, _, err := f.predict(512); err == nil {
		t.Fatal("out-of-domain query")
	}
	t.Logf("full mixture max evaluations=%d", maxEvaluations)
}

func TestSpikePredictiveInvalid(t *testing.T) {
	for _, l := range []spikeLaw{{offset: math.NaN()}, {variance: -1}, {inclusion: []float64{.5}}, {inclusion: []float64{2}, mean: []float64{0}, slabVariance: []float64{1}}, {inclusion: []float64{.5}, mean: []float64{math.Inf(1)}, slabVariance: []float64{1}}} {
		if _, _, err := l.probability(65537); err == nil {
			t.Fatal("bad law")
		}
	}
	for _, budget := range []int{-1, 0, 1, 65538} {
		if _, _, err := (spikeLaw{offset: 3}).probability(budget); err == nil {
			t.Fatal("bad/exhausted budget")
		}
	}
	if _, _, err := (spikeLaw{offset: 1e10}).probability(65537); err == nil {
		t.Fatal("resolution cap")
	}
	if _, _, err := (*spikeFit)(nil).predict(0); err == nil {
		t.Fatal("nil fit")
	}
}

func BenchmarkSpikeMixturePrediction(b *testing.B) {
	s, masks, _, _ := spikeLargeFixture()
	f, err := fitSpikeFixed(s, masks, .1, 1, 1024)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := f.predict(255); err != nil {
			b.Fatal(err)
		}
	}
}
