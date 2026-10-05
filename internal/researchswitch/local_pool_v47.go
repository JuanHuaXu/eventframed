package researchswitch

import (
	"errors"

	"github.com/JuanHuaXu/eventframed/internal/researchdispersion"
	"github.com/JuanHuaXu/eventframed/internal/researchmoment"
)

// LocalPool is an isolated research candidate. Only the expert-state filter
// is task-local; all three child learners still see every revealed outcome.
// A task is an observable member, never an oracle regime or future label.
type LocalPool struct {
	full, adaptive *researchdispersion.WindowObserver
	moment         *researchmoment.Model
	mixes          []*Model
	base           []float64
	rows           []poolRow
	issued         []int
	cfg            Config
	memberTrials   int
	used, pending  int
	clock          int64
	epoch          uint64
	poison         bool
}

type LocalTicket struct {
	owner *LocalPool
	epoch uint64
	slot  int
	q     float64
}

func (t LocalTicket) Forecast() float64 { return t.q }

// Global admission and per-task history limits are distinct. In particular,
// allocating the global Trials bound for EVERY task is not the stated budget.
func NewLocalPool(base []float64, cfg Config, memberTrials int, epoch uint64) (*LocalPool, error) {
	if len(cfg.Prior) != 3 || len(base) < 2 || len(base) > 200 || memberTrials < 1 || memberTrials > researchdispersion.MaxTrials || cfg.Trials < 1 || cfg.Trials > MaxTrials || cfg.Trials > len(base)*memberTrials || cfg.Pending < 1 || cfg.Pending > cfg.Trials {
		return nil, errors.New("invalid local pool caps or arity")
	}
	local := cfg
	local.Trials, local.Pending = memberTrials, min(cfg.Pending, memberTrials)
	first, err := New(local, epoch)
	if err != nil {
		return nil, err
	}
	full, err := researchdispersion.NewWindowObserver(base, epoch, cfg.Pending, "full")
	if err != nil {
		return nil, err
	}
	adaptive, err := researchdispersion.NewWindowObserver(base, epoch, cfg.Pending, "adaptive")
	if err != nil {
		return nil, err
	}
	moment, err := researchmoment.New(base, epoch, cfg.Pending, researchmoment.Config{Family: "rich", Prior: "moment", Strength: 2, Hazard: 1. / 16, Shared: true})
	if err != nil {
		return nil, err
	}
	mixes := make([]*Model, len(base))
	mixes[0] = first
	for i := 1; i < len(base); i++ {
		mixes[i], err = New(local, epoch)
		if err != nil {
			return nil, err
		}
	}
	cfg.Prior = append([]float64(nil), cfg.Prior...)
	return &LocalPool{full: full, adaptive: adaptive, moment: moment, mixes: mixes, base: append([]float64(nil), base...), rows: make([]poolRow, cfg.Trials), issued: make([]int, len(base)), cfg: cfg, memberTrials: memberTrials, epoch: epoch}, nil
}

func (p *LocalPool) advice(member int) ([3]float64, error) {
	var q [3]float64
	if p.poison || member < 0 || member >= len(p.base) {
		return q, errors.New("poisoned local pool or unknown member")
	}
	var err error
	q[0], err = p.full.Predict(member)
	if err != nil {
		return q, err
	}
	q[1], err = p.adaptive.Predict(member)
	if err != nil {
		return q, err
	}
	q[2], err = p.moment.Predict(member)
	return q, err
}

func (p *LocalPool) Predict(member int) (float64, error) {
	q, err := p.advice(member)
	if err != nil {
		return 0, err
	}
	return p.mixes[member].Predict(q[:])
}

// Only Issue advances this member's Markov clock. Other tasks can alter the
// shared expert advice, but cannot age/reset this task's expert-state weights.
func (p *LocalPool) Issue(member int, at int64) (LocalTicket, error) {
	if p.poison || member < 0 || member >= len(p.base) || at < p.clock || p.used == len(p.rows) || p.pending == p.cfg.Pending || p.issued[member] == p.memberTrials {
		return LocalTicket{}, errors.New("invalid or capped local issue")
	}
	q, err := p.advice(member)
	if err != nil {
		return LocalTicket{}, err
	}
	if _, err = p.mixes[member].Predict(q[:]); err != nil {
		return LocalTicket{}, err
	}
	r := poolRow{member: member, ordinal: p.issued[member] + 1, at: at, status: 1}
	r.full, err = p.full.Issue(member, at)
	if err == nil {
		r.adaptive, err = p.adaptive.Issue(member, at)
	}
	if err == nil {
		r.moment, err = p.moment.Issue(member, at)
	}
	if err == nil {
		r.mix, err = p.mixes[member].Issue(q[:], at)
	}
	if err != nil {
		p.poison, p.clock = true, at
		return LocalTicket{}, err
	}
	if r.full.Forecast() != q[0] || r.adaptive.Forecast() != q[1] || r.moment.Forecast() != q[2] {
		p.poison, p.clock = true, at
		return LocalTicket{}, errors.New("expert changed during local issue")
	}
	slot := p.used
	p.rows[slot] = r
	p.used++
	p.pending++
	p.issued[member]++
	p.clock = at
	return LocalTicket{p, p.epoch, slot, r.mix.Forecast()}, nil
}

func (p *LocalPool) validate(t LocalTicket, at int64) error {
	if p.poison || t.owner != p || t.epoch != p.epoch || t.slot < 0 || t.slot >= p.used || at < p.clock || at < p.rows[t.slot].at || p.rows[t.slot].status != 1 {
		return errors.New("invalid local owner epoch time or status")
	}
	return nil
}

// Delayed labels replay at the ORIGINAL task-local position. The returned
// receipt additionally binds the original GLOBAL nomination ordinal.
// Unexpected partial child failure fences the entire bundle, not just a task.
func (p *LocalPool) Resolve(t LocalTicket, useful bool, at int64) (PoolReceipt, error) {
	if err := p.validate(t, at); err != nil {
		return PoolReceipt{}, err
	}
	r := p.rows[t.slot]
	_, err := p.full.Resolve(r.full, useful, at)
	if err == nil {
		_, err = p.adaptive.Resolve(r.adaptive, useful, at)
	}
	if err == nil {
		_, err = p.moment.Resolve(r.moment, useful, at)
	}
	var receipt Receipt
	if err == nil {
		receipt, err = p.mixes[r.member].Resolve(r.mix, useful, at)
	}
	if err != nil {
		p.poison, p.clock = true, at
		return PoolReceipt{}, err
	}
	receipt.TrialOrdinal = t.slot + 1
	p.rows[t.slot].status = 2
	p.pending--
	p.clock = at
	return PoolReceipt{receipt, r.member, r.ordinal}, nil
}

func (p *LocalPool) Cancel(t LocalTicket, at int64) error {
	if err := p.validate(t, at); err != nil {
		return err
	}
	r := p.rows[t.slot]
	err := p.full.Cancel(r.full, at)
	if err == nil {
		err = p.adaptive.Cancel(r.adaptive, at)
	}
	if err == nil {
		err = p.moment.Cancel(r.moment, at)
	}
	if err == nil {
		err = p.mixes[r.member].Cancel(r.mix, at)
	}
	if err != nil {
		p.poison, p.clock = true, at
		return err
	}
	p.rows[t.slot].status = 3
	p.pending--
	p.clock = at
	return nil
}

func (p *LocalPool) Pending() int { return p.pending }

func (p *LocalPool) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= p.epoch || at < p.clock {
		return errors.New("invalid local epoch")
	}
	next, err := NewLocalPool(p.base, p.cfg, p.memberTrials, epoch)
	if err != nil {
		return err
	}
	next.clock = at
	*p = *next
	return nil
}
