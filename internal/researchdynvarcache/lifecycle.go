package researchdynvarcache

import "errors"

func (m *Model) validate(t Ticket, at int64, known bool) error {
	if t.owner != m || t.epoch != m.epoch || t.slot < 0 || t.slot >= m.count || at < m.clock {
		return errors.New("dispersion owner/epoch/time")
	}
	r := m.rows[t.slot]
	if t.second {
		if r.second != 1 || at < r.auditAt {
			return errors.New("dispersion second replay")
		}
	} else {
		want := uint8(1)
		if known {
			want = 2
		}
		if r.first != want || at < r.at {
			return errors.New("dispersion first replay")
		}
	}
	return nil
}
func (m *Model) Issue(i int, at int64) (Ticket, error) {
	if i < 0 || i >= len(m.base) || at < m.clock || m.pending >= m.cap || m.count >= len(m.rows) || m.members[i].count >= MaxTrials {
		return Ticket{}, errors.New("dispersion issue cap/time")
	}
	q, o, e := m.Predict(i)
	if e != nil {
		return Ticket{}, e
	}
	x := &m.members[i]
	next := x.latest
	if x.count > 0 {
		for k := range next.p {
			next.p[k], _, e = normalize(m.transition(i, k/3, next.p[k]))
			if e != nil {
				return Ticket{}, e
			}
		}
	}
	d, e := m.derive(i, next)
	if e != nil {
		return Ticket{}, e
	}
	n := m.count
	m.rows[n] = row{member: i, ordinal: x.count, at: at, clean: q, observed: o, first: 1}
	x.slots[x.count] = n
	x.count++
	x.latest = next
	x.derived = d
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
	next, e := m.prepare(r.member, t.slot, r)
	if e != nil {
		return Receipt{}, e
	}
	d, e := m.derive(r.member, next)
	if e != nil {
		return Receipt{}, e
	}
	w, e := m.rebuildOdds(r.member, &next)
	if e != nil {
		return Receipt{}, e
	}
	// Publish the revised factor and its full dependent posterior together.
	m.rows[t.slot] = r
	m.members[r.member].latest = next
	m.members[r.member].derived = d
	m.odds = w
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
		return Ticket{}, errors.New("dispersion second cap/type")
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
func (m *Model) Pending() int { return m.pending }
func (m *Model) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= m.epoch || at < m.clock {
		return errors.New("dispersion epoch")
	}
	next, e := New(m.base, epoch, m.cap, m.cfg)
	if e != nil {
		return e
	}
	next.clock = at
	*m = *next
	return nil
}
