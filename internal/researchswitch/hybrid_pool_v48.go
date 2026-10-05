package researchswitch

import "errors"

// HybridPool is an isolated aggregation strategy, NOT a joint posterior that
// treats child/head/scope uses of one label as independent observations.
// Local and global heads share one set of all-visible child learners.
type HybridPool struct {
	local         *LocalPool
	global, scope *compactModelV48
	globalHazard  float64
	scopeHazard   float64
}

type HybridTicket struct {
	owner *HybridPool
	epoch uint64
	slot  int
	q     float64
}

func (t HybridTicket) Forecast() float64 { return t.q }

func NewHybridPool(base []float64, cfg Config, memberTrials int, globalHazard, scopeHazard float64, epoch uint64) (*HybridPool, error) {
	globalCfg := cfg
	globalCfg.Hazard = globalHazard
	global, err := newCompactV48(compactConfigV48(globalCfg), epoch)
	if err != nil {
		return nil, err
	}
	scope, err := newCompactV48(compactConfigV48{Prior: []float64{.9, .1}, Hazard: scopeHazard, Trials: cfg.Trials, Pending: cfg.Pending}, epoch)
	if err != nil {
		return nil, err
	}
	local, err := NewLocalPool(base, cfg, memberTrials, epoch)
	if err != nil {
		return nil, err
	}
	return &HybridPool{local, global, scope, globalHazard, scopeHazard}, nil
}

func (p *HybridPool) advice(member int) ([3]float64, [2]float64, error) {
	q, err := p.local.advice(member)
	var heads [2]float64
	if err != nil {
		return q, heads, err
	}
	heads[0], err = p.global.Predict(q[:])
	if err == nil {
		heads[1], err = p.local.mixes[member].Predict(q[:])
	}
	return q, heads, err
}

func (p *HybridPool) Predict(member int) (float64, error) {
	_, heads, err := p.advice(member)
	if err != nil {
		return 0, err
	}
	return p.scope.Predict(heads[:])
}

// Advice at ALL levels is taken before mutating any learner and retained at
// the original nomination position. Unexpected partial failures fence all
// heads via the local owner; there is no cross-model rollback assertion.
func (p *HybridPool) Issue(member int, at int64) (HybridTicket, error) {
	l := p.local
	if l.poison || member < 0 || member >= len(l.base) || at < l.clock || l.used == len(l.rows) || l.pending == l.cfg.Pending || l.issued[member] == l.memberTrials {
		return HybridTicket{}, errors.New("invalid or capped hybrid issue")
	}
	q, heads, err := p.advice(member)
	if err != nil {
		return HybridTicket{}, err
	}
	if _, err = p.scope.Predict(heads[:]); err != nil {
		return HybridTicket{}, err
	}
	lt, err := l.Issue(member, at)
	var gt, st compactTicketV48
	if err == nil {
		gt, err = p.global.Issue(q[:], at)
	}
	if err == nil {
		st, err = p.scope.Issue(heads[:], at)
	}
	if err != nil {
		l.poison, l.clock = true, at
		return HybridTicket{}, err
	}
	if lt.Forecast() != heads[1] || gt.Forecast() != heads[0] || lt.slot != gt.slot || gt.slot != st.slot {
		l.poison, l.clock = true, at
		return HybridTicket{}, errors.New("hybrid advice or nomination position changed")
	}
	return HybridTicket{p, l.epoch, lt.slot, st.Forecast()}, nil
}

func (p *HybridPool) validate(t HybridTicket, at int64) error {
	if t.owner != p || t.epoch != p.local.epoch {
		return errors.New("hybrid owner or epoch")
	}
	return p.local.validate(LocalTicket{owner: p.local, epoch: t.epoch, slot: t.slot}, at)
}

func (p *HybridPool) Resolve(t HybridTicket, useful bool, at int64) (PoolReceipt, error) {
	if err := p.validate(t, at); err != nil {
		return PoolReceipt{}, err
	}
	l := p.local
	r, err := l.Resolve(LocalTicket{owner: l, epoch: t.epoch, slot: t.slot}, useful, at)
	if err == nil {
		_, err = p.global.Resolve(compactTicketV48{owner: p.global, epoch: t.epoch, slot: t.slot}, useful, at)
	}
	var receipt compactReceiptV48
	if err == nil {
		receipt, err = p.scope.Resolve(compactTicketV48{owner: p.scope, epoch: t.epoch, slot: t.slot}, useful, at)
	}
	if err != nil {
		l.poison, l.clock = true, at
		return PoolReceipt{}, err
	}
	// Auxiliary local/global head scores do not replace the served scope law.
	r.Receipt = Receipt(receipt)
	return r, nil
}

func (p *HybridPool) Cancel(t HybridTicket, at int64) error {
	if err := p.validate(t, at); err != nil {
		return err
	}
	l := p.local
	err := l.Cancel(LocalTicket{owner: l, epoch: t.epoch, slot: t.slot}, at)
	if err == nil {
		err = p.global.Cancel(compactTicketV48{owner: p.global, epoch: t.epoch, slot: t.slot}, at)
	}
	if err == nil {
		err = p.scope.Cancel(compactTicketV48{owner: p.scope, epoch: t.epoch, slot: t.slot}, at)
	}
	if err != nil {
		l.poison, l.clock = true, at
	}
	return err
}

func (p *HybridPool) Pending() int { return p.local.Pending() }

func (p *HybridPool) BeginEpoch(epoch uint64, at int64) error {
	l := p.local
	if epoch <= l.epoch || at < l.clock {
		return errors.New("invalid hybrid epoch")
	}
	next, err := NewHybridPool(l.base, l.cfg, l.memberTrials, p.globalHazard, p.scopeHazard, epoch)
	if err != nil {
		return err
	}
	next.local.clock, next.global.clock, next.scope.clock = at, at, at
	*p = *next
	return nil
}
