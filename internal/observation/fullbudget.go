package observation

import (
	"errors"
	"math/bits"
	"math/rand"
)

// RunFullBudget is a research-only counterfactual, not a serving policy.
func RunFullBudget(m *Model, reader Reader, epoch uint64) (Result, error) {
	const policy = "mmm"
	const seed int64 = 0
	valid := false
	for _, p := range Policies {
		if p == policy {
			valid = true
		}
	}
	if m == nil || reader == nil || !valid {
		return Result{}, errors.New("invalid model, reader, or policy")
	}
	budget := 6
	if policy == "exhaustive" {
		budget = 9
	}
	rng := rand.New(rand.NewSource(seed))
	var mask, values, attempted uint16
	result := Result{Probability: m.predict(0, 0)}
	for result.Cost < budget {
		if reader.Epoch() != epoch {
			return Result{}, errors.New("stale observation snapshot")
		}
		var candidates []View
		for scope := 0; scope < 3; scope++ {
			for depth := 0; depth < 3; depth++ {
				v := View{scope, depth}
				cost := bits.OnesCount16(v.Mask() &^ attempted)
				if cost == 0 || cost > budget-result.Cost {
					continue
				}
				if result.Cost == 0 && v != (View{0, 0}) {
					continue
				}
				if policy == "fixed" && v != (View{0, 0}) {
					continue
				}
				if policy == "scope" && depth != 0 {
					continue
				}
				if policy == "depth" && scope != 0 {
					continue
				}
				candidates = append(candidates, v)
			}
		}
		if len(candidates) == 0 {
			result.Stop = "no_affordable_view"
			break
		}
		chosen := candidates[0]
		if policy == "random" {
			chosen = candidates[rng.Intn(len(candidates))]
		}
		if policy == "breadth" || policy == "scope" {
			for _, v := range candidates {
				if v.Depth < chosen.Depth || (v.Depth == chosen.Depth && v.Scope < chosen.Scope) {
					chosen = v
				}
			}
		}
		if policy == "mmm" {
			best := -1.0
			for _, v := range candidates {
				added := v.Mask() &^ attempted
				score := m.value(mask, values, added) / float64(bits.OnesCount16(added))
				if score > best+1e-12 {
					best = score
					chosen = v
				}
			}
		}
		newMask, newValues, err := reader.Read(chosen)
		if err != nil {
			return Result{}, err
		}
		if reader.Epoch() != epoch {
			return Result{}, errors.New("snapshot changed during observation")
		}
		if newMask&^chosen.Mask() != 0 || newValues&^newMask != 0 {
			return Result{}, errors.New("reader returned undeclared coordinates")
		}
		if (newValues^values)&(newMask&mask) != 0 {
			return Result{}, errors.New("snapshot changed observed values")
		}
		result.Cost += bits.OnesCount16(chosen.Mask() &^ attempted)
		attempted |= chosen.Mask()
		mask |= newMask
		values |= newValues
		result.Probability = m.predict(mask, values)
		result.Trace = append(result.Trace, Step{chosen, attempted, mask, values, result.Probability})
	}
	if result.Stop == "" {
		result.Stop = "budget"
	}
	result.Observed = bits.OnesCount16(mask)
	return result, nil
}
