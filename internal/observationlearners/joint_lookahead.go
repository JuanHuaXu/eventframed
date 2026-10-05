package observationlearners

import (
	"errors"
	"math/bits"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// Each complete read makes the observed and paid masks identical. With one
// fixed total budget, remaining budget is therefore determined by the mask.
// Memoization is valid only within this immutable forecast snapshot.
type jointPlanner struct {
	law             observationMixture
	budget          int
	value           [19683]float64
	action          [19683]uint8 // 0 unseen, 1 terminal, 2+3*scope+depth for a view
	nodes, branches int
}

func newJointPlanner(m observationMixture, budget int) (*jointPlanner, error) {
	if budget < 1 || budget > 6 {
		return nil, errors.New("invalid lookahead budget")
	}
	if _, err := m.Forecast(0, 0); err != nil {
		return nil, err
	}
	return &jointPlanner{law: m, budget: budget}, nil
}

// solve minimizes expected terminal entropy, not immediate gain per cost.
// Equal values retain the first scope/depth in the frozen enumeration order.
// This is a small-domain exact planner, not an arbitrary-domain runtime bound.
func (p *jointPlanner) solve(mask, values uint16) (float64, uint8) {
	i := partialIndex(mask, values)
	if p.action[i] != 0 {
		return p.value[i], p.action[i]
	}
	p.nodes++
	c := p.law.cell(mask, values)
	prob := c.weighted / c.mass
	value, action := jointObservationEntropy(prob), uint8(1)
	used := bits.OnesCount16(mask)
	if used < p.budget && !(used > 0 && (prob <= .1 || prob >= .9)) {
		found := false
		for scope := 0; scope < 3; scope++ {
			for depth := 0; depth < 3; depth++ {
				view := observation.View{Scope: scope, Depth: depth}
				added := view.Mask() &^ mask
				cost := bits.OnesCount16(added)
				if cost == 0 || used+cost > p.budget || (mask == 0 && (scope != 0 || depth != 0)) {
					continue
				}
				expected := 0.
				for sub := added; ; sub = (sub - 1) & added {
					child := p.law.cell(mask|added, values|sub)
					terminal, _ := p.solve(mask|added, values|sub)
					expected += child.mass / c.mass * terminal
					p.branches++
					if sub == 0 {
						break
					}
				}
				if !found || expected < value-1e-12 {
					value, action, found = expected, uint8(2+3*scope+depth), true
				}
			}
		}
	}
	p.value[i], p.action[i] = value, action
	return value, action
}

type jointLookaheadResult struct {
	observation.Result
	Nodes, Branches int
}

func runJointLookahead(m observationMixture, reader observation.Reader, epoch uint64) (jointLookaheadResult, error) {
	if reader == nil {
		return jointLookaheadResult{}, errors.New("missing lookahead reader")
	}
	planner, err := newJointPlanner(m, 6)
	if err != nil {
		return jointLookaheadResult{}, err
	}
	p, _ := m.Forecast(0, 0)
	r := jointLookaheadResult{Result: observation.Result{Probability: p}}
	var mask, values uint16
	for {
		if reader.Epoch() != epoch {
			return jointLookaheadResult{}, errors.New("stale lookahead snapshot")
		}
		_, action := planner.solve(mask, values)
		if action == 1 {
			r.Stop = "no_affordable_view"
			if r.Cost >= 6 {
				r.Stop = "budget"
			} else if r.Cost > 0 && (r.Probability <= .1 || r.Probability >= .9) {
				r.Stop = "confidence"
			}
			break
		}
		view := observation.View{Scope: int(action-2) / 3, Depth: int(action-2) % 3}
		// Planning reads only joint-table cells; this is the sole evidence read.
		seen, v, err := reader.Read(view)
		if err != nil {
			return jointLookaheadResult{}, err
		}
		if reader.Epoch() != epoch || seen != view.Mask() || v&^seen != 0 || (v^values)&(seen&mask) != 0 {
			return jointLookaheadResult{}, errors.New("invalid lookahead observation")
		}
		r.Cost += bits.OnesCount16(seen &^ mask)
		mask |= seen
		values |= v
		r.Probability, _ = m.Forecast(mask, values)
		r.Trace = append(r.Trace, observation.Step{View: view, Attempted: mask, Observed: mask, Values: values, Probability: r.Probability})
	}
	r.Observed = bits.OnesCount16(mask)
	r.Nodes, r.Branches = planner.nodes, planner.branches
	return r, nil
}

func (s *routedObservationState) predictLookahead(sequence uint64, e *observationExperts, reader observation.Reader, epoch uint64) (jointLookaheadResult, error) {
	m, err := s.preview(sequence, e)
	if err != nil {
		return jointLookaheadResult{}, err
	}
	s.busy = true
	defer func() { s.busy = false }()
	r, err := runJointLookahead(m, reader, epoch)
	if err != nil {
		return jointLookaheadResult{}, err
	}
	last := r.Trace[len(r.Trace)-1]
	f, err := s.finish(sequence, m, last.Observed, last.Values)
	if err != nil {
		return jointLookaheadResult{}, err
	}
	r.Probability = f.P
	return r, nil
}
