package service

import (
	"context"
	"errors"
	"math"
	"sort"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchsparse"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type ResearchRankCandidate struct {
	Features uint16
	Baseline float64
	Sparse   *researchsparse.Features
}
type ResearchRankInput struct {
	Snapshot   model.Snapshot
	AsOf       time.Time
	Candidates []ResearchRankCandidate
}
type ResearchRankingPolicy struct {
	TenantID string
	// SparseASCII selects the separately validated public-pilot feature map.
	SparseASCII bool
	// DirectOrder reproduces offline probability ordering, instead of the
	// capped-delta experiment. Neither mode changes the properly-scored law.
	DirectOrder bool
	// Score must be read-only and cancellation-aware, using a frozen model.
	// It must not fit, consume feedback, or open a durable journal on retries.
	Score func(context.Context, ResearchRankInput) ([]float64, error)
}

func (s *Service) applyResearchRanking(ctx context.Context, request model.RecallRequest, snapshot model.Snapshot, candidates []model.Candidate) error {
	p := s.config.ResearchRanking
	if p.Score == nil || p.TenantID != request.TenantID {
		return nil
	}
	if len(candidates) > 200 {
		return errors.New("research frontier exceeds200")
	}
	in := ResearchRankInput{Snapshot: snapshot, AsOf: request.AsOf, Candidates: make([]ResearchRankCandidate, len(candidates))}
	for i, c := range candidates {
		base := c.Forecast.PreResidualLaw.Useful
		in.Candidates[i].Baseline = base
		if p.SparseASCII {
			features, e := researchsparse.Extract(request.Query, c.Event)
			if e != nil {
				return e
			}
			in.Candidates[i].Sparse = &features
		} else {
			features, e := researchmemory.Extract(request.Query, c.Event, base)
			if e != nil {
				return e
			}
			in.Candidates[i].Features = features
		}
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	scores, e := p.Score(ctx, in)
	if e != nil {
		return e
	}
	if e = ctx.Err(); e != nil {
		return e
	}
	if s.store.Snapshot(ctx) != snapshot {
		return store.ErrStaleSnapshot
	}
	if len(scores) != len(candidates) {
		return errors.New("incomplete research ranking")
	}
	for _, v := range scores {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return errors.New("invalid research probability")
		}
	}
	// Keep retrieval and existing residual deltas unchanged. This optional
	// search-order correction is separately visible and does not rewrite laws.
	for i := range candidates {
		old := candidates[i].Score
		candidates[i].Score = clamp(old+clamp(scores[i]-candidates[i].Forecast.PreResidualLaw.Useful, -.25, .25), 0, 1)
		if p.DirectOrder {
			candidates[i].Score = scores[i]
		}
		candidates[i].ResearchRankDelta = candidates[i].Score - old
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].Score > candidates[j].Score })
	return nil
}
