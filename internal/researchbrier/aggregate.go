// Package researchbrier isolates binary strong Brier aggregation. Its weights
// are loss-based strategy weights, not Bayesian evidence or split authority.
package researchbrier

import (
	"errors"
	"math"
)

const MaxExperts, MaxTrials = 8, 4096

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

func logSum(x []float64) float64 {
	maximum := x[0]
	for _, v := range x[1:] {
		maximum = math.Max(maximum, v)
	}
	sum := 0.
	for _, v := range x {
		sum += math.Exp(v - maximum)
	}
	return maximum + math.Log(sum)
}

// Forecast reduces Algorithm1 of Vovk and Zhdanov (2009) to two outcomes.
// Their categorical Brier loss is 2*(p-y)^2 in our single-coordinate scale.
// The output is a probability, not the arithmetic mixture of expert advice.
func Forecast(logWeights, advice []float64) (float64, error) {
	n := len(advice)
	if n < 1 || n > MaxExperts || len(logWeights) != n {
		return 0, errors.New("invalid Brier arity")
	}
	maximum, lo, hi := math.Inf(-1), 1., 0.
	for j, p := range advice {
		if !finite(p) || p < 0 || p > 1 || !finite(logWeights[j]) {
			return 0, errors.New("invalid Brier advice or weight")
		}
		maximum = math.Max(maximum, logWeights[j])
		lo = math.Min(lo, p)
		hi = math.Max(hi, p)
	}
	if lo == hi {
		return lo, nil
	}
	var zero, one [MaxExperts]float64
	for j, p := range advice {
		zero[j] = logWeights[j] - maximum - 2*p*p
		one[j] = logWeights[j] - maximum - 2*(1-p)*(1-p)
	}
	q := .5 + (logSum(one[:n])-logSum(zero[:n]))/4
	if !finite(q) {
		return 0, errors.New("invalid Brier substitution")
	}
	// Exact substitution is inside the experts' range. Projection repairs only
	// boundary roundoff (including sub-ulp endpoint advice), not model error.
	return math.Max(lo, math.Min(hi, q)), nil
}

// Model requires serialized ownership. One unresolved ticket blocks further
// nominations; wall-clock delay is allowed, but not further issued forecasts.
// Regret comparisons are per completed epoch, not across discarded tickets.
type Model struct {
	prior, logs     [MaxExperts]float64
	advice          [MaxExperts]float64
	n, cap, used    int
	epoch           uint64
	clock, issuedAt int64
	pending         bool
	q               float64
}
type Ticket struct {
	owner   *Model
	epoch   uint64
	ordinal int
	q       float64
}

func (t Ticket) Forecast() float64 { return t.q }

type Receipt struct {
	Epoch               uint64
	TrialOrdinal        int
	IssuedAt, ArrivedAt int64
	Forecast            float64
	Useful              bool
}

// New requires positive prior masses. Immediate single-pending admission is
// explicit; no delayed-replay or missing-label regret theorem is asserted.
func New(prior []float64, epoch uint64, capacity int) (*Model, error) {
	if len(prior) < 1 || len(prior) > MaxExperts || epoch == 0 || capacity < 1 || capacity > MaxTrials {
		return nil, errors.New("invalid Brier contract")
	}
	m := &Model{n: len(prior), cap: capacity, epoch: epoch}
	for j, p := range prior {
		if !finite(p) || p <= 0 {
			return nil, errors.New("invalid Brier prior")
		}
		m.logs[j] = math.Log(p)
	}
	norm := logSum(m.logs[:m.n])
	for j := 0; j < m.n; j++ {
		m.logs[j] -= norm
		m.prior[j] = m.logs[j]
	}
	return m, nil
}
func (m *Model) Predict(advice []float64) (float64, error) { return Forecast(m.logs[:m.n], advice) }
func (m *Model) Issue(advice []float64, at int64) (Ticket, error) {
	if m.pending || m.used == m.cap || at < m.clock {
		return Ticket{}, errors.New("Brier pending cap or time")
	}
	q, e := m.Predict(advice)
	if e != nil {
		return Ticket{}, e
	}
	copy(m.advice[:m.n], advice)
	m.used++
	m.q = q
	m.pending = true
	m.clock = at
	m.issuedAt = at
	return Ticket{m, m.epoch, m.used, q}, nil
}
func (m *Model) Resolve(t Ticket, useful bool, at int64) (Receipt, error) {
	if t.owner != m || t.epoch != m.epoch || t.ordinal != m.used || t.q != m.q || !m.pending || at < m.clock {
		return Receipt{}, errors.New("invalid Brier ticket or time")
	}
	y := 0.
	if useful {
		y = 1
	}
	next := m.logs
	for j := 0; j < m.n; j++ {
		d := m.advice[j] - y
		next[j] -= 2 * d * d
	}
	norm := logSum(next[:m.n])
	for j := 0; j < m.n; j++ {
		next[j] -= norm
		if !finite(next[j]) {
			return Receipt{}, errors.New("invalid Brier update")
		}
	}
	m.logs = next
	m.pending = false
	m.clock = at
	return Receipt{m.epoch, m.used, m.issuedAt, at, t.q, useful}, nil
}
func (m *Model) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= m.epoch || at < m.clock {
		return errors.New("invalid Brier epoch")
	}
	m.logs = m.prior
	m.advice = [MaxExperts]float64{}
	m.used = 0
	m.pending = false
	m.q = 0
	m.epoch = epoch
	m.clock = at
	m.issuedAt = 0
	return nil
}
