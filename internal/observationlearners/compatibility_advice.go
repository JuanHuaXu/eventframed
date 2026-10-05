package observationlearners

import (
	"errors"
	"math"
)

// Research-only compatibility modulation, not a reuse certificate. The common
// input law must be unchanged; role identity alone does not establish this.
func forecastCompatibility(old, next *observationExperts) ([5]float64, error) {
	gamma := [5]float64{1, 1, 1, 1, 1}
	if old == nil || next == nil || !old.ready || !next.ready {
		return gamma, errors.New("missing compatibility publication")
	}
	var affinity [4]float64
	same := [4]bool{true, true, true, true}
	for x := uint16(0); x < 512; x++ {
		i := partialIndex(511, x)
		mu := old.models[0].cells[i].mass / old.roots[0]
		if !jointClose(mu, next.models[0].cells[i].mass/next.roots[0]) {
			return gamma, errors.New("compatibility input law changed")
		}
		for j := range affinity {
			p, err := old.models[j].Forecast(511, x)
			if err != nil {
				return gamma, err
			}
			q, err := next.models[j].Forecast(511, x)
			if err != nil {
				return gamma, err
			}
			same[j] = same[j] && p == q
			affinity[j] += mu * (math.Sqrt(p*q) + math.Sqrt((1-p)*(1-q)))
		}
	}
	for j, a := range affinity {
		if math.IsNaN(a) || a <= 0 || a > 1+1e-12 {
			return gamma, errors.New("invalid compatibility affinity")
		}
		if !same[j] {
			gamma[j+1] = math.Pow(math.Min(1, a), 32)
		}
	}
	return gamma, nil
}

func validCompatibility(s *agedAdvice, gamma [5]float64) bool {
	if s == nil || !s.ready || s.aged {
		return false
	}
	for _, v := range gamma {
		if math.IsNaN(v) || v < 0 || v > 1 {
			return false
		}
	}
	return true
}

func compatibilityIdentity(gamma [5]float64) bool {
	return gamma == [5]float64{1, 1, 1, 1, 1}
}

func transferCompatibility(s *agedAdvice, gamma [5]float64) error {
	if !validCompatibility(s, gamma) {
		return errors.New("invalid compatibility transfer")
	}
	if compatibilityIdentity(gamma) {
		return nil
	}
	copy := *s
	maximum := math.Inf(-1)
	for _, v := range s.logs {
		maximum = math.Max(maximum, v)
	}
	sum := 0.
	for _, v := range s.logs {
		sum += math.Exp(v - maximum)
	}
	logSum := math.Log(sum)
	// Normalize in log space before role-specific scaling. Neither a common
	// log offset nor exp underflow is allowed to erase recoverable evidence.
	for j, p := range s.prior {
		copy.logs[j] = math.Log(p)
		if p > 0 && gamma[j] > 0 {
			copy.logs[j] += gamma[j] * ((s.logs[j] - maximum) - logSum - math.Log(p))
		}
	}
	*s = copy
	return nil
}

func observeCompatibility(s *agedAdvice, origin uint64, raw [4]float64, y bool, gamma [5]float64) error {
	if !validCompatibility(s, gamma) {
		return errors.New("invalid pending compatibility")
	}
	if compatibilityIdentity(gamma) {
		return observeLogAdvice(s, origin, raw, y, .001)
	}
	// Validate the entire original update before mutating any state.
	probe := *s
	if err := observeLogAdvice(&probe, origin, raw, y, .001); err != nil {
		return err
	}
	copy := *s
	for j, p := range [5]float64{.5, raw[0], raw[1], raw[2], raw[3]} {
		ll := math.Log1p(-p)
		if y {
			ll = math.Log(p)
		}
		// Attenuate a fixed-reference likelihood ratio, not a negative loss.
		// Ignoring a role's evidence must not reward it merely for avoiding loss.
		copy.logs[j] += gamma[j] * (ll - math.Log(.5))
	}
	w := copy.weights()
	for j := range copy.logs {
		copy.logs[j] = math.Log(.999*w[j] + .001*copy.prior[j])
	}
	*s = copy
	return nil
}

// Eight publication versions, fixed model roles, single owner. Each row caches
// log products from that issue version to the current publication. Publication
// updates at most 8*5 coefficients; each arrival reads just five coefficients.
type compatibilityAdviceJournal struct {
	agedAdviceJournal
	pendingTransfer bool
	logTransfer     [9][5]float64
}

func newCompatibilityAdviceJournal(pending bool) *compatibilityAdviceJournal {
	return &compatibilityAdviceJournal{agedAdviceJournal: *newAgedAdviceJournal(false, false), pendingTransfer: pending}
}

func (g *compatibilityAdviceJournal) publish(e *observationExperts) error {
	if g == nil || g.journal.busy {
		return errors.New("invalid compatibility publication")
	}
	copy := *g
	old := copy.journal.models
	if err := copy.agedAdviceJournal.publish(e); err != nil {
		return err
	}
	if copy.journal.version > 8 {
		return errors.New("compatibility version cap")
	}
	if old != nil {
		gamma, err := forecastCompatibility(old, e)
		if err != nil {
			return err
		}
		if err := transferCompatibility(&copy.advice, gamma); err != nil {
			return err
		}
		for v := uint64(1); v < copy.journal.version; v++ {
			for j, factor := range gamma {
				copy.logTransfer[v][j] += math.Log(factor)
			}
		}
	}
	*g = copy
	return nil
}

func (g *compatibilityAdviceJournal) deliver(origin uint64, y bool) error {
	if g == nil || g.journal.busy {
		return errors.New("invalid compatibility delivery")
	}
	copy := *g
	entry := copy.journal.entries[origin%routedJournalCapacity]
	if err := copy.journal.deliver(origin, y); err != nil {
		return err
	}
	if entry.version == 0 || entry.version > copy.journal.version || entry.version > 8 {
		return errors.New("invalid compatibility issue version")
	}
	gamma := [5]float64{1, 1, 1, 1, 1}
	if copy.pendingTransfer {
		for j, v := range copy.logTransfer[entry.version] {
			gamma[j] = math.Exp(v)
		}
	}
	if err := observeCompatibility(&copy.advice, origin, entry.bank.Experts, y, gamma); err != nil {
		return err
	}
	copy.received++
	*g = copy
	return nil
}
