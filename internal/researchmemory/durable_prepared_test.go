package researchmemory

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestDurablePreparedWarmReplayParity(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	var want []AdmissionResult
	var trainingControl []RecordedPrediction
	for _, prepared := range []bool{false, true} {
		open := OpenDurable
		if prepared {
			open = OpenDurablePreparedBatches
		}
		path := t.TempDir() + "/warm.sqlite"
		d, err := open(ctx, path, "tenant", "stream", 1, 42)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := d.log.(preparedBatchLedger); ok != prepared {
			t.Fatal("wrong ledger method")
		}
		for id := uint64(1); id <= 64; id++ {
			if _, _, err = d.Admit(ctx, id, uint16(id), .6, now); err != nil {
				t.Fatal(err)
			}
			r, err := d.Admission(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if !prepared {
				trainingControl = append(trainingControl, r)
			} else if !reflect.DeepEqual(r, trainingControl[id-1]) {
				t.Fatalf("training original %d differs under matched update schedule", id)
			}
			if _, err = d.Feedback(ctx, id, id%3 == 0, now); err != nil {
				t.Fatal(err)
			}
			// Match the forecast-producing history, not merely the final label
			// count: mixture updates depend on the original pre-outcome experts.
			if err = d.worker.WaitProcessed(ctx, id); err != nil {
				t.Fatal(err)
			}
		}
		if err = d.worker.WaitProcessed(ctx, 64); err != nil {
			t.Fatal(err)
		}
		requests := make([]AdmissionRequest, 200)
		ids := make([]uint64, 200)
		discards := make([]DiscardRequest, 200)
		for i := range requests {
			ids[i] = uint64(i + 65)
			requests[i] = AdmissionRequest{ID: ids[i], Features: uint16(i), Baseline: .6, At: now}
			discards[i] = DiscardRequest{ids[i], now}
		}
		got, err := d.AdmitBatchWithSnapshotReads(ctx, requests)
		if err != nil {
			t.Fatal(err)
		}
		if !got[0].Record.Ready {
			t.Fatal("fixture is not warm")
		}
		if !prepared {
			want = got
		} else if !reflect.DeepEqual(got, want) {
			t.Fatal("prepared changed warm originals")
		}
		if err = d.Close(); err != nil {
			t.Fatal(err)
		}
		d, err = open(ctx, path, "tenant", "stream", 1, 42)
		if err != nil {
			t.Fatal(err)
		}
		retried, err := d.AdmitBatchWithSnapshotReads(ctx, requests)
		if err != nil {
			t.Fatal(err)
		}
		stored, err := d.Admissions(ctx, ids)
		if err != nil {
			t.Fatal(err)
		}
		for i, r := range retried {
			if !r.Retry || !reflect.DeepEqual(r.Record, got[i].Record) || !reflect.DeepEqual(stored[i], got[i].Record) {
				t.Fatal("reopen original changed", i)
			}
		}
		if _, err = d.DiscardBatchWithSnapshotReads(ctx, discards); err != nil {
			t.Fatal(err)
		}
		if err = d.Close(); err != nil {
			t.Fatal(err)
		}
		d, err = open(ctx, path, "tenant", "stream", 1, 42)
		if err != nil {
			t.Fatal(err)
		}
		retries, err := d.DiscardBatchWithSnapshotReads(ctx, discards)
		if err != nil {
			t.Fatal(err)
		}
		for _, retry := range retries {
			if !retry {
				t.Fatal("terminal not replayed")
			}
		}
		if err = d.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDurableTrainingScheduleChangesOriginals(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	for _, queued := range []bool{false, true} {
		d, err := OpenDurable(ctx, t.TempDir()+"/schedule.sqlite", "tenant", "stream", 1, 42)
		if err != nil {
			t.Fatal(err)
		}
		func() {
			// Deliberately hold fitting until all forecasts are recorded. This
			// controls scheduling without changing input labels or SQL methods.
			if queued {
				d.worker.adapter.mu.Lock()
				defer d.worker.adapter.mu.Unlock()
			}
			for id := uint64(1); id <= 64; id++ {
				if _, _, err = d.Admit(ctx, id, uint16(id), .6, now); err != nil {
					t.Fatal(err)
				}
				if _, err = d.Feedback(ctx, id, id%3 == 0, now); err != nil {
					t.Fatal(err)
				}
				if !queued {
					if err = d.worker.WaitProcessed(ctx, id); err != nil {
						t.Fatal(err)
					}
				}
			}
		}()
		if err = d.worker.WaitProcessed(ctx, 64); err != nil {
			t.Fatal(err)
		}
		r, err := d.Admission(ctx, 64)
		if err != nil {
			t.Fatal(err)
		}
		if r.Ready == queued {
			t.Fatal("schedule did not distinguish original experts", queued, r.Ready)
		}
		if err = d.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
