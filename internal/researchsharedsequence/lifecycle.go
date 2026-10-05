package researchsharedsequence

import "errors"

func (m *Model) Issue(i int, at int64) (Ticket, error) {
	if i < 0 || i >= len(m.base) || at < m.clock || m.pending >= m.cap || m.count >= len(m.rows) || m.members[i].count >= MaxTrials {
		return Ticket{}, errors.New("shared issue cap/time")
	}
	q, o, e := m.Predict(i)
	if e != nil {
		return Ticket{}, e
	}
	x := &m.members[i]
	p := x.latest
	if m.needsLocal() && x.count > 0 {
		for h := 0; h < 3; h++ {
			p.p[h], _, e = normalizeLocal(m.localTransition(i, p.p[h]))
			if e != nil {
				return Ticket{}, e
			}
		}
	}
	next := m.latest
	if m.needsShared() {
		if m.count > 0 {
			next.p = m.sharedTransition(next.p)
		}
		next.p, _, e = normalizeShared(next.p)
		if e != nil {
			return Ticket{}, e
		}
	}
	n := m.count
	if m.needsShared() && n%Block == 0 {
		m.checkpoints[n/Block] = m.latest
	}
	m.rows[n] = row{member: i, ordinal: x.count, at: at, clean: q, observed: o, first: 1}
	x.slots[x.count] = n
	x.count++
	x.latest = p
	m.latest = next
	m.count++
	m.pending++
	m.clock = at
	return Ticket{owner: m, epoch: m.epoch, slot: n, forecast: q}, nil
}
func (m *Model) Resolve(t Ticket, value bool, at int64) (Receipt, error) {
	if e := m.validate(t, at, false); e != nil {
		return Receipt{}, e
	}
	r := m.rows[t.slot]
	if t.second {
		r.second, r.b = 2, value
	} else {
		r.first, r.a = 2, value
	}
	var local conditional
	var shared sharedPrepared
	var e error
	if m.needsLocal() {
		local, e = m.prepareLocal(r.member, t.slot, r)
		if e != nil {
			return Receipt{}, e
		}
		if _, _, e = weights3(local.log, noisePrior); e != nil {
			return Receipt{}, e
		}
	}
	if m.needsShared() {
		shared, e = m.prepareShared(t.slot, r)
		if e != nil {
			return Receipt{}, e
		}
	}
	var ls *conditional
	var ss *checkpoint
	if m.needsLocal() {
		ls = &local
	}
	if m.needsShared() {
		ss = &shared.latest
	}
	if _, e = m.currentView(r.member, ls, ss); e != nil {
		return Receipt{}, e
	}
	m.rows[t.slot] = r
	if m.needsLocal() {
		m.members[r.member].latest = local
	}
	if m.needsShared() {
		m.latest = shared.latest
		for j := shared.first; j <= shared.last; j++ {
			m.checkpoints[j] = shared.blocks[j]
		}
	}
	m.pending--
	m.clock = at
	measurement, issued, forecast := 1, r.at, r.observed
	if t.second {
		measurement, issued, forecast = 2, r.auditAt, r.auditForecast
	}
	return Receipt{r.member, r.ordinal + 1, measurement, m.epoch, issued, at, forecast, value}, nil
}
func (m *Model) Cancel(t Ticket, at int64) error {
	if e := m.validate(t, at, false); e != nil {
		return e
	}
	r := &m.rows[t.slot]
	if t.second {
		r.second = 3
	} else {
		r.first = 3
	}
	m.pending--
	m.clock = at
	return nil
}
func (m *Model) RequestSecond(t Ticket, at int64) (Ticket, error) {
	if t.second || m.pending >= m.cap {
		return Ticket{}, errors.New("shared second cap/type")
	}
	if e := m.validate(t, at, true); e != nil {
		return Ticket{}, e
	}
	o, e := m.Query(t, "forecast")
	if e != nil {
		return Ticket{}, e
	}
	r := &m.rows[t.slot]
	r.second, r.auditAt, r.auditForecast = 1, at, o.Observed
	m.pending++
	m.clock = at
	t.second, t.forecast = true, o.Observed
	return t, nil
}
