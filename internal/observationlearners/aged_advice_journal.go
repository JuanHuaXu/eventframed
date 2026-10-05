package observationlearners

import (
	"errors"
	"fmt"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// Single-owner research ablation. Advice learns on arrival; the frozen journal
// still orders, expires and version-binds the unchanged comparative tests.
// No new test allocation, relaxed rejection threshold or oracle reset is added.
type agedAdviceJournal struct {
	journal  routedFeedbackJournal
	advice   agedAdvice
	received uint64
}

func newAgedAdviceJournal(neutral, aged bool) *agedAdviceJournal {
	return &agedAdviceJournal{journal: *newRoleRoutedFeedbackJournal(), advice: newAgedAdvice(neutral, aged)}
}

func (g *agedAdviceJournal) setClock(clock uint64) error {
	if g == nil || g.journal.busy || (g.journal.stats.Issued < 256 && clock > g.journal.stats.Issued) {
		return errors.New("invalid journal event clock")
	}
	return g.advice.advance(clock)
}

func (g *agedAdviceJournal) publish(e *observationExperts) error {
	if g == nil || g.advice.clock != g.journal.stats.Issued {
		return errors.New("publication clock mismatch")
	}
	return g.journal.publish(e)
}

func adviceRoutedWeights(g *evidenceRouting, sequence uint64, raw [4]float64, w [5]float64) (routedForecast, error) {
	total := 0.
	for _, v := range w {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return routedForecast{}, errors.New("invalid advice mixture weight")
		}
		total += v
	}
	if math.Abs(total-1) > 1e-12 {
		return routedForecast{}, errors.New("invalid advice mixture total")
	}
	sum := 0.
	for j := 1; j < 5; j++ {
		sum += w[j]
	}
	if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
		return routedForecast{}, errors.New("invalid advice raw mass")
	}
	var inner [4]float64
	for j := range inner {
		inner[j] = w[j+1] / sum
	}
	f, err := g.predict(sequence, raw, inner)
	if err != nil {
		return f, err
	}
	// Preserve rejection authority: direct neutral mass cannot resurrect a raw
	// model, and raw mass still passes through the unchanged rejection routing.
	for j := range f.Weights {
		f.Weights[j] *= sum
	}
	f.Weights[0] += w[0]
	// Adding the direct-neutral and routed masses can leave 1+one ULP when
	// every raw expert is rejected. Normalize the composed law before validation.
	total = 0
	for _, v := range f.Weights {
		total += v
	}
	for j := range f.Weights {
		f.Weights[j] /= total
	}
	f.P = .5 * f.Weights[0]
	f.Neutral = true
	for j, p := range raw {
		f.P += f.Weights[j+1] * p
		f.Neutral = f.Neutral && f.Weights[j+1] == 0
	}
	return f, nil
}

func (g *agedAdviceJournal) predict(origin uint64, reader observation.Reader, epoch uint64) (observation.Result, error) {
	if g == nil || !g.advice.ready || g.journal.busy || g.journal.models == nil || origin != g.advice.clock || origin != g.journal.core.bank.issued || origin >= 256 || origin/32 != g.journal.publishedAt/32 {
		return observation.Result{}, errors.New("invalid advice journal issuance")
	}
	if g.journal.entries[origin%64].active {
		return observation.Result{}, errors.New("journal capacity backpressure")
	}
	copy := g.journal
	g.journal.busy = true
	defer func() { g.journal.busy = false }()
	weights := g.advice.weights()
	preview := copy.core.gate
	proposed, err := adviceRoutedWeights(&preview, origin, [4]float64{.5, .5, .5, .5}, weights)
	if err != nil {
		return observation.Result{}, err
	}
	m, err := copy.models.snapshot(proposed.Weights)
	if err != nil {
		return observation.Result{}, fmt.Errorf("advice snapshot %v: %w", proposed.Weights, err)
	}
	result, err := runJointObserver(m, reader, epoch)
	if err != nil {
		return observation.Result{}, err
	}
	last := result.Trace[len(result.Trace)-1]
	var raw [4]float64
	for j := range raw {
		raw[j], err = copy.models.models[j].Forecast(last.Observed, last.Values)
		if err != nil {
			return observation.Result{}, err
		}
	}
	served, err := adviceRoutedWeights(&copy.core.gate, origin, raw, weights)
	if err != nil {
		return observation.Result{}, err
	}
	if math.Abs(served.P-result.Probability) > 1e-12 {
		return observation.Result{}, errors.New("advice served law mismatch")
	}
	for j, w := range served.Weights {
		if math.Abs(w-m.weights[j]) > 1e-12 {
			return observation.Result{}, errors.New("advice acquisition law mismatch")
		}
	}
	// This bank is retained only as the frozen journal's lifecycle carrier.
	// Its temporary drain updates never choose the new policy's mixture weights.
	bank, err := copy.core.bank.predict(origin, raw[:])
	if err != nil {
		return observation.Result{}, err
	}
	copy.entries[origin%64] = routedJournalEntry{active: true, origin: origin, version: copy.version, bank: bank, gate: copy.core.gate.gate.pending, served: served.P}
	copy.core.bank.pending = brierBankForecast{}
	copy.core.bank.active = false
	copy.core.gate.gate.pending = comparativeForecast{}
	copy.core.gate.gate.active = false
	copy.stats.Issued++
	copy.stats.Pending++
	g.journal = copy
	result.Probability = served.P
	return result, nil
}

func (g *agedAdviceJournal) deliver(origin uint64, y bool) error {
	if g == nil || g.journal.busy {
		return errors.New("invalid advice journal delivery")
	}
	copy := *g
	entry := copy.journal.entries[origin%64]
	if err := copy.journal.deliver(origin, y); err != nil {
		return err
	}
	if err := copy.advice.observe(origin, entry.bank.Experts, y); err != nil {
		return err
	}
	copy.received++
	*g = copy
	return nil
}

func (g *agedAdviceJournal) expireBefore(cutoff uint64) error {
	if g == nil {
		return errors.New("nil advice journal")
	}
	return g.journal.expireBefore(cutoff)
}
