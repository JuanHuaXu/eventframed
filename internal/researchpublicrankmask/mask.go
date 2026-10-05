// Package researchpublicrankmask isolates a source-only feature ablation.
// Nomination cues may still build a frontier, but this learner cannot score them.
package researchpublicrankmask

import (
	"context"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
)

// SourceOnly validates every original feature before masking, returns an owned
// complete copy, and never changes base scores, source identities or order.
func SourceOnly(ctx context.Context, rows []r.Row) ([]r.Row, error) {
	if _, e := (r.Model{}).Rank(ctx, rows); e != nil {
		return nil, e
	}
	out := append([]r.Row(nil), rows...)
	for i := range out {
		out[i].Features[6] = 0
		out[i].Features[7] = 0
	}
	return out, ctx.Err()
}
