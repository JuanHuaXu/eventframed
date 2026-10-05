package researchwindowjournal

import (
	"errors"
	law "github.com/JuanHuaXu/eventframed/internal/researchretentionlaw"
)

type Bank struct {
	models     [Experts]*Model
	selector   *Selector
	epoch      uint64
	clock      int64
	windows    [Experts]int
	depth, cap int
}
type BankTicket struct {
	owner    *Bank
	epoch    uint64
	base     [Experts]Ticket
	selector SelectorTicket
	second   bool
	forecast float64
}

func (t BankTicket) Forecast() float64 { return t.forecast }

func NewBank(base []float64, epoch uint64, cap, depth int, windows [Experts]int) (*Bank, error) {
	for k, w := range windows {
		if w < 1 || w > len(base)*MaxTrials || (k > 0 && w <= windows[k-1]) {
			return nil, errors.New("bank ordered windows")
		}
	}
	first, e := New(base, epoch, cap, Config{Depth: depth, Window: windows[0]})
	if e != nil {
		return nil, e
	}
	b := &Bank{epoch: epoch, windows: windows, depth: depth, cap: cap}
	first.bankOwned = true
	b.models[0] = first
	for k := 1; k < Experts; k++ {
		b.models[k], e = first.newSibling(windows[k], epoch, cap)
		if e != nil {
			return nil, e
		}
	}
	b.selector, e = newSelector(len(base), epoch)
	if e != nil {
		return nil, e
	}
	return b, nil
}
func (b *Bank) forecasts(i int) (Forecasts, error) {
	var joints [Experts]law.Joint
	for k, m := range b.models {
		j, e := m.LatentJoint(i)
		if e != nil {
			return Forecasts{}, e
		}
		joints[k] = j
	}
	f, e := law.Forecasts(joints)
	if e != nil {
		return Forecasts{}, e
	}
	var out Forecasts
	for k, v := range f {
		out[k] = Expert{Clean: v.Clean, Joint: v.Joint}
	}
	return out, nil
}
func (b *Bank) Predict(i int) (float64, float64, error) {
	f, e := b.forecasts(i)
	if e != nil {
		return 0, 0, e
	}
	w, e := b.selector.Weights(i)
	if e != nil {
		return 0, 0, e
	}
	q, o := 0., 0.
	for k, p := range w {
		q += p * f[k].Clean
		o += p * (f[k].Joint[2] + f[k].Joint[3])
	}
	return q, o, nil
}
func (b *Bank) Pending() int { return b.models[0].pending }

type issuePlan struct {
	correction      plan
	changed         bool
	clean, observed float64
	slot            int
}

func (m *Model) prepareIssue(i int, at int64) (issuePlan, error) {
	var p issuePlan
	if i < 0 || i >= len(m.base) || at < m.clock || m.pending >= m.cap || m.issued[i] >= MaxTrials {
		return p, errors.New("bank issue cap/time")
	}
	var e error
	p.clean, p.observed, e = m.Predict(i)
	if e != nil {
		return p, e
	}
	expire := m.seq + 1 - m.cfg.Window
	if expire > 0 {
		x := m.trials[m.seqSlots[expire-1]]
		old := category(x)
		if old >= 0 {
			p.correction, e = m.prepare(x.member, old, -1)
			if e != nil {
				return p, e
			}
			p.changed = true
		}
	}
	p.slot = i*MaxTrials + m.issued[i]
	return p, nil
}
func (b *Bank) Issue(i int, at int64) (BankTicket, error) {
	if at < b.clock {
		return BankTicket{}, errors.New("bank issue time")
	}
	f, e := b.forecasts(i)
	if e != nil {
		return BankTicket{}, e
	}
	var plans [Experts]issuePlan
	for k, m := range b.models {
		plans[k], e = m.prepareIssue(i, at)
		if e != nil {
			return BankTicket{}, e
		}
		if !m.bankOwned || m.seq != b.models[0].seq || plans[k].slot != plans[0].slot {
			return BankTicket{}, errors.New("journal issue synchronization")
		}
	}
	// No child is committed until the selector also accepts the issued joint.
	selected, e := b.selector.Issue(i, at, f)
	if e != nil {
		return BankTicket{}, e
	}
	t := BankTicket{owner: b, epoch: b.epoch, selector: selected, forecast: selected.Forecast()}
	for k, m := range b.models {
		p := &plans[k]
		if p.changed {
			m.commit(&p.correction)
		}
		m.seq++
		m.issued[i]++
		m.pending++
		m.clock = at
		t.base[k] = Ticket{owner: m, epoch: m.epoch, slot: p.slot, q: p.clean}
	}
	// Canonical metadata is the BANK-issued law, not a selected child's law.
	m := b.models[0]
	m.seqSlots[m.seq-1] = plans[0].slot
	r := b.selector.members[i].rows[selected.ordinal]
	m.trials[plans[0].slot] = trial{seq: m.seq, member: i, ordinal: m.issued[i], first: 1, at: at, clean: t.forecast, observed: r.issuedFirst}
	b.clock = at
	return t, nil
}
func (b *Bank) validate(t BankTicket, at int64) error {
	if t.owner != b || t.epoch != b.epoch || at < b.clock || t.second != t.selector.second {
		return errors.New("bank owner/epoch/time")
	}
	if e := b.selector.validate(t.selector, at); e != nil {
		return e
	}
	for k, m := range b.models {
		if !m.bankOwned || t.base[k].slot != t.base[0].slot {
			return errors.New("journal receipt synchronization")
		}
		if t.second != t.base[k].second {
			return errors.New("bank child phase")
		}
		if e := m.validate(t.base[k], at, false); e != nil {
			return e
		}
	}
	return nil
}
func (b *Bank) Resolve(t BankTicket, value bool, at int64) (SelectorReceipt, error) {
	if e := b.validate(t, at); e != nil {
		return SelectorReceipt{}, e
	}
	i, n := t.selector.selectorMember, t.selector.ordinal
	r := b.selector.members[i].rows[n]
	if t.second {
		r.second = 2
		r.w2 = value
	} else {
		r.first = 2
		r.w1 = value
	}
	scratch, e := b.selector.prepare(i, n, r)
	if e != nil {
		return SelectorReceipt{}, e
	}
	var plans [Experts]plan
	var updates [Experts]trial
	var changed [Experts]bool
	for k, m := range b.models {
		x := m.trials[t.base[k].slot]
		old := category(x)
		if t.second {
			x.second = 2
			x.w2 = value
		} else {
			x.first = 2
			x.w1 = value
		}
		updates[k] = x
		if x.seq > m.seq-m.cfg.Window {
			plans[k], e = m.prepare(x.member, old, category(x))
			if e != nil {
				return SelectorReceipt{}, e
			}
			changed[k] = true
		}
	}
	// From here all writes are infallible and the serial publication is atomic.
	for k, m := range b.models {
		if changed[k] {
			m.commit(&plans[k])
		}
		m.pending--
		m.clock = at
	}
	b.models[0].trials[t.base[0].slot] = updates[0]
	b.selector.members[i].rows[n] = r
	for j := n; j < b.selector.members[i].count; j++ {
		b.selector.members[i].rows[j].posterior = scratch[j]
	}
	b.selector.clock = at
	b.clock = at
	measurement, issuedAt, forecast := 1, r.at, r.issuedFirst
	if t.second {
		measurement, issuedAt, forecast = 2, r.requestAt, r.issuedSecond
	}
	return SelectorReceipt{i, n + 1, measurement, b.epoch, issuedAt, at, forecast, value}, nil
}
func (b *Bank) Cancel(t BankTicket, at int64) error {
	if e := b.validate(t, at); e != nil {
		return e
	}
	for _, m := range b.models {
		m.pending--
		m.clock = at
	}
	x := &b.models[0].trials[t.base[0].slot]
	if t.second {
		x.second = 3
	} else {
		x.first = 3
	}
	r := &b.selector.members[t.selector.selectorMember].rows[t.selector.ordinal]
	if t.second {
		r.second = 3
	} else {
		r.first = 3
	}
	b.selector.clock = at
	b.clock = at
	return nil
}
func (b *Bank) QuerySecond(t BankTicket) (float64, error) {
	if t.owner != b || t.epoch != b.epoch || t.second {
		return 0, errors.New("bank query origin")
	}
	return b.selector.QuerySecond(t.selector)
}
func (b *Bank) RequestSecond(t BankTicket, at int64) (BankTicket, error) {
	if at < b.clock {
		return BankTicket{}, errors.New("bank request time")
	}
	q, e := b.QuerySecond(t)
	if e != nil {
		return BankTicket{}, e
	}
	for k, m := range b.models {
		if !m.bankOwned || t.base[k].slot != t.base[0].slot {
			return BankTicket{}, errors.New("journal request synchronization")
		}
		if m.pending >= m.cap {
			return BankTicket{}, errors.New("bank request cap")
		}
		if e := m.validate(t.base[k], at, true); e != nil {
			return BankTicket{}, e
		}
		if m.trials[t.base[k].slot].second != 0 {
			return BankTicket{}, errors.New("bank child request replay")
		}
	}
	selected, e := b.selector.RequestSecond(t.selector, at)
	if e != nil {
		return BankTicket{}, e
	}
	t.selector = selected
	t.second = true
	t.forecast = q
	for k, m := range b.models {
		m.pending++
		m.clock = at
		t.base[k] = Ticket{owner: m, epoch: m.epoch, slot: t.base[k].slot, second: true, q: q}
	}
	x := &b.models[0].trials[t.base[0].slot]
	x.second, x.auditAt, x.auditForecast = 1, at, q
	b.clock = at
	return t, nil
}
func (b *Bank) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= b.epoch || at < b.clock {
		return errors.New("bank epoch")
	}
	next, e := NewBank(b.models[0].base, epoch, b.cap, b.depth, b.windows)
	if e != nil {
		return e
	}
	next.clock = at
	next.selector.clock = at
	for _, m := range next.models {
		m.clock = at
	}
	*b = *next
	return nil
}
