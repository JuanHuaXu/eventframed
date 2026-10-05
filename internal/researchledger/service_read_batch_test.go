package researchledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func sourceReadFixture(t *testing.T) (*Ledger, string) {
	t.Helper()
	path := t.TempDir() + "/sources.sqlite"
	l, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	if err = l.EnableServiceIdentity(context.Background()); err != nil {
		t.Fatal(err)
	}
	return l, path
}

func TestServiceReadBatchScopeOrderAndPlan(t *testing.T) {
	l, _ := sourceReadFixture(t)
	ctx := context.Background()
	base := sourceFixture()
	keys := []ServiceIdentity{base}
	for i := 0; i < 5; i++ {
		k := base
		switch i {
		case 0:
			k.Tenant = "other"
		case 1:
			k.Stream = "other"
		case 2:
			k.Contract = "other"
		case 3:
			k.Journal = "other"
		case 4:
			k.Event = "other"
		}
		keys = append(keys, k)
	}
	var entries []AppendRequest
	for i, k := range keys {
		entries = append(entries, serviceEntry(fmt.Sprint(i+1), k))
	}
	if _, err := l.AppendBatchPrepared(ctx, entries); err != nil {
		t.Fatal(err)
	}
	missing := base
	missing.Event = "missing"
	requests := append([]ServiceIdentity{missing, base}, keys...)
	got, err := l.GetServiceAdmissions(ctx, requests)
	if err != nil || len(got) != len(requests) {
		t.Fatal(got, err)
	}
	for i, k := range requests {
		want, e := l.GetServiceAdmission(ctx, k)
		if got[i].Source != k {
			t.Fatal("unbound result")
		}
		if errors.Is(e, sql.ErrNoRows) {
			if got[i].Found || !reflect.DeepEqual(got[i].Entry, Entry{}) {
				t.Fatal("malformed miss")
			}
		} else if e != nil || !got[i].Found || !reflect.DeepEqual(got[i].Entry, want) {
			t.Fatal("point parity", i, e)
		}
	}
	got[1].Entry.Payload[0] = 'x'
	if got[2].Entry.Payload[0] == 'x' {
		t.Fatal("duplicate result alias")
	}
	if r, e := l.GetServiceAdmission(ctx, base); e != nil || r.Payload[0] == 'x' {
		t.Fatal("stored payload alias", e)
	}
	rows, err := l.db.Query("EXPLAIN QUERY PLAN "+serviceBatchLookupSQL, 1<<20, base.Tenant, base.Stream, base.Contract, base.Journal, base.Event)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	found := false
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err = rows.Scan(&id, &parent, &unused, &detail); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(detail, "SEARCH research_log USING INDEX "+serviceIndexName) && strings.Count(detail, "<expr>=?") == 5 {
			found = true
		}
	}
	if err = rows.Err(); err != nil || !found {
		t.Fatal("incomplete index search", err)
	}
}

func TestServiceReadBatchSnapshot(t *testing.T) {
	l, path := sourceReadFixture(t)
	ctx := context.Background()
	first, late := sourceFixture(), sourceFixture()
	late.Event = "late"
	a, b := serviceEntry("1", first), serviceEntry("2", late)
	if _, _, err := l.Append(ctx, a.Key, a.Kind, a.Payload); err != nil {
		t.Fatal(err)
	}
	u := url.URL{Scheme: "file", Path: path}
	writer, err := sql.Open("sqlite", u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	requests := []ServiceIdentity{first, late}
	got, err := l.getServiceAdmissions(ctx, requests, func(i int) {
		if i == 0 {
			key, _ := json.Marshal(b.Key)
			if _, e := writer.ExecContext(ctx, "INSERT INTO research_log(identity,kind,payload) VALUES(?,?,?)", string(key), b.Kind, []byte(b.Payload)); e != nil {
				t.Fatal(e)
			}
		}
	})
	if err != nil || !got[0].Found || got[1].Found {
		t.Fatal("mixed snapshot", got, err)
	}
	if got, err = l.GetServiceAdmissions(ctx, requests); err != nil || !got[1].Found {
		t.Fatal("new snapshot missed commit", got, err)
	}
}

func TestServiceReadBatchCapsAndCancellation(t *testing.T) {
	l, _ := sourceReadFixture(t)
	ctx := context.Background()
	k := sourceFixture()
	bad := k
	bad.Event = string([]byte{255})
	for _, req := range [][]ServiceIdentity{nil, make([]ServiceIdentity, 513), {ServiceIdentity{}}, {bad}} {
		if r, e := l.GetServiceAdmissions(ctx, req); e == nil || r != nil {
			t.Fatal("invalid input", e)
		}
	}
	full := make([]ServiceIdentity, 512)
	for i := range full {
		full[i] = k
	}
	if r, e := l.GetServiceAdmissions(ctx, full); e != nil || len(r) != 512 {
		t.Fatal("count limit", e)
	}
	field := strings.Repeat("\x00", 4096)
	for i := range full {
		full[i] = ServiceIdentity{field, field, field, field, field}
	}
	if r, e := l.GetServiceAdmissions(ctx, full); e == nil || r != nil || !strings.Contains(e.Error(), "request byte cap") {
		t.Fatal("request cap", e)
	}
	entry := serviceEntry("1", k)
	// Valid bound JSON exactly at the individual materialization cap.
	var value map[string]any
	if err := json.Unmarshal(entry.Payload, &value); err != nil {
		t.Fatal(err)
	}
	value["padding"] = ""
	raw, _ := json.Marshal(value)
	value["padding"] = strings.Repeat("x", (1<<20)-len(raw))
	entry.Payload, _ = json.Marshal(value)
	if len(entry.Payload) != 1<<20 {
		t.Fatal("fixture size")
	}
	if _, _, err := l.Append(ctx, entry.Key, entry.Kind, entry.Payload); err != nil {
		t.Fatal(err)
	}
	req := make([]ServiceIdentity, 8)
	for i := range req {
		req[i] = k
	}
	if r, e := l.GetServiceAdmissions(ctx, req); e != nil || len(r) != 8 {
		t.Fatal("exact payload cap", e)
	}
	missing := k
	missing.Event = "absent"
	if _, e := l.GetServiceAdmissions(ctx, append(req, missing)); e != nil {
		t.Fatal("zero-budget miss", e)
	}
	if r, e := l.GetServiceAdmissions(ctx, append(req, k)); e == nil || r != nil {
		t.Fatal("payload overflow", e)
	}
	c, cancel := context.WithCancel(ctx)
	if r, e := l.getServiceAdmissions(c, []ServiceIdentity{k, k}, func(int) { cancel() }); !errors.Is(e, context.Canceled) || r != nil {
		t.Fatal("partial canceled batch", e)
	}
	if r, e := l.GetServiceAdmissions(c, []ServiceIdentity{k}); !errors.Is(e, context.Canceled) || r != nil {
		t.Fatal("precancel", e)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("missing panic")
			}
		}()
		_, _ = l.getServiceAdmissions(ctx, []ServiceIdentity{k}, func(int) { panic("read boundary") })
	}()
	if _, e := l.GetServiceAdmissions(ctx, []ServiceIdentity{k}); e != nil {
		t.Fatal("transaction leak", e)
	}
}

func TestServiceReadBatchCorruptionAndIndexFailure(t *testing.T) {
	for _, kind := range []string{"oversized", "text", "binding", "identity", "missing-index", "ambiguous"} {
		t.Run(kind, func(t *testing.T) {
			l, _ := sourceReadFixture(t)
			ctx := context.Background()
			a, b := sourceFixture(), sourceFixture()
			b.Event = "bad"
			first, second := serviceEntry("1", a), serviceEntry("2", b)
			if _, err := l.AppendBatchPrepared(ctx, []AppendRequest{first, second}); err != nil {
				t.Fatal(err)
			}
			key, _ := json.Marshal(second.Key)
			switch kind {
			case "oversized":
				var v map[string]any
				json.Unmarshal(second.Payload, &v)
				v["padding"] = strings.Repeat("x", 1<<20)
				raw, _ := json.Marshal(v)
				if _, err := l.db.Exec("UPDATE research_log SET payload=? WHERE identity=?", raw, string(key)); err != nil {
					t.Fatal(err)
				}
			case "text":
				if _, err := l.db.Exec("UPDATE research_log SET payload=CAST(payload AS TEXT) WHERE identity=?", string(key)); err != nil {
					t.Fatal(err)
				}
			case "binding":
				var v map[string]any
				json.Unmarshal(second.Payload, &v)
				v["Binding"].(map[string]any)["Tenant"] = "other"
				raw, _ := json.Marshal(v)
				if _, err := l.db.Exec("UPDATE research_log SET payload=? WHERE identity=?", raw, string(key)); err != nil {
					t.Fatal(err)
				}
			case "identity":
				bad := second.Key
				bad.Event = strings.Repeat("x", 131073)
				raw, _ := json.Marshal(bad)
				if _, err := l.db.Exec("UPDATE research_log SET identity=? WHERE identity=?", string(raw), string(key)); err != nil {
					t.Fatal(err)
				}
			case "missing-index", "ambiguous":
				if _, err := l.db.Exec("DROP INDEX " + serviceIndexName); err != nil {
					t.Fatal(err)
				}
				if kind == "ambiguous" {
					if _, err := l.db.Exec("CREATE INDEX " + serviceIndexName + " ON research_log(kind)"); err != nil {
						t.Fatal(err)
					}
					dup := serviceEntry("3", b)
					if _, _, err := l.Append(ctx, dup.Key, dup.Kind, dup.Payload); err != nil {
						t.Fatal(err)
					}
				}
			}
			if r, err := l.GetServiceAdmissions(ctx, []ServiceIdentity{a, b}); err == nil || r != nil || errors.Is(err, sql.ErrNoRows) {
				t.Fatal("bad row/index treated as miss or partial success", r, err)
			}
		})
	}
}
