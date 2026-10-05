package observationlearners

import (
	"errors"
	"math"
)

type intervalBrierSlot struct {
	start        uint64
	weight, rate float64
	base         brierAggregator
	pending      brierForecast
}

// intervalBrier is a finite-horizon research specialization of SAOL.
// A single owner must supply every outcome before requesting another forecast.
// Ten slots cover the geometric intervals intersecting rounds1..512; this is
// not an unbounded daemon or the v94 lifetime-regret guarantee.
type intervalBrier struct {
	horizon, issued uint64
	active          bool
	slots           [10]intervalBrierSlot
	count           int
	mix             float64
}

func newIntervalBrier(horizon uint64) (*intervalBrier, error) {
	if horizon == 0 || horizon > 512 {
		return nil, errors.New("invalid interval horizon")
	}
	return &intervalBrier{horizon: horizon}, nil
}

func (s *intervalBrier) predict(sequence uint64, experts [2]float64) (float64, error) {
	if s == nil || s.horizon == 0 || s.active || sequence != s.issued || s.issued >= s.horizon {
		return 0, errors.New("invalid interval issuance")
	}
	for _, p := range experts {
		if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
			return 0, errors.New("invalid interval probability")
		}
	}
	t := sequence + 1
	total, mix := 0., 0.
	s.count = 0
	for level := range s.slots {
		length := uint64(1) << level
		if length > t {
			break
		}
		start := (t / length) * length
		slot := &s.slots[level]
		if slot.start != start {
			base, _ := newBrierAggregator(0)
			rate := math.Min(.5, 1/math.Sqrt(float64(length)))
			*slot = intervalBrierSlot{start: start, weight: rate, rate: rate, base: *base}
		}
		// The private state guarantees the base issuance contract; callers cannot
		// mutate an interval's sequence or substitute a different forecast later.
		f, e := slot.base.predict(t-start, experts)
		if e != nil {
			return 0, e
		}
		slot.pending = f
		total += slot.weight
		mix += slot.weight * f.P
		s.count++
	}
	s.mix = math.Max(0, math.Min(1, mix/total))
	s.active = true
	s.issued++
	return s.mix, nil
}

func (s *intervalBrier) observe(sequence uint64, outcome bool) error {
	if s == nil || !s.active || s.issued == 0 || sequence != s.issued-1 {
		return errors.New("invalid interval feedback")
	}
	y := 0.
	if outcome {
		y = 1
	}
	total, surrogate := 0., 0.
	var loss [10]float64
	for i := 0; i < s.count; i++ {
		slot := &s.slots[i]
		d := slot.pending.P - y
		loss[i] = d * d
		total += slot.weight
		surrogate += slot.weight * loss[i]
	}
	surrogate /= total
	// Use the weighted expert loss in the SAOL meta-update, not the mixed
	// forecast's smaller squared loss. Convexity connects the two evaluations.
	for i := 0; i < s.count; i++ {
		slot := &s.slots[i]
		slot.weight *= 1 + slot.rate*(surrogate-loss[i])
		if e := slot.base.observe(slot.pending.Sequence, outcome); e != nil {
			return e
		}
	}
	s.active = false
	return nil
}
