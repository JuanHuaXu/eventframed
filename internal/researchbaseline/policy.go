// Package researchbaseline declares prospective research references, not daemon defaults.
package researchbaseline

import (
	"errors"
	"math"
)

const (
	PrimaryArm    = "adaptive"
	SpeedControl  = "full"
	CoreBudgetMS  = 400.
	CompleteCells = 120
)

type Candidate struct {
	Name          string  `json:"name"`
	Cohort        string  `json:"cohort"`
	Cells         int     `json:"cells"`
	ExpectedBrier float64 `json:"expected_brier"`
	MeanCoreMS    float64 `json:"mean_core_ms"`
	WorstCoreMS   float64 `json:"worst_core_ms"`
	Comparable    bool    `json:"comparable"`
}

// Select minimizes recorded loss within the original complete-core budget.
// Component benchmarks and unlike cohorts cannot enter this comparison.
func Select(candidates []Candidate, cohort string) (Candidate, error) {
	if cohort == "" {
		return Candidate{}, errors.New("explicit cohort required")
	}
	var best Candidate
	found := false
	for _, c := range candidates {
		if !c.Comparable || c.Cohort != cohort {
			continue
		}
		if c.Name == "" || c.Cells != CompleteCells || math.IsNaN(c.ExpectedBrier) ||
			math.IsInf(c.ExpectedBrier, 0) || c.ExpectedBrier < 0 || c.ExpectedBrier > 1 ||
			math.IsNaN(c.MeanCoreMS) || math.IsInf(c.MeanCoreMS, 0) || c.MeanCoreMS <= 0 ||
			math.IsNaN(c.WorstCoreMS) || math.IsInf(c.WorstCoreMS, 0) || c.WorstCoreMS < c.MeanCoreMS {
			return Candidate{}, errors.New("invalid comparable evidence")
		}
		if c.WorstCoreMS > CoreBudgetMS {
			continue
		}
		if !found || c.ExpectedBrier < best.ExpectedBrier ||
			(c.ExpectedBrier == best.ExpectedBrier && c.MeanCoreMS < best.MeanCoreMS) {
			best, found = c, true
		}
	}
	if !found {
		return Candidate{}, errors.New("no complete budget-eligible baseline")
	}
	return best, nil
}
