package observationlearners

import (
	"errors"
	"math"
)

// Research-only log-score update. The journal supplies immutable issue-time
// probabilities and owns identity/censoring. With delayed arrival-order share
// transitions this is advice weighting, not event-clock Bayesian filtering.
func observeLogAdvice(s *agedAdvice, origin uint64, raw [4]float64, y bool, share float64) error {
	if s == nil || !s.ready || s.aged || origin > s.clock || origin >= 256 || math.IsNaN(share) || share < 0 || share >= 1 {
		return errors.New("invalid log-advice contract")
	}
	for _, p := range raw {
		if math.IsNaN(p) || p <= 0 || p >= 1 {
			return errors.New("log advice requires full-support forecasts")
		}
	}
	copy := *s
	row := [5]float64{.5, raw[0], raw[1], raw[2], raw[3]}
	maximum := math.Inf(-1)
	for j, p := range row {
		ll := math.Log1p(-p)
		if y {
			ll = math.Log(p)
		}
		copy.logs[j] += ll
		maximum = math.Max(maximum, copy.logs[j])
	}
	for j := range copy.logs {
		copy.logs[j] -= maximum
	}
	// At zero share retain log weights: converting an underflowed probability
	// back through log(0) would permanently erase recoverable expert evidence.
	if share > 0 {
		w := copy.weights()
		for j := range copy.logs {
			copy.logs[j] = math.Log((1-share)*w[j] + share*copy.prior[j])
		}
	}
	*s = copy
	return nil
}

// Reuse the verified acquisition and version-scoped gate machinery without
// changing the frozen Brier/age kernels. Only selector delivery is replaced.
type logAdviceJournal struct{ *agedAdviceJournal }

func newLogAdviceJournal(neutral bool) *logAdviceJournal {
	return &logAdviceJournal{newAgedAdviceJournal(neutral, false)}
}

func (g *logAdviceJournal) deliver(origin uint64, y bool) error {
	if g == nil || g.agedAdviceJournal == nil || g.journal.busy {
		return errors.New("invalid log-advice journal delivery")
	}
	copy := *g.agedAdviceJournal
	entry := copy.journal.entries[origin%routedJournalCapacity]
	if err := copy.journal.deliver(origin, y); err != nil {
		return err
	}
	if err := observeLogAdvice(&copy.advice, origin, entry.bank.Experts, y, .001); err != nil {
		return err
	}
	copy.received++
	*g.agedAdviceJournal = copy
	return nil
}
