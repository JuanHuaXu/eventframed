package researchledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"unicode/utf8"
)

// Source binds every result to its requested logical identity, including misses.
// A miss has Found=false and a zero Entry; a read failure returns no result slice.
type ServiceLookupResult struct {
	Source ServiceIdentity
	Entry  Entry
	Found  bool
}

const serviceBatchLookupSQL = `SELECT sequence,length(identity),CASE WHEN length(CAST(identity AS BLOB))<=131072 THEN identity ELSE NULL END,length(payload),typeof(payload),CASE WHEN typeof(payload)='blob' AND length(payload)<=? THEN payload ELSE NULL END FROM research_log INDEXED BY research_service_identity_v1 WHERE ` + serviceLookupWhere + ` LIMIT 2`

// GetServiceAdmissions reads at most 512 source keys, in caller order, from one
// transaction snapshot. Duplicate requests are allowed and count toward bounds.
// Encoded requests and returned payloads each have an independent 8MiB cap;
// individual payloads are at most 1MiB and identity envelopes at most 128KiB.
// Inputs remain immutable until return. Index activation and exclusive schema
// ownership are required; this resolves originals, not current evidence authority.
func (l *Ledger) GetServiceAdmissions(ctx context.Context, requests []ServiceIdentity) ([]ServiceLookupResult, error) {
	return l.getServiceAdmissions(ctx, requests, nil)
}

// The local hook exercises snapshot, cancellation and panic boundaries in tests.
func (l *Ledger) getServiceAdmissions(ctx context.Context, requests []ServiceIdentity, afterRead func(int)) ([]ServiceLookupResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(requests) == 0 || len(requests) > MaxBatchEntries {
		return nil, errors.New("invalid source lookup batch count")
	}
	size := 0
	for _, r := range requests {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for _, s := range []string{r.Tenant, r.Stream, r.Contract, r.Journal, r.Event} {
			if len(s) == 0 || len(s) > 4096 || !utf8.ValidString(s) {
				return nil, errors.New("invalid source lookup key")
			}
		}
		raw, err := json.Marshal(r)
		if err != nil {
			return nil, err
		}
		size += len(raw)
		if size > MaxBatchBytes {
			return nil, errors.New("source lookup request byte cap exceeded")
		}
	}
	tx, err := l.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, serviceBatchLookupSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	results := make([]ServiceLookupResult, len(requests))
	remaining := MaxBatchBytes
	for i, r := range requests {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		limit := remaining
		if limit > 1<<20 {
			limit = 1 << 20
		}
		rows, err := stmt.QueryContext(ctx, limit, r.Tenant, r.Stream, r.Contract, r.Journal, r.Event)
		if err != nil {
			return nil, err
		}
		entry, err := readServiceAdmission(rows, r, limit)
		if errors.Is(err, sql.ErrNoRows) {
			results[i] = ServiceLookupResult{Source: r}
		} else if err != nil {
			return nil, err
		} else {
			remaining -= len(entry.Payload)
			results[i] = ServiceLookupResult{Source: r, Entry: entry, Found: true}
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
