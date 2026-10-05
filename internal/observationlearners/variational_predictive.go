package observationlearners

import (
	"errors"
	"math"
)

// A fixed positive-weight normal quadrature on [-10,10], with 256 trapezoid
// panels. Normalization makes this a discrete probability measure. Numerical
// reference checks, not a uniform error theorem, establish its tested accuracy.
// The omitted standard-normal mass is < 1.6e-23; discretization is separate.
var variationalNormalWeights = func() [257]float64 {
	var weights [257]float64
	sum := 0.
	for i := range weights {
		z := -10 + float64(i)*(20./256)
		weights[i] = math.Exp(-z * z / 2)
		if i == 0 || i == 256 {
			weights[i] /= 2
		}
		sum += weights[i]
	}
	for i := range weights {
		weights[i] /= sum
	}
	return weights
}()

func variationalIntegral(mean, variance float64) (float64, error) {
	// Under the unit prior A >= I and ||x||^2=10, so x' Sigma x <=10.
	// Only roundoff-sized excess is tolerated; an unsupported variance fails.
	if math.IsNaN(mean) || math.IsInf(mean, 0) || math.IsNaN(variance) || variance < 0 || variance > 10+1e-10 {
		return 0, errors.New("invalid variational predictive moments")
	}
	if variance == 0 {
		return ridgeSigmoid(mean), nil
	}
	sd := math.Sqrt(math.Min(variance, 10))
	p := 0.
	for i, w := range variationalNormalWeights {
		z := -10 + float64(i)*(20./256)
		p += w * ridgeSigmoid(mean+sd*z)
	}
	return math.Max(0, math.Min(1, p)), nil
}

func (m *variationalLogistic) predict(bits uint16) (float64, error) {
	if m == nil || m.iterations < 1 || bits >= 512 {
		return 0, errors.New("invalid variational prediction")
	}
	mean, variance := m.moments(bits)
	p, err := variationalIntegral(mean, variance)
	if err != nil {
		return 0, err
	}
	// Match the MAP control's numerical forecast floor, without changing fits.
	return math.Max(ridgeProbabilityFloor, math.Min(1-ridgeProbabilityFloor, p)), nil
}

func (m *variationalLogistic) conditional(weights [512]float64) (*ConditionalForest, error) {
	out := new(ConditionalForest)
	for x, w := range weights {
		if math.IsNaN(w) || math.IsInf(w, 0) || w <= 0 || w > 1e6 {
			return nil, errors.New("invalid variational input weight")
		}
		p, err := m.predict(uint16(x))
		if err != nil {
			return nil, err
		}
		out.cells[partialIndex(511, uint16(x))] = conditionalCell{w, w * p}
	}
	// Compile all ternary restrictions from the same full-input law. The table
	// owns its numbers; later model changes cannot rewrite an issued forecast.
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
