package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"runtime"
	"testing"
)

func TestResearchSnapshotReadsLoadAccounting(t *testing.T) {
	for _, writes := range []int{0, 4} {
		r := researchGuardLoadProbeArm(t, "group4reads", 0, 8, writes, true)
		checkQueuedGuardLoad(t, r, writes)
		if r.Accepted == 0 {
			t.Error("no snapshot read admissions")
		}
	}
}

func TestResearchSnapshotReadsLoadExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SNAPSHOT_READS_ARTIFACT")
	if path == "" {
		t.Skip("opt-in snapshot reads load")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{"go.mod", "go.sum", "internal/service/research_snapshot_reads_load_test.go", "internal/service/research_grouped_consumer_test.go", "internal/service/research_queued_load_test.go", "internal/service/research_guard_load_test.go", "internal/service/research_batch_validation.go", "internal/researchpublicationstore/guard.go", "internal/researchpublicationstore/store.go", "internal/researchmemory/durable.go", "internal/researchmemory/durable_batch_admit.go", "internal/researchmemory/durable_batch_discard.go", "internal/researchmemory/durable_batch_read.go", "internal/researchmemory/record.go", "internal/researchledger/ledger.go", "internal/researchledger/batch.go", "internal/researchledger/read_batch.go", "docs/experiments/mmm-snapshot-reads-v36-protocol.md"} {
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
			mode := []string{"off", "group4", "group4transactions", "group4reads"}[(trial+order)%4]
			r := researchGuardLoadArm(t, mode, trial, 192)
			if err = enc.Encode(r); err != nil {
				t.Fatal(err)
			}
			if err = f.Sync(); err != nil {
				t.Fatal(err)
			}
			checkQueuedGuardLoad(t, r, 96)
			t.Logf("trial%d %s accepted%d expired%d dropped%d admits%d discards%d", trial, mode, r.Accepted, r.WaitExpired, r.Dropped, r.Admits, r.Discards)
		}
	}
}
