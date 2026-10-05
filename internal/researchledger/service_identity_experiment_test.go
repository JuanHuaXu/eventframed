package researchledger

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestServiceIdentityScalingExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SOURCE_INDEX_ARTIFACT")
	if path == "" {
		t.Skip("opt-in source lookup scaling")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{"go.mod", "go.sum", "internal/researchledger/ledger.go", "internal/researchledger/batch.go", "internal/researchledger/service_identity.go", "internal/researchledger/service_identity_test.go", "internal/researchledger/service_identity_experiment_test.go", "docs/experiments/mmm-source-index-v46-protocol.md"} {
		b, err := os.ReadFile("../../" + name)
		if err != nil {
			t.Fatal(err)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if err = enc.Encode(map[string]any{"kind": "header", "expected_cases": 3, "Sources": sources, "Hashes": hashes, "GoVersion": runtime.Version(), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for _, n := range []int{100, 1000, 10000} {
		l, err := Open(t.TempDir() + "/scaling.sqlite")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { l.Close() })
		if err = l.EnableServiceIdentity(ctx); err != nil {
			t.Fatal(err)
		}
		for begin := 1; begin <= n; begin += 200 {
			var requests []AppendRequest
			for i := begin; i < begin+200 && i <= n; i++ {
				k := sourceFixture()
				k.Event = fmt.Sprint(i)
				requests = append(requests, serviceEntry(fmt.Sprint(i), k))
			}
			if _, err = l.AppendBatchPrepared(ctx, requests); err != nil {
				t.Fatal(err)
			}
		}
		var hits, misses []int64
		for i := 0; i < 128; i++ {
			k := sourceFixture()
			k.Event = fmt.Sprint(1 + (i*7919)%n)
			start := time.Now()
			e, err := l.GetServiceAdmission(ctx, k)
			hits = append(hits, time.Since(start).Nanoseconds())
			if err != nil || e.Key.Event != k.Event {
				t.Fatal("hit", err)
			}
			k.Event = "missing"
			start = time.Now()
			_, err = l.GetServiceAdmission(ctx, k)
			misses = append(misses, time.Since(start).Nanoseconds())
			if !errors.Is(err, sql.ErrNoRows) {
				t.Fatal("miss", err)
			}
		}
		k := sourceFixture()
		rows, err := l.db.QueryContext(ctx, "EXPLAIN QUERY PLAN "+serviceLookupSQL, k.Tenant, k.Stream, k.Contract, k.Journal, k.Event)
		if err != nil {
			t.Fatal(err)
		}
		var plans []string
		for rows.Next() {
			var id, parent, unused int
			var detail string
			if err = rows.Scan(&id, &parent, &unused, &detail); err != nil {
				t.Fatal(err)
			}
			plans = append(plans, detail)
		}
		if err = rows.Err(); err != nil {
			t.Fatal(err)
		}
		rows.Close()
		if len(plans) != 1 || !strings.Contains(plans[0], "SEARCH research_log USING INDEX "+serviceIndexName) || strings.Count(plans[0], "<expr>=?") != 5 {
			t.Fatal("scan", plans)
		}
		var count int
		if err = l.db.QueryRowContext(ctx, "SELECT count(*) FROM research_log").Scan(&count); err != nil || count != n {
			t.Fatal("history count", count, err)
		}
		if err = l.Close(); err != nil {
			t.Fatal(err)
		}
		if err = enc.Encode(map[string]any{"History": n, "HitNS": hits, "MissNS": misses, "Plan": plans}); err != nil {
			t.Fatal(err)
		}
		if err = f.Sync(); err != nil {
			t.Fatal(err)
		}
		t.Logf("history%d hits%d misses%d %s", n, len(hits), len(misses), plans[0])
	}
}
