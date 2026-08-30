package packing

import (
	"math"
	"strings"
	"unicode"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

type Policy struct {
	AdaptiveEnabled  bool
	DiversityEnabled bool
	MaxPack          int
	MarginThreshold  float64
	EntropyThreshold float64
	DiversityPenalty float64
	PriorityFloor    float64
	PriorityPenalty  float64
}

type Result struct {
	Candidates []model.Candidate
	UsedTokens int
	Expanded   bool
}

func DefaultPolicy() Policy {
	return Policy{MaxPack: 20, MarginThreshold: .03, EntropyThreshold: .65, DiversityPenalty: .08, PriorityFloor: .8, PriorityPenalty: .02}
}

func Select(candidates []model.Candidate, posteriorKeys map[string]string, packK, recallK, tokenBudget int, policy Policy) Result {
	limit := min(packK, len(candidates))
	expanded := false
	if policy.AdaptiveEnabled && shouldExpand(candidates, packK, policy) {
		hardCap := min(min(2*packK, recallK), policy.MaxPack)
		limit = min(hardCap, len(candidates))
		expanded = limit > packK
	}
	ordered := candidates
	if policy.DiversityEnabled {
		ordered = diversify(candidates, posteriorKeys, limit, policy)
	}
	packed := make([]model.Candidate, 0, limit)
	usedTokens := 0
	for _, candidate := range ordered {
		if len(packed) >= limit {
			break
		}
		if usedTokens+candidate.EstimatedTokens > tokenBudget {
			continue
		}
		packed = append(packed, candidate)
		usedTokens += candidate.EstimatedTokens
	}
	return Result{Candidates: packed, UsedTokens: usedTokens, Expanded: expanded}
}

func shouldExpand(candidates []model.Candidate, packK int, policy Policy) bool {
	if packK <= 0 || len(candidates) <= packK {
		return false
	}
	inside, outside := candidates[packK-1], candidates[packK]
	margin := math.Abs(inside.Forecast.CorrectedLaw.Useful - outside.Forecast.CorrectedLaw.Useful)
	entropy := (bernoulliEntropy(inside.Forecast.CorrectedLaw.Useful) + bernoulliEntropy(outside.Forecast.CorrectedLaw.Useful)) / 2
	return margin < policy.MarginThreshold || entropy > policy.EntropyThreshold || outside.Event.Priority >= policy.PriorityFloor
}

func diversify(candidates []model.Candidate, posteriorKeys map[string]string, limit int, policy Policy) []model.Candidate {
	remaining := append([]model.Candidate(nil), candidates...)
	selected := make([]model.Candidate, 0, min(limit, len(candidates)))
	tokenSets := make(map[string]map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		tokenSets[candidate.Event.ID] = tokens(candidate.Event.Content)
	}
	for len(remaining) > 0 && len(selected) < limit {
		bestIndex := 0
		bestScore := math.Inf(-1)
		for index, candidate := range remaining {
			penalty := 0.0
			for _, prior := range selected {
				if distinctCertifiedBuckets(candidate, prior, posteriorKeys) {
					continue
				}
				penalty = math.Max(penalty, tokenJaccardSets(tokenSets[candidate.Event.ID], tokenSets[prior.Event.ID]))
			}
			maxPenalty := policy.DiversityPenalty
			if candidate.Event.Priority >= policy.PriorityFloor {
				maxPenalty = math.Min(maxPenalty, policy.PriorityPenalty)
			}
			score := candidate.Score - maxPenalty*penalty
			if score > bestScore {
				bestIndex, bestScore = index, score
			}
		}
		selected = append(selected, remaining[bestIndex])
		remaining = append(remaining[:bestIndex], remaining[bestIndex+1:]...)
	}
	selected = append(selected, remaining...)
	return selected
}

func distinctCertifiedBuckets(left, right model.Candidate, keys map[string]string) bool {
	leftKey, rightKey := keys[left.Event.ID], keys[right.Event.ID]
	return strings.HasPrefix(leftKey, "ap:") && strings.HasPrefix(rightKey, "ap:") && leftKey != rightKey
}

func bernoulliEntropy(probability float64) float64 {
	probability = math.Min(1-1e-12, math.Max(1e-12, probability))
	return -probability*math.Log(probability) - (1-probability)*math.Log(1-probability)
}

func tokenJaccard(left, right string) float64 {
	return tokenJaccardSets(tokens(left), tokens(right))
}

func tokenJaccardSets(leftTokens, rightTokens map[string]struct{}) float64 {
	if len(leftTokens) == 0 || len(rightTokens) == 0 {
		return 0
	}
	intersection := 0
	for token := range leftTokens {
		if _, ok := rightTokens[token]; ok {
			intersection++
		}
	}
	union := len(leftTokens) + len(rightTokens) - intersection
	return float64(intersection) / float64(union)
}

func tokens(value string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, token := range strings.FieldsFunc(strings.ToLower(value), func(current rune) bool { return !unicode.IsLetter(current) && !unicode.IsDigit(current) }) {
		if len(token) >= 2 {
			result[token] = struct{}{}
		}
	}
	return result
}
