package service

import (
	"context"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

// WithValidatedResearchBatchAdmission validates one bounded journal frontier
// under its as-of mutation guard before work is invoked. It grants no feedback
// authority and is not a cross-database transaction. work must not mutate/close
// the store. Records and request must remain caller-immutable until return.
func (s *Service) WithValidatedResearchBatchAdmission(ctx context.Context, records []researchmemory.RecordedPrediction, request model.RecallRequest, work func() error) error {
	if len(records) == 0 || len(records) > 256 || records[0].Binding == nil || work == nil {
		return errors.New("invalid bounded research batch")
	}
	guard, ok := s.store.(interface {
		WithResearchAsOfSnapshotWait(context.Context, model.Snapshot, time.Time, func() error) error
	})
	if !ok {
		return errors.New("store lacks as-of research snapshot guard")
	}
	return guard.WithResearchAsOfSnapshotWait(ctx, records[0].Binding.Snapshot, records[0].At, func() error {
		if err := s.ValidateResearchAdmissionBatch(ctx, records, request); err != nil {
			return err
		}
		return work()
	})
}

// ValidateResearchAdmissionBatch amortizes shared journal/query reads without
// caching authority across calls. Each record must belong to the same journal,
// snapshot and prediction time. As with the single-record validator, this alone
// is a point-in-time check, not a reservation for later persistence.
func (s *Service) ValidateResearchAdmissionBatch(ctx context.Context, records []researchmemory.RecordedPrediction, request model.RecallRequest) error {
	return s.validateResearchAdmissionBatch(ctx, records, request, s.store.GetEvents)
}

func (s *Service) validateResearchAdmissionBatch(ctx context.Context, records []researchmemory.RecordedPrediction, request model.RecallRequest, getEvents func(context.Context, string, []string, time.Time) ([]model.Event, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(records) == 0 || len(records) > 256 {
		return errors.New("invalid bounded research batch")
	}
	first := records[0]
	if first.Binding == nil {
		return errors.New("missing research binding")
	}
	binding := *first.Binding
	ids := make([]string, 0, len(records))
	wanted := make(map[string]int, len(records))
	predictions := make(map[uint64]bool, len(records))
	for i, r := range records {
		if err := r.Validate(); err != nil {
			return err
		}
		b := r.Binding
		if b == nil || b.Tenant != request.TenantID || b.Tenant != binding.Tenant || b.JournalID != binding.JournalID || b.Snapshot != binding.Snapshot || !r.At.Equal(first.At) || !r.At.Equal(request.AsOf) {
			return errors.New("research batch binding mismatch")
		}
		if _, exists := wanted[b.EventID]; exists || predictions[r.Prediction.ID] {
			return errors.New("duplicate research batch identity")
		}
		wanted[b.EventID] = i
		predictions[r.Prediction.ID] = true
		ids = append(ids, b.EventID)
	}
	compatible := func() bool {
		if ctx.Err() != nil {
			return false
		}
		if proof, ok := s.store.(interface {
			ResearchPublicationCompatible(context.Context, model.Snapshot, time.Time) bool
		}); ok {
			return proof.ResearchPublicationCompatible(ctx, binding.Snapshot, first.At)
		}
		return s.store.Snapshot(ctx) == binding.Snapshot
	}
	if !compatible() {
		return errors.New("stale research admission")
	}
	j, err := s.store.GetBayesianJournal(ctx, binding.Tenant, binding.JournalID)
	if err != nil {
		return err
	}
	if j.ID != binding.JournalID || j.TenantID != binding.Tenant || j.SessionID != request.SessionID || j.Snapshot != binding.Snapshot || !j.AsOf.Equal(first.At) {
		return errors.New("research journal mismatch")
	}
	query := frame.QueryText(request.Query)
	vector := request.Embedding
	if len(vector) == 0 {
		vector, err = embed.Query(s.embedder, query)
		if err != nil {
			return err
		}
	} else if len(vector) != s.embedder.Dimension() || request.EmbeddingModel != s.embedder.ModelKey() {
		return errors.New("research query model mismatch")
	}
	digest, err := recallQueryDigest(query, vector, s.embedder.ModelKey())
	if err != nil {
		return err
	}
	if digest != j.QueryDigest {
		return errors.New("research query digest mismatch")
	}
	matches := make([]int, len(records))
	for _, d := range j.Report.Decisions {
		if i, ok := wanted[d.EventID]; ok {
			matches[i]++
			if d.Forecast.PreResidualLaw.Useful != records[i].Outer[0] {
				return errors.New("research baseline mismatch")
			}
		}
	}
	for _, count := range matches {
		if count != 1 {
			return errors.New("research candidate absent or ambiguous")
		}
	}
	events, err := getEvents(ctx, binding.Tenant, ids, first.At)
	if err != nil {
		return err
	}
	if len(events) != len(records) {
		return errors.New("research event unavailable")
	}
	seen := make(map[string]bool, len(events))
	for _, event := range events {
		i, ok := wanted[event.ID]
		if !ok || seen[event.ID] {
			return errors.New("research event absent or ambiguous")
		}
		seen[event.ID] = true
		features, err := researchmemory.Extract(request.Query, event, records[i].Outer[0])
		if err != nil {
			return err
		}
		if features != records[i].Prediction.Features {
			return errors.New("research feature mismatch")
		}
	}
	if !compatible() {
		return errors.New("research dependencies changed during validation")
	}
	return nil
}
