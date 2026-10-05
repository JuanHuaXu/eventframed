package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"runtime"
	"testing"
)

func checkQueuedGuardLoad(t *testing.T, r researchGuardLoadResult, writes int) {
	t.Helper()
	if len(r.Errors) > 0 || len(r.ReadNS) != r.Requests || len(r.WriteNS) != writes {
		t.Errorf("request accounting: %+v", r)
		return
	}
	if r.Mode == "off" {
		return
	}
	if r.Attempts+r.Dropped != uint64(r.Requests) || r.Attempts != r.Accepted+r.Busy+r.Stale+r.WaitExpired || r.Validated != 50*r.Accepted {
		t.Errorf("admission accounting: %+v", r)
	}
	if r.Mode == "group4durable" || r.Mode == "group4transactions" || r.Mode == "group4reads" || r.Mode == "group4postdiscard" || r.Mode == "group4postverify" {
		if r.Admits != r.Validated || r.Discards != r.Admits {
			t.Error("durable counters mismatch")
		}
	} else if r.Admits != 0 || r.Discards != 0 {
		t.Error("unexpected ledger activity")
	}
	if writes == 0 && r.Accepted == 0 {
		t.Error("no read-only admissions")
	}
	grouped := r.Mode == "group1" || r.Mode == "group4" || r.Mode == "group4durable" || r.Mode == "group4transactions" || r.Mode == "group4reads" || r.Mode == "group4postdiscard" || r.Mode == "group4postverify"
	if !grouped && uint64(len(r.Phases)) != r.Attempts {
		t.Error("phase count mismatch")
	}
	if grouped && len(r.GroupSizes) != len(r.Phases) {
		t.Fatal("group phase count mismatch")
	}
	var entered, accepted uint64
	var attempts uint64
	for i, p := range r.Phases {
		if p.VerifiedDiscardNS < 0 || (!r.CombinedSourceCleanup && p.VerifiedDiscardNS != 0) {
			t.Error("invalid verified cleanup phase")
		}
		if r.CombinedSourceCleanup && (p.DurableVerifyNS != 0 || p.DurableDiscardNS != 0) {
			t.Error("combined cleanup double counted")
		}
		weight := uint64(1)
		if grouped {
			limit := 4
			if r.Mode == "group1" {
				limit = 1
			}
			if r.GroupSizes[i] < 1 || r.GroupSizes[i] > limit {
				t.Fatal("group cap violated")
			}
			weight = uint64(r.GroupSizes[i])
		}
		attempts += weight
		if p.QueueNS < 0 || p.PrefetchNS < 0 || p.EntryNS < 0 || p.CallbackNS < 0 || p.TotalNS < 0 || p.PostGuardNS < 0 {
			t.Error("negative monotonic duration")
		}
		if p.Entered {
			inside := p.DurableAdmitNS + p.DurableVerifyNS + p.DurableDiscardNS
			if r.Mode == "group4postdiscard" || r.Mode == "group4postverify" {
				inside -= p.DurableDiscardNS
				outside := p.DurableDiscardNS
				outside += p.VerifiedDiscardNS
				if r.Mode == "group4postverify" {
					inside -= p.DurableVerifyNS
					outside += p.DurableVerifyNS
				}
				if outside > p.PostGuardNS {
					t.Error("discard escaped post-guard accounting")
				}
			} else if p.PostGuardNS != 0 {
				t.Error("unexpected post-guard work")
			}
			if p.DurableAdmitNS < 0 || p.DurableVerifyNS < 0 || p.DurableDiscardNS < 0 || inside > p.CallbackNS {
				t.Error("durable phase containment violated")
			}
			entered += weight
			if p.TotalNS < p.EntryNS+p.CallbackNS+p.PostGuardNS {
				t.Error("phase duration exceeds guard total")
			}
		}
		if p.Accepted {
			accepted += weight
			if !p.Entered {
				t.Error("accepted without callback")
			}
		}
	}
	if attempts != r.Attempts || accepted != r.Accepted || entered < accepted {
		t.Error("phase outcomes mismatch")
	}
}

func TestResearchQueuedLoadAccounting(t *testing.T) {
	for _, writes := range []int{0, 4} {
		r := researchGuardLoadProbeArm(t, "queued", 0, 8, writes, true)
		checkQueuedGuardLoad(t, r, writes)
	}
}

func TestResearchQueuedLoadExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_QUEUED_LOAD_ARTIFACT")
	if path == "" {
		t.Skip("opt-in queued guard load")
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{"go.mod", "go.sum", "internal/service/research_queued_load_test.go", "internal/service/research_guard_load_test.go", "internal/service/research_durable_validation.go", "internal/researchpublicationstore/guard.go", "internal/researchpublicationstore/store.go", "internal/researchpublicationstore/mutations.go", "docs/experiments/mmm-queued-load-v22-protocol.md"} {
		b, e := os.ReadFile("../../" + name)
		if e != nil {
			t.Fatal(e)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if e = enc.Encode(map[string]any{"kind": "header", "expected_arms": 9, "Sources": sources, "Hashes": hashes, "GoVersion": runtime.Version(), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "GOMAXPROCS": runtime.GOMAXPROCS(0), "wait_budget_ms": 20}); e != nil {
		t.Fatal(e)
	}
	for trial := 0; trial < 3; trial++ {
		for order := 0; order < 3; order++ {
			mode := []string{"off", "validate", "queued"}[(trial+order)%3]
			r := researchGuardLoadArm(t, mode, trial, 192)
			if e = enc.Encode(r); e != nil {
				t.Fatal(e)
			}
			if e = f.Sync(); e != nil {
				t.Fatal(e)
			}
			checkQueuedGuardLoad(t, r, 96)
			t.Logf("trial%d %s accepted%d busy%d stale%d expired%d dropped%d", trial, mode, r.Accepted, r.Busy, r.Stale, r.WaitExpired, r.Dropped)
		}
	}
}
