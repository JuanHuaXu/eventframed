// Package researchmagnitude is an isolated rank-only magnitude fusion pilot.
// Its utilities are not forecast probabilities or answer confidence.
package researchmagnitude

import (
	"context"
	"errors"
	"math"
)

// Scores preserves lexical score gaps, combining coverage with incumbent rank
// utility. Input order, not a usefulness probability, defines the incumbent.
// Stable descending sorting belongs to the existing research service hook.
func Scores(ctx context.Context, lexical []float64, weight float64) ([]float64, error) {
	if len(lexical) > 200 || math.IsNaN(weight) || math.IsInf(weight, 0) || weight < 0 || weight > 1 {
		return nil, errors.New("magnitude fusion contract violation")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := make([]float64, len(lexical))
	for i, v := range lexical {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return nil, errors.New("invalid lexical coverage")
		}
		out[i] = (1-weight)/float64(i+1) + weight*v
	}
	return out, nil
}
