package researchswitch

import "errors"

// ScoredHybridV51 changes only the served outer head. The original local/global
// heads and child learners retain their advice/update contracts. Nested loss
// weights do not assert multiple independent Bayesian uses of one observation.
type ScoredHybridV51 struct {
	local        *LocalPool
	global       *compactModelV48
	scope        *scoreModelV51
	globalHazard float64
	scopeHazard  float64
	scoreMode    string
}

type ScoredHybridTicketV51 struct {
	owner *ScoredHybridV51
	epoch uint64
	slot  int
	q     float64
}

func (t ScoredHybridTicketV51) Forecast() float64 { return t.q }

func NewScoredHybridV51(base []float64, cfg Config, memberTrials int, globalHazard, scopeHazard float64, epoch uint64, mode string) (*ScoredHybridV51, error) {
	globalCfg := cfg
	globalCfg.Hazard = globalHazard
	global, err := newCompactV48(compactConfigV48(globalCfg), epoch)
	if err != nil {
		return nil, err
	}
	scope, err := newScoreV51(scoreConfigV51{Prior: []float64{.9, .1}, Hazard: scopeHazard, Trials: cfg.Trials, Pending: cfg.Pending, Mode: mode}, epoch)
	if err != nil {
		return nil, err
	}
	local, err := newMemoLocalV49(base, cfg, memberTrials, epoch)
	if err != nil {
		return nil, err
	}
	return &ScoredHybridV51{local, global, scope, globalHazard, scopeHazard, mode}, nil
}

func (p *ScoredHybridV51) advice(member int) ([3]float64, [2]float64, error) {
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

func (p *ScoredHybridV51) Predict(member int) (float64, error) {
	_, heads, err := p.advice(member)
	if err != nil {
		return 0, err
	}
	return p.scope.Predict(heads[:])
}

// All three levels retain AS-OF original advice. An unexpected partial child
// failure fences the entire bundle; this is not cross-model atomic rollback.
func (p *ScoredHybridV51) Issue(member int, at int64) (ScoredHybridTicketV51, error) {
	l := p.local
	if l.poison || member < 0 || member >= len(l.base) || at < l.clock || l.used == len(l.rows) || l.pending == l.cfg.Pending || l.issued[member] == l.memberTrials {
		return ScoredHybridTicketV51{}, errors.New("invalid or capped scored issue")
	}
	q, heads, err := p.advice(member)
	if err != nil {
		return ScoredHybridTicketV51{}, err
	}
	if _, err = p.scope.Predict(heads[:]); err != nil {
		return ScoredHybridTicketV51{}, err
	}
	lt, err := l.Issue(member, at)
	var gt compactTicketV48
	var st scoreTicketV51
	if err == nil {
		gt, err = p.global.Issue(q[:], at)
	}
	if err == nil {
		st, err = p.scope.Issue(heads[:], at)
	}
	if err != nil {
		l.poison, l.clock = true, at
		return ScoredHybridTicketV51{}, err
	}
	if lt.Forecast() != heads[1] || gt.Forecast() != heads[0] || lt.slot != gt.slot || gt.slot != st.slot {
		l.poison, l.clock = true, at
		return ScoredHybridTicketV51{}, errors.New("scored advice or position changed")
	}
	return ScoredHybridTicketV51{p, l.epoch, lt.slot, st.Forecast()}, nil
}

func (p *ScoredHybridV51) validate(t ScoredHybridTicketV51, at int64) error {
	if t.owner != p || t.epoch != p.local.epoch {
		return errors.New("scored owner or epoch")
	}
	return p.local.validate(LocalTicket{owner: p.local, epoch: t.epoch, slot: t.slot}, at)
}

func (p *ScoredHybridV51) Resolve(t ScoredHybridTicketV51, useful bool, at int64) (PoolReceipt, error) {
	if err := p.validate(t, at); err != nil {
		return PoolReceipt{}, err
	}
	l := p.local
	r, err := l.Resolve(LocalTicket{owner: l, epoch: t.epoch, slot: t.slot}, useful, at)
	if err == nil {
		_, err = p.global.Resolve(compactTicketV48{owner: p.global, epoch: t.epoch, slot: t.slot}, useful, at)
	}
	var receipt scoreReceiptV51
	if err == nil {
		receipt, err = p.scope.Resolve(scoreTicketV51{owner: p.scope, epoch: t.epoch, slot: t.slot}, useful, at)
	}
	if err != nil {
		l.poison, l.clock = true, at
		return PoolReceipt{}, err
	}
	// The receipt is the served outer law, never an auxiliary head's forecast.
	r.Receipt = Receipt(receipt)
	return r, nil
}

func (p *ScoredHybridV51) Cancel(t ScoredHybridTicketV51, at int64) error {
	if err := p.validate(t, at); err != nil {
		return err
	}
	l := p.local
	err := l.Cancel(LocalTicket{owner: l, epoch: t.epoch, slot: t.slot}, at)
	if err == nil {
		err = p.global.Cancel(compactTicketV48{owner: p.global, epoch: t.epoch, slot: t.slot}, at)
	}
	if err == nil {
		err = p.scope.Cancel(scoreTicketV51{owner: p.scope, epoch: t.epoch, slot: t.slot}, at)
	}
	if err != nil {
		l.poison, l.clock = true, at
	}
	return err
}

func (p *ScoredHybridV51) Pending() int { return p.local.Pending() }

func (p *ScoredHybridV51) BeginEpoch(epoch uint64, at int64) error {
	l := p.local
	if epoch <= l.epoch || at < l.clock {
		return errors.New("invalid scored epoch")
	}
	next, err := NewScoredHybridV51(l.base, l.cfg, l.memberTrials, p.globalHazard, p.scopeHazard, epoch, p.scoreMode)
	if err != nil {
		return err
	}
	next.local.clock, next.global.clock, next.scope.clock = at, at, at
	*p = *next
	return nil
}
