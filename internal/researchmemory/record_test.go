package researchmemory

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

func TestDurableOriginalForecastReplay(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "replay.sqlite")
	log, e := researchledger.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	live := New(1, 42)
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	type label struct {
		ID        uint64
		Useful    bool
		Available time.Time
	}
	// Predictions are batched before labels: later labels must not retroactively
	// change the earlier experts. More than256 lifetime identities are retained.
	for batch := 0; batch < 20; batch++ {
		var ids []uint64
		for i := 0; i < 16; i++ {
			p, e := live.Predict(uint16((batch*16+i)%512), .6, 1, now)
			if e != nil {
				t.Fatal(e)
			}
			r, e := live.Record(p.ID)
			if e != nil {
				t.Fatal(e)
			}
			b, e := json.Marshal(r)
			if e != nil {
				t.Fatal(e)
			}
			key := researchledger.Key{Tenant: "tenant", Journal: "journal", Event: fmt.Sprint(p.ID), Contract: RecordContract}
			if _, _, e = log.Append(ctx, key, "admit", b); e != nil {
				t.Fatal(e)
			}
			ids = append(ids, p.ID)
		}
		for i, id := range ids {
			l := label{id, (batch+i)%3 != 0, now}
			b, _ := json.Marshal(l)
			key := researchledger.Key{Tenant: "tenant", Journal: "journal", Event: fmt.Sprint(id), Contract: RecordContract}
			if _, _, e = log.Append(ctx, key, "feedback", b); e != nil {
				t.Fatal(e)
			}
			if e = live.Feedback(id, l.Useful, 1, now); e != nil {
				t.Fatal(e)
			}
		}
	}
	if e = log.Close(); e != nil {
		t.Fatal(e)
	}
	log, e = researchledger.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer log.Close()
	restored := New(1, 42)
	var after int64
	for {
		page, e := log.ReadAfter(ctx, after, 37)
		if e != nil {
			t.Fatal(e)
		}
		if len(page) == 0 {
			break
		}
		for _, entry := range page {
			if entry.Kind == "admit" {
				var r RecordedPrediction
				if e = json.Unmarshal(entry.Payload, &r); e != nil {
					t.Fatal(e)
				}
				e = restored.RestorePrediction(r)
			} else {
				var l label
				if e = json.Unmarshal(entry.Payload, &l); e != nil {
					t.Fatal(e)
				}
				e = restored.Feedback(l.ID, l.Useful, 1, l.Available)
			}
			if e != nil {
				t.Fatal(e)
			}
			after = entry.Sequence
		}
	}
	if n, p := restored.Counts(); n != 320 || p != 0 || after != 640 {
		t.Fatal(n, p, after)
	}
	for x := uint16(0); x < 512; x++ {
		a, e := live.Freeze().Score(x, .6, 1, now)
		b, f := restored.Freeze().Score(x, .6, 1, now)
		if e != nil || f != nil || a != b {
			t.Fatal("replay forecast differs", x, a, b, e, f)
		}
	}
}

func TestRecordRejectsCorruptionAndOrder(t *testing.T) {
	a := New(1, 42)
	p, e := a.Predict(1, .6, 1, time.Now())
	if e != nil {
		t.Fatal(e)
	}
	r, e := a.Record(p.ID)
	if e != nil {
		t.Fatal(e)
	}
	for _, mutate := range []func(*RecordedPrediction){func(r *RecordedPrediction) { r.Contract = "unknown" }, func(r *RecordedPrediction) { r.Seed++ }, func(r *RecordedPrediction) { r.Prediction.ID++ }, func(r *RecordedPrediction) { r.Prediction.Epoch++ }, func(r *RecordedPrediction) { r.Outer[2] = math.NaN() }, func(r *RecordedPrediction) { r.At = time.Time{} }, func(r *RecordedPrediction) { r.Prediction.Probability = .1 }} {
		bad := r
		mutate(&bad)
		if e = New(1, 42).RestorePrediction(bad); e == nil {
			t.Fatal("bad record accepted", bad)
		}
	}
	b := New(1, 42)
	if e = b.RestorePrediction(r); e != nil {
		t.Fatal(e)
	}
	if e = b.RestorePrediction(r); e == nil {
		t.Fatal("duplicate restored")
	}
}
