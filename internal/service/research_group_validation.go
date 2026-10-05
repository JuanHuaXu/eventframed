package service

import (
	"context"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

// WithValidatedResearchGroupAdmission shares event reads within one held guard,
// not service authority across calls. Each journal/query/feature check remains.
// Inputs must remain immutable; work must not mutate/close the guarded store.
func (s *Service) WithValidatedResearchGroupAdmission(ctx context.Context, groups [][]researchmemory.RecordedPrediction, requests []model.RecallRequest, work func() error) error {
	if len(groups) == 0 || len(groups) > 4 || len(requests) != len(groups) || len(groups[0]) == 0 || groups[0][0].Binding == nil || work == nil {
		return errors.New("invalid research group")
	}
	guard, ok := s.store.(interface {
		WithResearchAsOfSnapshotWait(context.Context, model.Snapshot, time.Time, func() error) error
	})
	if !ok {
		return errors.New("store lacks as-of research snapshot guard")
	}
	first := groups[0][0]
	return guard.WithResearchAsOfSnapshotWait(ctx, first.Binding.Snapshot, first.At, func() error {
		if err := s.validateResearchAdmissionGroups(ctx, groups, requests); err != nil {
			return err
		}
		return work()
	})
}

// Caller holds the service mutation guard. The read union never escapes this
// invocation, and it does not cache query-specific features or journal checks.
func (s *Service) validateResearchAdmissionGroups(ctx context.Context, groups [][]researchmemory.RecordedPrediction, requests []model.RecallRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(groups) == 0 || len(groups) > 4 || len(groups) != len(requests) {
		return errors.New("invalid bounded research groups")
	}
	tenant, at := requests[0].TenantID, requests[0].AsOf
	if tenant == "" || at.IsZero() {
		return errors.New("invalid group scope")
	}
	var ids []string
	wanted := map[string]bool{}
	predictions := map[uint64]bool{}
	total := 0
	for i, group := range groups {
		if len(group) == 0 || len(group) > 256 || requests[i].TenantID != tenant || !requests[i].AsOf.Equal(at) {
			return errors.New("research group scope mismatch")
		}
		total += len(group)
		if total > 256 {
			return errors.New("research group record cap")
		}
		for _, r := range group {
			if err := r.Validate(); err != nil {
				return err
			}
			if r.Binding == nil || r.Binding.Tenant != tenant || !r.At.Equal(at) || predictions[r.Prediction.ID] {
				return errors.New("research group identity mismatch")
			}
			predictions[r.Prediction.ID] = true
			if !wanted[r.Binding.EventID] {
				wanted[r.Binding.EventID] = true
				ids = append(ids, r.Binding.EventID)
			}
		}
	}
	events, err := s.store.GetEvents(ctx, tenant, ids, at)
	if err != nil {
		return err
	}
	if len(events) != len(ids) {
		return errors.New("research group event unavailable")
	}
	byID := make(map[string]model.Event, len(events))
	for _, event := range events {
		if !wanted[event.ID] {
			return errors.New("unexpected group event")
		}
		if _, exists := byID[event.ID]; exists {
			return errors.New("ambiguous group event")
		}
		byID[event.ID] = event
	}
	get := func(ctx context.Context, t string, keys []string, when time.Time) ([]model.Event, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if t != tenant || !when.Equal(at) {
			return nil, errors.New("group read scope mismatch")
		}
		out := make([]model.Event, 0, len(keys))
		for _, key := range keys {
			event, ok := byID[key]
			if !ok {
				return nil, errors.New("group event missing")
			}
			out = append(out, event)
		}
		return out, nil
	}
	for i, group := range groups {
		if err := s.validateResearchAdmissionBatch(ctx, group, requests[i], get); err != nil {
			return err
		}
	}
	return nil
}
