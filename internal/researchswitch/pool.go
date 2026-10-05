package researchswitch

import (
	"errors"

	"github.com/JuanHuaXu/eventframed/internal/researchdispersion"
	"github.com/JuanHuaXu/eventframed/internal/researchmoment"
)

// Pool uses ALL-visible-data learning experts, not local/reset experts.
// Its mixer likelihood scores original expert advice exactly once. Child
// models receive the same revealed label, not fictitiously independent copies.
type Pool struct {
	full, adaptive *researchdispersion.WindowObserver
	moment         *researchmoment.Model
	mix            *Model
	base           []float64
	rows           []poolRow
	issued         []int
	clock          int64
	epoch          uint64
	poison         bool
}

type poolRow struct {
	full, adaptive  researchdispersion.DelayedTicket
	moment          researchmoment.Ticket
	mix             Ticket
	member, ordinal int
	at              int64
	status          uint8
}

type PoolTicket struct {
	owner *Pool
	epoch uint64
	slot  int
	q     float64
}

func (t PoolTicket) Forecast() float64 { return t.q }

type PoolReceipt struct {
	Receipt
	Member, MemberOrdinal int
}

func NewPool(base []float64, cfg Config, epoch uint64) (*Pool, error) {
	if cfg.Trials > len(base)*researchdispersion.MaxTrials {
		return nil, errors.New("pool capacity exceeds member evidence cap")
	}
	// This is a declared THREE-expert pool, never an arbitrary unbound advice
	// vector. Its prior may be frozen by the study but arity cannot vary.
	if len(cfg.Prior) != 3 {
		return nil, errors.New("pool requires Full Adaptive and rich-moment experts")
	}
	mix, err := New(cfg, epoch)
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
	return &Pool{full: full, adaptive: adaptive, moment: moment, mix: mix, base: append([]float64(nil), base...), rows: make([]poolRow, cfg.Trials), issued: make([]int, len(base)), epoch: epoch}, nil
}

func (p *Pool) advice(member int) ([3]float64, error) {
	var q [3]float64
	if p.poison || member < 0 || member >= len(p.base) {
		return q, errors.New("poisoned pool or unknown member")
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

func (p *Pool) Predict(member int) (float64, error) {
	q, err := p.advice(member)
	if err != nil {
		return 0, err
	}
	return p.mix.Predict(q[:])
}

// Children expose no cross-model transactional prepare API. Admission is
// prevalidated, but any unexpected child failure fences the whole pool until
// explicit epoch replacement. It MUST NOT emit from a partly updated bundle.
func (p *Pool) Issue(member int, at int64) (PoolTicket, error) {
	if p.poison || member < 0 || member >= len(p.base) || at < p.clock || p.mix.used == len(p.rows) || p.mix.pending == p.mix.cap || p.issued[member] == researchdispersion.MaxTrials {
		return PoolTicket{}, errors.New("invalid or capped pool issue")
	}
	q, err := p.advice(member)
	if err != nil {
		return PoolTicket{}, err
	}
	if _, err = p.mix.Predict(q[:]); err != nil {
		return PoolTicket{}, err
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
		r.mix, err = p.mix.Issue(q[:], at)
	}
	if err != nil {
		p.poison, p.clock = true, at
		return PoolTicket{}, err
	}
	if r.full.Forecast() != q[0] || r.adaptive.Forecast() != q[1] || r.moment.Forecast() != q[2] {
		p.poison, p.clock = true, at
		return PoolTicket{}, errors.New("expert changed during pool issue")
	}
	slot := p.mix.used - 1
	p.rows[slot] = r
	p.issued[member]++
	p.clock = at
	return PoolTicket{p, p.epoch, slot, r.mix.Forecast()}, nil
}

func (p *Pool) validate(t PoolTicket, at int64) error {
	if p.poison || t.owner != p || t.epoch != p.epoch || t.slot < 0 || t.slot >= p.mix.used || at < p.clock || at < p.rows[t.slot].at || p.rows[t.slot].status != 1 {
		return errors.New("invalid pool owner epoch time or status")
	}
	return nil
}

func (p *Pool) Resolve(t PoolTicket, useful bool, at int64) (PoolReceipt, error) {
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
		receipt, err = p.mix.Resolve(r.mix, useful, at)
	}
	if err != nil {
		p.poison, p.clock = true, at
		return PoolReceipt{}, err
	}
	p.rows[t.slot].status = 2
	p.clock = at
	return PoolReceipt{receipt, r.member, r.ordinal}, nil
}

func (p *Pool) Cancel(t PoolTicket, at int64) error {
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
		err = p.mix.Cancel(r.mix, at)
	}
	if err != nil {
		p.poison, p.clock = true, at
		return err
	}
	p.rows[t.slot].status = 3
	p.clock = at
	return nil
}

func (p *Pool) Pending() int { return p.mix.Pending() }

func (p *Pool) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= p.epoch || at < p.clock {
		return errors.New("invalid pool epoch")
	}
	cfg := Config{Prior: append([]float64(nil), p.mix.prior[:p.mix.n]...), Hazard: p.mix.hazard, Trials: len(p.rows), Pending: p.mix.cap}
	next, err := NewPool(p.base, cfg, epoch)
	if err != nil {
		return err
	}
	// An unchanged caller-visible logical clock fences all rebuilt children.
	next.clock = at
	*p = *next
	return nil
}
