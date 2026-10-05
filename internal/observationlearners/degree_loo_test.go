package observationlearners

import (
	"fmt"
	"math"
	"math/bits"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// Research-only LOO fitting of kernel variances, not posterior uncertainty.
// Hyperparameter selection consumes the window's labels; LOO is the fitting
// objective, never the independent performance estimate.
func degreeLOOEvaluate(s []observation.Sample, u [5]float64, gradient, clipped bool) (learnedDegreeEvaluation, [64]float64, error) {
	var out learnedDegreeEvaluation
	var means [64]float64
	n := len(s)
	if n < 1 || n > 64 || u[4] != 0 {
		return out, means, fmt.Errorf("LOO sample count or fixed noise")
	}
	var weights [4]float64
	for i := range weights {
		lo, hi := learnedDegreeBounds(i)
		if math.IsNaN(u[i]) || math.IsInf(u[i], 0) || u[i] < lo || u[i] > hi {
			return out, means, fmt.Errorf("LOO parameter")
		}
		weights[i] = math.Exp(u[i])
	}
	for _, x := range s {
		if x.Bits >= 512 {
			return out, means, fmt.Errorf("LOO input")
		}
	}
	var basis [10][4]float64
	var kernel [10]float64
	for h := range basis {
		basis[h] = learnedDegreeBasis(h)
		kernel[h] = 1
		for d, g := range basis[h] {
			kernel[h] += weights[d] * g
		}
	}
	var c, l, inverse [64][64]float64
	var y, forward [64]float64
	for i, x := range s {
		y[i] = -1
		if x.Outcome {
			y[i] = 1
		}
		for j := 0; j <= i; j++ {
			v := kernel[bits.OnesCount16(x.Bits^s[j].Bits)]
			if i == j {
				v++
			}
			c[i][j], c[j][i] = v, v
			for k := 0; k < j; k++ {
				v -= l[i][k] * l[j][k]
			}
			if i == j {
				if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
					return out, means, fmt.Errorf("LOO Cholesky")
				}
				l[i][j] = math.Sqrt(v)
			} else {
				l[i][j] = v / l[j][j]
			}
		}
		v := y[i]
		for j := 0; j < i; j++ {
			v -= l[i][j] * forward[j]
		}
		forward[i] = v / l[i][i]
	}
	for i := n - 1; i >= 0; i-- {
		v := forward[i]
		for j := i + 1; j < n; j++ {
			v -= l[j][i] * out.Alpha[j]
		}
		out.Alpha[i] = v / l[i][i]
	}
	for col := 0; col < n; col++ {
		var z [64]float64
		for i := 0; i < n; i++ {
			v := 0.
			if i == col {
				v = 1
			}
			for j := 0; j < i; j++ {
				v -= l[i][j] * z[j]
			}
			z[i] = v / l[i][i]
		}
		for i := n - 1; i >= 0; i-- {
			v := z[i]
			for j := i + 1; j < n; j++ {
				v -= l[j][i] * inverse[j][col]
			}
			inverse[i][col] = v / l[i][i]
		}
	}
	var lossSlope [64]float64
	for i := 0; i < n; i++ {
		v := -y[i]
		for j := 0; j < n; j++ {
			v += c[i][j] * out.Alpha[j]
		}
		out.Residual = math.Max(out.Residual, math.Abs(v))
		d := inverse[i][i]
		if d <= 0 || math.IsNaN(d) || math.IsInf(d, 0) {
			return out, means, fmt.Errorf("LOO inverse diagonal")
		}
		r := out.Alpha[i] / d
		means[i] = y[i] - r
		if clipped {
			p := (1 + means[i]) / 2
			q := math.Max(1e-12, math.Min(1-1e-12, p))
			error := q - (1+y[i])/2
			out.Value += error * error / float64(n)
			if p > 1e-12 && p < 1-1e-12 {
				lossSlope[i] = -error / float64(n)
			}
		} else {
			out.Value += r * r / (4 * float64(n))
			lossSlope[i] = r / (2 * float64(n))
		}
	}
	if gradient {
		for d := 0; d < 4; d++ {
			// Z = D*V; only its diagonal after V*Z and Z*y are needed.
			var z [64][64]float64
			for i := 0; i < n; i++ {
				for k := 0; k < n; k++ {
					v := weights[d] * basis[bits.OnesCount16(s[i].Bits^s[k].Bits)][d]
					for j := 0; j < n; j++ {
						z[i][j] += v * inverse[k][j]
					}
				}
			}
			var da [64]float64
			for i := 0; i < n; i++ {
				for j := 0; j < n; j++ {
					da[i] += z[i][j] * y[j]
				}
			}
			for i := 0; i < n; i++ {
				aPrime, diagonalPrime := 0., 0.
				for k := 0; k < n; k++ {
					aPrime -= inverse[i][k] * da[k]
					diagonalPrime -= inverse[i][k] * z[k][i]
				}
				diag := inverse[i][i]
				rPrime := (aPrime*diag - out.Alpha[i]*diagonalPrime) / (diag * diag)
				out.Gradient[d] += lossSlope[i] * rPrime
			}
		}
	}
	if out.Residual > 1e-8 || math.IsNaN(out.Residual) {
		return out, means, fmt.Errorf("LOO solve residual")
	}
	return out, means, degreeBFGSFinite(out, gradient)
}

func fitDegreeLOO(s []observation.Sample, clipped bool) (*spectralModel, degreeBFGSStats, error) {
	start := [5]float64{math.Log(9), math.Log(9), math.Log(9), math.Log(9), 0}
	state, st, e := degreeBoxMinimize(func(u [5]float64, g bool) (learnedDegreeEvaluation, error) {
		v, _, err := degreeLOOEvaluate(s, u, g, clipped)
		return v, err
	}, start, false, 64)
	if e != nil {
		return nil, st, e
	}
	spectralInit()
	m := new(spectralModel)
	m.Residual = state.Residual
	for j, mask := range spectralMasks {
		m.Features = append(m.Features, j)
		variance := 1.
		if mask != 0 {
			d := bits.OnesCount16(mask)
			variance = math.Exp(st.Logs[d-1]) / []float64{1, 9, 36, 84, 126}[d]
		}
		for i, r := range s {
			m.Coeff[j] += variance * state.Alpha[i] * spectralPhi[r.Bits][j]
		}
	}
	return m, st, nil
}

func TestDegreeLOOIdentities(t *testing.T) {
	for _, n := range []int{1, 2, 7, 32, 64} {
		s := make([]observation.Sample, n)
		for i := range s {
			s[i] = observation.Sample{Bits: uint16((i % 11) * 19), Outcome: i%3 == 0}
		}
		u := [5]float64{-3, -1, .5, 1, 0}
		_, mu, e := degreeLOOEvaluate(s, u, true, false)
		if e != nil {
			t.Fatal(e)
		}
		for i := range s {
			others := append([]observation.Sample(nil), s[:i]...)
			others = append(others, s[i+1:]...)
			want := 0.
			if len(others) > 0 {
				fit, err := learnedDegreeEvaluate(others, u, false)
				if err != nil {
					t.Fatal(err)
				}
				for j, x := range others {
					k := 1.
					for d, g := range learnedDegreeBasis(bits.OnesCount16(x.Bits ^ s[i].Bits)) {
						k += math.Exp(u[d]) * g
					}
					want += k * fit.Alpha[j]
				}
			}
			if math.Abs(want-mu[i]) > 1e-10 {
				t.Fatal("explicit LOO", n, i, want, mu[i])
			}
			flip := append([]observation.Sample(nil), s...)
			flip[i].Outcome = !flip[i].Outcome
			_, got, err := degreeLOOEvaluate(flip, u, false, false)
			if err != nil || math.Abs(got[i]-mu[i]) > 1e-10 {
				t.Fatal("held label leak", n, i, err)
			}
		}
		for _, clipped := range []bool{false, true} {
			base, _, err := degreeLOOEvaluate(s, u, true, clipped)
			if err != nil {
				t.Fatal(err)
			}
			for d := 0; d < 4; d++ {
				a, b := u, u
				a[d] += 1e-5
				b[d] -= 1e-5
				p, _, err := degreeLOOEvaluate(s, a, false, clipped)
				if err != nil {
					t.Fatal(err)
				}
				q, _, err := degreeLOOEvaluate(s, b, false, clipped)
				if err != nil {
					t.Fatal(err)
				}
				if math.Abs((p.Value-q.Value)/2e-5-base.Gradient[d]) > 1e-7 {
					t.Fatal("LOO gradient", n, d, clipped)
				}
			}
		}
	}
}

func TestDegreeLOOContracts(t *testing.T) {
	for _, s := range [][]observation.Sample{nil, make([]observation.Sample, 65), {{Bits: 512}}} {
		if _, _, e := degreeLOOEvaluate(s, [5]float64{}, true, true); e == nil {
			t.Fatal("invalid sample accepted")
		}
	}
	for _, x := range []float64{math.NaN(), math.Inf(1), math.Log(17)} {
		if _, _, e := degreeLOOEvaluate([]observation.Sample{{}}, [5]float64{x}, true, true); e == nil {
			t.Fatal("invalid parameter accepted")
		}
	}
	if _, _, e := degreeLOOEvaluate([]observation.Sample{{}}, [5]float64{0, 0, 0, 0, math.NaN()}, true, true); e == nil {
		t.Fatal("invalid noise")
	}
	for _, clipped := range []bool{false, true} {
		s := make([]observation.Sample, 16)
		for i := range s {
			s[i] = observation.Sample{Bits: uint16(i * 19), Outcome: i%3 == 0}
		}
		copy := append([]observation.Sample(nil), s...)
		m, st, e := fitDegreeLOO(s, clipped)
		if e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(copy, s) || st.Final > st.Initial || st.Evaluations > 1345 || st.Logs[4] != 0 {
			t.Fatal("fit invariant", st)
		}
		for i := 1; i < len(st.Trace); i++ {
			if st.Trace[i] > st.Trace[i-1]+1e-13 {
				t.Fatal("descent")
			}
		}
		for x := 0; x < 512; x++ {
			if _, e := m.predict(uint16(x)); e != nil {
				t.Fatal(e)
			}
		}
		t.Logf("clipped=%v stop=%s iterations=%d loss=%g->%g", clipped, st.Stop, st.Iterations, st.Initial, st.Final)
	}
}

func TestDegreeLOOClippingBranches(t *testing.T) {
	// A mostly linear majority fit extrapolates past the signed-label bounds;
	// the equal-degree parity fixture did not exercise clipping at all.
	u := [5]float64{math.Log(9), -5, -5, -5, 0}
	saturated := 0
	for seed := uint32(1); seed <= 8; seed++ {
		state := seed
		s := make([]observation.Sample, 32)
		for i := range s {
			state = 1664525*state + 1013904223
			x := uint16(state >> 23)
			s[i] = observation.Sample{Bits: x, Outcome: bits.OnesCount16(x) >= 5}
		}
		for _, clipped := range []bool{false, true} {
			base, mu, e := degreeLOOEvaluate(s, u, true, clipped)
			if e != nil {
				t.Fatal(e)
			}
			for _, m := range mu[:len(s)] {
				if math.Abs(math.Abs(m)-1) < 1e-5 {
					t.Fatal("fixture too close to clipping knot")
				}
				if clipped && math.Abs(m) > 1 {
					saturated++
				}
			}
			for d := 0; d < 4; d++ {
				a, b := u, u
				a[d] += 1e-5
				b[d] -= 1e-5
				p, _, e := degreeLOOEvaluate(s, a, false, clipped)
				if e != nil {
					t.Fatal(e)
				}
				q, _, e := degreeLOOEvaluate(s, b, false, clipped)
				if e != nil {
					t.Fatal(e)
				}
				if math.Abs((p.Value-q.Value)/2e-5-base.Gradient[d]) > 1e-7 {
					t.Fatal("clipping gradient", seed, clipped, d)
				}
			}
		}
	}
	if saturated == 0 {
		t.Fatal("no saturated branches exercised")
	}
	t.Logf("saturated held-out predictions=%d", saturated)
}

func BenchmarkDegreeLOOFit(b *testing.B) {
	spectralInit()
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 19 % 512), Outcome: i%3 == 0}
	}
	for _, clipped := range []bool{false, true} {
		b.Run(fmt.Sprint(clipped), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, _, e := fitDegreeLOO(s, clipped); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}
