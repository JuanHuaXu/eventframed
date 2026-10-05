package researchmemory

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

// BoundLabel pairs an original forecast record with its terminal feedback.
// It is not self-authenticating: the caller must still verify the durable pair,
// surviving source and feature semantics against the target snapshot.
type BoundLabel struct {
	Prediction RecordedPrediction
	Feedback   RecordedFeedback
}

type BoundLabelValidator func(context.Context, model.Snapshot, BoundLabel) (bool, error)

// SealedBoundBootstrap keeps the rebuilt model and its evidence commitment
// together. Only the validating rebuild can construct one outside this package.
type SealedBoundBootstrap struct {
	state *sealedBoundState
}

type sealedBoundState struct {
	mu             sync.Mutex
	used           bool
	adapter        *Adapter
	seal           []byte
	tenant, stream string
	epoch          uint64
	seed           int64
	motion         bool
	origin         model.Snapshot
}

// RebuildFromBoundLabels reconstructs only component learners in a new research
// epoch. The caller owns a quiescent, ordered source log and must verify that the
// target snapshot remains current before using the result. Neither unresolved
// predictions nor old mixture weights cross the epoch boundary.
func RebuildFromBoundLabels(ctx context.Context, target model.Snapshot, tenant string, epoch uint64, seed int64, cutoff time.Time, labels []BoundLabel, validate BoundLabelValidator) (*Adapter, int, error) {
	a, n, _, err := rebuildBoundLabels(ctx, target, tenant, epoch, seed, cutoff, labels, validate, "", "", nil, nil)
	return a, n, err
}

// RebuildSealedBoundLabels derives the epoch commitment while validation and
// component reconstruction observe the same records. The HMAC key remains with
// the caller; the seal alone cannot authenticate an unvalidated source.
func RebuildSealedBoundLabels(ctx context.Context, target model.Snapshot, tenant, stream string, epoch uint64, seed int64, cutoff time.Time, labels []BoundLabel, validate BoundLabelValidator, keyID string, key []byte) (*SealedBoundBootstrap, int, error) {
	if stream == "" || keyID == "" || len(key) < 32 {
		return nil, 0, errors.New("invalid bound bootstrap key")
	}
	a, n, seal, err := rebuildBoundLabels(ctx, target, tenant, epoch, seed, cutoff, labels, validate, stream, keyID, key, nil)
	if err != nil {
		return nil, 0, err
	}
	return &SealedBoundBootstrap{state: &sealedBoundState{adapter: a, seal: seal, tenant: tenant, stream: stream, epoch: epoch, seed: seed}}, n, nil
}

// RebuildMotionSealedBoundLabels validates against the current target while
// committing to the original bootstrap snapshot. Future-only target motion
// can preserve the seal, but the caller must prove as-of compatibility first.
func RebuildMotionSealedBoundLabels(ctx context.Context, target, origin model.Snapshot, tenant, stream string, epoch uint64, seed int64, cutoff time.Time, labels []BoundLabel, validate BoundLabelValidator, keyID string, key []byte) (*SealedBoundBootstrap, int, error) {
	if stream == "" || keyID == "" || len(key) < 32 || origin.RuntimeVersion == 0 || origin.ContractVersion == 0 {
		return nil, 0, errors.New("invalid motion bootstrap contract")
	}
	a, n, seal, err := rebuildBoundLabels(ctx, target, tenant, epoch, seed, cutoff, labels, validate, stream, keyID, key, &origin)
	if err != nil {
		return nil, 0, err
	}
	return &SealedBoundBootstrap{state: &sealedBoundState{adapter: a, seal: seal, tenant: tenant, stream: stream, epoch: epoch, seed: seed, motion: true, origin: origin}}, n, nil
}

func writeBoundSeal(h hash.Hash, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(raw)))
	_, _ = h.Write(length[:])
	_, _ = h.Write(raw)
	return nil
}

func rebuildBoundLabels(ctx context.Context, target model.Snapshot, tenant string, epoch uint64, seed int64, cutoff time.Time, labels []BoundLabel, validate BoundLabelValidator, stream, keyID string, key []byte, origin *model.Snapshot) (*Adapter, int, []byte, error) {
	if ctx == nil || validate == nil || tenant == "" || epoch == 0 || cutoff.IsZero() || target.RuntimeVersion == 0 || target.ContractVersion == 0 || len(labels) > 256 {
		return nil, 0, nil, errors.New("invalid bound-label rebuild contract")
	}
	var seal hash.Hash
	if key != nil {
		seal = hmac.New(sha256.New, key)
		var err error
		if origin == nil {
			err = writeBoundSeal(seal, struct {
				Contract string
				Target   model.Snapshot
				Tenant   string
				Stream   string
				Epoch    uint64
				Seed     int64
				Cutoff   time.Time
				KeyID    string
			}{"eventframed-bound-bootstrap-v1", target, tenant, stream, epoch, seed, cutoff, keyID})
		} else {
			err = writeBoundSeal(seal, struct {
				Contract string
				Origin   model.Snapshot
				Tenant   string
				Stream   string
				Epoch    uint64
				Seed     int64
				Cutoff   time.Time
				KeyID    string
			}{"eventframed-bound-motion-bootstrap-v1", *origin, tenant, stream, epoch, seed, cutoff, keyID})
		}
		if err != nil {
			return nil, 0, nil, err
		}
	}
	a := New(epoch, seed)
	seen := make(map[boundLabelKey]struct{}, len(labels))
	seenIDs := make(map[uint64]struct{}, len(labels))
	var previous time.Time
	for i, r := range labels {
		if err := ctx.Err(); err != nil {
			return nil, 0, nil, err
		}
		if err := r.Prediction.Validate(); err != nil {
			return nil, 0, nil, fmt.Errorf("bound label %d: %w", i, err)
		}
		binding := r.Prediction.Binding
		if binding == nil {
			return nil, 0, nil, fmt.Errorf("unbound label %d", i)
		}
		key := boundLabelKey{binding.JournalID, binding.EventID}
		id := r.Prediction.Prediction.ID
		_, duplicateKey := seen[key]
		_, duplicateID := seenIDs[id]
		if duplicateKey || duplicateID || binding.Tenant != tenant || r.Feedback.ID != id || r.Prediction.Prediction.Epoch == 0 || r.Feedback.Available.IsZero() || r.Feedback.Available.Before(r.Prediction.At) || r.Feedback.Available.Before(previous) || r.Feedback.Available.After(cutoff) {
			return nil, 0, nil, fmt.Errorf("invalid bound label %d", i)
		}
		seen[key] = struct{}{}
		seenIDs[id] = struct{}{}
		previous = r.Feedback.Available
		retain, err := validate(ctx, target, r)
		if err != nil {
			return nil, 0, nil, fmt.Errorf("validate bound label %d: %w", i, err)
		}
		if seal != nil {
			if err = writeBoundSeal(seal, struct {
				Label  BoundLabel
				Retain bool
			}{r, retain}); err != nil {
				return nil, 0, nil, err
			}
		}
		if retain {
			a.samples = append(a.samples, observation.Sample{Bits: r.Prediction.Prediction.Features, Outcome: r.Feedback.Useful})
			a.lastFeedback = r.Feedback.Available
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, 0, nil, err
	}
	a.total = len(a.samples)
	if a.total >= 32 {
		var err error
		shortSamples := a.samples[max(0, a.total-64):]
		if a.short, err = observation.Fit(shortSamples); err != nil {
			return nil, 0, nil, err
		}
		if a.long, err = observation.Fit(a.samples); err != nil {
			return nil, 0, nil, err
		}
		a.forest = observationlearners.NewForest(seed)
		for _, sample := range shortSamples {
			a.forest.Update(sample.Bits, sample.Outcome)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, 0, nil, err
	}
	if seal != nil {
		return a, a.total, seal.Sum(nil), nil
	}
	return a, a.total, nil, nil
}

type boundLabelKey struct{ journal, event string }
