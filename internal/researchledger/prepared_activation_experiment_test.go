package researchledger

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// This test-only schema keeps prepared originals invisible until a separate,
// atomic source-index/acceptance transaction. No daemon opens these tables.
func activationTrialSchema(ctx context.Context, l *Ledger) error {
	for _, ddl := range []string{
		`CREATE TABLE activation_trial_prepare(id INTEGER PRIMARY KEY, identity TEXT NOT NULL, source TEXT NOT NULL, prediction_id INTEGER NOT NULL, version INTEGER NOT NULL, payload BLOB NOT NULL)`,
		`CREATE TABLE activation_trial_accept(source TEXT PRIMARY KEY, identity TEXT NOT NULL UNIQUE, prediction_id INTEGER NOT NULL UNIQUE, prepare_id INTEGER NOT NULL UNIQUE, version INTEGER NOT NULL)`,
	} {
		if _, err := l.db.ExecContext(ctx, ddl); err != nil {
			return err
		}
	}
	return nil
}

func activationTrialPrepare(ctx context.Context, l *Ledger, version int, requests []AppendRequest) ([]int64, error) {
	if len(requests) == 0 || len(requests) > 256 {
		return nil, errors.New("invalid prepare size")
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO activation_trial_prepare(identity,source,prediction_id,version,payload) VALUES(?,?,?,?,?)`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	ids := make([]int64, len(requests))
	seen := make(map[string]bool, len(requests))
	for i, r := range requests {
		identity, source, err := envelopeReferenceSource(r)
		if err != nil || seen[source] {
			return nil, errors.New("invalid or duplicate prepared source")
		}
		seen[source] = true
		var predictionID int64
		if _, err = fmt.Sscan(r.Key.Event, &predictionID); err != nil || predictionID <= 0 {
			return nil, errors.New("invalid prepared prediction ID")
		}
		result, err := stmt.ExecContext(ctx, identity, source, predictionID, version, []byte(r.Payload))
		if err != nil {
			return nil, err
		}
		ids[i], err = result.LastInsertId()
		if err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}

// Caller must hold the publication guard from version check through commit.
// The test invokes this as one guarded phase; SQL does not certify the store.
func activationTrialAccept(ctx context.Context, l *Ledger, version int, prepared []int64) error {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var next int64
	if err = tx.QueryRowContext(ctx, `SELECT coalesce(max(prediction_id),0) FROM activation_trial_accept`).Scan(&next); err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO activation_trial_accept(source,identity,prediction_id,prepare_id,version) VALUES(?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, id := range prepared {
		var identity, source string
		var predictionID int64
		var captured int
		var payload []byte
		err = tx.QueryRowContext(ctx, `SELECT identity,source,prediction_id,version,payload FROM activation_trial_prepare WHERE id=?`, id).
			Scan(&identity, &source, &predictionID, &captured, &payload)
		if err != nil {
			return err
		}
		if captured != version || predictionID != next+1 || len(payload) > 1<<20 || !json.Valid(payload) {
			return errors.New("prepared record stale, out of order, or invalid")
		}
		if _, err = stmt.ExecContext(ctx, source, identity, predictionID, id, version); err != nil {
			return err
		}
		next++
	}
	return tx.Commit()
}

func activationTrialGet(ctx context.Context, l *Ledger, source string) (int64, []byte, error) {
	var predictionID int64
	var payload []byte
	err := l.db.QueryRowContext(ctx, `SELECT a.prediction_id,p.payload FROM activation_trial_accept a LEFT JOIN activation_trial_prepare p ON p.id=a.prepare_id WHERE a.source=?`, source).
		Scan(&predictionID, &payload)
	if err != nil {
		return 0, nil, err
	}
	if predictionID <= 0 || len(payload) == 0 || len(payload) > 1<<20 || !json.Valid(payload) {
		return 0, nil, errors.New("accepted original missing or corrupt")
	}
	return predictionID, payload, nil
}

func activationTrialResolve(ctx context.Context, l *Ledger, request AppendRequest) (int64, error) {
	identity, source, err := envelopeReferenceSource(request)
	if err != nil {
		return 0, err
	}
	var storedIdentity string
	var payload []byte
	var predictionID int64
	err = l.db.QueryRowContext(ctx, `SELECT a.identity,a.prediction_id,p.payload FROM activation_trial_accept a LEFT JOIN activation_trial_prepare p ON p.id=a.prepare_id WHERE a.source=?`, source).
		Scan(&storedIdentity, &predictionID, &payload)
	if err != nil {
		return 0, err
	}
	if storedIdentity != identity || predictionID <= 0 || !bytes.Equal(payload, request.Payload) {
		return 0, errors.New("conflicting or corrupt accepted retry")
	}
	return predictionID, nil
}

func TestPreparedActivationSemantics(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/activation.sqlite"
	l, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = activationTrialSchema(ctx, l); err != nil {
		t.Fatal(err)
	}
	a := envelopeReferenceFixture(1, 1)
	ids, err := activationTrialPrepare(ctx, l, 7, []AppendRequest{a})
	if err != nil {
		t.Fatal(err)
	}
	_, sourceA, _ := envelopeReferenceSource(a)
	if _, _, err = activationTrialGet(ctx, l, sourceA); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("prepared original visible before acceptance: %v", err)
	}
	if err = l.Close(); err != nil {
		t.Fatal(err)
	}
	l, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if _, _, err = activationTrialGet(ctx, l, sourceA); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("orphaned prepare visible after reopen: %v", err)
	}
	if err = activationTrialAccept(ctx, l, 8, ids); err == nil {
		t.Fatal("accepted stale version")
	}
	if err = activationTrialAccept(ctx, l, 7, ids); err != nil {
		t.Fatal(err)
	}
	if got, payload, err := activationTrialGet(ctx, l, sourceA); err != nil || got != 1 || !bytes.Equal(payload, a.Payload) {
		t.Fatal("accepted original mismatch", got, err)
	}
	if err = l.Close(); err != nil {
		t.Fatal(err)
	}
	l, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, payload, err := activationTrialGet(ctx, l, sourceA); err != nil || got != 1 || !bytes.Equal(payload, a.Payload) {
		t.Fatal("uncertain acknowledgment did not replay exact original", got, err)
	}
	if got, err := activationTrialResolve(ctx, l, a); err != nil || got != 1 {
		t.Fatal("exact retry did not resolve original", got, err)
	}
	changed := a
	changed.Payload = json.RawMessage(`{"Binding":{"Tenant":"public","JournalID":"public-session","EventID":"1"},"Value":"changed"}`)
	if _, err := activationTrialResolve(ctx, l, changed); err == nil {
		t.Fatal("changed retry resolved old original")
	}
	b := envelopeReferenceFixture(2, 2)
	bids, err := activationTrialPrepare(ctx, l, 7, []AppendRequest{b})
	if err != nil {
		t.Fatal(err)
	}
	if err = activationTrialAccept(ctx, l, 7, bids); err != nil {
		t.Fatal(err)
	}
	conflict := envelopeReferenceFixture(4, 2)
	c := envelopeReferenceFixture(3, 3)
	badIDs, err := activationTrialPrepare(ctx, l, 7, []AppendRequest{c, conflict})
	if err != nil {
		t.Fatal(err)
	}
	if err = activationTrialAccept(ctx, l, 7, badIDs); err == nil {
		t.Fatal("accepted duplicate source")
	}
	_, sourceC, _ := envelopeReferenceSource(c)
	if _, _, err = activationTrialGet(ctx, l, sourceC); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("late conflict partially accepted batch", err)
	}
}

type activationTrialResult struct {
	Trial, Size, Verified int
	Staged                bool
	PrepareNS, GuardNS    []int64
}

func TestPreparedActivationExperiment(t *testing.T) {
	artifact := os.Getenv("EVENTFRAME_PREPARED_ACTIVATION_ARTIFACT")
	if artifact == "" {
		t.Skip("opt-in prepared activation experiment")
	}
	f, err := os.OpenFile(artifact, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ctx := context.Background()
	enc := json.NewEncoder(f)
	if err = enc.Encode(map[string]any{"kind": "header", "scope": "test-only SQLite WAL/FULL, no service guard", "sizes": []int{50, 200}, "batches_per_cell": 32}); err != nil {
		t.Fatal(err)
	}
	for trial := 0; trial < 3; trial++ {
		for _, size := range []int{50, 200} {
			for order := 0; order < 2; order++ {
				staged := (trial+order)%2 == 1
				path := t.TempDir() + "/activation.sqlite"
				l, err := Open(path)
				if err != nil {
					t.Fatal(err)
				}
				if staged {
					err = activationTrialSchema(ctx, l)
				} else {
					err = l.EnableServiceIdentity(ctx)
				}
				if err != nil {
					t.Fatal(err)
				}
				var synchronous int
				var journal string
				if err = l.db.QueryRow("PRAGMA synchronous").Scan(&synchronous); err != nil || synchronous != 2 {
					t.Fatal("not FULL", synchronous, err)
				}
				if err = l.db.QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil || journal != "wal" {
					t.Fatal("not WAL", journal, err)
				}
				r := activationTrialResult{Trial: trial, Size: size, Staged: staged}
				for batch := 0; batch < 32; batch++ {
					requests := make([]AppendRequest, size)
					for i := range requests {
						requests[i] = envelopeReferenceFixture(batch*size+i+1, batch*size+i+1)
						var body map[string]any
						if err = json.Unmarshal(requests[i].Payload, &body); err != nil {
							t.Fatal(err)
						}
						body["Padding"] = strings.Repeat("x", 1024)
						requests[i].Payload, err = json.Marshal(body)
						if err != nil {
							t.Fatal(err)
						}
					}
					if staged {
						start := time.Now()
						ids, prepareErr := activationTrialPrepare(ctx, l, 7, requests)
						r.PrepareNS = append(r.PrepareNS, time.Since(start).Nanoseconds())
						if prepareErr != nil {
							t.Fatal(prepareErr)
						}
						start = time.Now()
						err = activationTrialAccept(ctx, l, 7, ids)
						r.GuardNS = append(r.GuardNS, time.Since(start).Nanoseconds())
					} else {
						start := time.Now()
						_, err = l.AppendBatchPrepared(ctx, requests)
						r.GuardNS = append(r.GuardNS, time.Since(start).Nanoseconds())
					}
					if err != nil {
						t.Fatal(err)
					}
				}
				if err = l.Close(); err != nil {
					t.Fatal(err)
				}
				l, err = Open(path)
				if err != nil {
					t.Fatal(err)
				}
				for i := 1; i <= size*32; i++ {
					want := envelopeReferenceFixture(i, i)
					var body map[string]any
					if err = json.Unmarshal(want.Payload, &body); err != nil {
						t.Fatal(err)
					}
					body["Padding"] = strings.Repeat("x", 1024)
					want.Payload, err = json.Marshal(body)
					if err != nil {
						t.Fatal(err)
					}
					if staged {
						_, source, _ := envelopeReferenceSource(want)
						got, payload, getErr := activationTrialGet(ctx, l, source)
						if getErr != nil || got != int64(i) || !bytes.Equal(payload, want.Payload) {
							t.Fatal("staged original mismatch", i, getErr)
						}
					} else {
						got, getErr := l.GetServiceAdmission(ctx, ServiceIdentity{"public", "stream", "reference-v1", "public-session", fmt.Sprint(i)})
						if getErr != nil || got.Sequence != int64(i) || !bytes.Equal(got.Payload, want.Payload) {
							t.Fatal("control original mismatch", i, getErr)
						}
					}
					r.Verified++
				}
				if err = l.Close(); err != nil {
					t.Fatal(err)
				}
				if err = enc.Encode(r); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
