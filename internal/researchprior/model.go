// Package researchprior is an isolated finite hierarchical rate model.
// Its posterior predicts outcomes; it neither authenticates sources nor grants
// Anti-Pigeon sharing, changepoint, graph, or production authority.
package researchprior

import (
	"errors"
	"math"
)

const Atoms, Families, MaxTrials = 21, 3, 64

type Config struct {
	Center           string
	Strength, Hazard float64
	Shared           bool
}

type row [Families][Atoms]float64
type trial struct {
	status   uint8
	useful   bool
	at       int64
	forecast float64
}

type Ticket struct {
	owner *Model
	epoch uint64
	slot  int
	q     float64
}

func (t Ticket) Forecast() float64 { return t.q }

type Receipt struct {
	Member, TrialOrdinal int
	Epoch                uint64
	IssuedAt, ArrivedAt  int64
	Forecast             float64
	Useful               bool
}

// Model has independent member chains conditional on one calibration family
// (Shared), or separate calibration families per member (private control).
// Both modes have the SAME prior marginal. Delayed emissions enter their
// original member position. It is single-owner, not concurrency-safe.
type Model struct {
	base            []float64
	cfg             Config
	epoch           uint64
	clock           int64
	cap, pending    int
	issued          []int
	trials          []trial
	prior           []row
	forward         []row
	evidence        [][Families]float64
	global          [Families]float64
	weights         [Families]float64
	private         [][Families]float64
	scratch         [MaxTrials]row
	scratchEvidence [MaxTrials][Families]float64
}

func atom(z int) float64             { return (float64(z) + .5) / Atoms }
func familyPrior() [Families]float64 { return [Families]float64{.8, .1, .1} }
func center(b float64, mode string) float64 {
	if mode == "inverse" {
		return (.9*b - .05) / .8
	}
	return b
}

func New(base []float64, epoch uint64, pendingCap int, cfg Config) (*Model, error) {
	if len(base) < 2 || len(base) > 200 || epoch == 0 || pendingCap < 1 || pendingCap > len(base)*MaxTrials ||
		(cfg.Center != "raw" && cfg.Center != "inverse") || !finite(cfg.Strength) || cfg.Strength <= 0 || cfg.Strength > 32 ||
		!finite(cfg.Hazard) || cfg.Hazard < 0 || cfg.Hazard >= 1 {
		return nil, errors.New("invalid prior contract")
	}
	for _, b := range base {
		if !finite(b) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("invalid prior baseline")
		}
	}
	n := len(base) * MaxTrials
	m := &Model{base: append([]float64(nil), base...), cfg: cfg, epoch: epoch, cap: pendingCap, issued: make([]int, len(base)), trials: make([]trial, n), prior: make([]row, len(base)), forward: make([]row, n), evidence: make([][Families]float64, n), private: make([][Families]float64, len(base)), weights: familyPrior()}
	for i, b := range base {
		p := center(b, cfg.Center)
		for h, mu := range []float64{p, 1 - p, .5} {
			sum := 0.
			for z := 0; z < Atoms; z++ {
				q := atom(z)
				v := math.Pow(q, cfg.Strength*mu-1) * math.Pow(1-q, cfg.Strength*(1-mu)-1)
				m.prior[i][h][z] = v
				sum += v
			}
			if !finite(sum) || sum <= 0 {
				return nil, errors.New("invalid prior normalization")
			}
			for z := 0; z < Atoms; z++ {
				m.prior[i][h][z] /= sum
			}
		}
		m.private[i] = familyPrior()
	}
	return m, nil
}
func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func (m *Model) next(member int) row {
	w := m.prior[member]
	if n := m.issued[member]; n > 0 {
		last := m.forward[member*MaxTrials+n-1]
		for h := 0; h < Families; h++ {
			for z := 0; z < Atoms; z++ {
				w[h][z] = (1-m.cfg.Hazard)*last[h][z] + m.cfg.Hazard*w[h][z]
			}
		}
	}
	return w
}
func (m *Model) Predict(member int) (float64, error) {
	if member < 0 || member >= len(m.base) {
		return 0, errors.New("unknown prior member")
	}
	w := m.next(member)
	weights := m.private[member]
	if m.cfg.Shared {
		weights = m.weights
	}
	q := 0.
	for h := 0; h < Families; h++ {
		for z := 0; z < Atoms; z++ {
			q += weights[h] * w[h][z] * atom(z)
		}
	}
	if !finite(q) || q <= 0 || q >= 1 {
		return 0, errors.New("invalid prior forecast")
	}
	return q, nil
}
func (m *Model) Issue(member int, at int64) (Ticket, error) {
	if member < 0 || member >= len(m.base) || at < m.clock || m.pending >= m.cap || m.issued[member] >= MaxTrials {
		return Ticket{}, errors.New("invalid or capped prior issue")
	}
	q, e := m.Predict(member)
	if e != nil {
		return Ticket{}, e
	}
	slot := member*MaxTrials + m.issued[member]
	m.forward[slot] = m.next(member)
	if m.issued[member] > 0 {
		m.evidence[slot] = m.evidence[slot-1]
	}
	m.trials[slot] = trial{status: 1, at: at, forecast: q}
	m.issued[member]++
	m.pending++
	m.clock = at
	return Ticket{m, m.epoch, slot, q}, nil
}
func (m *Model) validate(t Ticket, at int64) error {
	if t.owner != m || t.epoch != m.epoch || t.slot < 0 || t.slot >= len(m.trials) || at < m.clock {
		return errors.New("invalid prior ticket owner/epoch/time")
	}
	x := m.trials[t.slot]
	if x.status != 1 || at < x.at {
		return errors.New("replayed cancelled or premature prior ticket")
	}
	return nil
}
func posterior(logs [Families]float64) ([Families]float64, error) {
	w := familyPrior()
	maximum := math.Inf(-1)
	for h := range w {
		w[h] = logs[h] + math.Log(w[h])
		if !finite(w[h]) {
			return w, errors.New("invalid hyper evidence")
		}
		maximum = math.Max(maximum, w[h])
	}
	sum := 0.
	for h := range w {
		w[h] = math.Exp(w[h] - maximum)
		sum += w[h]
	}
	if !finite(sum) || sum <= 0 {
		return w, errors.New("invalid hyper normalization")
	}
	for h := range w {
		w[h] /= sum
	}
	return w, nil
}

// Resolve replaces this member's OLD marginal likelihood contribution with
// its NEW one, never multiplies old evidence twice. Scratch rows and both
// posterior alternatives validate before publishing any change. Other member
// conditionals remain intact; only the shared family weights can move them.
func (m *Model) Resolve(t Ticket, useful bool, at int64) (Receipt, error) {
	if e := m.validate(t, at); e != nil {
		return Receipt{}, e
	}
	i, start := t.slot/MaxTrials, t.slot%MaxTrials
	w := m.prior[i]
	logs := [Families]float64{}
	if start > 0 {
		w = m.forward[t.slot-1]
		logs = m.evidence[t.slot-1]
	}
	for j := start; j < m.issued[i]; j++ {
		if j > 0 {
			for h := 0; h < Families; h++ {
				for z := 0; z < Atoms; z++ {
					w[h][z] = (1-m.cfg.Hazard)*w[h][z] + m.cfg.Hazard*m.prior[i][h][z]
				}
			}
		}
		slot := i*MaxTrials + j
		x := m.trials[slot]
		if slot == t.slot || x.status == 2 {
			y := x.useful
			if slot == t.slot {
				y = useful
			}
			for h := 0; h < Families; h++ {
				sum := 0.
				for z := 0; z < Atoms; z++ {
					p := atom(z)
					if !y {
						p = 1 - p
					}
					w[h][z] *= p
					sum += w[h][z]
				}
				if !finite(sum) || sum <= 0 {
					return Receipt{}, errors.New("invalid member evidence")
				}
				logs[h] += math.Log(sum)
				for z := 0; z < Atoms; z++ {
					w[h][z] /= sum
				}
			}
		}
		m.scratch[j] = w
		m.scratchEvidence[j] = logs
	}
	last := i*MaxTrials + m.issued[i] - 1
	global := m.global
	for h := range global {
		global[h] += logs[h] - m.evidence[last][h]
	}
	shared, e := posterior(global)
	if e != nil {
		return Receipt{}, e
	}
	private, e := posterior(logs)
	if e != nil {
		return Receipt{}, e
	}
	for j := start; j < m.issued[i]; j++ {
		slot := i*MaxTrials + j
		m.forward[slot] = m.scratch[j]
		m.evidence[slot] = m.scratchEvidence[j]
	}
	x := m.trials[t.slot]
	m.trials[t.slot].status = 2
	m.trials[t.slot].useful = useful
	m.global, m.weights, m.private[i] = global, shared, private
	m.pending--
	m.clock = at
	return Receipt{i, start + 1, m.epoch, x.at, at, x.forecast, useful}, nil
}
func (m *Model) Pending() int { return m.pending }
func (m *Model) Cancel(t Ticket, at int64) error {
	if e := m.validate(t, at); e != nil {
		return e
	}
	m.trials[t.slot].status = 3
	m.pending--
	m.clock = at
	return nil
}
func (m *Model) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= m.epoch || at < m.clock {
		return errors.New("invalid prior epoch")
	}
	next, e := New(m.base, epoch, m.cap, m.cfg)
	if e != nil {
		return e
	}
	next.clock = at
	*m = *next
	return nil
}
