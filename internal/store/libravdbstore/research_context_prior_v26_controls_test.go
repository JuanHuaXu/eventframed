package libravdbstore

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type priorFixtureStoreV26 struct {
	store.EventStore
	journal   model.BayesianJournalEntry
	posterior model.BayesianPosterior
	snapshot  model.Snapshot
}

func (s *priorFixtureStoreV26) GetBayesianJournal(context.Context, string, string) (model.BayesianJournalEntry, error) {
	return s.journal, nil
}

func (s *priorFixtureStoreV26) GetBayesianPosterior(context.Context, string, string) (model.BayesianPosterior, error) {
	return s.posterior, nil
}

func (s *priorFixtureStoreV26) Snapshot(context.Context) model.Snapshot { return s.snapshot }

func priorFixtureV26(t *testing.T) (*contextPriorStoreV26, *priorFixtureStoreV26, context.Context) {
	t.Helper()
	origin := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	req := model.RecallRequest{TenantID: "tenant-a", Query: "public fixture", Embedding: []float32{1, 0}, EmbeddingModel: "public-model", AsOf: origin.Add(5 * time.Second)}
	ctx, err := priorContextV26(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	scope := ctx.Value(priorScopeKeyV26{}).(priorScopeV26)
	fake := &priorFixtureStoreV26{
		snapshot:  model.Snapshot{PolicyVersion: 11, EvidenceEpoch: 217},
		posterior: model.BayesianPosterior{TenantID: "tenant-a", PosteriorKey: "event-a", Alpha: 2, Beta: 1, EffectiveSupport: 1, EvidenceEpoch: 217, Certified: true, UpdatedAt: origin.Add(3 * time.Second)},
	}
	fake.journal = model.BayesianJournalEntry{ID: "origin-a", TenantID: "tenant-a", AsOf: origin.Add(2 * time.Second), QueryDigest: scope.digest, Snapshot: fake.snapshot,
		Report: model.BayesianShadowReport{JournalID: "origin-a", JournalDurable: true, SelectionSupportCertified: true, OmittedInfluenceCertified: true,
			Decisions: []model.BayesianDecision{{EventID: "event-a", PosteriorKey: "event-a", Activated: true,
				Forecast: model.ForecastBundle{HorizonKey: model.RetrievalUsefulnessHorizon, BaseLaw: model.BernoulliLaw{Useful: .9, NotUseful: .1}}}}},
	}
	adapter := &contextPriorStoreV26{EventStore: fake}
	if err := adapter.bind(context.Background(), "tenant-a", "origin-a"); err != nil {
		t.Fatal(err)
	}
	return adapter, fake, ctx
}

func TestResearchContextPriorV26IdentityAndCounts(t *testing.T) {
	adapter, fake, ctx := priorFixtureV26(t)
	got, err := adapter.GetBayesianPosterior(ctx, "tenant-a", "event-a")
	if err != nil || math.Abs(got.Mean()-2.8/3) > 1e-14 || fake.posterior.Alpha != 2 {
		t.Fatal("prior did not preserve owned sufficient counts", got, err)
	}
	fake.posterior.Alpha, fake.posterior.Beta, fake.posterior.EffectiveSupport = 1, 1, 0
	got, err = adapter.GetBayesianPosterior(ctx, "tenant-a", "event-a")
	if err != nil || math.Abs(got.Mean()-.9) > 1e-14 {
		t.Fatal("zero-evidence identity", got, err)
	}
}

func TestResearchContextPriorV26Validity(t *testing.T) {
	for _, item := range []struct {
		name   string
		mutate func(*priorFixtureStoreV26)
	}{
		{"epoch", func(s *priorFixtureStoreV26) { s.snapshot.EvidenceEpoch++ }},
		{"policy", func(s *priorFixtureStoreV26) { s.snapshot.PolicyVersion++ }},
		{"stale-posterior", func(s *priorFixtureStoreV26) { s.posterior.EvidenceEpoch-- }},
		{"future-posterior", func(s *priorFixtureStoreV26) { s.posterior.UpdatedAt = s.posterior.UpdatedAt.Add(time.Hour) }},
		{"pre-anchor", func(s *priorFixtureStoreV26) { s.posterior.UpdatedAt = s.posterior.UpdatedAt.Add(-time.Minute) }},
		{"uncertified", func(s *priorFixtureStoreV26) { s.posterior.Certified = false }},
		{"fractional", func(s *priorFixtureStoreV26) { s.posterior.Alpha = 1.5; s.posterior.EffectiveSupport = .5 }},
		{"support-mismatch", func(s *priorFixtureStoreV26) { s.posterior.EffectiveSupport = 2 }},
		{"nan", func(s *priorFixtureStoreV26) { s.posterior.Alpha = math.NaN() }},
		{"infinity", func(s *priorFixtureStoreV26) { s.posterior.Beta = math.Inf(1) }},
		{"negative", func(s *priorFixtureStoreV26) { s.posterior.Alpha = .5 }},
		{"wrong-key", func(s *priorFixtureStoreV26) { s.posterior.PosteriorKey = "other-key" }},
		{"wrong-tenant", func(s *priorFixtureStoreV26) { s.posterior.TenantID = "other-tenant" }},
		{"working", func(s *priorFixtureStoreV26) { s.posterior.WorkingBelief = &model.WorkingBelief{} }},
		{"cap", func(s *priorFixtureStoreV26) { s.posterior.Alpha = 1000002; s.posterior.EffectiveSupport = 1000001 }},
	} {
		t.Run(item.name, func(t *testing.T) {
			adapter, fake, ctx := priorFixtureV26(t)
			item.mutate(fake)
			if _, err := adapter.GetBayesianPosterior(ctx, "tenant-a", "event-a"); !errors.Is(err, store.ErrPosteriorNotFound) {
				t.Fatal("invalid authority escaped", err)
			}
		})
	}
}
