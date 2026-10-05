package researchledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func serviceEntry(id string, k ServiceIdentity) AppendRequest {
	p, _ := json.Marshal(struct {
		Binding struct{ Tenant, JournalID, EventID string }
		Value   int
	}{Binding: struct{ Tenant, JournalID, EventID string }{k.Tenant, k.Journal, k.Event}, Value: 7})
	return AppendRequest{Key{k.Tenant, k.Stream, id, k.Contract}, "admit", p}
}
func sourceFixture() ServiceIdentity {
	return ServiceIdentity{"tenant", "stream", "contract", "journal", "event"}
}

func TestServiceIdentityAtomicityAndReopen(t *testing.T) {
	for _, mode := range []string{"single", "batch", "prepared", "conditional"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			path := t.TempDir() + "/source.sqlite"
			l, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { l.Close() }()
			if err = l.EnableServiceIdentity(ctx); err != nil {
				t.Fatal(err)
			}
			k := sourceFixture()
			original := serviceEntry("1", k)
			if _, _, err = l.Append(ctx, original.Key, original.Kind, original.Payload); err != nil {
				t.Fatal(err)
			}
			dup := serviceEntry("2", k)
			write := l.AppendBatch
			if mode == "prepared" {
				write = l.AppendBatchPrepared
			}
			if mode == "conditional" {
				write = l.AppendBatchConditionalInsert
				if ack, e := write(ctx, []AppendRequest{original}); e != nil || len(ack) != 1 || !ack[0].Retry {
					t.Fatal("conditional source retry", ack, e)
				}
			}
			if mode == "single" {
				_, _, err = l.Append(ctx, dup.Key, dup.Kind, dup.Payload)
			} else {
				other := k
				other.Event = "other"
				_, err = write(ctx, []AppendRequest{serviceEntry("3", other), dup})
			}
			if err == nil {
				t.Fatal("duplicate source accepted")
			}
			rows, err := l.ReadAfter(ctx, 0, 256)
			if err != nil || len(rows) != 1 {
				t.Fatal("partial conflict commit", err, len(rows))
			}
			if _, retry, err := l.Append(ctx, original.Key, original.Kind, original.Payload); err != nil || !retry {
				t.Fatal("exact retry", err)
			}
			if err = l.Close(); err != nil {
				t.Fatal(err)
			}
			l, err = Open(path)
			if err != nil {
				t.Fatal(err)
			}
			if err = l.EnableServiceIdentity(ctx); err != nil {
				t.Fatal("idempotent enable", err)
			}
			got, err := l.GetServiceAdmission(ctx, k)
			if err != nil || got.Key != original.Key || !reflect.DeepEqual(got.Payload, original.Payload) {
				t.Fatal("lost original", err)
			}
			if _, _, err = l.Append(ctx, dup.Key, dup.Kind, dup.Payload); err == nil {
				t.Fatal("index lost on reopen")
			}
			if _, _, err = l.Append(ctx, original.Key, "feedback", json.RawMessage(`{"discard":true}`)); err != nil {
				t.Fatal(err)
			}
			if _, err = l.GetServiceAdmission(ctx, k); err != nil {
				t.Fatal("terminal erased identity", err)
			}
		})
	}
}

func TestServiceIdentityScopeAndPlan(t *testing.T) {
	l, err := Open(t.TempDir() + "/scope.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	ctx := context.Background()
	if err = l.EnableServiceIdentity(ctx); err != nil {
		t.Fatal(err)
	}
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
	var req []AppendRequest
	for i, k := range keys {
		req = append(req, serviceEntry(fmt.Sprint(i+1), k))
	}
	if _, err = l.AppendBatchPrepared(ctx, req); err != nil {
		t.Fatal(err)
	}
	for i, k := range keys {
		e, err := l.GetServiceAdmission(ctx, k)
		if err != nil || e.Key != req[i].Key {
			t.Fatal("scope collision", i, err)
		}
	}
	missing := base
	missing.Event = "missing"
	if _, err = l.GetServiceAdmission(ctx, missing); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("missing", err)
	}
	rows, err := l.db.Query("EXPLAIN QUERY PLAN "+serviceLookupSQL, base.Tenant, base.Stream, base.Contract, base.Journal, base.Event)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	search := false
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err = rows.Scan(&id, &parent, &unused, &detail); err != nil {
			t.Fatal(err)
		}
		t.Log(detail)
		if strings.Contains(detail, "SEARCH research_log USING INDEX "+serviceIndexName) && strings.Count(detail, "<expr>=?") == 5 {
			search = true
		}
	}
	if err = rows.Err(); err != nil || !search {
		t.Fatal("unindexed/incomplete lookup", err)
	}
}

func TestServiceIdentityActivationRefusesBadHistory(t *testing.T) {
	for _, kind := range []string{"duplicate", "invalid-binding", "wrong-index", "canceled"} {
		t.Run(kind, func(t *testing.T) {
			l, err := Open(t.TempDir() + "/old.sqlite")
			if err != nil {
				t.Fatal(err)
			}
			defer l.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			k := sourceFixture()
			req := []AppendRequest{serviceEntry("1", k)}
			switch kind {
			case "duplicate":
				req = append(req, serviceEntry("2", k))
			case "invalid-binding":
				req[0].Payload = json.RawMessage(`{"Binding":{"Tenant":"other","JournalID":"journal","EventID":"event"}}`)
			}
			if _, err = l.AppendBatch(ctx, req); err != nil {
				t.Fatal(err)
			}
			before, err := l.ReadAfter(ctx, 0, 256)
			if err != nil {
				t.Fatal(err)
			}
			if kind == "wrong-index" {
				if _, err = l.db.Exec("CREATE INDEX " + serviceIndexName + " ON research_log(kind)"); err != nil {
					t.Fatal(err)
				}
			}
			if kind == "canceled" {
				cancel()
			}
			if err = l.EnableServiceIdentity(ctx); err == nil {
				t.Fatal("unsafe activation")
			}
			after, err := l.ReadAfter(context.Background(), 0, 256)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatal("history rewritten", err)
			}
			if kind != "wrong-index" {
				var n int
				if err = l.db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name=?", serviceIndexName).Scan(&n); err != nil || n != 0 {
					t.Fatal("partial activation", err, n)
				}
			}
		})
	}
}

func TestServiceIdentityConcurrentClaim(t *testing.T) {
	l, err := Open(t.TempDir() + "/claims.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	ctx := context.Background()
	if err = l.EnableServiceIdentity(ctx); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	accepted := 0
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r := serviceEntry(fmt.Sprint(i+1), sourceFixture())
			if _, _, err := l.Append(ctx, r.Key, r.Kind, r.Payload); err == nil {
				mu.Lock()
				accepted++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	if accepted != 1 {
		t.Fatal("claim count", accepted)
	}
	if _, err = l.GetServiceAdmission(ctx, sourceFixture()); err != nil {
		t.Fatal(err)
	}
}

func TestServiceIdentityLookupBounds(t *testing.T) {
	l, err := Open(t.TempDir() + "/bounds.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	ctx := context.Background()
	k := sourceFixture()
	if _, err = l.GetServiceAdmission(ctx, k); err == nil || errors.Is(err, sql.ErrNoRows) {
		t.Fatal("missing index treated as miss", err)
	}
	if err = l.EnableServiceIdentity(ctx); err != nil {
		t.Fatal(err)
	}
	bad := k
	bad.Event = strings.Repeat("x", 4097)
	if _, err = l.GetServiceAdmission(ctx, bad); err == nil {
		t.Fatal("oversized key")
	}
	bad.Event = string([]byte{0xff})
	if _, err = l.GetServiceAdmission(ctx, bad); err == nil {
		t.Fatal("invalid UTF8 key")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = l.GetServiceAdmission(canceled, k); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	r := serviceEntry("1", k)
	// Inject a valid but oversized bound blob through SQL, outside normal append
	// validation, to verify the read envelope rejects before materializing it.
	var payload map[string]any
	if err = json.Unmarshal(r.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	payload["Large"] = strings.Repeat("x", 1<<20)
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	identity, _ := json.Marshal(r.Key)
	if _, err = l.db.Exec("INSERT INTO research_log(identity,kind,payload) VALUES(?,?,?)", string(identity), "admit", raw); err != nil {
		t.Fatal(err)
	}
	if _, err = l.GetServiceAdmission(ctx, k); err == nil {
		t.Fatal("oversized original returned")
	}
}

func TestServiceIdentityCrashHelper(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SOURCE_CHILD")
	if path == "" {
		t.Skip("subprocess only")
	}
	l, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = l.EnableServiceIdentity(context.Background()); err != nil {
		t.Fatal(err)
	}
	var hook func()
	if os.Getenv("EVENTFRAME_SOURCE_PHASE") == "before" {
		hook = func() { os.Exit(23) }
	}
	if _, err = l.appendBatchMode(context.Background(), []AppendRequest{serviceEntry("1", sourceFixture())}, true, hook); err != nil {
		t.Fatal(err)
	}
	os.Exit(23)
}

func TestServiceIdentityCrashBoundary(t *testing.T) {
	for _, phase := range []string{"before", "after"} {
		t.Run(phase, func(t *testing.T) {
			path := t.TempDir() + "/crash.sqlite"
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestServiceIdentityCrashHelper$")
			cmd.Env = append(os.Environ(), "EVENTFRAME_SOURCE_CHILD="+path, "EVENTFRAME_SOURCE_PHASE="+phase)
			out, err := cmd.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 23 {
				t.Fatalf("child %v %s", err, out)
			}
			l, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer l.Close()
			if err = l.EnableServiceIdentity(ctx); err != nil {
				t.Fatal(err)
			}
			_, err = l.GetServiceAdmission(ctx, sourceFixture())
			if phase == "before" {
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatal(err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			r := serviceEntry("1", sourceFixture())
			_, retry, err := l.Append(ctx, r.Key, r.Kind, r.Payload)
			if err != nil || retry != (phase == "after") {
				t.Fatal("recovery", retry, err)
			}
			if _, err = l.GetServiceAdmission(ctx, sourceFixture()); err != nil {
				t.Fatal(err)
			}
		})
	}
}
