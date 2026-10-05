package researchledger

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

type envelopeReadStats struct{ Bodies, Slices, BodyBytes int }

// Cache lifetime is one transaction. Once its byte allowance is exhausted,
// slice reads preserve support for dispersed requests rather than rejecting
// otherwise valid batches. Output copies have their own aggregate byte cap.
func envelopeReferenceBatch(ctx context.Context, l *Ledger, requests []ServiceIdentity, cacheCap int) ([]ServiceLookupResult, envelopeReadStats, error) {
	var stats envelopeReadStats
	if cacheCap < 0 || cacheCap > MaxBatchBytes || len(requests) == 0 || len(requests) > MaxBatchEntries {
		return nil, stats, errors.New("invalid bounds")
	}
	keys := make([]string, len(requests))
	inputBytes := 0
	for i, r := range requests {
		for _, s := range []string{r.Tenant, r.Stream, r.Contract, r.Journal, r.Event} {
			if len(s) == 0 || len(s) > 4096 || !utf8.ValidString(s) {
				return nil, stats, errors.New("invalid source")
			}
		}
		b, err := json.Marshal(r)
		if err != nil {
			return nil, stats, err
		}
		keys[i] = string(b)
		inputBytes += len(b)
		if inputBytes > MaxBatchBytes {
			return nil, stats, errors.New("input cap")
		}
	}
	tx, err := l.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, stats, err
	}
	defer tx.Rollback()
	meta, err := tx.PrepareContext(ctx, `SELECT sequence,CASE WHEN length(CAST(identity AS BLOB))<=131072 THEN identity ELSE NULL END,envelope,start,size FROM envelope_trial_ref WHERE source=?`)
	if err != nil {
		return nil, stats, err
	}
	defer meta.Close()
	bodyStmt, err := tx.PrepareContext(ctx, `SELECT length(body),typeof(body),CASE WHEN typeof(body)='blob' AND length(body)<=? THEN body ELSE NULL END FROM envelope_trial_body WHERE id=?`)
	if err != nil {
		return nil, stats, err
	}
	defer bodyStmt.Close()
	sliceStmt, err := tx.PrepareContext(ctx, `SELECT CASE WHEN typeof(body)='blob' AND ?>=0 AND ?>0 AND ?<=1048576 AND ?<=length(body)-? THEN substr(body,?+1,?) ELSE NULL END FROM envelope_trial_body WHERE id=?`)
	if err != nil {
		return nil, stats, err
	}
	defer sliceStmt.Close()
	cache := map[int64][]byte{}
	out := make([]ServiceLookupResult, len(requests))
	remaining := MaxBatchBytes
	for i, r := range requests {
		if err := ctx.Err(); err != nil {
			return nil, stats, err
		}
		out[i].Source = r
		var seq, env, start, size int64
		var identity sql.NullString
		err := meta.QueryRowContext(ctx, keys[i]).Scan(&seq, &identity, &env, &start, &size)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, stats, err
		}
		if seq <= 0 || env <= 0 || !identity.Valid || start < 0 || size <= 0 || size > 1<<20 || size > int64(remaining) {
			return nil, stats, errors.New("invalid reference")
		}
		var payload []byte
		body, cached := cache[env]
		if !cached {
			var length int64
			var kind string
			if err := bodyStmt.QueryRowContext(ctx, cacheCap-stats.BodyBytes, env).Scan(&length, &kind, &body); err != nil {
				return nil, stats, err
			}
			if kind != "blob" || length <= 0 || length > MaxBatchBytes || start > length-size {
				return nil, stats, errors.New("invalid body")
			}
			if body != nil {
				cache[env] = body
				stats.BodyBytes += len(body)
				stats.Bodies++
			}
		}
		if body != nil {
			if start > int64(len(body))-size {
				return nil, stats, errors.New("invalid offset")
			}
			payload = append([]byte(nil), body[start:start+size]...)
		} else {
			if err := sliceStmt.QueryRowContext(ctx, start, size, size, start, size, start, size, env).Scan(&payload); err != nil {
				return nil, stats, err
			}
			stats.Slices++
		}
		if int64(len(payload)) != size {
			return nil, stats, errors.New("invalid slice")
		}
		var k Key
		if err := json.Unmarshal([]byte(identity.String), &k); err != nil {
			return nil, stats, err
		}
		id, source, err := envelopeReferenceSource(AppendRequest{k, "admit", payload})
		if err != nil || id != identity.String || source != keys[i] {
			return nil, stats, errors.New("source/payload mismatch")
		}
		remaining -= len(payload)
		out[i] = ServiceLookupResult{Source: r, Found: true, Entry: Entry{Sequence: seq, Kind: "admit", Key: k, Payload: payload}}
	}
	if err := tx.Commit(); err != nil {
		return nil, stats, err
	}
	return out, stats, nil
}

func TestEnvelopeBatchRead(t *testing.T) {
	ctx := context.Background()
	l, err := Open(t.TempDir() + "/batch.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if err := envelopeReferenceSchema(ctx, l); err != nil {
		t.Fatal(err)
	}
	var originals []AppendRequest
	var sources []ServiceIdentity
	for i := 1; i <= 4; i++ {
		r := envelopeReferenceFixture(i, i)
		originals = append(originals, r)
		_, s, _ := envelopeReferenceSource(r)
		var source ServiceIdentity
		json.Unmarshal([]byte(s), &source)
		sources = append(sources, source)
	}
	if _, err := envelopeReferenceAppend(ctx, l, originals[:2], nil); err != nil {
		t.Fatal(err)
	}
	if _, err := envelopeReferenceAppend(ctx, l, originals[2:], nil); err != nil {
		t.Fatal(err)
	}
	missing := sources[0]
	missing.Event = "missing"
	requests := []ServiceIdentity{sources[3], sources[0], missing, sources[3], sources[1]}
	for _, cap := range []int{0, MaxBatchBytes} {
		t.Run(fmt.Sprint(cap), func(t *testing.T) {
			got, stats, err := envelopeReferenceBatch(ctx, l, requests, cap)
			if err != nil || len(got) != 5 {
				t.Fatal(got, err)
			}
			for i, index := range []int{3, 0, -1, 3, 1} {
				if got[i].Source != requests[i] {
					t.Fatal("order")
				}
				if index < 0 {
					if got[i].Found || got[i].Entry.Sequence != 0 {
						t.Fatal("miss")
					}
					continue
				}
				if !got[i].Found || got[i].Entry.Sequence != int64(index+1) || !bytes.Equal(got[i].Entry.Payload, originals[index].Payload) {
					t.Fatal("original")
				}
			}
			got[0].Entry.Payload[0] = '!'
			if got[3].Entry.Payload[0] == '!' {
				t.Fatal("aliased duplicate output")
			}
			if cap == 0 && (stats.Bodies != 0 || stats.Slices != 4) {
				t.Fatal(stats)
			}
			if cap > 0 && (stats.Bodies != 2 || stats.Slices != 0) {
				t.Fatal(stats)
			}
		})
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if out, _, err := envelopeReferenceBatch(cancelled, l, requests, MaxBatchBytes); err == nil || out != nil {
		t.Fatal("cancel", err)
	}
	// Corrupt a valid-range reference to another original: structural bounds
	// alone must not let the wrong source escape.
	if _, err := l.db.Exec(`UPDATE envelope_trial_ref SET start=0,size=? WHERE sequence=2`, len(originals[0].Payload)); err != nil {
		t.Fatal(err)
	}
	if out, _, err := envelopeReferenceBatch(ctx, l, []ServiceIdentity{sources[0], sources[1]}, MaxBatchBytes); err == nil || out != nil {
		t.Fatal("wrong-source slice escaped")
	}
}

func TestEnvelopeBatchOutputCap(t *testing.T) {
	ctx := context.Background()
	l, err := Open(t.TempDir() + "/cap.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if err := envelopeReferenceSchema(ctx, l); err != nil {
		t.Fatal(err)
	}
	r := envelopeReferenceFixture(1, 1)
	var body map[string]any
	if err := json.Unmarshal(r.Payload, &body); err != nil {
		t.Fatal(err)
	}
	body["Padding"] = strings.Repeat("x", 800000)
	r.Payload, err = json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := envelopeReferenceAppend(ctx, l, []AppendRequest{r}, nil); err != nil {
		t.Fatal(err)
	}
	_, source, err := envelopeReferenceSource(r)
	if err != nil {
		t.Fatal(err)
	}
	var key ServiceIdentity
	if err := json.Unmarshal([]byte(source), &key); err != nil {
		t.Fatal(err)
	}
	requests := make([]ServiceIdentity, 12)
	for i := range requests {
		requests[i] = key
	}
	for _, cap := range []int{0, MaxBatchBytes} {
		out, stats, err := envelopeReferenceBatch(ctx, l, requests, cap)
		if err == nil || out != nil || stats.BodyBytes > cap {
			t.Fatal("output cap bypass", stats, err)
		}
	}
}
