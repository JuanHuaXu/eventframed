package observationlearners

import "fmt"

// These are separately fitted counterfactual actions, not one posterior that
// receives all purchased labels. Only the offline scorer may rank their value.
type regimeEnvelopeRecord struct {
	Phase, Case, Index, Schedule int
	Choices                      []int
	Bundles                      []regimeOutcomeRecord
	ActualFits                   int
	Error                        string `json:",omitempty"`
}

func regimeEnvelopeChoices(input softV120Record) ([]int, error) {
	if len(input.Steps) != 256 {
		return nil, fmt.Errorf("invalid source length")
	}
	choices := []int{-1}
	for j := 152; j < 160; j++ {
		s := input.Steps[j]
		if s.Missing || j+s.Delay > 160 {
			choices = append(choices, j)
		}
	}
	return choices, nil
}

func runRegimeEnvelope(input softV120Record) regimeEnvelopeRecord {
	r := regimeEnvelopeRecord{Phase: input.Phase, Case: input.Case, Index: input.Index, Schedule: input.Schedule}
	choices, err := regimeEnvelopeChoices(input)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	r.Choices = choices
	for start := 0; start < len(choices); start += 4 {
		d := regimeOutcomeDecision{Clock: 160}
		for arm := range d.Selected {
			d.Selected[arm] = choices[min(start+arm, len(choices)-1)]
			if d.Selected[arm] >= 0 {
				d.Costs[arm] = 1
			}
		}
		bundle := publishRegimeOutcome(input, d)
		r.Bundles = append(r.Bundles, bundle)
		r.ActualFits += bundle.ActualFits
		if bundle.Error != "" {
			r.Error = bundle.Error
			return r
		}
	}
	return r
}
