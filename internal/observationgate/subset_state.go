package observationgate

import (
	"errors"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

// subsetState retains the count model while testing learned marginalization.
// It is an isolated research composition, with no serving publication rights.
type subsetState struct {
	base           *observation.Model
	enabled, split bool
	mix, inner     bayes.ForecastMix
	next           int
	pending        *subsetPending
}
type subsetPending struct {
	p                      observationpreserved.Prediction
	inner                  [4]float64
	innerWeights           [4]float64
	available, subsetGuide bool
}

func (s *subsetState) predict(reader observation.Reader, models observationpreserved.Models, subset *observationlearners.ConditionalForest, sequence int, seed int64) (observationpreserved.Prediction, error) {
	if s.base == nil || s.pending != nil || sequence != s.next {
		return observationpreserved.Prediction{}, errors.New("invalid subset prediction order")
	}
	long := models.Pooled
	if s.split {
		long = models.Local
	}
	ms := [3]*observation.Model{s.base, models.Short, long}
	w := s.mix.Weights
	if w == [4]float64{} {
		w = [4]float64{.7, .1, .1, .1}
	}
	iw := s.inner.Weights
	if iw == [4]float64{} {
		iw = [4]float64{.7, .1, .1, .1}
	}
	guide := 0
	for j := 1; j < 3; j++ {
		if ms[j] != nil && w[j] > w[guide] {
			guide = j
		}
	}
	available := s.enabled && subset != nil && models.Short != nil
	useSubset := available && guide == 1 && iw[1]+iw[2]+iw[3] > iw[0]
	var view observation.Result
	var err error
	if useSubset {
		view, err = observationlearners.RunConditionalObserver(subset, reader, reader.Epoch())
	} else {
		view, err = observation.Run(ms[guide], reader, reader.Epoch(), "mmm", seed)
	}
	if err != nil {
		return observationpreserved.Prediction{}, err
	}
	last := view.Trace[len(view.Trace)-1]
	p := observationpreserved.Prediction{Weights: w, Mask: last.Observed, Values: last.Values, Cost: view.Cost, Version: models.Version, Guide: guide, Split: s.split, Experts: [4]float64{.5, .5, .5, .5}}
	for j, m := range ms {
		if m != nil {
			p.Experts[j], err = m.ForecastObserved(p.Mask, p.Values)
			if err != nil {
				return observationpreserved.Prediction{}, err
			}
		}
	}
	inside := [4]float64{p.Experts[1], .5, .5, .5}
	if available {
		v, e := subset.Forecast(p.Mask, p.Values)
		if e != nil {
			return observationpreserved.Prediction{}, e
		}
		inside[1], inside[2], inside[3] = v, v, v
		p.Experts[1] = s.inner.Forecast(inside)
	}
	p.P = s.mix.Forecast(p.Experts)
	s.pending = &subsetPending{p: p, inner: inside, innerWeights: iw, available: available, subsetGuide: useSubset}
	return p, nil
}
func (s *subsetState) observe(sequence int, y, authorized bool) error {
	if s.pending == nil || sequence != s.next {
		return errors.New("missing or duplicate subset feedback")
	}
	s.mix = s.mix.Observe(s.pending.p.Experts, y, 1)
	if s.pending.available {
		s.inner = s.inner.Observe(s.pending.inner, y, 1)
	}
	// Revocation only changes the pooled slot; it cannot erase the incumbent
	// or the independently learned count/subset comparison in the short slot.
	if authorized && !s.split {
		s.split = true
		s.mix.Weights[2] = math.Min(.1, s.mix.Weights[2])
		sum := 0.
		for _, w := range s.mix.Weights {
			sum += w
		}
		for j := range s.mix.Weights {
			s.mix.Weights[j] /= sum
		}
	}
	s.pending = nil
	s.next++
	return nil
}
