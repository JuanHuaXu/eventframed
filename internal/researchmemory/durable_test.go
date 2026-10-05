package researchmemory

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestDurableAcknowledgmentAndRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "consumer.sqlite")
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	d, e := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
	if e != nil {
		t.Fatal(e)
	}
	a := New(1, 42)
	for id := uint64(1); id <= 48; id++ {
		p, dup, e := d.Admit(ctx, id, uint16(id), .6, now)
		if e != nil || dup {
			t.Fatal(p, dup, e)
		}
		r, e := d.worker.Record(id)
		if e != nil {
			t.Fatal(e)
		}
		if e = a.RestorePrediction(r); e != nil {
			t.Fatal(e)
		}
		q, dup, e := d.Admit(ctx, id, uint16(id), .6, now)
		if e != nil || !dup || p != q {
			t.Fatal("retry changed forecast", p, q, dup, e)
		}
		if _, e = d.Feedback(ctx, id, true, now.Add(-time.Second)); e == nil {
			t.Fatal("early label accepted")
		}
		if dup, e = d.Feedback(ctx, id, id%3 != 0, now); e != nil || dup {
			t.Fatal(dup, e)
		}
		if e = a.Feedback(id, id%3 != 0, 1, now); e != nil {
			t.Fatal(e)
		}
		if dup, e = d.Feedback(ctx, id, id%3 != 0, now); e != nil || !dup {
			t.Fatal("feedback retry", dup, e)
		}
	}
	if e = d.Close(); e != nil {
		t.Fatal(e)
	}
	d, e = OpenDurable(ctx, path, "tenant", "stream", 1, 42)
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	if n, f, p, _ := d.worker.Counts(); n != 48 || f != 0 || p != 0 {
		t.Fatal(n, f, p)
	}
	for x := uint16(0); x < 512; x++ {
		want, e := a.Freeze().Score(x, .6, 1, now)
		got, f := d.worker.Snapshot().Score(x, .6, 1, now)
		if e != nil || f != nil || want != got {
			t.Fatal(x, want, got, e, f)
		}
	}
	if _, dup, e := d.Admit(ctx, 1, 1, .6, now); e != nil || !dup {
		t.Fatal("reopened retry", dup, e)
	}
	if _, e = d.Feedback(ctx, 1, false, now); e == nil {
		t.Fatal("conflicting outcome accepted")
	}
	if p, dup, e := d.Admit(ctx, 49, 49, .6, now); e != nil || dup || p.ID != 49 {
		t.Fatal(p, dup, e)
	}
}
