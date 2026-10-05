package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func checkGuardIsolation(t *testing.T, r researchGuardLoadResult, writes int) {
	t.Helper()
	if len(r.Errors) > 0 || len(r.ReadNS) != r.Requests || len(r.WriteNS) != writes || r.Attempts+r.Dropped != uint64(r.Requests) || r.Attempts != r.Accepted+r.Busy+r.Stale || r.Validated != 50*r.Accepted {
		t.Errorf("isolation accounting: %+v", r)
	}
	if writes == 0 && r.Accepted == 0 {
		t.Error("read-only control never entered validation")
	}
	if r.Admits != 0 || r.Discards != 0 {
		t.Error("unexpected ledger operation")
	}
}

func TestResearchGuardIsolationAccounting(t *testing.T) {
	for _, prefetch := range []bool{false, true} {
		for _, writes := range []int{0, 4} {
			r := researchGuardLoadProbeArm(t, "validate", 0, 8, writes, prefetch)
			checkGuardIsolation(t, r, writes)
		}
	}
}

func TestResearchGuardIsolationExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_GUARD_ISOLATION_ARTIFACT")
	if path == "" {
		t.Skip("opt-in guard isolation")
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{"internal/service/research_guard_isolation_test.go", "internal/service/research_guard_load_test.go", "internal/service/research_durable_validation.go", "internal/researchpublicationstore/guard.go", "internal/researchpublicationstore/store.go", "internal/researchpublicationstore/mutations.go", "docs/experiments/mmm-guard-isolation-v20-protocol.md"} {
		b, e := os.ReadFile("../../" + name)
		if e != nil {
			t.Fatal(e)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if e = enc.Encode(map[string]any{"kind": "header", "expected_arms": 12, "Sources": sources, "Hashes": hashes}); e != nil {
		t.Fatal(e)
	}
	cells := []struct {
		writes   int
		prefetch bool
	}{{0, true}, {32, true}, {0, false}, {32, false}}
	for trial := 0; trial < 3; trial++ {
		for order := 0; order < 4; order++ {
			cell := cells[(trial+order)%4]
			r := researchGuardLoadProbeArm(t, "validate", trial, 64, cell.writes, cell.prefetch)
			if e = enc.Encode(map[string]any{"writes": cell.writes, "prefetch": cell.prefetch, "result": r}); e != nil {
				t.Fatal(e)
			}
			if e = f.Sync(); e != nil {
				t.Fatal(e)
			}
			checkGuardIsolation(t, r, cell.writes)
			t.Logf("trial%d writes%d prefetch%v accepted%d busy%d stale%d", trial, cell.writes, cell.prefetch, r.Accepted, r.Busy, r.Stale)
		}
	}
}
