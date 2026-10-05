package researchledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"unicode/utf8"
)

const materializedBatchSQL = `SELECT sequence,length(identity),CASE WHEN length(CAST(identity AS BLOB))<=131072 THEN identity ELSE NULL END,length(payload),typeof(payload),CASE WHEN typeof(payload)='blob' AND length(payload)<=? THEN payload ELSE NULL END FROM materialized_source_trial WHERE source=? LIMIT 2`

func materializedSourceBatch(ctx context.Context, l *Ledger, requests []ServiceIdentity, afterRead func(int)) ([]ServiceLookupResult, error) {
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
	stmt, err := tx.PrepareContext(ctx, materializedBatchSQL)
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
		key, err := json.Marshal([]string{r.Tenant, r.Stream, r.Contract, r.Journal, r.Event})
		if err != nil {
			return nil, err
		}
		rows, err := stmt.QueryContext(ctx, limit, string(key))
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
