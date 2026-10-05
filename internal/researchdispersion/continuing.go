package researchdispersion

import (
	"errors"
	"math"
)

const RateAtoms = 21
const ratePriorStrength = 4.

// ContinuingObserver has a separate finite-state latent rate chain per member.
// It integrates rate uncertainty instead of freezing four-label point templates
// or forcing one reversal state across every member. It grants no AP authority.
// This is exact for the declared discrete model, not continuous-rate BOCPD.
type ContinuingObserver struct {
	ledger  *DelayedShape
	hazard  float64
	prior   [][RateAtoms]float64
	forward [][RateAtoms]float64
	labels  []bool
	scratch [MaxTrials][RateAtoms]float64
}

func NewContinuingObserver(base []float64, epoch uint64, pendingCap int, hazard float64) (*ContinuingObserver, error) {
	if hazard < 0 || hazard >= 1 || math.IsNaN(hazard) || math.IsInf(hazard, 0) {
		return nil, errors.New("invalid continuing hazard")
	}
	ledger, err := NewDelayedShape(base, epoch, pendingCap)
	if err != nil {
		return nil, err
	}
	m := &ContinuingObserver{ledger: ledger, hazard: hazard, prior: make([][RateAtoms]float64, len(base)), forward: make([][RateAtoms]float64, len(ledger.trials)), labels: make([]bool, len(ledger.trials))}
	for i, b := range base {
		var sum float64
		for z := 0; z < RateAtoms; z++ {
			q := rateAtom(z)
			w := math.Pow(q, ratePriorStrength*b-1) * math.Pow(1-q, ratePriorStrength*(1-b)-1)
			m.prior[i][z] = w
			sum += w
		}
		if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
			return nil, errors.New("invalid continuing prior")
		}
		for z := 0; z < RateAtoms; z++ {
			m.prior[i][z] /= sum
		}
	}
	return m, nil
}
func rateAtom(z int) float64 { return (float64(z) + .5) / RateAtoms }
func (m *ContinuingObserver) next(member int) [RateAtoms]float64 {
	w := m.prior[member]
	if n := int(m.ledger.issued[member]); n > 0 {
		last := m.forward[member*MaxTrials+n-1]
		for z := 0; z < RateAtoms; z++ {
			w[z] = (1-m.hazard)*last[z] + m.hazard*w[z]
		}
	}
	return w
}
func (m *ContinuingObserver) Predict(member int) (float64, error) {
	if member < 0 || member >= len(m.prior) {
		return 0, errors.New("unknown continuing member")
	}
	w := m.next(member)
	q := 0.
	for z, p := range w {
		q += p * rateAtom(z)
	}
	if q <= 0 || q >= 1 || math.IsNaN(q) || math.IsInf(q, 0) {
		return 0, errors.New("invalid continuing forecast")
	}
	return q, nil
}
func (m *ContinuingObserver) Issue(member int, at int64) (DelayedTicket, error) {
	q, err := m.Predict(member)
	if err != nil {
		return DelayedTicket{}, err
	}
	w := m.next(member)
	t, err := m.ledger.Issue(member, at)
	if err != nil {
		return DelayedTicket{}, err
	}
	m.forward[t.slot] = w
	m.ledger.trials[t.slot].forecast, t.q = q, q
	return t, nil
}

// Late outcomes enter their ORIGINAL member position. Unknown/cancelled labels
// have unit emission. Replay prepares private rows before publishing anything;
// scoring always uses the privately retained original issued forecast.
func (m *ContinuingObserver) Resolve(t DelayedTicket, useful bool, at int64) (DelayedReceipt, error) {
	if err := m.ledger.validate(t, at); err != nil {
		return DelayedReceipt{}, err
	}
	member, ordinal := t.slot/MaxTrials, t.slot%MaxTrials
	w := m.prior[member]
	if ordinal > 0 {
		w = m.forward[t.slot-1]
	}
	for j := ordinal; j < int(m.ledger.issued[member]); j++ {
		if j > 0 {
			for z := 0; z < RateAtoms; z++ {
				w[z] = (1-m.hazard)*w[z] + m.hazard*m.prior[member][z]
			}
		}
		slot := member*MaxTrials + j
		if slot == t.slot || m.ledger.trials[slot].status == 2 {
			y := m.labels[slot]
			if slot == t.slot {
				y = useful
			}
			sum := 0.
			for z := 0; z < RateAtoms; z++ {
				w[z] *= likelihood(rateAtom(z), y)
				sum += w[z]
			}
			if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
				return DelayedReceipt{}, errors.New("invalid continuing normalization")
			}
			for z := 0; z < RateAtoms; z++ {
				w[z] /= sum
			}
		}
		m.scratch[j] = w
	}
	for j := ordinal; j < int(m.ledger.issued[member]); j++ {
		m.forward[member*MaxTrials+j] = m.scratch[j]
	}
	x := m.ledger.trials[t.slot]
	m.labels[t.slot] = useful
	m.ledger.trials[t.slot].status = 2
	m.ledger.pending--
	m.ledger.clock = at
	return DelayedReceipt{Member: member, TrialOrdinal: ordinal + 1, Epoch: m.ledger.epoch, IssuedAt: x.issuedAt, ArrivedAt: at, Forecast: x.forecast, Useful: useful}, nil
}
func (m *ContinuingObserver) Pending() int                           { return m.ledger.Pending() }
func (m *ContinuingObserver) Cancel(t DelayedTicket, at int64) error { return m.ledger.Cancel(t, at) }
func (m *ContinuingObserver) BeginEpoch(epoch uint64, at int64) error {
	if epoch <= m.ledger.epoch || at < m.ledger.clock {
		return errors.New("nonincreasing epoch or backward time")
	}
	next, err := NewContinuingObserver(m.ledger.base, epoch, m.ledger.cap, m.hazard)
	if err != nil {
		return err
	}
	next.ledger.clock = at
	*m = *next
	return nil
}
