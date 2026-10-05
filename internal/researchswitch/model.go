// Package researchswitch implements an isolated delayed Bernoulli expert-state
// filter. It is not a production learner, Anti-Pigeon certificate, or detector.
package researchswitch

import (
	"errors"
	"math"
	"sort"
)

const MaxExperts, MaxTrials = 8, 4096
const ProbabilityFloor = 1e-6

type Config struct {
	Prior           []float64
	Hazard          float64
	Trials, Pending int
}

type row struct {
	advice, logPosterior [MaxExperts]float64
	at                   int64
	forecast             float64
	status               uint8
	useful               bool
}

type Model struct {
	prior, logPrior [MaxExperts]float64
	logReset        [MaxExperts]float64
	logStay         float64
	rows            []row
	work            [][MaxExperts]float64
	known           []int
	hazard          float64
	epoch           uint64
	clock           int64
	n, used, cap    int
	pending         int
}

// Tickets are opaque, owner/epoch-bound handles to privately retained advice.
// Callers cannot replace original predictions when delayed evidence arrives.
type Ticket struct {
	owner *Model
	epoch uint64
	slot  int
	q     float64
}

func (t Ticket) Forecast() float64 { return t.q }

type Receipt struct {
	Epoch               uint64
	TrialOrdinal        int
	IssuedAt, ArrivedAt int64
	Forecast            float64
	Useful              bool
}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

func New(cfg Config, epoch uint64) (*Model, error) {
	if len(cfg.Prior) < 2 || len(cfg.Prior) > MaxExperts || !finite(cfg.Hazard) || cfg.Hazard < 0 || cfg.Hazard > 1 || cfg.Trials < 1 || cfg.Trials > MaxTrials || cfg.Pending < 1 || cfg.Pending > cfg.Trials || epoch == 0 {
		return nil, errors.New("invalid switching contract")
	}
	sum := 0.
	for _, w := range cfg.Prior {
		if !finite(w) || w <= 0 || w > 1 {
			return nil, errors.New("invalid expert prior")
		}
		sum += w
	}
	if math.Abs(sum-1) > 1e-12 {
		return nil, errors.New("expert prior must sum to one")
	}
	m := &Model{rows: make([]row, cfg.Trials), work: make([][MaxExperts]float64, cfg.Trials), known: make([]int, 0, cfg.Trials), hazard: cfg.Hazard, epoch: epoch, n: len(cfg.Prior), cap: cfg.Pending}
	copy(m.prior[:], cfg.Prior)
	// Normalize accepted round-off once so unit-likelihood transitions remain
	// probability laws even at the configured simplex tolerance.
	for i := 0; i < m.n; i++ {
		m.prior[i] /= sum
		m.logPrior[i] = math.Log(m.prior[i])
	}
	// The hazard and prior are immutable within an epoch. Cache only their
	// transition constants; retain the original floating-point expression.
	if m.hazard > 0 && m.hazard < 1 {
		m.logStay = math.Log1p(-m.hazard)
		for i := 0; i < m.n; i++ {
			m.logReset[i] = math.Log(m.hazard) + m.logPrior[i]
		}
	}
	return m, nil
}

func (m *Model) transitionLogs(previous [MaxExperts]float64) [MaxExperts]float64 {
	if m.hazard == 0 {
		return previous
	}
	if m.hazard == 1 {
		return m.logPrior
	}
	var next [MaxExperts]float64
	for i := 0; i < m.n; i++ {
		a, b := m.logStay+previous[i], m.logReset[i]
		maximum := math.Max(a, b)
		next[i] = maximum + math.Log(math.Exp(a-maximum)+math.Exp(b-maximum))
	}
	return next
}

func (m *Model) nextLogs() [MaxExperts]float64 {
	if m.used == 0 {
		return m.logPrior
	}
	return m.transitionLogs(m.logsAt(m.used - 1))
}

// Unknown/cancelled positions have unit likelihood. The reset-prior matrix
// composes exactly: T^k = (1-alpha)^k I + (1-(1-alpha)^k) 1*pi.
// Expm1 retains small reset mass; zero/full hazard remain exact endpoints.
func (m *Model) advanceLogs(previous [MaxExperts]float64, steps int) [MaxExperts]float64 {
	if steps == 0 || m.hazard == 0 {
		return previous
	}
	if steps == 1 {
		return m.transitionLogs(previous)
	}
	if m.hazard == 1 {
		return m.logPrior
	}
	stay := float64(steps) * m.logStay
	reset := math.Log(-math.Expm1(stay))
	var next [MaxExperts]float64
	for i := 0; i < m.n; i++ {
		a, b := stay+previous[i], reset+m.logPrior[i]
		maximum := math.Max(a, b)
		next[i] = maximum + math.Log(math.Exp(a-maximum)+math.Exp(b-maximum))
	}
	return next
}

func (m *Model) logsAt(slot int) [MaxExperts]float64 {
	k := sort.SearchInts(m.known, slot+1) - 1
	if k < 0 {
		// Position zero has the prior without a preceding transition.
		return m.advanceLogs(m.logPrior, slot)
	}
	anchor := m.known[k]
	return m.advanceLogs(m.rows[anchor].logPosterior, slot-anchor)
}

// Linear weights may round to zero for output, but finite latent log weights
// MUST remain recoverable after contrary evidence. Do not feed rounded masses
// back into the filter; that would make zero an unintended absorbing state.
func (m *Model) probabilities(logs [MaxExperts]float64) [MaxExperts]float64 {
	var weights [MaxExperts]float64
	for i := 0; i < m.n; i++ {
		weights[i] = math.Exp(logs[i])
	}
	return weights
}

// Advice must be produced from the caller's AS-OF visible history by all
// experts. Their generation/training cost is additional to this filter cost.
func (m *Model) Predict(advice []float64) (float64, error) {
	if len(advice) != m.n {
		return 0, errors.New("expert advice arity")
	}
	w := m.probabilities(m.nextLogs())
	q := 0.
	for i, p := range advice {
		if !finite(p) || p < ProbabilityFloor || p > 1-ProbabilityFloor || !finite(w[i]) || w[i] < 0 {
			return 0, errors.New("invalid expert prediction")
		}
		q += w[i] * p
	}
	if !finite(q) || q <= 0 || q >= 1 {
		return 0, errors.New("invalid switched forecast")
	}
	return q, nil
}

func (m *Model) Issue(advice []float64, at int64) (Ticket, error) {
	if at < m.clock || m.used == len(m.rows) || m.pending == m.cap {
		return Ticket{}, errors.New("invalid or capped switching issue")
	}
	q, err := m.Predict(advice)
	if err != nil {
		return Ticket{}, err
	}
	r := row{at: at, forecast: q, status: 1, logPosterior: m.nextLogs()}
	copy(r.advice[:], advice)
	slot := m.used
	m.rows[slot] = r
	m.used++
	m.pending++
	m.clock = at
	return Ticket{m, m.epoch, slot, q}, nil
}

func (m *Model) validate(t Ticket, at int64) error {
	if t.owner != m || t.epoch != m.epoch || t.slot < 0 || t.slot >= m.used || at < m.clock || at < m.rows[t.slot].at || m.rows[t.slot].status != 1 {
		return errors.New("invalid switching owner epoch time or status")
	}
	return nil
}

// Replay at ORIGINAL nomination positions, never arrival order. Cached expert
// advice is immutable; only hidden-state forward messages may be revised.
// Scratch messages make a numerical failure atomic. Unknown/cancelled spans
// consume every original transition but need no per-position replay.
func (m *Model) Resolve(t Ticket, useful bool, at int64) (Receipt, error) {
	if err := m.validate(t, at); err != nil {
		return Receipt{}, err
	}
	insert := sort.SearchInts(m.known, t.slot)
	previous, position := m.logPrior, 0
	if insert > 0 {
		position = m.known[insert-1]
		previous = m.rows[position].logPosterior
	}
	for k := insert; k <= len(m.known); k++ {
		j := t.slot
		if k > insert {
			j = m.known[k-1]
		}
		w := m.advanceLogs(previous, j-position)
		r := m.rows[j]
		y := r.useful
		if j == t.slot {
			y = useful
		}
		for i := 0; i < m.n; i++ {
			p := r.advice[i]
			if !y {
				p = 1 - p
			}
			if !finite(p) || p < ProbabilityFloor/2 || p > 1 {
				return Receipt{}, errors.New("invalid retained expert likelihood")
			}
			w[i] += math.Log(p)
		}
		maximum := math.Inf(-1)
		for i := 0; i < m.n; i++ {
			if !finite(w[i]) {
				return Receipt{}, errors.New("invalid forward message")
			}
			maximum = math.Max(maximum, w[i])
		}
		sum := 0.
		for i := 0; i < m.n; i++ {
			sum += math.Exp(w[i] - maximum)
		}
		if !finite(sum) || sum <= 0 {
			return Receipt{}, errors.New("invalid switching normalization")
		}
		normalizer := maximum + math.Log(sum)
		for i := 0; i < m.n; i++ {
			w[i] -= normalizer
		}
		m.work[j] = w
		previous, position = w, j
	}
	for k := insert; k <= len(m.known); k++ {
		j := t.slot
		if k > insert {
			j = m.known[k-1]
		}
		m.rows[j].logPosterior = m.work[j]
	}
	m.known = append(m.known, 0)
	copy(m.known[insert+1:], m.known[insert:])
	m.known[insert] = t.slot
	r := m.rows[t.slot]
	m.rows[t.slot].status, m.rows[t.slot].useful = 2, useful
	m.pending--
	m.clock = at
	return Receipt{m.epoch, t.slot + 1, r.at, at, r.forecast, useful}, nil
}

func (m *Model) Cancel(t Ticket, at int64) error {
	if err := m.validate(t, at); err != nil {
		return err
	}
	m.rows[t.slot].status = 3
	m.pending--
	m.clock = at
	return nil
}

func (m *Model) Pending() int { return m.pending }

func (m *Model) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= m.epoch || at < m.clock {
		return errors.New("invalid switching epoch")
	}
	cfg := Config{Prior: append([]float64(nil), m.prior[:m.n]...), Hazard: m.hazard, Trials: len(m.rows), Pending: m.cap}
	next, err := New(cfg, epoch)
	if err != nil {
		return err
	}
	next.clock = at
	*m = *next
	return nil
}
