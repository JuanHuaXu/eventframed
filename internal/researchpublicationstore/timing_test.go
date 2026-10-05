package researchpublicationstore

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

func TestGateTimingCancellationErrorAndPanic(t *testing.T) {
	s, err := New(memorystore.New())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if s.timing != nil {
		t.Fatal("default enabled timing")
	}
	recorder, err := NewGateRecorder(20)
	if err != nil {
		t.Fatal(err)
	}
	s.timing = recorder
	bg := context.Background()
	snapshot := s.Snapshot(bg)
	if err = s.writer.Acquire(bg, 1); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(bg, 5*time.Millisecond)
	err = s.WithResearchAsOfSnapshotWait(ctx, snapshot, time.Now(), func() error { t.Error("expired callback ran"); return nil })
	cancel()
	s.writer.Release(1)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(bg, time.Second)
	defer cancel()
	want := errors.New("fixture callback")
	start := time.Now()
	err = s.WithResearchSnapshotWait(ctx, snapshot, func() error { time.Sleep(time.Millisecond); return want })
	elapsed := time.Since(start)
	if !errors.Is(err, want) {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("panic swallowed")
			}
		}()
		_ = s.WithResearchSnapshot(ctx, snapshot, func() error { panic("fixture") })
	}()
	if !s.writer.TryAcquire(1) {
		t.Fatal("panic retained gate")
	}
	s.writer.Release(1)
	if _, err = s.BindBayesianPolicy(bg, "timing-fixture"); err != nil {
		t.Fatal(err)
	}
	spans, dropped := recorder.Snapshot()
	if dropped != 0 || len(spans) != 4 {
		t.Fatal(spans, dropped)
	}
	if spans[0].Entered || spans[0].HeldNS != 0 || spans[0].WaitNS <= 0 {
		t.Fatal("bad canceled measurement", spans[0])
	}
	if !spans[1].Entered || spans[1].HeldNS <= 0 || spans[1].WaitNS+spans[1].HeldNS > elapsed.Nanoseconds() {
		t.Fatal("bad phase boundaries", spans[1], elapsed)
	}
	if !spans[2].Entered || spans[3].Kind != "general" || !spans[3].Entered {
		t.Fatal(spans)
	}
	spans[1].Kind = "changed"
	again, _ := recorder.Snapshot()
	if again[1].Kind == "changed" {
		t.Fatal("snapshot aliases recorder")
	}
}

func TestGateRecorderBoundAndCapabilities(t *testing.T) {
	if _, e := NewGateRecorder(0); e == nil {
		t.Fatal("accepted invalid limit")
	}
	r, _ := NewGateRecorder(3)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				r.record(GateSpan{Kind: "fixture"})
				r.Snapshot()
			}
		}()
	}
	wg.Wait()
	spans, dropped := r.Snapshot()
	if len(spans) != 3 || dropped != 77 {
		t.Fatal(len(spans), dropped)
	}
	for _, backend := range []store.EventStore{memorystore.New(), struct{ store.EventStore }{memorystore.New()}} {
		recorder, _ := NewGateRecorder(4)
		wrapped, e := WrapMeasured(backend, recorder)
		if e != nil {
			t.Fatal(e)
		}
		_, want := backend.(store.VectorEventStore)
		_, got := wrapped.(store.VectorEventStore)
		if got != want {
			t.Fatal("capability changed")
		}
		if e = wrapped.Close(); e != nil {
			t.Fatal(e)
		}
		spans, dropped := recorder.Snapshot()
		if len(spans) != 1 || spans[0].Kind != "close" || dropped != 0 {
			t.Fatal(spans, dropped)
		}
	}
	if _, e := WrapMeasured(memorystore.New(), nil); e == nil {
		t.Fatal("accepted nil recorder")
	}
}
