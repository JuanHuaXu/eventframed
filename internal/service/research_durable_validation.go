package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

// WithValidatedResearchAdmission keeps owned store mutations out of a bounded
// validation-plus-ledger operation. Unsupported stores fail closed. The callback
// must not mutate this store; later feedback still requires a fresh guard.
func (s *Service) WithValidatedResearchAdmission(ctx context.Context, r researchmemory.RecordedPrediction, request model.RecallRequest, work func() error) error {
	if r.Binding == nil || work == nil {
		return errors.New("missing bound research operation")
	}
	guard, ok := s.store.(interface {
		WithResearchSnapshot(context.Context, model.Snapshot, func() error) error
	})
	if !ok {
		return errors.New("store lacks research snapshot guard")
	}
	return guard.WithResearchSnapshot(ctx, r.Binding.Snapshot, func() error {
		if e := s.ValidateResearchAdmission(ctx, r, request); e != nil {
			return e
		}
		return work()
	})
}

// WithValidatedResearchAsOfAdmission binds temporal guard authority to the
// original prediction time, then rechecks the request/journal/event under lock.
// It remains research-only; later feedback needs its own fresh authority check.
func (s *Service) WithValidatedResearchAsOfAdmission(ctx context.Context, r researchmemory.RecordedPrediction, request model.RecallRequest, work func() error) error {
	if r.Binding == nil || work == nil {
		return errors.New("missing bound research operation")
	}
	guard, ok := s.store.(interface {
		WithResearchAsOfSnapshotWait(context.Context, model.Snapshot, time.Time, func() error) error
	})
	if !ok {
		return errors.New("store lacks as-of research snapshot guard")
	}
	return guard.WithResearchAsOfSnapshotWait(ctx, r.Binding.Snapshot, r.At, func() error {
		if err := s.ValidateResearchAdmission(ctx, r, request); err != nil {
			return err
		}
		return work()
	})
}

// WithValidatedResearchCandidateAsOfAdmission checks the committed source
// before a forecast exists. Its callback journals the actual forecast.
func (s *Service) WithValidatedResearchCandidateAsOfAdmission(ctx context.Context, binding researchmemory.ServiceBinding, at time.Time, features uint16, baseline float64, request model.RecallRequest, work func() error) error {
	if work == nil {
		return errors.New("missing research candidate operation")
	}
	guard, ok := s.store.(interface {
		WithResearchAsOfSnapshotWait(context.Context, model.Snapshot, time.Time, func() error) error
	})
	if !ok {
		return errors.New("store lacks as-of research snapshot guard")
	}
	return guard.WithResearchAsOfSnapshotWait(ctx, binding.Snapshot, at, func() error {
		if err := s.validateResearchCandidate(ctx, binding, at, features, baseline, request); err != nil {
			return err
		}
		return work()
	})
}

// ValidateResearchAdmission checks actual journal/query/event inputs at a point
// in time. It neither authenticates labels nor reserves the snapshot against a
// later write. A durable bridge must revalidate around its state transitions.
func (s *Service) ValidateResearchAdmission(ctx context.Context, r researchmemory.RecordedPrediction, request model.RecallRequest) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	if e := r.Validate(); e != nil {
		return e
	}
	if r.Binding == nil {
		return errors.New("research request binding mismatch")
	}
	return s.validateResearchCandidate(ctx, *r.Binding, r.At, r.Prediction.Features, r.Outer[0], request)
}

func (s *Service) validateResearchCandidate(ctx context.Context, binding researchmemory.ServiceBinding, at time.Time, features uint16, baseline float64, request model.RecallRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if binding.Tenant == "" || binding.JournalID == "" || binding.EventID == "" || len(binding.Tenant) > 4096 || len(binding.JournalID) > 4096 || len(binding.EventID) > 4096 || binding.Snapshot.RuntimeVersion == 0 || binding.Snapshot.ContractVersion == 0 || features >= 512 || math.IsNaN(baseline) || math.IsInf(baseline, 0) || baseline < 0 || baseline > 1 || binding.Tenant != request.TenantID || at.IsZero() || !at.Equal(request.AsOf) {
		return errors.New("research request binding mismatch")
	}
	compatible := func() bool {
		if ctx.Err() != nil {
			return false
		}
		if proof, ok := s.store.(interface {
			ResearchPublicationCompatible(context.Context, model.Snapshot, time.Time) bool
		}); ok {
			return proof.ResearchPublicationCompatible(ctx, binding.Snapshot, at)
		}
		return s.store.Snapshot(ctx) == binding.Snapshot
	}
	if !compatible() {
		return errors.New("stale research admission")
	}
	j, e := s.store.GetBayesianJournal(ctx, binding.Tenant, binding.JournalID)
	if e != nil {
		return e
	}
	if j.ID != binding.JournalID || j.TenantID != binding.Tenant || j.SessionID != request.SessionID || j.Snapshot != binding.Snapshot || !j.AsOf.Equal(at) {
		return errors.New("research journal mismatch")
	}
	vector := request.Embedding
	if len(vector) == 0 {
		vector, e = embed.Query(s.embedder, frame.QueryText(request.Query))
		if e != nil {
			return e
		}
	} else if len(vector) != s.embedder.Dimension() || request.EmbeddingModel != s.embedder.ModelKey() {
		return errors.New("research query model mismatch")
	}
	digest, e := recallQueryDigest(frame.QueryText(request.Query), vector, s.embedder.ModelKey())
	if e != nil {
		return e
	}
	if digest != j.QueryDigest {
		return errors.New("research query digest mismatch")
	}
	matches := 0
	for _, d := range j.Report.Decisions {
		if d.EventID == binding.EventID {
			matches++
			if d.Forecast.PreResidualLaw.Useful != baseline {
				return errors.New("research baseline mismatch")
			}
		}
	}
	if matches != 1 {
		return errors.New("research candidate absent or ambiguous")
	}
	events, e := s.store.GetEvents(ctx, binding.Tenant, []string{binding.EventID}, at)
	if e != nil {
		return e
	}
	if len(events) != 1 || events[0].ID != binding.EventID {
		return errors.New("research event unavailable")
	}
	extracted, e := researchmemory.Extract(request.Query, events[0], baseline)
	if e != nil {
		return e
	}
	if extracted != features {
		return errors.New("research feature mismatch")
	}
	if !compatible() {
		return errors.New("research dependencies changed during validation")
	}
	return nil
}
