package researchcalendar

import (
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"math"
)

// PackTaskLexical uses lexical scores only inside selection. Adaptive expansion
// reads the ORIGINAL boundary laws/order before lexical permutation. Returned
// records are restored in full; token and evidence limits still use packing.
func PackTaskLexical(candidates []model.Candidate, plan PriorityPlan, lex LexicalResult, keys map[string]string, packK, recallK, budget int, policy packing.Policy) (packing.Result, error) {
	n := len(candidates)
	if n == 0 || n > 200 || len(plan.Decisions) != n || len(lex.Order) != n || len(lex.Scores) != n || lex.Method != "what-lexical-v2" || packK <= 0 || recallK < packK || recallK > 200 || budget <= 0 {
		return packing.Result{}, errors.New("invalid lexical packing contract")
	}
	if policy.AdaptiveEnabled && policy.MaxPack <= 0 {
		return packing.Result{}, errors.New("invalid adaptive cap")
	}
	original := map[string]model.Candidate{}
	used := make([]bool, n)
	local := make([]model.Candidate, n)
	rebound := plan
	rebound.Order = make([]int, n)
	rebound.Decisions = make([]TemporalDecision, n)
	for i, c := range candidates {
		if c.Event.ID == "" || plan.Decisions[i].EventID != c.Event.ID || math.IsNaN(c.Score) || math.IsInf(c.Score, 0) || c.Score < 0 || c.Score > 1 {
			return packing.Result{}, errors.New("invalid original candidate")
		}
		if _, exists := original[c.Event.ID]; exists {
			return packing.Result{}, errors.New("duplicate original candidate")
		}
		original[c.Event.ID] = c
	}
	for i, j := range lex.Order {
		if j < 0 || j >= n || used[j] {
			return packing.Result{}, errors.New("invalid lexical permutation")
		}
		used[j] = true
		score := lex.Scores[j]
		if math.IsNaN(score) || math.IsInf(score, 0) || score < 0 || score > 1+1e-12 {
			return packing.Result{}, errors.New("invalid lexical score")
		}
		local[i] = candidates[j]
		local[i].Score = math.Min(1, score)
		rebound.Decisions[i] = plan.Decisions[j]
		rebound.Order[i] = i
	}
	expanded := packing.ResearchExpansion(candidates, packK, recallK, policy)
	if expanded {
		packK = min(2*packK, recallK, policy.MaxPack)
	}
	policy.AdaptiveEnabled = false
	result, err := PackTaskRole(local, rebound, keys, packK, recallK, budget, policy)
	if err != nil {
		return packing.Result{}, err
	}
	result.Expanded = expanded
	for i, c := range result.Candidates {
		result.Candidates[i] = original[c.Event.ID]
	}
	return result, nil
}
