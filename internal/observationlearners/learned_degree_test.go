package observationlearners

import (
	"fmt"
	"math"
	"math/bits"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type learnedDegreeStats struct {
	Logs                              [5]float64
	Initial, Final, ProjectedGradient float64
	Evaluations, Iterations           int
	Stop                              string
	Trace                             []float64
}
type learnedDegreeEvaluation struct {
	Value    float64
	Gradient [5]float64
	Alpha    [64]float64
	Residual float64
}

func learnedDegreeBasis(h int) [4]float64 {
	e := [5]float64{1}
	for i := 0; i < 9; i++ {
		v := 1.
		if i < h {
			v = -1
		}
		for d := 4; d >= 1; d-- {
			e[d] += v * e[d-1]
		}
	}
	return [4]float64{e[1] / 9, e[2] / 36, e[3] / 84, e[4] / 126}
}
func learnedDegreeBounds(i int) (float64, float64) {
	if i == 4 {
		return math.Log(.01), math.Log(4)
	}
	return math.Log(1e-4), math.Log(16)
}
func learnedDegreeEvaluate(s []observation.Sample, u [5]float64, gradient bool) (learnedDegreeEvaluation, error) {
	var out learnedDegreeEvaluation
	n := len(s)
	if n < 1 || n > 64 {
		return out, fmt.Errorf("learned degree samples")
	}
	var weights [5]float64
	for i, v := range u {
		lo, hi := learnedDegreeBounds(i)
		if math.IsNaN(v) || math.IsInf(v, 0) || v < lo || v > hi {
			return out, fmt.Errorf("learned degree parameter")
		}
		weights[i] = math.Exp(v)
	}
	var kernel [10]float64
	var basis [10][4]float64
	for h := range kernel {
		basis[h] = learnedDegreeBasis(h)
		kernel[h] = 1
		for d, g := range basis[h] {
			kernel[h] += weights[d] * g
		}
	}
	for _, r := range s {
		if r.Bits >= 512 {
			return out, fmt.Errorf("learned degree input")
		}
	}
	var c, l [64][64]float64
	var y, z [64]float64
	for i, r := range s {
		y[i] = -1
		if r.Outcome {
			y[i] = 1
		}
		for j := 0; j <= i; j++ {
			v := kernel[bits.OnesCount16(r.Bits^s[j].Bits)]
			if i == j {
				v += weights[4]
			}
			c[i][j], c[j][i] = v, v
			for k := 0; k < j; k++ {
				v -= l[i][k] * l[j][k]
			}
			if i == j {
				if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
					return out, fmt.Errorf("learned degree Cholesky")
				}
				l[i][j] = math.Sqrt(v)
				out.Value += math.Log(l[i][j])
			} else {
				l[i][j] = v / l[j][j]
			}
		}
		v := y[i]
		for j := 0; j < i; j++ {
			v -= l[i][j] * z[j]
		}
		z[i] = v / l[i][i]
	}
	for i := n - 1; i >= 0; i-- {
		v := z[i]
		for j := i + 1; j < n; j++ {
			v -= l[j][i] * out.Alpha[j]
		}
		out.Alpha[i] = v / l[i][i]
		out.Value += y[i] * out.Alpha[i] / 2
	}
	for i := 0; i < n; i++ {
		v := -y[i]
		for j := 0; j < n; j++ {
			v += c[i][j] * out.Alpha[j]
		}
		out.Residual = math.Max(out.Residual, math.Abs(v))
	}
	if math.IsNaN(out.Value) || math.IsInf(out.Value, 0) || math.IsNaN(out.Residual) || out.Residual > 1e-8 {
		return out, fmt.Errorf("learned degree numerical residual")
	}
	if gradient {
		// Traces need only one inverse column at a time; no persistent dense inverse.
		for col := 0; col < n; col++ {
			var a, b [64]float64
			for i := 0; i < n; i++ {
				v := 0.
				if i == col {
					v = 1
				}
				for j := 0; j < i; j++ {
					v -= l[i][j] * a[j]
				}
				a[i] = v / l[i][i]
			}
			for i := n - 1; i >= 0; i-- {
				v := a[i]
				for j := i + 1; j < n; j++ {
					v -= l[j][i] * b[j]
				}
				b[i] = v / l[i][i]
			}
			for row := 0; row < n; row++ {
				defect := (b[row] - out.Alpha[row]*out.Alpha[col]) / 2
				for d, g := range basis[bits.OnesCount16(s[row].Bits^s[col].Bits)] {
					out.Gradient[d] += defect * weights[d] * g
				}
				if row == col {
					out.Gradient[4] += defect * weights[4]
				}
			}
		}
		for _, g := range out.Gradient {
			if math.IsNaN(g) || math.IsInf(g, 0) {
				return out, fmt.Errorf("learned degree gradient")
			}
		}
	}
	return out, nil
}

func fitLearnedDegree(s []observation.Sample, learnNoise bool) (*spectralModel, learnedDegreeStats, error) {
	stats := learnedDegreeStats{Logs: [5]float64{math.Log(9), math.Log(9), math.Log(9), math.Log(9), 0}}
	var state learnedDegreeEvaluation
	for iteration := 0; iteration <= 16; iteration++ {
		var err error
		state, err = learnedDegreeEvaluate(s, stats.Logs, true)
		stats.Evaluations++
		if err != nil {
			return nil, stats, err
		}
		if iteration == 0 {
			stats.Initial = state.Value
			stats.Trace = append(stats.Trace, state.Value)
		}
		stats.Final = state.Value
		stats.Iterations = iteration
		var direction [5]float64
		norm, pg := 0., 0.
		for i, g := range state.Gradient {
			if i == 4 && !learnNoise {
				continue
			}
			lo, hi := learnedDegreeBounds(i)
			u := stats.Logs[i]
			pg = math.Max(pg, math.Abs(u-math.Max(lo, math.Min(hi, u-g))))
			if (u == lo && g > 0) || (u == hi && g < 0) {
				g = 0
			}
			direction[i] = g
			norm = math.Max(norm, math.Abs(g))
		}
		stats.ProjectedGradient = pg
		if pg <= 1e-6 {
			stats.Stop = "projected-gradient"
			break
		}
		if iteration == 16 {
			stats.Stop = "iteration-cap"
			break
		}
		accepted := false
		for backtrack := 0; backtrack < 8; backtrack++ {
			trial := stats.Logs
			slope := 0.
			step := math.Ldexp(1, -backtrack) / math.Max(1, norm)
			for i, d := range direction {
				if i == 4 && !learnNoise {
					continue
				}
				lo, hi := learnedDegreeBounds(i)
				trial[i] = math.Max(lo, math.Min(hi, trial[i]-step*d))
				slope += state.Gradient[i] * (trial[i] - stats.Logs[i])
			}
			if slope >= 0 {
				continue
			}
			candidate, err := learnedDegreeEvaluate(s, trial, false)
			stats.Evaluations++
			if err != nil {
				return nil, stats, err
			}
			if candidate.Value <= state.Value+1e-4*slope {
				stats.Logs = trial
				stats.Trace = append(stats.Trace, candidate.Value)
				accepted = true
				break
			}
		}
		if !accepted {
			stats.Stop = "line-search-cap"
			break
		}
	}
	spectralInit()
	m := new(spectralModel)
	m.Residual = state.Residual
	for j, mask := range spectralMasks {
		m.Features = append(m.Features, j)
		variance := 1.
		if mask != 0 {
			d := bits.OnesCount16(mask)
			variance = math.Exp(stats.Logs[d-1]) / []float64{1, 9, 36, 84, 126}[d]
		}
		for i, r := range s {
			m.Coeff[j] += variance * state.Alpha[i] * spectralPhi[r.Bits][j]
		}
	}
	return m, stats, nil
}

func TestLearnedDegreeGradient(t *testing.T) {
	for _, n := range []int{1, 7, 32, 64} {
		s := make([]observation.Sample, n)
		for i := range s {
			s[i] = observation.Sample{Bits: uint16(i * 19 % 512), Outcome: i%3 == 0}
		}
		for _, u := range [][5]float64{{0, 0, 0, 0, 0}, {-3, -1, .5, 1, -2}} {
			e, err := learnedDegreeEvaluate(s, u, true)
			if err != nil {
				t.Fatal(err)
			}
			for i := range u {
				a, b := u, u
				a[i] += 1e-5
				b[i] -= 1e-5
				p, err := learnedDegreeEvaluate(s, a, false)
				if err != nil {
					t.Fatal(err)
				}
				q, err := learnedDegreeEvaluate(s, b, false)
				if err != nil {
					t.Fatal(err)
				}
				fd := (p.Value - q.Value) / 2e-5
				if math.Abs(fd-e.Gradient[i]) > 2e-6 {
					t.Fatal("gradient", n, i, fd, e.Gradient[i])
				}
			}
		}
	}
}

func TestLearnedDegreeContracts(t *testing.T) {
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 19 % 512), Outcome: i%3 == 0}
	}
	u := [5]float64{math.Log(9), math.Log(9), math.Log(9), math.Log(9), 0}
	initial, e := learnedDegreeEvaluate(s, u, false)
	if e != nil {
		t.Fatal(e)
	}
	old, e := fitDegree(s, 0)
	if e != nil {
		t.Fatal(e)
	}
	for x := 0; x < 512; x++ {
		v := 0.
		for i, r := range s {
			g := learnedDegreeBasis(bits.OnesCount16(uint16(x) ^ r.Bits))
			k := 1.
			for _, z := range g {
				k += 9 * z
			}
			v += k * initial.Alpha[i]
		}
		p := math.Max(1e-12, math.Min(1-1e-12, (1+v)/2))
		q, e := old.predict(uint16(x))
		if e != nil || math.Abs(p-q) > 1e-11 {
			t.Fatal("initial agreement", p, q, e)
		}
	}
	for _, noise := range []bool{false, true} {
		m, a, e := fitLearnedDegree(s, noise)
		if e != nil {
			t.Fatal(e)
		}
		if a.Evaluations > 145 || a.Iterations > 16 || a.Final > a.Initial || a.Stop == "" {
			t.Fatal(a)
		}
		for i := 1; i < len(a.Trace); i++ {
			if a.Trace[i] > a.Trace[i-1] {
				t.Fatal("not monotone")
			}
		}
		for i, v := range a.Logs {
			lo, hi := learnedDegreeBounds(i)
			if v < lo || v > hi {
				t.Fatal("parameter bound")
			}
		}
		if !noise && a.Logs[4] != 0 {
			t.Fatal("noise moved")
		}
		for x := 0; x < 512; x++ {
			if _, e := m.predict(uint16(x)); e != nil {
				t.Fatal(e)
			}
		}
	}
	for _, bad := range [][]observation.Sample{nil, make([]observation.Sample, 65), {{Bits: 512}}} {
		if _, _, e := fitLearnedDegree(bad, false); e == nil {
			t.Fatal("bad input accepted")
		}
	}
}

func TestLearnedDegreeExtremeInputs(t *testing.T) {
	for _, truth := range []bool{false, true} {
		s := make([]observation.Sample, 64)
		for i := range s {
			s[i] = observation.Sample{Bits: 511, Outcome: truth}
		}
		for _, noise := range []bool{false, true} {
			m, a, e := fitLearnedDegree(s, noise)
			if e != nil {
				t.Fatal(e)
			}
			p, e := m.predict(511)
			if e != nil || (p > .5) != truth || a.Final > a.Initial {
				t.Fatal(p, a, e)
			}
		}
	}
	s := []observation.Sample{{Bits: 0, Outcome: false}, {Bits: 0, Outcome: true}}
	for _, noise := range []bool{false, true} {
		m, _, e := fitLearnedDegree(s, noise)
		if e != nil {
			t.Fatal(e)
		}
		p, e := m.predict(0)
		if e != nil || math.Abs(p-.5) > 1e-11 {
			t.Fatal("conflicting duplicates", p, e)
		}
	}
	for _, bad := range []float64{math.NaN(), math.Inf(1), math.Log(17)} {
		u := [5]float64{bad, 0, 0, 0, 0}
		if _, e := learnedDegreeEvaluate(s, u, true); e == nil {
			t.Fatal("invalid parameter accepted")
		}
	}
}

func BenchmarkLearnedDegreeFit(b *testing.B) {
	spectralInit()
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 19 % 512), Outcome: i%3 == 0}
	}
	for _, noise := range []bool{false, true} {
		b.Run(fmt.Sprint(noise), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, _, e := fitLearnedDegree(s, noise); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}
