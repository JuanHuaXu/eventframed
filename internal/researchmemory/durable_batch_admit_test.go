package researchmemory

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

type singleOnlyLog struct{ durableLog }

func TestDurableBatchWarmOriginals(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	path := t.TempDir() + "/warm.sqlite"
	d, err := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	for id := uint64(1); id <= 64; id++ {
		if _, _, err = d.Admit(ctx, id, uint16(id), .6, now); err != nil {
			t.Fatal(err)
		}
		if _, err = d.Feedback(ctx, id, id%3 == 0, now); err != nil {
			t.Fatal(err)
		}
	}
	wait, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = d.worker.WaitProcessed(wait, 64); err != nil {
		t.Fatal(err)
	}
	requests := make([]AdmissionRequest, 128)
	for i := range requests {
		requests[i] = AdmissionRequest{ID: uint64(65 + i), Features: uint16(i), Baseline: .6, At: now}
	}
	results, err := d.AdmitBatch(ctx, requests)
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range results {
		if r.Retry || !r.Record.Ready {
			t.Fatal("warm original missing", i)
		}
		owned, err := d.worker.Record(requests[i].ID)
		if err != nil || !reflect.DeepEqual(owned, r.Record) {
			t.Fatal("owned original changed", i, err)
		}
		saved, err := d.Admission(ctx, requests[i].ID)
		if err != nil || !reflect.DeepEqual(saved, r.Record) {
			t.Fatal("stored original changed", i, err)
		}
	}
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	d, err = OpenDurable(ctx, path, "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	retried, err := d.AdmitBatch(ctx, requests)
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range retried {
		if !r.Retry || !reflect.DeepEqual(r.Record, results[i].Record) {
			t.Fatal("replay recomputed original", i)
		}
	}
	if n, f, p, q := d.worker.Counts(); n != 64 || f != 0 || p != 128 || q != 0 {
		t.Fatal("wrong warm replay", n, f, p, q)
	}
}

func TestDurableBatchAdmissionFailureRecovery(t *testing.T) {
	for _, commit := range []bool{false, true} {
		for _, panics := range []bool{false, true} {
			t.Run(fmt.Sprintf("commit=%t/panic=%t", commit, panics), func(t *testing.T) {
				ctx := context.Background()
				now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
				path := t.TempDir() + "/failure.sqlite"
				d, err := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
				if err != nil {
					t.Fatal(err)
				}
				requests := []AdmissionRequest{{ID: 1, Features: 1, Baseline: .6, At: now}, {ID: 2, Features: 2, Baseline: .6, At: now}}
				d.log = faultyBatchLog{d.log, commit, panics}
				func() {
					defer func() {
						r := recover()
						if (r != nil) != panics {
							t.Error("panic mismatch", r)
						}
					}()
					result, err := d.AdmitBatch(ctx, requests)
					if result != nil || err == nil {
						t.Error("failed batch acknowledged", result, err)
					}
				}()
				if !d.stopped {
					t.Fatal("uncertain owner remained active")
				}
				originals := make([]RecordedPrediction, 2)
				for i := range originals {
					originals[i], err = d.worker.Record(uint64(i + 1))
					if err != nil {
						t.Fatal(err)
					}
				}
				if _, err = d.AdmitBatch(ctx, requests); err == nil {
					t.Fatal("continued uncertain owner")
				}
				if err = d.Close(); err != nil {
					t.Fatal(err)
				}
				d, err = OpenDurable(ctx, path, "tenant", "stream", 1, 42)
				if err != nil {
					t.Fatal(err)
				}
				defer d.Close()
				_, _, pending, _ := d.worker.Counts()
				want := 0
				if commit {
					want = 2
				}
				if pending != want {
					t.Fatal("partial recovery", pending, want)
				}
				results, err := d.AdmitBatch(ctx, requests)
				if err != nil {
					t.Fatal(err)
				}
				for i, r := range results {
					if r.Retry != commit || !reflect.DeepEqual(r.Record, originals[i]) {
						t.Fatal("retry changed original", i)
					}
				}
			})
		}
	}
}

func TestDurableBatchAdmissionPreflight(t *testing.T) {
	for _, kind := range []string{"empty", "count", "duplicate", "gap", "features", "baseline", "time", "unsupported", "cancel"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
			d, err := OpenDurable(ctx, t.TempDir()+"/reject.sqlite", "tenant", "stream", 1, 42)
			if err != nil {
				t.Fatal(err)
			}
			defer d.Close()
			requests := []AdmissionRequest{{ID: 1, Features: 1, Baseline: .6, At: now}, {ID: 2, Features: 2, Baseline: .6, At: now}}
			switch kind {
			case "empty":
				requests = nil
			case "count":
				requests = make([]AdmissionRequest, 257)
			case "duplicate":
				requests[1].ID = 1
			case "gap":
				requests[1].ID = 3
			case "features":
				requests[1].Features = 512
			case "baseline":
				requests[1].Baseline = math.NaN()
			case "time":
				requests[1].At = time.Time{}
			case "unsupported":
				d.log = singleOnlyLog{d.log}
			case "cancel":
				cancel()
			}
			if results, err := d.AdmitBatch(ctx, requests); err == nil || results != nil {
				t.Fatal("bad preflight accepted", err)
			}
			if d.stopped || d.worker.next != 0 || len(d.worker.pending) != 0 {
				t.Fatal("preflight mutated worker")
			}
		})
	}
}

func TestDurableBatchBindingCapacityAndMixedRetry(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	d, err := OpenDurable(ctx, t.TempDir()+"/bindings.sqlite", "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	binding := ServiceBinding{Tenant: "tenant", JournalID: "journal", EventID: "public-event", Snapshot: model.Snapshot{RuntimeVersion: 1, ContractVersion: 1}}
	request := AdmissionRequest{ID: 1, Features: 1, Baseline: .6, At: now, Binding: &binding}
	results, err := d.AdmitBatch(ctx, []AdmissionRequest{request})
	if err != nil {
		t.Fatal(err)
	}
	binding.EventID = "mutated-input"
	if results[0].Record.Binding.EventID != "public-event" {
		t.Fatal("input aliases returned original")
	}
	results[0].Record.Binding.EventID = "mutated-output"
	saved, err := d.Admission(ctx, 1)
	if err != nil || saved.Binding.EventID != "public-event" {
		t.Fatal("output aliases stored original", err)
	}
	request.Binding = saved.Binding
	newRequest := AdmissionRequest{ID: 2, Features: 2, Baseline: .6, At: now}
	badRetry := request
	badRetry.Baseline = .7
	if _, err = d.AdmitBatch(ctx, []AdmissionRequest{newRequest, badRetry}); err == nil {
		t.Fatal("conflicting late retry accepted")
	}
	if d.stopped || d.worker.next != 1 {
		t.Fatal("late preflight staged partial state")
	}
	results, err = d.AdmitBatch(ctx, []AdmissionRequest{request, newRequest})
	if err != nil || !results[0].Retry || results[1].Retry {
		t.Fatal("mixed retry mismatch", err)
	}
	remaining := make([]AdmissionRequest, 254)
	for i := range remaining {
		remaining[i] = AdmissionRequest{ID: uint64(i + 3), Features: uint16(i), Baseline: .6, At: now}
	}
	if _, err = d.AdmitBatch(ctx, remaining); err != nil {
		t.Fatal("exact pending capacity rejected", err)
	}
	if _, err = d.AdmitBatch(ctx, []AdmissionRequest{{ID: 257, Features: 1, Baseline: .6, At: now}}); err == nil {
		t.Fatal("pending capacity exceeded")
	}
	if d.stopped || d.worker.next != 256 {
		t.Fatal("capacity rejection mutated owner")
	}
	discards := make([]DiscardRequest, 256)
	for i := range discards {
		discards[i] = DiscardRequest{uint64(i + 1), now}
	}
	if _, err = d.DiscardBatch(ctx, discards[:1]); err != nil {
		t.Fatal(err)
	}
	retries, err := d.DiscardBatch(ctx, discards)
	if err != nil || !retries[0] {
		t.Fatal("mixed discard retry", err)
	}
	for _, retry := range retries[1:] {
		if retry {
			t.Fatal("new discard marked retry")
		}
	}
}
