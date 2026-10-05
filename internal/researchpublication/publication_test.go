package researchpublication

import (
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestPendingAndCommittedBoundaries(t *testing.T) {
	base := model.Snapshot{RuntimeVersion: 1, EvidenceEpoch: 1}
	now := time.Now()
	for _, kind := range []Kind{General, Ingestion} {
		for _, offset := range []time.Duration{-time.Second, 0, time.Second} {
			p := New(base)
			ticket, e := p.Begin(kind, now.Add(offset))
			if e != nil {
				t.Fatal(e)
			}
			want := kind == Ingestion && offset > 0
			if p.Compatible(base, now) != want {
				t.Fatal("pending boundary", kind, offset)
			}
			next := base
			next.RuntimeVersion++
			if kind == Ingestion {
				next.EvidenceEpoch++
			} else {
				next.GraphVersion++
			}
			if e = p.Commit(ticket, next); e != nil {
				t.Fatal(e)
			}
			if p.Compatible(base, now) != want {
				t.Fatal("committed boundary", kind, offset)
			}
			if !p.Compatible(next, now) {
				t.Fatal("current snapshot rejected")
			}
			if p.Commit(ticket, next) == nil {
				t.Fatal("duplicate commit")
			}
		}
	}
}

func TestAbortAndMismatchQuarantine(t *testing.T) {
	base := model.Snapshot{RuntimeVersion: 1}
	now := time.Now()
	for _, uncertain := range []bool{false, true} {
		p := New(base)
		ticket, _ := p.Begin(General, time.Time{})
		if p.Abort(ticket+1, uncertain) == nil {
			t.Fatal("wrong owner aborted")
		}
		if e := p.Abort(ticket, uncertain); e != nil {
			t.Fatal(e)
		}
		if p.Compatible(base, now) == uncertain {
			t.Fatal("abort confidence boundary")
		}
	}
	p := New(base)
	ticket, _ := p.Begin(Ingestion, now.Add(time.Hour))
	bad := base
	bad.RuntimeVersion++
	bad.EvidenceEpoch++
	bad.GraphVersion++
	if p.Commit(ticket, bad) == nil || p.Compatible(base, now) || p.Compatible(bad, now) {
		t.Fatal("unexpected backend mutation not quarantined")
	}
	if _, e := p.Begin(General, now); e == nil {
		t.Fatal("quarantine reset itself")
	}
}

func TestPublicationConcurrentReadersAndHistory(t *testing.T) {
	base := model.Snapshot{RuntimeVersion: 1, EvidenceEpoch: 1}
	p := New(base)
	now := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				if !p.Compatible(base, now) {
					t.Error("compatible future stream rejected")
					return
				}
			}
		}()
	}
	current := base
	for i := 0; i < 100; i++ {
		ticket, e := p.Begin(Ingestion, now.Add(time.Hour))
		if e != nil {
			t.Fatal(e)
		}
		current.RuntimeVersion++
		current.EvidenceEpoch++
		if e = p.Commit(ticket, current); e != nil {
			t.Fatal(e)
		}
	}
	wg.Wait()
	for i := 100; i < 4097; i++ {
		ticket, _ := p.Begin(Ingestion, now.Add(time.Hour))
		current.RuntimeVersion++
		current.EvidenceEpoch++
		if e := p.Commit(ticket, current); e != nil {
			t.Fatal(e)
		}
	}
	if p.Compatible(base, now) {
		t.Fatal("evicted history accepted")
	}
}

func TestOverflowAndZeroTime(t *testing.T) {
	base := model.Snapshot{RuntimeVersion: ^uint64(0), EvidenceEpoch: ^uint64(0)}
	p := New(base)
	now := time.Now()
	if p.Compatible(base, time.Time{}) {
		t.Fatal("zero cutoff")
	}
	if _, e := p.Begin(Ingestion, time.Time{}); e == nil {
		t.Fatal("missing ingestion time")
	}
	ticket, _ := p.Begin(Ingestion, now.Add(time.Hour))
	if p.Commit(ticket, model.Snapshot{}) == nil {
		t.Fatal("wrapped version committed")
	}
}

func TestRestoredMotionIsValidatedAndDetached(t *testing.T) {
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	current := model.Snapshot{RuntimeVersion: 2, EvidenceEpoch: 2}
	old := model.Snapshot{RuntimeVersion: 1, EvidenceEpoch: 1}
	motion := map[uint64]time.Time{2: now.Add(time.Hour)}
	p, err := NewWithMotion(current, motion)
	if err != nil || !p.Compatible(old, now) {
		t.Fatalf("valid motion rejected: %v", err)
	}
	delete(motion, 2)
	if !p.Compatible(old, now) {
		t.Fatal("caller mutation changed published history")
	}
	for _, bad := range []map[uint64]time.Time{{0: now}, {3: now}, {2: time.Time{}}} {
		if _, err := NewWithMotion(current, bad); err == nil {
			t.Fatalf("invalid restored motion admitted: %+v", bad)
		}
	}
}
