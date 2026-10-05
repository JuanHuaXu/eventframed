package observationlearners

import (
	"errors"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// Research ablation: the advice selector learns at label arrival, whereas
// version-scoped tests retain the original prefix-release order. This is not
// an ordinary fixed-order posterior or an inherited delayed-regret guarantee.
type arrivalRoutedJournal struct {
	journal  routedFeedbackJournal
	received uint64
}

func newArrivalRoutedJournal() *arrivalRoutedJournal {
	return &arrivalRoutedJournal{journal: *newRoleRoutedFeedbackJournal()}
}

func (g *arrivalRoutedJournal) publish(e *observationExperts) error {
	if g == nil {
		return errors.New("nil arrival journal")
	}
	return g.journal.publish(e)
}

func (g *arrivalRoutedJournal) predict(origin uint64, r observation.Reader, epoch uint64) (observation.Result, error) {
	if g == nil {
		return observation.Result{}, errors.New("nil arrival journal")
	}
	return g.journal.predict(origin, r, epoch)
}

func (g *arrivalRoutedJournal) deliver(origin uint64, y bool) error {
	if g == nil {
		return errors.New("nil arrival journal")
	}
	copy := g.journal
	entry := copy.entries[origin%routedJournalCapacity]
	bank := copy.core.bank
	// The frozen journal validates identity/duplicates and advances only the
	// allowed prefix. Its temporary bank updates do not enter test calculations.
	if err := copy.deliver(origin, y); err != nil {
		return err
	}
	// Restore the selector, then apply exactly this newly arrived advice loss.
	// Never apply the buffered-prefix losses twice when the head later clears.
	bank.pending = entry.bank
	bank.active = true
	if err := bank.observe(origin, y); err != nil {
		return err
	}
	copy.core.bank = bank
	g.journal = copy
	g.received++
	return nil
}

func (g *arrivalRoutedJournal) expireBefore(cutoff uint64) error {
	if g == nil {
		return errors.New("nil arrival journal")
	}
	copy := g.journal
	bank := copy.core.bank
	if err := copy.expireBefore(cutoff); err != nil {
		return err
	}
	// Releasing already-delivered evidence affects the gate, not the selector:
	// these losses were accounted for at their own arrivals.
	copy.core.bank = bank
	g.journal = copy
	return nil
}
