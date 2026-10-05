package researchmemory

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestDurableDiscardDoesNotLearn(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "discard.sqlite")
	now := durableFixtureTime
	d, e := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
	if e != nil {
		t.Fatal(e)
	}
	for id := uint64(1); id <= 300; id++ {
		if _, _, e = d.Admit(ctx, id, uint16(id%512), .6, now); e != nil {
			t.Fatal(e)
		}
		if dup, e := d.Discard(ctx, id, now); e != nil || dup {
			t.Fatal(dup, e)
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
	if n, f, p, q := d.worker.Counts(); n != 0 || f != 0 || p != 0 || q != 0 {
		t.Fatal("discard became evidence", n, f, p, q)
	}
	if dup, e := d.Discard(ctx, 1, now); e != nil || !dup {
		t.Fatal(dup, e)
	}
	if _, e = d.Feedback(ctx, 1, false, now); e == nil {
		t.Fatal("discard converted to label")
	}
	if _, _, e = d.Admit(ctx, 301, 3, .6, now); e != nil {
		t.Fatal(e)
	}
	if _, e = d.Discard(ctx, 301, now.Add(-time.Second)); e == nil {
		t.Fatal("early discard")
	}
	if _, e = d.Feedback(ctx, 301, true, now); e != nil {
		t.Fatal(e)
	}
	if _, e = d.Discard(ctx, 301, now); e == nil {
		t.Fatal("label converted to discard")
	}
	wait, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	if e = d.worker.WaitProcessed(wait, 1); e != nil {
		t.Fatal(e)
	}
}
