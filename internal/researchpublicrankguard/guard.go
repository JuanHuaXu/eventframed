// Package researchpublicrankguard rejects incomplete native rank advice without
// dropping the already bound search frontier. It is research-only serving glue.
package researchpublicrankguard

import (
	"context"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchpublicpool"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

type Result struct {
	Candidates     []retrieval.Candidate
	NativeComplete bool
}

// CompleteOrSearch is not a repair of native scoring. Incomplete advice earns
// no rank benefit: retain the entire original search order unchanged. Invalid
// or out-of-frontier advice remains an error, not a permissive fallback.
func CompleteOrSearch(ctx context.Context, r *researchpublicpool.Registry, nominated, advice []retrieval.Candidate, asOf time.Time, cap int) (Result, error) {
	if r == nil {
		return Result{}, errors.New("nil source registry")
	}
	if _, err := r.Bind(ctx, nominated, asOf, cap); err != nil {
		return Result{}, err
	}
	if _, err := r.Bind(ctx, advice, asOf, cap); err != nil {
		return Result{}, err
	}
	set := make(map[string]bool, len(nominated))
	for _, c := range nominated {
		set[c.ID] = true
	}
	for _, c := range advice {
		if !set[c.ID] {
			return Result{}, errors.New("native advice outside nomination")
		}
	}
	complete := len(advice) == len(nominated)
	chosen := nominated
	if complete {
		chosen = advice
	}
	// The original full-frontier invariant is enforced on the emitted permutation.
	if _, err := r.BindRanked(ctx, nominated, chosen, asOf, cap); err != nil {
		return Result{}, err
	}
	owned := make([]retrieval.Candidate, len(chosen))
	for i, c := range chosen {
		owned[i] = c
		owned[i].Metadata = append([]byte(nil), c.Metadata...)
	}
	return Result{Candidates: owned, NativeComplete: complete}, nil
}
