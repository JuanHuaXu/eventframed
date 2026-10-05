package researchmemory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

type batchDurableLog interface {
	AppendBatch(context.Context, []researchledger.AppendRequest) ([]researchledger.AppendResult, error)
}

type DiscardRequest struct {
	ID        uint64
	Available time.Time
}

// DiscardBatch atomically stores up to 256 explicit unlabeled terminals, then
// releases their pending predictions. Inputs must be immutable until return.
// Exact retries preserve terminal dates. No evidence clock or model is updated.
// A failed/uncertain append stops this owner until close/replay; it is not safe
// to assume that the transaction rolled back merely because no result returned.
func (d *Durable) DiscardBatch(ctx context.Context, requests []DiscardRequest) ([]bool, error) {
	return d.discardBatch(ctx, requests, false)
}

// DiscardBatchWithSnapshotReads changes only preflight retrieval; typed terminal
// and uncertainty semantics remain identical to DiscardBatch.
func (d *Durable) DiscardBatchWithSnapshotReads(ctx context.Context, requests []DiscardRequest) ([]bool, error) {
	return d.discardBatch(ctx, requests, true)
}

func (d *Durable) discardBatch(ctx context.Context, requests []DiscardRequest, snapshotReads bool) ([]bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.stopped {
		return nil, errors.New("durable stream stopped")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(requests) == 0 || len(requests) > 256 {
		return nil, errors.New("invalid discard batch size")
	}
	log, ok := d.log.(batchDurableLog)
	if !ok {
		return nil, errors.New("log lacks atomic batch append")
	}
	var stored []researchledger.LookupResult
	if snapshotReads {
		ids := make([]uint64, len(requests))
		for i, r := range requests {
			ids[i] = r.ID
		}
		var err error
		stored, err = d.lookupBatch(ctx, ids, "feedback")
		if err != nil {
			return nil, err
		}
	}
	seen := make(map[uint64]bool, len(requests))
	entries := make([]researchledger.AppendRequest, len(requests))
	retries := make([]bool, len(requests))
	for i, request := range requests {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if request.ID == 0 || seen[request.ID] || request.Available.IsZero() {
			return nil, errors.New("invalid duplicate discard request")
		}
		seen[request.ID] = true
		key := d.key(request.ID)
		var old researchledger.Entry
		var err error
		if snapshotReads {
			old = stored[i].Entry
			if !stored[i].Found {
				err = sql.ErrNoRows
			}
		} else {
			old, err = d.log.Get(ctx, key, "feedback")
		}
		if err == nil {
			var r RecordedDiscard
			if err = decodeCanonical(old.Payload, &r); err != nil || !r.Discard || r.ID != request.ID || !r.Available.Equal(request.Available) {
				return nil, errors.New("conflicting discard retry or labeled prediction")
			}
			retries[i] = true
			// Equivalent timestamp instants can have different JSON encodings.
			// A retry must preserve the canonical terminal already in the log.
			entries[i] = researchledger.AppendRequest{Key: key, Kind: "feedback", Payload: old.Payload}
			continue
		} else {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			p, err := d.worker.Record(request.ID)
			if err != nil || request.Available.Before(p.At) {
				return nil, errors.New("invalid discard")
			}
		}
		raw, err := json.Marshal(RecordedDiscard{request.ID, true, request.Available})
		if err != nil {
			return nil, err
		}
		entries[i] = researchledger.AppendRequest{Key: key, Kind: "feedback", Payload: raw}
	}
	// Stop before crossing the uncertain durable boundary, including panics.
	d.stopped = true
	results, err := log.AppendBatch(ctx, entries)
	if err != nil {
		return nil, err
	}
	if len(results) != len(requests) {
		return nil, errors.New("invalid batch append acknowledgment")
	}
	for i, r := range results {
		if r.Sequence <= 0 || r.Retry != retries[i] {
			return nil, errors.New("inconsistent batch append acknowledgment")
		}
	}
	for i, request := range requests {
		if !retries[i] {
			d.worker.Discard(request.ID)
		}
	}
	d.stopped = false
	return retries, nil
}
