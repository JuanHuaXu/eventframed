package observationlearners

import (
	"context"
	"fmt"
	"math"
	"reflect"
)

type regimeDisjointRecord struct {
	Original, Candidate regimeOutcomeRecord
	ExtraBatches        int
	Error               string `json:",omitempty"`
}

// Retain the original decision before replacing only its virtual target probes.
// Query-label marginal probabilities must not change with this replacement.
func disjointRegimeDecision(input softV120Record, original regimeOutcomeDecision) (regimeOutcomeDecision, error) {
	d := original
	d.Probes = make([]uint16, 8)
	if len(input.Steps) != 256 || d.Clock != 160 {
		return d, fmt.Errorf("invalid disjoint input")
	}
	for i := range d.Probes {
		d.Probes[i] = input.Steps[137+i].X
		if d.Probes[i] >= 512 {
			return d, fmt.Errorf("invalid disjoint probe")
		}
	}
	if len(d.Pool) == 0 {
		return d, nil
	}
	h, pool, _, err := regimeQueryTapeView(input, 160)
	if err != nil {
		return d, err
	}
	if !reflect.DeepEqual(pool, d.Pool) {
		return d, fmt.Errorf("changed candidate pool")
	}
	batch, err := fitRegimeQueryBatch(context.Background(), -16, 160, h, pool, d.Probes)
	if err != nil {
		return d, err
	}
	if len(batch.Values) != len(original.Values) {
		return d, fmt.Errorf("changed candidate count")
	}
	entropy, gains := make([]float64, len(pool)), make([]float64, len(pool))
	for i, v := range batch.Values {
		for y := 0; y < 2; y++ {
			if math.Abs(v.Mass[y]-original.Values[i].Mass[y]) > 1e-10 || math.Abs(v.LogEvidence[y]-original.Values[i].LogEvidence[y]) > 1e-10 {
				return d, fmt.Errorf("probe-dependent marginal")
			}
		}
		p := v.Mass[1]
		entropy[i] = -p*math.Log(p) - (1-p)*math.Log1p(-p)
		gains[i] = v.Gain
	}
	unchanged, err := chooseRegimeMaximum(pool, entropy)
	if err != nil {
		return d, err
	}
	if unchanged != original.Selected[2] {
		return d, fmt.Errorf("probe-dependent entropy selection")
	}
	d.Selected[3], err = chooseRegimeMaximum(pool, gains)
	if err != nil {
		return d, err
	}
	d.Values = batch.Values
	d.BaseLogEvidence = batch.BaseLogEvidence
	return d, nil
}

func runRegimeDisjoint(input softV120Record) regimeDisjointRecord {
	r := regimeDisjointRecord{Original: runRegimeOutcome(input)}
	if r.Original.Error != "" {
		r.Error = r.Original.Error
		return r
	}
	d, err := disjointRegimeDecision(input, r.Original.Decision)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	r.Candidate = publishRegimeOutcome(input, d)
	if r.Candidate.Error != "" {
		r.Error = r.Candidate.Error
		return r
	}
	if len(d.Pool) > 0 {
		r.ExtraBatches = 1
	}
	for arm := 0; arm < 3; arm++ {
		if r.Candidate.Decision.Selected[arm] != r.Original.Decision.Selected[arm] || r.Candidate.Decision.Costs[arm] != r.Original.Decision.Costs[arm] || !reflect.DeepEqual(r.Candidate.Origins[arm], r.Original.Origins[arm]) || r.Candidate.Predictions[arm] != r.Original.Predictions[arm] || r.Candidate.AtPublication[arm] != r.Original.AtPublication[arm] || r.Candidate.LogEvidence[arm] != r.Original.LogEvidence[arm] {
			r.Error = "baseline changed"
			return r
		}
	}
	return r
}
