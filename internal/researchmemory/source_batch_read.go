package researchmemory

import (
	"context"
	"database/sql"
	"errors"
	"reflect"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

// OpenSourceOwnerBatchReads opts admission/discard source resolution into one
// transaction per batch. Replay, actual originals, writes and authority remain
// unchanged. The default constructor and single-item lookup stay the control.
func OpenSourceOwnerBatchReads(ctx context.Context, path, tenant, stream string, epoch uint64, seed int64) (*SourceOwner, error) {
	o, err := OpenSourceOwner(ctx, path, tenant, stream, epoch, seed)
	if err != nil {
		return nil, err
	}
	o.batchReads = true
	return o, nil
}

type SourceReference struct{ JournalID, EventID string }
type sourceResolution struct {
	record RecordedPrediction
	found  bool
}

// LookupBatch explicitly opts original readback into one snapshot. It requires
// every source to exist; missing/corrupt members return no partial slice. Reading
// an original never grants current service or feedback authority. Inputs must
// stay immutable until return. Duplicate read requests are allowed within caps.
func (o *SourceOwner) LookupBatch(ctx context.Context, refs []SourceReference) ([]RecordedPrediction, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	resolved, err := o.batchOriginals(ctx, refs)
	if err != nil {
		return nil, err
	}
	result := make([]RecordedPrediction, len(resolved))
	for i, r := range resolved {
		if !r.found {
			return nil, sql.ErrNoRows
		}
		result[i] = r.record
	}
	return result, nil
}

// Called under the private owner lock. The ledger snapshot cannot race another
// owner operation, and no caller-provided learner ID participates in resolution.
func (o *SourceOwner) batchOriginals(ctx context.Context, refs []SourceReference) ([]sourceResolution, error) {
	if o.d.closed || o.d.stopped {
		return nil, errors.New("source owner stopped")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(refs) == 0 || len(refs) > 256 {
		return nil, errors.New("invalid source read count")
	}
	keys := make([]researchledger.ServiceIdentity, len(refs))
	for i, r := range refs {
		keys[i] = o.key(r.JournalID, r.EventID)
	}
	rows, err := o.log.GetServiceAdmissions(ctx, keys)
	if err != nil {
		return nil, err
	}
	decoded, err := o.decodeSourceBatch(keys, rows)
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return decoded, nil
}

func (o *SourceOwner) decodeSourceBatch(keys []researchledger.ServiceIdentity, rows []researchledger.ServiceLookupResult) ([]sourceResolution, error) {
	if len(keys) == 0 || len(keys) > 256 || len(rows) != len(keys) {
		return nil, errors.New("invalid source read response count")
	}
	result := make([]sourceResolution, len(keys))
	total := 0
	for i, row := range rows {
		if row.Source != keys[i] {
			return nil, errors.New("source read response identity mismatch")
		}
		if !row.Found {
			if !reflect.DeepEqual(row.Entry, researchledger.Entry{}) {
				return nil, errors.New("malformed missing source")
			}
			continue
		}
		total += len(row.Entry.Payload)
		if len(row.Entry.Payload) == 0 || len(row.Entry.Payload) > 1<<20 || total > researchledger.MaxBatchBytes {
			return nil, errors.New("source read response byte cap")
		}
		r, err := o.sourceOriginal(row.Entry, keys[i].Journal, keys[i].Event)
		if err != nil {
			return nil, err
		}
		result[i] = sourceResolution{record: r, found: true}
	}
	return result, nil
}
