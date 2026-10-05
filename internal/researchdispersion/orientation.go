package researchdispersion

import (
	"errors"
	"math"
)

const orientationWarm = 4
const orientationHazard = 1. / 600

type orientationNode struct {
	member   int
	observed bool
	useful   bool
	forward  float64 // P(reversed here | arrived emissions through this position)
}

// OrientationFilter is an exact two-state filter conditional on a frozen
// plug-in template. It does not integrate template uncertainty or identify
// causes. The hidden state advances per nomination, not per wall-clock tick.
type OrientationFilter struct {
	template []float64
	nodes    []orientationNode
	scratch  []float64
	hazard   float64
}

func NewOrientationFilter(template []float64, hazard float64, cap int) (*OrientationFilter, error) {
	if len(template) == 0 || cap < 1 || cap > len(template)*MaxTrials || hazard < 0 || hazard >= .5 || math.IsNaN(hazard) {
		return nil, errors.New("invalid orientation contract")
	}
	for _, p := range template {
		if p <= 0 || p >= 1 || math.IsNaN(p) {
			return nil, errors.New("invalid orientation template")
		}
	}
	return &OrientationFilter{template: append([]float64(nil), template...), nodes: make([]orientationNode, 0, cap), scratch: make([]float64, cap), hazard: hazard}, nil
}

func (m *OrientationFilter) next() float64 {
	p := 0.
	if len(m.nodes) > 0 {
		p = m.nodes[len(m.nodes)-1].forward
	}
	return m.hazard + (1-2*m.hazard)*p
}

func (m *OrientationFilter) Predict(member int) (float64, error) {
	if member < 0 || member >= len(m.template) {
		return 0, errors.New("unknown orientation member")
	}
	p, f := m.template[member], m.next()
	q := (1-f)*p + f*(1-p)
	if q <= 0 || q >= 1 || math.IsNaN(q) {
		return 0, errors.New("invalid orientation forecast")
	}
	return q, nil
}

func (m *OrientationFilter) nominate(member int) (int, error) {
	if len(m.nodes) == cap(m.nodes) {
		return 0, errors.New("capped orientation history")
	}
	if _, err := m.Predict(member); err != nil {
		return 0, err
	}
	i := len(m.nodes)
	m.nodes = append(m.nodes, orientationNode{member: member, forward: m.next()})
	return i, nil
}

// Observe inserts a late emission at its ORIGINAL nomination position and
// replays only the affected suffix. Unobserved positions have unit emission.
// Scratch work is private; no failed normalization publishes partial state.
func (m *OrientationFilter) Observe(index int, useful bool) error {
	if index < 0 || index >= len(m.nodes) || m.nodes[index].observed {
		return errors.New("unknown or replayed orientation outcome")
	}
	f := 0.
	if index > 0 {
		f = m.nodes[index-1].forward
	}
	for j := index; j < len(m.nodes); j++ {
		x := m.nodes[j]
		f = m.hazard + (1-2*m.hazard)*f
		if x.observed || j == index {
			y := x.useful
			if j == index {
				y = useful
			}
			p := m.template[x.member]
			a, b := (1-f)*likelihood(p, y), f*likelihood(1-p, y)
			if a+b <= 0 || math.IsNaN(a+b) || math.IsInf(a+b, 0) {
				return errors.New("invalid orientation normalization")
			}
			f = b / (a + b)
		}
		if f < 0 || f > 1 || math.IsNaN(f) {
			return errors.New("invalid orientation state")
		}
		m.scratch[j] = f
	}
	for j := index; j < len(m.nodes); j++ {
		m.nodes[j].forward = m.scratch[j]
	}
	m.nodes[index].observed, m.nodes[index].useful = true, useful
	return nil
}

// OrientationObserver learns a template from the first four genuine trials
// of EACH member. Only after every such outcome arrives is it frozen. Later
// evidence filters a declared returning/reversed-pattern model, never AP
// authority. The anchor control uses hazard=0 and isolates template freezing.
type OrientationObserver struct {
	ledger       *DelayedShape
	anchor       *RollingShape
	filter       *OrientationFilter
	order        []int
	indices      []int
	labels       []bool
	anchorLabels int
	mode         string
}

func NewOrientationObserver(base []float64, epoch uint64, cap int, mode string) (*OrientationObserver, error) {
	if mode != "anchor" && mode != "orientation" {
		return nil, errors.New("invalid orientation mode")
	}
	ledger, err := NewDelayedShape(base, epoch, cap)
	if err != nil {
		return nil, err
	}
	anchor, err := NewRollingShape(base, MaxTrials)
	if err != nil {
		return nil, err
	}
	m := &OrientationObserver{ledger: ledger, anchor: anchor, mode: mode, indices: make([]int, len(ledger.trials)), labels: make([]bool, len(ledger.trials)), order: make([]int, 0, len(ledger.trials))}
	for i := range m.indices {
		m.indices[i] = -1
	}
	return m, nil
}

func (m *OrientationObserver) Predict(member int) (float64, error) {
	if m.filter == nil {
		return m.anchor.Predict(member)
	}
	return m.filter.Predict(member)
}

func (m *OrientationObserver) Issue(member int, at int64) (DelayedTicket, error) {
	q, err := m.Predict(member)
	if err != nil {
		return DelayedTicket{}, err
	}
	if m.filter != nil && len(m.filter.nodes) == cap(m.filter.nodes) {
		return DelayedTicket{}, errors.New("capped orientation history")
	}
	t, err := m.ledger.Issue(member, at)
	if err != nil {
		return DelayedTicket{}, err
	}
	m.ledger.trials[t.slot].forecast, t.q = q, q
	m.order = append(m.order, t.slot)
	// No nomination or outcome is duplicated: pre-freeze nominations are
	// replayed once when the last anchor label arrives.
	if m.filter != nil {
		// Predict and capacity checks already validated this nomination.
		m.indices[t.slot] = len(m.filter.nodes)
		m.filter.nodes = append(m.filter.nodes, orientationNode{member: member, forward: m.filter.next()})
	}
	return t, nil
}

func (m *OrientationObserver) prepareFreeze(plan *rollingPlan) (*OrientationFilter, []int, error) {
	n := len(m.ledger.base)
	template := make([]float64, n)
	for i := range template {
		row := m.anchor.model.cache[i*Hypotheses*ShapeKernels : (i+1)*Hypotheses*ShapeKernels]
		if i == plan.member {
			row = plan.row[:]
		}
		for z, p := range row {
			template[i] += plan.weights[z] * p
		}
	}
	hazard := 0.
	if m.mode == "orientation" {
		hazard = orientationHazard
	}
	f, err := NewOrientationFilter(template, hazard, n*(MaxTrials-orientationWarm))
	if err != nil {
		return nil, nil, err
	}
	indices := make([]int, len(m.indices))
	for i := range indices {
		indices[i] = -1
	}
	for _, slot := range m.order {
		if slot%MaxTrials < orientationWarm {
			continue
		}
		index, err := f.nominate(slot / MaxTrials)
		if err != nil {
			return nil, nil, err
		}
		indices[slot] = index
		if m.ledger.trials[slot].status == 2 {
			if err := f.Observe(index, m.labels[slot]); err != nil {
				return nil, nil, err
			}
		}
	}
	return f, indices, nil
}

func (m *OrientationObserver) Resolve(t DelayedTicket, useful bool, at int64) (DelayedReceipt, error) {
	if err := m.ledger.validate(t, at); err != nil {
		return DelayedReceipt{}, err
	}
	i, ordinal := t.slot/MaxTrials, t.slot%MaxTrials+1
	if ordinal <= orientationWarm {
		plan, err := m.anchor.prepare(i, ordinal, useful)
		if err != nil {
			return DelayedReceipt{}, err
		}
		var f *OrientationFilter
		var indices []int
		if m.anchorLabels+1 == len(m.ledger.base)*orientationWarm {
			f, indices, err = m.prepareFreeze(&plan)
			if err != nil {
				return DelayedReceipt{}, err
			}
		}
		m.anchor.commit(&plan)
		m.anchorLabels++
		if f != nil {
			m.filter, m.indices = f, indices
		}
	} else if m.filter != nil {
		if err := m.filter.Observe(m.indices[t.slot], useful); err != nil {
			return DelayedReceipt{}, err
		}
	}
	x := m.ledger.trials[t.slot]
	m.labels[t.slot] = useful
	m.ledger.trials[t.slot].status = 2
	m.ledger.pending--
	m.ledger.clock = at
	return DelayedReceipt{Member: i, TrialOrdinal: ordinal, Epoch: m.ledger.epoch, IssuedAt: x.issuedAt, ArrivedAt: at, Forecast: x.forecast, Useful: useful}, nil
}

func (m *OrientationObserver) Pending() int { return m.ledger.Pending() }
func (m *OrientationObserver) Cancel(t DelayedTicket, at int64) error {
	return m.ledger.Cancel(t, at)
}

func (m *OrientationObserver) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= m.ledger.epoch || at < m.ledger.clock {
		return errors.New("nonincreasing epoch or backward time")
	}
	next, err := NewOrientationObserver(m.ledger.base, epoch, m.ledger.cap, m.mode)
	if err != nil {
		return err
	}
	next.ledger.clock = at
	*m = *next
	return nil
}
