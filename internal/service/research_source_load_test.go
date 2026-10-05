package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
)

// A cold validation preview is not the original. Only the owner's emitted record
// is retained for readback; no caller learner ID crosses SourceOwner.Admit.
func researchPersistSourceGroup(ctx context.Context, o *researchmemory.SourceOwner, groups [][]researchmemory.RecordedPrediction, phase *researchGuardPhase, result *researchGuardLoadResult) error {
	var requests []researchmemory.SourceAdmissionRequest
	for _, g := range groups {
		for _, r := range g {
			if r.Ready || r.Binding == nil {
				return errors.New("source load requires cold bound previews")
			}
			requests = append(requests, researchmemory.SourceAdmissionRequest{Features: r.Prediction.Features, Baseline: r.Outer[0], At: r.At, Binding: *r.Binding})
		}
	}
	start := time.Now()
	admitted, err := o.Admit(ctx, requests)
	phase.DurableAdmitNS += time.Since(start).Nanoseconds()
	if err != nil {
		return err
	}
	if len(admitted) != len(requests) {
		return errors.New("source load admission count mismatch")
	}
	index := 0
	for i, g := range groups {
		for j, preview := range g {
			r := admitted[index]
			if r.Retry || r.Record.Ready || r.Record.Prediction.ID != result.Admits+uint64(index)+1 {
				return errors.New("unexpected source load identity or warm record")
			}
			// Check the cold envelope independent of the preview's temporary ID.
			// Never write this adjusted preview to storage or use it as the original.
			preview.Prediction.ID = r.Record.Prediction.ID
			if !reflect.DeepEqual(preview, r.Record) {
				return errors.New("source load original differs from cold envelope")
			}
			groups[i][j] = r.Record
			index++
		}
	}
	result.Admits += uint64(len(admitted))
	return nil
}

func researchFinishSourceGroup(ctx context.Context, o *researchmemory.SourceOwner, groups [][]researchmemory.RecordedPrediction, phase *researchGuardPhase, result *researchGuardLoadResult) error {
	if result.CombinedSourceCleanup {
		start := time.Now()
		var requests []researchmemory.VerifiedSourceDiscardRequest
		for _, g := range groups {
			for _, r := range g {
				requests = append(requests, researchmemory.VerifiedSourceDiscardRequest{Original: r, Available: r.At})
			}
		}
		retries, err := o.VerifyAndDiscardBatch(ctx, requests)
		phase.VerifiedDiscardNS += time.Since(start).Nanoseconds()
		if err != nil {
			return err
		}
		if len(retries) != len(requests) {
			return errors.New("verified cleanup count mismatch")
		}
		for _, retry := range retries {
			if retry {
				return errors.New("unexpected verified cleanup retry")
			}
		}
		result.Discards += uint64(len(retries))
		return nil
	}
	start := time.Now()
	var discards []researchmemory.SourceDiscardRequest
	var batched []researchmemory.RecordedPrediction
	if result.BatchSourceReads {
		var refs []researchmemory.SourceReference
		for _, g := range groups {
			for _, r := range g {
				if r.Binding == nil {
					return errors.New("missing source original binding")
				}
				refs = append(refs, researchmemory.SourceReference{JournalID: r.Binding.JournalID, EventID: r.Binding.EventID})
			}
		}
		var err error
		batched, err = o.LookupBatch(ctx, refs)
		if err != nil {
			return err
		}
		if len(batched) != len(refs) {
			return errors.New("source batch readback count mismatch")
		}
	}
	index := 0
	for _, g := range groups {
		for _, r := range g {
			if r.Binding == nil {
				return errors.New("missing source original binding")
			}
			var saved researchmemory.RecordedPrediction
			var err error
			if result.BatchSourceReads {
				saved = batched[index]
			} else {
				saved, err = o.Lookup(ctx, r.Binding.JournalID, r.Binding.EventID)
			}
			index++
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(saved, r) {
				return errors.New("source load readback mismatch")
			}
			discards = append(discards, researchmemory.SourceDiscardRequest{JournalID: r.Binding.JournalID, EventID: r.Binding.EventID, Available: r.At})
		}
	}
	phase.DurableVerifyNS += time.Since(start).Nanoseconds()
	start = time.Now()
	retries, err := o.DiscardBatch(ctx, discards)
	phase.DurableDiscardNS += time.Since(start).Nanoseconds()
	if err != nil {
		return err
	}
	if len(retries) != len(discards) {
		return errors.New("source load discard count mismatch")
	}
	for _, retry := range retries {
		if retry {
			return errors.New("unexpected source load terminal retry")
		}
	}
	result.Discards += uint64(len(retries))
	return nil
}

func TestResearchSourceLoadAccounting(t *testing.T) {
	for _, writes := range []int{0, 4} {
		r := researchGuardLoadSourceArm(t, "group4postverify", 0, 8, writes, true, true, true, true)
		checkQueuedGuardLoad(t, r, writes)
		if !r.SourceOwner || r.Accepted == 0 || r.Admits == 0 || r.Admits != r.Discards {
			t.Fatal("source path not exercised", r)
		}
	}
}

func TestResearchSourceLoadExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SOURCE_LOAD_ARTIFACT")
	if path == "" {
		t.Skip("opt-in source owner load experiment")
	}
	runResearchSourceLoad(t, path, false)
}

func TestResearchSourceBatchLoadAccounting(t *testing.T) {
	for _, writes := range []int{0, 4} {
		r := researchGuardLoadSourceReadArm(t, "group4postverify", 0, 8, writes, true, true, true, true, true)
		checkQueuedGuardLoad(t, r, writes)
		if !r.SourceOwner || !r.BatchSourceReads || r.Accepted == 0 || r.Admits == 0 || r.Admits != r.Discards {
			t.Fatal("batch source path not exercised")
		}
	}
}

func TestResearchSourceBatchLoadExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SOURCE_BATCH_LOAD_ARTIFACT")
	if path == "" {
		t.Skip("opt-in batch source owner load")
	}
	runResearchSourceLoad(t, path, true)
}

func runResearchSourceLoad(t *testing.T, path string, batchStudy bool) {
	runResearchSourceCleanupLoad(t, path, batchStudy, false)
}

func TestResearchVerifiedCleanupAccounting(t *testing.T) {
	for _, writes := range []int{0, 4} {
		r := researchGuardLoadCleanupArm(t, "group4postverify", 0, 8, writes, true, true, true, true, true, true)
		checkQueuedGuardLoad(t, r, writes)
		if !r.CombinedSourceCleanup || r.Accepted == 0 || r.Admits == 0 || r.Admits != r.Discards {
			t.Fatal("verified cleanup not exercised")
		}
		for _, p := range r.Phases {
			if p.Accepted && p.VerifiedDiscardNS <= 0 {
				t.Fatal("missing combined timing")
			}
		}
	}
}

func TestResearchVerifiedCleanupExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_VERIFIED_CLEANUP_ARTIFACT")
	if path == "" {
		t.Skip("opt-in verified source cleanup")
	}
	runResearchSourceCleanupLoad(t, path, true, true)
}

func runResearchSourceCleanupLoad(t *testing.T, path string, batchStudy, combinedStudy bool) {
	runResearchResolvedSourceLoad(t, path, batchStudy, combinedStudy, false)
}

func TestResearchResolvedAdmissionAccounting(t *testing.T) {
	for _, writes := range []int{0, 4} {
		r := researchGuardLoadResolvedArm(t, "group4postverify", 0, 8, writes, true, true, true, true, true, true, true)
		checkQueuedGuardLoad(t, r, writes)
		if !r.ResolvedSourceAdmissions || !r.CombinedSourceCleanup || r.Accepted == 0 || r.Admits == 0 || r.Admits != r.Discards {
			t.Fatal("resolved admission path not exercised")
		}
	}
}

func TestResearchResolvedAdmissionExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_RESOLVED_ADMISSION_ARTIFACT")
	if path == "" {
		t.Skip("opt-in resolved source admission study")
	}
	runResearchResolvedSourceLoad(t, path, true, true, true)
}

func runResearchResolvedSourceLoad(t *testing.T, path string, batchStudy, combinedStudy, resolvedStudy bool) {
	runResearchOfferedSourceLoad(t, path, batchStudy, combinedStudy, resolvedStudy, false)
}

func runResearchOfferedSourceLoad(t *testing.T, path string, batchStudy, combinedStudy, resolvedStudy, offeredStudy bool) {
	runResearchWrapperSourceLoad(t, path, batchStudy, combinedStudy, resolvedStudy, offeredStudy, false)
}

func runResearchWrapperSourceLoad(t *testing.T, path string, batchStudy, combinedStudy, resolvedStudy, offeredStudy, idleStudy bool) {
	runResearchGuardOnlySourceLoad(t, path, batchStudy, combinedStudy, resolvedStudy, offeredStudy, idleStudy, false)
}

func runResearchGuardOnlySourceLoad(t *testing.T, path string, batchStudy, combinedStudy, resolvedStudy, offeredStudy, idleStudy, guardStudy bool) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{"go.mod", "go.sum", "internal/service/research_source_load_test.go", "internal/service/research_grouped_consumer_test.go", "internal/service/research_guard_load_test.go", "internal/service/research_queued_load_test.go", "internal/service/research_batch_validation.go", "internal/service/research_group_validation.go", "internal/researchpublicationstore/guard.go", "internal/researchpublicationstore/store.go", "internal/researchmemory/source_owner.go", "internal/researchmemory/durable.go", "internal/researchmemory/durable_prepared.go", "internal/researchmemory/durable_batch_admit.go", "internal/researchmemory/durable_batch_discard.go", "internal/researchmemory/durable_batch_read.go", "internal/researchmemory/record.go", "internal/researchmemory/replay.go", "internal/researchledger/service_identity.go", "internal/researchledger/ledger.go", "internal/researchledger/batch.go", "internal/researchledger/read_batch.go", "docs/experiments/mmm-source-load-v50-protocol.md"} {
		b, e := os.ReadFile("../../" + name)
		if e != nil {
			t.Fatal(e)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	for _, name := range []string{"internal/researchmemory/source_batch_read.go", "internal/researchmemory/source_batch_read_test.go", "internal/researchmemory/source_verified_discard.go", "internal/researchmemory/source_verified_discard_test.go", "internal/researchmemory/durable_discard_encoding_test.go", "docs/experiments/mmm-verified-cleanup-v54-protocol.md", "internal/researchledger/service_read_batch.go", "docs/experiments/mmm-source-batch-load-v53-protocol.md"} {
		b, e := os.ReadFile("../../" + name)
		if e != nil {
			t.Fatal(e)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	arms := 3
	if resolvedStudy {
		if !combinedStudy || !batchStudy {
			t.Fatal("resolved study requires combined four-arm comparison")
		}
		for _, name := range []string{"internal/researchmemory/source_resolved_admit.go", "internal/researchmemory/source_resolved_admit_test.go", "docs/experiments/mmm-resolved-admission-v55-protocol.md"} {
			b, e := os.ReadFile("../../" + name)
			if e != nil {
				t.Fatal(e)
			}
			sources[name] = string(b)
			h := sha256.Sum256(b)
			hashes[name] = hex.EncodeToString(h[:])
		}
	}
	if combinedStudy && !batchStudy {
		t.Fatal("combined study requires four-arm batch comparison")
	}
	if batchStudy {
		arms = 4
	}
	schedules := []*researchLoadSchedule{nil}
	if offeredStudy {
		if !resolvedStudy {
			t.Fatal("offered comparison requires resolved arms")
		}
		schedules = []*researchLoadSchedule{{2 * time.Millisecond, 4 * time.Millisecond}, {5 * time.Millisecond, 10 * time.Millisecond}, {10 * time.Millisecond, 20 * time.Millisecond}}
		for _, name := range []string{"internal/service/research_offered_load_test.go", "docs/experiments/mmm-offered-load-v57-protocol.md"} {
			b, e := os.ReadFile("../../" + name)
			if e != nil {
				t.Fatal(e)
			}
			sources[name] = string(b)
			h := sha256.Sum256(b)
			hashes[name] = hex.EncodeToString(h[:])
		}
	}
	if idleStudy {
		if !offeredStudy {
			t.Fatal("idle wrapper study requires offered schedule")
		}
		arms = 5
		schedules = []*researchLoadSchedule{{5 * time.Millisecond, 10 * time.Millisecond}}
		for _, name := range []string{"internal/service/research_idle_wrapper_test.go", "internal/service/service.go", "internal/model/api.go", "internal/service/research_frontier.go", "internal/researchpublicationstore/mutations.go", "docs/experiments/mmm-idle-wrapper-v58-protocol.md"} {
			b, e := os.ReadFile("../../" + name)
			if e != nil {
				t.Fatal(e)
			}
			sources[name] = string(b)
			h := sha256.Sum256(b)
			hashes[name] = hex.EncodeToString(h[:])
		}
	}
	enc := json.NewEncoder(f)
	if guardStudy {
		if !idleStudy {
			t.Fatal("guard-only comparison requires idle controls")
		}
		arms = 6
		for _, name := range []string{"internal/service/research_guard_only_test.go", "docs/experiments/mmm-guard-only-v59-protocol.md"} {
			b, e := os.ReadFile("../../" + name)
			if e != nil {
				t.Fatal(e)
			}
			sources[name] = string(b)
			h := sha256.Sum256(b)
			hashes[name] = hex.EncodeToString(h[:])
		}
	}
	if err = enc.Encode(map[string]any{"kind": "header", "guard_only_study": guardStudy, "idle_wrapper_study": idleStudy, "offered_arrival_study": offeredStudy, "schedules": schedules, "resolved_admission_study": resolvedStudy, "combined_cleanup_study": combinedStudy, "batch_source_read_study": batchStudy, "expected_arms": arms * 3 * len(schedules), "Sources": sources, "Hashes": hashes, "GoVersion": runtime.Version(), "GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH, "GOMAXPROCS": runtime.GOMAXPROCS(0), "wait_budget_ms": 20}); err != nil {
		t.Fatal(err)
	}
	for _, schedule := range schedules {
		for trial := 0; trial < 3; trial++ {
			for order := 0; order < arms; order++ {
				arm := (trial + order) % arms
				idle := idleStudy && arm == 1
				guardOnly := guardStudy && arm == 2
				if guardStudy && arm >= 2 {
					arm--
				}
				if idleStudy && arm > 0 {
					arm--
				}
				mode := "group4postverify"
				if arm == 0 {
					mode = "off"
				}
				if guardOnly {
					mode = "group4"
				}
				r := researchGuardLoadWrapperArm(t, mode, trial, 192, 96, true, arm != 0, arm != 0 || guardOnly, arm >= 2, arm == 3 || (combinedStudy && arm == 2), combinedStudy && (arm == 3 || (resolvedStudy && arm == 2)), resolvedStudy && arm == 3, schedule, idle)
				if err = enc.Encode(r); err != nil {
					t.Fatal(err)
				}
				if err = f.Sync(); err != nil {
					t.Fatal(err)
				}
				checkQueuedGuardLoad(t, r, 96)
				if guardOnly {
					checkResearchGuardOnly(t, r)
				}
				if idleStudy {
					t.Logf("idle_wrapper%t", r.IdlePublicationWrapper)
				}
				if offeredStudy {
					if err = researchCheckOffered(r, 96); err != nil {
						t.Fatal(err)
					}
					t.Logf("offered_read_ms%d", r.OfferedReadEveryNS/int64(time.Millisecond))
				}
				t.Logf("trial%d %s source%t batch%t combined%t resolved%t accepted%d dropped%d expired%d admits%d discards%d", trial, mode, r.SourceOwner, r.BatchSourceReads, r.CombinedSourceCleanup, r.ResolvedSourceAdmissions, r.Accepted, r.Dropped, r.WaitExpired, r.Admits, r.Discards)
			}
		}
	}
}
