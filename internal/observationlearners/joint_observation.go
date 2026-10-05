package observationlearners

import (
	"errors"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// Publication owns copies. A caller replacing a fitted source table cannot
// rewrite a previously prepared observation law. The fixed domain is 3^9 cells.
type observationExperts struct {
	models [4]ConditionalForest
	roots  [4]float64
	ready  bool
}

func newObservationExperts(models [4]*ConditionalForest) (*observationExperts, error) {
	e := new(observationExperts)
	for j, m := range models {
		if m == nil {
			return nil, errors.New("missing observation expert")
		}
		e.models[j] = *m
		e.roots[j] = e.models[j].cells[0].mass
		for i, c := range e.models[j].cells {
			if math.IsNaN(c.mass) || math.IsInf(c.mass, 0) || c.mass <= 0 || math.IsNaN(c.weighted) || math.IsInf(c.weighted, 0) || c.weighted <= 0 || c.weighted >= c.mass {
				return nil, errors.New("invalid full-support joint table")
			}
			// Validate the same disjoint-child identity used by the archived table.
			v, p := i, 1
			for bit := 0; bit < 9; bit++ {
				if v%3 == 0 {
					a, b := e.models[j].cells[i+p], e.models[j].cells[i+2*p]
					if !jointClose(c.mass, a.mass+b.mass) || !jointClose(c.weighted, a.weighted+b.weighted) {
						return nil, errors.New("incoherent joint table")
					}
					break
				}
				v /= 3
				p *= 3
			}
		}
	}
	for i := range e.models[0].cells {
		mass := e.models[0].cells[i].mass / e.roots[0]
		if mass <= 0 || math.IsNaN(mass) || math.IsInf(mass, 0) {
			return nil, errors.New("unnormalizable input law")
		}
		for j := 1; j < 4; j++ {
			if !jointClose(mass, e.models[j].cells[i].mass/e.roots[j]) {
				return nil, errors.New("observation experts have different input laws")
			}
		}
	}
	e.ready = true
	return e, nil
}

func jointClose(a, b float64) bool {
	return !math.IsNaN(a) && !math.IsNaN(b) && !math.IsInf(a, 0) && !math.IsInf(b, 0) && math.Abs(a-b) <= 1e-12*math.Max(math.Abs(a), math.Abs(b))
}

type observationMixture struct {
	experts *observationExperts
	weights [5]float64 // neutral followed by the four raw experts
}

func (e *observationExperts) snapshot(weights [5]float64) (observationMixture, error) {
	if e == nil || !e.ready {
		return observationMixture{}, errors.New("missing expert publication")
	}
	total := 0.
	for _, w := range weights {
		if math.IsNaN(w) || math.IsInf(w, 0) || w < 0 || w > 1 {
			return observationMixture{}, errors.New("invalid observation mixture weight")
		}
		total += w
	}
	if math.Abs(total-1) > 1e-12 {
		return observationMixture{}, errors.New("observation weights do not sum to one")
	}
	for i := range weights {
		weights[i] /= total
	}
	return observationMixture{experts: e, weights: weights}, nil
}

func (s observationMixture) cell(mask, values uint16) conditionalCell {
	i := partialIndex(mask, values)
	mass := s.experts.models[0].cells[i].mass / s.experts.roots[0]
	weighted := s.weights[0] * .5 * mass
	for j := 0; j < 4; j++ {
		weighted += s.weights[j+1] * (s.experts.models[j].cells[i].weighted / s.experts.roots[j])
	}
	return conditionalCell{mass: mass, weighted: weighted}
}

func (s observationMixture) Forecast(mask, values uint16) (float64, error) {
	if s.experts == nil || !s.experts.ready || mask >= 512 || values&^mask != 0 {
		return 0, errors.New("invalid mixture observation")
	}
	c := s.cell(mask, values)
	return math.Max(0, math.Min(1, c.weighted/c.mass)), nil
}

// Hypothetical inspection copies the routing state; it does not issue evidence.
// Only finish commits the actual acquired mask to the bank and routing gate.
type routedObservationState struct {
	bank brierBank
	gate evidenceRouting
	busy bool
}

func newRoutedObservationState() *routedObservationState {
	b, _ := newBrierBank([]float64{.95, .05 / 3, .05 / 3, .05 / 3}, .001)
	return &routedObservationState{bank: *b}
}

func (s *routedObservationState) preview(sequence uint64, e *observationExperts) (observationMixture, error) {
	if s == nil || s.busy || s.bank.count != 4 || s.bank.active || sequence != s.bank.issued {
		return observationMixture{}, errors.New("invalid observation preview")
	}
	copy := s.gate
	f, err := copy.predict(sequence, [4]float64{.5, .5, .5, .5}, s.bank.weights())
	if err != nil {
		return observationMixture{}, err
	}
	return e.snapshot(f.Weights)
}

func (s *routedObservationState) finish(sequence uint64, m observationMixture, mask, values uint16) (routedForecast, error) {
	if s == nil || m.experts == nil {
		return routedForecast{}, errors.New("missing observation state")
	}
	want, err := m.Forecast(mask, values)
	if err != nil {
		return routedForecast{}, err
	}
	var p [4]float64
	for j := range p {
		p[j], err = m.experts.models[j].Forecast(mask, values)
		if err != nil {
			return routedForecast{}, err
		}
	}
	copy := *s
	b, err := copy.bank.predict(sequence, p[:])
	if err != nil {
		return routedForecast{}, err
	}
	f, err := copy.gate.predict(sequence, p, b.Weights)
	if err != nil {
		return routedForecast{}, err
	}
	if math.Abs(f.P-want) > 1e-12 {
		return routedForecast{}, errors.New("observation/served law mismatch")
	}
	// Equal endpoint probabilities do not establish equal acquisition laws.
	for i, w := range f.Weights {
		if math.Abs(w-m.weights[i]) > 1e-12 {
			return routedForecast{}, errors.New("observation mixture weight mismatch")
		}
	}
	*s = copy
	return f, nil
}

func (s *routedObservationState) predict(sequence uint64, e *observationExperts, reader observation.Reader, epoch uint64) (observation.Result, error) {
	m, err := s.preview(sequence, e)
	if err != nil {
		return observation.Result{}, err
	}
	s.busy = true
	defer func() { s.busy = false }()
	r, err := runJointObserver(m, reader, epoch)
	if err != nil {
		return observation.Result{}, err
	}
	last := r.Trace[len(r.Trace)-1]
	f, err := s.finish(sequence, m, last.Observed, last.Values)
	if err != nil {
		return observation.Result{}, err
	}
	r.Probability = f.P
	return r, nil
}

func (s *routedObservationState) observe(sequence uint64, y bool) error {
	if s == nil || s.busy {
		return errors.New("invalid observation feedback")
	}
	copy := *s
	if err := copy.bank.observe(sequence, y); err != nil {
		return err
	}
	if err := copy.gate.observe(sequence, y); err != nil {
		return err
	}
	*s = copy
	return nil
}
