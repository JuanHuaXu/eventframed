package observationlearners

import (
	"errors"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type markovAdviceEntry struct {
	active, known, censored, outcome bool
	origin                           uint64
	raw                              [4]float64
}

// Research-only finite-state refilter. The checkpoint is immediately before
// base; current is immediately after the transition out of the last issue.
// Wall-clock flushes do not create transitions for unissued events.
type markovAdvice struct {
	ready               bool
	alpha               float64
	base, next          uint64
	checkpoint, current agedAdvice
	entries             [routedJournalCapacity]markovAdviceEntry
}

func newMarkovAdvice(alpha float64) (markovAdvice, error) {
	if math.IsNaN(alpha) || alpha < 0 || alpha > 1 {
		return markovAdvice{}, errors.New("invalid Markov transition")
	}
	s := newAgedAdvice(false, false)
	return markovAdvice{ready: true, alpha: alpha, checkpoint: s, current: s}, nil
}

// A missing outcome has unit emission, not a negative label. The same
// stochastic transition follows every issue, whether its label has arrived.
func markovAdviceStep(s *agedAdvice, e markovAdviceEntry, alpha float64) error {
	s.clock = e.origin
	if alpha == 1 {
		for j, p := range s.prior {
			s.logs[j] = math.Log(p)
		}
		return nil
	}
	if e.known {
		return observeLogAdvice(s, e.origin, e.raw, e.outcome, alpha)
	}
	if alpha > 0 {
		w := s.weights()
		for j := range s.logs {
			s.logs[j] = math.Log((1-alpha)*w[j] + alpha*s.prior[j])
		}
	}
	return nil
}

func (f *markovAdvice) issue(origin uint64, raw [4]float64) error {
	if f == nil || !f.ready || origin != f.next || origin >= 256 || f.next-f.base >= routedJournalCapacity {
		return errors.New("invalid Markov issue or capacity")
	}
	for _, p := range raw {
		if math.IsNaN(p) || p <= 0 || p >= 1 {
			return errors.New("invalid Markov emission")
		}
	}
	copy := *f
	e := markovAdviceEntry{active: true, origin: origin, raw: raw}
	if copy.entries[origin%routedJournalCapacity].active {
		return errors.New("occupied Markov entry")
	}
	if err := markovAdviceStep(&copy.current, e, copy.alpha); err != nil {
		return err
	}
	copy.entries[origin%routedJournalCapacity] = e
	copy.next++
	*f = copy
	return nil
}

// Recomputing messages does not readmit evidence. A label is admitted once;
// subsequent passes only evaluate the already-declared conditional model.
func (f *markovAdvice) rebuild() error {
	s := f.checkpoint
	base := f.base
	for origin := f.base; origin < f.next; origin++ {
		i := origin % routedJournalCapacity
		e := f.entries[i]
		if !e.active || e.origin != origin || (e.known && e.censored) {
			return errors.New("Markov suffix gap")
		}
		if err := markovAdviceStep(&s, e, f.alpha); err != nil {
			return err
		}
		if origin == base && (e.known || e.censored) {
			f.checkpoint = s
			f.entries[i] = markovAdviceEntry{}
			base++
		}
	}
	f.base, f.current = base, s
	return nil
}

func (f *markovAdvice) deliver(origin uint64, y bool) error {
	if f == nil || !f.ready || origin < f.base || origin >= f.next {
		return errors.New("invalid Markov delivery")
	}
	e := f.entries[origin%routedJournalCapacity]
	if !e.active || e.origin != origin || e.known || e.censored {
		return errors.New("settled Markov delivery")
	}
	copy := *f
	copy.entries[origin%routedJournalCapacity].known = true
	copy.entries[origin%routedJournalCapacity].outcome = y
	if err := copy.rebuild(); err != nil {
		return err
	}
	*f = copy
	return nil
}

func (f *markovAdvice) expireBefore(cutoff uint64) error {
	if f == nil || !f.ready || cutoff > f.next {
		return errors.New("invalid Markov expiry")
	}
	copy := *f
	for i, e := range copy.entries {
		if e.active && !e.known && e.origin < cutoff {
			copy.entries[i].censored = true
		}
	}
	if err := copy.rebuild(); err != nil {
		return err
	}
	*f = copy
	return nil
}

type markovAdviceJournal struct {
	agedAdviceJournal
	filter markovAdvice
}

func newMarkovAdviceJournal() *markovAdviceJournal {
	f, _ := newMarkovAdvice(.001)
	return &markovAdviceJournal{agedAdviceJournal: *newAgedAdviceJournal(false, false), filter: f}
}

func (g *markovAdviceJournal) syncAdvice() error {
	if g.filter.next != g.journal.stats.Issued || g.filter.base != g.journal.drainAt || g.filter.next-g.filter.base != g.journal.stats.Pending {
		return errors.New("Markov journal frontier mismatch")
	}
	// Preserve the journal's wall clock while updating its next-issue law.
	g.advice.logs = g.filter.current.logs
	return nil
}

func (g *markovAdviceJournal) predict(origin uint64, reader observation.Reader, epoch uint64) (observation.Result, error) {
	if g == nil || g.journal.busy {
		return observation.Result{}, errors.New("invalid Markov prediction")
	}
	copy := *g
	// The reader can call back into the original owner, not just this copy.
	g.journal.busy = true
	defer func() { g.journal.busy = false }()
	r, err := copy.agedAdviceJournal.predict(origin, reader, epoch)
	if err != nil {
		return r, err
	}
	e := copy.journal.entries[origin%routedJournalCapacity]
	if err := copy.filter.issue(origin, e.bank.Experts); err != nil {
		return observation.Result{}, err
	}
	if err := copy.syncAdvice(); err != nil {
		return observation.Result{}, err
	}
	*g = copy
	return r, nil
}

func (g *markovAdviceJournal) deliver(origin uint64, y bool) error {
	if g == nil || g.journal.busy {
		return errors.New("invalid Markov journal delivery")
	}
	copy := *g
	if err := copy.journal.deliver(origin, y); err != nil {
		return err
	}
	if err := copy.filter.deliver(origin, y); err != nil {
		return err
	}
	if err := copy.syncAdvice(); err != nil {
		return err
	}
	copy.received++
	*g = copy
	return nil
}

func (g *markovAdviceJournal) expireBefore(cutoff uint64) error {
	if g == nil || g.journal.busy {
		return errors.New("invalid Markov journal expiry")
	}
	copy := *g
	if err := copy.journal.expireBefore(cutoff); err != nil {
		return err
	}
	if err := copy.filter.expireBefore(cutoff); err != nil {
		return err
	}
	if err := copy.syncAdvice(); err != nil {
		return err
	}
	*g = copy
	return nil
}
