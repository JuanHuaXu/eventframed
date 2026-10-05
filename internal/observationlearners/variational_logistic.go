package observationlearners

import (
	"errors"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// Research-only batch Jaakkola-Jordan approximation with a fresh N(0,I)
// prior for every window. Prediction numerically averages over these Gaussian
// moments; neither the variational approximation nor its quadrature is exact.
type variationalLogistic struct {
	mean            [ridgeDimension]float64
	cov             [ridgeDimension][ridgeDimension]float64
	bound, residual float64
	iterations      int
}

func variationalLambda(x float64) float64 {
	if x < 1e-6 {
		return .125 - x*x/96
	}
	return math.Tanh(x/2) / (4 * x)
}

// One Gaussian coordinate maximization. xi belongs to the current window,
// never a previous cached posterior; retaining both would double-count labels.
func variationalStep(samples []observation.Sample, xi []float64) (variationalLogistic, error) {
	var m variationalLogistic
	if len(samples) != len(xi) {
		return m, errors.New("variational size mismatch")
	}
	var a [ridgeDimension][ridgeDimension]float64
	var b [ridgeDimension]float64
	for j := range a {
		a[j][j] = 1
	}
	for i, s := range samples {
		if s.Bits >= 512 || xi[i] < 0 || math.IsNaN(xi[i]) || math.IsInf(xi[i], 0) {
			return m, errors.New("invalid variational input")
		}
		x := ridgeFeatures(s.Bits)
		l := variationalLambda(xi[i])
		y := -.5
		if s.Outcome {
			y = .5
		}
		for j := range b {
			b[j] += y * x[j]
			for k := range b {
				a[j][k] += 2 * l * x[j] * x[k]
			}
		}
		m.bound += -ridgeNLL(xi[i], true) - xi[i]/2 + l*xi[i]*xi[i]
	}
	var err error
	m.mean, err = ridgeSolve(a, b)
	if err != nil {
		return m, err
	}
	for k := range b {
		var unit [ridgeDimension]float64
		unit[k] = 1
		col, err := ridgeSolve(a, unit)
		if err != nil {
			return m, err
		}
		for j := range b {
			m.cov[j][k] = col[j]
		}
	}
	// Cholesky diagonal gives half logdet(A), without determinant overflow.
	var lower [ridgeDimension][ridgeDimension]float64
	for j := range b {
		for k := 0; k <= j; k++ {
			v := a[j][k]
			for h := 0; h < k; h++ {
				v -= lower[j][h] * lower[k][h]
			}
			if j == k {
				if !(v > 0) {
					return m, errors.New("invalid variational precision")
				}
				lower[j][k] = math.Sqrt(v)
				m.bound -= math.Log(lower[j][k])
			} else {
				lower[j][k] = v / lower[k][k]
			}
		}
	}
	m.bound += ridgeDot(b, m.mean) / 2
	if math.IsNaN(m.bound) || math.IsInf(m.bound, 0) {
		return m, errors.New("invalid variational bound")
	}
	return m, nil
}

func (m variationalLogistic) moments(bits uint16) (float64, float64) {
	x := ridgeFeatures(bits)
	v := 0.
	for j := range x {
		for k := range x {
			v += x[j] * m.cov[j][k] * x[k]
		}
	}
	return ridgeDot(x, m.mean), v
}

func fitVariationalLogistic(samples []observation.Sample) (*variationalLogistic, error) {
	return fitVariationalLogisticLimit(samples, 1024)
}

func fitVariationalLogisticLimit(samples []observation.Sample, limit int) (*variationalLogistic, error) {
	if len(samples) > 256 || limit < 1 || limit > 1024 {
		return nil, errors.New("invalid variational bounds")
	}
	xi := make([]float64, len(samples))
	for i := range xi {
		xi[i] = 1
	}
	previous := math.Inf(-1)
	for iteration := 1; iteration <= limit; iteration++ {
		m, err := variationalStep(samples, xi)
		if err != nil {
			return nil, err
		}
		if m.bound < previous-1e-10*(1+math.Abs(previous)) {
			return nil, errors.New("variational bound decreased")
		}
		previous = m.bound
		for i, s := range samples {
			mean, v := m.moments(s.Bits)
			if !(v > 0) || math.IsInf(v, 0) {
				return nil, errors.New("invalid variational variance")
			}
			next := math.Sqrt(v + mean*mean)
			m.residual = math.Max(m.residual, math.Abs(next-xi[i])/(1+next))
			xi[i] = next
		}
		m.iterations = iteration
		if m.residual <= 1e-10 {
			return &m, nil
		}
	}
	return nil, errors.New("variational iteration cap reached before convergence")
}
