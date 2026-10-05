package observationlearners

import (
	"errors"
	"math"
)

// Research-only five-role advice selector. Neutrality is an ordinary competing
// forecast, not evidence that any raw model is false. Discounting uses origin
// age, never arrival age. Neither mode claims an ordinary Bayesian posterior.
type agedAdvice struct {
	ready, aged         bool
	clock               uint64
	prior, logs, losses [5]float64 // neutral, generic64, Boolean64, generic32, Boolean32
}

func newAgedAdvice(neutral, aged bool) agedAdvice {
	s := agedAdvice{ready: true, aged: aged}
	mass := 0.
	if neutral {
		mass = .05
	}
	s.prior = [5]float64{mass, (1 - mass) * .95, (1 - mass) * .05 / 3, (1 - mass) * .05 / 3, (1 - mass) * .05 / 3}
	for j, p := range s.prior {
		s.logs[j] = math.Log(p)
	}
	return s
}

func (s *agedAdvice) advance(clock uint64) error {
	if s == nil || !s.ready || clock < s.clock || clock > 288 {
		return errors.New("invalid advice clock")
	}
	if s.aged {
		factor := math.Exp2(-float64(clock-s.clock) / 32)
		for j := range s.losses {
			s.losses[j] *= factor
		}
	}
	s.clock = clock
	return nil
}

func (s *agedAdvice) weights() [5]float64 {
	logs := s.logs
	if s.aged {
		for j, p := range s.prior {
			logs[j] = math.Log(p) - .5*s.losses[j]
		}
	}
	var w [5]float64
	maximum := math.Inf(-1)
	for _, v := range logs {
		maximum = math.Max(maximum, v)
	}
	sum := 0.
	for j, v := range logs {
		w[j] = math.Exp(v - maximum)
		sum += w[j]
	}
	for j := range w {
		w[j] /= sum
	}
	return w
}

// The journal owns packet identity, single-use and censoring. This helper only
// consumes validated immutable issue-time forecasts, at the current event clock.
func (s *agedAdvice) observe(origin uint64, raw [4]float64, y bool) error {
	if s == nil || !s.ready || origin > s.clock || origin >= 256 {
		return errors.New("invalid advice origin")
	}
	for _, p := range raw {
		if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
			return errors.New("invalid advice probability")
		}
	}
	target := 0.
	if y {
		target = 1
	}
	p := [5]float64{.5, raw[0], raw[1], raw[2], raw[3]}
	if s.aged {
		factor := math.Exp2(-float64(s.clock-origin) / 32)
		for j, x := range p {
			s.losses[j] += factor * (x - target) * (x - target)
		}
	} else {
		maximum := math.Inf(-1)
		for j, x := range p {
			s.logs[j] -= .5 * (x - target) * (x - target)
			maximum = math.Max(maximum, s.logs[j])
		}
		for j := range s.logs {
			s.logs[j] -= maximum
		}
		w := s.weights()
		for j := range s.logs {
			s.logs[j] = math.Log(.999*w[j] + .001*s.prior[j])
		}
	}
	return nil
}
