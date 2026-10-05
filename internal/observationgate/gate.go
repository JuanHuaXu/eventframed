// Package observationgate tests predictable evidence betting, outside serving.
package observationgate

import (
	"errors"
	"math"
)

type signed struct {
	Log                    [5]float64
	Adaptive, Sum, Squares float64
}

type Gate struct {
	Starts [8][2]signed
	Next   int
	Alert  [4]bool
}

func logMean(xs []float64) float64 {
	m := xs[0]
	for _, x := range xs[1:] {
		m = math.Max(m, x)
	}
	s := 0.
	for _, x := range xs {
		s += math.Exp(x - m)
	}
	return m + math.Log(s/float64(len(xs)))
}

// rate is computed before the revealing sample enters Sum or Squares.
func (s *signed) rate() float64 { return math.Max(.05, math.Min(.8, s.Sum/(1+s.Squares))) }

func (g *Gate) Observe(d float64) ([4]bool, error) {
	if math.IsNaN(d) || math.IsInf(d, 0) || d < -1 || d > 1 {
		return g.Alert, errors.New("invalid paired difference")
	}
	for j := range g.Starts {
		if g.Next < j*64 {
			continue
		}
		var wealth [2][4]float64
		for k, sign := range []float64{1, -1} {
			s := &g.Starts[j][k]
			x := sign*d - .15
			s.Adaptive += math.Log1p(s.rate() * x)
			for a, rate := range []float64{.05, .15, .25, .5, .8} {
				s.Log[a] += math.Log1p(rate * x)
			}
			wealth[k] = [4]float64{s.Log[2], logMean(s.Log[:]), s.Adaptive, logMean([]float64{s.Log[2], s.Adaptive})}
			s.Sum += x
			s.Squares += x * x
		}
		for a := range g.Alert {
			if logMean([]float64{wealth[0][a], wealth[1][a]}) >= math.Log(800) {
				g.Alert[a] = true
			}
		}
	}
	g.Next++
	return g.Alert, nil
}
