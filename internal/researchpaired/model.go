// Package researchpaired is an isolated original-position joint-measurement
// learner. Its source-independence assumption is not an authentication claim.
package researchpaired

import (
	"errors"
	prior "github.com/JuanHuaXu/eventframed/internal/researchnoisemoment"
	"math"
)

const Atoms, Families, Components, MaxTrials = 21, 28, 84, 64

var noiseRates = [3]float64{0, .1, .2}
var noisePrior = [3]float64{.8, .1, .1}

type Config struct{ Strength, Hazard float64 }
type trial struct {
	first, second                  uint8
	w1, w2                         bool
	at, auditAt                    int64
	clean, observed, auditForecast float64
}
type Ticket struct {
	owner  *Model
	epoch  uint64
	slot   int
	second bool
	q      float64
}

func (t Ticket) Forecast() float64 { return t.q }

type Receipt struct {
	Member, Ordinal, Measurement int
	Epoch                        uint64
	IssuedAt, ArrivedAt          int64
	Forecast                     float64
	Value                        bool
}
type Option struct{ Observed, Uncertainty, Information, EdgeCut float64 }
type Model struct {
	base                []float64
	cfg                 Config
	epoch               uint64
	clock               int64
	cap, pending        int
	issued              []int
	trials              []trial
	prior, latest, logs []float64
	weights             [Components]float64
}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func atom(z int) float64    { return (float64(z) + .5) / Atoms }
func familyMean(b float64, i, n, h int) float64 {
	switch h {
	case 0:
		return b
	case 1:
		return 1 - b
	case 2:
		return .5
	}
	j := h - 3
	return math.Max(.025, math.Min(.975, .1+.2*float64(j/5)+(-.8+.4*float64(j%5))*float64(i)/float64(n-1)))
}
func familyPrior(h int) float64 {
	switch h {
	case 0:
		return .72
	case 1, 2:
		return .09
	default:
		return .004
	}
}
func entropy(p float64) float64 {
	if p == 0 || p == 1 {
		return 0
	}
	return -p*math.Log(p) - (1-p)*math.Log1p(-p)
}
func New(base []float64, epoch uint64, cap int, cfg Config) (*Model, error) {
	if len(base) < 2 || len(base) > 200 || epoch == 0 || cap < 1 || cap > 2*len(base)*MaxTrials || !finite(cfg.Strength) || cfg.Strength <= 0 || cfg.Strength > 32 || !finite(cfg.Hazard) || cfg.Hazard < 0 || cfg.Hazard >= 1 {
		return nil, errors.New("paired contract")
	}
	for _, b := range base {
		if !finite(b) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("paired baseline")
		}
	}
	m := &Model{base: append([]float64(nil), base...), cfg: cfg, epoch: epoch, cap: cap, issued: make([]int, len(base)), trials: make([]trial, len(base)*MaxTrials), prior: make([]float64, len(base)*Families*Atoms), latest: make([]float64, len(base)*Components*Atoms), logs: make([]float64, len(base)*Components)}
	var cache [128]struct {
		mu float64
		p  [Atoms]float64
	}
	used, next := 0, 0
	for i, b := range base {
		for h := 0; h < Families; h++ {
			mu := familyMean(b, i, len(base), h)
			p := [Atoms]float64{}
			found := false
			for k := 0; k < used; k++ {
				if cache[k].mu == mu {
					p = cache[k].p
					found = true
					break
				}
			}
			if !found {
				var e error
				p, e = prior.DiscretePrior(mu, cfg.Strength, "moment")
				if e != nil {
					return nil, e
				}
				cache[next].mu, cache[next].p = mu, p
				next = (next + 1) % len(cache)
				used = min(used+1, len(cache))
			}
			copy(m.prior[(i*Families+h)*Atoms:], p[:])
			for eta := 0; eta < 3; eta++ {
				c := eta*Families + h
				copy(m.latest[(i*Components+c)*Atoms:], p[:])
				m.weights[c] = noisePrior[eta] * familyPrior(h)
			}
		}
	}
	return m, nil
}
func likelihood(p, eta float64, x trial) float64 {
	if x.first != 2 {
		return 1
	}
	if x.second != 2 {
		q := eta + (1-2*eta)*p
		if !x.w1 {
			q = 1 - q
		}
		return q
	}
	a, b := eta, eta
	if x.w1 {
		a = 1 - eta
	}
	if x.w2 {
		b = 1 - eta
	}
	return p*a*b + (1-p)*(1-a)*(1-b)
}
func (m *Model) next(i, c, z int) float64 {
	k := (i*Components+c)*Atoms + z
	p := m.prior[(i*Families+c%Families)*Atoms+z]
	if m.issued[i] == 0 {
		return p
	}
	return (1-m.cfg.Hazard)*m.latest[k] + m.cfg.Hazard*p
}
func (m *Model) Predict(i int) (float64, float64, error) {
	if i < 0 || i >= len(m.base) {
		return 0, 0, errors.New("paired member")
	}
	q, o := 0., 0.
	for c, w := range m.weights {
		eta := noiseRates[c/Families]
		for z := 0; z < Atoms; z++ {
			mass := w * m.next(i, c, z)
			q += mass * atom(z)
			o += mass * (eta + (1-2*eta)*atom(z))
		}
	}
	if !finite(q) || !finite(o) || q <= 0 || q >= 1 || o <= 0 || o >= 1 {
		return 0, 0, errors.New("paired forecast")
	}
	return q, o, nil
}
func (m *Model) NoiseWeights() [3]float64 {
	var out [3]float64
	for c, w := range m.weights {
		out[c/Families] += w
	}
	return out
}
func (m *Model) Issue(i int, at int64) (Ticket, error) {
	if i < 0 || i >= len(m.base) || at < m.clock || m.pending >= m.cap || m.issued[i] >= MaxTrials {
		return Ticket{}, errors.New("paired issue cap/time")
	}
	q, o, e := m.Predict(i)
	if e != nil {
		return Ticket{}, e
	}
	for c := 0; c < Components; c++ {
		for z := 0; z < Atoms; z++ {
			m.latest[(i*Components+c)*Atoms+z] = m.next(i, c, z)
		}
	}
	slot := i*MaxTrials + m.issued[i]
	m.trials[slot] = trial{first: 1, at: at, clean: q, observed: o}
	m.issued[i]++
	m.pending++
	m.clock = at
	return Ticket{owner: m, epoch: m.epoch, slot: slot, q: q}, nil
}

// integrate recomputes ONE member's complete factor history into scratch. A
// genuine zero-probability component is not a NaN and may be eliminated.
func (m *Model) integrate(i, override int, replacement trial) ([Components][Atoms]float64, [Components]float64, error) {
	var rows [Components][Atoms]float64
	var logs [Components]float64
	for c := 0; c < Components; c++ {
		p := m.prior[(i*Families+c%Families)*Atoms : (i*Families+c%Families+1)*Atoms]
		copy(rows[c][:], p)
		for j := 0; j < m.issued[i]; j++ {
			slot := i*MaxTrials + j
			x := m.trials[slot]
			if slot == override {
				x = replacement
			}
			if j > 0 {
				for z := range rows[c] {
					rows[c][z] = (1-m.cfg.Hazard)*rows[c][z] + m.cfg.Hazard*p[z]
				}
			}
			sum := 0.
			for z := range rows[c] {
				rows[c][z] *= likelihood(atom(z), noiseRates[c/Families], x)
				sum += rows[c][z]
			}
			if !finite(sum) || sum < 0 {
				return rows, logs, errors.New("paired numerical evidence fault")
			}
			if sum == 0 {
				logs[c] = math.Inf(-1)
				copy(rows[c][:], p)
				break
			}
			logs[c] += math.Log(sum)
			for z := range rows[c] {
				rows[c][z] /= sum
			}
		}
	}
	return rows, logs, nil
}
func (m *Model) normalized(i int, logs [Components]float64) ([Components]float64, error) {
	var w [Components]float64
	maximum := math.Inf(-1)
	// Full sum avoids Inf-Inf when replacing an impossible member's marginal.
	for c := range w {
		v := math.Log(noisePrior[c/Families] * familyPrior(c%Families))
		for member := range m.base {
			l := m.logs[member*Components+c]
			if member == i {
				l = logs[c]
			}
			if math.IsNaN(l) || math.IsInf(l, 1) {
				return w, errors.New("paired nonfinite global evidence")
			}
			v += l
		}
		w[c] = v
		maximum = math.Max(maximum, v)
	}
	if !finite(maximum) {
		return w, errors.New("paired all-model failure")
	}
	sum := 0.
	for c := range w {
		w[c] = math.Exp(w[c] - maximum)
		sum += w[c]
	}
	if !finite(sum) || sum <= 0 {
		return w, errors.New("paired weight normalization")
	}
	for c := range w {
		w[c] /= sum
	}
	return w, nil
}
func (m *Model) validate(t Ticket, at int64, known bool) error {
	if t.owner != m || t.epoch != m.epoch || t.slot < 0 || t.slot >= len(m.trials) || at < m.clock {
		return errors.New("paired owner/epoch/time")
	}
	x := m.trials[t.slot]
	if t.second {
		if x.second != 1 || at < x.auditAt {
			return errors.New("paired second replay/time")
		}
	} else {
		want := uint8(1)
		if known {
			want = 2
		}
		if x.first != want || at < x.at {
			return errors.New("paired first replay/time")
		}
	}
	return nil
}
func (m *Model) Resolve(t Ticket, value bool, at int64) (Receipt, error) {
	if e := m.validate(t, at, false); e != nil {
		return Receipt{}, e
	}
	x := m.trials[t.slot]
	if t.second {
		x.second, x.w2 = 2, value
	} else {
		x.first, x.w1 = 2, value
	}
	i := t.slot / MaxTrials
	rows, logs, e := m.integrate(i, t.slot, x)
	if e != nil {
		return Receipt{}, e
	}
	weights, e := m.normalized(i, logs)
	if e != nil {
		return Receipt{}, e
	}
	for c := range rows {
		copy(m.latest[(i*Components+c)*Atoms:], rows[c][:])
		m.logs[i*Components+c] = logs[c]
	}
	m.weights = weights
	m.trials[t.slot] = x
	m.pending--
	m.clock = at
	measurement, issuedAt, forecast := 1, x.at, x.clean
	if t.second {
		measurement, issuedAt, forecast = 2, x.auditAt, x.auditForecast
	}
	return Receipt{Member: i, Ordinal: t.slot%MaxTrials + 1, Measurement: measurement, Epoch: m.epoch, IssuedAt: issuedAt, ArrivedAt: at, Forecast: forecast, Value: value}, nil
}
func (m *Model) Cancel(t Ticket, at int64) error {
	if e := m.validate(t, at, false); e != nil {
		return e
	}
	if t.second {
		m.trials[t.slot].second = 3
	} else {
		m.trials[t.slot].first = 3
	}
	m.pending--
	m.clock = at
	return nil
}
func (m *Model) Pending() int { return m.pending }
func (m *Model) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= m.epoch || at < m.clock {
		return errors.New("paired epoch")
	}
	next, e := New(m.base, epoch, m.cap, m.cfg)
	if e != nil {
		return e
	}
	next.clock = at
	*m = *next
	return nil
}

// Options predicts the second observer at the ORIGINAL outcome position,
// conditioned on all evidence already revealed, including later known frames.
func (m *Model) Options(t Ticket) (Option, error) {
	return m.options(t, "all")
}

// Query computes only the requested policy statistic. Random audit requests
// need an origin-bound forecast, not unused information-selection work.
func (m *Model) Query(t Ticket, mode string) (Option, error) { return m.options(t, mode) }
func (m *Model) options(t Ticket, mode string) (Option, error) {
	if mode != "all" && mode != "forecast" && mode != "uncertainty" && mode != "information" && mode != "falsification" {
		return Option{}, errors.New("paired query mode")
	}
	if t.second || t.owner != m || t.epoch != m.epoch || t.slot < 0 || t.slot >= len(m.trials) {
		return Option{}, errors.New("paired option identity")
	}
	x := m.trials[t.slot]
	if x.first != 2 || x.second != 0 {
		return Option{}, errors.New("paired option availability")
	}
	i := t.slot / MaxTrials
	x.second, x.w2 = 2, true
	_, yes, e := m.integrate(i, t.slot, x)
	if e != nil {
		return Option{}, e
	}
	x.w2 = false
	_, no, e := m.integrate(i, t.slot, x)
	if e != nil {
		return Option{}, e
	}
	q, conditionalEntropy := 0., 0.
	var probabilities [Components]float64
	for c, w := range m.weights {
		if w == 0 {
			continue
		}
		old := m.logs[i*Components+c]
		if !finite(old) {
			return Option{}, errors.New("paired live component impossible")
		}
		py, pn := math.Exp(yes[c]-old), math.Exp(no[c]-old)
		if !finite(py) || !finite(pn) || py < 0 || pn < 0 || math.Abs(py+pn-1) > 2e-10 {
			return Option{}, errors.New("paired conditional normalization")
		}
		py = math.Min(1, py)
		probabilities[c] = py
		q += w * py
		if mode == "all" || mode == "information" {
			conditionalEntropy += w * entropy(py)
		}
	}
	if !finite(q) || q < 0 || q > 1+2e-12 {
		return Option{}, errors.New("paired option forecast")
	}
	q = math.Min(1, q)
	if mode == "forecast" {
		return Option{Observed: q}, nil
	}
	u := entropy(q)
	if mode == "uncertainty" {
		return Option{Observed: q, Uncertainty: u}, nil
	}
	edge := 0.
	if mode == "all" || mode == "falsification" {
		for c, w := range m.weights {
			py := probabilities[c]
			term := -1.
			if q > 0 {
				term += py * py / q
			}
			if q < 1 {
				term += (1 - py) * (1 - py) / (1 - q)
			}
			edge += w * w * term
		}
		if !finite(edge) || edge < -2e-10 {
			return Option{}, errors.New("paired negative concentration gain")
		}
		edge = math.Max(0, edge)
	}
	if mode == "falsification" {
		return Option{Observed: q, EdgeCut: edge}, nil
	}
	info := u - conditionalEntropy
	if !finite(info) || info < -2e-10 {
		return Option{}, errors.New("paired negative information")
	}
	return Option{Observed: q, Uncertainty: u, Information: math.Max(0, info), EdgeCut: edge}, nil
}
func (m *Model) RequestAudit(t Ticket, at int64) (Ticket, error) {
	if t.second || m.pending >= m.cap {
		return Ticket{}, errors.New("paired audit cap/type")
	}
	if e := m.validate(t, at, true); e != nil {
		return Ticket{}, e
	}
	option, e := m.Query(t, "forecast")
	if e != nil {
		return Ticket{}, e
	}
	m.trials[t.slot].second = 1
	m.trials[t.slot].auditAt = at
	m.trials[t.slot].auditForecast = option.Observed
	m.pending++
	m.clock = at
	return Ticket{owner: m, epoch: m.epoch, slot: t.slot, second: true, q: option.Observed}, nil
}
