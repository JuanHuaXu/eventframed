package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

// PreparedResearchBoundWorker is a single-use, research-only new-epoch handoff.
// Its model and seal are derived from the complete cutoff-eligible durable
// terminal stream, with source validation against one target snapshot.
type PreparedResearchBoundWorker struct {
	mu        sync.Mutex
	used      bool
	service   *Service
	source    *researchmemory.Durable
	target    model.Snapshot
	sequence  int64
	bootstrap *researchmemory.SealedBoundBootstrap
	keyID     string
	key       []byte
	motion    bool
	origin    model.Snapshot
	cutoff    time.Time
}

// ResearchBoundWorker owns a sealed, continuing research epoch. Its handle
// does not expose the raw Durable, so service-originating admissions and
// feedback must cross a fresh source guard. It never changes served answers.
type ResearchBoundWorker struct {
	mu      sync.Mutex
	closed  bool
	service *Service
	durable *researchmemory.Durable
	keyID   string
	key     []byte
}

type researchSourceAsOfGuard interface {
	WithResearchSourceAsOfSnapshotWait(context.Context, model.Snapshot, time.Time, string, string, func(model.Snapshot) error) error
}

func (s *Service) PrepareResearchBoundWorker(ctx context.Context, source *researchmemory.Durable, plan ResearchBoundPlan, stream string) (*PreparedResearchBoundWorker, error) {
	return s.prepareResearchBoundWorker(ctx, source, plan, stream, "", false)
}

// PrepareResearchMotionBoundWorker reads the persisted origin of a motion-mode
// epoch, or fixes the current target as origin for a new empty ledger.
func (s *Service) PrepareResearchMotionBoundWorker(ctx context.Context, source *researchmemory.Durable, plan ResearchBoundPlan, stream, path string) (*PreparedResearchBoundWorker, error) {
	return s.prepareResearchBoundWorker(ctx, source, plan, stream, path, true)
}

func (s *Service) prepareResearchBoundWorker(ctx context.Context, source *researchmemory.Durable, plan ResearchBoundPlan, stream, path string, motion bool) (*PreparedResearchBoundWorker, error) {
	if s == nil || source == nil || ctx == nil || plan.Tenant == "" || plan.Tenant != source.Tenant() || stream == "" || plan.Epoch <= source.Epoch() || plan.KeyID == "" || len(plan.Key) < 32 {
		return nil, errors.New("invalid research bound worker plan")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.store.Snapshot(ctx) != plan.Target {
		return nil, errors.New("research worker target moved before preparation")
	}
	origin := plan.Target
	if motion {
		if path == "" {
			return nil, errors.New("motion worker path required")
		}
		ledger, err := researchledger.Open(path)
		if err != nil {
			return nil, err
		}
		seal, rawOrigin, readErr := ledger.MotionBootstrap(ctx)
		closeErr := ledger.Close()
		if err := errors.Join(readErr, closeErr); err != nil {
			return nil, err
		}
		if seal != nil {
			if err := json.Unmarshal(rawOrigin, &origin); err != nil {
				return nil, err
			}
			canonical, err := json.Marshal(origin)
			if err != nil || !bytes.Equal(canonical, rawOrigin) || origin.RuntimeVersion == 0 || origin.ContractVersion == 0 {
				return nil, errors.New("invalid stored motion origin")
			}
		}
		proof, ok := s.store.(interface {
			ResearchPublicationCompatible(context.Context, model.Snapshot, time.Time) bool
		})
		if !ok || !proof.ResearchPublicationCompatible(ctx, origin, plan.Cutoff) {
			return nil, errors.New("bootstrap origin is not as-of compatible")
		}
	}
	labels, sequence, err := source.BoundLabelsWithSequence(ctx, plan.Cutoff)
	if err != nil {
		return nil, err
	}
	validate := func(ctx context.Context, target model.Snapshot, label researchmemory.BoundLabel) (bool, error) {
		return s.ValidateResearchTransferSource(ctx, target, label.Prediction, plan.KeyID, plan.Key, plan.Cutoff)
	}
	var bootstrap *researchmemory.SealedBoundBootstrap
	if motion {
		bootstrap, _, err = researchmemory.RebuildMotionSealedBoundLabels(ctx, plan.Target, origin, plan.Tenant, stream, plan.Epoch, plan.Seed, plan.Cutoff, labels, validate, plan.KeyID, plan.Key)
	} else {
		bootstrap, _, err = researchmemory.RebuildSealedBoundLabels(ctx, plan.Target, plan.Tenant, stream, plan.Epoch, plan.Seed, plan.Cutoff, labels, validate, plan.KeyID, plan.Key)
	}
	if err != nil {
		return nil, err
	}
	if s.store.Snapshot(ctx) != plan.Target {
		return nil, errors.New("research worker target moved during preparation")
	}
	return &PreparedResearchBoundWorker{service: s, source: source, target: plan.Target, sequence: sequence, bootstrap: bootstrap, keyID: plan.KeyID, key: append([]byte(nil), plan.Key...), motion: motion, origin: origin, cutoff: plan.Cutoff}, nil
}

// OpenPreparedResearchBoundWorker replays privately, then uses store-before-
// source guards to certify the handoff before exposing the worker. It is not
// a hot-path API and may leave a sealed, unexposed log on failed certification.
func (s *Service) OpenPreparedResearchBoundWorker(ctx context.Context, prepared *PreparedResearchBoundWorker, path string) (*ResearchBoundWorker, error) {
	if s == nil || ctx == nil || prepared == nil || prepared.service != s || prepared.source == nil {
		return nil, errors.New("invalid prepared research worker")
	}
	if _, ok := ctx.Deadline(); !ok {
		return nil, errors.New("research worker open requires a deadline")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	prepared.mu.Lock()
	defer prepared.mu.Unlock()
	if prepared.used {
		return nil, errors.New("prepared research worker already consumed")
	}
	prepared.used = true
	guard, ok := s.store.(interface {
		WithResearchSnapshotWait(context.Context, model.Snapshot, func() error) error
	})
	if !ok {
		return nil, errors.New("store lacks exact research worker guard")
	}
	validate := func(ctx context.Context, original researchmemory.RecordedPrediction) error {
		if prepared.motion {
			proof, ok := s.store.(interface {
				ResearchPublicationCompatible(context.Context, model.Snapshot, time.Time) bool
			})
			if original.Binding == nil || !ok || !proof.ResearchPublicationCompatible(ctx, original.Binding.Snapshot, original.At) {
				return errors.New("bound worker original as-of history changed")
			}
		}
		valid, err := s.ValidateResearchTransferSource(ctx, prepared.target, original, prepared.keyID, prepared.key, original.At)
		if err != nil {
			return err
		}
		if !valid {
			return errors.New("bound worker original source changed")
		}
		return nil
	}
	// Replay may need its own continuity guard. Keep the worker private until
	// the final store-before-source guard certifies the prepared transition.
	var worker *researchmemory.Durable
	var err error
	if prepared.motion {
		validateOrigin := func(ctx context.Context, origin model.Snapshot) error {
			proof, ok := s.store.(interface {
				ResearchPublicationCompatible(context.Context, model.Snapshot, time.Time) bool
			})
			if origin != prepared.origin || s.store.Snapshot(ctx) != prepared.target || !ok || !proof.ResearchPublicationCompatible(ctx, origin, prepared.cutoff) {
				return errors.New("bound worker motion origin changed")
			}
			return nil
		}
		worker, err = researchmemory.OpenMotionBoundDurable(ctx, path, prepared.bootstrap, validateOrigin, validate)
	} else {
		worker, err = researchmemory.OpenBoundDurable(ctx, path, prepared.bootstrap, validate)
	}
	if err != nil {
		return nil, err
	}
	err = guard.WithResearchSnapshotWait(ctx, prepared.target, func() error {
		return prepared.source.WithUnchangedSequence(ctx, prepared.sequence, func() error {
			if prepared.motion {
				proof, ok := s.store.(interface {
					ResearchPublicationCompatible(context.Context, model.Snapshot, time.Time) bool
				})
				if !ok || !proof.ResearchPublicationCompatible(ctx, prepared.origin, prepared.cutoff) {
					return errors.New("bound worker motion proof changed before handoff")
				}
			}
			return ctx.Err()
		})
	})
	if err != nil {
		if worker != nil {
			_ = worker.Close()
		}
		return nil, err
	}
	return &ResearchBoundWorker{service: s, durable: worker, keyID: prepared.keyID, key: prepared.key}, nil
}

// Admit validates an observed candidate against the committed query journal
// and current event under the store guard, then journals the actual forecast.
// The source index makes exact retries idempotent across process restarts.
func (w *ResearchBoundWorker) Admit(ctx context.Context, request model.RecallRequest, observed ResearchFrontierObservation, candidate ResearchFrontierCandidate) (researchmemory.Prediction, bool, error) {
	if w == nil || ctx == nil {
		return researchmemory.Prediction{}, false, errors.New("invalid bound worker admission")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed || observed.JournalID == "" || observed.Snapshot.RuntimeVersion == 0 || !observed.AsOf.Equal(request.AsOf) || request.TenantID != w.durable.Tenant() {
		return researchmemory.Prediction{}, false, errors.New("stale or mismatched bound worker candidate")
	}
	seen := 0
	for _, c := range observed.Candidates {
		if c.EventID == candidate.EventID {
			seen++
			if c != candidate {
				return researchmemory.Prediction{}, false, errors.New("changed bound worker candidate")
			}
		}
	}
	if seen != 1 {
		return researchmemory.Prediction{}, false, errors.New("bound worker candidate not nominated once")
	}
	binding := researchmemory.ServiceBinding{Tenant: request.TenantID, JournalID: observed.JournalID, EventID: candidate.EventID, Snapshot: observed.Snapshot}
	guard, ok := w.service.store.(researchSourceAsOfGuard)
	if !ok {
		return researchmemory.Prediction{}, false, errors.New("bound worker lacks held source continuity proof")
	}
	var forecast researchmemory.Prediction
	var retry bool
	err := guard.WithResearchSourceAsOfSnapshotWait(ctx, binding.Snapshot, observed.AsOf, binding.Tenant, binding.EventID, func(target model.Snapshot) error {
		if err := w.service.validateResearchCandidate(ctx, binding, observed.AsOf, candidate.Features, candidate.Baseline, request); err != nil {
			return err
		}
		if w.service.store.Snapshot(ctx) != target {
			return errors.New("bound worker source snapshot moved")
		}
		journal, err := w.service.store.GetBayesianJournal(ctx, binding.Tenant, binding.JournalID)
		if err != nil {
			return err
		}
		events, err := w.service.store.GetEvents(ctx, binding.Tenant, []string{binding.EventID}, observed.AsOf)
		if err != nil {
			return err
		}
		if len(events) != 1 || events[0].ID != binding.EventID {
			return errors.New("bound worker source unavailable")
		}
		witness, err := researchmemory.NewSourceWitness(w.keyID, w.key, binding, journal.QueryDigest, events[0], candidate.Features, candidate.Baseline)
		if err != nil {
			return err
		}
		forecast, retry, err = w.durable.AdmitSourceBoundWithWitness(ctx, candidate.Features, candidate.Baseline, observed.AsOf, binding, witness)
		return err
	})
	return forecast, retry, err
}

// Feedback consumes only externally verified usefulness. It revalidates the
// original event/query before admitting the terminal record; the probability
// or rank cannot label itself.
func (w *ResearchBoundWorker) Feedback(ctx context.Context, request model.RecallRequest, journal, event string, useful bool, available time.Time) (bool, error) {
	if w == nil || ctx == nil {
		return false, errors.New("invalid bound worker feedback")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return false, errors.New("bound worker closed")
	}
	original, err := w.durable.AdmissionBySource(ctx, journal, event)
	if err != nil {
		return false, err
	}
	if original.Binding == nil {
		return false, errors.New("bound worker feedback lacks source binding")
	}
	guard, ok := w.service.store.(researchSourceAsOfGuard)
	if !ok {
		return false, errors.New("bound worker lacks held source continuity proof")
	}
	var retry bool
	err = guard.WithResearchSourceAsOfSnapshotWait(ctx, original.Binding.Snapshot, original.At, original.Binding.Tenant, original.Binding.EventID, func(target model.Snapshot) error {
		valid, err := w.service.validateResearchTransferSourceEvidence(ctx, target, original, w.keyID, w.key, original.At)
		if err != nil {
			return err
		}
		if !valid {
			return errors.New("bound worker feedback source changed")
		}
		if err := w.service.ValidateResearchAdmission(ctx, original, request); err != nil {
			return err
		}
		if w.service.store.Snapshot(ctx) != target {
			return errors.New("bound worker source snapshot moved")
		}
		var labelErr error
		retry, labelErr = w.durable.Feedback(ctx, original.Prediction.ID, useful, available)
		return labelErr
	})
	return retry, err
}

func (w *ResearchBoundWorker) Counts() (completed, failed uint64, pending, queued int) {
	return w.durable.Counts()
}

func (w *ResearchBoundWorker) WaitProcessed(ctx context.Context, target uint64) error {
	return w.durable.WaitProcessed(ctx, target)
}

func (w *ResearchBoundWorker) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return nil
	}
	w.closed = true
	err := w.durable.Close()
	for i := range w.key {
		w.key[i] = 0
	}
	return err
}
