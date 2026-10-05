// Package researchhybridref independently visits every original position.
// It imports no candidate/cache/span helper and supports only two/three heads.
package researchhybridref

import (
	"errors"
	"math"
)

func End(prior []float64, alpha float64, advice [][]float64, observed map[int]bool) ([]float64, error) {
	n := len(prior)
	if n < 2 || n > 3 || len(advice) > 4096 || math.IsNaN(alpha) || alpha < 0 || alpha > 1 {
		return nil, errors.New("reference domain")
	}
	sum := 0.
	u := make([]float64, n)
	for h, p := range prior {
		if math.IsNaN(p) || p <= 0 || p > 1 {
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
			if math.IsNaN(p) || p <= 0 || p >= 1 {
				return nil, errors.New("reference probability")
			}
			if y, ok := observed[j]; ok {
				if !y {
					p = 1 - p
				}
				u[h] += math.Log(p)
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
