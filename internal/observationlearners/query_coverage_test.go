package observationlearners

import (
	"context"
	"fmt"
	"math"
)

type coverageQueryBranch struct {
	Origin            int
	Mass, LogEvidence [2]float64
	Conditional       [2][512]float64
	Gain              [2]float64
}

type coverageQueryRecord struct {
	Phase, Case, Index, Schedule int
	Origins, Pool                []int
	BaseLogEvidence              float64
	Base                         [512]float64
	Weights                      [2][512]uint16
	Branches                     []coverageQueryBranch
	Selected                     [2]int
	Fits                         int
	Error                        string `json:",omitempty"`
}

// The histogram is built from visible inputs, not from a generator/input-law
// oracle. Retaining counts matters when a dependent input occurs repeatedly.
func coverageQueryWeights(input softV120Record) ([2][512]uint16, error) {
	var weights [2][512]uint16
	if len(input.Steps) != 256 {
		return weights, fmt.Errorf("invalid coverage tape")
	}
	for j := 0; j <= 160; j++ {
		x := input.Steps[j].X
		if x >= 512 {
			return weights, fmt.Errorf("invalid coverage input")
		}
		weights[0][x]++
		if j >= 129 {
			weights[1][x]++
		}
	}
	return weights, nil
}

// This explicit-refit reference owns each branch. It deliberately does not
// change the already audited eight-probe batch or production model contracts.
func coverageQueryValue(ctx context.Context, s *regimeQueryState, origin int) (coverageQueryBranch, error) {
	r := coverageQueryBranch{Origin: origin}
	if s == nil || origin < s.left || origin >= s.clock || s.history[origin-s.left].Arrives >= 0 {
		return r, fmt.Errorf("invalid coverage query")
	}
	for y := 0; y < 2; y++ {
		if err := ctx.Err(); err != nil {
			return r, err
		}
		history := append([]segmentPacket(nil), s.history...)
		history[origin-s.left].Outcome = y == 1
		history[origin-s.left].Arrives = s.clock
		m, err := fitSegmentPosteriorBatch(ctx, s.left, s.clock, history, 64, .01, .95)
		if err != nil {
			return r, err
		}
		if len(m.origins) != len(s.base.origins)+1 {
			return r, fmt.Errorf("coverage conditioning evicted evidence")
		}
		pos := 0
		for _, j := range m.origins {
			if j == origin {
				continue
			}
			if pos >= len(s.base.origins) || s.base.origins[pos] != j {
				return r, fmt.Errorf("coverage support changed")
			}
			pos++
		}
		if pos != len(s.base.origins) {
			return r, fmt.Errorf("coverage support lost")
		}
		r.LogEvidence[y] = m.logEvidence
		r.Mass[y] = math.Exp(m.logEvidence - s.base.logEvidence)
		for x, p := range m.predictions {
			r.Conditional[y][x] = .005 + .99*p
		}
	}
	if math.Abs(r.Mass[0]+r.Mass[1]-1) > 1e-10 {
		return r, fmt.Errorf("coverage mass incoherent")
	}
	for x, p := range s.base.predictions {
		base := .005 + .99*p
		if math.Abs(r.Mass[0]*r.Conditional[0][x]+r.Mass[1]*r.Conditional[1][x]-base) > 1e-10 {
			return r, fmt.Errorf("coverage marginal incoherent")
		}
	}
	return r, nil
}

func coverageQueryGain(base [512]float64, b coverageQueryBranch, weights [512]uint16) (float64, error) {
	if math.IsNaN(b.Mass[0]) || math.IsNaN(b.Mass[1]) || b.Mass[0] < 0 || b.Mass[1] < 0 || b.Mass[0] > 1 || b.Mass[1] > 1 || math.Abs(b.Mass[0]+b.Mass[1]-1) > 1e-10 {
		return 0, fmt.Errorf("invalid coverage mass")
	}
	total := 0
	for _, n := range weights {
		total += int(n)
	}
	if total == 0 {
		return 0, fmt.Errorf("empty coverage weights")
	}
	gain, reduction := 0., 0.
	for x, n := range weights {
		if n == 0 {
			continue
		}
		p := base[x]
		if math.IsNaN(p) || p < 0 || p > 1 {
			return 0, fmt.Errorf("invalid coverage base")
		}
		remaining, local := 0., 0.
		for y := 0; y < 2; y++ {
			q := b.Conditional[y][x]
			if math.IsNaN(q) || q < 0 || q > 1 || math.IsNaN(b.Mass[y]) || b.Mass[y] < 0 || b.Mass[y] > 1 {
				return 0, fmt.Errorf("invalid coverage conditional")
			}
			local += b.Mass[y] * (q - p) * (q - p)
			remaining += b.Mass[y] * q * (1 - q)
		}
		w := float64(n) / float64(total)
		gain += w * local
		reduction += w * (p*(1-p) - remaining)
	}
	if math.IsNaN(gain) || math.Abs(gain-reduction) > 1e-10 {
		return 0, fmt.Errorf("coverage risk incoherent")
	}
	return gain, nil
}

func runCoverageQuery(ctx context.Context, input softV120Record) coverageQueryRecord {
	r := coverageQueryRecord{Phase: input.Phase, Case: input.Case, Index: input.Index, Schedule: input.Schedule, Selected: [2]int{-1, -1}}
	if err := ctx.Err(); err != nil {
		r.Error = err.Error()
		return r
	}
	var err error
	r.Weights, err = coverageQueryWeights(input)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	h, pool, _, err := regimeQueryTapeView(input, 160)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	r.Pool = pool
	for i, p := range h {
		if p.Arrives >= 0 {
			r.Origins = append(r.Origins, i-16)
		}
	}
	if len(pool) == 0 {
		return r
	}
	s, err := newRegimeQueryState(ctx, -16, 160, h)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	r.Fits++
	r.BaseLogEvidence = s.base.logEvidence
	for x, p := range s.base.predictions {
		r.Base[x] = .005 + .99*p
	}
	for _, j := range pool {
		b, err := coverageQueryValue(ctx, s, j)
		if err != nil {
			r.Error = err.Error()
			return r
		}
		r.Fits += 2
		for a, w := range r.Weights {
			b.Gain[a], err = coverageQueryGain(r.Base, b, w)
			if err != nil {
				r.Error = err.Error()
				return r
			}
		}
		r.Branches = append(r.Branches, b)
	}
	for a := range r.Selected {
		scores := make([]float64, len(pool))
		for i, b := range r.Branches {
			scores[i] = b.Gain[a]
		}
		r.Selected[a], err = chooseRegimeMaximum(pool, scores)
		if err != nil {
			r.Error = err.Error()
			return r
		}
	}
	return r
}
