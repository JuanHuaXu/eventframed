package researchcalendar

import (
	"errors"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
)

// PackPriority keeps selection-only priority separate from exposed numeric
// scores. It delegates token, diversity and evidence occupancy rules to packing.
// The plan must belong to this exact candidate ordering and current snapshot;
// this component validates identities, not the caller-owned snapshot lifecycle.
func PackPriority(candidates []model.Candidate, plan PriorityPlan, keys map[string]string, packK, recallK, budget int, policy packing.Policy) (packing.Result, error) {
	n := len(candidates)
	if n == 0 || n > 200 || len(plan.Order) != n || len(plan.Decisions) != n || plan.Method != "research/calendar-priority-v1" || plan.CalibrationStatus != "not_evaluated" || packK <= 0 || recallK < packK || recallK > 200 || budget <= 0 {
		return packing.Result{}, errors.New("invalid priority packing contract")
	}
	for _, v := range []float64{policy.DiversityPenalty, policy.PriorityPenalty} {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return packing.Result{}, errors.New("unsupported priority packing penalty")
		}
	}
	if policy.AdaptiveEnabled && policy.MaxPack <= 0 {
		return packing.Result{}, errors.New("invalid adaptive cap")
	}
	seen := make(map[string]bool, n)
	indices := make([]bool, n)
	bad := 0
	for i, c := range candidates {
		if c.Event.ID == "" || seen[c.Event.ID] || plan.Decisions[i].EventID != c.Event.ID || math.IsNaN(c.Score) || math.IsInf(c.Score, 0) || c.Score < 0 || c.Score > 1 {
			return packing.Result{}, errors.New("priority candidate mismatch")
		}
		seen[c.Event.ID] = true
		switch plan.Decisions[i].State {
		case "contradicted":
			bad++
		case "compatible", "unknown":
		default:
			return packing.Result{}, errors.New("invalid temporal state")
		}
		j := plan.Order[i]
		if j < 0 || j >= n || indices[j] {
			return packing.Result{}, errors.New("invalid priority permutation")
		}
		indices[j] = true
	}
	if plan.AllContradicted != (bad == n) {
		return packing.Result{}, errors.New("contradiction flag mismatch")
	}
	if bad == 0 || bad == n {
		return packing.Select(candidates, keys, packK, recallK, budget, policy), nil
	}
	// Expansion uses original scores, before any selection-only priority encoding.
	expanded := false
	if policy.AdaptiveEnabled {
		expanded = packing.Select(candidates, keys, packK, recallK, budget, policy).Expanded
		if expanded {
			packK = min(2*packK, recallK, policy.MaxPack)
		}
	}
	local := make([]model.Candidate, n)
	original := make(map[string]model.Candidate, n)
	for i, j := range plan.Order {
		local[i] = candidates[j]
		original[local[i].Event.ID] = candidates[j]
		// With scores in[0,1] and penalties in[0,1], offset3 enforces
		// lexicographic class priority inside the existing diversity selector.
		if plan.Decisions[j].State != "contradicted" {
			local[i].Score += 3
		}
	}
	policy.AdaptiveEnabled = false
	out := packing.Select(local, keys, packK, recallK, budget, policy)
	out.Expanded = expanded
	for i, c := range out.Candidates {
		out.Candidates[i] = original[c.Event.ID]
	}
	return out, nil
}
