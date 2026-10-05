package researchledger

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// Admission-only, test-owned prototype. Raw SQL is not an authorized writer:
// source-to-payload consistency depends on this private preparation boundary.
func materializedSourceSchema(l *Ledger) error {
	_, err := l.db.Exec(`CREATE TABLE materialized_source_trial(sequence INTEGER PRIMARY KEY AUTOINCREMENT,identity TEXT NOT NULL UNIQUE,source TEXT NOT NULL UNIQUE,payload BLOB NOT NULL)`)
	return err
}

func materializedSourceAppend(ctx context.Context, l *Ledger, requests []AppendRequest, beforeCommit func()) ([]AppendResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(requests) == 0 || len(requests) > MaxBatchEntries {
		return nil, errors.New("invalid batch count")
	}
	identities, sources := make([]string, len(requests)), make([]string, len(requests))
	total := 0
	for i, r := range requests {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		source, err := canonicalSourceKey(r)
		if err != nil {
			return nil, err
		}
		id, err := json.Marshal(r.Key)
		if err != nil {
			return nil, err
		}
		total += len(id) + len(r.Kind) + len(r.Payload)
		if total > MaxBatchBytes {
			return nil, errors.New("batch byte cap")
		}
		identities[i], sources[i] = string(id), source
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	lookup, err := tx.PrepareContext(ctx, `SELECT sequence,source,payload FROM materialized_source_trial WHERE identity=?`)
	if err != nil {
		return nil, err
	}
	defer lookup.Close()
	insert, err := tx.PrepareContext(ctx, `INSERT INTO materialized_source_trial(identity,source,payload) VALUES(?,?,?)`)
	if err != nil {
		return nil, err
	}
	defer insert.Close()
	results := make([]AppendResult, len(requests))
	for i, r := range requests {
		var seq int64
		var oldSource string
		var old []byte
		err := lookup.QueryRowContext(ctx, identities[i]).Scan(&seq, &oldSource, &old)
		if err == nil {
			if oldSource != sources[i] || !bytes.Equal(old, r.Payload) {
				return nil, errors.New("conflicting retry")
			}
			results[i] = AppendResult{Sequence: seq, Retry: true}
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		result, err := insert.ExecContext(ctx, identities[i], sources[i], []byte(r.Payload))
		if err != nil {
			return nil, err
		}
		seq, err = result.LastInsertId()
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

func materializedSourceRead(ctx context.Context, l *Ledger, source string) (Entry, error) {
	var e Entry
	var identity string
	err := l.db.QueryRowContext(ctx, `SELECT sequence,identity,payload FROM materialized_source_trial WHERE source=?`, source).Scan(&e.Sequence, &identity, &e.Payload)
	if err != nil {
		return e, err
	}
	if err = json.Unmarshal([]byte(identity), &e.Key); err != nil {
		return e, err
	}
	e.Kind = "admit"
	derived, err := canonicalSourceKey(AppendRequest{Key: e.Key, Kind: e.Kind, Payload: e.Payload})
	if err != nil || derived != source {
		return Entry{}, errors.New("source payload mismatch")
	}
	return e, nil
}

func TestMaterializedSourceSemantics(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/materialized.sqlite"
	l, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { l.Close() }()
	if err := materializedSourceSchema(l); err != nil {
		t.Fatal(err)
	}
	a := serviceEntry("1", ServiceIdentity{"tenant", "stream", "contract", "alpha", "event"})
	acks, err := materializedSourceAppend(ctx, l, []AppendRequest{a}, nil)
	if err != nil || len(acks) != 1 || acks[0].Retry {
		t.Fatal("first", err)
	}
	first := acks[0].Sequence
	for _, mode := range []string{"new key same source", "equivalent escape", "same key changed bytes"} {
		b := a
		if mode != "same key changed bytes" {
			b.Key.Event = "2"
		}
		if mode != "new key same source" {
			b.Payload = []byte(strings.Replace(string(b.Payload), `"alpha"`, `"\u0061lpha"`, 1))
		}
		if _, err := materializedSourceAppend(ctx, l, []AppendRequest{phaseFixture(3, 1)[0], b}, nil); err == nil {
			t.Fatal("accepted", mode)
		}
		var count int
		if err := l.db.QueryRow(`SELECT count(*) FROM materialized_source_trial`).Scan(&count); err != nil || count != 1 {
			t.Fatal("partial commit", err, count)
		}
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	_, err = materializedSourceAppend(cancelCtx, l, phaseFixture(3, 1), cancel)
	cancel()
	if err == nil {
		t.Fatal("canceled commit")
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	l, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	acks, err = materializedSourceAppend(ctx, l, []AppendRequest{a}, nil)
	if err != nil || len(acks) != 1 || !acks[0].Retry || acks[0].Sequence != first {
		t.Fatal("retry after reopen", err)
	}
	source, err := canonicalSourceKey(a)
	if err != nil {
		t.Fatal(err)
	}
	e, err := materializedSourceRead(ctx, l, source)
	if err != nil || e.Key != a.Key || !bytes.Equal(e.Payload, a.Payload) {
		t.Fatal("read", err)
	}
	if _, err := l.db.Exec(`UPDATE materialized_source_trial SET source='wrong'`); err != nil {
		t.Fatal(err)
	}
	if _, err := materializedSourceRead(ctx, l, "wrong"); err == nil {
		t.Fatal("unbound source accepted")
	}
}
