package researchledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestReadBatchOrderMissingAndBounds(t *testing.T) {
	ctx := context.Background()
	l, err := Open(t.TempDir() + "/reads.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	rows := batchFixture()
	if _, err = l.AppendBatch(ctx, rows); err != nil {
		t.Fatal(err)
	}
	missing := rows[0].Key
	missing.Contract = "other"
	requests := []LookupRequest{{rows[2].Key, "admit"}, {missing, "admit"}, {rows[0].Key, "feedback"}, {rows[2].Key, "admit"}}
	got, err := l.GetBatch(ctx, requests)
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range requests {
		want, err := l.Get(ctx, r.Key, r.Kind)
		if errors.Is(err, sql.ErrNoRows) {
			if got[i].Found || got[i].Entry.Key != r.Key {
				t.Fatal("missing mismatch")
			}
		} else if err != nil || !got[i].Found || !reflect.DeepEqual(got[i].Entry, want) {
			t.Fatal("single parity", i, err)
		}
	}
	got[0].Entry.Payload[0] = 'x'
	again, err := l.Get(ctx, requests[0].Key, requests[0].Kind)
	if err != nil || again.Payload[0] == 'x' || got[3].Entry.Payload[0] == 'x' {
		t.Fatal("payload alias", err)
	}
	for _, bad := range [][]LookupRequest{nil, make([]LookupRequest, 513), {{Key: missing, Kind: "unknown"}}, {{Kind: "admit"}}} {
		if r, err := l.GetBatch(ctx, bad); err == nil || r != nil {
			t.Fatal("invalid input accepted")
		}
	}
	key := Key{"tenant", "large", "1", "v1"}
	payload := json.RawMessage(`"` + strings.Repeat("x", (1<<20)-2) + `"`)
	if _, _, err = l.Append(ctx, key, "admit", payload); err != nil {
		t.Fatal(err)
	}
	large := make([]LookupRequest, 8)
	for i := range large {
		large[i] = LookupRequest{key, "admit"}
	}
	if _, err = l.GetBatch(ctx, large); err != nil {
		t.Fatal("exact payload cap rejected", err)
	}
	if r, err := l.GetBatch(ctx, append(large, large[0])); err == nil || r != nil {
		t.Fatal("aggregate payload overflow accepted")
	}
	full := make([]LookupRequest, 512)
	for i := range full {
		full[i] = LookupRequest{missing, "admit"}
	}
	if r, err := l.GetBatch(ctx, full); err != nil || len(r) != 512 {
		t.Fatal("count boundary", err)
	}
}

func TestReadBatchTransactionSnapshot(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/snapshot.sqlite"
	l, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	first, late := Key{"tenant", "s", "first", "v1"}, Key{"tenant", "s", "late", "v1"}
	if _, _, err = l.Append(ctx, first, "admit", json.RawMessage(`true`)); err != nil {
		t.Fatal(err)
	}
	u := url.URL{Scheme: "file", Path: path}
	other, err := sql.Open("sqlite", u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	requests := []LookupRequest{{first, "admit"}, {late, "admit"}}
	got, err := l.getBatch(ctx, requests, func(i int) {
		if i == 0 {
			key, _ := json.Marshal(late)
			if _, err := other.ExecContext(ctx, "INSERT INTO research_log(identity,kind,payload) VALUES(?,?,?)", string(key), "admit", []byte(`false`)); err != nil {
				t.Fatal(err)
			}
		}
	})
	if err != nil || !got[0].Found || got[1].Found {
		t.Fatal("mixed transaction snapshots", got, err)
	}
	got, err = l.GetBatch(ctx, requests)
	if err != nil || !got[1].Found {
		t.Fatal("new transaction missed committed row", err)
	}
}

func TestReadBatchCancellationAndCorruption(t *testing.T) {
	l, err := Open(t.TempDir() + "/cancel.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	ctx := context.Background()
	rows := batchFixture()
	if _, err = l.AppendBatch(ctx, rows); err != nil {
		t.Fatal(err)
	}
	requests := []LookupRequest{{rows[0].Key, "admit"}, {rows[2].Key, "admit"}}
	c, cancel := context.WithCancel(ctx)
	if r, err := l.getBatch(c, requests, func(int) { cancel() }); !errors.Is(err, context.Canceled) || r != nil {
		t.Fatal("partial canceled result", err)
	}
	if _, err = l.GetBatch(ctx, requests); err != nil {
		t.Fatal("canceled read retained transaction", err)
	}
	key, _ := json.Marshal(rows[2].Key)
	if _, err = l.db.Exec("UPDATE research_log SET payload=zeroblob(?) WHERE identity=? AND kind='admit'", (1<<20)+1, string(key)); err != nil {
		t.Fatal(err)
	}
	if r, err := l.GetBatch(ctx, requests); err == nil || r != nil {
		t.Fatal("oversized stored blob materialized as valid")
	}
}

func TestReadBatchRequestCapPanicAndStorageType(t *testing.T) {
	l, err := Open(t.TempDir() + "/edge.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	ctx := context.Background()
	rows := batchFixture()
	if _, err = l.AppendBatch(ctx, rows); err != nil {
		t.Fatal(err)
	}
	request := LookupRequest{rows[0].Key, "admit"}
	oversized := make([]LookupRequest, 512)
	field := strings.Repeat("\x00", 4096)
	for i := range oversized {
		oversized[i] = LookupRequest{Key{field, field, field, field}, "admit"}
	}
	if r, err := l.GetBatch(ctx, oversized); err == nil || r != nil || !strings.Contains(err.Error(), "request byte cap") {
		t.Fatal("encoded request bound", err)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("missing read hook panic")
			}
		}()
		_, _ = l.getBatch(ctx, []LookupRequest{request}, func(int) { panic("read boundary") })
	}()
	if _, err = l.GetBatch(ctx, []LookupRequest{request}); err != nil {
		t.Fatal("panic retained read transaction", err)
	}
	key, _ := json.Marshal(request.Key)
	if _, err = l.db.Exec("UPDATE research_log SET payload='text-not-blob' WHERE identity=? AND kind='admit'", string(key)); err != nil {
		t.Fatal(err)
	}
	if r, err := l.GetBatch(ctx, []LookupRequest{request}); err == nil || r != nil {
		t.Fatal("invalid stored type accepted")
	}
	c, cancel := context.WithCancel(ctx)
	cancel()
	if r, err := l.GetBatch(c, []LookupRequest{request}); !errors.Is(err, context.Canceled) || r != nil {
		t.Fatal("pre-cancel ignored", err)
	}
}
