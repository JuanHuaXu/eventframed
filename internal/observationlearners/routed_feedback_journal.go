package observationlearners

import (
	"errors"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

const routedJournalCapacity = 64

type routedJournalEntry struct {
	active, ready, censored, outcome bool
	origin, version                  uint64
	bank                             brierBankForecast
	gate                             comparativeForecast
	served                           float64
}

type routedJournalStats struct{ Issued, Pending, Applied, BankOnly, Stale, Censored uint64 }

// Single-owner research adapter. Results may arrive out of order, but selector
// and test updates drain in origin order. A missing head needs an explicit,
// predeclared expiry decision; it is never silently turned into a negative label.
// The 256-issue horizon retains eight test allocations. It does not establish
// conditional-null validity under an arbitrary delay or missingness process.
type routedFeedbackJournal struct {
	core                          routedObservationState
	models                        *observationExperts
	entries                       [routedJournalCapacity]routedJournalEntry
	stats                         routedJournalStats
	version, publishedAt, drainAt uint64
	busy                          bool
	carrySelector                 bool
}

func newRoutedFeedbackJournal() *routedFeedbackJournal {
	return &routedFeedbackJournal{core: *newRoutedObservationState()}
}

// Learner roles remain generic64, Boolean64, generic32 and Boolean32 across
// publications. This variant can score their historical advice without claiming
// that a new fitted model issued it. Version-specific tests never receive it.
func newRoleRoutedFeedbackJournal() *routedFeedbackJournal {
	g := newRoutedFeedbackJournal()
	g.carrySelector = true
	return g
}

func (g *routedFeedbackJournal) publish(e *observationExperts) error {
	if g == nil || g.busy || e == nil || !e.ready || g.core.bank.count != 4 {
		return errors.New("invalid journal publication")
	}
	n := g.core.bank.issued
	if n >= 256 || n%32 != 0 || (g.models != nil && g.publishedAt == n) {
		return errors.New("publication outside frozen cadence")
	}
	g.models, g.publishedAt = e, n
	g.version++
	// Reset at publication, before any late feedback could touch the new tests.
	g.core.gate.gate.tests = [4]comparativeTest{}
	g.core.gate.credits = [4][5]float64{}
	return nil
}

func (g *routedFeedbackJournal) predict(origin uint64, reader observation.Reader, epoch uint64) (observation.Result, error) {
	if g == nil || g.busy || g.models == nil || origin != g.core.bank.issued || origin >= 256 || origin/32 != g.publishedAt/32 {
		return observation.Result{}, errors.New("invalid journal issuance")
	}
	if g.entries[origin%routedJournalCapacity].active {
		return observation.Result{}, errors.New("journal capacity backpressure")
	}
	g.busy = true
	defer func() { g.busy = false }()
	copy := g.core
	r, err := copy.predict(origin, g.models, reader, epoch)
	if err != nil {
		return observation.Result{}, err
	}
	entry := routedJournalEntry{active: true, origin: origin, version: g.version, bank: copy.bank.pending, gate: copy.gate.gate.pending, served: r.Probability}
	// Detach immutable pending forecasts from the single-pending kernels. Only
	// drain reinstalls them, without recomputing under a newer model or weight.
	copy.bank.pending = brierBankForecast{}
	copy.bank.active = false
	copy.gate.gate.pending = comparativeForecast{}
	copy.gate.gate.active = false
	g.core = copy
	g.entries[origin%routedJournalCapacity] = entry
	g.stats.Issued++
	g.stats.Pending++
	return r, nil
}

func (g *routedFeedbackJournal) deliver(origin uint64, y bool) error {
	if g == nil || g.busy || origin >= g.core.bank.issued {
		return errors.New("invalid journal delivery")
	}
	e := g.entries[origin%routedJournalCapacity]
	if !e.active || e.origin != origin || e.ready || e.censored {
		return errors.New("unknown or settled journal delivery")
	}
	copy := *g
	copy.entries[origin%routedJournalCapacity].ready = true
	copy.entries[origin%routedJournalCapacity].outcome = y
	if err := copy.drain(); err != nil {
		return err
	}
	*g = copy
	return nil
}

// Expiry is an external policy decision. Already delivered buffered evidence
// is preserved; only waiting labels before cutoff are censored.
func (g *routedFeedbackJournal) expireBefore(cutoff uint64) error {
	if g == nil || g.busy || cutoff > g.core.bank.issued {
		return errors.New("invalid journal expiry")
	}
	copy := *g
	for i, e := range copy.entries {
		if e.active && !e.ready && e.origin < cutoff {
			copy.entries[i].censored = true
		}
	}
	if err := copy.drain(); err != nil {
		return err
	}
	*g = copy
	return nil
}

func (g *routedFeedbackJournal) drain() error {
	for g.drainAt < g.core.bank.issued {
		i := g.drainAt % routedJournalCapacity
		e := g.entries[i]
		if !e.active || e.origin != g.drainAt {
			return errors.New("journal origin gap")
		}
		if !e.ready && !e.censored {
			break
		}
		if e.censored {
			g.stats.Censored++
		} else if e.version != g.version {
			// The label may separately enter an as-of training audit. It must not
			// certify or weight a newly fitted expert using its old prediction.
			if g.carrySelector {
				g.core.bank.pending = e.bank
				g.core.bank.active = true
				if err := g.core.bank.observe(e.origin, e.outcome); err != nil {
					return err
				}
				g.stats.BankOnly++
			} else {
				g.stats.Stale++
			}
		} else {
			g.core.bank.pending = e.bank
			g.core.bank.active = true
			g.core.gate.gate.pending = e.gate
			g.core.gate.gate.active = true
			if err := g.core.observe(e.origin, e.outcome); err != nil {
				return err
			}
			g.stats.Applied++
		}
		g.entries[i] = routedJournalEntry{}
		g.stats.Pending--
		g.drainAt++
	}
	return nil
}
