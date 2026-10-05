package researchledger

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

const MaxBatchEntries = 512
const MaxBatchBytes = 8 << 20

type AppendRequest struct {
	Key     Key
	Kind    string
	Payload json.RawMessage
}

type AppendResult struct {
	Sequence int64
	Retry    bool
}

// AppendBatch commits a bounded ordered set atomically. Exact retries retain
// their original sequences; conflicting bytes roll back the whole transaction.
// Inputs must remain immutable until return. An error returns no acknowledgments;
// a commit error can be uncertain, so callers must resolve it by replay/retry.
// This validates ledger structure, not evidence or forecast semantics.
func (l *Ledger) AppendBatch(ctx context.Context, requests []AppendRequest) ([]AppendResult, error) {
	return l.appendBatch(ctx, requests, nil)
}

// beforeCommit is a package-private fault-injection boundary. Production entry
// uses nil; no global hook can alter a concurrent ledger operation.
func (l *Ledger) appendBatch(ctx context.Context, requests []AppendRequest, beforeCommit func()) ([]AppendResult, error) {
	return l.appendBatchMode(ctx, requests, false, beforeCommit)
}

// AppendBatchPrepared reuses statements within one transaction only. It retains
// the same ordering, retry and durable acknowledgment contract as AppendBatch.
// The original API remains the independent query-per-record control.
func (l *Ledger) AppendBatchPrepared(ctx context.Context, requests []AppendRequest) ([]AppendResult, error) {
	return l.appendBatchMode(ctx, requests, true, nil)
}

func (l *Ledger) appendBatchMode(ctx context.Context, requests []AppendRequest, prepared bool, beforeCommit func()) ([]AppendResult, error) {
	return l.appendBatchStrategy(ctx, requests, prepared, false, beforeCommit)
}

// AppendBatchConditionalInsert is an opt-in statement-count experiment. Fresh
// admissions use one conditional INSERT; retries still compare exact bytes.
// No uniqueness failure is ignored, and feedback retains its admission check.
func (l *Ledger) AppendBatchConditionalInsert(ctx context.Context, requests []AppendRequest) ([]AppendResult, error) {
	return l.appendBatchStrategy(ctx, requests, true, true, nil)
}

func (l *Ledger) appendBatchStrategy(ctx context.Context, requests []AppendRequest, prepared, conditional bool, beforeCommit func()) ([]AppendResult, error) {
	return l.appendBatchPlan(ctx, requests, prepared, conditional, false, beforeCommit)
}

func (l *Ledger) appendBatchPlan(ctx context.Context, requests []AppendRequest, prepared, conditional, bulk bool, beforeCommit func()) ([]AppendResult, error) {
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
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
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
	if beforeCommit != nil {
		beforeCommit()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return results, nil
}
