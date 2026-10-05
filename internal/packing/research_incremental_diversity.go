package packing

import (
	"math"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// researchDiversifyIncremental preserves greedy selection and stable ties.
// After each selection only one new pair can increase a survivor's maximum
// penalty. This removes repeated comparisons with earlier selected items.
// It is not wired into Select; frozen production and experiment paths stay put.
func researchDiversifyIncremental(candidates []model.Candidate, posteriorKeys map[string]string, limit int, policy Policy) []model.Candidate {
	remaining := append([]model.Candidate(nil), candidates...)
	penalties := make([]float64, len(remaining))
	selected := make([]model.Candidate, 0, min(limit, len(candidates)))
	tokenSets := make(map[string]map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		tokenSets[candidate.Event.ID] = tokens(candidate.Event.FrameText())
	}
	for len(remaining) > 0 && len(selected) < limit {
		bestIndex := 0
		bestScore := math.Inf(-1)
		for index, candidate := range remaining {
			maxPenalty := policy.DiversityPenalty
			if candidate.Event.Priority >= policy.PriorityFloor {
				maxPenalty = math.Min(maxPenalty, policy.PriorityPenalty)
			}
			score := candidate.Score - maxPenalty*penalties[index]
			if score > bestScore {
				bestIndex, bestScore = index, score
			}
		}
		prior := remaining[bestIndex]
		selected = append(selected, prior)
		remaining = append(remaining[:bestIndex], remaining[bestIndex+1:]...)
		penalties = append(penalties[:bestIndex], penalties[bestIndex+1:]...)
		if len(selected) >= limit {
			break
		}
		for index, candidate := range remaining {
			if distinctCertifiedBuckets(candidate, prior, posteriorKeys) {
				continue
			}
			penalties[index] = math.Max(penalties[index], tokenJaccardSets(tokenSets[candidate.Event.ID], tokenSets[prior.Event.ID]))
		}
	}
	return append(selected, remaining...)
}
