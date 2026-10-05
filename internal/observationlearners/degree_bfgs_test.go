package observationlearners

import (
	"fmt"
	"math"
	"math/bits"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type degreeBFGSStats struct {
	learnedDegreeStats
	HessianResets, QPPatterns int
}
type degreeMatrix [5][5]float64

func degreeBFGSFinite(e learnedDegreeEvaluation, gradient bool) error {
	if math.IsNaN(e.Value) || math.IsInf(e.Value, 0) {
		return fmt.Errorf("BFGS nonfinite objective")
	}
	if gradient {
		for _, g := range e.Gradient {
			if math.IsNaN(g) || math.IsInf(g, 0) {
				return fmt.Errorf("BFGS nonfinite gradient")
			}
		}
	}
	return nil
}

func degreeIdentity() degreeMatrix {
	var b degreeMatrix
	for i := range b {
		b[i][i] = 1
	}
	return b
}
func degreeSmallSolve(b degreeMatrix, rhs [5]float64, free []int) ([5]float64, error) {
	var l degreeMatrix
	var z, out [5]float64
	for ii, i := range free {
		for jj := 0; jj <= ii; jj++ {
			j := free[jj]
			v := b[i][j]
			for k := 0; k < jj; k++ {
				v -= l[ii][k] * l[jj][k]
			}
			if ii == jj {
				if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
					return out, fmt.Errorf("BFGS principal Hessian")
				}
				l[ii][jj] = math.Sqrt(v)
			} else {
				l[ii][jj] = v / l[jj][jj]
			}
		}
		v := rhs[i]
		for j := 0; j < ii; j++ {
			v -= l[ii][j] * z[j]
		}
		z[ii] = v / l[ii][ii]
	}
	for ii := len(free) - 1; ii >= 0; ii-- {
		i := free[ii]
		v := z[ii]
		for jj := ii + 1; jj < len(free); jj++ {
			v -= l[jj][ii] * out[free[jj]]
		}
		out[i] = v / l[ii][ii]
		if math.IsNaN(out[i]) || math.IsInf(out[i], 0) {
			return out, fmt.Errorf("BFGS nonfinite step")
		}
	}
	return out, nil
}

func degreeBoxStep(b degreeMatrix, g, u [5]float64, learnNoise bool) ([5]float64, int, error) {
	var best [5]float64
	if _, e := degreeSmallSolve(b, [5]float64{}, []int{0, 1, 2, 3, 4}); e != nil {
		return best, 0, e
	}
	dim := 4
	if learnNoise {
		dim = 5
	}
	patterns := 1
	for i := 0; i < dim; i++ {
		patterns *= 3
	}
	bestValue := 0.
	for pattern := 0; pattern < patterns; pattern++ {
		var step, rhs [5]float64
		var free []int
		v := pattern
		for i := 0; i < dim; i++ {
			state := v % 3
			v /= 3
			lo, hi := learnedDegreeBounds(i)
			switch state {
			case 0:
				free = append(free, i)
			case 1:
				step[i] = lo - u[i]
			case 2:
				step[i] = hi - u[i]
			}
		}
		for _, i := range free {
			rhs[i] = -g[i]
			for j := 0; j < dim; j++ {
				rhs[i] -= b[i][j] * step[j]
			}
		}
		solved, e := degreeSmallSolve(b, rhs, free)
		if e != nil {
			return best, pattern + 1, e
		}
		feasible := true
		for _, i := range free {
			step[i] = solved[i]
			lo, hi := learnedDegreeBounds(i)
			if u[i]+step[i] < lo || u[i]+step[i] > hi {
				feasible = false
				break
			}
		}
		if !feasible {
			continue
		}
		value := 0.
		for i := 0; i < dim; i++ {
			value += g[i] * step[i]
			for j := 0; j < dim; j++ {
				value += step[i] * b[i][j] * step[j] / 2
			}
		}
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return best, pattern + 1, fmt.Errorf("BFGS quadratic value")
		}
		if value < bestValue {
			bestValue = value
			best = step
		}
	}
	return best, patterns, nil
}

func degreeDampedUpdate(b degreeMatrix, s, y [5]float64) (degreeMatrix, bool) {
	var bs [5]float64
	sy, sbs := 0., 0.
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			bs[i] += b[i][j] * s[j]
		}
		sy += s[i] * y[i]
		sbs += s[i] * bs[i]
	}
	if sbs <= 1e-18 || math.IsNaN(sbs) || math.IsInf(sbs, 0) || math.IsNaN(sy) || math.IsInf(sy, 0) {
		return degreeIdentity(), true
	}
	if sy < .2*sbs {
		theta := .8 * sbs / (sbs - sy)
		for i := range y {
			y[i] = theta*y[i] + (1-theta)*bs[i]
		}
	}
	sy = 0
	for i := range y {
		sy += s[i] * y[i]
	}
	if sy <= 1e-18 {
		return degreeIdentity(), true
	}
	var next degreeMatrix
	for i := 0; i < 5; i++ {
		for j := 0; j <= i; j++ {
			v := b[i][j] - bs[i]*bs[j]/sbs + y[i]*y[j]/sy
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return degreeIdentity(), true
			}
			next[i][j], next[j][i] = v, v
		}
	}
	if _, e := degreeSmallSolve(next, [5]float64{}, []int{0, 1, 2, 3, 4}); e != nil {
		return degreeIdentity(), true
	}
	return next, false
}

func degreeBoxMinimize(evaluate func([5]float64, bool) (learnedDegreeEvaluation, error), start [5]float64, noise bool, limit int) (learnedDegreeEvaluation, degreeBFGSStats, error) {
	st := degreeBFGSStats{learnedDegreeStats: learnedDegreeStats{Logs: start}}
	var state learnedDegreeEvaluation
	if limit < 1 || limit > 64 || (!noise && start[4] != 0) {
		return state, st, fmt.Errorf("BFGS bounds")
	}
	for i, x := range start {
		lo, hi := learnedDegreeBounds(i)
		if math.IsNaN(x) || math.IsInf(x, 0) || x < lo || x > hi {
			return state, st, fmt.Errorf("BFGS initial point")
		}
	}
	var err error
	state, err = evaluate(start, true)
	st.Evaluations++
	if err == nil {
		err = degreeBFGSFinite(state, true)
	}
	if err != nil {
		return state, st, err
	}
	st.Initial = state.Value
	st.Trace = append(st.Trace, state.Value)
	b := degreeIdentity()
	for iteration := 0; iteration <= limit; iteration++ {
		st.Iterations = iteration
		st.Final = state.Value
		st.ProjectedGradient = 0
		g := state.Gradient
		if !noise {
			g[4] = 0
		}
		for i, v := range g {
			if i == 4 && !noise {
				continue
			}
			lo, hi := learnedDegreeBounds(i)
			st.ProjectedGradient = math.Max(st.ProjectedGradient, math.Abs(st.Logs[i]-math.Max(lo, math.Min(hi, st.Logs[i]-v))))
		}
		if st.ProjectedGradient <= 1e-6 {
			st.Stop = "projected-gradient"
			break
		}
		if iteration == limit {
			st.Stop = "iteration-cap"
			break
		}
		step, patterns, e := degreeBoxStep(b, g, st.Logs, noise)
		st.QPPatterns += patterns
		if e != nil {
			return state, st, e
		}
		accepted := false
		oldLogs := st.Logs
		oldGradient := g
		for bt := 0; bt < 20; bt++ {
			trial := oldLogs
			slope := 0.
			scale := math.Ldexp(1, -bt)
			for i, d := range step {
				if i == 4 && !noise {
					continue
				}
				lo, hi := learnedDegreeBounds(i)
				trial[i] = math.Max(lo, math.Min(hi, oldLogs[i]+scale*d))
				slope += g[i] * (trial[i] - oldLogs[i])
			}
			if slope >= 0 {
				continue
			}
			candidate, e := evaluate(trial, false)
			st.Evaluations++
			if e == nil {
				e = degreeBFGSFinite(candidate, false)
			}
			if e != nil {
				return state, st, e
			}
			if candidate.Value <= state.Value+1e-4*slope {
				st.Logs = trial
				accepted = true
				break
			}
		}
		if !accepted {
			st.Stop = "line-search-cap"
			break
		}
		state, err = evaluate(st.Logs, true)
		st.Evaluations++
		if err == nil {
			err = degreeBFGSFinite(state, true)
		}
		if err != nil {
			return state, st, err
		}
		st.Trace = append(st.Trace, state.Value)
		var s, y [5]float64
		for i := range s {
			s[i] = st.Logs[i] - oldLogs[i]
			y[i] = state.Gradient[i] - oldGradient[i]
		}
		if !noise {
			y[4] = 0
		}
		var reset bool
		b, reset = degreeDampedUpdate(b, s, y)
		if reset {
			st.HessianResets++
		}
	}
	return state, st, nil
}

func fitDegreeBFGS(s []observation.Sample, noise bool) (*spectralModel, degreeBFGSStats, error) {
	start := [5]float64{math.Log(9), math.Log(9), math.Log(9), math.Log(9), 0}
	state, st, e := degreeBoxMinimize(func(u [5]float64, g bool) (learnedDegreeEvaluation, error) { return learnedDegreeEvaluate(s, u, g) }, start, noise, 64)
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

func TestDegreeBFGSQuadratics(t *testing.T) {
	for _, boundary := range []bool{false, true} {
		for _, noise := range []bool{false, true} {
			target := [5]float64{-2, 1, -1, .5, -1}
			if boundary {
				target = [5]float64{-20, 10, -20, 10, -20}
			}
			eval := func(u [5]float64, _ bool) (learnedDegreeEvaluation, error) {
				var out learnedDegreeEvaluation
				for i, v := range u {
					d := v - target[i]
					weight := float64(uint64(1) << i)
					out.Value += weight * d * d / 2
					out.Gradient[i] = weight * d
				}
				return out, nil
			}
			_, st, e := degreeBoxMinimize(eval, [5]float64{}, noise, 64)
			if e != nil {
				t.Fatal(e)
			}
			if st.Stop != "projected-gradient" {
				t.Fatal("quadratic convergence", boundary, noise, st)
			}
			for i, v := range st.Logs {
				lo, hi := learnedDegreeBounds(i)
				want := math.Max(lo, math.Min(hi, target[i]))
				if i == 4 && !noise {
					want = 0
				}
				if math.Abs(v-want) > 2e-6 {
					t.Fatal("quadratic optimum", v, want)
				}
			}
			for i := 1; i < len(st.Trace); i++ {
				if st.Trace[i] > st.Trace[i-1] {
					t.Fatal("descent")
				}
			}
		}
	}
	b := degreeIdentity()
	b[0][0] = -1
	if _, _, e := degreeBoxStep(b, [5]float64{1}, [5]float64{}, true); e == nil {
		t.Fatal("indefinite accepted")
	}
}

func TestDegreeBFGSCurvature(t *testing.T) {
	s := [5]float64{1, 2, 3, 4, 5}
	y := [5]float64{-1, -2, -3, -4, -5}
	damped, reset := degreeDampedUpdate(degreeIdentity(), s, y)
	if reset {
		t.Fatal("unexpected one-step reset")
	}
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			want := -.8 * s[i] * s[j] / 55
			if i == j {
				want++
			}
			if math.Abs(damped[i][j]-want) > 1e-14 {
				t.Fatal("damping formula")
			}
		}
	}
	b := degreeIdentity()
	resets := 0
	for k := 1; k <= 100; k++ {
		var s, y [5]float64
		for i := 0; i < 5; i++ {
			s[i] = math.Sin(float64(k + i))
			y[i] = -s[i] + .1*math.Cos(float64(2*k+i))
		}
		var reset bool
		b, reset = degreeDampedUpdate(b, s, y)
		if reset {
			resets++
			if b != degreeIdentity() {
				t.Fatal("invalid safety reset", k)
			}
		}
		if _, e := degreeSmallSolve(b, [5]float64{}, []int{0, 1, 2, 3, 4}); e != nil {
			t.Fatal(e)
		}
	}
	if resets == 0 {
		t.Fatal("stress did not exercise safety reset")
	}
}
