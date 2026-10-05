package observationlearners

import (
	"errors"
	"math"
)

const snapshotAdviceCapacity = 80

var snapshotRolePrior = [4]float64{.95, .05 / 3, .05 / 3, .05 / 3}

type snapshotAdviceState struct {
	ids     [8]uint64
	logs    [8]float64
	version int
}

type snapshotAdviceEvent struct {
	publication, known, censored, outcome bool
	origin                                uint64
	version                               int
	ids                                   [8]uint64
	raw                                   [8]float64
}

func newSnapshotAdviceState() snapshotAdviceState {
	s := snapshotAdviceState{version: -1}
	for i := range s.logs {
		s.logs[i] = math.Inf(-1)
	}
	return s
}

func snapshotLogAdd(a, b float64) float64 {
	if math.IsInf(a, -1) {
		return b
	}
	if math.IsInf(b, -1) {
		return a
	}
	if a < b {
		a, b = b, a
	}
	return a + math.Log1p(math.Exp(b-a))
}

func (s *snapshotAdviceState) normalize() error {
	total := math.Inf(-1)
	for i, v := range s.logs {
		if math.IsNaN(v) || math.IsInf(v, 1) || s.ids[i] == 0 && !math.IsInf(v, -1) {
			return errors.New("invalid snapshot log mass")
		}
		total = snapshotLogAdd(total, v)
	}
	if math.IsInf(total, -1) {
		return errors.New("empty snapshot law")
	}
	for i := range s.logs {
		s.logs[i] -= total
	}
	return nil
}

func (s snapshotAdviceState) weights() ([8]float64, error) {
	var w [8]float64
	if err := s.normalize(); err != nil {
		return w, err
	}
	for i := range w {
		w[i] = math.Exp(s.logs[i])
	}
	return w, nil
}

// Publication is a stochastic transition, not conditioning on survival.
// A survivor keeps (1-rho) of its mass; every retired path enters the new bank.
// Historical messages retain their original identities when slots are reused.
func snapshotAdviceStep(s *snapshotAdviceState, e snapshotAdviceEvent, alpha float64) error {
	if math.IsNaN(alpha) || alpha < 0 || alpha > 1 {
		return errors.New("invalid snapshot share")
	}
	if e.publication {
		if e.version != s.version+1 || e.version < 0 || e.version >= 8 || e.origin != uint64(32*e.version) {
			return errors.New("invalid snapshot admission")
		}
		start := 4 * (e.version % 2)
		if e.version == 0 {
			for j, p := range snapshotRolePrior {
				s.ids[j] = uint64(j + 1)
				s.logs[j] = math.Log(p)
			}
		} else {
			if err := s.normalize(); err != nil {
				return err
			}
			rho := 1 / float64(e.version+1)
			retired := math.Inf(-1)
			for j := start; j < start+4; j++ {
				retired = snapshotLogAdd(retired, s.logs[j])
			}
			incoming := snapshotLogAdd(math.Log(rho), math.Log1p(-rho)+retired)
			for j := range s.logs {
				s.logs[j] += math.Log1p(-rho)
			}
			for j, p := range snapshotRolePrior {
				s.ids[start+j] = uint64(4*e.version + j + 1)
				s.logs[start+j] = incoming + math.Log(p)
			}
		}
		s.version = e.version
		return s.normalize()
	}
	if s.version < 0 || int(e.origin/32) != s.version || e.ids != s.ids || e.known && e.censored {
		return errors.New("snapshot issue identity mismatch")
	}
	for i, id := range e.ids {
		p := e.raw[i]
		if id == 0 {
			if p != 0 {
				return errors.New("inactive snapshot emission")
			}
			continue
		}
		if math.IsNaN(p) || p <= 0 || p >= 1 {
			return errors.New("invalid snapshot emission")
		}
		if e.known {
			if e.outcome {
				s.logs[i] += math.Log(p)
			} else {
				s.logs[i] += math.Log1p(-p)
			}
		}
	}
	if err := s.normalize(); err != nil {
		return err
	}
	if alpha > 0 {
		banks := 1.
		if s.version > 0 {
			banks = 2
		}
		for i, id := range s.ids {
			if id == 0 {
				continue
			}
			prior := snapshotRolePrior[(id-1)%4] / banks
			s.logs[i] = snapshotLogAdd(math.Log1p(-alpha)+s.logs[i], math.Log(alpha)+math.Log(prior))
		}
	}
	return s.normalize()
}

// Research component only, externally serialized. Publication nodes have no
// observation likelihood. The bounded suffix supports exact delayed refiltering;
// it is not a record of independent re-admissions of the same evidence.
type snapshotAdvice struct {
	ready                              bool
	alpha                              float64
	head, tail, baseOrigin, nextOrigin uint64
	checkpoint, current                snapshotAdviceState
	events                             [snapshotAdviceCapacity]snapshotAdviceEvent
}

func newSnapshotAdvice(alpha float64) (snapshotAdvice, error) {
	if math.IsNaN(alpha) || alpha < 0 || alpha > 1 {
		return snapshotAdvice{}, errors.New("invalid snapshot share")
	}
	s := newSnapshotAdviceState()
	return snapshotAdvice{ready: true, alpha: alpha, checkpoint: s, current: s}, nil
}

func (f *snapshotAdvice) append(e snapshotAdviceEvent) error {
	if f.tail-f.head >= snapshotAdviceCapacity {
		return errors.New("snapshot suffix capacity")
	}
	if err := snapshotAdviceStep(&f.current, e, f.alpha); err != nil {
		return err
	}
	f.events[f.tail%snapshotAdviceCapacity] = e
	f.tail++
	if e.publication && f.tail-f.head == 1 {
		f.checkpoint = f.current
		f.events[f.head%snapshotAdviceCapacity] = snapshotAdviceEvent{}
		f.head++
	}
	return nil
}

func (f *snapshotAdvice) publish(version int) error {
	if f == nil || !f.ready || version < 0 || version >= 8 || f.nextOrigin != uint64(32*version) {
		return errors.New("invalid snapshot publication clock")
	}
	c := *f
	if err := c.append(snapshotAdviceEvent{publication: true, version: version, origin: f.nextOrigin}); err != nil {
		return err
	}
	*f = c
	return nil
}

func (f *snapshotAdvice) issue(origin uint64, ids [8]uint64, raw [8]float64) error {
	if f == nil || !f.ready || origin != f.nextOrigin || origin >= 256 || f.nextOrigin-f.baseOrigin >= 64 {
		return errors.New("invalid snapshot issue or capacity")
	}
	c := *f
	if err := c.append(snapshotAdviceEvent{origin: origin, ids: ids, raw: raw}); err != nil {
		return err
	}
	c.nextOrigin++
	*f = c
	return nil
}

func (f *snapshotAdvice) rebuild() error {
	s := f.checkpoint
	head := f.head
	for seq := f.head; seq < f.tail; seq++ {
		e := f.events[seq%snapshotAdviceCapacity]
		if err := snapshotAdviceStep(&s, e, f.alpha); err != nil {
			return err
		}
		if seq == head && (e.publication || e.known || e.censored) {
			f.checkpoint = s
			if !e.publication {
				f.baseOrigin = e.origin + 1
			}
			f.events[seq%snapshotAdviceCapacity] = snapshotAdviceEvent{}
			head++
		}
	}
	f.head, f.current = head, s
	return nil
}

func (f *snapshotAdvice) deliver(origin uint64, y bool) error {
	if f == nil || !f.ready || origin < f.baseOrigin || origin >= f.nextOrigin {
		return errors.New("invalid snapshot delivery")
	}
	c := *f
	found := false
	for seq := c.head; seq < c.tail; seq++ {
		e := &c.events[seq%snapshotAdviceCapacity]
		if !e.publication && e.origin == origin {
			if e.known || e.censored {
				return errors.New("settled snapshot delivery")
			}
			e.known, e.outcome = true, y
			found = true
			break
		}
	}
	if !found {
		return errors.New("missing snapshot origin")
	}
	if err := c.rebuild(); err != nil {
		return err
	}
	*f = c
	return nil
}

func (f *snapshotAdvice) expireBefore(cutoff uint64) error {
	if f == nil || !f.ready || cutoff > f.nextOrigin {
		return errors.New("invalid snapshot expiry")
	}
	c := *f
	for seq := c.head; seq < c.tail; seq++ {
		e := &c.events[seq%snapshotAdviceCapacity]
		if !e.publication && !e.known && e.origin < cutoff {
			e.censored = true
		}
	}
	if err := c.rebuild(); err != nil {
		return err
	}
	*f = c
	return nil
}
