package observationlearners

import (
	"context"
	"math"
)

type populationQueryBranch struct {
	Origin                          int
	Origins                         []int
	Redundant, ActualY              bool
	Q                               float64
	Population, Sample, LogEvidence [2]float64
	AtPublication, Predictions      [2][31]float64
}
type populationQueryRecord struct {
	Phase, Case, Index, Schedule int
	Branches                     []populationQueryBranch
	ActualFits                   int
	Error                        string `json:",omitempty"`
}

// No oracle object is accepted here. Changing an answer only changes the one
// explicitly purchased label, never natural evidence or other candidate labels.
func fitPopulationQueryBranch(input softV120Record, origin int, answer bool) (*segmentPosterior, []int, bool, error) {
	h, origins, redundant, err := regimeOutcomePublication(input, origin)
	if err != nil {
		return nil, nil, false, err
	}
	if origin >= 0 && !redundant {
		h[origin+16].Outcome = answer
	}
	m, err := fitSegmentPosteriorBatch(context.Background(), -16, 161, h, 64, .01, .95)
	return m, origins, redundant, err
}
func runPopulationQuery(input softV120Record) populationQueryRecord {
	r := populationQueryRecord{Phase: input.Phase, Case: input.Case, Index: input.Index, Schedule: input.Schedule}
	choices, err := regimeEnvelopeChoices(input)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	o, err := populationOracle(input)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	k, err := makePopulationKernel(o)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	for j, s := range input.Steps {
		if s.X >= 512 || math.IsNaN(s.Q) || s.Q < 0 || s.Q > 1 {
			r.Error = "invalid stored truth/input"
			return r
		}
		q, e := o.truth(s.X, j)
		if e != nil || math.Abs(q-s.Q) > 1e-14 {
			r.Error = "stored truth mismatch"
			return r
		}
	}
	base, baseOrigins, _, err := fitPopulationQueryBranch(input, -1, false)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	r.ActualFits++
	for _, origin := range choices {
		b := populationQueryBranch{Origin: origin, Q: .5, Origins: baseOrigins}
		if origin >= 0 {
			s := input.Steps[origin]
			b.ActualY = s.Y
			b.Q = s.Q
			b.Redundant = !s.Missing && origin+s.Delay <= 161
		}
		for y := 0; y < 2; y++ {
			m := base
			if origin >= 0 && !b.Redundant {
				m, b.Origins, _, err = fitPopulationQueryBranch(input, origin, y == 1)
				if err != nil {
					r.Error = err.Error()
					return r
				}
				r.ActualFits++
			}
			b.Population[y], err = k.risk(m.predictions)
			if err != nil {
				r.Error = err.Error()
				return r
			}
			b.LogEvidence[y] = m.logEvidence
			for i := 0; i < 31; i++ {
				s := input.Steps[161+i]
				p := m.predictions[s.X]
				b.AtPublication[y][i] = p
				p = .5 + math.Pow(.99, float64(i))*(p-.5)
				b.Predictions[y][i] = p
				b.Sample[y] += ((p-s.Q)*(p-s.Q) + s.Q*(1-s.Q)) / 31
			}
		}
		r.Branches = append(r.Branches, b)
	}
	return r
}
