// Package researchfusion contains bounded, label-free research rank fusion.
// Scores are search-order utilities, never calibrated outcome probabilities.
package researchfusion

import (
	"context"
	"errors"
	"math"
	"sort"
)

const MaxFrontier = 200
const Offset = 60

// Scores treats input order as the incumbent permutation and constructs the
// second permutation from lexical coverage, with incumbent-order tie breaking.
// A positive protectedPrefix preserves only that pre-packing prefix's SET.
// Correlation/token-budget packing can still change the eventual packet set.
func Scores(ctx context.Context, lexical []float64, protectedPrefix int) ([]float64, error) {
	if len(lexical) > MaxFrontier || protectedPrefix < 0 || protectedPrefix > MaxFrontier {
		return nil, errors.New("fusion cap violation")
	}
	order := make([]int, len(lexical))
	for i, v := range lexical {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return nil, errors.New("invalid lexical coverage")
		}
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool { return lexical[order[i]] > lexical[order[j]] })
	out := make([]float64, len(lexical))
	for j, i := range order {
		v := float64(Offset+1) / 2 * (1/float64(Offset+i+1) + 1/float64(Offset+j+1))
		if protectedPrefix > 0 {
			v /= 2
			if i < protectedPrefix {
				v += .5
			}
		}
		out[i] = v
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
