package observationlearners

import (
	"errors"
	"math"
)

// Diagnostic only: callers construct this positive-semidefinite quadratic
// from full-input forecast rows. No oracle object is used by a serving policy.
type mixtureProfile struct {
	A [5][5]float64
	B [5]float64
	C float64
}

func makeMixtureProfile(rows [][5]float64, truth []float64) (mixtureProfile, error) {
	var p mixtureProfile
	if len(rows) == 0 || len(rows) != len(truth) {
		return p, errors.New("invalid oracle rows")
	}
	for k, row := range rows {
		q := truth[k]
		if math.IsNaN(q) || q < 0 || q > 1 {
			return p, errors.New("invalid oracle truth")
		}
		for _, v := range row {
			if math.IsNaN(v) || v < 0 || v > 1 {
				return p, errors.New("invalid oracle forecast")
			}
		}
		p.C += q / float64(len(rows))
		for i, x := range row {
			p.B[i] += q * x / float64(len(rows))
			for j, y := range row {
				p.A[i][j] += x * y / float64(len(rows))
			}
		}
	}
	return p, nil
}

func (p mixtureProfile) risk(w [5]float64) float64 {
	v := p.C
	for i := range w {
		v -= 2 * w[i] * p.B[i]
		for j := range w {
			v += w[i] * w[j] * p.A[i][j]
		}
	}
	return v
}

type mixtureOptimum struct {
	Weights          [5]float64
	Risk, Lower, Gap float64
}

// Enumerate all nonempty faces of a simplex with at most five vertices.
// A near-singular face is skipped, but success is never inferred from that:
// the returned point must have a convex first-order gap <=1e-8.
func (p mixtureProfile) minimum(n int) (mixtureOptimum, error) {
	best := mixtureOptimum{Risk: math.Inf(1)}
	if n < 1 || n > 5 {
		return best, errors.New("invalid oracle dimension")
	}
	for support := 1; support < 1<<n; support++ {
		var ids []int
		for i := 0; i < n; i++ {
			if support&(1<<i) != 0 {
				ids = append(ids, i)
			}
		}
		k := len(ids)
		var matrix [6][7]float64
		for i, ii := range ids {
			for j, jj := range ids {
				matrix[i][j] = 2 * p.A[ii][jj]
			}
			matrix[i][k] = 1
			matrix[i][k+1] = 2 * p.B[ii]
			matrix[k][i] = 1
		}
		matrix[k][k+1] = 1
		x, ok := solveMixtureSystem(matrix, k+1)
		if !ok {
			continue
		}
		var w [5]float64
		sum := 0.
		valid := true
		for i, j := range ids {
			if !finiteMixture(x[i]) || x[i] < -1e-10 {
				valid = false
				break
			}
			w[j] = math.Max(0, x[i])
			sum += w[j]
		}
		if !valid || sum <= 0 {
			continue
		}
		for i := range w {
			w[i] /= sum
		}
		v := p.risk(w)
		if finiteMixture(v) && v < best.Risk {
			best = mixtureOptimum{Weights: w, Risk: v}
		}
	}
	if !finiteMixture(best.Risk) {
		return best, errors.New("no feasible oracle candidate")
	}
	var grad [5]float64
	minimum, dot := math.Inf(1), 0.
	for i := 0; i < n; i++ {
		grad[i] = -2 * p.B[i]
		for j := 0; j < n; j++ {
			grad[i] += 2 * p.A[i][j] * best.Weights[j]
		}
		minimum = math.Min(minimum, grad[i])
		dot += grad[i] * best.Weights[i]
	}
	best.Gap = math.Max(0, dot-minimum)
	best.Lower = best.Risk - best.Gap
	if !finiteMixture(best.Gap) || best.Gap > 1e-8 {
		return best, errors.New("oracle numerical gap unresolved")
	}
	return best, nil
}

func finiteMixture(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

func solveMixtureSystem(a [6][7]float64, n int) ([6]float64, bool) {
	var out [6]float64
	for col := 0; col < n; col++ {
		pivot := col
		for row := col + 1; row < n; row++ {
			if math.Abs(a[row][col]) > math.Abs(a[pivot][col]) {
				pivot = row
			}
		}
		if !finiteMixture(a[pivot][col]) || math.Abs(a[pivot][col]) < 1e-12 {
			return out, false
		}
		a[col], a[pivot] = a[pivot], a[col]
		scale := a[col][col]
		for j := col; j <= n; j++ {
			a[col][j] /= scale
		}
		for row := 0; row < n; row++ {
			if row != col {
				scale := a[row][col]
				for j := col; j <= n; j++ {
					a[row][j] -= scale * a[col][j]
				}
			}
		}
	}
	for i := 0; i < n; i++ {
		out[i] = a[i][n]
	}
	return out, true
}
