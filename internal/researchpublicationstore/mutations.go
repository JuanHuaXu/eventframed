package researchpublicationstore

import (
	"context"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchpublication"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

func (s *Store) Put(ctx context.Context, e model.Event, v []float32, d string) (store.PutResult, error) {
	return mutateWithTouches(s, researchpublication.Ingestion, e.AvailableAt, func() (store.PutResult, error) { return s.EventStore.Put(ctx, e, v, d) }, func(r store.PutResult, after model.Snapshot) {
		if !r.Duplicate {
			s.touchEvent(e.TenantID, e.ID, after)
		}
	})
}
func (s *Store) BindBayesianPolicy(ctx context.Context, d string) (model.Snapshot, error) {
	return mutate(s, researchpublication.General, time.Time{}, func() (model.Snapshot, error) { return s.EventStore.BindBayesianPolicy(ctx, d) })
}
func (s *Store) PutComposition(ctx context.Context, e model.Event, v []float32, d string, b model.Snapshot) (store.PutResult, error) {
	return mutateWithTouches(s, researchpublication.General, time.Time{}, func() (store.PutResult, error) { return s.EventStore.PutComposition(ctx, e, v, d, b) }, func(r store.PutResult, after model.Snapshot) {
		if !r.Duplicate {
			s.touchEvent(e.TenantID, e.ID, after)
		}
	})
}
func (s *Store) Delete(ctx context.Context, t, e string) (store.DeleteResult, error) {
	return mutateWithTouches(s, researchpublication.General, time.Time{}, func() (store.DeleteResult, error) { return s.EventStore.Delete(ctx, t, e) }, func(r store.DeleteResult, after model.Snapshot) {
		if r.Deleted {
			s.touchEvent(t, e, after)
		}
	})
}
func (s *Store) DeleteComposition(ctx context.Context, t, e, r string, at time.Time) (store.CompositionDeleteResult, error) {
	return mutateWithTouches(s, researchpublication.General, time.Time{}, func() (store.CompositionDeleteResult, error) { return s.EventStore.DeleteComposition(ctx, t, e, r, at) }, func(r store.CompositionDeleteResult, after model.Snapshot) {
		if r.Deleted {
			s.touchEvent(t, e, after)
		}
	})
}
func (s *Store) DeleteBefore(ctx context.Context, t string, at time.Time, n int) (store.RetentionResult, error) {
	return mutateWithTouches(s, researchpublication.General, time.Time{}, func() (store.RetentionResult, error) { return s.EventStore.DeleteBefore(ctx, t, at, n) }, func(r store.RetentionResult, after model.Snapshot) {
		for _, id := range r.DeletedIDs {
			s.touchEvent(t, id, after)
		}
	})
}
func (s *Store) Backup(ctx context.Context, d string) error {
	_, e := mutate(s, researchpublication.General, time.Time{}, func() (struct{}, error) { return struct{}{}, s.EventStore.Backup(ctx, d) })
	return e
}
func (s *Store) Compact(ctx context.Context) error {
	_, e := mutate(s, researchpublication.General, time.Time{}, func() (struct{}, error) { return struct{}{}, s.EventStore.Compact(ctx) })
	return e
}
func (s *Store) PublishSelectionCertificate(ctx context.Context, c model.SelectionSupportCertificate) (model.Snapshot, error) {
	return mutate(s, researchpublication.General, time.Time{}, func() (model.Snapshot, error) { return s.EventStore.PublishSelectionCertificate(ctx, c) })
}
func (s *Store) PublishAntiPigeonCertificate(ctx context.Context, c model.AntiPigeonCertificate) (model.Snapshot, error) {
	return mutate(s, researchpublication.General, time.Time{}, func() (model.Snapshot, error) { return s.EventStore.PublishAntiPigeonCertificate(ctx, c) })
}
func (s *Store) PublishOmittedInfluenceCertificate(ctx context.Context, c model.OmittedInfluenceCertificate) (model.Snapshot, error) {
	return mutate(s, researchpublication.General, time.Time{}, func() (model.Snapshot, error) { return s.EventStore.PublishOmittedInfluenceCertificate(ctx, c) })
}
func (s *Store) ApplyBayesianOutcome(ctx context.Context, r model.BayesianOutcomeRequest, p, parent, d string, w float64, c bayes.ChangePolicy, g bayes.GroupPolicy, o model.ResidualObservation, rp residual.Policy) (store.BayesianOutcomeResult, error) {
	return mutate(s, researchpublication.General, time.Time{}, func() (store.BayesianOutcomeResult, error) {
		return s.EventStore.ApplyBayesianOutcome(ctx, r, p, parent, d, w, c, g, o, rp)
	})
}

type graphResult struct {
	graph    model.PredictiveGraph
	snapshot model.Snapshot
}

func (s *Store) PublishPredictiveSnap(ctx context.Context, r model.PredictiveSnapRecord) (model.PredictiveGraph, model.Snapshot, error) {
	v, e := mutate(s, researchpublication.General, time.Time{}, func() (graphResult, error) {
		g, s, e := s.EventStore.PublishPredictiveSnap(ctx, r)
		return graphResult{g, s}, e
	})
	return v.graph, v.snapshot, e
}
func (s *Store) RollbackPredictiveSnap(ctx context.Context, t, id, r string) (model.PredictiveGraph, model.Snapshot, error) {
	v, e := mutate(s, researchpublication.General, time.Time{}, func() (graphResult, error) {
		g, s, e := s.EventStore.RollbackPredictiveSnap(ctx, t, id, r)
		return graphResult{g, s}, e
	})
	return v.graph, v.snapshot, e
}
func (s *Store) PutAgencyProposal(ctx context.Context, r model.AgencyProposalRecord, d string, c, n int, at time.Time) (store.AgencyPutResult, error) {
	return mutate(s, researchpublication.General, time.Time{}, func() (store.AgencyPutResult, error) { return s.EventStore.PutAgencyProposal(ctx, r, d, c, n, at) })
}

type claimResult struct {
	records  []model.AgencyProposalRecord
	snapshot model.Snapshot
}

func (s *Store) ClaimAgencyProposals(ctx context.Context, t, c string, at time.Time, n int, lease time.Duration) ([]model.AgencyProposalRecord, model.Snapshot, error) {
	v, e := mutate(s, researchpublication.General, time.Time{}, func() (claimResult, error) {
		r, s, e := s.EventStore.ClaimAgencyProposals(ctx, t, c, at, n, lease)
		return claimResult{r, s}, e
	})
	return v.records, v.snapshot, e
}
func (s *Store) ResolveAgencyProposal(ctx context.Context, r model.ResolveAgencyProposalRequest, at time.Time) (store.AgencyResolveResult, error) {
	return mutate(s, researchpublication.General, time.Time{}, func() (store.AgencyResolveResult, error) { return s.EventStore.ResolveAgencyProposal(ctx, r, at) })
}
