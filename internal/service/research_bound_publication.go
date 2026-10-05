package service

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

// ResearchBoundSlot is an opt-in, research-only immutable model pointer.
// Readers must use Score, which checks the owning store's as-of authority.
type ResearchBoundSlot struct {
	current atomic.Pointer[ResearchBoundPublication]
}

type ResearchBoundPublication struct {
	Target        model.Snapshot
	Epoch         uint64
	Cutoff        time.Time
	Retained      int
	TerminalCount int
	model         researchmemory.Frozen
}

type ResearchBoundPlan struct {
	Target model.Snapshot
	Tenant string
	Epoch  uint64
	Seed   int64
	Cutoff time.Time
	KeyID  string
	Key    []byte
}

type PreparedResearchBound struct {
	service   *Service
	slot      *ResearchBoundSlot
	durable   *researchmemory.Durable
	base      *ResearchBoundPublication
	candidate *ResearchBoundPublication
	sequence  int64
}

func (slot *ResearchBoundSlot) Current() (ResearchBoundPublication, bool) {
	if slot == nil {
		return ResearchBoundPublication{}, false
	}
	p := slot.current.Load()
	if p == nil {
		return ResearchBoundPublication{}, false
	}
	return *p, true
}

// PrepareResearchBound reads every cutoff-eligible terminal from the durable
// stream and validates sources against one exact target. It does not publish.
func (s *Service) PrepareResearchBound(ctx context.Context, slot *ResearchBoundSlot, durable *researchmemory.Durable, plan ResearchBoundPlan) (*PreparedResearchBound, error) {
	if s == nil || slot == nil || durable == nil || ctx == nil || plan.Tenant == "" || plan.Tenant != durable.Tenant() || plan.KeyID == "" || len(plan.Key) < 32 {
		return nil, errors.New("invalid research bound preparation")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.store.Snapshot(ctx) != plan.Target {
		return nil, errors.New("research bound target moved before preparation")
	}
	base := slot.current.Load()
	if base != nil && plan.Epoch <= base.Epoch {
		return nil, errors.New("research bound epoch must advance")
	}
	labels, sequence, err := durable.BoundLabelsWithSequence(ctx, plan.Cutoff)
	if err != nil {
		return nil, err
	}
	validate := func(ctx context.Context, target model.Snapshot, label researchmemory.BoundLabel) (bool, error) {
		return s.ValidateResearchTransferSource(ctx, target, label.Prediction, plan.KeyID, plan.Key, plan.Cutoff)
	}
	adapter, retained, err := researchmemory.RebuildFromBoundLabels(ctx, plan.Target, plan.Tenant, plan.Epoch, plan.Seed, plan.Cutoff, labels, validate)
	if err != nil {
		return nil, err
	}
	if s.store.Snapshot(ctx) != plan.Target {
		return nil, errors.New("research bound target moved during preparation")
	}
	candidate := &ResearchBoundPublication{Target: plan.Target, Epoch: plan.Epoch, Cutoff: plan.Cutoff, Retained: retained, TerminalCount: len(labels), model: adapter.Freeze()}
	return &PreparedResearchBound{service: s, slot: slot, durable: durable, base: base, candidate: candidate, sequence: sequence}, nil
}

// PublishPreparedResearchBound linearizes the swap while both the event store
// and durable stream are unchanged. Store-before-durable lock order matches
// guarded feedback admission; the callback performs only an atomic CAS.
func (s *Service) PublishPreparedResearchBound(ctx context.Context, prepared *PreparedResearchBound) error {
	if s == nil || ctx == nil || prepared == nil || prepared.service != s || prepared.slot == nil || prepared.durable == nil || prepared.candidate == nil {
		return errors.New("invalid prepared research publication")
	}
	guard, ok := s.store.(interface {
		WithResearchSnapshotWait(context.Context, model.Snapshot, func() error) error
	})
	if !ok {
		return errors.New("store lacks exact research publication guard")
	}
	return guard.WithResearchSnapshotWait(ctx, prepared.candidate.Target, func() error {
		return prepared.durable.WithUnchangedSequence(ctx, prepared.sequence, func() error {
			if prepared.base != nil && prepared.candidate.Epoch <= prepared.base.Epoch {
				return errors.New("research publication epoch did not advance")
			}
			if !prepared.slot.current.CompareAndSwap(prepared.base, prepared.candidate) {
				return errors.New("research publication slot changed")
			}
			return nil
		})
	})
}

// Score refuses stale source authority. A future-only ingestion may leave an
// earlier as-of query valid; a general mutation or already-available event may
// not. The caller supplies a deadline and chooses the fallback on rejection.
func (slot *ResearchBoundSlot) Score(ctx context.Context, s *Service, features uint16, baseline float64, at time.Time) (float64, error) {
	if slot == nil || s == nil || ctx == nil {
		return 0, errors.New("invalid research bound score")
	}
	p := slot.current.Load()
	if p == nil {
		return 0, errors.New("no published research bound model")
	}
	guard, ok := s.store.(interface {
		WithResearchAsOfSnapshotWait(context.Context, model.Snapshot, time.Time, func() error) error
	})
	if !ok {
		return 0, errors.New("store lacks as-of research score guard")
	}
	var score float64
	err := guard.WithResearchAsOfSnapshotWait(ctx, p.Target, at, func() error {
		if slot.current.Load() != p {
			return errors.New("research bound model changed before score")
		}
		var scoreErr error
		score, scoreErr = p.model.Score(features, baseline, p.Epoch, at)
		return scoreErr
	})
	if err != nil {
		return 0, err
	}
	return score, nil
}
