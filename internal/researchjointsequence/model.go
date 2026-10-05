// Package researchjointsequence is an isolated finite joint sequence model.
// Each member has a static noise hypothesis and a changing rate state. This
// declared model supplies forecasts AND optional-observation probabilities;
// it does not authenticate the model or provide Anti-Pigeon authority.
package researchjointsequence

import (
	"errors"
	"math"
)

const Atoms, States, MaxTrials, MaxMembers = 22, 66, 64, 200

var noise = [3]float64{0, .1, .2}
var noisePrior = [3]float64{.8, .1, .1}

type Config struct{ Hazard float64 }
type distribution [States]float64
type row struct {
	posterior                      distribution
	at, auditAt                    int64
	clean, observed, auditForecast float64
	first, second                  uint8
	w1, w2                         bool
}
type member struct {
	rows  [MaxTrials]row
	count int
}
type Model struct {
	base         []float64
	priors       [][Atoms]float64
	factors      [][States][6]float64
	members      []member
	cfg          Config
	epoch        uint64
	clock        int64
	cap, pending int
}
type Ticket struct {
	owner           *Model
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
type Option struct{ Observed, Uncertainty, Information, EdgeCut, Value float64 }

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func bit(x bool) int {
	if x {
		return 1
	}
	return 0
}
func rate(base float64, z int) float64 {
	if z == 0 {
		return base
	}
	return float64(z-1) / 20
}
func normalize(p distribution) (distribution, error) {
	sum := 0.
	for _, x := range p {
		if !finite(x) || x < 0 {
			return distribution{}, errors.New("joint invalid mass")
		}
		sum += x
	}
	if !finite(sum) || sum <= 0 {
		return distribution{}, errors.New("joint zero support")
	}
	for j := range p {
		p[j] /= sum
	}
	return p, nil
}
func New(base []float64, epoch uint64, cap int, cfg Config) (*Model, error) {
	if len(base) < 2 || len(base) > MaxMembers || epoch == 0 || cap < 1 || cap > 2*len(base)*MaxTrials || !finite(cfg.Hazard) || cfg.Hazard < 0 || cfg.Hazard > 1 {
		return nil, errors.New("joint constructor contract")
	}
	for _, b := range base {
		if !finite(b) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("joint baseline")
		}
	}
	m := &Model{base: append([]float64(nil), base...), priors: make([][Atoms]float64, len(base)), factors: make([][States][6]float64, len(base)), members: make([]member, len(base)), cfg: cfg, epoch: epoch, cap: cap}
	for i, b := range base {
		// Finite Polya prior preserves the supplied mean; no fitting labels used.
		weights := [21]float64{}
		weights[0] = 1
		for n := 0; n < 20; n++ {
			weights[0] *= (1 - b + float64(n)) / (1 + float64(n))
		}
		for z := 0; z < 20; z++ {
			weights[z+1] = weights[z] * float64(20-z) / float64(z+1) * (b + float64(z)) / (1 - b + float64(19-z))
		}
		sum := 0.
		for _, v := range weights {
			sum += v
		}
		m.priors[i][0] = .8
		for z, v := range weights {
			m.priors[i][z+1] = .2 * v / sum
		}
		mean := 0.
		for z, p := range m.priors[i] {
			mean += p * rate(b, z)
		}
		if !finite(sum) || sum <= 0 || math.Abs(mean-b) > 2e-14 {
			return nil, errors.New("joint prior mean")
		}
		for h, eta := range noise {
			for z := 0; z < Atoms; z++ {
				p := rate(b, z)
				q := eta + (1-2*eta)*p
				state := h*Atoms + z
				m.factors[i][state][0], m.factors[i][state][1] = 1-q, q
				for a := 0; a < 2; a++ {
					for c := 0; c < 2; c++ {
						one, two := eta, eta
						if a == 1 {
							one = 1 - eta
						}
						if c == 1 {
							two = 1 - eta
						}
						m.factors[i][state][2+2*a+c] = p*one*two + (1-p)*(1-one)*(1-two)
					}
				}
			}
		}
	}
	return m, nil
}
func (m *Model) initial(i int) distribution {
	var p distribution
	for h, w := range noisePrior {
		for z, v := range m.priors[i] {
			p[h*Atoms+z] = w * v
		}
	}
	return p
}
func (m *Model) transition(i int, p distribution) distribution {
	var out distribution
	for h := 0; h < 3; h++ {
		sum := 0.
		for z := 0; z < Atoms; z++ {
			sum += p[h*Atoms+z]
		}
		for z, v := range m.priors[i] {
			out[h*Atoms+z] = (1-m.cfg.Hazard)*p[h*Atoms+z] + m.cfg.Hazard*sum*v
		}
	}
	return out
}
func (m *Model) next(i int, p distribution) (float64, float64, error) {
	p, e := normalize(m.transition(i, p))
	if e != nil {
		return 0, 0, e
	}
	q, o := 0., 0.
	for state, w := range p {
		q += w * rate(m.base[i], state%Atoms)
		o += w * m.factors[i][state][1]
	}
	return q, o, nil
}
func (m *Model) Predict(i int) (float64, float64, error) {
	if i < 0 || i >= len(m.base) {
		return 0, 0, errors.New("joint member")
	}
	x := &m.members[i]
	p := m.initial(i)
	if x.count > 0 {
		p = x.rows[x.count-1].posterior
	}
	return m.next(i, p)
}
func category(x row) int {
	if x.first != 2 {
		return -1
	}
	if x.second == 2 {
		return 2 + 2*bit(x.w1) + bit(x.w2)
	}
	return bit(x.w1)
}
func (m *Model) Issue(i int, at int64) (Ticket, error) {
	if i < 0 || i >= len(m.base) || at < m.clock || m.pending >= m.cap || m.members[i].count >= MaxTrials {
		return Ticket{}, errors.New("joint issue cap/time")
	}
	q, o, e := m.Predict(i)
	if e != nil {
		return Ticket{}, e
	}
	x := &m.members[i]
	p := m.initial(i)
	if x.count > 0 {
		p = m.transition(i, x.rows[x.count-1].posterior)
	}
	p, e = normalize(p)
	if e != nil {
		return Ticket{}, e
	}
	n := x.count
	x.rows[n] = row{posterior: p, first: 1, at: at, clean: q, observed: o}
	x.count++
	m.pending++
	m.clock = at
	return Ticket{owner: m, epoch: m.epoch, member: i, ordinal: n, forecast: q}, nil
}
func (m *Model) validate(t Ticket, at int64, known bool) error {
	if t.owner != m || t.epoch != m.epoch || t.member < 0 || t.member >= len(m.base) || t.ordinal < 0 || t.ordinal >= m.members[t.member].count || at < m.clock {
		return errors.New("joint owner/epoch/time")
	}
	x := m.members[t.member].rows[t.ordinal]
	want := uint8(1)
	if known {
		want = 2
	}
	if t.second {
		if x.second != 1 || at < x.auditAt {
			return errors.New("joint second replay/time")
		}
	} else if x.first != want || at < x.at {
		return errors.New("joint first replay/time")
	}
	return nil
}

// Delayed factors replace their original emission, then replay a bounded suffix.
// Nothing publishes until every normalization succeeds.
func (m *Model) prepare(i, n int, replacement row) ([MaxTrials]distribution, error) {
	var scratch [MaxTrials]distribution
	x := &m.members[i]
	p := m.initial(i)
	if n > 0 {
		p = x.rows[n-1].posterior
	}
	for j := n; j < x.count; j++ {
		if j > 0 {
			p = m.transition(i, p)
		}
		r := x.rows[j]
		if j == n {
			r = replacement
		}
		if c := category(r); c >= 0 {
			for state := range p {
				p[state] *= m.factors[i][state][c]
			}
		}
		var e error
		p, e = normalize(p)
		if e != nil {
			return scratch, e
		}
		scratch[j] = p
	}
	return scratch, nil
}
func (m *Model) Resolve(t Ticket, value bool, at int64) (Receipt, error) {
	if e := m.validate(t, at, false); e != nil {
		return Receipt{}, e
	}
	x := &m.members[t.member]
	r := x.rows[t.ordinal]
	if t.second {
		r.second, r.w2 = 2, value
	} else {
		r.first, r.w1 = 2, value
	}
	scratch, e := m.prepare(t.member, t.ordinal, r)
	if e != nil {
		return Receipt{}, e
	}
	x.rows[t.ordinal] = r
	for j := t.ordinal; j < x.count; j++ {
		x.rows[j].posterior = scratch[j]
	}
	m.pending--
	m.clock = at
	measurement, issued, forecast := 1, r.at, r.observed
	if t.second {
		measurement, issued, forecast = 2, r.auditAt, r.auditForecast
	}
	return Receipt{t.member, t.ordinal + 1, measurement, m.epoch, issued, at, forecast, value}, nil
}
func (m *Model) Cancel(t Ticket, at int64) error {
	if e := m.validate(t, at, false); e != nil {
		return e
	}
	r := &m.members[t.member].rows[t.ordinal]
	if t.second {
		r.second = 3
	} else {
		r.first = 3
	}
	m.pending--
	m.clock = at
	return nil
}
func (m *Model) Pending() int { return m.pending }
func (m *Model) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= m.epoch || at < m.clock {
		return errors.New("joint epoch")
	}
	next, e := New(m.base, epoch, m.cap, m.cfg)
	if e != nil {
		return e
	}
	next.clock = at
	*m = *next
	return nil
}

// Smooth the original state using only NOW-visible factors. Transpose action
// of the reset transition avoids a dense 66x66 multiply; H never resets here.
func (m *Model) origin(t Ticket) (distribution, error) {
	if t.second || t.owner != m || t.epoch != m.epoch || t.member < 0 || t.member >= len(m.base) || t.ordinal < 0 || t.ordinal >= m.members[t.member].count {
		return distribution{}, errors.New("joint query origin")
	}
	x := &m.members[t.member]
	r := x.rows[t.ordinal]
	if r.first != 2 || r.second != 0 {
		return distribution{}, errors.New("joint query availability")
	}
	back := distribution{}
	for state := range back {
		back[state] = 1
	}
	for n := x.count - 1; n > t.ordinal; n-- {
		if c := category(x.rows[n]); c >= 0 {
			for state := range back {
				back[state] *= m.factors[t.member][state][c]
			}
		}
		var next distribution
		for h := 0; h < 3; h++ {
			sum := 0.
			for z, p := range m.priors[t.member] {
				sum += p * back[h*Atoms+z]
			}
			for z := 0; z < Atoms; z++ {
				next[h*Atoms+z] = (1-m.cfg.Hazard)*back[h*Atoms+z] + m.cfg.Hazard*sum
			}
		}
		var e error
		back, e = normalize(next)
		if e != nil {
			return distribution{}, e
		}
	}
	p := r.posterior
	for state := range p {
		p[state] *= back[state]
	}
	return normalize(p)
}
func (m *Model) conditional(t Ticket) (float64, distribution, error) {
	p, e := m.origin(t)
	if e != nil {
		return 0, p, e
	}
	a := bit(m.members[t.member].rows[t.ordinal].w1)
	q := 0.
	for state, w := range p {
		den := m.factors[t.member][state][a]
		if den == 0 {
			if w > 0 {
				return 0, p, errors.New("joint unsupported first")
			}
			continue
		}
		q += w * m.factors[t.member][state][2+2*a+1] / den
	}
	if !finite(q) || q < 0 || q > 1+2e-12 {
		return 0, p, errors.New("joint query probability")
	}
	return math.Min(1, q), p, nil
}
func (m *Model) QuerySecond(t Ticket) (float64, error) { q, _, e := m.conditional(t); return q, e }
func (m *Model) RequestSecond(t Ticket, at int64) (Ticket, error) {
	if t.second || m.pending >= m.cap {
		return Ticket{}, errors.New("joint request cap/type")
	}
	if e := m.validate(t, at, true); e != nil {
		return Ticket{}, e
	}
	q, e := m.QuerySecond(t)
	if e != nil {
		return Ticket{}, e
	}
	r := &m.members[t.member].rows[t.ordinal]
	r.second, r.auditAt, r.auditForecast = 1, at, q
	m.pending++
	m.clock = at
	t.second, t.forecast = true, q
	return t, nil
}
func entropy(p float64) float64 {
	if p == 0 || p == 1 {
		return 0
	}
	return -p*math.Log(p) - (1-p)*math.Log1p(-p)
}
func (m *Model) Query(t Ticket, mode string) (Option, error) {
	if mode != "forecast" && mode != "uncertainty" && mode != "information" && mode != "falsification" && mode != "predictive" {
		return Option{}, errors.New("joint query mode")
	}
	q, p, e := m.conditional(t)
	if e != nil {
		return Option{}, e
	}
	out := Option{Observed: q}
	if mode == "forecast" {
		return out, nil
	}
	out.Uncertainty = entropy(q)
	if mode == "uncertainty" {
		return out, nil
	}
	a := bit(m.members[t.member].rows[t.ordinal].w1)
	if mode == "information" || mode == "falsification" {
		out.Information = out.Uncertainty
		for state, w := range p {
			den := m.factors[t.member][state][a]
			if w == 0 {
				continue
			}
			py := m.factors[t.member][state][2+2*a+1] / den
			out.Information -= w * entropy(py)
			term := -1.
			if q > 0 {
				term += py * py / q
			}
			if q < 1 {
				term += (1 - py) * (1 - py) / (1 - q)
			}
			out.EdgeCut += w * w * term
		}
		if out.Information < -2e-10 || out.EdgeCut < -2e-10 {
			return Option{}, errors.New("joint negative information")
		}
		out.Information = math.Max(0, out.Information)
		out.EdgeCut = math.Max(0, out.EdgeCut)
		return out, nil
	}
	before, _, e := m.Predict(t.member)
	if e != nil {
		return Option{}, e
	}
	means := [2]float64{before, before}
	for value := 0; value < 2; value++ {
		if (value == 1 && q == 0) || (value == 0 && q == 1) {
			continue
		}
		r := m.members[t.member].rows[t.ordinal]
		r.second, r.w2 = 2, value == 1
		scratch, e := m.prepare(t.member, t.ordinal, r)
		if e != nil {
			return Option{}, e
		}
		means[value], _, e = m.next(t.member, scratch[m.members[t.member].count-1])
		if e != nil {
			return Option{}, e
		}
	}
	if math.Abs(q*means[1]+(1-q)*means[0]-before) > 2e-10 {
		return Option{}, errors.New("joint predictive tower")
	}
	// Only this member's future law changes in the declared independent model.
	out.Value = (q*(means[1]-before)*(means[1]-before) + (1-q)*(means[0]-before)*(means[0]-before)) / float64(len(m.base))
	return out, nil
}
