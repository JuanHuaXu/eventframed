// Package researchretention studies bounded, member-specific window selection.
// It is a serial research component, not a production learner or certificate.
package researchretention

import (
	"errors"
	"math"
)

const Experts, MaxMembers, MaxTrials = 3, 200, 64

// Expert is copied at issue. Joint indexes 2*W1+W2 and measures ONE latent Y.
// The caller must produce it before either measurement arrives.
type Expert struct {
	Clean float64
	Joint [4]float64
}
type Forecasts [Experts]Expert
type Weights [Experts]float64
type row struct {
	forecasts                              Forecasts
	posterior                              Weights
	first, second                          uint8
	w1, w2                                 bool
	at, requestAt                          int64
	issuedClean, issuedFirst, issuedSecond float64
}
type member struct {
	rows  [MaxTrials]row
	count int
}
type Selector struct {
	members []member
	epoch   uint64
	clock   int64
}
type Ticket struct {
	owner           *Selector
	epoch           uint64
	member, ordinal int
	second          bool
	forecast        float64
}

func (t Ticket) Forecast() float64 { return t.forecast }

type Receipt struct {
	Member, Ordinal, Measurement int
	Epoch                        uint64
	IssuedAt, ArrivedAt          int64
	Forecast                     float64
	Value                        bool
}

func New(members int, epoch uint64) (*Selector, error) {
	if members < 1 || members > MaxMembers || epoch == 0 {
		return nil, errors.New("retention constructor cap/epoch")
	}
	return &Selector{members: make([]member, members), epoch: epoch}, nil
}
func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func bit(x bool) int {
	if x {
		return 1
	}
	return 0
}
func uniform() Weights { return Weights{1. / 3, 1. / 3, 1. / 3} }
func transition(w Weights, ordinal int) Weights {
	if ordinal == 1 {
		return uniform()
	}
	a := 1 / float64(ordinal)
	for k := range w {
		w[k] = (1-a)*w[k] + a/Experts
	}
	return w
}
func normalize(w Weights) (Weights, error) {
	sum := 0.
	for _, p := range w {
		if !finite(p) || p < 0 {
			return Weights{}, errors.New("retention invalid weight")
		}
		sum += p
	}
	if !finite(sum) || sum <= 0 {
		return Weights{}, errors.New("retention zero evidence")
	}
	for k := range w {
		w[k] /= sum
	}
	return w, nil
}
func validateForecasts(f Forecasts) error {
	for _, e := range f {
		if !finite(e.Clean) || e.Clean < 0 || e.Clean > 1 {
			return errors.New("retention clean forecast")
		}
		sum := 0.
		for _, p := range e.Joint {
			if !finite(p) || p < 0 || p > 1 {
				return errors.New("retention joint forecast")
			}
			sum += p
		}
		if math.Abs(sum-1) > 2e-12 {
			return errors.New("retention joint normalization")
		}
	}
	return nil
}
func factor(r row) Weights {
	w := Weights{1, 1, 1}
	if r.first != 2 {
		return w
	}
	a := 2 * bit(r.w1)
	for k, e := range r.forecasts {
		w[k] = e.Joint[a] + e.Joint[a+1]
		if r.second == 2 {
			w[k] = e.Joint[a+bit(r.w2)]
		}
	}
	return w
}
func (s *Selector) Weights(i int) (Weights, error) {
	if i < 0 || i >= len(s.members) {
		return Weights{}, errors.New("retention member")
	}
	m := &s.members[i]
	if m.count == 0 {
		return uniform(), nil
	}
	return transition(m.rows[m.count-1].posterior, m.count+1), nil
}
func (s *Selector) Issue(i int, at int64, f Forecasts) (Ticket, error) {
	if i < 0 || i >= len(s.members) || at < s.clock || s.members[i].count >= MaxTrials {
		return Ticket{}, errors.New("retention issue cap/time")
	}
	if e := validateForecasts(f); e != nil {
		return Ticket{}, e
	}
	w, e := s.Weights(i)
	if e != nil {
		return Ticket{}, e
	}
	q, first := 0., 0.
	for k, p := range w {
		q += p * f[k].Clean
		first += p * (f[k].Joint[2] + f[k].Joint[3])
	}
	m := &s.members[i]
	n := m.count
	m.rows[n] = row{forecasts: f, posterior: w, first: 1, at: at, issuedClean: q, issuedFirst: first}
	m.count++
	s.clock = at
	return Ticket{owner: s, epoch: s.epoch, member: i, ordinal: n, forecast: q}, nil
}
func (s *Selector) validate(t Ticket, at int64) error {
	if t.owner != s || t.epoch != s.epoch || t.member < 0 || t.member >= len(s.members) || t.ordinal < 0 || t.ordinal >= s.members[t.member].count || at < s.clock {
		return errors.New("retention owner/epoch/time")
	}
	r := s.members[t.member].rows[t.ordinal]
	if at < r.at {
		return errors.New("retention origin time")
	}
	if t.second {
		if r.second != 1 || at < r.requestAt {
			return errors.New("retention second state/time")
		}
	} else if r.first != 1 {
		return errors.New("retention first replay")
	}
	return nil
}

// Replay only the affected member from the changed origin, in a bounded scratch
// array. Publication and lifecycle flags commit together after all normalization.
func (s *Selector) prepare(i, n int, replacement row) ([MaxTrials]Weights, error) {
	var scratch [MaxTrials]Weights
	m := &s.members[i]
	previous := uniform()
	if n > 0 {
		previous = m.rows[n-1].posterior
	}
	for j := n; j < m.count; j++ {
		r := m.rows[j]
		if j == n {
			r = replacement
		}
		w := transition(previous, j+1)
		f := factor(r)
		for k := range w {
			w[k] *= f[k]
		}
		var e error
		previous, e = normalize(w)
		if e != nil {
			return scratch, e
		}
		scratch[j] = previous
	}
	return scratch, nil
}
func (s *Selector) Resolve(t Ticket, value bool, at int64) (Receipt, error) {
	if e := s.validate(t, at); e != nil {
		return Receipt{}, e
	}
	m := &s.members[t.member]
	r := m.rows[t.ordinal]
	if t.second {
		r.second = 2
		r.w2 = value
	} else {
		r.first = 2
		r.w1 = value
	}
	scratch, e := s.prepare(t.member, t.ordinal, r)
	if e != nil {
		return Receipt{}, e
	}
	m.rows[t.ordinal] = r
	for n := t.ordinal; n < m.count; n++ {
		m.rows[n].posterior = scratch[n]
	}
	s.clock = at
	measurement, issuedAt, forecast := 1, r.at, r.issuedFirst
	if t.second {
		measurement, issuedAt, forecast = 2, r.requestAt, r.issuedSecond
	}
	return Receipt{t.member, t.ordinal + 1, measurement, s.epoch, issuedAt, at, forecast, value}, nil
}
func (s *Selector) Cancel(t Ticket, at int64) error {
	if e := s.validate(t, at); e != nil {
		return e
	}
	r := &s.members[t.member].rows[t.ordinal]
	if t.second {
		r.second = 3
	} else {
		r.first = 3
	}
	s.clock = at
	return nil
}

// OriginWeights smooths the original label with all NOW-visible evidence.
// It cannot look at a requested packet's unreceived value.
func (s *Selector) OriginWeights(t Ticket) (Weights, error) {
	if t.second || t.owner != s || t.epoch != s.epoch || t.member < 0 || t.member >= len(s.members) || t.ordinal < 0 || t.ordinal >= s.members[t.member].count {
		return Weights{}, errors.New("retention smoothing origin")
	}
	m := &s.members[t.member]
	r := m.rows[t.ordinal]
	if r.first != 2 || r.second != 0 {
		return Weights{}, errors.New("retention smoothing state")
	}
	back := Weights{1, 1, 1}
	for n := m.count - 1; n > t.ordinal; n-- {
		f := factor(m.rows[n])
		a := 1 / float64(n+1)
		sum := 0.
		for k := range back {
			sum += f[k] * back[k]
		}
		var next Weights
		for k := range back {
			next[k] = (1-a)*f[k]*back[k] + a*sum/Experts
		}
		var e error
		back, e = normalize(next)
		if e != nil {
			return Weights{}, e
		}
	}
	w := r.posterior
	for k := range w {
		w[k] *= back[k]
	}
	return normalize(w)
}
func (s *Selector) QuerySecond(t Ticket) (float64, error) {
	w, e := s.OriginWeights(t)
	if e != nil {
		return 0, e
	}
	r := s.members[t.member].rows[t.ordinal]
	a := 2 * bit(r.w1)
	q := 0.
	for k, p := range w {
		marginal := r.forecasts[k].Joint[a] + r.forecasts[k].Joint[a+1]
		if marginal == 0 {
			if p > 0 {
				return 0, errors.New("retention impossible first evidence")
			}
			continue
		}
		q += p * r.forecasts[k].Joint[a+1] / marginal
	}
	if !finite(q) || q < 0 || q > 1+2e-12 {
		return 0, errors.New("retention conditional forecast")
	}
	return math.Min(q, 1), nil
}
func (s *Selector) RequestSecond(t Ticket, at int64) (Ticket, error) {
	if at < s.clock {
		return Ticket{}, errors.New("retention request time")
	}
	q, e := s.QuerySecond(t)
	if e != nil {
		return Ticket{}, e
	}
	r := &s.members[t.member].rows[t.ordinal]
	r.second = 1
	r.requestAt = at
	r.issuedSecond = q
	s.clock = at
	t.second = true
	t.forecast = q
	return t, nil
}
func (s *Selector) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= s.epoch || at < s.clock {
		return errors.New("retention epoch")
	}
	next, e := New(len(s.members), epoch)
	if e != nil {
		return e
	}
	next.clock = at
	*s = *next
	return nil
}
