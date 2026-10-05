package service

import (
	"context"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func bridgeFixture(t *testing.T) (*Service, *ResearchFeedbackBridge, ResearchFrontierObservation) {
	t.Helper()
	s, now := shadowService(t, ResearchShadowPolicy{})
	tap, _ := NewResearchFrontierTap("tenant-a", 1)
	t.Cleanup(tap.Close)
	s.config.ResearchFrontier = tap
	b, e := NewResearchFeedbackBridge(s, "tenant-a", 42)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(b.Close)
	shadowRecall(t, s, now)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	in, e := tap.Take(ctx)
	if e != nil {
		t.Fatal(e)
	}
	return s, b, in
}

func TestResearchFeedbackBindingsAndMissing(t *testing.T) {
	s, b, in := bridgeFixture(t)
	ctx := context.Background()
	p, e := b.Admit(ctx, in)
	if e != nil || len(p) != 3 {
		t.Fatal(e)
	}
	if _, e = b.Admit(ctx, in); e == nil {
		t.Fatal("duplicate observation admitted")
	}
	// A second successful Recall can yield the same journal: still one evidence set.
	shadowRecall(t, s, in.AsOf)
	again, e := s.config.ResearchFrontier.Take(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if again.JournalID != in.JournalID {
		t.Fatal("fixture no longer exercises repeated journal")
	}
	if _, e = b.Admit(ctx, again); e == nil {
		t.Fatal("repeated request duplicated evidence")
	}
	first := in.Candidates[0].EventID
	if b.Feedback(ctx, "wrong", first, true, in.AsOf) == nil || b.Feedback(ctx, in.JournalID, "wrong", true, in.AsOf) == nil {
		t.Fatal("unbound evidence accepted")
	}
	if b.Feedback(ctx, in.JournalID, first, true, in.AsOf.Add(-time.Second)) == nil {
		t.Fatal("early feedback accepted")
	}
	if e = b.Feedback(ctx, in.JournalID, first, true, in.AsOf.Add(time.Second)); e != nil {
		t.Fatal(e)
	}
	if b.Feedback(ctx, in.JournalID, first, true, in.AsOf.Add(time.Second)) == nil {
		t.Fatal("duplicate feedback accepted")
	}
	b.Discard(in.JournalID, in.Candidates[1].EventID)
	if b.Feedback(ctx, in.JournalID, in.Candidates[1].EventID, false, in.AsOf.Add(time.Second)) == nil {
		t.Fatal("discard learned negative")
	}
	if e = b.Feedback(ctx, in.JournalID, in.Candidates[2].EventID, false, in.AsOf.Add(2*time.Second)); e != nil {
		t.Fatal(e)
	}
	until := time.Now().Add(time.Second)
	for {
		n, failed, pending, _ := b.worker.Counts()
		if failed != 0 {
			t.Fatal("worker failed")
		}
		if n == 2 && pending == 0 {
			break
		}
		if time.Now().After(until) {
			t.Fatal("worker timeout")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestResearchFeedbackAtomicAdmissionAndStale(t *testing.T) {
	s, b, in := bridgeFixture(t)
	ctx := context.Background()
	bad := in
	bad.Candidates = append([]ResearchFrontierCandidate(nil), in.Candidates...)
	bad.Candidates[1].Features = 512
	if _, e := b.Admit(ctx, bad); e == nil {
		t.Fatal("invalid features accepted")
	}
	_, _, pending, _ := b.worker.Counts()
	if pending != 0 || len(b.seen) != 0 {
		t.Fatal("partial admission remained")
	}
	bad.Candidates[1] = bad.Candidates[0]
	if _, e := b.Admit(ctx, bad); e == nil {
		t.Fatal("duplicate member accepted")
	}
	bad = in
	bad.Candidates = append([]ResearchFrontierCandidate(nil), in.Candidates...)
	bad.Candidates[0].Baseline = .123456789
	if _, e := b.Admit(ctx, bad); e == nil {
		t.Fatal("modified baseline accepted")
	}
	if _, e := b.Admit(ctx, in); e != nil {
		t.Fatal("valid retry failed", e)
	}
	ev := testutil.Event("new-dependency", "public changed fixture", in.AsOf)
	if _, e := s.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: ev.ID, Event: ev}); e != nil {
		t.Fatal(e)
	}
	if b.Feedback(ctx, in.JournalID, in.Candidates[0].EventID, true, in.AsOf) == nil {
		t.Fatal("stale dependencies accepted")
	}
	if _, e := b.Score(ctx, 1, .6, in.AsOf); e == nil {
		t.Fatal("stale score exposed")
	}
	if _, e := b.Admit(ctx, in); e == nil {
		t.Fatal("stale admission")
	}
}

func TestResearchFeedbackTenantAndCapacity(t *testing.T) {
	s, b, in := bridgeFixture(t)
	ctx := context.Background()
	other, e := NewResearchFeedbackBridge(s, "other-tenant", 42)
	if e != nil {
		t.Fatal(e)
	}
	defer other.Close()
	if _, e = other.Admit(ctx, in); e == nil {
		t.Fatal("cross tenant admitted")
	}
	for i := 0; i < 256; i++ {
		b.seen[string(rune(i+1))] = true
	}
	if _, e = b.Admit(ctx, in); e == nil {
		t.Fatal("replay tombstone bound ignored")
	}
	b.Close()
	if _, e = b.Score(ctx, 1, .6, in.AsOf); e == nil {
		t.Fatal("closed score")
	}
}

func TestResearchFeedbackBridgeRefits(t *testing.T) {
	s, b, in := bridgeFixture(t)
	ctx := context.Background()
	control := researchmemory.New(1, 42)
	start := in.AsOf
	for round := 0; round < 16; round++ {
		if round > 0 {
			shadowRecall(t, s, start.Add(time.Duration(round)*time.Second))
			var e error
			in, e = s.config.ResearchFrontier.Take(ctx)
			if e != nil {
				t.Fatal(e)
			}
		}
		predictions, e := b.Admit(ctx, in)
		if e != nil {
			t.Fatal(e)
		}
		for i, c := range in.Candidates {
			p, e := control.Predict(c.Features, c.Baseline, 1, in.AsOf)
			if e != nil || p != predictions[i] {
				t.Fatal("pre-feedback mismatch", e)
			}
		}
		for i, c := range in.Candidates {
			// Declared fixture outcomes; this does not estimate real usefulness.
			useful := c.EventID != "shadow-b"
			if e := b.Feedback(ctx, in.JournalID, c.EventID, useful, in.AsOf); e != nil {
				t.Fatal(e)
			}
			if e := control.Feedback(predictions[i].ID, useful, 1, in.AsOf); e != nil {
				t.Fatal(e)
			}
		}
		until := time.Now().Add(time.Second)
		for {
			n, f, _, _ := b.worker.Counts()
			if f != 0 {
				t.Fatal("fit failed")
			}
			if n == uint64((round+1)*3) {
				break
			}
			if time.Now().After(until) {
				t.Fatal("fit timeout")
			}
			time.Sleep(time.Millisecond)
		}
	}
	for x := uint16(0); x < 512; x++ {
		p, e := b.Score(ctx, x, .6, start.Add(time.Hour))
		q, f := control.Freeze().Score(x, .6, 1, start.Add(time.Hour))
		if e != nil || f != nil || p != q {
			t.Fatal("refit mismatch", x, p, q, e, f)
		}
	}
}
