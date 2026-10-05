// Package observationpreserved is an isolated research composition, not serving.
package observationpreserved

import (
	"errors"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
)

var Arms = []string{"frozen_mmm", "mix_breadth_ap", "mix_random_ap", "mix_mmm_no_ap", "mix_mmm_local", "mix_mmm_ap"}

// Evidence implements eight predeclared two-sided betting processes. Validity
// concerns a bounded conditional mean of paired correctness, not target diameter.
type Evidence struct {
	Plus, Minus [8]float64
	Next        int
	Max         float64
}

func (e *Evidence) Observe(reference, live bool) bool {
	d := 0.
	if reference {
		d++
	}
	if live {
		d--
	}
	for j := range e.Plus {
		if e.Next < j*64 {
			continue
		}
		if e.Next == j*64 {
			e.Plus[j], e.Minus[j] = 1, 1
		}
		e.Plus[j] *= 1 + .25*(d-.15)
		e.Minus[j] *= 1 + .25*(-d-.15)
		e.Max = math.Max(e.Max, (e.Plus[j]+e.Minus[j])/2)
	}
	e.Next++
	return e.Max >= 800
}

type Models struct {
	Short, Local, Pooled *observation.Model
	Version              int
}
type Prediction struct {
	P       float64    `json:"p"`
	Experts [4]float64 `json:"experts"`
	Weights [4]float64 `json:"weights_before_share"`
	Mask    uint16     `json:"mask"`
	Values  uint16     `json:"values"`
	Cost    int        `json:"cost"`
	Version int        `json:"version"`
	Guide   int        `json:"guide"`
	Split   bool       `json:"split"`
}
type State struct {
	Arm     string
	base    *observation.Model
	mix     bayes.ForecastMix
	split   bool
	next    int
	pending *Prediction
}

func New(base *observation.Model, arm string) (*State, error) {
	valid := false
	for _, a := range Arms {
		if a == arm {
			valid = true
		}
	}
	if base == nil || !valid {
		return nil, errors.New("invalid preserved experiment configuration")
	}
	return &State{Arm: arm, base: base}, nil
}
func (s *State) Predict(r observation.Reader, m Models, sequence int, seed int64) (Prediction, error) {
	if s.pending != nil || sequence != s.next {
		return Prediction{}, errors.New("prediction order violation")
	}
	long := m.Pooled
	if s.split || s.Arm == "mix_mmm_local" {
		long = m.Local
	}
	models := [3]*observation.Model{s.base, m.Short, long}
	w := s.mix.Weights
	if w == [4]float64{} {
		w = [4]float64{.7, .1, .1, .1}
	}
	guide := 0
	if s.Arm != "frozen_mmm" {
		for j := 1; j < 3; j++ {
			if models[j] != nil && w[j] > w[guide] {
				guide = j
			}
		}
	}
	policy := "mmm"
	if s.Arm == "mix_breadth_ap" {
		policy = "breadth"
	}
	if s.Arm == "mix_random_ap" {
		policy = "random"
	}
	inspection, err := observation.Run(models[guide], r, r.Epoch(), policy, seed)
	if err != nil {
		return Prediction{}, err
	}
	last := inspection.Trace[len(inspection.Trace)-1]
	p := Prediction{Weights: w, Mask: last.Observed, Values: last.Values, Cost: inspection.Cost, Version: m.Version, Guide: guide, Split: s.split}
	p.Experts = [4]float64{.5, .5, .5, .5}
	for j, model := range models {
		if model != nil {
			p.Experts[j], err = model.ForecastObserved(p.Mask, p.Values)
			if err != nil {
				return Prediction{}, err
			}
		}
	}
	p.P = s.mix.Forecast(p.Experts)
	if s.Arm == "frozen_mmm" {
		p.P = p.Experts[0]
	}
	s.pending = &p
	return p, nil
}

// Feedback uses the journaled predictions, never recomputes a model on its own
// revealing label. Sharing revocation changes only the pooled challenger slot.
func (s *State) Observe(sequence int, outcome, splitAuthorized bool) error {
	if s.pending == nil || sequence != s.next {
		return errors.New("missing or duplicate feedback")
	}
	if s.Arm != "frozen_mmm" {
		s.mix = s.mix.Observe(s.pending.Experts, outcome, 1)
	}
	if splitAuthorized && !s.split && (s.Arm == "mix_mmm_ap" || s.Arm == "mix_breadth_ap" || s.Arm == "mix_random_ap") {
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
