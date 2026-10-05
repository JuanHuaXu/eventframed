// Package researchcalibration is a bounded, research-only hierarchical
// predictor. It supplies neither provenance nor sharing/certificate authority.
package researchcalibration

import (
	"errors"
	"math"
	"math/rand"
)

const Hypotheses = 27

type Model struct {
	base         []float64
	p, entropy   []float64
	weights      [Hypotheses]float64
	seen, useful []bool
	count        int
}

func New(base []float64) (*Model, error) {
	if len(base) < 2 || len(base) > 200 {
		return nil, errors.New("unsupported frontier")
	}
	m := &Model{base: append([]float64(nil), base...), p: make([]float64, len(base)*Hypotheses), entropy: make([]float64, len(base)*Hypotheses), seen: make([]bool, len(base)), useful: make([]bool, len(base))}
	m.weights[0], m.weights[1] = .1, .8
	for h := 2; h < Hypotheses; h++ {
		m.weights[h] = .004
	}
	for i, b := range base {
		if math.IsNaN(b) || math.IsInf(b, 0) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("unsupported baseline")
		}
		r := float64(i) / float64(len(base)-1)
		row := m.p[i*Hypotheses : (i+1)*Hypotheses]
		row[0], row[1] = b, (.9*b-.05)/.8
		h := 2
		for _, a := range []float64{.1, .3, .5, .7, .9} {
			for _, c := range []float64{-.8, -.4, 0, .4, .8} {
				row[h] = math.Max(.02, math.Min(.98, a+c*r))
				h++
			}
		}
		for h, p := range row {
			m.entropy[i*Hypotheses+h] = Entropy(p)
		}
	}
	return m, nil
}

func Entropy(p float64) float64 {
	if p <= 0 || p >= 1 {
		return 0
	}
	return -p*math.Log(p) - (1-p)*math.Log1p(-p)
}

func (m *Model) Predict(i int) (float64, error) {
	if i < 0 || i >= len(m.base) {
		return 0, errors.New("unknown candidate")
	}
	q := 0.0
	for h, w := range m.weights {
		p := m.p[i*Hypotheses+h]
		if m.seen[i] {
			y := 0.0
			if m.useful[i] {
				y = 1
			}
			p = (2*p + y) / 3
		}
		q += w * p
	}
	return q, nil
}

// Marginalizing each event rate gives the same likelihood used by the
// posterior-predictive kernel. One distinct event observation is supported.
func (m *Model) Observe(i int, useful bool) error {
	if i < 0 || i >= len(m.base) || m.seen[i] {
		return errors.New("unknown or duplicate evidence")
	}
	var logs [Hypotheses]float64
	maximum := math.Inf(-1)
	for h, w := range m.weights {
		p := m.p[i*Hypotheses+h]
		if !useful {
			p = 1 - p
		}
		logs[h] = math.Log(w) + math.Log(p)
		maximum = math.Max(maximum, logs[h])
	}
	sum := 0.0
	for h := range logs {
		logs[h] = math.Exp(logs[h] - maximum)
		sum += logs[h]
	}
	if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
		return errors.New("invalid posterior normalization")
	}
	for h, v := range logs {
		m.weights[h] = v / sum
	}
	m.seen[i], m.useful[i] = true, useful
	m.count++
	return nil
}

func (m *Model) Select(policy string, rng *rand.Rand) (int, float64, error) {
	if m.count == len(m.base) {
		return -1, 0, errors.New("frontier exhausted")
	}
	if policy == "random" {
		if rng == nil {
			return -1, 0, errors.New("random policy needs RNG")
		}
		n := rng.Intn(len(m.base) - m.count)
		for i, seen := range m.seen {
			if !seen {
				if n == 0 {
					return i, 0, nil
				}
				n--
			}
		}
	}
	if policy == "stratified" {
		size, bits := 1, 0
		for size < len(m.base) {
			size *= 2
			bits++
		}
		for k := 0; k < size; k++ {
			index, x := 0, k
			for j := 0; j < bits; j++ {
				index = (index << 1) | (x & 1)
				x >>= 1
			}
			if index < len(m.base) && !m.seen[index] {
				return index, 0, nil
			}
		}
	}
	best, score := -1, math.Inf(-1)
	for i, seen := range m.seen {
		if seen {
			continue
		}
		if policy == "head" {
			return i, 0, nil
		}
		if policy != "uncertainty" && policy != "information" {
			return -1, 0, errors.New("unknown policy")
		}
		q, _ := m.Predict(i)
		value := Entropy(q)
		if policy == "information" {
			for h, w := range m.weights {
				value -= w * m.entropy[i*Hypotheses+h]
			}
			value = math.Max(0, value)
		}
		if value > score {
			best, score = i, value
		}
	}
	return best, score, nil
}

// Costs are frozen conservative modeled units, not hardware FLOPs or ns.
func SelectionCost(n int, policy string) int64 {
	switch policy {
	case "head", "random":
		return int64(n + 1)
	case "stratified":
		size := 1
		for size < n {
			size *= 2
		}
		return int64(size)
	case "uncertainty":
		return int64(n*(3*Hypotheses+12) + n)
	case "information":
		return int64(n*(6*Hypotheses+12) + n)
	default:
		return -1
	}
}
