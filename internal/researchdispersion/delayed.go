package researchdispersion

import "errors"

// DelayedShape is a single-owner research adapter for stationary, genuine
// repeated trials. It binds issued forecasts to feedback identities, not
// source authenticity. Outcome-dependent missingness needs a separate model.
type DelayedShape struct {
	model        *ShapeModel
	base         []float64
	trials       []delayedTrial
	issued       []uint16
	epoch        uint64
	clock        int64
	pending, cap int
}

type delayedTrial struct {
	status   uint8 // 0 unissued, 1 pending, 2 resolved, 3 cancelled
	issuedAt int64
	forecast float64
}

// DelayedTicket is opaque and owner/epoch-bound. The adapter retains the
// authoritative forecast privately; callers cannot replace it on resolution.
type DelayedTicket struct {
	owner *DelayedShape
	epoch uint64
	slot  int
	q     float64
}

func (t DelayedTicket) Forecast() float64 { return t.q }

type DelayedReceipt struct {
	Member, TrialOrdinal int
	Epoch                uint64
	IssuedAt, ArrivedAt  int64
	Forecast             float64
	Useful               bool
}

func NewDelayedShape(base []float64, epoch uint64, pendingCap int) (*DelayedShape, error) {
	if epoch == 0 || pendingCap < 1 || pendingCap > len(base)*MaxTrials {
		return nil, errors.New("invalid epoch or pending cap")
	}
	m, err := NewShape(base)
	if err != nil {
		return nil, err
	}
	return &DelayedShape{model: m, base: append([]float64(nil), base...), trials: make([]delayedTrial, len(base)*MaxTrials), issued: make([]uint16, len(base)), epoch: epoch, cap: pendingCap}, nil
}

func (d *DelayedShape) Predict(member int) (float64, error) {
	return d.model.Predict(member)
}

func (d *DelayedShape) Pending() int { return d.pending }

// Issue sees no outcome. Times are monotonically increasing logical as-of
// ticks; several events can share a tick. Admission failure mutates nothing.
func (d *DelayedShape) Issue(member int, at int64) (DelayedTicket, error) {
	if member < 0 || member >= len(d.issued) || at < d.clock || d.pending >= d.cap || int(d.issued[member]) >= MaxTrials {
		return DelayedTicket{}, errors.New("invalid, backward-time or capped issue")
	}
	q, err := d.model.Predict(member)
	if err != nil {
		return DelayedTicket{}, err
	}
	slot := member*MaxTrials + int(d.issued[member])
	d.trials[slot] = delayedTrial{status: 1, issuedAt: at, forecast: q}
	d.issued[member]++
	d.pending++
	d.clock = at
	return DelayedTicket{owner: d, epoch: d.epoch, slot: slot, q: q}, nil
}

func (d *DelayedShape) validate(t DelayedTicket, at int64) error {
	if t.owner != d || t.epoch != d.epoch || t.slot < 0 || t.slot >= len(d.trials) || at < d.clock {
		return errors.New("invalid, stale or backward-time ticket")
	}
	x := d.trials[t.slot]
	if x.status != 1 || at < x.issuedAt {
		return errors.New("unissued, replayed, cancelled or premature ticket")
	}
	return nil
}

// Resolve uses arrival counts for the exchangeable likelihood, not the issue
// ordinal. Scoring retains the issued forecast, never today's updated law.
func (d *DelayedShape) Resolve(t DelayedTicket, useful bool, at int64) (DelayedReceipt, error) {
	if err := d.validate(t, at); err != nil {
		return DelayedReceipt{}, err
	}
	member := t.slot / MaxTrials
	x := d.trials[t.slot]
	if err := d.model.Observe(member, int(d.model.n[member])+1, useful); err != nil {
		return DelayedReceipt{}, err
	}
	d.trials[t.slot].status = 2
	d.pending--
	d.clock = at
	return DelayedReceipt{Member: member, TrialOrdinal: t.slot%MaxTrials + 1, Epoch: d.epoch, IssuedAt: x.issuedAt, ArrivedAt: at, Forecast: x.forecast, Useful: useful}, nil
}

// Cancel is explicit censoring, not a negative label. Its issue ordinal is
// permanently consumed in this epoch so the same trial cannot be reissued.
func (d *DelayedShape) Cancel(t DelayedTicket, at int64) error {
	if err := d.validate(t, at); err != nil {
		return err
	}
	d.trials[t.slot].status = 3
	d.pending--
	d.clock = at
	return nil
}

// BeginEpoch discards old evidence/pending work explicitly. It is NOT a
// changepoint detector; the external controller must justify the boundary.
func (d *DelayedShape) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= d.epoch || at < d.clock {
		return errors.New("nonincreasing epoch or backward time")
	}
	m, err := NewShape(d.base)
	if err != nil {
		return err
	}
	d.model = m
	clear(d.trials)
	clear(d.issued)
	d.pending, d.epoch, d.clock = 0, epoch, at
	return nil
}
