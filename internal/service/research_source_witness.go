package service

import (
	"context"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

// ValidateResearchTransferSource checks an opt-in original against the current
// store. It only decides whether this one historical source still matches; it
// does not authorize a new learner publication or authenticate its label.
func (s *Service) ValidateResearchTransferSource(ctx context.Context, target model.Snapshot, r researchmemory.RecordedPrediction, keyID string, key []byte, cutoff time.Time) (bool, error) {
	if s == nil || cutoff.IsZero() || r.Binding == nil || r.Witness == nil {
		return false, errors.New("missing research source witness")
	}
	if err := r.Validate(); err != nil {
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if s.store.Snapshot(ctx) != target {
		return false, errors.New("research target snapshot moved")
	}
	binding := r.Binding
	lineage, ok := s.store.(interface {
		ResearchEventUnchanged(context.Context, model.Snapshot, model.Snapshot, string, string) (bool, error)
	})
	if !ok {
		return false, errors.New("store lacks research event continuity proof")
	}
	continuous, err := lineage.ResearchEventUnchanged(ctx, binding.Snapshot, target, binding.Tenant, binding.EventID)
	if err != nil || !continuous {
		return false, err
	}
	match, err := s.validateResearchTransferSourceEvidence(ctx, target, r, keyID, key, cutoff)
	if err != nil {
		return false, err
	}
	continuous, err = lineage.ResearchEventUnchanged(ctx, binding.Snapshot, target, binding.Tenant, binding.EventID)
	if err != nil || !continuous {
		return false, err
	}
	return match, nil
}

// validateResearchTransferSourceEvidence runs only after the caller has
// established source continuity. A held source/as-of guard may call it without
// reentering the writer permit.
func (s *Service) validateResearchTransferSourceEvidence(ctx context.Context, target model.Snapshot, r researchmemory.RecordedPrediction, keyID string, key []byte, cutoff time.Time) (bool, error) {
	if s == nil || cutoff.IsZero() || r.Binding == nil || r.Witness == nil {
		return false, errors.New("missing research source witness")
	}
	if err := r.Validate(); err != nil {
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if s.store.Snapshot(ctx) != target {
		return false, errors.New("research target snapshot moved")
	}
	binding := r.Binding
	j, err := s.store.GetBayesianJournal(ctx, binding.Tenant, binding.JournalID)
	if err != nil {
		return false, err
	}
	if j.ID != binding.JournalID || j.TenantID != binding.Tenant || j.Snapshot != binding.Snapshot || !j.AsOf.Equal(r.At) {
		return false, errors.New("research source journal mismatch")
	}
	matches := 0
	for _, decision := range j.Report.Decisions {
		if decision.EventID == binding.EventID {
			matches++
			if decision.Forecast.PreResidualLaw.Useful != r.Outer[0] {
				return false, errors.New("research source baseline mismatch")
			}
		}
	}
	if matches != 1 {
		return false, errors.New("research source absent from original journal")
	}
	events, err := s.store.GetEvents(ctx, binding.Tenant, []string{binding.EventID}, cutoff)
	if errors.Is(err, store.ErrEventNotFound) {
		if s.store.Snapshot(ctx) != target {
			return false, errors.New("research target snapshot moved")
		}
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if len(events) != 1 || events[0].ID != binding.EventID {
		return false, errors.New("research source lookup mismatch")
	}
	match, err := r.Witness.Verify(keyID, key, *binding, j.QueryDigest, events[0], r.Prediction.Features, r.Outer[0])
	if err != nil {
		return false, err
	}
	if s.store.Snapshot(ctx) != target {
		return false, errors.New("research target snapshot moved")
	}
	return match, nil
}
