package researchledger

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
)

// AppendBatchBulkAdmissions is opt-in. Only all-admission batches use grouped
// VALUES; mixed batches retain ordered admission/feedback validation.
func (l *Ledger) AppendBatchBulkAdmissions(ctx context.Context, r []AppendRequest) ([]AppendResult, error) {
	return l.appendBatchPlan(ctx, r, true, false, true, nil)
}

func identityArgs(ids []string) (string, []any) {
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return strings.TrimSuffix(strings.Repeat("?,", len(ids)), ","), args
}

// Caller already validated count/byte limits and owns this transaction. Every
// receipt is read by identity; no assumption of contiguous row IDs is used.
func bulkAdmissionTx(ctx context.Context, tx *sql.Tx, r []AppendRequest, ids []string, beforeCommit func()) ([]AppendResult, error) {
	marks, args := identityArgs(ids)
	rows, err := tx.QueryContext(ctx, "SELECT identity,sequence,payload FROM research_log WHERE kind='admit' AND identity IN ("+marks+")", args...)
	if err != nil {
		return nil, err
	}
	type stored struct {
		seq     int64
		payload []byte
	}
	old := map[string]stored{}
	for rows.Next() {
		var id string
		var v stored
		if err = rows.Scan(&id, &v.seq, &v.payload); err != nil {
			rows.Close()
			return nil, err
		}
		old[id] = v
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	results := make([]AppendResult, len(r))
	first := map[string]int{}
	var fresh []int
	for i, id := range ids {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		if j, ok := first[id]; ok {
			if !bytes.Equal(r[i].Payload, r[j].Payload) {
				return nil, errors.New("conflicting ledger retry")
			}
			results[i].Retry = true
			continue
		}
		first[id] = i
		if v, ok := old[id]; ok {
			if !bytes.Equal(r[i].Payload, v.payload) {
				return nil, errors.New("conflicting ledger retry")
			}
			results[i] = AppendResult{v.seq, true}
		} else {
			fresh = append(fresh, i)
		}
	}
	// 128 rows cap each SQL statement at384 bound values; total stays <=512.
	for begin := 0; begin < len(fresh); begin += 128 {
		end := begin + 128
		if end > len(fresh) {
			end = len(fresh)
		}
		values := strings.TrimSuffix(strings.Repeat("(?,?,?),", end-begin), ",")
		args := make([]any, 0, 3*(end-begin))
		for _, i := range fresh[begin:end] {
			args = append(args, ids[i], "admit", []byte(r[i].Payload))
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO research_log(identity,kind,payload) VALUES "+values, args...); err != nil {
			return nil, err
		}
	}
	sequences := map[string]int64{}
	if len(fresh) > 0 {
		freshIDs := make([]string, len(fresh))
		for j, i := range fresh {
			freshIDs[j] = ids[i]
		}
		marks, args = identityArgs(freshIDs)
		rows, err = tx.QueryContext(ctx, "SELECT identity,sequence FROM research_log WHERE kind='admit' AND identity IN ("+marks+")", args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id string
			var seq int64
			if err = rows.Scan(&id, &seq); err != nil {
				rows.Close()
				return nil, err
			}
			sequences[id] = seq
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
		if len(sequences) != len(fresh) {
			return nil, errors.New("bulk receipt count mismatch")
		}
	}
	for i, id := range ids {
		if v, ok := old[id]; ok {
			results[i].Sequence = v.seq
		} else {
			results[i].Sequence = sequences[id]
		}
		if results[i].Sequence <= 0 {
			return nil, errors.New("invalid bulk receipt")
		}
	}
	if beforeCommit != nil {
		beforeCommit()
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return results, nil
}
