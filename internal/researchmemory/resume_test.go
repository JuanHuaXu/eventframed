package researchmemory

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

func TestResumePendingAndContinue(t *testing.T) {
	ctx := context.Background()
	log, e := researchledger.Open(filepath.Join(t.TempDir(), "resume.sqlite"))
	if e != nil {
		t.Fatal(e)
	}
	defer log.Close()
	a := New(1, 42)
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	var pendingIDs []uint64
	for i := 0; i < 80; i++ {
		p, e := a.Predict(uint16(i), .6, 1, now)
		if e != nil {
			t.Fatal(e)
		}
		r, e := a.Record(p.ID)
		if e != nil {
			t.Fatal(e)
		}
		raw, _ := json.Marshal(r)
		key := researchledger.Key{Tenant: "tenant", Journal: "stream", Event: fmt.Sprint(p.ID), Contract: RecordContract}
		if _, _, e = log.Append(ctx, key, "admit", raw); e != nil {
			t.Fatal(e)
		}
		if i < 64 {
			f := RecordedFeedback{p.ID, i%3 != 0, now}
			raw, _ = json.Marshal(f)
			if _, _, e = log.Append(ctx, key, "feedback", raw); e != nil {
				t.Fatal(e)
			}
			if e = a.Feedback(p.ID, f.Useful, 1, now); e != nil {
				t.Fatal(e)
			}
		} else {
			pendingIDs = append(pendingIDs, p.ID)
		}
	}
	b, e := ResumeBackground(ctx, log, "tenant", "stream", 1, 42, 64)
	if e != nil {
		t.Fatal(e)
	}
	defer b.Close()
	if n, f, p, q := b.Counts(); n != 64 || f != 0 || p != 16 || q != 0 {
		t.Fatal(n, f, p, q)
	}
	for _, id := range pendingIDs {
		want, e := a.Record(id)
		if e != nil {
			t.Fatal(e)
		}
		got, e := b.Record(id)
		if e != nil || want != got {
			t.Fatal("lost original record", id, e)
		}
	}
	// Abandon one unresolved prediction without inventing negative evidence.
	last := pendingIDs[len(pendingIDs)-1]
	a.Discard(last)
	b.Discard(last)
	for _, id := range pendingIDs[:len(pendingIDs)-1] {
		if e = b.Feedback(id, true, 1, now); e != nil {
			t.Fatal(e)
		}
		if e = a.Feedback(id, true, 1, now); e != nil {
			t.Fatal(e)
		}
	}
	wait, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if e = b.WaitProcessed(wait, 79); e != nil {
		t.Fatal(e)
	}
	if e = b.Feedback(last, true, 1, now); e == nil {
		t.Fatal("discarded record reused")
	}
	b.adapter.mu.Lock()
	left := len(b.adapter.pending)
	b.adapter.mu.Unlock()
	if left != 0 {
		t.Fatal("pending ownership leaked", left)
	}
	for x := uint16(0); x < 512; x++ {
		want, e := a.Freeze().Score(x, .6, 1, now)
		got, f := b.Snapshot().Score(x, .6, 1, now)
		if e != nil || f != nil || want != got {
			t.Fatal("continuation differs", x, want, got, e, f)
		}
	}
	p, e := b.Predict(99, .6, 1, now)
	if e != nil || p.ID != 81 {
		t.Fatal("identity not continued", p, e)
	}
	if _, e = ResumeBackground(ctx, log, "tenant", "stream", 2, 42, 64); e == nil {
		t.Fatal("wrong epoch resumed")
	}
}
