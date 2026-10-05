package researchledger

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"unicode/utf8"
)

// Test-only normalized envelope prototype. No daemon constructor installs this
// schema. Every original keeps its exact key, source, bytes and stable sequence.
func envelopeReferenceSchema(ctx context.Context, l *Ledger) error {
	for _, ddl := range []string{
		`CREATE TABLE envelope_trial_body(id INTEGER PRIMARY KEY, body BLOB NOT NULL)`,
		`CREATE TABLE envelope_trial_ref(sequence INTEGER PRIMARY KEY AUTOINCREMENT, identity TEXT NOT NULL UNIQUE, source TEXT NOT NULL UNIQUE, envelope INTEGER NOT NULL, start INTEGER NOT NULL CHECK(start>=0), size INTEGER NOT NULL CHECK(size>0 AND size<=1048576))`,
	} {
		if _, err := l.db.ExecContext(ctx, ddl); err != nil {
			return err
		}
	}
	return nil
}

func envelopeReferenceSource(r AppendRequest) (string, string, error) {
	if r.Kind != "admit" || len(r.Payload) == 0 || len(r.Payload) > 1<<20 || !json.Valid(r.Payload) {
		return "", "", errors.New("invalid original")
	}
	var b struct {
		Binding struct{ Tenant, JournalID, EventID string }
	}
	if err := json.Unmarshal(r.Payload, &b); err != nil {
		return "", "", err
	}
	if b.Binding.Tenant != r.Key.Tenant {
		return "", "", errors.New("tenant mismatch")
	}
	for _, v := range []string{r.Key.Tenant, r.Key.Journal, r.Key.Event, r.Key.Contract, b.Binding.JournalID, b.Binding.EventID} {
		if len(v) == 0 || len(v) > 4096 || !utf8.ValidString(v) {
			return "", "", errors.New("invalid identity")
		}
	}
	id, err := json.Marshal(r.Key)
	if err != nil {
		return "", "", err
	}
	source, err := json.Marshal(ServiceIdentity{r.Key.Tenant, r.Key.Journal, r.Key.Contract, b.Binding.JournalID, b.Binding.EventID})
	return string(id), string(source), err
}

// A left join exposes missing bodies as corruption, never as an absent key.
// Slice bounds are checked before materializing at most one original payload.
const envelopeReferenceLookup = `SELECT r.sequence,r.source,r.start,r.size,length(b.body),
CASE WHEN r.start>=0 AND r.size>0 AND r.size<=1048576 AND r.start+r.size<=length(b.body)
THEN substr(b.body,r.start+1,r.size) ELSE NULL END
FROM envelope_trial_ref r LEFT JOIN envelope_trial_body b ON b.id=r.envelope WHERE r.identity=?`

func envelopeReferenceRead(row *sql.Row) (int64, string, []byte, error) {
	var seq, start, size int64
	var length sql.NullInt64
	var source string
	var payload []byte
	err := row.Scan(&seq, &source, &start, &size, &length, &payload)
	if err != nil {
		return 0, "", nil, err
	}
	if seq <= 0 || !length.Valid || start < 0 || size <= 0 || size > 1<<20 || start > length.Int64-size || int64(len(payload)) != size {
		return 0, "", nil, errors.New("corrupt envelope reference")
	}
	return seq, source, payload, nil
}

func envelopeReferenceAppend(ctx context.Context, l *Ledger, requests []AppendRequest, beforeCommit func() error) ([]AppendResult, error) {
	if len(requests) == 0 || len(requests) > MaxBatchEntries {
		return nil, errors.New("invalid count")
	}
	ids, sources := make([]string, len(requests)), make([]string, len(requests))
	seenID, seenSource := map[string]bool{}, map[string]bool{}
	total := 0
	for i, r := range requests {
		var err error
		ids[i], sources[i], err = envelopeReferenceSource(r)
		if err != nil {
			return nil, err
		}
		if seenID[ids[i]] || seenSource[sources[i]] {
			return nil, errors.New("duplicate batch source")
		}
		seenID[ids[i]], seenSource[sources[i]] = true, true
		total += len(ids[i]) + len(sources[i]) + len(r.Payload)
		if total > MaxBatchBytes {
			return nil, errors.New("batch too large")
		}
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	lookup, err := tx.PrepareContext(ctx, envelopeReferenceLookup)
	if err != nil {
		return nil, err
	}
	defer lookup.Close()
	out := make([]AppendResult, len(requests))
	var body []byte
	starts := make([]int, len(requests))
	for i, r := range requests {
		seq, source, payload, e := envelopeReferenceRead(lookup.QueryRowContext(ctx, ids[i]))
		if e == nil {
			if source != sources[i] || !bytes.Equal(payload, r.Payload) {
				return nil, errors.New("conflicting retry")
			}
			out[i] = AppendResult{Sequence: seq, Retry: true}
			continue
		}
		if !errors.Is(e, sql.ErrNoRows) {
			return nil, e
		}
		starts[i] = len(body)
		body = append(body, r.Payload...)
	}
	if len(body) > 0 {
		result, e := tx.ExecContext(ctx, `INSERT INTO envelope_trial_body(body) VALUES(?)`, body)
		if e != nil {
			return nil, e
		}
		envelope, e := result.LastInsertId()
		if e != nil {
			return nil, e
		}
		insert, e := tx.PrepareContext(ctx, `INSERT INTO envelope_trial_ref(identity,source,envelope,start,size) VALUES(?,?,?,?,?)`)
		if e != nil {
			return nil, e
		}
		defer insert.Close()
		for i, r := range requests {
			if out[i].Retry {
				continue
			}
			result, e := insert.ExecContext(ctx, ids[i], sources[i], envelope, starts[i], len(r.Payload))
			if e != nil {
				return nil, e
			}
			seq, e := result.LastInsertId()
			if e != nil {
				return nil, e
			}
			out[i] = AppendResult{Sequence: seq}
		}
	}
	if beforeCommit != nil {
		if err := beforeCommit(); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func envelopeReferenceFixture(id int, source int) AppendRequest {
	raw, _ := json.Marshal(struct {
		Binding struct{ Tenant, JournalID, EventID string }
		Value   string
	}{
		Binding: struct{ Tenant, JournalID, EventID string }{"public", "public-session", fmt.Sprint(source)}, Value: fmt.Sprintf("public fact %d", id),
	})
	return AppendRequest{Key{"public", "stream", fmt.Sprint(id), "reference-v1"}, "admit", raw}
}

func TestEnvelopeReferenceSemantics(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/reference.sqlite"
	l, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if err := envelopeReferenceSchema(ctx, l); err != nil {
		t.Fatal(err)
	}
	appendBatch := func(rs []AppendRequest) ([]AppendResult, error) { return envelopeReferenceAppend(ctx, l, rs, nil) }
	a, b, c := envelopeReferenceFixture(1, 1), envelopeReferenceFixture(2, 2), envelopeReferenceFixture(3, 3)
	acks, err := appendBatch([]AppendRequest{a, b})
	if err != nil || len(acks) != 2 {
		t.Fatal(acks, err)
	}
	first := acks[0].Sequence
	acks, err = appendBatch([]AppendRequest{a, c})
	if err != nil || !acks[0].Retry || acks[0].Sequence != first || acks[1].Retry {
		t.Fatal("mixed retry", acks, err)
	}
	for name, rs := range map[string][]AppendRequest{
		"duplicate within":         {a, a},
		"same source different ID": {envelopeReferenceFixture(4, 1)},
		"late conflict rollback":   {envelopeReferenceFixture(4, 4), envelopeReferenceFixture(5, 2)},
		"changed bytes":            {{a.Key, a.Kind, json.RawMessage(`{"Binding":{"Tenant":"public","JournalID":"public-session","EventID":"1"},"Value":"changed"}`)}},
		"late invalid":             {envelopeReferenceFixture(4, 4), {Key{}, "admit", json.RawMessage(`{}`)}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := appendBatch(rs); err == nil {
				t.Fatal("accepted invalid batch")
			}
		})
	}
	fault := errors.New("rollback fault")
	if _, err := envelopeReferenceAppend(ctx, l, []AppendRequest{envelopeReferenceFixture(4, 4)}, func() error { return fault }); !errors.Is(err, fault) {
		t.Fatal(err)
	}
	var count int
	if err := l.db.QueryRow(`SELECT count(*) FROM envelope_trial_ref`).Scan(&count); err != nil || count != 3 {
		t.Fatal("partial refs", count, err)
	}
	if err := l.db.QueryRow(`SELECT count(*) FROM envelope_trial_body`).Scan(&count); err != nil || count != 2 {
		t.Fatal("orphan body", count, err)
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	l, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	for _, r := range []AppendRequest{a, b, c} {
		id, source, err := envelopeReferenceSource(r)
		if err != nil {
			t.Fatal(err)
		}
		_, got, payload, err := envelopeReferenceRead(l.db.QueryRowContext(ctx, envelopeReferenceLookup, id))
		if err != nil || got != source || !bytes.Equal(payload, r.Payload) {
			t.Fatal("reopen", err)
		}
	}
	// Acknowledgment can be lost after a successful commit: an exact retry
	// must return the persisted sequence rather than append a new original.
	acks, err = appendBatch([]AppendRequest{a})
	if err != nil || !acks[0].Retry || acks[0].Sequence != first {
		t.Fatal("lost ack", acks, err)
	}
	if _, err := l.db.Exec(`UPDATE envelope_trial_ref SET start=999999 WHERE sequence=?`, first); err != nil {
		t.Fatal(err)
	}
	if _, err := appendBatch([]AppendRequest{a}); err == nil {
		t.Fatal("corrupt offset accepted")
	}
	if _, err := l.db.Exec(`DELETE FROM envelope_trial_body`); err != nil {
		t.Fatal(err)
	}
	if _, err := appendBatch([]AppendRequest{c}); err == nil {
		t.Fatal("missing body accepted")
	}
}
