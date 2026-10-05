package researchmemory

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"time"
)

type VerifiedSourceDiscardRequest struct {
	Original  RecordedPrediction
	Available time.Time
}

// VerifyAndDiscardBatch reads each persisted original once, compares the entire
// expected record, then atomically records unlabeled terminals. Expected learner
// IDs are assertions, never authority: IDs used for cleanup come from the source
// index and canonical stored record. Every member must pass before any write.
// The owner lock spans resolution, comparison and terminal commit. No evidence,
// model-serving or current service authority is granted. Inputs stay immutable
// until return; uncertainty still stops this owner until close and replay.
func (o *SourceOwner) VerifyAndDiscardBatch(ctx context.Context, requests []VerifiedSourceDiscardRequest) ([]bool, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.d.closed || o.d.stopped {
		return nil, errors.New("source owner stopped")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(requests) == 0 || len(requests) > 256 {
		return nil, errors.New("invalid verified discard count")
	}
	refs := make([]SourceReference, len(requests))
	seen := make(map[SourceReference]bool, len(requests))
	for i, req := range requests {
		if err := req.Original.Validate(); err != nil {
			return nil, err
		}
		b := req.Original.Binding
		if b == nil || b.Tenant != o.d.tenant || req.Available.IsZero() || req.Available.Before(req.Original.At) {
			return nil, errors.New("invalid verified discard binding/time")
		}
		ref := SourceReference{b.JournalID, b.EventID}
		if seen[ref] {
			return nil, errors.New("duplicate verified discard source")
		}
		seen[ref] = true
		refs[i] = ref
	}
	originals, err := o.batchOriginals(ctx, refs)
	if err != nil {
		return nil, err
	}
	resolved := make([]DiscardRequest, len(requests))
	for i, original := range originals {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !original.found {
			return nil, sql.ErrNoRows
		}
		expected := requests[i].Original
		if !original.record.At.Equal(expected.At) {
			return nil, errors.New("verified discard original time mismatch")
		}
		// Monotonic clock readings and location pointers do not survive JSON.
		// Only their representation is normalized; the instant must match first.
		expected.At = original.record.At
		if !reflect.DeepEqual(original.record, expected) {
			return nil, errors.New("verified discard original mismatch")
		}
		resolved[i] = DiscardRequest{ID: original.record.Prediction.ID, Available: requests[i].Available}
	}
	return o.d.DiscardBatchWithSnapshotReads(ctx, resolved)
}
