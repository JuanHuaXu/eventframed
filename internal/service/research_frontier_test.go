package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

func TestResearchFrontierLearnerHandoff(t *testing.T) {
	base, now := shadowService(t, ResearchShadowPolicy{})
	s, _ := shadowService(t, ResearchShadowPolicy{})
	tap, e := NewResearchFrontierTap("tenant-a", 1)
	if e != nil {
		t.Fatal(e)
	}
	defer tap.Close()
	s.config.ResearchFrontier = tap
	worker, e := researchmemory.NewBackground(1, 42, 64)
	if e != nil {
		t.Fatal(e)
	}
	defer worker.Close()
	control := researchmemory.New(1, 42)
	var completed uint64
	// Fixture labels test plumbing, not accuracy: they are never read from rank.
	labels := map[string]bool{"shadow-a": true, "shadow-b": false, "shadow-c": true}
	for iteration := 0; iteration < 16; iteration++ {
		at := now.Add(time.Duration(iteration) * time.Second)
		want, got := shadowRecall(t, base, at), shadowRecall(t, s, at)
		if !reflect.DeepEqual(want, got) {
			t.Fatal("tap changed served packet")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		in, e := tap.Take(ctx)
		cancel()
		if e != nil {
			t.Fatal(e)
		}
		if len(in.Candidates) != 3 || got.Packed >= len(in.Candidates) || in.JournalID != got.BayesianShadow.JournalID || in.Snapshot != got.Snapshot || in.AsOf != at {
			t.Fatal("incorrect prepacking handoff")
		}
		if in.Snapshot != s.store.Snapshot(context.Background()) {
			t.Fatal("changed dependency snapshot")
		}
		var ids []uint64
		for _, c := range in.Candidates {
			p, e := worker.Predict(c.Features, c.Baseline, 1, in.AsOf)
			if e != nil {
				t.Fatal(e)
			}
			q, e := control.Predict(c.Features, c.Baseline, 1, in.AsOf)
			if e != nil || p != q {
				t.Fatal("prediction differs", e)
			}
			ids = append(ids, p.ID)
		}
		// Release externally supplied labels only after every candidate forecast.
		for i, c := range in.Candidates {
			useful, ok := labels[c.EventID]
			if !ok {
				t.Fatal("unbound fixture outcome")
			}
			if e := worker.Feedback(ids[i], useful, 1, at); e != nil {
				t.Fatal(e)
			}
			if e := control.Feedback(ids[i], useful, 1, at); e != nil {
				t.Fatal(e)
			}
			completed++
		}
		until := time.Now().Add(time.Second)
		for {
			n, failed, _, _ := worker.Counts()
			if failed != 0 {
				t.Fatal("worker failed")
			}
			if n == completed {
				break
			}
			if time.Now().After(until) {
				t.Fatal("worker timeout")
			}
			time.Sleep(time.Millisecond)
		}
	}
	if completed != 48 {
		t.Fatal("missing labels")
	}
	for x := uint16(0); x < 512; x++ {
		p, e := worker.Snapshot().Score(x, .6, 1, now.Add(time.Hour))
		q, f := control.Freeze().Score(x, .6, 1, now.Add(time.Hour))
		if e != nil || f != nil || p != q {
			t.Fatal("learned model mismatch")
		}
	}
}

type frontierJournalFailure struct {
	store.EventStore
	fail, attempts int
}

func (f *frontierJournalFailure) PutBayesianJournal(ctx context.Context, j model.BayesianJournalEntry) error {
	f.attempts++
	if f.attempts <= f.fail {
		return store.ErrStaleSnapshot
	}
	return f.EventStore.PutBayesianJournal(ctx, j)
}

func TestResearchFrontierCommitAndBackpressure(t *testing.T) {
	s, now := shadowService(t, ResearchShadowPolicy{})
	tap, _ := NewResearchFrontierTap("tenant-a", 1)
	defer tap.Close()
	s.config.ResearchFrontier = tap
	f := &frontierJournalFailure{EventStore: s.store, fail: 5}
	s.store = f
	r := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "public query", AsOf: now, RecallK: 3, PackK: 2, TokenBudget: 1000}
	if _, e := s.Recall(context.Background(), r); !errors.Is(e, store.ErrStaleSnapshot) {
		t.Fatal("expected failed commit", e)
	}
	if len(tap.queue) != 0 {
		t.Fatal("failed attempts escaped")
	}
	f.attempts = 0
	f.fail = 3
	if _, e := s.Recall(context.Background(), r); e != nil {
		t.Fatal(e)
	}
	if f.attempts != 4 || len(tap.queue) != 1 {
		t.Fatal("retries duplicated handoff")
	}
	shadowRecall(t, s, now)
	if tap.Dropped() != 1 {
		t.Fatal("backpressure not counted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := tap.Take(ctx); !errors.Is(e, context.Canceled) {
		t.Fatal("cancelled consumer accepted")
	}
	if len(tap.queue) != 1 {
		t.Fatal("cancelled consumer removed queued observation")
	}
	tap.Close()
	shadowRecall(t, s, now)
	if tap.Dropped() != 2 {
		t.Fatal("closed tap affected serving")
	}
}

func TestResearchFrontierIsolation(t *testing.T) {
	p, _ := NewResearchFrontierTap("tenant-a", 1)
	defer p.Close()
	s := &Service{config: Config{ResearchFrontier: p}}
	cs := []model.Candidate{{}}
	cs[0].Event.ID = "fixture"
	cs[0].Event.What.Value = "public query"
	r := model.RecallRequest{TenantID: "other", Query: "public query", AsOf: time.Now()}
	s.tapResearchFrontier(r, model.BayesianJournalEntry{}, cs)
	if len(p.queue) != 0 {
		t.Fatal("tenant crossed")
	}
	r.TenantID = "tenant-a"
	s.tapResearchFrontier(r, model.BayesianJournalEntry{}, cs)
	in, e := p.Take(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	in.Candidates[0].EventID = "mutated"
	if cs[0].Event.ID != "fixture" {
		t.Fatal("tap shared candidate memory")
	}
	s.tapResearchFrontier(r, model.BayesianJournalEntry{}, make([]model.Candidate, 201))
	if p.Dropped() != 1 || len(p.queue) != 0 {
		t.Fatal("oversize not rejected")
	}
}

func BenchmarkResearchFrontierTap(b *testing.B) {
	for _, mode := range []string{"disabled", "drained", "full"} {
		b.Run(mode, func(b *testing.B) {
			p, _ := NewResearchFrontierTap("tenant-a", 1)
			defer p.Close()
			s := &Service{}
			if mode != "disabled" {
				s.config.ResearchFrontier = p
			}
			r := model.RecallRequest{TenantID: "tenant-a", Query: "public query", AsOf: time.Now()}
			cs := make([]model.Candidate, 200)
			for i := range cs {
				cs[i].Event.What.Value = "public query reference"
				cs[i].Forecast.PreResidualLaw.Useful = .6
			}
			if mode == "full" {
				s.tapResearchFrontier(r, model.BayesianJournalEntry{}, cs)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s.tapResearchFrontier(r, model.BayesianJournalEntry{}, cs)
				if mode == "drained" {
					<-p.queue
				}
			}
		})
	}
}
