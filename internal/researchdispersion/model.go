// Package researchdispersion tests a finite repeated-trial hierarchy in
// isolation. A trial ordinal prevents replay, not false source independence.
// It grants no sharing, provenance, changepoint or production authority.
package researchdispersion

import (
	"errors"
	"math"
)

const (
	Hypotheses = 27
	Strengths  = 5
	MaxTrials  = 64
)

// Zero denotes the point-mass limit phi=p, never Beta(0,0).
var concentrations = [Strengths]float64{.5, 2, 8, 32, 0}

type Model struct {
	p          []float64
	n, success []uint16
	logs, w    [Hypotheses * Strengths]float64
}

// New retains V27's mean family and its priors. Adaptive dispersion has an
// independent uniform prior over five strengths, frozen before outcomes.
func New(base []float64, mode string) (*Model, error) {
	if len(base) < 2 || len(base) > 200 || mode != "adaptive" && mode != "fixed2" && mode != "shared" {
		return nil, errors.New("unsupported frontier or mode")
	}
	m := &Model{p: make([]float64, len(base)*Hypotheses), n: make([]uint16, len(base)), success: make([]uint16, len(base))}
	for i, b := range base {
		if math.IsNaN(b) || math.IsInf(b, 0) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("unsupported baseline")
		}
		row := m.p[i*Hypotheses : (i+1)*Hypotheses]
		row[0], row[1] = b, (.9*b-.05)/.8
		h := 2
		for _, a := range []float64{.1, .3, .5, .7, .9} {
			for _, c := range []float64{-.8, -.4, 0, .4, .8} {
				row[h] = math.Max(.02, math.Min(.98, a+c*float64(i)/float64(len(base)-1)))
				h++
			}
		}
	}
	for h := 0; h < Hypotheses; h++ {
		prior := .004
		if h == 0 {
			prior = .1
		} else if h == 1 {
			prior = .8
		}
		for k := 0; k < Strengths; k++ {
			z := h*Strengths + k
			m.logs[z] = math.Inf(-1)
			if mode == "adaptive" || mode == "fixed2" && k == 1 || mode == "shared" && k == 4 {
				m.w[z] = prior
				if mode == "adaptive" {
					m.w[z] /= Strengths
				}
				m.logs[z] = math.Log(m.w[z])
			}
		}
	}
	return m, nil
}

func (m *Model) conditional(i, h, k int) float64 {
	p := m.p[i*Hypotheses+h]
	c := concentrations[k]
	if c == 0 || m.n[i] == 0 {
		return p
	}
	return (c*p + float64(m.success[i])) / (c + float64(m.n[i]))
}

func (m *Model) Predict(i int) (float64, error) {
	if i < 0 || i >= len(m.n) {
		return 0, errors.New("unknown member")
	}
	q := 0.0
	for h := 0; h < Hypotheses; h++ {
		for k := 0; k < Strengths; k++ {
			q += m.w[h*Strengths+k] * m.conditional(i, h, k)
		}
	}
	return q, nil
}

// Observe accepts the next genuine trial of this member. Its evidence
// likelihood is precisely the pre-outcome kernel from the same joint model.
// Callers must establish independence/identity outside this toy ordinal gate.
func (m *Model) Observe(i int, ordinal int, useful bool) error {
	if i < 0 || i >= len(m.n) || ordinal != int(m.n[i])+1 || ordinal > MaxTrials {
		return errors.New("unknown, replayed, out-of-order or capped trial")
	}
	var logs, weights [Hypotheses * Strengths]float64
	maximum := math.Inf(-1)
	for h := 0; h < Hypotheses; h++ {
		for k := 0; k < Strengths; k++ {
			z := h*Strengths + k
			p := m.conditional(i, h, k)
			if !useful {
				p = 1 - p
			}
			logs[z] = m.logs[z] + math.Log(p)
			maximum = math.Max(maximum, logs[z])
		}
	}
	sum := 0.0
	for z := range logs {
		weights[z] = math.Exp(logs[z] - maximum)
		sum += weights[z]
	}
	if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
		return errors.New("invalid normalization")
	}
	// Commit only after valid normalization; finite log weights can recover
	// from numerical underflow in the exposed probability representation.
	logSum := math.Log(sum)
	for z := range logs {
		m.logs[z] = logs[z] - maximum - logSum
		m.w[z] = weights[z] / sum
	}
	m.n[i]++
	if useful {
		m.success[i]++
	}
	return nil
}

// Dispersion is a copy of the marginal strength weights, not a certificate.
func (m *Model) Dispersion() (weights [Strengths]float64) {
	for h := 0; h < Hypotheses; h++ {
		for k := range weights {
			weights[k] += m.w[h*Strengths+k]
		}
	}
	return weights
}
