package observationlearners

import (
	"errors"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// Rate is fixed along a path, but uncertain. Logs retain the joint evidence;
// normalizing each four-role component separately would erase rate learning.
const hazardRateCount = 11

type hazardAdviceState struct {
	logs  [hazardRateCount][5]float64
	prior [5]float64
}

func hazardRates() [hazardRateCount]float64 {
	return [hazardRateCount]float64{0, .001, 1, .5, .25, .125, .0625, .03125, .015625, .0078125, .00390625}
}

func hazardPrior() [hazardRateCount]float64 {
	p := [hazardRateCount]float64{}
	for i := range p {
		p[i] = .05
	}
	p[1] = .5
	return p
}

func hazardLogAdd(a, b float64) float64 {
	m := math.Max(a, b)
	if math.IsInf(m, -1) {
		return m
	}
	return m + math.Log(math.Exp(a-m)+math.Exp(b-m))
}

func (s *hazardAdviceState) weights() [5]float64 {
	var logs [5]float64
	for j := range logs {
		logs[j] = math.Inf(-1)
		for h := range s.logs {
			logs[j] = hazardLogAdd(logs[j], s.logs[h][j])
		}
	}
	a := newAgedAdvice(false, false)
	a.logs = logs
	return a.weights()
}

func (s *hazardAdviceState) rateWeights() [hazardRateCount]float64 {
	var logs, weights [hazardRateCount]float64
	total := math.Inf(-1)
	for h := range logs {
		logs[h] = math.Inf(-1)
		for _, v := range s.logs[h] {
			logs[h] = hazardLogAdd(logs[h], v)
		}
		total = hazardLogAdd(total, logs[h])
	}
	for h, v := range logs {
		weights[h] = math.Exp(v - total)
	}
	return weights
}

type hazardAdvice struct {
	ready               bool
	base, next          uint64
	checkpoint, current hazardAdviceState
	entries             [routedJournalCapacity]markovAdviceEntry
}

func newHazardAdvice(prior [hazardRateCount]float64) (hazardAdvice, error) {
	total := 0.
	for _, p := range prior {
		if math.IsNaN(p) || p < 0 || p > 1 {
			return hazardAdvice{}, errors.New("invalid hazard prior")
		}
		total += p
	}
	if math.Abs(total-1) > 1e-12 {
		return hazardAdvice{}, errors.New("unnormalized hazard prior")
	}
	s := hazardAdviceState{prior: newAgedAdvice(false, false).prior}
	for h, p := range prior {
		for j, q := range s.prior {
			s.logs[h][j] = math.Log(p/total) + math.Log(q)
		}
	}
	return hazardAdvice{ready: true, checkpoint: s, current: s}, nil
}

// Every rate uses the same immutable emission. Transitions preserve the rate
// marginal; only evidence can change it. A missing label has unit emission.
func hazardAdviceStep(s *hazardAdviceState, e markovAdviceEntry) error {
	copy := *s
	rates := hazardRates()
	maximum := math.Inf(-1)
	for h, alpha := range rates {
		row := copy.logs[h]
		mass := math.Inf(-1)
		for j := 1; j < 5; j++ {
			if e.known {
				p := e.raw[j-1]
				ll := math.Log1p(-p)
				if e.outcome {
					ll = math.Log(p)
				}
				row[j] += ll
			}
			mass = hazardLogAdd(mass, row[j])
		}
		for j := range row {
			v := math.Inf(-1)
			if !math.IsInf(mass, -1) && s.prior[j] > 0 {
				if alpha == 0 {
					// Preserve recoverable log evidence when conditional weights underflow.
					v = row[j]
				} else {
					conditional := (1-alpha)*math.Exp(row[j]-mass) + alpha*s.prior[j]
					// Restore the rate's evidence mass after its conditional transition.
					// Omitting this term would erase learning about the rate itself.
					v = mass + math.Log(conditional)
				}
			}
			copy.logs[h][j] = v
			maximum = math.Max(maximum, v)
		}
	}
	if math.IsInf(maximum, 0) || math.IsNaN(maximum) {
		return errors.New("invalid hazard joint evidence")
	}
	for h := range copy.logs {
		for j := range copy.logs[h] {
			copy.logs[h][j] -= maximum
		}
	}
	*s = copy
	return nil
}

func (f *hazardAdvice) issue(origin uint64, raw [4]float64) error {
	if f == nil || !f.ready || origin != f.next || origin >= 256 || f.next-f.base >= routedJournalCapacity {
		return errors.New("invalid Hazard issue or capacity")
	}
	for _, p := range raw {
		if math.IsNaN(p) || p <= 0 || p >= 1 {
			return errors.New("invalid Hazard emission")
		}
	}
	copy := *f
	e := markovAdviceEntry{active: true, origin: origin, raw: raw}
	if copy.entries[origin%routedJournalCapacity].active {
		return errors.New("occupied Hazard entry")
	}
	if err := hazardAdviceStep(&copy.current, e); err != nil {
		return err
	}
	copy.entries[origin%routedJournalCapacity] = e
	copy.next++
	*f = copy
	return nil
}

// Recomputing messages does not readmit evidence. A label is admitted once;
// subsequent passes only evaluate the already-declared conditional model.
func (f *hazardAdvice) rebuild() error {
	s := f.checkpoint
	base := f.base
	for origin := f.base; origin < f.next; origin++ {
		i := origin % routedJournalCapacity
		e := f.entries[i]
		if !e.active || e.origin != origin || (e.known && e.censored) {
			return errors.New("Hazard suffix gap")
		}
		if err := hazardAdviceStep(&s, e); err != nil {
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

func (f *hazardAdvice) deliver(origin uint64, y bool) error {
	if f == nil || !f.ready || origin < f.base || origin >= f.next {
		return errors.New("invalid Hazard delivery")
	}
	e := f.entries[origin%routedJournalCapacity]
	if !e.active || e.origin != origin || e.known || e.censored {
		return errors.New("settled Hazard delivery")
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

func (f *hazardAdvice) expireBefore(cutoff uint64) error {
	if f == nil || !f.ready || cutoff > f.next {
		return errors.New("invalid Hazard expiry")
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

type hazardAdviceJournal struct {
	agedAdviceJournal
	filter hazardAdvice
}

func newHazardAdviceJournal() *hazardAdviceJournal {
	f, _ := newHazardAdvice(hazardPrior())
	return &hazardAdviceJournal{agedAdviceJournal: *newAgedAdviceJournal(false, false), filter: f}
}

func (g *hazardAdviceJournal) syncAdvice() error {
	if g.filter.next != g.journal.stats.Issued || g.filter.base != g.journal.drainAt || g.filter.next-g.filter.base != g.journal.stats.Pending {
		return errors.New("Hazard journal frontier mismatch")
	}
	// Preserve the journal's wall clock while updating its next-issue law.
	w := g.filter.current.weights()
	for j, p := range w {
		g.advice.logs[j] = math.Log(p)
	}
	return nil
}

func (g *hazardAdviceJournal) predict(origin uint64, reader observation.Reader, epoch uint64) (observation.Result, error) {
	if g == nil || g.journal.busy {
		return observation.Result{}, errors.New("invalid Hazard prediction")
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

func (g *hazardAdviceJournal) deliver(origin uint64, y bool) error {
	if g == nil || g.journal.busy {
		return errors.New("invalid Hazard journal delivery")
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

func (g *hazardAdviceJournal) expireBefore(cutoff uint64) error {
	if g == nil || g.journal.busy {
		return errors.New("invalid Hazard journal expiry")
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
