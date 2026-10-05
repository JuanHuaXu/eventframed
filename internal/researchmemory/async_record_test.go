package researchmemory

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestReplayPreservesLaggingBackgroundForecasts(t *testing.T) {
	b, e := NewBackground(1, 42, 256)
	if e != nil {
		t.Fatal(e)
	}
	defer b.Close()
	now := time.Now()
	type op struct {
		Record *RecordedPrediction
		ID     uint64
		Useful bool
	}
	var log []op
	// Block fitting deliberately, while predictions and queued feedback continue.
	// This makes lag deterministic rather than relying on scheduler luck.
	func() {
		b.adapter.mu.Lock()
		defer b.adapter.mu.Unlock()
		for i := 0; i < 64; i++ {
			p, e := b.Predict(uint16(i), .6, 1, now)
			if e != nil {
				t.Fatal(e)
			}
			r, e := b.Record(p.ID)
			if e != nil {
				t.Fatal(e)
			}
			if r.Ready {
				t.Fatal("fixture did not retain cold snapshot")
			}
			// JSON roundtrip exercises the actual durable representation.
			raw, e := json.Marshal(r)
			if e != nil {
				t.Fatal(e)
			}
			var saved RecordedPrediction
			if e = json.Unmarshal(raw, &saved); e != nil {
				t.Fatal(e)
			}
			log = append(log, op{Record: &saved})
			useful := i%3 != 0
			if e = b.Feedback(p.ID, useful, 1, now); e != nil {
				t.Fatal(e)
			}
			log = append(log, op{ID: p.ID, Useful: useful})
			if _, e = b.Record(p.ID); e == nil {
				t.Fatal("consumed record exported")
			}
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if e = b.WaitProcessed(ctx, 64); e != nil {
		t.Fatal(e)
	}
	correct, naive := New(1, 42), New(1, 42)
	different := 0
	for _, v := range log {
		if v.Record != nil {
			if e = correct.RestorePrediction(*v.Record); e != nil {
				t.Fatal(e)
			}
			p, e := naive.Predict(v.Record.Prediction.Features, v.Record.Outer[0], 1, now)
			if e != nil {
				t.Fatal(e)
			}
			rec, e := naive.Record(p.ID)
			if e != nil {
				t.Fatal(e)
			}
			if rec.Ready != v.Record.Ready || rec.Outer != v.Record.Outer || rec.Inner != v.Record.Inner {
				different++
			}
		} else {
			if e = correct.Feedback(v.ID, v.Useful, 1, now); e != nil {
				t.Fatal(e)
			}
			if e = naive.Feedback(v.ID, v.Useful, 1, now); e != nil {
				t.Fatal(e)
			}
		}
	}
	if different == 0 {
		t.Fatal("negative control did not exercise stale experts")
	}
	forecastDifferences := 0
	for x := uint16(0); x < 512; x++ {
		live, e := b.Snapshot().Score(x, .6, 1, now)
		got, f := correct.Freeze().Score(x, .6, 1, now)
		bad, g := naive.Freeze().Score(x, .6, 1, now)
		if e != nil || f != nil || g != nil || live != got {
			t.Fatal("original-record replay differs", x, live, got, e, f, g)
		}
		if live != bad {
			forecastDifferences++
		}
	}
	if forecastDifferences == 0 {
		t.Fatal("recompute control failed to distinguish final state")
	}
	t.Logf("recomputed experts differ at %d admissions; final forecasts differ in %d/512 states", different, forecastDifferences)
}
