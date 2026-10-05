package observationlearners

import (
	"errors"
	"math"
)

type brierBankForecast struct {
	Sequence         uint64
	Experts, Weights [4]float64
	P                float64
}

// brierBank generalizes the verified two-expert research kernel, with a hard
// four-expert cap. Forecasts are immutable values; each outcome is required
// exactly once before the next forecast. This is not a delayed-feedback API.
type brierBank struct {
	count       int
	active      bool
	issued      uint64
	share       float64
	prior, logs [4]float64
	pending     brierBankForecast
}

func newBrierBank(prior []float64, share float64) (*brierBank, error) {
	if len(prior) < 1 || len(prior) > 4 || math.IsNaN(share) || math.IsInf(share, 0) || share < 0 || share >= 1 {
		return nil, errors.New("invalid Brier bank contract")
	}
	sum := 0.
	for _, p := range prior {
		if math.IsNaN(p) || math.IsInf(p, 0) || p <= 0 || p > 1 {
			return nil, errors.New("invalid Brier bank prior")
		}
		sum += p
	}
	if math.Abs(sum-1) > 1e-12 {
		return nil, errors.New("Brier prior does not sum to one")
	}
	s := &brierBank{count: len(prior), share: share}
	for i, p := range prior {
		s.prior[i] = p / sum
		s.logs[i] = math.Log(s.prior[i])
	}
	return s, nil
}

func (s *brierBank) weights() [4]float64 {
	var w [4]float64
	maximum := s.logs[0]
	for i := 1; i < s.count; i++ {
		maximum = math.Max(maximum, s.logs[i])
	}
	sum := 0.
	for i := 0; i < s.count; i++ {
		w[i] = math.Exp(s.logs[i] - maximum)
		sum += w[i]
	}
	for i := 0; i < s.count; i++ {
		w[i] /= sum
	}
	return w
}

func (s *brierBank) predict(sequence uint64, experts []float64) (brierBankForecast, error) {
	if s == nil || s.count < 1 || s.active || sequence != s.issued || s.issued == math.MaxUint64 || len(experts) != s.count {
		return brierBankForecast{}, errors.New("invalid Brier bank issuance")
	}
	for _, p := range experts {
		if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
			return brierBankForecast{}, errors.New("invalid Brier bank forecast")
		}
	}
	f := brierBankForecast{Sequence: sequence, Weights: s.weights()}
	for i, p := range experts {
		f.Experts[i] = p
		f.P += f.Weights[i] * p
	}
	f.P = math.Max(0, math.Min(1, f.P))
	s.pending = f
	s.active = true
	s.issued++
	return f, nil
}

func (s *brierBank) observe(sequence uint64, outcome bool) error {
	if s == nil || !s.active || sequence != s.pending.Sequence {
		return errors.New("invalid Brier bank feedback")
	}
	y := 0.
	if outcome {
		y = 1
	}
	maximum := math.Inf(-1)
	for i := 0; i < s.count; i++ {
		d := s.pending.Experts[i] - y
		s.logs[i] -= .5 * d * d
		maximum = math.Max(maximum, s.logs[i])
	}
	for i := 0; i < s.count; i++ {
		s.logs[i] -= maximum
	}
	if s.share > 0 {
		w := s.weights()
		for i := 0; i < s.count; i++ {
			s.logs[i] = math.Log((1-s.share)*w[i] + s.share*s.prior[i])
		}
	}
	s.pending = brierBankForecast{}
	s.active = false
	return nil
}

// The comparator is expert0 under the actual normalized prior, not a constant
// copied from a predecessor with a different number or allocation of experts.
func (s *brierBank) genericBound(n uint64) float64 {
	if n == 0 {
		return 0
	}
	return (-math.Log(s.prior[0]) - float64(n-1)*math.Log1p(-s.share)) / .5
}
