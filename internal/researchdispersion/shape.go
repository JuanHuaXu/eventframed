package researchdispersion

import (
	"errors"
	"math"
)

const ShapeKernels = Strengths + 9

// ShapeModel averages a fixed catalogue of rate distributions. Kernel
// weights predict future trials; they do not authenticate or license groups.
type ShapeModel struct {
	p, cache   []float64
	n, success []uint16
	logs, w    [Hypotheses * ShapeKernels]float64
}

// NewShape retains the original mean family. Half the shape prior retains
// V34's dispersion mixture; half covers nine mean-preserving two-point laws.
func NewShape(base []float64) (*ShapeModel, error) {
	old, err := New(base, "adaptive")
	if err != nil {
		return nil, err
	}
	m := &ShapeModel{p: old.p, n: old.n, success: old.success, cache: make([]float64, len(base)*Hypotheses*ShapeKernels)}
	for h := 0; h < Hypotheses; h++ {
		prior := 0.
		for k := 0; k < Strengths; k++ {
			prior += old.w[h*Strengths+k]
		}
		for k := 0; k < ShapeKernels; k++ {
			mass := .5 / 9
			if k < Strengths {
				mass = .5 / Strengths
			}
			z := h*ShapeKernels + k
			m.w[z] = prior * mass
			m.logs[z] = math.Log(m.w[z])
			for i := range base {
				m.cache[i*Hypotheses*ShapeKernels+z] = m.p[i*Hypotheses+h]
			}
		}
	}
	return m, nil
}

func atomBounds(p float64, kernel int) (float64, float64) {
	d := .05 * float64(kernel-Strengths+1)
	return math.Max(.01, p-d), math.Min(.99, p+d)
}

func shapeConditional(p float64, kernel, n, s int) float64 {
	if n == 0 {
		return p
	}
	if kernel < Strengths {
		c := concentrations[kernel]
		if c == 0 {
			return p
		}
		return (c*p + float64(s)) / (c + float64(n))
	}
	lo, hi := atomBounds(p, kernel)
	// Prior high-mass=(p-lo)/(hi-lo) preserves the mean even near bounds.
	logOdds := math.Log((p-lo)/(hi-p)) + float64(s)*math.Log(hi/lo) + float64(n-s)*math.Log((1-hi)/(1-lo))
	high := 0.
	if logOdds >= 0 {
		high = 1 / (1 + math.Exp(-logOdds))
	} else {
		x := math.Exp(logOdds)
		high = x / (1 + x)
	}
	return lo*(1-high) + hi*high
}

func (m *ShapeModel) Predict(i int) (float64, error) {
	if i < 0 || i >= len(m.n) {
		return 0, errors.New("unknown member")
	}
	row := m.cache[i*Hypotheses*ShapeKernels : (i+1)*Hypotheses*ShapeKernels]
	q := 0.
	for z, p := range row {
		q += m.w[z] * p
	}
	return q, nil
}

// Observe uses the cached pre-outcome kernel for evidence, then refreshes
// only this member's conditional kernels after successful normalization.
func (m *ShapeModel) Observe(i, ordinal int, useful bool) error {
	if i < 0 || i >= len(m.n) || ordinal != int(m.n[i])+1 || ordinal > MaxTrials {
		return errors.New("unknown, replayed, out-of-order or capped trial")
	}
	row := m.cache[i*Hypotheses*ShapeKernels : (i+1)*Hypotheses*ShapeKernels]
	var logs, weights [Hypotheses * ShapeKernels]float64
	maximum := math.Inf(-1)
	for z, p := range row {
		if !useful {
			p = 1 - p
		}
		logs[z] = m.logs[z] + math.Log(p)
		maximum = math.Max(maximum, logs[z])
	}
	sum := 0.
	for z, v := range logs {
		weights[z] = math.Exp(v - maximum)
		sum += weights[z]
	}
	if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
		return errors.New("invalid normalization")
	}
	logSum := math.Log(sum)
	for z := range logs {
		m.logs[z], m.w[z] = logs[z]-maximum-logSum, weights[z]/sum
	}
	m.n[i]++
	if useful {
		m.success[i]++
	}
	for h := 0; h < Hypotheses; h++ {
		p := m.p[i*Hypotheses+h]
		for k := 0; k < ShapeKernels; k++ {
			row[h*ShapeKernels+k] = shapeConditional(p, k, int(m.n[i]), int(m.success[i]))
		}
	}
	return nil
}

func (m *ShapeModel) Shapes() (weights [ShapeKernels]float64) {
	for z, mass := range m.w {
		weights[z%ShapeKernels] += mass
	}
	return weights
}
