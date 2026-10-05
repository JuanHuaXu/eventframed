// Package researchpublicfrontierunion separates expanded nomination from scoring.
// It is isolated research: no law, belief, store or production boundary changes.
package researchpublicfrontierunion

import (
	"context"
	"errors"
	fb "github.com/JuanHuaXu/eventframed/internal/researchpublicfeedback"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
	"sort"
)

const MaxUnion = 400

func OriginalScores(ctx context.Context, x *fb.Index, q string, a, b []fb.Hit) ([]fb.Hit, error) {
	if ctx == nil || x == nil || len(a) > 200 || len(b) > 200 {
		return nil, errors.New("invalid union context")
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	seen := map[string]bool{}
	hits := []fb.Hit{}
	for _, h := range a {
		if seen[h.ID] {
			return nil, errors.New("duplicate original nominee")
		}
		seen[h.ID] = true
		hits = append(hits, h)
	}
	seenB := map[string]bool{}
	for _, h := range b {
		if seenB[h.ID] {
			return nil, errors.New("duplicate expanded nominee")
		}
		seenB[h.ID] = true
		if !seen[h.ID] {
			seen[h.ID] = true
			hits = append(hits, h)
		}
	}
	// Validate both entire legs, even duplicates whose expanded scores are unused.
	if _, e := x.Rerank(ctx, q, a); e != nil {
		return nil, e
	}
	if _, e := x.Rerank(ctx, q, b); e != nil {
		return nil, e
	}
	out := make([]fb.Hit, len(hits))
	score := map[string]float64{}
	for first := 0; first < len(hits); first += 200 {
		end := first + 200
		if end > len(hits) {
			end = len(hits)
		}
		ranked, e := x.Rerank(ctx, q, hits[first:end])
		if e != nil {
			return nil, e
		}
		for _, h := range ranked {
			score[h.ID] = h.Score
		}
	}
	for j, h := range hits {
		out[j] = fb.Hit{ID: h.ID, Score: score[h.ID]}
	}
	return out, ctx.Err()
}

// Score the full union before the declared cap. Existing scalar Model.Score
// enforces unchanged feature/base/weight semantics for each candidate.
func Rank(ctx context.Context, m r.Model, rows []r.Row, cap int) ([]r.Scored, error) {
	if ctx == nil || len(rows) > MaxUnion || cap < 1 || cap > 200 {
		return nil, errors.New("invalid union rank context")
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	seen := map[string]bool{}
	out := make([]r.Scored, len(rows))
	for j, row := range rows {
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		if seen[row.ID] {
			return nil, errors.New("duplicate union row")
		}
		seen[row.ID] = true
		score, e := m.Score(row)
		if e != nil {
			return nil, e
		}
		out[j] = r.Scored{ID: row.ID, Score: score}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].ID < out[j].ID
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > cap {
		out = out[:cap]
	}
	return out, ctx.Err()
}
