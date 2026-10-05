package observationgate

import (
	"errors"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

const forecastJournalCapacity = 64

type forecastJournalEntry struct {
	active     bool
	origin     int
	generation uint64
	forecast   subsetPending
}
type forecastJournalStats struct{ Applied, Stale, Censored, Pending int }

// forecastJournal is owned by one serialized research stream. Issuance and
// feedback have separate clocks; immutable probabilities bridge the two.
type forecastJournal struct {
	base           *observation.Model
	enabled, split bool
	mix, inner     bayes.ForecastMix
	models         observationpreserved.Models
	subset         *observationlearners.ConditionalForest
	generation     uint64
	issued         int
	entries        [forecastJournalCapacity]forecastJournalEntry
	stats          forecastJournalStats
}

func (g *forecastJournal) publish(models observationpreserved.Models, subset *observationlearners.ConditionalForest) error {
	if models.Version <= g.models.Version || (subset != nil && models.Short == nil) {
		return errors.New("invalid model publication")
	}
	// References point to immutable fitted objects. Pending entries retain
	// probabilities, so publication cannot rewrite their old predictions.
	g.models, g.subset = models, subset
	return nil
}

func (g *forecastJournal) predict(reader observation.Reader, origin int, seed int64) (observationpreserved.Prediction, error) {
	if g.base == nil || origin != g.issued || origin < 0 || origin == int(^uint(0)>>1) {
		return observationpreserved.Prediction{}, errors.New("invalid journal issuance")
	}
	slot := &g.entries[origin%forecastJournalCapacity]
	if slot.active {
		return observationpreserved.Prediction{}, errors.New("forecast journal backpressure")
	}
	// Reuse the frozen prediction transition on a value snapshot. Its temporary
	// pending field is copied into the journal only after every read succeeds.
	s := subsetState{base: g.base, enabled: g.enabled, split: g.split, mix: g.mix, inner: g.inner, next: origin}
	p, e := s.predict(reader, g.models, g.subset, origin, seed)
	if e != nil {
		return observationpreserved.Prediction{}, e
	}
	*slot = forecastJournalEntry{active: true, origin: origin, generation: g.generation, forecast: *s.pending}
	g.issued++
	g.stats.Pending++
	return p, nil
}

func (g *forecastJournal) entry(origin int) (*forecastJournalEntry, error) {
	if origin < 0 || origin >= g.issued {
		return nil, errors.New("unknown forecast origin")
	}
	s := &g.entries[origin%forecastJournalCapacity]
	if !s.active || s.origin != origin {
		return nil, errors.New("settled or overwritten forecast origin")
	}
	return s, nil
}

func (g *forecastJournal) deliver(origin int, y, authorized bool) (string, error) {
	entry, e := g.entry(origin)
	if e != nil {
		return "", e
	}
	if entry.forecast.p.Version != g.models.Version || entry.generation != g.generation {
		*entry = forecastJournalEntry{}
		g.stats.Pending--
		g.stats.Stale++
		return "stale", nil
	}
	// Expert updates use the probabilities originally emitted, even if feedback
	// arrives out of order. They are online selector updates, not a claim of an
	// exact fixed-model Bayesian posterior for event-time ordering.
	pending := entry.forecast
	s := subsetState{base: g.base, enabled: g.enabled, split: g.split, mix: g.mix, inner: g.inner, next: origin, pending: &pending}
	if e := s.observe(origin, y, authorized); e != nil {
		return "", e
	}
	if s.split != g.split {
		g.generation++
	}
	g.mix, g.inner, g.split = s.mix, s.inner, s.split
	*entry = forecastJournalEntry{}
	g.stats.Pending--
	g.stats.Applied++
	return "applied", nil
}

func (g *forecastJournal) censor(origin int) error {
	entry, e := g.entry(origin)
	if e != nil {
		return e
	}
	*entry = forecastJournalEntry{}
	g.stats.Pending--
	g.stats.Censored++
	return nil
}

func (g *forecastJournal) expireBefore(cutoff int) (int, error) {
	if cutoff < 0 || cutoff > g.issued {
		return 0, errors.New("invalid forecast expiry cutoff")
	}
	n := 0
	for i := range g.entries {
		e := &g.entries[i]
		if e.active && e.origin < cutoff {
			*e = forecastJournalEntry{}
			n++
		}
	}
	g.stats.Pending -= n
	g.stats.Censored += n
	return n, nil
}
