package researchmemory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

type AdmissionRequest struct {
	ID       uint64
	Features uint16
	Baseline float64
	At       time.Time
	Binding  *ServiceBinding
}

type AdmissionResult struct {
	Record RecordedPrediction
	Retry  bool
}

// AdmitBatch stages actual owned-worker forecasts and atomically records them.
// Up to 256 unique IDs are allowed; new IDs must follow the durable sequence.
// Each forecast keeps the snapshot actually used by Predict, which can differ
// between records while asynchronous fitting proceeds. No preview is relabeled.
// Caller inputs must be immutable until return. Unsupported logs reject before
// staging; staging/commit uncertainty stops this owner until close and replay.
func (d *Durable) AdmitBatch(ctx context.Context, requests []AdmissionRequest) ([]AdmissionResult, error) {
	return d.admitBatch(ctx, requests, false)
}

// AdmitBatchWithSnapshotReads uses bounded transactional preflight reads. The
// write/original-forecast contract is identical; the per-key API remains control.
func (d *Durable) AdmitBatchWithSnapshotReads(ctx context.Context, requests []AdmissionRequest) ([]AdmissionResult, error) {
	return d.admitBatch(ctx, requests, true)
}

func (d *Durable) admitBatch(ctx context.Context, requests []AdmissionRequest, snapshotReads bool) ([]AdmissionResult, error) {
	return d.admitBatchResolved(ctx, requests, snapshotReads, nil)
}

// resolved is private, call-local source resolution from SourceOwner while its
// exclusive lock is held. It is never a caller hint or a retained cache. The
// worker can publish fits, but only this owner allocates IDs; AppendBatch still
// checks exact persisted identities and retry bytes before acknowledging anything.
func (d *Durable) admitBatchResolved(ctx context.Context, requests []AdmissionRequest, snapshotReads bool, resolved []sourceResolution) ([]AdmissionResult, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	var phaseStart time.Time
	if d.measurement != nil {
		phaseStart = time.Now()
	}
	if d.closed || d.stopped {
		return nil, errors.New("durable stream stopped")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(requests) == 0 || len(requests) > 256 {
		return nil, errors.New("invalid admission batch size")
	}
	if resolved != nil && len(resolved) != len(requests) {
		return nil, errors.New("invalid resolved admission count")
	}
	log, ok := d.log.(batchDurableLog)
	if !ok {
		return nil, errors.New("log lacks atomic batch append")
	}
	var stored []researchledger.LookupResult
	if snapshotReads && resolved == nil {
		ids := make([]uint64, len(requests))
		for i, r := range requests {
			ids[i] = r.ID
		}
		var err error
		stored, err = d.lookupBatch(ctx, ids, "admit")
		if err != nil {
			return nil, err
		}
	}
	d.worker.mu.Lock()
	next, pending := d.worker.next, len(d.worker.pending)
	d.worker.mu.Unlock()
	seen := make(map[uint64]bool, len(requests))
	results := make([]AdmissionResult, len(requests))
	bindings := make([]*ServiceBinding, len(requests))
	for i, request := range requests {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if request.ID == 0 || seen[request.ID] {
			return nil, errors.New("duplicate or invalid admission ID")
		}
		seen[request.ID] = true
		if request.Binding != nil {
			b := *request.Binding
			if err := b.validate(); err != nil {
				return nil, err
			}
			if b.Tenant != d.tenant {
				return nil, errors.New("binding tenant mismatch")
			}
			bindings[i] = &b
		}
		var old researchledger.Entry
		var r RecordedPrediction
		var err error
		if resolved != nil {
			if bindings[i] == nil {
				return nil, errors.New("resolved admission requires source binding")
			}
			if resolved[i].found {
				r = resolved[i].record
			} else {
				if !reflect.DeepEqual(resolved[i].record, RecordedPrediction{}) {
					return nil, errors.New("malformed missing admission resolution")
				}
				err = sql.ErrNoRows
			}
		} else if snapshotReads {
			old = stored[i].Entry
			if !stored[i].Found {
				err = sql.ErrNoRows
			}
		} else {
			old, err = d.log.Get(ctx, d.key(request.ID), "admit")
		}
		if err == nil {
			if resolved == nil {
				if err = decodeCanonical(old.Payload, &r); err != nil {
					return nil, err
				}
			}
			if err = r.Validate(); err != nil {
				return nil, err
			}
			if r.Prediction.ID != request.ID || r.Prediction.Features != request.Features || r.Outer[0] != request.Baseline || !r.At.Equal(request.At) || !reflect.DeepEqual(r.Binding, bindings[i]) {
				return nil, errors.New("conflicting admission retry")
			}
			results[i] = AdmissionResult{Record: r, Retry: true}
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if next == ^uint64(0) || request.ID != next+1 || pending >= 256 {
			return nil, errors.New("admission batch order or capacity mismatch")
		}
		if _, err = d.worker.Snapshot().Score(request.Features, request.Baseline, d.epoch, request.At); err != nil {
			return nil, err
		}
		next++
		pending++
	}
	// Once staging starts, any error or panic can leave the in-memory stream
	// ahead of its log. Do not resume it optimistically, even on a known rollback.
	if d.measurement != nil {
		d.measurement.PreflightNS += time.Since(phaseStart).Nanoseconds()
		phaseStart = time.Now()
	}
	d.stopped = true
	entries := make([]researchledger.AppendRequest, len(requests))
	for i, request := range requests {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !results[i].Retry {
			p, err := d.worker.Predict(request.Features, request.Baseline, d.epoch, request.At)
			if err != nil {
				return nil, err
			}
			if p.ID != request.ID {
				return nil, errors.New("staged admission ID mismatch")
			}
			r, err := d.worker.Record(p.ID)
			if err != nil {
				return nil, err
			}
			r.Binding = bindings[i]
			results[i].Record = r
		}
		raw, err := json.Marshal(results[i].Record)
		if err != nil {
			return nil, err
		}
		entries[i] = researchledger.AppendRequest{Key: d.key(request.ID), Kind: "admit", Payload: raw}
	}
	if d.measurement != nil {
		d.measurement.StageNS += time.Since(phaseStart).Nanoseconds()
		phaseStart = time.Now()
	}
	ack, err := log.AppendBatch(ctx, entries)
	if d.measurement != nil {
		d.measurement.AppendNS += time.Since(phaseStart).Nanoseconds()
	}
	if err != nil {
		return nil, err
	}
	if len(ack) != len(results) {
		return nil, errors.New("invalid batch append acknowledgment")
	}
	for i, r := range ack {
		if r.Sequence <= 0 || r.Retry != results[i].Retry {
			return nil, errors.New("inconsistent batch append acknowledgment")
		}
	}
	d.stopped = false
	return results, nil
}
