// Package researchtree tests a rolling paired-evidence feature-partition model.
// It is not a CTW suffix source, an Anti-Pigeon certificate, or production code.
package researchtree

import (
	"errors"
	"math"
	"sort"

	prior "github.com/JuanHuaXu/eventframed/internal/researchnoisemoment"
	paired "github.com/JuanHuaXu/eventframed/internal/researchpaired"
)

const Atoms, MaxDepth, MaxNodes, MaxTrials = 21, 7, 255, 64

var noise = [3]float64{0, .1, .2}
var noisePrior = [3]float64{.8, .1, .1}

type Config struct {
	Depth, Window int
	Strength      float64
}
type node struct {
	counts                 [6]int
	leaf, tree, mean, stop [3]float64
}
type trial struct {
	seq, member, ordinal           int
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

type Receipt = paired.Receipt
type Option = paired.Option
type PredictiveOption struct{ Observed, Value float64 }
type Model struct {
	base          []float64
	cfg           Config
	epoch         uint64
	clock         int64
	cap, pending  int
	seq           int
	issued, ranks []int
	trials        []trial
	seqSlots      []int
	logPrior      [MaxNodes][Atoms]float64
	priorMean     [MaxNodes]float64
	factors       [3][Atoms][6]float64
	nodes         [MaxNodes]node
	weights       [3]float64
}
type plan struct {
	nodes   [MaxNodes]node
	weights [3]float64
}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func atom(z int) float64    { return (float64(z) + .5) / Atoms }
func New(base []float64, epoch uint64, cap int, cfg Config) (*Model, error) {
	if len(base) < 2 || len(base) > 200 || epoch == 0 || cap < 1 || cap > 2*len(base)*MaxTrials || cfg.Depth < 0 || cfg.Depth > MaxDepth || cfg.Window < 1 || cfg.Window > len(base)*MaxTrials || !finite(cfg.Strength) || cfg.Strength <= 0 || cfg.Strength > 32 {
		return nil, errors.New("tree contract")
	}
	for _, b := range base {
		if !finite(b) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("tree baseline")
		}
	}
	m := &Model{base: append([]float64(nil), base...), cfg: cfg, epoch: epoch, cap: cap, issued: make([]int, len(base)), ranks: make([]int, len(base)), trials: make([]trial, len(base)*MaxTrials), seqSlots: make([]int, len(base)*MaxTrials)}
	order := make([]int, len(base))
	total := 0.
	for i := range base {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool { return base[order[i]] < base[order[j]] })
	var sums [MaxNodes]float64
	var sizes [MaxNodes]int
	firstEqual := 0
	for rank, i := range order {
		total += base[i]
		if rank == 0 || base[i] != base[order[rank-1]] {
			firstEqual = rank
		}
		// Equal public features cannot acquire an arbitrary ID-dependent split.
		m.ranks[i] = firstEqual * (1 << cfg.Depth) / len(base)
		for n := ((1 << cfg.Depth) - 1) + m.ranks[i]; ; n = (n - 1) / 2 {
			sums[n] += base[i]
			sizes[n]++
			if n == 0 {
				break
			}
		}
	}
	for n := 0; n < (1<<(cfg.Depth+1))-1; n++ {
		mu := total / float64(len(base))
		if sizes[n] > 0 {
			mu = sums[n] / float64(sizes[n])
		}
		p, e := prior.DiscretePrior(mu, cfg.Strength, "moment")
		if e != nil {
			return nil, e
		}
		for z, v := range p {
			m.logPrior[n][z] = math.Log(v)
			m.priorMean[n] += v * atom(z)
		}
	}
	for eta, v := range noise {
		for z := 0; z < Atoms; z++ {
			p := atom(z)
			q := v + (1-2*v)*p
			m.factors[eta][z][0], m.factors[eta][z][1] = math.Log(1-q), math.Log(q)
			for a := 0; a < 2; a++ {
				for b := 0; b < 2; b++ {
					one, two := v, v
					if a == 1 {
						one = 1 - v
					}
					if b == 1 {
						two = 1 - v
					}
					m.factors[eta][z][2+2*a+b] = math.Log(p*one*two + (1-p)*(1-one)*(1-two))
				}
			}
		}
	}
	for n := (1 << (cfg.Depth + 1)) - 2; n >= 0; n-- {
		if e := m.refresh(&m.nodes, n); e != nil {
			return nil, e
		}
	}
	var e error
	m.weights, e = m.normalize(&m.nodes)
	return m, e
}
func logAdd(a, b float64) float64 {
	if math.IsInf(a, -1) {
		return b
	}
	if math.IsInf(b, -1) {
		return a
	}
	hi, lo := math.Max(a, b), math.Min(a, b)
	return hi + math.Log1p(math.Exp(lo-hi))
}
func (m *Model) refresh(nodes *[MaxNodes]node, n int) error {
	x := &nodes[n]
	for eta := 0; eta < 3; eta++ {
		maximum := math.Inf(-1)
		var terms [Atoms]float64
		for z := range terms {
			v := m.logPrior[n][z]
			if math.IsNaN(v) || math.IsInf(v, 1) {
				return errors.New("tree prior fault")
			}
			for k, count := range x.counts {
				if count < 0 || count > m.cfg.Window {
					return errors.New("tree count fault")
				}
				if count > 0 {
					v += float64(count) * m.factors[eta][z][k]
				}
			}
			if math.IsNaN(v) || math.IsInf(v, 1) {
				return errors.New("tree evidence fault")
			}
			terms[z] = v
			maximum = math.Max(maximum, v)
		}
		x.leaf[eta], x.mean[eta] = math.Inf(-1), m.priorMean[n]
		if finite(maximum) {
			sum, mean := 0., 0.
			for z, v := range terms {
				v = math.Exp(v - maximum)
				sum += v
				mean += v * atom(z)
			}
			if !finite(sum) || sum <= 0 {
				return errors.New("tree leaf normalization")
			}
			x.leaf[eta], x.mean[eta] = maximum+math.Log(sum), mean/sum
		}
		x.tree[eta], x.stop[eta] = x.leaf[eta], 1
		if n < (1<<m.cfg.Depth)-1 {
			leaf := x.leaf[eta] - math.Ln2
			split := nodes[2*n+1].tree[eta] + nodes[2*n+2].tree[eta] - math.Ln2
			x.tree[eta] = logAdd(leaf, split)
			if finite(x.tree[eta]) {
				x.stop[eta] = math.Exp(leaf - x.tree[eta])
			}
		}
		if math.IsNaN(x.tree[eta]) || math.IsInf(x.tree[eta], 1) || !finite(x.mean[eta]) || !finite(x.stop[eta]) {
			return errors.New("tree node fault")
		}
	}
	return nil
}
func (m *Model) normalize(nodes *[MaxNodes]node) ([3]float64, error) {
	var w [3]float64
	maximum := math.Inf(-1)
	for eta := range w {
		w[eta] = math.Log(noisePrior[eta]) + nodes[0].tree[eta]
		if math.IsNaN(w[eta]) || math.IsInf(w[eta], 1) {
			return w, errors.New("tree root fault")
		}
		maximum = math.Max(maximum, w[eta])
	}
	if !finite(maximum) {
		return w, errors.New("tree root support")
	}
	sum := 0.
	for eta := range w {
		w[eta] = math.Exp(w[eta] - maximum)
		sum += w[eta]
	}
	if !finite(sum) || sum <= 0 {
		return w, errors.New("tree root normalization")
	}
	for eta := range w {
		w[eta] /= sum
	}
	return w, nil
}
func (m *Model) predict(i int, nodes *[MaxNodes]node, weights [3]float64) (float64, float64, error) {
	if i < 0 || i >= len(m.base) {
		return 0, 0, errors.New("tree member")
	}
	q, o := 0., 0.
	for eta, w := range weights {
		if w == 0 {
			continue
		}
		n := ((1 << m.cfg.Depth) - 1) + m.ranks[i]
		v := nodes[n].mean[eta]
		for n > 0 {
			n = (n - 1) / 2
			x := nodes[n]
			v = x.stop[eta]*x.mean[eta] + (1-x.stop[eta])*v
		}
		q += w * v
		o += w * (noise[eta] + (1-2*noise[eta])*v)
	}
	if !finite(q) || !finite(o) || q <= 0 || q >= 1 || o <= 0 || o >= 1 {
		return 0, 0, errors.New("tree forecast")
	}
	return q, o, nil
}
func (m *Model) Predict(i int) (float64, float64, error) { return m.predict(i, &m.nodes, m.weights) }
func (m *Model) NoiseWeights() [3]float64                { return m.weights }
func (m *Model) Pending() int                            { return m.pending }
func category(x trial) int {
	if x.first != 2 {
		return -1
	}
	a, b := 0, 0
	if x.w1 {
		a = 1
	}
	if x.w2 {
		b = 1
	}
	if x.second == 2 {
		return 2 + 2*a + b
	}
	return a
}

// prepare replaces one original outcome factor on its feature path. Full node
// scratch makes failed arithmetic atomic; no trial or posterior is committed.
func (m *Model) prepare(i, old, next int) (plan, error) {
	p := plan{nodes: m.nodes}
	for n := ((1 << m.cfg.Depth) - 1) + m.ranks[i]; ; n = (n - 1) / 2 {
		if old >= 0 {
			p.nodes[n].counts[old]--
		}
		if next >= 0 {
			p.nodes[n].counts[next]++
		}
		if e := m.refresh(&p.nodes, n); e != nil {
			return p, e
		}
		if n == 0 {
			break
		}
	}
	var e error
	p.weights, e = m.normalize(&p.nodes)
	return p, e
}
func (m *Model) Issue(i int, at int64) (Ticket, error) {
	if i < 0 || i >= len(m.base) || at < m.clock || m.pending >= m.cap || m.issued[i] >= MaxTrials {
		return Ticket{}, errors.New("tree issue cap/time")
	}
	q, o, e := m.Predict(i)
	if e != nil {
		return Ticket{}, e
	}
	var p plan
	changed := false
	expire := m.seq + 1 - m.cfg.Window
	if expire > 0 {
		x := m.trials[m.seqSlots[expire-1]]
		k := category(x)
		if k >= 0 {
			p, e = m.prepare(x.member, k, -1)
			if e != nil {
				return Ticket{}, e
			}
			changed = true
		}
	}
	if changed {
		m.nodes, m.weights = p.nodes, p.weights
	}
	slot := i*MaxTrials + m.issued[i]
	m.seq++
	m.seqSlots[m.seq-1] = slot
	m.trials[slot] = trial{seq: m.seq, member: i, ordinal: m.issued[i] + 1, first: 1, at: at, clean: q, observed: o}
	m.issued[i]++
	m.pending++
	m.clock = at
	return Ticket{owner: m, epoch: m.epoch, slot: slot, q: q}, nil
}
func (m *Model) validate(t Ticket, at int64, known bool) error {
	if t.owner != m || t.epoch != m.epoch || t.slot < 0 || t.slot >= len(m.trials) || at < m.clock {
		return errors.New("tree owner/epoch/time")
	}
	x := m.trials[t.slot]
	want := uint8(1)
	if known {
		want = 2
	}
	if t.second {
		if x.second != 1 || at < x.auditAt {
			return errors.New("tree second replay/time")
		}
	} else if x.first != want || at < x.at {
		return errors.New("tree first replay/time")
	}
	return nil
}
func (m *Model) Resolve(t Ticket, value bool, at int64) (Receipt, error) {
	if e := m.validate(t, at, false); e != nil {
		return Receipt{}, e
	}
	x := m.trials[t.slot]
	old := category(x)
	if t.second {
		x.second, x.w2 = 2, value
	} else {
		x.first, x.w1 = 2, value
	}
	if x.seq > m.seq-m.cfg.Window {
		p, e := m.prepare(x.member, old, category(x))
		if e != nil {
			return Receipt{}, e
		}
		m.nodes, m.weights = p.nodes, p.weights
	}
	m.trials[t.slot] = x
	m.pending--
	m.clock = at
	measurement, issuedAt, forecast := 1, x.at, x.clean
	if t.second {
		measurement, issuedAt, forecast = 2, x.auditAt, x.auditForecast
	}
	return Receipt{Member: x.member, Ordinal: x.ordinal, Measurement: measurement, Epoch: m.epoch, IssuedAt: issuedAt, ArrivedAt: at, Forecast: forecast, Value: value}, nil
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
func (m *Model) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= m.epoch || at < m.clock {
		return errors.New("tree epoch")
	}
	next, e := New(m.base, epoch, m.cap, m.cfg)
	if e != nil {
		return e
	}
	next.clock = at
	*m = *next
	return nil
}
func entropy(q float64) float64 {
	if q == 0 || q == 1 {
		return 0
	}
	return -q*math.Log(q) - (1-q)*math.Log1p(-q)
}
func (m *Model) Query(t Ticket, mode string) (Option, error) {
	o, _, e := m.query(t, mode, nil)
	return o, e
}
func (m *Model) query(t Ticket, mode string, targets []float64) (Option, float64, error) {
	if mode != "forecast" && mode != "uncertainty" && mode != "information" && mode != "falsification" && mode != "predictive" {
		return Option{}, 0, errors.New("tree query mode")
	}
	if t.second || t.owner != m || t.epoch != m.epoch || t.slot < 0 || t.slot >= len(m.trials) {
		return Option{}, 0, errors.New("tree query origin")
	}
	x := m.trials[t.slot]
	if x.first != 2 || x.second != 0 {
		return Option{}, 0, errors.New("tree query availability")
	}
	// Expired outcomes are outside this suffix model: no artificial update or
	// extrapolation of a no-longer-bound likelihood is permitted.
	if x.seq <= m.seq-m.cfg.Window {
		return Option{}, 0, errors.New("tree expired origin")
	}
	x.second, x.w2 = 2, true
	yes, e := m.prepare(x.member, category(m.trials[t.slot]), category(x))
	if e != nil {
		return Option{}, 0, e
	}
	x.w2 = false
	no, e := m.prepare(x.member, category(m.trials[t.slot]), category(x))
	if e != nil {
		return Option{}, 0, e
	}
	q, h := 0., 0.
	var probabilities [3]float64
	for eta, w := range m.weights {
		if w == 0 {
			continue
		}
		old := m.nodes[0].tree[eta]
		py, pn := math.Exp(yes.nodes[0].tree[eta]-old), math.Exp(no.nodes[0].tree[eta]-old)
		if !finite(py) || !finite(pn) || py < 0 || pn < 0 || math.Abs(py+pn-1) > 2e-10 {
			return Option{}, 0, errors.New("tree conditional normalization")
		}
		py = math.Min(1, py)
		probabilities[eta] = py
		q += w * py
		h += w * entropy(py)
	}
	if !finite(q) || q < 0 || q > 1+2e-10 {
		return Option{}, 0, errors.New("tree option forecast")
	}
	q = math.Min(1, q)
	o := Option{Observed: q}
	if mode == "forecast" {
		return o, 0, nil
	}
	if mode == "predictive" {
		value := 0.
		for i, target := range targets {
			before, _, e := m.Predict(i)
			if e != nil {
				return Option{}, 0, e
			}
			one, two := before, before
			if q > 0 {
				one, _, e = m.predict(i, &yes.nodes, yes.weights)
				if e != nil {
					return Option{}, 0, e
				}
			}
			if q < 1 {
				two, _, e = m.predict(i, &no.nodes, no.weights)
				if e != nil {
					return Option{}, 0, e
				}
			}
			if math.Abs(q*one+(1-q)*two-before) > 2e-10 {
				return Option{}, 0, errors.New("tree predictive tower")
			}
			value += target * (q*(one-before)*(one-before) + (1-q)*(two-before)*(two-before))
		}
		if !finite(value) || value < 0 || value > .25+2e-10 {
			return Option{}, 0, errors.New("tree predictive value")
		}
		return o, value, nil
	}
	u := entropy(q)
	if mode == "uncertainty" {
		o.Uncertainty = u
		return o, 0, nil
	}
	if mode == "information" {
		o.Information = u - h
		if o.Information < -2e-10 {
			return Option{}, 0, errors.New("tree negative information")
		}
		o.Information = math.Max(0, o.Information)
		o.Uncertainty = u
		return o, 0, nil
	}
	edge := 0.
	for eta, w := range m.weights {
		py := probabilities[eta]
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
		return Option{}, 0, errors.New("tree negative concentration")
	}
	o.EdgeCut = math.Max(0, edge)
	return o, 0, nil
}
func (m *Model) PredictionValues(origins []Ticket, targets []float64) ([]PredictiveOption, error) {
	if len(origins) == 0 || len(origins) > len(m.base) || len(targets) != len(m.base) {
		return nil, errors.New("tree value shape")
	}
	sum := 0.
	for _, v := range targets {
		if !finite(v) || v < 0 || v > 1 {
			return nil, errors.New("tree value weight")
		}
		sum += v
	}
	if math.Abs(sum-1) > 1e-12 {
		return nil, errors.New("tree value normalization")
	}
	seen := map[int]bool{}
	out := make([]PredictiveOption, len(origins))
	for i, t := range origins {
		if seen[t.slot] {
			return nil, errors.New("tree duplicate value origin")
		}
		seen[t.slot] = true
		o, v, e := m.query(t, "predictive", targets)
		if e != nil {
			return nil, e
		}
		out[i] = PredictiveOption{Observed: o.Observed, Value: v}
	}
	return out, nil
}
func (m *Model) RequestAudit(t Ticket, at int64) (Ticket, error) {
	if t.second || m.pending >= m.cap {
		return Ticket{}, errors.New("tree audit cap/type")
	}
	if e := m.validate(t, at, true); e != nil {
		return Ticket{}, e
	}
	o, e := m.Query(t, "forecast")
	if e != nil {
		return Ticket{}, e
	}
	m.trials[t.slot].second = 1
	m.trials[t.slot].auditAt = at
	m.trials[t.slot].auditForecast = o.Observed
	m.pending++
	m.clock = at
	return Ticket{owner: m, epoch: m.epoch, slot: t.slot, second: true, q: o.Observed}, nil
}
