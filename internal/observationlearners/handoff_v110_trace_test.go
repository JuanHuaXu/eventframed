package observationlearners

import (
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
)

type handoffV110Step struct {
	Origin                   int
	Advice, Routed           [5]float64
	Raw, MatchedRaw, FullRaw [4]float64
	Served, Matched, Full    float64
	Mask, MatchedMask        uint16
}
type handoffV110Update struct {
	Origin, Clock, Arrival      int
	IssueVersion, UpdateVersion uint64
	Raw                         [4]float64
	Before, After               [5]float64
	Outcome                     bool
}
type handoffV110Arm struct {
	Steps   []handoffV110Step
	Updates []handoffV110Update
}
type handoffV110Profile struct {
	Clock          int
	Rows           [512][4]float64
	Truth          [512]float64
	Movement, Risk [4]float64
}
type handoffV110Trace struct {
	Arms     [2]handoffV110Arm
	Profiles []handoffV110Profile
}
type tracingV110Journal struct {
	*logAdviceJournal
	parent   *logV109Record
	trace    *handoffV110Arm
	profiles *[]handoffV110Profile
}

func newTracingV110Journal(neutral bool, parent *logV109Record, trace *handoffV110Arm, profiles *[]handoffV110Profile) *tracingV110Journal {
	return &tracingV110Journal{newLogAdviceJournal(neutral), parent, trace, profiles}
}
func (g *tracingV110Journal) statsSnapshot() routedJournalStats { return g.journal.stats }
func (g *tracingV110Journal) selectorCount() uint64             { return g.received }

func handoffV110Truth(r *logV109Record, x uint16, clock int) float64 {
	name, rule := "majority3", r.Rules[0]
	if r.Case == "parity_to_majority" {
		name = "parity4"
	}
	if clock >= 128 {
		rule = r.Rules[1]
		if r.Case == "majority_to_parity" {
			name = "parity4"
		} else {
			name = "majority3"
		}
	}
	return stackV93Truth(x, rule, name)
}

func (g *tracingV110Journal) publish(e *observationExperts) error {
	old := g.journal.models
	if err := g.logAdviceJournal.publish(e); err != nil {
		return err
	}
	if g.profiles == nil {
		return nil
	}
	p := handoffV110Profile{Clock: int(g.journal.stats.Issued)}
	for x := uint16(0); x < 512; x++ {
		q := handoffV110Truth(g.parent, x, p.Clock)
		p.Truth[x] = q
		for j := 0; j < 4; j++ {
			v, err := e.models[j].Forecast(511, x)
			if err != nil {
				return err
			}
			p.Rows[x][j] = v
			p.Risk[j] += ((v-q)*(v-q) + q*(1-q)) / 512
			if old != nil {
				u, err := old.models[j].Forecast(511, x)
				if err != nil {
					return err
				}
				p.Movement[j] += math.Abs(v-u) / 512
			}
		}
	}
	*g.profiles = append(*g.profiles, p)
	return nil
}

func (g *tracingV110Journal) predict(origin uint64, r observation.Reader, epoch uint64) (observation.Result, error) {
	w := g.advice.weights()
	gate := g.journal.core.gate
	f, err := adviceRoutedWeights(&gate, origin, [4]float64{.5, .5, .5, .5}, w)
	if err != nil {
		return observation.Result{}, err
	}
	got, err := g.logAdviceJournal.predict(origin, r, epoch)
	if err != nil {
		return got, err
	}
	old := g.parent.Steps[origin]
	entry := g.journal.entries[origin%64]
	s := handoffV110Step{Origin: int(origin), Advice: w, Routed: f.Weights, Raw: entry.bank.Experts, Served: got.Probability, Mask: got.Trace[len(got.Trace)-1].Observed, MatchedMask: old.Mask[0]}
	s.Matched = .5 * s.Routed[0]
	s.Full = s.Matched
	for j := 0; j < 4; j++ {
		// These read-only counterfactuals use the parent's simulator input only
		// after issuance. They never enter acquisition, advice or the evidence gate.
		s.MatchedRaw[j], err = g.journal.models.models[j].Forecast(s.MatchedMask, old.X&s.MatchedMask)
		if err != nil {
			return got, err
		}
		s.FullRaw[j], err = g.journal.models.models[j].Forecast(511, old.X)
		if err != nil {
			return got, err
		}
		s.Matched += s.Routed[j+1] * s.MatchedRaw[j]
		s.Full += s.Routed[j+1] * s.FullRaw[j]
	}
	check := .5 * s.Routed[0]
	for j, p := range s.Raw {
		check += s.Routed[j+1] * p
	}
	if math.Abs(check-s.Served) > 1e-12 {
		return got, fmt.Errorf("traced law mismatch")
	}
	g.trace.Steps = append(g.trace.Steps, s)
	return got, nil
}

func (g *tracingV110Journal) deliver(origin uint64, y bool) error {
	e := g.journal.entries[origin%64]
	u := handoffV110Update{Origin: int(origin), Clock: int(g.advice.clock), Arrival: int(origin) + g.parent.Steps[origin].Delay, IssueVersion: e.version, UpdateVersion: g.journal.version, Raw: e.bank.Experts, Before: g.advice.weights(), Outcome: y}
	if err := g.logAdviceJournal.deliver(origin, y); err != nil {
		return err
	}
	u.After = g.advice.weights()
	g.trace.Updates = append(g.trace.Updates, u)
	return nil
}
