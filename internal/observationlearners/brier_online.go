package observationlearners

import (
	"errors"
	"math"
)

const brierRate = .5
const brierGenericPrior = .95

type brierForecast struct {
	Sequence         uint64
	Experts, Weights [2]float64
	P                float64
}

// brierAggregator is a single-owner immediate-feedback research state machine.
// Its weights describe expert performance, not posterior probabilities of truth.
type brierAggregator struct {
	initialized, active bool
	share               float64
	logWeights          [2]float64
	issued              uint64
	pending             brierForecast
}

func newBrierAggregator(share float64) (*brierAggregator, error) {
	if math.IsNaN(share) || math.IsInf(share, 0) || share < 0 || share >= 1 {
		return nil, errors.New("invalid Brier sharing rate")
	}
	return &brierAggregator{initialized: true, share: share, logWeights: [2]float64{math.Log(brierGenericPrior), math.Log1p(-brierGenericPrior)}}, nil
}

func (s *brierAggregator) weights() [2]float64 {
	maximum := math.Max(s.logWeights[0], s.logWeights[1])
	w := [2]float64{math.Exp(s.logWeights[0] - maximum), math.Exp(s.logWeights[1] - maximum)}
	total := w[0] + w[1]
	w[0] /= total
	w[1] /= total
	return w
}

func (s *brierAggregator) predict(sequence uint64, experts [2]float64) (brierForecast, error) {
	if s == nil || !s.initialized || s.active || sequence != s.issued || s.issued == math.MaxUint64 {
		return brierForecast{}, errors.New("invalid Brier issuance")
	}
	for _, p := range experts {
		if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
			return brierForecast{}, errors.New("invalid Brier probability")
		}
	}
	w := s.weights()
	f := brierForecast{Sequence: sequence, Experts: experts, Weights: w, P: math.Max(0, math.Min(1, w[0]*experts[0]+w[1]*experts[1]))}
	s.pending = f
	s.active = true
	s.issued++
	return f, nil
}

func (s *brierAggregator) observe(sequence uint64, outcome bool) error {
	if s == nil || !s.initialized || !s.active || sequence != s.pending.Sequence {
		return errors.New("invalid Brier feedback")
	}
	y := 0.
	if outcome {
		y = 1
	}
	for k, p := range s.pending.Experts {
		d := p - y
		s.logWeights[k] -= brierRate * d * d
	}
	maximum := math.Max(s.logWeights[0], s.logWeights[1])
	for k := range s.logWeights {
		s.logWeights[k] -= maximum
	}
	if s.share > 0 {
		w := s.weights()
		prior := [2]float64{brierGenericPrior, 1 - brierGenericPrior}
		for k := range w {
			s.logWeights[k] = math.Log((1-s.share)*w[k] + s.share*prior[k])
		}
	}
	s.active = false
	s.pending = brierForecast{}
	return nil
}

// brierGenericBound includes the no-switch path penalty of fixed sharing.
// It does not cover omitted feedback, state resets or a different comparator.
func brierGenericBound(n uint64, share float64) float64 {
	if n == 0 {
		return 0
	}
	return (-math.Log(brierGenericPrior) - float64(n-1)*math.Log1p(-share)) / brierRate
}
