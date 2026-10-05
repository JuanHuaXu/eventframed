package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"runtime"
	"testing"
)

func TestResearchUnionLoadAccounting(t *testing.T) {
	for _, writes := range []int{0, 4} {
		r := researchGuardLoadValidationArm(t, "group4postverify", 0, 8, writes, true, true, true)
		checkQueuedGuardLoad(t, r, writes)
		if !r.PreparedWrites || !r.UnionReads || r.Accepted == 0 {
			t.Fatal("union path not exercised")
		}
	}
}

func TestResearchUnionLoadExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_UNION_ARTIFACT")
	if path == "" {
		t.Skip("opt-in union load")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{"go.mod", "go.sum", "internal/service/research_union_load_test.go", "internal/service/research_grouped_consumer_test.go", "internal/service/research_guard_load_test.go", "internal/service/research_queued_load_test.go", "internal/service/research_batch_validation.go", "internal/service/research_group_validation.go", "internal/researchpublicationstore/guard.go", "internal/researchpublicationstore/store.go", "internal/researchmemory/durable.go", "internal/researchmemory/durable_prepared.go", "internal/researchmemory/durable_batch_admit.go", "internal/researchmemory/durable_batch_discard.go", "internal/researchmemory/durable_batch_read.go", "internal/researchmemory/record.go", "internal/researchledger/ledger.go", "internal/researchledger/batch.go", "internal/researchledger/read_batch.go", "docs/experiments/mmm-union-reads-v42-protocol.md"} {
		b, err := os.ReadFile("../../" + name)
		if err != nil {
			t.Fatal(err)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if err = enc.Encode(map[string]any{"kind": "header", "expected_arms": 12, "Sources": sources, "Hashes": hashes, "GoVersion": runtime.Version(), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "GOMAXPROCS": runtime.GOMAXPROCS(0), "wait_budget_ms": 20}); err != nil {
		t.Fatal(err)
	}
	for trial := 0; trial < 3; trial++ {
		for order := 0; order < 4; order++ {
			arm := (trial + order) % 4
			mode := []string{"off", "group4", "group4postverify", "group4postverify"}[arm]
			r := researchGuardLoadValidationArm(t, mode, trial, 192, 96, true, arm >= 2, arm == 3)
			if err = enc.Encode(r); err != nil {
				t.Fatal(err)
			}
			if err = f.Sync(); err != nil {
				t.Fatal(err)
			}
			checkQueuedGuardLoad(t, r, 96)
			t.Logf("trial%d %s prepared%t union%t accepted%d dropped%d expired%d admits%d discards%d", trial, mode, r.PreparedWrites, r.UnionReads, r.Accepted, r.Dropped, r.WaitExpired, r.Admits, r.Discards)
		}
	}
}
