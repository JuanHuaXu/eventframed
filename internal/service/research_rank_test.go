package service

import (
	"context"
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchsparse"
)

func TestSparseDirectRankingBeforePacking(t *testing.T) {
	s, now := shadowService(t, ResearchShadowPolicy{})
	w := make([]float64, researchsparse.Dimension)
	w[0] = -4
	w[7] = 8
	f, e := researchsparse.New(w)
	if e != nil {
		t.Fatal(e)
	}
	s.config.ResearchRanking = ResearchRankingPolicy{TenantID: "tenant-a", SparseASCII: true, DirectOrder: true, Score: func(ctx context.Context, in ResearchRankInput) ([]float64, error) {
		if len(in.Candidates) != 20 {
			t.Fatal("truncated sparse frontier")
		}
		out := make([]float64, len(in.Candidates))
		for i, c := range in.Candidates {
			if c.Sparse == nil {
				t.Fatal("missing sparse features")
			}
			p, e := f.Score(*c.Sparse)
			if e != nil {
				return nil, e
			}
			out[i] = p
		}
		return out, nil
	}}
	cs := make([]model.Candidate, 20)
	for i := range cs {
		cs[i].Score = .9 - float64(i)*.001
		cs[i].RetrievalScore = cs[i].Score
		cs[i].Event.What.Value = "unrelated text"
		cs[i].Forecast.PreResidualLaw.Useful = .8
	}
	cs[19].Event.What.Value = "unique query"
	wantRetrieval := cs[19].RetrievalScore
	wantLaw := cs[19].Forecast
	if e = s.applyResearchRanking(context.Background(), model.RecallRequest{TenantID: "tenant-a", Query: "unique query", AsOf: now}, s.store.Snapshot(context.Background()), cs); e != nil {
		t.Fatal(e)
	}
	if cs[0].Event.What.Value != "unique query" || cs[0].Score < .98 {
		t.Fatal("direct mode did not preserve learned order")
	}
	if cs[0].RetrievalScore != wantRetrieval || cs[0].Forecast != wantLaw {
		t.Fatal("direct mode changed backend or law")
	}
}

func TestResearchRankRecallFrozenBaseline(t *testing.T) {
	base, now := shadowService(t, ResearchShadowPolicy{})
	s, _ := shadowService(t, ResearchShadowPolicy{})
	a := researchmemory.New(1, 42)
	f := a.Freeze()
	seen := 0
	s.config.ResearchRanking = ResearchRankingPolicy{TenantID: "tenant-a", Score: func(ctx context.Context, in ResearchRankInput) ([]float64, error) {
		seen = len(in.Candidates)
		p := make([]float64, seen)
		for i, c := range in.Candidates {
			if e := ctx.Err(); e != nil {
				return nil, e
			}
			v, e := f.Score(c.Features, c.Baseline, 1, in.AsOf)
			if e != nil {
				return nil, e
			}
			p[i] = v
		}
		return p, nil
	}}
	want, got := shadowRecall(t, base, now), shadowRecall(t, s, now)
	if seen != 3 {
		t.Fatal("callback did not see pre-packing frontier", seen)
	}
	if !reflect.DeepEqual(want.Candidates, got.Candidates) {
		t.Fatal("untrained adapter changed candidates")
	}
	if got.PacketAnswerCertainty != 0 || got.PacketConfidence != 0 {
		t.Fatal("experimental packing inherited calibration")
	}
	labels, pending := a.Counts()
	if labels != 0 || pending != 0 {
		t.Fatal("retrieval fitted or journaled evidence")
	}
}

func TestResearchRankFrontierAndIsolation(t *testing.T) {
	s, now := shadowService(t, ResearchShadowPolicy{})
	cs := make([]model.Candidate, 20)
	for i := range cs {
		cs[i].Score = .6 - float64(i)*.001
		cs[i].RetrievalScore = cs[i].Score
		cs[i].Forecast.PreResidualLaw.Useful = .5
	}
	before := append([]model.Candidate(nil), cs...)
	s.config.ResearchRanking = ResearchRankingPolicy{TenantID: "tenant-a", Score: func(_ context.Context, in ResearchRankInput) ([]float64, error) {
		if len(in.Candidates) != 20 {
			t.Fatal("frontier truncated")
		}
		p := make([]float64, 20)
		for i := range p {
			p[i] = .5
			in.Candidates[i].Baseline = 0
		}
		p[19] = .9
		return p, nil
	}}
	r := model.RecallRequest{TenantID: "tenant-a", Query: "public query", AsOf: now}
	if e := s.applyResearchRanking(context.Background(), r, s.store.Snapshot(context.Background()), cs); e != nil {
		t.Fatal(e)
	}
	if cs[0].RetrievalScore != before[19].RetrievalScore || math.Abs(cs[0].ResearchRankDelta-.25) > 1e-12 {
		t.Fatal("tail not promoted correctly")
	}
	for i, c := range cs {
		original := before[max(0, i-1)]
		if i == 0 {
			original = before[19]
		}
		if c.Forecast != original.Forecast || c.RankDelta != original.RankDelta {
			t.Fatal("rank-only hook changed law or base delta")
		}
		if i > 0 && c.ResearchRankDelta != 0 {
			t.Fatal("callback mutated correction baseline")
		}
	}
}

func TestResearchRankRejectsAtomically(t *testing.T) {
	s, now := shadowService(t, ResearchShadowPolicy{})
	for _, scores := range [][]float64{{.8}, {.8, math.NaN()}, {.8, 1.1}} {
		cs := make([]model.Candidate, 2)
		before := append([]model.Candidate(nil), cs...)
		s.config.ResearchRanking = ResearchRankingPolicy{TenantID: "tenant-a", Score: func(context.Context, ResearchRankInput) ([]float64, error) { return scores, nil }}
		e := s.applyResearchRanking(context.Background(), model.RecallRequest{TenantID: "tenant-a", Query: "public query", AsOf: now}, s.store.Snapshot(context.Background()), cs)
		if e == nil || !reflect.DeepEqual(cs, before) {
			t.Fatal("invalid results partially applied")
		}
	}
	called := false
	s.config.ResearchRanking.Score = func(context.Context, ResearchRankInput) ([]float64, error) { called = true; return nil, nil }
	if e := s.applyResearchRanking(context.Background(), model.RecallRequest{TenantID: "other"}, model.Snapshot{}, nil); e != nil || called {
		t.Fatal("cross-tenant hook ran")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e := s.applyResearchRanking(ctx, model.RecallRequest{TenantID: "tenant-a"}, model.Snapshot{}, nil); e == nil || called {
		t.Fatal("cancelled hook ran")
	}
}
