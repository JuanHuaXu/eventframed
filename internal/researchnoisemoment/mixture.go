package researchnoisemoment

import (
	"errors"
	"math"
)

// Mixture compares three alternative joint models, not independent evidence
// copies. One latent noise rate is shared across this isolated epoch.
type Mixture struct {
	children [3]*Model
	prior    [3]float64
	epoch    uint64
	clock    int64
	poison   bool
	served   []float64
}
type MixtureTicket struct {
	owner           *Mixture
	epoch           uint64
	member, ordinal int
	at              int64
	q               float64
	children        [3]Ticket
}

func (t MixtureTicket) Forecast() float64 { return t.q }

func NewMixture(base []float64, epoch uint64, cap int, cfg Config) (*Mixture, error) {
	if !cfg.Shared || cfg.Noise != 0 {
		return nil, errors.New("mixture requires shared family and implicit noise grid")
	}
	m := &Mixture{prior: [3]float64{.8, .1, .1}, epoch: epoch}
	for j, eta := range []float64{0, .1, .2} {
		cfg.Noise = eta
		child, e := NewMemoV49(base, epoch, cap, cfg)
		if e != nil {
			return nil, e
		}
		m.children[j] = child
	}
	m.served = make([]float64, len(base)*MaxTrials)
	return m, nil
}

func (m *Mixture) Weights() ([3]float64, error) {
	var w [3]float64
	if m.poison {
		return w, errors.New("poisoned noise bundle")
	}
	maximum := math.Inf(-1)
	for j, c := range m.children {
		l, e := c.LogEvidence()
		if e != nil {
			return w, e
		}
		w[j] = math.Log(m.prior[j]) + l
		maximum = math.Max(maximum, w[j])
	}
	sum := 0.
	for j := range w {
		w[j] = math.Exp(w[j] - maximum)
		sum += w[j]
	}
	if !finite(sum) || sum <= 0 {
		return w, errors.New("noise evidence normalization")
	}
	for j := range w {
		w[j] /= sum
	}
	return w, nil
}

func (m *Mixture) predict(i int, observed bool) (float64, error) {
	w, e := m.Weights()
	if e != nil {
		return 0, e
	}
	q := 0.
	for j, c := range m.children {
		v, e := c.Predict(i)
		if e != nil {
			return 0, e
		}
		if observed {
			v = c.cfg.Noise + (1-2*c.cfg.Noise)*v
		}
		q += w[j] * v
	}
	if !finite(q) || q <= 0 || q >= 1 {
		return 0, errors.New("noise mixture forecast")
	}
	return q, nil
}
func (m *Mixture) Predict(i int) (float64, error)         { return m.predict(i, false) }
func (m *Mixture) ObservedPredict(i int) (float64, error) { return m.predict(i, true) }

func (m *Mixture) Issue(i int, at int64) (MixtureTicket, error) {
	c := m.children[0]
	if m.poison || i < 0 || i >= len(c.base) || at < m.clock || c.pending >= c.cap || c.issued[i] >= MaxTrials {
		return MixtureTicket{}, errors.New("invalid or capped noise mixture issue")
	}
	q, e := m.Predict(i)
	if e != nil {
		return MixtureTicket{}, e
	}
	t := MixtureTicket{owner: m, epoch: m.epoch, member: i, ordinal: c.issued[i] + 1, at: at, q: q}
	for j, c := range m.children {
		t.children[j], e = c.Issue(i, at)
		if e != nil {
			m.poison, m.clock = true, at
			return MixtureTicket{}, e
		}
	}
	m.served[t.children[0].slot] = q
	m.clock = at
	return t, nil
}
func (m *Mixture) validate(t MixtureTicket, at int64) error {
	if m.poison || t.owner != m || t.epoch != m.epoch || at < m.clock {
		return errors.New("noise mixture owner epoch time")
	}
	for j, c := range m.children {
		if e := c.validate(t.children[j], at); e != nil {
			return e
		}
		if t.children[j].slot != t.children[0].slot {
			return errors.New("noise mixture child identity mismatch")
		}
	}
	return nil
}
func (m *Mixture) Resolve(t MixtureTicket, y bool, at int64) (Receipt, error) {
	if e := m.validate(t, at); e != nil {
		return Receipt{}, e
	}
	var original Receipt
	for j, c := range m.children {
		r, e := c.Resolve(t.children[j], y, at)
		if e != nil {
			m.poison, m.clock = true, at
			return Receipt{}, e
		}
		if j == 0 {
			original = r
		}
	}
	// Predictive receipts bind the originally served CLEAN law, not noisy label law.
	m.clock = at
	original.Forecast = m.served[t.children[0].slot]
	return original, nil
}
func (m *Mixture) Cancel(t MixtureTicket, at int64) error {
	if e := m.validate(t, at); e != nil {
		return e
	}
	for j, c := range m.children {
		if e := c.Cancel(t.children[j], at); e != nil {
			m.poison, m.clock = true, at
			return e
		}
	}
	m.clock = at
	return nil
}
func (m *Mixture) Pending() int { return m.children[0].Pending() }
func (m *Mixture) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= m.epoch || at < m.clock {
		return errors.New("invalid noise mixture epoch")
	}
	c := m.children[0]
	cfg := c.cfg
	cfg.Noise = 0
	next, e := NewMixture(c.base, epoch, c.cap, cfg)
	if e != nil {
		return e
	}
	next.clock = at
	for _, v := range next.children {
		v.clock = at
	}
	*m = *next
	return nil
}
