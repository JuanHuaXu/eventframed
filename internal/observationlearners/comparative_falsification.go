package observationlearners

import (
	"errors"
	"math"
)

type comparativeTest struct {
	Count     int
	LogActive [4]float64
	Rejected  bool
}

// Every alternative is a predictable probability law, not an independent vote.
// Each active sum mixes all32 predeclared starts; dormant starts remain at1.
func (s *comparativeTest) update(p float64, q [4]float64, y bool) float64 {
	logP := math.Log1p(-p)
	if y {
		logP = math.Log(p)
	}
	count := s.Count + 1
	dormant := math.Inf(-1)
	if count < 32 {
		dormant = math.Log(float64(32 - count))
	}
	combined := math.Inf(-1)
	for j, alt := range q {
		previous := s.LogActive[j]
		if s.Count == 0 {
			previous = math.Inf(-1)
		}
		logQ := math.Log1p(-alt)
		if y {
			logQ = math.Log(alt)
		}
		s.LogActive[j] = logQ - logP + falsificationLogAdd(previous, 0)
		logMass := -math.Log(6)
		if j == 0 {
			logMass = -math.Ln2
		}
		term := logMass + falsificationLogAdd(s.LogActive[j], dormant) - math.Log(32)
		combined = falsificationLogAdd(combined, term)
	}
	s.Count = count
	if combined >= math.Log(6400) {
		s.Rejected = true
	}
	return combined
}

type comparativeForecast struct {
	Sequence uint64
	Raw      [4]float64
	Accepted [4]bool
	P        float64
	Neutral  bool
}

// comparativeFalsification is a single-owner256-step research policy. Eight fixed
// 32-step publications and two views x four experts spend the declared0.01
// family error budget. Unrejected versions are NOT certified true.
type comparativeFalsification struct {
	issued  uint64
	active  bool
	tests   [4]comparativeTest
	pending comparativeForecast
}

func (s *comparativeFalsification) predict(sequence uint64, p, w [4]float64) (comparativeForecast, error) {
	if s == nil || s.active || sequence != s.issued || sequence >= 256 {
		return comparativeForecast{}, errors.New("invalid falsification issuance")
	}
	total := 0.
	for i := range p {
		// This research contract requires full support; no silent probability floor
		// is introduced into the likelihood ratio or its conditional-law null.
		if math.IsNaN(p[i]) || math.IsInf(p[i], 0) || p[i] <= 0 || p[i] >= 1 || math.IsNaN(w[i]) || math.IsInf(w[i], 0) || w[i] < 0 || w[i] > 1 {
			return comparativeForecast{}, errors.New("invalid falsification probability or weight")
		}
		total += w[i]
	}
	if math.Abs(total-1) > 1e-12 {
		return comparativeForecast{}, errors.New("invalid falsification weights")
	}
	if sequence%32 == 0 {
		s.tests = [4]comparativeTest{}
	}
	f := comparativeForecast{Sequence: sequence, Raw: p}
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

func (s *comparativeFalsification) observe(sequence uint64, y bool) error {
	if s == nil || !s.active || sequence != s.pending.Sequence {
		return errors.New("invalid falsification feedback")
	}
	for i, p := range s.pending.Raw {
		q := [4]float64{.5}
		j := 1
		for k, v := range s.pending.Raw {
			if k != i {
				q[j] = v
				j++
			}
		}
		s.tests[i].update(p, q, y)
	}
	s.pending = comparativeForecast{}
	s.active = false
	return nil
}
