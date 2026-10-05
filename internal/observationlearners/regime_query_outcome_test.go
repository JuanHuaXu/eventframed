package observationlearners

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math"
)

type regimeOutcomeDecision struct {
	Clock           int
	Origins, Pool   []int
	Probes          []uint16
	Values          []regimeQueryValue
	BaseLogEvidence float64 `json:",omitempty"`
	Selected, Costs [4]int
}

func chooseRegimeMaximum(pool []int, scores []float64) (int, error) {
	if len(pool) == 0 || len(pool) != len(scores) {
		return -1, fmt.Errorf("invalid selector input")
	}
	best := math.Inf(-1)
	for _, s := range scores {
		if math.IsNaN(s) || math.IsInf(s, 0) {
			return -1, fmt.Errorf("invalid selector score")
		}
		best = math.Max(best, s)
	}
	winner := int(^uint(0) >> 1)
	for i, s := range scores {
		if best-s <= 1e-10 && pool[i] < winner {
			winner = pool[i]
		}
	}
	return winner, nil
}

// This phase cannot read a paid answer. It ends with explicit selections that
// are passed into the separate reveal/publication phase below.
func decideRegimeOutcome(input softV120Record) (regimeOutcomeDecision, error) {
	d := regimeOutcomeDecision{Clock: 160, Selected: [4]int{-1, -1, -1, -1}}
	h, pool, probes, err := regimeQueryTapeView(input, d.Clock)
	if err != nil {
		return d, err
	}
	d.Pool, d.Probes = pool, probes
	for i, p := range h {
		if p.Arrives >= 0 {
			d.Origins = append(d.Origins, i-16)
		}
	}
	if len(pool) == 0 {
		return d, nil
	}
	batch, err := fitRegimeQueryBatch(context.Background(), -16, d.Clock, h, pool, probes)
	if err != nil {
		return d, err
	}
	d.Values = batch.Values
	d.BaseLogEvidence = batch.BaseLogEvidence
	var lowest [32]byte
	for i, j := range pool {
		hash := sha256.Sum256([]byte(fmt.Sprintf("regime-query-outcome-v1:%d:%d:%d:%d:%d", input.Phase, input.Case, input.Index, d.Clock, j)))
		if i == 0 || bytes.Compare(hash[:], lowest[:]) < 0 || (hash == lowest && j < d.Selected[1]) {
			lowest = hash
			d.Selected[1] = j
		}
	}
	entropies, gains := make([]float64, len(pool)), make([]float64, len(pool))
	for i, v := range d.Values {
		p := v.Mass[1]
		entropies[i] = -p*math.Log(p) - (1-p)*math.Log1p(-p)
		gains[i] = v.Gain
	}
	d.Selected[2], err = chooseRegimeMaximum(pool, entropies)
	if err != nil {
		return d, err
	}
	d.Selected[3], err = chooseRegimeMaximum(pool, gains)
	if err != nil {
		return d, err
	}
	d.Costs = [4]int{0, 1, 1, 1}
	return d, nil
}

type regimeOutcomeRecord struct {
	Phase, Case, Index, Schedule int
	Decision                     regimeOutcomeDecision
	Origins                      [4][]int
	Redundant                    [4]bool
	AtPublication, Predictions   [4][31]float64
	LogEvidence                  [4]float64
	ActualFits                   int
	Error                        string `json:",omitempty"`
}

func regimeOutcomePublication(input softV120Record, selected int) ([]segmentPacket, []int, bool, error) {
	if len(input.Steps) != 256 || selected < -1 || selected >= 160 || (selected >= 0 && selected < 152) {
		return nil, nil, false, fmt.Errorf("invalid publication selection")
	}
	if selected >= 0 {
		s := input.Steps[selected]
		if !s.Missing && selected+s.Delay <= 160 {
			return nil, nil, false, fmt.Errorf("selected already-known label")
		}
	}
	var available []int
	for j := -16; j < 161; j++ {
		if j < 0 || j == selected || (!input.Steps[j].Missing && j+input.Steps[j].Delay <= 161) {
			available = append(available, j)
		}
	}
	available = available[max(0, len(available)-64):]
	history := make([]segmentPacket, 177)
	for i := range history {
		j := i - 16
		history[i].Arrives = -1
		if j < 0 {
			history[i].Bits = input.Initial[j+16].Bits
		} else {
			history[i].Bits = input.Steps[j].X
		}
	}
	for _, j := range available {
		history[j+16].Arrives = 161
		if j < 0 {
			history[j+16].Outcome = input.Initial[j+16].Outcome
		} else {
			history[j+16].Outcome = input.Steps[j].Y
		}
	}
	redundant := false
	if selected >= 0 {
		s := input.Steps[selected]
		redundant = !s.Missing && selected+s.Delay <= 161
	}
	return history, available, redundant, nil
}

// Publish after reveal. Other natural arrivals and the64-label cap can change
// support, so this is measured practical value, not the hypothetical Bayes law.
func publishRegimeOutcome(input softV120Record, decision regimeOutcomeDecision) regimeOutcomeRecord {
	r := regimeOutcomeRecord{Phase: input.Phase, Case: input.Case, Index: input.Index, Schedule: input.Schedule, Decision: decision}
	if decision.Clock != 160 {
		r.Error = "invalid decision clock"
		return r
	}
	type supportKey struct {
		N       int
		Origins [64]int
	}
	cache := map[supportKey]*segmentPosterior{}
	for arm, selected := range decision.Selected {
		h, origins, redundant, err := regimeOutcomePublication(input, selected)
		if err != nil {
			r.Error = err.Error()
			return r
		}
		r.Origins[arm], r.Redundant[arm] = origins, redundant
		key := supportKey{N: len(origins)}
		copy(key.Origins[:], origins)
		m := cache[key]
		if m == nil {
			m, err = fitSegmentPosteriorBatch(context.Background(), -16, 161, h, 64, .01, .95)
			if err != nil {
				r.Error = err.Error()
				return r
			}
			cache[key] = m
			r.ActualFits++
		}
		r.LogEvidence[arm] = m.logEvidence
		for i := 0; i < 31; i++ {
			x := input.Steps[161+i].X
			if x >= 512 {
				r.Error = "invalid future query input"
				return r
			}
			p := m.predictions[x]
			r.AtPublication[arm][i] = p
			r.Predictions[arm][i] = .5 + math.Pow(.99, float64(i))*(p-.5)
		}
	}
	return r
}

func runRegimeOutcome(input softV120Record) regimeOutcomeRecord {
	d, err := decideRegimeOutcome(input)
	if err != nil {
		return regimeOutcomeRecord{Phase: input.Phase, Case: input.Case, Index: input.Index, Schedule: input.Schedule, Error: err.Error()}
	}
	return publishRegimeOutcome(input, d)
}
