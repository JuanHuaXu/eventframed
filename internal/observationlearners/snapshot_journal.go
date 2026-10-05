package observationlearners

import (
	"errors"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type snapshotJournalEntry struct {
	active, known, censored, outcome bool
	origin                           uint64
	epoch                            int
	ids                              [8]uint64
	raw                              [8]float64
	served                           float64
}

// Delayed advice refilters on arrival; rejection evidence drains in origin order.
// The two consumers share immutable issued probabilities, not publication slots.
type snapshotJournal struct {
	bank                     snapshotBank
	gate                     snapshotGate
	clock, drainAt, received uint64
	busy                     bool
	stats                    routedJournalStats
	entries                  [64]snapshotJournalEntry
}

func newSnapshotJournal() *snapshotJournal {
	return &snapshotJournal{bank: *newSnapshotBank()}
}

func (g *snapshotJournal) setClock(clock uint64) error {
	if g == nil || g.busy || clock < g.clock || clock > 288 || g.stats.Issued < 256 && clock > g.stats.Issued {
		return errors.New("invalid snapshot clock")
	}
	g.clock = clock
	return nil
}

func (g *snapshotJournal) publish(e *observationExperts) error {
	if g == nil || g.busy || e == nil || !e.ready || g.clock != g.stats.Issued || g.clock%32 != 0 {
		return errors.New("invalid snapshot journal publication")
	}
	c := *g
	var models [4]*ConditionalForest
	for j := range models {
		models[j] = &e.models[j]
	}
	version := int(g.clock / 32)
	if err := c.bank.publish(version, models); err != nil {
		return err
	}
	if err := c.gate.reset(version, c.bank.filter.current.ids, snapshotGateBoundary); err != nil {
		return err
	}
	*g = c
	return nil
}

func (g *snapshotJournal) consistent() error {
	f := &g.bank.filter
	if f.nextOrigin != g.stats.Issued || f.baseOrigin != g.drainAt || g.stats.Pending != g.stats.Issued-g.drainAt || g.stats.Issued != g.stats.Pending+g.stats.Applied+g.stats.BankOnly+g.stats.Stale+g.stats.Censored {
		return errors.New("snapshot journal accounting mismatch")
	}
	return nil
}

func (g *snapshotJournal) predict(origin uint64, reader observation.Reader, epoch uint64) (observation.Result, error) {
	if g == nil || g.busy || origin != g.clock || origin != g.stats.Issued || origin >= 256 || g.entries[origin%64].active {
		return observation.Result{}, errors.New("invalid snapshot prediction")
	}
	c := *g
	g.busy = true
	defer func() { g.busy = false }()
	base, err := c.bank.snapshot()
	if err != nil {
		return observation.Result{}, err
	}
	w, err := c.gate.route(base.ids, base.weights)
	if err != nil {
		return observation.Result{}, err
	}
	law := snapshotRoutedLaw{base: base, weights: w}
	result, err := runSnapshotObserver(law, reader, epoch)
	if err != nil {
		return observation.Result{}, err
	}
	if len(result.Trace) == 0 {
		return observation.Result{}, errors.New("empty snapshot observation")
	}
	last := result.Trace[len(result.Trace)-1]
	raw, err := base.raw(last.Observed, last.Values)
	if err != nil {
		return observation.Result{}, err
	}
	served := .5 * w[0]
	for i, p := range raw {
		served += w[i+1] * p
	}
	if math.Abs(served-result.Probability) > 1e-12 {
		return observation.Result{}, errors.New("snapshot acquisition/served law mismatch")
	}
	if err := c.bank.issue(origin, base, last.Observed, last.Values); err != nil {
		return observation.Result{}, err
	}
	c.entries[origin%64] = snapshotJournalEntry{active: true, origin: origin, epoch: c.gate.epoch, ids: base.ids, raw: raw, served: served}
	c.stats.Issued++
	c.stats.Pending++
	if err := c.consistent(); err != nil {
		return observation.Result{}, err
	}
	*g = c
	result.Probability = served
	return result, nil
}

func (g *snapshotJournal) drain() error {
	for g.drainAt < g.stats.Issued {
		i := g.drainAt % 64
		e := g.entries[i]
		if !e.active || e.origin != g.drainAt {
			return errors.New("snapshot journal gap")
		}
		if !e.known && !e.censored {
			break
		}
		if e.censored {
			g.stats.Censored++
		} else if e.epoch == g.gate.epoch {
			if err := g.gate.observe(e.epoch, e.ids, e.raw, e.outcome); err != nil {
				return err
			}
			g.stats.Applied++
		} else {
			g.stats.BankOnly++
		}
		g.stats.Pending--
		g.drainAt++
		g.entries[i] = snapshotJournalEntry{}
	}
	return g.consistent()
}

func (g *snapshotJournal) deliver(origin uint64, y bool) error {
	if g == nil || g.busy || origin >= g.stats.Issued || origin < g.drainAt {
		return errors.New("invalid snapshot journal feedback")
	}
	e := g.entries[origin%64]
	if !e.active || e.origin != origin || e.known || e.censored {
		return errors.New("settled snapshot journal feedback")
	}
	c := *g
	if err := c.bank.filter.deliver(origin, y); err != nil {
		return err
	}
	c.entries[origin%64].known = true
	c.entries[origin%64].outcome = y
	c.received++
	if err := c.drain(); err != nil {
		return err
	}
	*g = c
	return nil
}

func (g *snapshotJournal) expireBefore(cutoff uint64) error {
	if g == nil || g.busy || cutoff > g.stats.Issued {
		return errors.New("invalid snapshot journal expiry")
	}
	c := *g
	if err := c.bank.filter.expireBefore(cutoff); err != nil {
		return err
	}
	for i := range c.entries {
		e := &c.entries[i]
		if e.active && !e.known && e.origin < cutoff {
			e.censored = true
		}
	}
	if err := c.drain(); err != nil {
		return err
	}
	*g = c
	return nil
}

func (g *snapshotJournal) statsSnapshot() routedJournalStats { return g.stats }
func (g *snapshotJournal) selectorCount() uint64             { return g.received }
