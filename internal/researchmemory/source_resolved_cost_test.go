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
	"strings"
	"testing"
	"time"
)

type resolvedAdmissionCostResult struct {
	ExclusiveLocking                                          bool
	BulkAdmissions                                            bool
	ConditionalInsert                                         bool
	AdmitPhases, RetryPhases                                  []admissionPhases
	Kind                                                      string
	Size, Trial, Cycles, TrainingLabels, Originals, Terminals int
	Resolved                                                  bool
	TrainingHash, OriginalHash                                string
	AdmitNS, RetryNS, VerifiedDiscardNS                       []int64
	RestartNS, Bytes                                          int64
}

func resolvedAdmissionCostArm(t *testing.T, size, trial, cycles int, warm, resolved bool) resolvedAdmissionCostResult {
	return resolvedAdmissionMeasuredArm(t, size, trial, cycles, warm, resolved, false)
}

func resolvedAdmissionMeasuredArm(t *testing.T, size, trial, cycles int, warm, resolved, measure bool) resolvedAdmissionCostResult {
	return resolvedAdmissionStrategyArm(t, size, trial, cycles, warm, resolved, measure, false)
}

func resolvedAdmissionStrategyArm(t *testing.T, size, trial, cycles int, warm, resolved, measure, conditional bool) resolvedAdmissionCostResult {
	return resolvedAdmissionBulkArm(t, size, trial, cycles, warm, resolved, measure, conditional, false)
}

func resolvedAdmissionBulkArm(t *testing.T, size, trial, cycles int, warm, resolved, measure, conditional, bulk bool) resolvedAdmissionCostResult {
	return resolvedAdmissionLockArm(t, size, trial, cycles, warm, resolved, measure, conditional, bulk, false)
}

func resolvedAdmissionLockArm(t *testing.T, size, trial, cycles int, warm, resolved, measure, conditional, bulk, exclusive bool) resolvedAdmissionCostResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	path := t.TempDir() + "/cost.sqlite"
	open := OpenSourceOwnerBatchReads
	if resolved {
		open = OpenSourceOwnerResolvedAdmissions
	}
	if conditional {
		if !resolved {
			t.Fatal("conditional study requires resolved owner")
		}
		open = OpenSourceOwnerConditionalInsert
	}
	if bulk {
		if !resolved || conditional {
			t.Fatal("invalid bulk arm")
		}
		open = OpenSourceOwnerBulkAdmissions
	}
	if exclusive {
		if !resolved || conditional || bulk {
			t.Fatal("invalid locking arm")
		}
		open = OpenSourceOwnerExclusive
	}
	o, err := open(ctx, path, "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if o != nil {
			o.Close()
		}
	}()
	r := resolvedAdmissionCostResult{Kind: "result", Size: size, Trial: trial, Cycles: cycles, Resolved: resolved}
	r.ConditionalInsert = conditional
	r.BulkAdmissions = bulk
	r.ExclusiveLocking = exclusive
	if warm {
		r.TrainingLabels = 64
	}
	trainingHash, originalHash := sha256.New(), sha256.New()
	trainingEncoder, originalEncoder := json.NewEncoder(trainingHash), json.NewEncoder(originalHash)
	for i := 1; i <= r.TrainingLabels; i++ {
		req := sourceRequest()
		req.Features = uint16(i)
		req.At = req.At.Add(time.Duration(i) * time.Second)
		req.Binding.JournalID = "public-training"
		req.Binding.EventID = fmt.Sprint(i)
		got, e := o.Admit(ctx, []SourceAdmissionRequest{req})
		if e != nil || len(got) != 1 || got[0].Retry || got[0].Record.Prediction.ID != uint64(i) {
			t.Fatal("training admission", got, e)
		}
		if e = trainingEncoder.Encode(got[0].Record); e != nil {
			t.Fatal(e)
		}
		// Test-only synthetic labels, never a new public feedback authority.
		if _, e = o.d.Feedback(ctx, uint64(i), i%3 == 0, req.At.Add(time.Nanosecond)); e != nil {
			t.Fatal(e)
		}
		if e = o.d.worker.WaitProcessed(ctx, uint64(i)); e != nil {
			t.Fatal(e)
		}
	}
	var first, last RecordedPrediction
	var lastCleanup []VerifiedSourceDiscardRequest
	for cycle := 0; cycle < cycles; cycle++ {
		requests := make([]SourceAdmissionRequest, size)
		for i := range requests {
			req := sourceRequest()
			req.Features = uint16((cycle*size + i) % 512)
			req.At = req.At.Add(time.Hour + time.Duration(cycle)*time.Second)
			req.Binding.JournalID = fmt.Sprintf("public-measured-%d", cycle)
			req.Binding.EventID = fmt.Sprint(i)
			requests[i] = req
		}
		if measure {
			o.d.measurement = &admissionPhases{}
		}
		start := time.Now()
		originals, e := o.Admit(ctx, requests)
		r.AdmitNS = append(r.AdmitNS, time.Since(start).Nanoseconds())
		if measure {
			r.AdmitPhases = append(r.AdmitPhases, *o.d.measurement)
			o.d.measurement = &admissionPhases{}
		}
		if e != nil || len(originals) != size {
			t.Fatal("admit", e)
		}
		start = time.Now()
		retries, e := o.Admit(ctx, requests)
		r.RetryNS = append(r.RetryNS, time.Since(start).Nanoseconds())
		if measure {
			r.RetryPhases = append(r.RetryPhases, *o.d.measurement)
			o.d.measurement = nil
		}
		if e != nil || len(retries) != size {
			t.Fatal("retry", e)
		}
		cleanup := make([]VerifiedSourceDiscardRequest, size)
		for i, v := range originals {
			id := uint64(r.TrainingLabels + cycle*size + i + 1)
			if v.Retry || !retries[i].Retry || !reflect.DeepEqual(v.Record, retries[i].Record) || v.Record.Prediction.ID != id || v.Record.Ready != warm {
				t.Fatal("original/retry mismatch", cycle, i)
			}
			if e = originalEncoder.Encode(v.Record); e != nil {
				t.Fatal(e)
			}
			cleanup[i] = VerifiedSourceDiscardRequest{Original: v.Record, Available: v.Record.At}
		}
		if cycle == 0 {
			first = originals[0].Record
		}
		last = originals[size-1].Record
		start = time.Now()
		terminal, e := o.VerifyAndDiscardBatch(ctx, cleanup)
		r.VerifiedDiscardNS = append(r.VerifiedDiscardNS, time.Since(start).Nanoseconds())
		if e != nil || len(terminal) != size {
			t.Fatal("verified cleanup", e)
		}
		for _, retry := range terminal {
			if retry {
				t.Fatal("unexpected terminal retry")
			}
		}
		lastCleanup = cleanup
		r.Originals += size
		r.Terminals += size
	}
	r.TrainingHash = hex.EncodeToString(trainingHash.Sum(nil))
	r.OriginalHash = hex.EncodeToString(originalHash.Sum(nil))
	check := func() {
		n, f, p, q := o.d.worker.Counts()
		if n != uint64(r.TrainingLabels) || f != 0 || p != 0 || q != 0 || o.d.worker.next != uint64(r.TrainingLabels+r.Originals) {
			t.Fatal("lifecycle counts", n, f, p, q, o.d.worker.next)
		}
	}
	check()
	if err = o.Close(); err != nil {
		t.Fatal(err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	r.Bytes = stat.Size()
	start := time.Now()
	o, err = open(ctx, path, "tenant", "stream", 1, 42)
	r.RestartNS = time.Since(start).Nanoseconds()
	if err != nil {
		t.Fatal(err)
	}
	check()
	for _, want := range []RecordedPrediction{first, last} {
		got, e := o.Lookup(ctx, want.Binding.JournalID, want.Binding.EventID)
		if e != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("replayed original", e)
		}
	}
	terminal, err := o.VerifyAndDiscardBatch(ctx, lastCleanup)
	if err != nil || len(terminal) != size {
		t.Fatal("replayed terminal", err)
	}
	for _, retry := range terminal {
		if !retry {
			t.Fatal("lost terminal")
		}
	}
	check()
	return r
}

func TestResolvedAdmissionCostAccounting(t *testing.T) {
	for _, warm := range []bool{false, true} {
		a := resolvedAdmissionCostArm(t, 5, 0, 2, warm, false)
		b := resolvedAdmissionCostArm(t, 5, 0, 2, warm, true)
		if a.OriginalHash != b.OriginalHash || a.TrainingHash != b.TrainingHash || a.Originals != 10 || b.Terminals != 10 {
			t.Fatal("unmatched experiment histories")
		}
	}
}

func TestResolvedAdmissionCostExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_RESOLVED_COST_ARTIFACT")
	if path == "" {
		t.Skip("opt-in isolated resolved admission cost")
	}
	runResolvedAdmissionCost(t, path, false)
}

func TestResearchAdmissionPhasesAccounting(t *testing.T) {
	for _, warm := range []bool{false, true} {
		a := resolvedAdmissionMeasuredArm(t, 5, 0, 2, warm, true, false)
		b := resolvedAdmissionMeasuredArm(t, 5, 0, 2, warm, true, true)
		if a.OriginalHash != b.OriginalHash || a.TrainingHash != b.TrainingHash {
			t.Fatal("measurement changed records")
		}
		for i, p := range b.AdmitPhases {
			if p.SourceNS <= 0 || p.PreflightNS <= 0 || p.StageNS <= 0 || p.AppendNS <= 0 || p.SourceNS+p.PreflightNS+p.StageNS+p.AppendNS > b.AdmitNS[i] {
				t.Fatal("admission phase containment")
			}
		}
		for i, p := range b.RetryPhases {
			if p.SourceNS <= 0 || p.PreflightNS <= 0 || p.StageNS <= 0 || p.AppendNS <= 0 || p.SourceNS+p.PreflightNS+p.StageNS+p.AppendNS > b.RetryNS[i] {
				t.Fatal("retry phase containment")
			}
		}
	}
}

func TestResearchAdmissionPhasesExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ADMISSION_PHASES_ARTIFACT")
	if path == "" {
		t.Skip("opt-in admission phase measurement")
	}
	runResolvedAdmissionCost(t, path, true)
}

func runResolvedAdmissionCost(t *testing.T, path string, measure bool) {
	runResolvedAdmissionStrategyCost(t, path, measure, false, false)
}

func TestConditionalAdmissionAccounting(t *testing.T) {
	for _, warm := range []bool{false, true} {
		a := resolvedAdmissionStrategyArm(t, 5, 0, 2, warm, true, false, false)
		b := resolvedAdmissionStrategyArm(t, 5, 0, 2, warm, true, false, true)
		if a.OriginalHash != b.OriginalHash || a.TrainingHash != b.TrainingHash || a.Originals != b.Originals || a.Terminals != b.Terminals {
			t.Fatal("conditional history mismatch")
		}
	}
}

func TestConditionalAdmissionCostExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_CONDITIONAL_COST_ARTIFACT")
	if path == "" {
		t.Skip("opt-in conditional insert cost")
	}
	runResolvedAdmissionStrategyCost(t, path, false, true, false)
}

func TestBulkAdmissionAccounting(t *testing.T) {
	for _, warm := range []bool{false, true} {
		a := resolvedAdmissionBulkArm(t, 5, 0, 2, warm, true, false, false, false)
		b := resolvedAdmissionBulkArm(t, 5, 0, 2, warm, true, false, false, true)
		if a.OriginalHash != b.OriginalHash || a.TrainingHash != b.TrainingHash {
			t.Fatal("bulk history mismatch")
		}
	}
}

func TestBulkAdmissionCostExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_BULK_COST_ARTIFACT")
	if path == "" {
		t.Skip("opt-in bulk admission cost")
	}
	runResolvedAdmissionStrategyCost(t, path, false, false, true)
}

func runResolvedAdmissionStrategyCost(t *testing.T, path string, measure, conditionalStudy, bulkStudy bool) {
	runResolvedAdmissionLockCost(t, path, measure, conditionalStudy, bulkStudy, false)
}

func TestExclusiveSourceAccounting(t *testing.T) {
	for _, warm := range []bool{false, true} {
		a := resolvedAdmissionLockArm(t, 5, 0, 2, warm, true, false, false, false, false)
		b := resolvedAdmissionLockArm(t, 5, 0, 2, warm, true, false, false, false, true)
		if a.OriginalHash != b.OriginalHash || a.TrainingHash != b.TrainingHash || a.Originals != b.Originals || a.Terminals != b.Terminals {
			t.Fatal("exclusive parity")
		}
	}
}

func TestExclusiveSourceCostExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_EXCLUSIVE_COST_ARTIFACT")
	if path == "" {
		t.Skip("opt-in exclusive ledger study")
	}
	runResolvedAdmissionLockCost(t, path, false, false, false, true)
}

func runResolvedAdmissionLockCost(t *testing.T, path string, measure, conditionalStudy, bulkStudy, exclusiveStudy bool) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	files := []string{"go.mod", "go.sum", "docs/experiments/mmm-resolved-cost-v56-protocol.md", "internal/researchmemory/source_resolved_cost_test.go", "internal/researchmemory/source_owner_test.go"}
	if measure {
		files = append(files, "docs/experiments/mmm-admission-phases-v65-protocol.md")
	}
	if conditionalStudy {
		files = append(files, "docs/experiments/mmm-conditional-insert-v66-protocol.md", "internal/researchledger/conditional_insert_test.go", "internal/researchledger/service_identity_test.go", "internal/researchmemory/source_conditional_insert_test.go")
	}
	if bulkStudy {
		files = append(files, "docs/experiments/mmm-bulk-admission-v67-protocol.md", "internal/researchledger/bulk_admission_test.go", "internal/researchmemory/source_bulk_admission_test.go")
	}
	if exclusiveStudy {
		files = append(files, "docs/experiments/mmm-exclusive-ledger-v69-protocol.md", "internal/researchledger/exclusive_test.go")
	}
	for _, dir := range []string{"internal/researchmemory", "internal/researchledger"} {
		entries, e := os.ReadDir("../../" + dir)
		if e != nil {
			t.Fatal(e)
		}
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
				files = append(files, dir+"/"+entry.Name())
			}
		}
	}
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range files {
		b, e := os.ReadFile("../../" + name)
		if e != nil {
			t.Fatal(e)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	orders := 2
	if measure {
		orders = 1
	}
	if err = enc.Encode(map[string]any{"Kind": "header", "MeasuredPhases": measure, "ExpectedCells": 12 * orders, "Cycles": 32, "Go": runtime.Version(), "OS": runtime.GOOS, "Arch": runtime.GOARCH, "CPUs": runtime.NumCPU(), "GOMAXPROCS": runtime.GOMAXPROCS(0), "Sources": sources, "Hashes": hashes}); err != nil {
		t.Fatal(err)
	}
	for _, size := range []int{50, 200} {
		for _, warm := range []bool{false, true} {
			for trial := 0; trial < 3; trial++ {
				var pair []resolvedAdmissionCostResult
				for order := 0; order < orders; order++ {
					r := resolvedAdmissionLockArm(t, size, trial, 32, warm, exclusiveStudy || bulkStudy || conditionalStudy || measure || (trial+order)%2 == 1, measure, conditionalStudy && (trial+order)%2 == 1, bulkStudy && (trial+order)%2 == 1, exclusiveStudy && (trial+order)%2 == 1)
					if err = enc.Encode(r); err != nil {
						t.Fatal(err)
					}
					if err = f.Sync(); err != nil {
						t.Fatal(err)
					}
					pair = append(pair, r)
					t.Logf("size%d warm%t trial%d resolved%t originals%d terminals%d", size, warm, trial, r.Resolved, r.Originals, r.Terminals)
				}
				if len(pair) == 2 && (pair[0].TrainingHash != pair[1].TrainingHash || pair[0].OriginalHash != pair[1].OriginalHash) {
					t.Fatal("unmatched original histories")
				}
			}
		}
	}
}
