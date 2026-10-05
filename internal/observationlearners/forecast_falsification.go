package observationlearners

import (
	"errors"
	"math"
)

type forecastTest struct {
	Count     int
	LogActive float64
	Rejected  bool
}

func falsificationLogAdd(a, b float64) float64 {
	if math.IsInf(a, -1) {
		return b
	}
	if math.IsInf(b, -1) {
		return a
	}
	return math.Max(a, b) + math.Log1p(math.Exp(-math.Abs(a-b)))
}

// Equal mass is assigned in advance to all32 possible starts. Dormant starts
// retain value1; the recurrence sums active likelihood ratios without32 loops.
func (s *forecastTest) update(p float64, y bool) float64 {
	logProbability := math.Log1p(-p)
	if y {
		logProbability = math.Log(p)
	}
	previous := s.LogActive
	if s.Count == 0 {
		previous = math.Inf(-1)
	}
	s.LogActive = -math.Ln2 - logProbability + falsificationLogAdd(previous, 0)
	s.Count++
	dormant := math.Inf(-1)
	if s.Count < 32 {
		dormant = math.Log(float64(32 - s.Count))
	}
	logEvidence := falsificationLogAdd(s.LogActive, dormant) - math.Log(32)
	if logEvidence >= math.Log(6400) {
		s.Rejected = true
	}
	return logEvidence
}

type falsificationForecast struct {
	Sequence uint64
	Raw      [4]float64
	Accepted [4]bool
	P        float64
	Neutral  bool
}

// forecastFalsification is a single-owner256-step research policy. Eight fixed
// 32-step publications and two views x four experts spend the declared0.01
// family error budget. Unrejected versions are NOT certified true.
type forecastFalsification struct {
	issued  uint64
	active  bool
	tests   [4]forecastTest
	pending falsificationForecast
}

func (s *forecastFalsification) predict(sequence uint64, p, w [4]float64) (falsificationForecast, error) {
	if s == nil || s.active || sequence != s.issued || sequence >= 256 {
		return falsificationForecast{}, errors.New("invalid falsification issuance")
	}
	total := 0.
	for i := range p {
		// This research contract requires full support; no silent probability floor
		// is introduced into the likelihood ratio or its conditional-law null.
		if math.IsNaN(p[i]) || math.IsInf(p[i], 0) || p[i] <= 0 || p[i] >= 1 || math.IsNaN(w[i]) || math.IsInf(w[i], 0) || w[i] < 0 || w[i] > 1 {
			return falsificationForecast{}, errors.New("invalid falsification probability or weight")
		}
		total += w[i]
	}
	if math.Abs(total-1) > 1e-12 {
		return falsificationForecast{}, errors.New("invalid falsification weights")
	}
	if sequence%32 == 0 {
		s.tests = [4]forecastTest{}
	}
	f := falsificationForecast{Sequence: sequence, Raw: p}
	mass := 0.
	for i := range p {
		f.Accepted[i] = !s.tests[i].Rejected
		if f.Accepted[i] {
			mass += w[i]
			f.P += w[i] * p[i]
		}
	}
	if mass == 0 {
		f.P = .5
		f.Neutral = true
	} else {
		f.P = math.Max(0, math.Min(1, f.P/mass))
	}
	s.pending = f
	s.active = true
	s.issued++
	return f, nil
}

func (s *forecastFalsification) observe(sequence uint64, y bool) error {
	if s == nil || !s.active || sequence != s.pending.Sequence {
		return errors.New("invalid falsification feedback")
	}
	for i, p := range s.pending.Raw {
		s.tests[i].update(p, y)
	}
	s.pending = falsificationForecast{}
	s.active = false
	return nil
}
