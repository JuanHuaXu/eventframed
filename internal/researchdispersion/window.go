package researchdispersion

import (
	"errors"
	"math"
	"math/bits"
)

// RollingShape is a truncated-evidence working posterior, not ordinary Bayes
// on the full history. Each member retains its newest arrived issue ordinals.
type RollingShape struct {
	model                  *ShapeModel
	width                  int
	seen, retained, labels []uint64
}

type rollingPlan struct {
	member                 int
	seen, retained, labels uint64
	n, s                   uint16
	ignored                bool
	logs, weights, row     [Hypotheses * ShapeKernels]float64
}

func NewRollingShape(base []float64, width int) (*RollingShape, error) {
	if width < 1 || width > MaxTrials {
		return nil, errors.New("invalid window width")
	}
	m, err := NewShape(base)
	if err != nil {
		return nil, err
	}
	return &RollingShape{model: m, width: width, seen: make([]uint64, len(base)), retained: make([]uint64, len(base)), labels: make([]uint64, len(base))}, nil
}

func (m *RollingShape) Predict(member int) (float64, error) {
	return m.model.Predict(member)
}

func likelihood(q float64, useful bool) float64 {
	if useful {
		return q
	}
	return 1 - q
}

// prepare removes an old likelihood conditional on the remaining member
// evidence, then adds the new one under that SAME remaining set. Using the
// old full-set forecast as a divisor would not undo its marginal likelihood.
func (m *RollingShape) prepare(member, ordinal int, useful bool) (rollingPlan, error) {
	p := rollingPlan{member: member}
	if member < 0 || member >= len(m.seen) || ordinal < 1 || ordinal > MaxTrials {
		return p, errors.New("invalid member or trial identity")
	}
	bit := uint64(1) << (ordinal - 1)
	if m.seen[member]&bit != 0 {
		return p, errors.New("replayed trial")
	}
	p.seen, p.retained, p.labels = m.seen[member]|bit, m.retained[member], m.labels[member]
	if useful {
		p.labels |= bit
	}
	n, s := int(m.model.n[member]), int(m.model.success[member])
	remove, oldUseful := false, false
	if n == m.width {
		oldest := bits.TrailingZeros64(p.retained)
		if ordinal-1 < oldest {
			// A late old label is acknowledged once but must not evict newer
			// evidence merely because its network arrival happened last.
			p.ignored = true
			return p, nil
		}
		oldBit := uint64(1) << oldest
		oldUseful, remove = p.labels&oldBit != 0, true
		p.retained &^= oldBit
		n--
		if oldUseful {
			s--
		}
	}
	p.retained |= bit
	p.n, p.s = uint16(n+1), uint16(s)
	if useful {
		p.s++
	}
	maximum := math.Inf(-1)
	for h := 0; h < Hypotheses; h++ {
		mean := m.model.p[member*Hypotheses+h]
		for k := 0; k < ShapeKernels; k++ {
			z := h*ShapeKernels + k
			q := shapeConditional(mean, k, n, s)
			p.logs[z] = m.model.logs[z] + math.Log(likelihood(q, useful))
			if remove {
				p.logs[z] -= math.Log(likelihood(q, oldUseful))
			}
			maximum = math.Max(maximum, p.logs[z])
			p.row[z] = shapeConditional(mean, k, int(p.n), int(p.s))
		}
	}
	sum := 0.
	for z, value := range p.logs {
		p.weights[z] = math.Exp(value - maximum)
		sum += p.weights[z]
	}
	if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
		return p, errors.New("invalid window normalization")
	}
	logSum := math.Log(sum)
	for z := range p.logs {
		p.logs[z] -= maximum + logSum
		p.weights[z] /= sum
	}
	return p, nil
}

func (m *RollingShape) commit(p *rollingPlan) {
	i := p.member
	m.seen[i], m.retained[i], m.labels[i] = p.seen, p.retained, p.labels
	if p.ignored {
		return
	}
	m.model.logs, m.model.w = p.logs, p.weights
	m.model.n[i], m.model.success[i] = p.n, p.s
	copy(m.model.cache[i*Hypotheses*ShapeKernels:(i+1)*Hypotheses*ShapeKernels], p.row[:])
}

func (m *RollingShape) Observe(member, ordinal int, useful bool) error {
	p, err := m.prepare(member, ordinal, useful)
	if err != nil {
		return err
	}
	m.commit(&p)
	return nil
}

var windowPrior = [4]float64{.85, .05, .05, .05}
var windowWidths = [4]int{64, 4, 8, 16}

const windowShare = 1. / 600

// WindowObserver shares V36's sealed identity ledger but replaces its law.
// Adaptive weights track original issued expert losses with fixed-share;
// they are not posterior probabilities of correct ontologies or AP authority.
type WindowObserver struct {
	ledger   *DelayedShape
	children []*RollingShape
	weights  [4]float64
	original [][4]float64
	mode     string
}

func NewWindowObserver(base []float64, epoch uint64, cap int, mode string) (*WindowObserver, error) {
	widths := []int(nil)
	switch mode {
	case "full":
		widths = []int{64}
	case "fixed4":
		widths = []int{4}
	case "fixed8":
		widths = []int{8}
	case "fixed16":
		widths = []int{16}
	case "adaptive":
		widths = windowWidths[:]
	default:
		return nil, errors.New("invalid window mode")
	}
	ledger, err := NewDelayedShape(base, epoch, cap)
	if err != nil {
		return nil, err
	}
	m := &WindowObserver{ledger: ledger, original: make([][4]float64, len(base)*MaxTrials), mode: mode, weights: [4]float64{1}}
	for _, width := range widths {
		child, err := NewRollingShape(base, width)
		if err != nil {
			return nil, err
		}
		m.children = append(m.children, child)
	}
	if mode == "adaptive" {
		m.weights = windowPrior
	}
	return m, nil
}

func (m *WindowObserver) forecasts(member int) ([4]float64, float64, error) {
	var row [4]float64
	q := 0.
	for j, child := range m.children {
		value, err := child.Predict(member)
		if err != nil {
			return row, 0, err
		}
		row[j], q = value, q+m.weights[j]*value
	}
	return row, q, nil
}

func (m *WindowObserver) Predict(member int) (float64, error) {
	_, q, err := m.forecasts(member)
	return q, err
}

func (m *WindowObserver) Issue(member int, at int64) (DelayedTicket, error) {
	row, q, err := m.forecasts(member)
	if err != nil {
		return DelayedTicket{}, err
	}
	ticket, err := m.ledger.Issue(member, at)
	if err != nil {
		return DelayedTicket{}, err
	}
	// Replace the ledger's unused baseline law before returning this issue.
	// No baseline observation is made; only the new law is scored/emitted.
	m.ledger.trials[ticket.slot].forecast, ticket.q = q, q
	m.original[ticket.slot] = row
	return ticket, nil
}

func (m *WindowObserver) Resolve(t DelayedTicket, useful bool, at int64) (DelayedReceipt, error) {
	if err := m.ledger.validate(t, at); err != nil {
		return DelayedReceipt{}, err
	}
	i, ordinal := t.slot/MaxTrials, t.slot%MaxTrials+1
	weights := m.weights
	if m.mode == "adaptive" {
		sum := 0.
		for j, q := range m.original[t.slot] {
			weights[j] *= likelihood(q, useful)
			sum += weights[j]
		}
		if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
			return DelayedReceipt{}, errors.New("invalid expert normalization")
		}
		for j := range weights {
			weights[j] = (1-windowShare)*weights[j]/sum + windowShare*windowPrior[j]
		}
	}
	// Validate all child transitions before committing any one child, so a
	// failed normalization cannot publish a partially updated expert bundle.
	var plans [4]rollingPlan
	for j, child := range m.children {
		p, err := child.prepare(i, ordinal, useful)
		if err != nil {
			return DelayedReceipt{}, err
		}
		plans[j] = p
	}
	for j, child := range m.children {
		child.commit(&plans[j])
	}
	m.weights = weights
	x := m.ledger.trials[t.slot]
	m.ledger.trials[t.slot].status = 2
	m.ledger.pending--
	m.ledger.clock = at
	return DelayedReceipt{Member: i, TrialOrdinal: ordinal, Epoch: m.ledger.epoch, IssuedAt: x.issuedAt, ArrivedAt: at, Forecast: x.forecast, Useful: useful}, nil
}

func (m *WindowObserver) Pending() int { return m.ledger.Pending() }

func (m *WindowObserver) Cancel(t DelayedTicket, at int64) error {
	return m.ledger.Cancel(t, at)
}

func (m *WindowObserver) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= m.ledger.epoch || at < m.ledger.clock {
		return errors.New("nonincreasing epoch or backward time")
	}
	next, err := NewWindowObserver(m.ledger.base, epoch, m.ledger.cap, m.mode)
	if err != nil {
		return err
	}
	next.ledger.clock = at
	*m = *next
	return nil
}
