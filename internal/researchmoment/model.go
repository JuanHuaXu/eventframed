// Package researchmoment tests declared-mean finite priors in isolation.
// Shared calibration evidence is not source authentication or Anti-Pigeon authority.
package researchmoment

import (
	"errors"
	"math"
)

const Atoms, MaxFamilies, MaxTrials = 21, 28, 64

type Config struct {
	Family, Prior    string
	Strength, Hazard float64
	Shared           bool
}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func atom(z int) float64    { return (float64(z) + .5) / Atoms }

// DiscretePrior returns density-grid masses or their minimum-KL projection
// onto the declared mean. Strength changes higher moments, not only variance.
// No observations enter this constructor. Boundary means are rejected.
func DiscretePrior(mu, strength float64, mode string) ([Atoms]float64, error) {
	var logBase, out [Atoms]float64
	if !finite(mu) || mu <= atom(0) || mu >= atom(Atoms-1) || !finite(strength) || strength <= 0 || strength > 32 || (mode != "density" && mode != "moment") {
		return out, errors.New("invalid finite prior")
	}
	for z := range logBase {
		a := atom(z)
		logBase[z] = (strength*mu-1)*math.Log(a) + (strength*(1-mu)-1)*math.Log1p(-a)
	}
	weights := func(lambda float64) float64 {
		maximum := math.Inf(-1)
		for z, v := range logBase {
			maximum = math.Max(maximum, v+lambda*atom(z))
		}
		sum, mean := 0., 0.
		for z, v := range logBase {
			out[z] = math.Exp(v + lambda*atom(z) - maximum)
			sum += out[z]
		}
		for z := range out {
			out[z] /= sum
			mean += out[z] * atom(z)
		}
		return mean
	}
	if mode == "density" {
		weights(0)
	} else {
		lo, hi := -4096., 4096.
		if weights(lo) >= mu || weights(hi) <= mu {
			return out, errors.New("unbracketed mean")
		}
		for j := 0; j < 80; j++ {
			mid := (lo + hi) / 2
			if weights(mid) < mu {
				lo = mid
			} else {
				hi = mid
			}
		}
		if mean := weights((lo + hi) / 2); !finite(mean) || math.Abs(mean-mu) > 2e-12 {
			return out, errors.New("mean projection failed")
		}
	}
	sum := 0.
	for _, v := range out {
		if !finite(v) || v < 0 {
			return out, errors.New("invalid prior mass")
		}
		sum += v
	}
	if math.Abs(sum-1) > 2e-12 {
		return out, errors.New("invalid prior sum")
	}
	return out, nil
}

func familyMean(b float64, member, members, h int) float64 {
	switch h {
	case 0:
		return b
	case 1:
		return 1 - b
	case 2:
		return .5
	}
	j := h - 3
	a := .1 + .2*float64(j/5)
	c := -.8 + .4*float64(j%5)
	return math.Max(.025, math.Min(.975, a+c*float64(member)/float64(members-1)))
}
func familyWeights(h int) [MaxFamilies]float64 {
	w := [MaxFamilies]float64{.8, .1, .1}
	if h == MaxFamilies {
		w[0], w[1], w[2] = .72, .09, .09
		for j := 3; j < h; j++ {
			w[j] = .004
		}
	}
	return w
}

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

// Model retains only the latest conditional row, plus a bounded original
// trial ledger. A delayed emission recomputes the member's full as-of history.
// This trades more replay work for O(MHQ+ML) memory instead of O(MHQL).
// Unknown/cancelled emissions are unit likelihoods but still consume a time
// transition. Missingness must be ignorable. This is single-owner code.
type Model struct {
	base                []float64
	cfg                 Config
	epoch               uint64
	clock               int64
	h, cap, pending     int
	issued              []int
	trials              []trial
	prior, latest, logs []float64
	global, weights     [MaxFamilies]float64
}

func New(base []float64, epoch uint64, pendingCap int, cfg Config) (*Model, error) {
	if len(base) < 2 || len(base) > 200 || epoch == 0 || pendingCap < 1 || pendingCap > len(base)*MaxTrials || (cfg.Family != "narrow" && cfg.Family != "rich") || (cfg.Prior != "density" && cfg.Prior != "moment") || !finite(cfg.Strength) || cfg.Strength <= 0 || cfg.Strength > 32 || !finite(cfg.Hazard) || cfg.Hazard < 0 || cfg.Hazard >= 1 {
		return nil, errors.New("invalid moment contract")
	}
	for _, b := range base {
		if !finite(b) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("invalid moment baseline")
		}
	}
	h := 3
	if cfg.Family == "rich" {
		h = MaxFamilies
	}
	m := &Model{base: append([]float64(nil), base...), cfg: cfg, epoch: epoch, h: h, cap: pendingCap, issued: make([]int, len(base)), trials: make([]trial, len(base)*MaxTrials), prior: make([]float64, len(base)*h*Atoms), latest: make([]float64, len(base)*h*Atoms), logs: make([]float64, len(base)*h), weights: familyWeights(h)}
	for i, b := range base {
		for j := 0; j < h; j++ {
			p, e := DiscretePrior(familyMean(b, i, len(base), j), cfg.Strength, cfg.Prior)
			if e != nil {
				return nil, e
			}
			copy(m.prior[(i*h+j)*Atoms:], p[:])
		}
	}
	copy(m.latest, m.prior)
	return m, nil
}
func (m *Model) next(i, h, z int) float64 {
	k := (i*m.h+h)*Atoms + z
	if m.issued[i] == 0 {
		return m.prior[k]
	}
	return (1-m.cfg.Hazard)*m.latest[k] + m.cfg.Hazard*m.prior[k]
}
func (m *Model) posterior(logs [MaxFamilies]float64) ([MaxFamilies]float64, error) {
	w := familyWeights(m.h)
	maximum := math.Inf(-1)
	for h := 0; h < m.h; h++ {
		w[h] = math.Log(w[h]) + logs[h]
		if !finite(w[h]) {
			return w, errors.New("invalid family evidence")
		}
		maximum = math.Max(maximum, w[h])
	}
	sum := 0.
	for h := 0; h < m.h; h++ {
		w[h] = math.Exp(w[h] - maximum)
		sum += w[h]
	}
	if !finite(sum) || sum <= 0 {
		return w, errors.New("invalid family normalization")
	}
	for h := 0; h < m.h; h++ {
		w[h] /= sum
	}
	return w, nil
}
func (m *Model) Predict(i int) (float64, error) {
	if i < 0 || i >= len(m.base) {
		return 0, errors.New("unknown moment member")
	}
	w := m.weights
	if !m.cfg.Shared {
		var logs [MaxFamilies]float64
		copy(logs[:m.h], m.logs[i*m.h:(i+1)*m.h])
		var e error
		w, e = m.posterior(logs)
		if e != nil {
			return 0, e
		}
	}
	q := 0.
	for h := 0; h < m.h; h++ {
		for z := 0; z < Atoms; z++ {
			q += w[h] * m.next(i, h, z) * atom(z)
		}
	}
	if !finite(q) || q <= 0 || q >= 1 {
		return 0, errors.New("invalid moment forecast")
	}
	return q, nil
}
func (m *Model) Issue(i int, at int64) (Ticket, error) {
	if i < 0 || i >= len(m.base) || at < m.clock || m.pending >= m.cap || m.issued[i] >= MaxTrials {
		return Ticket{}, errors.New("invalid or capped moment issue")
	}
	q, e := m.Predict(i)
	if e != nil {
		return Ticket{}, e
	}
	for h := 0; h < m.h; h++ {
		for z := 0; z < Atoms; z++ {
			m.latest[(i*m.h+h)*Atoms+z] = m.next(i, h, z)
		}
	}
	slot := i*MaxTrials + m.issued[i]
	m.trials[slot] = trial{status: 1, at: at, forecast: q}
	m.issued[i]++
	m.pending++
	m.clock = at
	return Ticket{m, m.epoch, slot, q}, nil
}
func (m *Model) validate(t Ticket, at int64) error {
	if t.owner != m || t.epoch != m.epoch || t.slot < 0 || t.slot >= len(m.trials) || at < m.clock {
		return errors.New("invalid moment owner/epoch/time")
	}
	x := m.trials[t.slot]
	if x.status != 1 || at < x.at {
		return errors.New("replayed cancelled or premature moment ticket")
	}
	return nil
}

// Resolve replaces the member's old marginal evidence; it never multiplies
// the same label twice. Validate scratch state before publishing any mutation.
func (m *Model) Resolve(t Ticket, useful bool, at int64) (Receipt, error) {
	if e := m.validate(t, at); e != nil {
		return Receipt{}, e
	}
	i := t.slot / MaxTrials
	var rows [MaxFamilies][Atoms]float64
	var logs [MaxFamilies]float64
	for h := 0; h < m.h; h++ {
		copy(rows[h][:], m.prior[(i*m.h+h)*Atoms:(i*m.h+h+1)*Atoms])
	}
	for j := 0; j < m.issued[i]; j++ {
		slot := i*MaxTrials + j
		x := m.trials[slot]
		for h := 0; h < m.h; h++ {
			if j > 0 {
				for z := 0; z < Atoms; z++ {
					rows[h][z] = (1-m.cfg.Hazard)*rows[h][z] + m.cfg.Hazard*m.prior[(i*m.h+h)*Atoms+z]
				}
			}
			if slot == t.slot || x.status == 2 {
				y := x.useful
				if slot == t.slot {
					y = useful
				}
				sum := 0.
				for z := 0; z < Atoms; z++ {
					p := atom(z)
					if !y {
						p = 1 - p
					}
					rows[h][z] *= p
					sum += rows[h][z]
				}
				if !finite(sum) || sum <= 0 {
					return Receipt{}, errors.New("invalid moment emission normalization")
				}
				logs[h] += math.Log(sum)
				for z := 0; z < Atoms; z++ {
					rows[h][z] /= sum
				}
			}
		}
	}
	global := m.global
	for h := 0; h < m.h; h++ {
		global[h] += logs[h] - m.logs[i*m.h+h]
	}
	w, e := m.posterior(global)
	if e != nil {
		return Receipt{}, e
	}
	// Private-mode prediction also must be valid before publication.
	if _, e = m.posterior(logs); e != nil {
		return Receipt{}, e
	}
	for h := 0; h < m.h; h++ {
		copy(m.latest[(i*m.h+h)*Atoms:], rows[h][:])
		m.logs[i*m.h+h] = logs[h]
	}
	x := m.trials[t.slot]
	m.trials[t.slot].status = 2
	m.trials[t.slot].useful = useful
	m.global, m.weights = global, w
	m.pending--
	m.clock = at
	return Receipt{i, t.slot%MaxTrials + 1, m.epoch, x.at, at, x.forecast, useful}, nil
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
		return errors.New("invalid moment epoch")
	}
	next, e := New(m.base, epoch, m.cap, m.cfg)
	if e != nil {
		return e
	}
	next.clock = at
	*m = *next
	return nil
}
