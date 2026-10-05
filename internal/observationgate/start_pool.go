package observationgate

import "math"

// StartPoolGate is a research alternative to any-window thresholding. Windows
// share observations, so wealth is averaged, never multiplied as independent
// evidence. The conditional null and fixed monitoring budget remain unchanged.
type StartPoolGate struct {
	base  Gate
	Alert [4]bool // fixed, grid, pooled fixed, pooled grid
}

func (g *StartPoolGate) Observe(d float64) ([4]bool, error) {
	a, err := g.base.Observe(d)
	if err != nil {
		return g.Alert, err
	}
	g.Alert[0], g.Alert[1] = a[0], a[1]
	var fixed, grid [8]float64
	for j := range g.base.Starts {
		// A future deterministic start holds wealth one, whose log is zero.
		if g.base.Next-1 < j*64 {
			continue
		}
		var fs, gs [2]float64
		for k, s := range g.base.Starts[j] {
			fs[k], gs[k] = s.Log[2], logMean(s.Log[:])
		}
		fixed[j], grid[j] = logMean(fs[:]), logMean(gs[:])
	}
	if logMean(fixed[:]) >= math.Log(100) {
		g.Alert[2] = true
	}
	if logMean(grid[:]) >= math.Log(100) {
		g.Alert[3] = true
	}
	return g.Alert, nil
}
