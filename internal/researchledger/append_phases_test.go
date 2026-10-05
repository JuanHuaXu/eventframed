package researchledger

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type appendPhases struct{ TotalNS, PreflightNS, BeginNS, PrepareNS, RowsNS, CommitNS int64 }

// Test-only phase attribution. The prepared transaction, source index, retry,
// atomicity and commit semantics are unchanged; no validity guard is released.
func timedPreparedAppend(l *Ledger, ctx context.Context, requests []AppendRequest, trace *appendPhases, beforeCommit func()) ([]AppendResult, error) {
	const prepared, conditional, bulk = true, false, false
	started := time.Now()
	defer func() { trace.TotalNS = time.Since(started).Nanoseconds() }()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(requests) == 0 || len(requests) > MaxBatchEntries {
		return nil, errors.New("invalid ledger batch count")
	}
	identities := make([]string, len(requests))
	total := 0
	for i, request := range requests {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for _, field := range []string{request.Key.Tenant, request.Key.Journal, request.Key.Event, request.Key.Contract} {
			if len(field) == 0 || len(field) > 4096 {
				return nil, errors.New("invalid ledger identity")
			}
		}
		if (request.Kind != "admit" && request.Kind != "feedback") || len(request.Payload) > 1<<20 || !json.Valid(request.Payload) {
			return nil, errors.New("invalid ledger record")
		}
		identity, err := json.Marshal(request.Key)
		if err != nil {
			return nil, err
		}
		total += len(identity) + len(request.Kind) + len(request.Payload)
		if total > MaxBatchBytes {
			return nil, errors.New("ledger batch byte cap exceeded")
		}
		identities[i] = string(identity)
	}
	trace.PreflightNS = time.Since(started).Nanoseconds()
	beginStart := time.Now()
	tx, err := l.db.BeginTx(ctx, nil)
	trace.BeginNS = time.Since(beginStart).Nanoseconds()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	prepareStart := time.Now()
	if bulk {
		allAdmissions := true
		for _, r := range requests {
			allAdmissions = allAdmissions && r.Kind == "admit"
		}
		if allAdmissions {
			return bulkAdmissionTx(ctx, tx, requests, identities, beforeCommit)
		}
	}
	const lookupSQL = "SELECT sequence,payload FROM research_log WHERE identity=? AND kind=?"
	const admissionSQL = "SELECT COUNT(*) FROM research_log WHERE identity=? AND kind='admit'"
	const insertSQL = "INSERT INTO research_log(identity,kind,payload) VALUES(?,?,?)"
	var statements [3]*sql.Stmt
	if prepared {
		for i, query := range []string{lookupSQL, admissionSQL, insertSQL} {
			statements[i], err = tx.PrepareContext(ctx, query)
			if err != nil {
				return nil, err
			}
			defer statements[i].Close()
		}
	}
	var fresh *sql.Stmt
	if conditional {
		fresh, err = tx.PrepareContext(ctx, "INSERT INTO research_log(identity,kind,payload) SELECT ?,?,? WHERE NOT EXISTS (SELECT 1 FROM research_log WHERE identity=? AND kind=?)")
		if err != nil {
			return nil, err
		}
		defer fresh.Close()
	}
	trace.PrepareNS = time.Since(prepareStart).Nanoseconds()
	rowStart := time.Now()
	results := make([]AppendResult, len(requests))
	for i, request := range requests {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var seq int64
		var old []byte
		var err error
		if conditional && request.Kind == "admit" {
			insert, e := fresh.ExecContext(ctx, identities[i], request.Kind, []byte(request.Payload), identities[i], request.Kind)
			if e != nil {
				return nil, e
			}
			n, e := insert.RowsAffected()
			if e != nil {
				return nil, e
			}
			if n == 1 {
				seq, e := insert.LastInsertId()
				if e != nil || seq <= 0 {
					return nil, errors.New("invalid conditional insert sequence")
				}
				results[i] = AppendResult{Sequence: seq}
				continue
			}
			if n != 0 {
				return nil, errors.New("invalid conditional insert count")
			}
		}
		if prepared {
			err = statements[0].QueryRowContext(ctx, identities[i], request.Kind).Scan(&seq, &old)
		} else {
			err = tx.QueryRowContext(ctx, lookupSQL, identities[i], request.Kind).Scan(&seq, &old)
		}
		if err == nil {
			if !bytes.Equal(old, request.Payload) {
				return nil, errors.New("conflicting ledger retry")
			}
			results[i] = AppendResult{Sequence: seq, Retry: true}
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if conditional && request.Kind == "admit" {
			return nil, errors.New("conditional retry row missing")
		}
		if request.Kind == "feedback" {
			var n int
			if prepared {
				err = statements[1].QueryRowContext(ctx, identities[i]).Scan(&n)
			} else {
				err = tx.QueryRowContext(ctx, admissionSQL, identities[i]).Scan(&n)
			}
			if err != nil {
				return nil, err
			}
			if n != 1 {
				return nil, errors.New("feedback without admission")
			}
		}
		var insert sql.Result
		if prepared {
			insert, err = statements[2].ExecContext(ctx, identities[i], request.Kind, []byte(request.Payload))
		} else {
			insert, err = tx.ExecContext(ctx, insertSQL, identities[i], request.Kind, []byte(request.Payload))
		}
		if err != nil {
			return nil, err
		}
		seq, err = insert.LastInsertId()
		if err != nil {
			return nil, err
		}
		results[i] = AppendResult{Sequence: seq}
	}
	trace.RowsNS = time.Since(rowStart).Nanoseconds()
	if beforeCommit != nil {
		beforeCommit()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	commitStart := time.Now()
	err = tx.Commit()
	trace.CommitNS = time.Since(commitStart).Nanoseconds()
	if err != nil {
		return nil, err
	}
	return results, nil
}
