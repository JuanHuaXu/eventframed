package observationlearners

import (
	"errors"
	"fmt"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

const ridgeDimension = 10
const ridgeIterations = 32
const ridgeTolerance = 1e-8
const ridgeProbabilityFloor = 1e-12

// ridgeLogistic is a research-only penalized-likelihood plug-in, not an
// integrated Bayesian predictive distribution. It retains no training samples.
type ridgeLogistic struct {
	beta  [ridgeDimension]float64
	stats ridgeFitStats
	ready bool
}

type ridgeFitStats struct {
	Iterations, Evaluations int
	GradientInf, Objective  float64
}

func ridgeFeatures(bits uint16) [ridgeDimension]float64 {
	var x [ridgeDimension]float64
	x[0] = 1
	for i := 1; i < ridgeDimension; i++ {
		x[i] = 2*float64((bits>>(i-1))&1) - 1
	}
	return x
}

func ridgeSigmoid(z float64) float64 {
	if z >= 0 {
		return 1 / (1 + math.Exp(-z))
	}
	e := math.Exp(z)
	return e / (1 + e)
}

func ridgeDot(a, b [ridgeDimension]float64) float64 {
	z := 0.
	for i := range a {
		z += a[i] * b[i]
	}
	return z
}

func ridgeNLL(z float64, y bool) float64 {
	if y {
		z = -z
	}
	return math.Max(z, 0) + math.Log1p(math.Exp(-math.Abs(z)))
}

// The objective uses a SUM of likelihood losses and lambda=1, including the
// intercept. This makes the Hessian positive definite even for constant labels
// or rank-deficient inputs. It differs from an unpenalized-intercept GLM fit.
func ridgeObjective(samples []observation.Sample, beta [ridgeDimension]float64) float64 {
	f := ridgeDot(beta, beta) / 2
	for _, s := range samples {
		f += ridgeNLL(ridgeDot(beta, ridgeFeatures(s.Bits)), s.Outcome)
	}
	return f
}

// Evaluate F(beta+delta)-F(beta) without subtracting two large totals near the
// optimum. The small-step softplus identity uses log1p/expm1; larger steps use
// stable losses to avoid exponential overflow. Kahan accumulation preserves
// cancellation between sample and penalty contributions.
func ridgeObjectiveDelta(samples []observation.Sample, beta, delta [ridgeDimension]float64) float64 {
	sum, compensation := 0., 0.
	add := func(v float64) {
		y := v - compensation
		t := sum + y
		compensation = (t - sum) - y
		sum = t
	}
	for i, d := range delta {
		add(beta[i]*d + d*d/2)
	}
	for _, s := range samples {
		x := ridgeFeatures(s.Bits)
		z, dz := ridgeDot(beta, x), ridgeDot(delta, x)
		if math.Abs(dz) <= 1 {
			if s.Outcome {
				z, dz = -z, -dz
			}
			add(math.Log1p(ridgeSigmoid(z) * math.Expm1(dz)))
		} else {
			add(ridgeNLL(z+dz, s.Outcome) - ridgeNLL(z, s.Outcome))
		}
	}
	return sum
}

func ridgeDerivatives(samples []observation.Sample, beta [ridgeDimension]float64) ([ridgeDimension]float64, [ridgeDimension][ridgeDimension]float64) {
	g := beta
	var h [ridgeDimension][ridgeDimension]float64
	for i := range h {
		h[i][i] = 1
	}
	for _, s := range samples {
		x := ridgeFeatures(s.Bits)
		z := ridgeDot(beta, x)
		p, complement := ridgeSigmoid(z), ridgeSigmoid(-z)
		residual := p
		if s.Outcome {
			residual = -complement
		}
		w := p * complement
		for i := range g {
			g[i] += residual * x[i]
			for j := 0; j <= i; j++ {
				h[i][j] += w * x[i] * x[j]
			}
		}
	}
	for i := range h {
		for j := 0; j < i; j++ {
			h[j][i] = h[i][j]
		}
	}
	return g, h
}

// A fixed ten-dimensional Cholesky solve avoids forming an inverse. The ridge
// term supplies strict positive definiteness; numerical failure is explicit.
func ridgeSolve(h [ridgeDimension][ridgeDimension]float64, g [ridgeDimension]float64) ([ridgeDimension]float64, error) {
	var l [ridgeDimension][ridgeDimension]float64
	var y, out [ridgeDimension]float64
	for i := range l {
		for j := 0; j <= i; j++ {
			v := h[i][j]
			for k := 0; k < j; k++ {
				v -= l[i][k] * l[j][k]
			}
			if i == j {
				if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
					return out, errors.New("ridge Hessian is not positive definite")
				}
				l[i][j] = math.Sqrt(v)
			} else {
				l[i][j] = v / l[j][j]
			}
		}
		v := g[i]
		for j := 0; j < i; j++ {
			v -= l[i][j] * y[j]
		}
		y[i] = v / l[i][i]
	}
	for i := ridgeDimension - 1; i >= 0; i-- {
		v := y[i]
		for j := i + 1; j < ridgeDimension; j++ {
			v -= l[j][i] * out[j]
		}
		out[i] = v / l[i][i]
		if math.IsNaN(out[i]) || math.IsInf(out[i], 0) {
			return [ridgeDimension]float64{}, errors.New("non-finite ridge direction")
		}
	}
	return out, nil
}

func fitRidgeLogistic(samples []observation.Sample) (*ridgeLogistic, error) {
	return fitRidgeLogisticLimit(samples, ridgeIterations)
}

func fitRidgeLogisticLimit(samples []observation.Sample, limit int) (*ridgeLogistic, error) {
	if len(samples) == 0 || len(samples) > 256 || limit < 1 || limit > ridgeIterations {
		return nil, errors.New("invalid ridge fitting bounds")
	}
	for _, s := range samples {
		if s.Bits >= 512 {
			return nil, errors.New("invalid ridge input")
		}
	}
	m := new(ridgeLogistic)
	f := ridgeObjective(samples, m.beta)
	m.stats.Evaluations = 1
	for iteration := 0; iteration <= limit; iteration++ {
		g, h := ridgeDerivatives(samples, m.beta)
		norm := 0.
		for _, v := range g {
			norm = math.Max(norm, math.Abs(v))
		}
		if norm <= ridgeTolerance {
			m.stats.Iterations, m.stats.GradientInf, m.stats.Objective = iteration, norm, f
			m.ready = true
			return m, nil
		}
		if iteration == limit {
			return nil, errors.New("ridge iteration cap reached before convergence")
		}
		direction, err := ridgeSolve(h, g)
		if err != nil {
			return nil, err
		}
		descent := ridgeDot(g, direction)
		if descent <= 0 || math.IsNaN(descent) || math.IsInf(descent, 0) {
			return nil, errors.New("invalid ridge descent direction")
		}
		accepted := false
		step := 1.
		lastValue := math.NaN()
		for backtrack := 0; backtrack < 24; backtrack++ {
			trial := m.beta
			var delta [ridgeDimension]float64
			for i := range trial {
				delta[i] = -step * direction[i]
				trial[i] += delta[i]
			}
			value := ridgeObjective(samples, trial)
			change := ridgeObjectiveDelta(samples, m.beta, delta)
			lastValue = value
			m.stats.Evaluations++
			if !math.IsNaN(value) && !math.IsInf(value, 0) && change <= -1e-4*step*descent {
				m.beta, f, accepted = trial, value, true
				break
			}
			step /= 2
		}
		if !accepted {
			return nil, fmt.Errorf("ridge line search exhausted: iteration=%d gradient=%.17g objective=%.17g trial=%.17g decrement=%.17g", iteration, norm, f, lastValue, descent)
		}
	}
	return nil, errors.New("unreachable ridge fit state")
}

func (m *ridgeLogistic) predict(x uint16) (float64, error) {
	if m == nil || !m.ready || x >= 512 {
		return 0, errors.New("invalid ridge prediction")
	}
	z := ridgeDot(m.beta, ridgeFeatures(x))
	if math.IsNaN(z) || math.IsInf(z, 0) {
		return 0, errors.New("non-finite ridge logit")
	}
	// The forecast floor is a declared numerical guard, not a training label or
	// evidence of calibration. Fitting still uses the unmodified logistic loss.
	return math.Max(ridgeProbabilityFloor, math.Min(1-ridgeProbabilityFloor, ridgeSigmoid(z))), nil
}

func (m *ridgeLogistic) conditional(weights [512]float64) (*ConditionalForest, error) {
	out := new(ConditionalForest)
	for x, w := range weights {
		if math.IsNaN(w) || math.IsInf(w, 0) || w <= 0 || w > 1e6 {
			return nil, errors.New("invalid ridge input weight")
		}
		p, err := m.predict(uint16(x))
		if err != nil {
			return nil, err
		}
		out.cells[partialIndex(511, uint16(x))] = conditionalCell{w, w * p}
	}
	// Match the existing ternary completion construction; no current query or
	// outcome can influence this detached compiled law.
	for i := len(out.cells) - 1; i >= 0; i-- {
		v, p := i, 1
		for bit := 0; bit < 9; bit++ {
			if v%3 == 0 {
				a, b := out.cells[i+p], out.cells[i+2*p]
				out.cells[i] = conditionalCell{a.mass + b.mass, a.weighted + b.weighted}
				break
			}
			v /= 3
			p *= 3
		}
	}
	return out, nil
}
