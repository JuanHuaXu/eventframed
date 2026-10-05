package researchstats

import "errors"

// BrierPair contains two probabilities forecast before the same Bernoulli
// outcome. Outcome's bool type excludes nonbinary labels, not label leakage.
type BrierPair struct {
	Control, Candidate float64
	Outcome            bool
}

// PairedBrierMean averages (Control-Y)^2-(Candidate-Y)^2 over one nonempty,
// prespecified stream window, including warm-up/adaptation if so specified.
// It weights clocks equally within this stream. Independent-clock assumptions
// are unnecessary for bounding the stream mean, but this is not a within-stream
// CS. Missing outcomes must follow a separately justified prespecified policy.
func PairedBrierMean(pairs []BrierPair) (float64, error) {
	if len(pairs) == 0 || uint64(len(pairs)) > maxUnits {
		return 0, errors.New("empty or oversized paired stream")
	}
	var state meanState
	for _, pair := range pairs {
		if !finite(pair.Control) || !finite(pair.Candidate) ||
			pair.Control < 0 || pair.Control > 1 || pair.Candidate < 0 || pair.Candidate > 1 {
			return 0, errors.New("Brier probabilities must be finite and in [0,1]")
		}
		y := 0.
		if pair.Outcome {
			y = 1
		}
		control, candidate := pair.Control-y, pair.Candidate-y
		state.add(control*control - candidate*candidate)
	}
	return state.mean(), nil
}

// AddPairedStream validates the complete stream before adding its single mean.
// Different stream lengths still receive equal stream-level weight.
func (s *ConfidenceSequence) AddPairedStream(comparison string, pairs []BrierPair) error {
	mean, err := PairedBrierMean(pairs)
	if err != nil {
		return err
	}
	return s.AddStream(comparison, mean)
}
