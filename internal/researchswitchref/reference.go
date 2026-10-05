// Package researchswitchref is an independent batch matrix reference.
// It imports neither the candidate nor production packages.
package researchswitchref

import (
	"errors"
	"math"
)

// Filter sums joint paths by dense matrix multiplication without intermediate
// normalization. This intentionally slower small-history reference shares no
// candidate forward-cache, transition helper, ticket or emission code.
func Filter(prior []float64, alpha float64, advice [][]float64, observed map[int]bool) ([]float64, error) {
	n := len(prior)
	if n < 2 || n > 8 || len(advice) > 64 || math.IsNaN(alpha) || alpha < 0 || alpha > 1 {
		return nil, errors.New("invalid batch reference contract")
	}
	u := append([]float64(nil), prior...)
	for j, q := range advice {
		if len(q) != n {
			return nil, errors.New("reference advice arity")
		}
		if j > 0 {
			v := make([]float64, n)
			for g := 0; g < n; g++ {
				for h := 0; h < n; h++ {
					transition := alpha * prior[h]
					if g == h {
						transition += 1 - alpha
					}
					v[h] += u[g] * transition
				}
			}
			u = v
		}
		if y, ok := observed[j]; ok {
			for h := range u {
				p := q[h]
				if !y {
					p = 1 - p
				}
				u[h] *= p
			}
		}
	}
	sum := 0.
	for _, x := range u {
		sum += x
	}
	if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
		return nil, errors.New("invalid batch reference sum")
	}
	for h := range u {
		u[h] /= sum
	}
	return u, nil
}

// Prefixes is a separately normalized dense-matrix reference for long sparse
// histories. It visits EVERY position; it has no unit-span shortcut or cache.
// Linear masses are suitable for these sparse tests, not extreme-revival ones.
func Prefixes(prior []float64, alpha float64, advice [][]float64, observed map[int]bool) ([][]float64, error) {
	n := len(prior)
	if n < 2 || n > 8 || len(advice) > 4096 || math.IsNaN(alpha) || alpha < 0 || alpha > 1 {
		return nil, errors.New("invalid prefix reference contract")
	}
	u := append([]float64(nil), prior...)
	result := make([][]float64, len(advice))
	for j, q := range advice {
		if len(q) != n {
			return nil, errors.New("prefix advice arity")
		}
		if j > 0 {
			v := make([]float64, n)
			for g := 0; g < n; g++ {
				for h := 0; h < n; h++ {
					t := alpha * prior[h]
					if g == h {
						t += 1 - alpha
					}
					v[h] += u[g] * t
				}
			}
			u = v
		}
		if y, ok := observed[j]; ok {
			for h, p := range q {
				if !y {
					p = 1 - p
				}
				u[h] *= p
			}
		}
		sum := 0.
		for _, p := range u {
			sum += p
		}
		if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
			return nil, errors.New("invalid prefix mass")
		}
		for h := range u {
			u[h] /= sum
		}
		result[j] = append([]float64(nil), u...)
	}
	return result, nil
}

// EndWeights independently visits all original positions with a dense reset
// matrix. Log messages preserve static-expert revival. For positive study
// hazards, global rescaling precedes dense multiplication (no span shortcut).
func EndWeights(prior []float64, alpha float64, advice [][3]float64, observed map[int]bool) ([3]float64, error) {
	var u [3]float64
	if len(prior) != 3 || len(advice) > 4096 || math.IsNaN(alpha) || alpha < 0 || alpha > 1 {
		return u, errors.New("invalid study matrix contract")
	}
	var matrix [3][3]float64
	for g := 0; g < 3; g++ {
		u[g] = math.Log(prior[g])
		for h := 0; h < 3; h++ {
			matrix[g][h] = alpha * prior[h]
			if g == h {
				matrix[g][h] += 1 - alpha
			}
		}
	}
	for j, q := range advice {
		if j > 0 && alpha > 0 {
			maximum := math.Max(u[0], math.Max(u[1], u[2]))
			var v [3]float64
			for g := 0; g < 3; g++ {
				mass := math.Exp(u[g] - maximum)
				for h := 0; h < 3; h++ {
					v[h] += mass * matrix[g][h]
				}
			}
			for h := range u {
				if v[h] <= 0 {
					return u, errors.New("study matrix underflow; cannot certify")
				}
				u[h] = maximum + math.Log(v[h])
			}
		}
		if y, ok := observed[j]; ok {
			for h, p := range q {
				if !y {
					p = 1 - p
				}
				if p <= 0 || p >= 1 || math.IsNaN(p) {
					return u, errors.New("invalid study matrix evidence")
				}
				u[h] += math.Log(p)
			}
		}
		maximum := math.Max(u[0], math.Max(u[1], u[2]))
		z := maximum + math.Log(math.Exp(u[0]-maximum)+math.Exp(u[1]-maximum)+math.Exp(u[2]-maximum))
		for h := range u {
			u[h] -= z
		}
	}
	for h := range u {
		u[h] = math.Exp(u[h])
	}
	return u, nil
}
