package researchmemory

import (
	"testing"
	"time"
)

func awaitLabels(t *testing.T, b *Background, n uint64) {
	t.Helper()
	until := time.Now().Add(5 * time.Second)
	for time.Now().Before(until) {
		completed, failed, _, _ := b.Counts()
		if failed != 0 {
			t.Fatal("background fit failed", failed)
		}
		if completed == n {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("background did not finish")
}

func TestBackgroundSynchronousParity(t *testing.T) {
	b, err := NewBackground(1, 42, 64)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	a := New(1, 42)
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	cold := b.Snapshot()
	// Batches deliberately journal several predictions before any feedback.
	for batch := 0; batch < 8; batch++ {
		var ids [16]uint64
		for i := 0; i < 16; i++ {
			x := uint16(batch*16 + i)
			p, e := b.Predict(x, .6, 1, now)
			if e != nil {
				t.Fatal(e)
			}
			q, e := a.Predict(x, .6, 1, now)
			if e != nil || p != q {
				t.Fatal("journal mismatch", p, q, e)
			}
			ids[i] = p.ID
		}
		for i, id := range ids {
			if e := b.Feedback(id, i%2 == 0, 1, now); e != nil {
				t.Fatal(e)
			}
			if e := a.Feedback(id, i%2 == 0, 1, now); e != nil {
				t.Fatal(e)
			}
		}
		awaitLabels(t, b, uint64((batch+1)*16))
		for x := uint16(0); x < 512; x++ {
			p, e := b.Snapshot().Score(x, .6, 1, now)
			q, f := a.Freeze().Score(x, .6, 1, now)
			if e != nil || f != nil || p != q {
				t.Fatal("snapshot mismatch", x, p, q, e, f)
			}
		}
	}
	if p, e := cold.Score(1, .6, 1, now); e != nil || p != .6 {
		t.Fatal("old snapshot mutated")
	}
}

func TestBackgroundRejectsAndCloses(t *testing.T) {
	if b, e := NewBackground(1, 1, 0); e == nil {
		b.Close()
		t.Fatal("invalid capacity")
	}
	b, _ := NewBackground(1, 1, 1)
	now := time.Now()
	p, e := b.Predict(1, .6, 1, now)
	if e != nil {
		t.Fatal(e)
	}
	if b.Feedback(p.ID, true, 2, now) == nil || b.Feedback(p.ID, true, 1, now.Add(-time.Second)) == nil {
		t.Fatal("invalid feedback admitted")
	}
	if e = b.Feedback(p.ID, true, 1, now); e != nil {
		t.Fatal(e)
	}
	if b.Feedback(p.ID, true, 1, now) == nil {
		t.Fatal("duplicate admitted")
	}
	awaitLabels(t, b, 1)
	for i := 0; i < 256; i++ {
		if _, e = b.Predict(1, .6, 1, now); e != nil {
			t.Fatal(e)
		}
	}
	if _, e = b.Predict(1, .6, 1, now); e == nil {
		t.Fatal("unbounded pending")
	}
	b.Discard(p.ID + 1)
	if _, e = b.Predict(1, .6, 1, now); e != nil {
		t.Fatal("discard did not free pending slot", e)
	}
	b.Close()
	b.Close()
	if _, e = b.Predict(1, .6, 1, now); e == nil {
		t.Fatal("prediction after close")
	}
	_, _, pending, queued := b.Counts()
	if pending != 0 || queued != 0 {
		t.Fatal("close retained work")
	}
}

func TestBackgroundQueueRetry(t *testing.T) {
	b, _ := NewBackground(1, 42, 1)
	defer b.Close()
	now := time.Now()
	var ids [3]uint64
	for i := range ids {
		p, e := b.Predict(1, .6, 1, now)
		if e != nil {
			t.Fatal(e)
		}
		ids[i] = p.ID
	}
	// Hold only the worker's fit lock, not the prediction/admission lock.
	b.adapter.mu.Lock()
	locked := true
	defer func() {
		if locked {
			b.adapter.mu.Unlock()
		}
	}()
	if e := b.Feedback(ids[0], true, 1, now); e != nil {
		t.Fatal(e)
	}
	until := time.Now().Add(time.Second)
	for len(b.jobs) != 0 && time.Now().Before(until) {
		time.Sleep(time.Millisecond)
	}
	if len(b.jobs) != 0 {
		t.Fatal("worker did not dequeue")
	}
	if e := b.Feedback(ids[1], true, 1, now); e != nil {
		t.Fatal(e)
	}
	if b.Feedback(ids[2], true, 1, now) == nil {
		t.Fatal("full queue accepted")
	}
	_, _, pending, _ := b.Counts()
	if pending != 1 {
		t.Fatal("rejected label lost journal")
	}
	b.adapter.mu.Unlock()
	locked = false
	awaitLabels(t, b, 2)
	if e := b.Feedback(ids[2], true, 1, now); e != nil {
		t.Fatal("retry failed", e)
	}
	awaitLabels(t, b, 3)
}

var backgroundScoreSink float64

func TestBackgroundReadersDuringRefits(t *testing.T) {
	b, _ := NewBackground(1, 42, 256)
	defer b.Close()
	now := time.Now()
	var ids [128]uint64
	for i := range ids {
		p, e := b.Predict(uint16(i), .6, 1, now)
		if e != nil {
			t.Fatal(e)
		}
		ids[i] = p.ID
	}
	done := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		for {
			select {
			case <-done:
				result <- nil
				return
			default:
			}
			f := b.Snapshot()
			for x := uint16(0); x < 512; x++ {
				if _, e := f.Score(x, .6, 1, now); e != nil {
					result <- e
					return
				}
			}
		}
	}()
	defer func() {
		close(done)
		if e := <-result; e != nil {
			t.Error(e)
		}
	}()
	for i, id := range ids {
		if e := b.Feedback(id, i%2 == 0, 1, now); e != nil {
			t.Fatal(e)
		}
	}
	awaitLabels(t, b, 128)
}

func BenchmarkBackgroundFrozenScore(b *testing.B) {
	a := New(1, 42)
	now := time.Now()
	for i := 0; i < 64; i++ {
		p, e := a.Predict(uint16(i), .6, 1, now)
		if e != nil {
			b.Fatal(e)
		}
		if e = a.Feedback(p.ID, i%2 == 0, 1, now); e != nil {
			b.Fatal(e)
		}
	}
	worker, e := NewBackground(1, 42, 64)
	if e != nil {
		b.Fatal(e)
	}
	defer worker.Close()
	f := a.Freeze()
	worker.current.Store(&f)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, e := worker.Snapshot().Score(uint16(i%512), .6, 1, now)
		if e != nil {
			b.Fatal(e)
		}
		backgroundScoreSink = p
	}
}
