// Isolated V60 lifecycle, mechanically retained from audited V58.
package researchwindowjournal

import (
	"errors"
	"math"
)

func (m *Model) Issue(i int, at int64) (Ticket, error) {
	if m.bankOwned {
		return Ticket{}, errors.New("journal bank mutation authority")
	}
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
		m.commit(&p)
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
	if m.bankOwned {
		return Receipt{}, errors.New("journal bank mutation authority")
	}
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
		m.commit(&p)
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
	if m.bankOwned {
		return errors.New("journal bank mutation authority")
	}
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
	if m.bankOwned {
		return errors.New("journal bank mutation authority")
	}
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
		py, pn := math.Exp(m.nodeAt(0, &yes).tree[eta]-old), math.Exp(m.nodeAt(0, &no).tree[eta]-old)
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
				one, _, e = m.predict(i, &yes, yes.weights)
				if e != nil {
					return Option{}, 0, e
				}
			}
			if q < 1 {
				two, _, e = m.predict(i, &no, no.weights)
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
	if m.bankOwned {
		return Ticket{}, errors.New("journal bank mutation authority")
	}
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
