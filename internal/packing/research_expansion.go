package packing

import "github.com/JuanHuaXu/eventframed/internal/model"

// ResearchExpansion exposes only the existing selection-size decision for the
// opt-in priority experiment. It does not perform diversity or evidence work.
// The boundary reads forecast laws and event priority, not numeric rank scores.
func ResearchExpansion(candidates []model.Candidate, packK, recallK int, policy Policy) bool {
	if !policy.AdaptiveEnabled || !shouldExpand(candidates, packK, policy) {
		return false
	}
	limit := min(2*packK, recallK, policy.MaxPack, len(candidates))
	return limit > packK
}
