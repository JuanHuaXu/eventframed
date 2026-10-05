package researchledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

type LookupRequest struct {
	Key  Key
	Kind string
}
type LookupResult struct {
	Entry Entry
	Found bool
}

// GetBatch reads up to 512 keys in caller order from one transaction snapshot.
// Missing rows are explicit, not read failures. Encoded request bytes and returned
// payload bytes each have an independent 8MiB cap; individual payloads retain
// the 1MiB write bound. Duplicate requests are allowed and count toward bounds.
// Inputs must remain immutable until return. This grants no service authority.
func (l *Ledger) GetBatch(ctx context.Context, requests []LookupRequest) ([]LookupResult, error) {
	return l.getBatch(ctx, requests, nil)
}

// The per-call hook tests snapshot/cancellation boundaries without global state.
func (l *Ledger) getBatch(ctx context.Context, requests []LookupRequest, afterRead func(int)) ([]LookupResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(requests) == 0 || len(requests) > MaxBatchEntries {
		return nil, errors.New("invalid lookup batch count")
	}
	identities := make([]string, len(requests))
	size := 0
	for i, r := range requests {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for _, s := range []string{r.Key.Tenant, r.Key.Journal, r.Key.Event, r.Key.Contract} {
			if len(s) == 0 || len(s) > 4096 {
				return nil, errors.New("invalid lookup identity")
			}
		}
		if r.Kind != "admit" && r.Kind != "feedback" {
			return nil, errors.New("invalid lookup kind")
		}
		b, err := json.Marshal(r.Key)
		if err != nil {
			return nil, err
		}
		size += len(b) + len(r.Kind)
		if size > MaxBatchBytes {
			return nil, errors.New("lookup request byte cap exceeded")
		}
		identities[i] = string(b)
	}
	tx, err := l.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// Check length inside SQL before materializing an oversized/corrupted BLOB.
	stmt, err := tx.PrepareContext(ctx, `SELECT sequence,length(payload),typeof(payload),CASE WHEN typeof(payload)='blob' AND length(payload)<=? THEN payload ELSE NULL END FROM research_log WHERE identity=? AND kind=?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	results := make([]LookupResult, len(requests))
	remaining := MaxBatchBytes
	for i, r := range requests {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		entry := Entry{Key: r.Key, Kind: r.Kind}
		limit := remaining
		if limit > 1<<20 {
			limit = 1 << 20
		}
		var length int64
		var storageType string
		err := stmt.QueryRowContext(ctx, limit, identities[i], r.Kind).Scan(&entry.Sequence, &length, &storageType, &entry.Payload)
		if errors.Is(err, sql.ErrNoRows) {
			results[i] = LookupResult{Entry: Entry{Key: r.Key, Kind: r.Kind}}
		} else {
			if err != nil {
				return nil, err
			}
			if storageType != "blob" || length < 0 || length > int64(limit) || int64(len(entry.Payload)) != length || entry.Sequence <= 0 {
				return nil, errors.New("invalid or oversized lookup payload")
			}
			remaining -= len(entry.Payload)
			results[i] = LookupResult{Entry: entry, Found: true}
		}
		if afterRead != nil {
			afterRead(i)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return results, nil
}
