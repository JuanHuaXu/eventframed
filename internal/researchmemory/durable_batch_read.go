package researchmemory

import (
	"context"
	"database/sql"
	"errors"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

type batchReadDurableLog interface {
	GetBatch(context.Context, []researchledger.LookupRequest) ([]researchledger.LookupResult, error)
}

// lookupBatch is called only while this owner holds d.mu. It validates the
// response envelope before missing rows may influence admission/terminal logic.
func (d *Durable) lookupBatch(ctx context.Context, ids []uint64, kind string) ([]researchledger.LookupResult, error) {
	if len(ids) == 0 || len(ids) > 256 {
		return nil, errors.New("invalid durable lookup count")
	}
	log, ok := d.log.(batchReadDurableLog)
	if !ok {
		return nil, errors.New("log lacks bounded batch reads")
	}
	requests := make([]researchledger.LookupRequest, len(ids))
	seen := make(map[uint64]bool, len(ids))
	for i, id := range ids {
		if id == 0 || seen[id] {
			return nil, errors.New("invalid durable lookup ID")
		}
		seen[id] = true
		requests[i] = researchledger.LookupRequest{Key: d.key(id), Kind: kind}
	}
	rows, err := log.GetBatch(ctx, requests)
	if err != nil {
		return nil, err
	}
	if len(rows) != len(requests) {
		return nil, errors.New("invalid batch read response count")
	}
	total := 0
	for i, r := range rows {
		if r.Entry.Key != requests[i].Key || r.Entry.Kind != kind {
			return nil, errors.New("batch read identity mismatch")
		}
		if !r.Found {
			if r.Entry.Sequence != 0 || len(r.Entry.Payload) != 0 {
				return nil, errors.New("malformed missing record")
			}
			continue
		}
		if r.Entry.Sequence <= 0 || len(r.Entry.Payload) == 0 || len(r.Entry.Payload) > 1<<20 {
			return nil, errors.New("invalid batch read record")
		}
		total += len(r.Entry.Payload)
		if total > researchledger.MaxBatchBytes {
			return nil, errors.New("batch read byte cap exceeded")
		}
	}
	return rows, nil
}

// Admissions reads original records together, not freshly recomputed forecasts.
// Missing or invalid members fail the entire read. This is not evidence authority.
func (d *Durable) Admissions(ctx context.Context, ids []uint64) ([]RecordedPrediction, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.stopped {
		return nil, errors.New("durable stream stopped")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := d.lookupBatch(ctx, ids, "admit")
	if err != nil {
		return nil, err
	}
	results := make([]RecordedPrediction, len(ids))
	for i, row := range rows {
		if !row.Found {
			return nil, sql.ErrNoRows
		}
		r := &results[i]
		if err := decodeCanonical(row.Entry.Payload, r); err != nil {
			return nil, err
		}
		if err := r.Validate(); err != nil {
			return nil, err
		}
		if r.Prediction.ID != ids[i] || (r.Binding != nil && r.Binding.Tenant != d.tenant) {
			return nil, errors.New("stored binding mismatch")
		}
	}
	return results, nil
}
