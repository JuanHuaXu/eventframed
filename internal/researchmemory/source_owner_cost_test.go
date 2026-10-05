package researchmemory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"
)

type sourceOwnerCostResult struct {
	BatchDiscard                                    bool
	Kind                                            string
	Size, Trial, Cycles, Originals, TerminalRecords int
	SourceOwner                                     bool
	AdmitNS, RetryNS, LookupNS, DiscardNS, CycleNS  []int64
	RestartNS, Bytes                                int64
}

func sourceOwnerCostArm(t *testing.T, size, trial int, source bool) sourceOwnerCostResult {
	return sourceOwnerCostVariantArm(t, size, trial, source, false)
}

func sourceOwnerCostVariantArm(t *testing.T, size, trial int, source, batchDiscard bool) sourceOwnerCostResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	path := t.TempDir() + "/cost.sqlite"
	r := sourceOwnerCostResult{Kind: "result", Size: size, Trial: trial, Cycles: 32, SourceOwner: source, BatchDiscard: batchDiscard}
	if batchDiscard && !source {
		t.Fatal("source batch discard requires source owner")
	}
	var owner *SourceOwner
	var control *Durable
	var err error
	open := func() {
		if source {
			owner, err = OpenSourceOwner(ctx, path, "tenant", "stream", 1, 42)
		} else {
			control, err = OpenDurablePreparedBatches(ctx, path, "tenant", "stream", 1, 42)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	closeOwner := func() {
		if source {
			err = owner.Close()
		} else {
			err = control.Close()
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	open()
	defer closeOwner()
	var first, last RecordedPrediction
	for cycle := 0; cycle < r.Cycles; cycle++ {
		requests := make([]SourceAdmissionRequest, size)
		legacy := make([]AdmissionRequest, size)
		ids := make([]uint64, size)
		terminals := make([]DiscardRequest, size)
		sourceTerminals := make([]SourceDiscardRequest, size)
		for i := range requests {
			id := uint64(cycle*size + i + 1)
			req := sourceRequest()
			req.Features = uint16(id % 512)
			req.Binding.JournalID = fmt.Sprintf("public-journal-%d", cycle)
			req.Binding.EventID = fmt.Sprintf("public-event-%d", i)
			requests[i] = req
			legacy[i] = AdmissionRequest{ID: id, Features: req.Features, Baseline: req.Baseline, At: req.At, Binding: &requests[i].Binding}
			ids[i] = id
			terminals[i] = DiscardRequest{ID: id, Available: req.At}
			sourceTerminals[i] = SourceDiscardRequest{req.Binding.JournalID, req.Binding.EventID, req.At}
		}
		admit := func() ([]AdmissionResult, error) {
			if source {
				return owner.Admit(ctx, requests)
			}
			return control.AdmitBatchWithSnapshotReads(ctx, legacy)
		}
		start := time.Now()
		originals, e := admit()
		r.AdmitNS = append(r.AdmitNS, time.Since(start).Nanoseconds())
		if e != nil || len(originals) != size {
			t.Fatal("admission", e)
		}
		start = time.Now()
		retries, e := admit()
		r.RetryNS = append(r.RetryNS, time.Since(start).Nanoseconds())
		if e != nil || len(retries) != size {
			t.Fatal("retry", e)
		}
		start = time.Now()
		var readback []RecordedPrediction
		if source {
			readback = make([]RecordedPrediction, size)
			for i, req := range requests {
				readback[i], e = owner.Lookup(ctx, req.Binding.JournalID, req.Binding.EventID)
				if e != nil {
					t.Fatal("source readback", e)
				}
			}
		} else {
			readback, e = control.Admissions(ctx, ids)
		}
		r.LookupNS = append(r.LookupNS, time.Since(start).Nanoseconds())
		if e != nil || len(readback) != size {
			t.Fatal("readback", e)
		}
		for i, v := range originals {
			if v.Retry || v.Record.Prediction.ID != ids[i] || !retries[i].Retry || !reflect.DeepEqual(v.Record, retries[i].Record) || !reflect.DeepEqual(v.Record, readback[i]) {
				t.Fatal("original/retry mismatch", cycle, i)
			}
		}
		if cycle == 0 {
			first = originals[0].Record
		}
		last = originals[size-1].Record
		start = time.Now()
		var discarded []bool
		if batchDiscard {
			discarded, e = owner.DiscardBatch(ctx, sourceTerminals)
		} else if source {
			discarded = make([]bool, size)
			for i, req := range requests {
				discarded[i], e = owner.Discard(ctx, req.Binding.JournalID, req.Binding.EventID, req.At)
				if e != nil {
					t.Fatal("source discard", e)
				}
			}
		} else {
			discarded, e = control.DiscardBatchWithSnapshotReads(ctx, terminals)
		}
		r.DiscardNS = append(r.DiscardNS, time.Since(start).Nanoseconds())
		if e != nil || len(discarded) != size {
			t.Fatal("discard", e)
		}
		for _, retry := range discarded {
			if retry {
				t.Fatal("unexpected terminal retry")
			}
		}
		r.CycleNS = append(r.CycleNS, r.AdmitNS[cycle]+r.RetryNS[cycle]+r.LookupNS[cycle]+r.DiscardNS[cycle])
		r.Originals += size
		r.TerminalRecords += size
	}
	closeOwner()
	stat, e := os.Stat(path)
	if e != nil {
		t.Fatal(e)
	}
	r.Bytes = stat.Size()
	start := time.Now()
	open()
	r.RestartNS = time.Since(start).Nanoseconds()
	var worker *Background
	if source {
		worker = owner.d.worker
	} else {
		worker = control.worker
	}
	if n, f, p, q := worker.Counts(); n != 0 || f != 0 || p != 0 || q != 0 {
		t.Fatal("invented evidence or lost terminal", n, f, p, q)
	}
	if worker.next != uint64(r.Originals) {
		t.Fatal("duplicate/lost original", worker.next, r.Originals)
	}
	for _, want := range []RecordedPrediction{first, last} {
		var got RecordedPrediction
		if source {
			got, e = owner.Lookup(ctx, want.Binding.JournalID, want.Binding.EventID)
		} else {
			got, e = control.Admission(ctx, want.Prediction.ID)
		}
		if e != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("reopened original mismatch", e)
		}
	}
	return r
}

func TestSourceOwnerCostStudy(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SOURCE_OWNER_COST_ARTIFACT")
	if path == "" {
		t.Skip("opt-in source owner cost study")
	}
	runSourceCostStudy(t, path, false)
}

func TestSourceDiscardCostStudy(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SOURCE_DISCARD_COST_ARTIFACT")
	if path == "" {
		t.Skip("opt-in source batch discard study")
	}
	runSourceCostStudy(t, path, true)
}

func runSourceCostStudy(t *testing.T, path string, batchStudy bool) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{
		"docs/experiments/mmm-source-owner-cost-v48-protocol.md",
		"docs/experiments/mmm-source-discard-cost-v49-protocol.md", "internal/researchmemory/source_discard_test.go",
		"internal/researchmemory/source_owner_cost_test.go", "internal/researchmemory/source_owner_test.go",
		"internal/researchmemory/source_owner.go", "internal/researchmemory/durable.go",
		"internal/researchmemory/durable_prepared.go", "internal/researchmemory/durable_batch_admit.go",
		"internal/researchmemory/durable_batch_read.go", "internal/researchmemory/durable_batch_discard.go",
		"internal/researchmemory/durable_discard.go", "internal/researchmemory/replay.go",
		"internal/researchmemory/record.go", "internal/researchmemory/background.go",
		"internal/researchmemory/adapter.go", "internal/researchmemory/frozen.go",
		"internal/researchledger/service_identity.go", "internal/researchledger/ledger.go",
		"internal/researchledger/batch.go", "internal/researchledger/read_batch.go", "go.mod", "go.sum",
	} {
		b, e := os.ReadFile("../../" + name)
		if e != nil {
			t.Fatal(e)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if err = enc.Encode(map[string]any{"Kind": "header", "BatchDiscardStudy": batchStudy, "ExpectedCells": 12, "CyclesPerCell": 32, "Go": runtime.Version(), "OS": runtime.GOOS, "Arch": runtime.GOARCH, "CPUs": runtime.NumCPU(), "GOMAXPROCS": runtime.GOMAXPROCS(0), "Sources": sources, "Hashes": hashes}); err != nil {
		t.Fatal(err)
	}
	for _, size := range []int{50, 200} {
		for trial := 0; trial < 3; trial++ {
			for order := 0; order < 2; order++ {
				var r sourceOwnerCostResult
				if batchStudy {
					r = sourceOwnerCostVariantArm(t, size, trial, true, (trial+order)%2 == 1)
				} else {
					r = sourceOwnerCostArm(t, size, trial, (trial+order)%2 == 1)
				}
				if err = enc.Encode(r); err != nil {
					t.Fatal(err)
				}
				if err = f.Sync(); err != nil {
					t.Fatal(err)
				}
				t.Logf("size%d trial%d source%t batchDiscard%t originals%d restart_ms%.3f", size, trial, r.SourceOwner, r.BatchDiscard, r.Originals, float64(r.RestartNS)/1e6)
			}
		}
	}
}
