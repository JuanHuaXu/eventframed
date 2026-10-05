// Package researchscoreref uses dense position-wise messages and positive-part
// root search. It imports no candidate, cache, span, or closed-form helper.
package researchscoreref

import (
	"errors"
	"math"
)

func valid(mode string) bool {
	return mode == "log_mean" || mode == "log_strong" || mode == "brier_mean" || mode == "brier_strong"
}

func End(prior []float64, alpha float64, advice [][]float64, observed map[int]bool, mode string) ([]float64, error) {
	n := len(prior)
	if !valid(mode) || n < 2 || n > 3 || len(advice) > 4096 || math.IsNaN(alpha) || math.IsInf(alpha, 0) || alpha < 0 || alpha > 1 {
		return nil, errors.New("reference domain")
	}
	sum := 0.
	u := make([]float64, n)
	for h, p := range prior {
		if math.IsNaN(p) || math.IsInf(p, 0) || p <= 0 || p > 1 {
			return nil, errors.New("reference prior")
		}
		sum += p
		u[h] = math.Log(p)
	}
	if math.Abs(sum-1) > 1e-12 {
		return nil, errors.New("reference prior sum")
	}
	for j := range observed {
		if j < 0 || j >= len(advice) {
			return nil, errors.New("reference observation position")
		}
	}
	pi := append([]float64(nil), prior...)
	for h := range pi {
		pi[h] /= sum
		u[h] = math.Log(pi[h])
	}
	var matrix [3][3]float64
	for g := 0; g < n; g++ {
		for h := 0; h < n; h++ {
			matrix[g][h] = alpha * pi[h]
			if g == h {
				matrix[g][h] += 1 - alpha
			}
		}
	}
	for j, q := range advice {
		if len(q) != n {
			return nil, errors.New("reference advice arity")
		}
		if j > 0 && alpha > 0 {
			maximum := math.Inf(-1)
			for _, x := range u {
				maximum = math.Max(maximum, x)
			}
			var v [3]float64
			for g, x := range u {
				mass := math.Exp(x - maximum)
				for h := range u {
					v[h] += mass * matrix[g][h]
				}
			}
			for h := range u {
				if v[h] <= 0 {
					return nil, errors.New("reference transition underflow")
				}
				u[h] = maximum + math.Log(v[h])
			}
		}
		for h, p := range q {
			if math.IsNaN(p) || math.IsInf(p, 0) || p <= 0 || p >= 1 {
				return nil, errors.New("reference probability")
			}
			if y, ok := observed[j]; ok {
				if mode == "brier_mean" || mode == "brier_strong" {
					outcome := 0.
					if y {
						outcome = 1
					}
					d := p - outcome
					u[h] -= 2 * d * d
				} else {
					if !y {
						p = 1 - p
					}
					u[h] += math.Log(p)
				}
			}
		}
		maximum := math.Inf(-1)
		for _, x := range u {
			maximum = math.Max(maximum, x)
		}
		total := 0.
		for _, x := range u {
			total += math.Exp(x - maximum)
		}
		for h := range u {
			u[h] -= maximum + math.Log(total)
		}
	}
	for h := range u {
		u[h] = math.Exp(u[h])
	}
	return u, nil
}

func Forecast(weights, advice []float64, mode string) (float64, error) {
	if !valid(mode) || len(weights) != len(advice) || len(advice) < 2 || len(advice) > 3 {
		return 0, errors.New("reference forecast domain")
	}
	total, q := 0., 0.
	for h, p := range advice {
		w := weights[h]
		if math.IsNaN(w) || math.IsInf(w, 0) || w < 0 || math.IsNaN(p) || math.IsInf(p, 0) || p <= 0 || p >= 1 {
			return 0, errors.New("reference forecast input")
		}
		total += w
		q += w * p
	}
	if math.Abs(total-1) > 1e-10 {
		return 0, errors.New("reference forecast mass")
	}
	if mode == "log_mean" || mode == "brier_mean" {
		return q, nil
	}
	var g [2]float64
	for y := 0; y < 2; y++ {
		mass := 0.
		for h, p := range advice {
			d := p - float64(y)
			mass += weights[h] / total * math.Exp(-2*d*d)
		}
		g[y] = -math.Log(mass)
	}
	lo, hi := math.Min(g[0], g[1])-2, math.Max(g[0], g[1])+2
	for k := 0; k < 96; k++ {
		s := (lo + hi) / 2
		if math.Max(s-g[0], 0)+math.Max(s-g[1], 0) < 2 {
			lo = s
		} else {
			hi = s
		}
	}
	return math.Max((lo+hi)/2-g[1], 0) / 2, nil
}
