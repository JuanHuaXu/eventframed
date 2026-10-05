package observationlearners

import (
	"errors"
	"math"
)

type routedForecast struct {
	Sequence uint64
	P        float64
	// Index0 is neutrality; indices1..4 map to raw experts0..3.
	Weights  [5]float64
	Accepted [4]bool
	Neutral  bool
}

// Evidence credits allocate forecast weight; they do not certify alternatives
// true. The comparative test budget and its single-owner lifecycle are unchanged.
type evidenceRouting struct {
	gate    comparativeFalsification
	credits [4][5]float64 // log credits frozen at the first crossing per version
}

func (s *evidenceRouting) predict(sequence uint64, p, w [4]float64) (routedForecast, error) {
	if s == nil {
		return routedForecast{}, errors.New("nil evidence routing")
	}
	f, err := s.gate.predict(sequence, p, w)
	if err != nil {
		return routedForecast{}, err
	}
	if sequence%32 == 0 {
		s.credits = [4][5]float64{}
	}
	r := routedForecast{Sequence: sequence, Accepted: f.Accepted}
	for i, accepted := range f.Accepted {
		if accepted {
			r.Weights[i+1] += w[i]
			continue
		}
		// Filter recipients before normalization. Never route recursively through
		// another rejected expert; neutral mass remains an eligible fallback.
		z := math.Inf(-1)
		valid := true
		for j, c := range s.credits[i] {
			if j > 0 && !f.Accepted[j-1] {
				continue
			}
			if math.IsNaN(c) || math.IsInf(c, 1) {
				valid = false
				break
			}
			z = falsificationLogAdd(z, c)
		}
		if !valid || math.IsInf(z, -1) {
			r.Weights[0] += w[i]
			continue
		}
		for j, c := range s.credits[i] {
			if j == 0 || f.Accepted[j-1] {
				r.Weights[j] += w[i] * math.Exp(c-z)
			}
		}
	}
	total := 0.
	for _, w := range r.Weights {
		total += w
	}
	for j := range r.Weights {
		r.Weights[j] /= total
	}
	r.P = .5 * r.Weights[0]
	r.Neutral = true
	for j, p := range p {
		r.P += r.Weights[j+1] * p
		r.Neutral = r.Neutral && r.Weights[j+1] == 0
	}
	r.P = math.Max(0, math.Min(1, r.P))
	return r, nil
}

func (s *evidenceRouting) observe(sequence uint64, y bool) error {
	if s == nil {
		return errors.New("nil evidence routing")
	}
	before := s.gate.tests
	if err := s.gate.observe(sequence, y); err != nil {
		return err
	}
	for i, test := range s.gate.tests {
		if !test.Rejected || before[i].Rejected {
			continue
		}
		for j := range s.credits[i] {
			s.credits[i][j] = math.Inf(-1)
		}
		dormant := math.Inf(-1)
		if test.Count < 32 {
			dormant = math.Log(float64(32 - test.Count))
		}
		z := math.Inf(-1)
		alt := 0
		for recipient := 0; recipient < 5; recipient++ {
			if recipient == i+1 {
				continue
			}
			mass := -math.Log(6)
			if recipient == 0 {
				mass = -math.Ln2
			}
			c := mass + falsificationLogAdd(test.LogActive[alt], dormant) - math.Log(32)
			s.credits[i][recipient] = c
			z = falsificationLogAdd(z, c)
			alt++
		}
		for j := range s.credits[i] {
			s.credits[i][j] -= z
		}
	}
	return nil
}
