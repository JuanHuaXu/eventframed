package researchpublicrankguard

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchpublicframe"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicpool"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

func fixture(t *testing.T) (*researchpublicpool.Registry, []retrieval.Candidate, time.Time) {
	t.Helper()
	clock := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	r, err := researchpublicpool.New(context.Background(), []researchpublicframe.Record{{ID: "a", Title: "One", Text: "Public source one"}, {ID: "b", Title: "Two", Text: "Public source two"}, {ID: "c", Title: "Three", Text: "Public source three"}},
		researchpublicframe.Config{TenantID: "t", SessionID: "s", ImportedAt: clock}, "public-scientific")
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entries()
	rows := []retrieval.Candidate{e[0].Candidate, e[1].Candidate, e[2].Candidate}
	return r, rows, clock
}
func TestCompleteAdviceOrWholeSearch(t *testing.T) {
	r, rows, clock := fixture(t)
	ctx := context.Background()
	for _, advice := range [][]retrieval.Candidate{nil, {}, rows[:1], rows[:2]} {
		out, err := CompleteOrSearch(ctx, r, rows, advice, clock, 200)
		if err != nil || out.NativeComplete || len(out.Candidates) != 3 {
			t.Fatalf("incomplete advice not rejected: %v", err)
		}
		for i, c := range out.Candidates {
			if c.ID != rows[i].ID || c.Score != rows[i].Score {
				t.Fatal("changed original fallback order")
			}
		}
		out.Candidates[0].Metadata[0] = 'x'
		if rows[0].Metadata[0] == 'x' {
			t.Fatal("fallback aliases input")
		}
	}
	reversed := []retrieval.Candidate{rows[2], rows[1], rows[0]}
	out, err := CompleteOrSearch(ctx, r, rows, reversed, clock, 200)
	if err != nil || !out.NativeComplete || out.Candidates[0].ID != rows[2].ID {
		t.Fatal("complete advice lost")
	}
}
func TestCorruptAdviceNeverFallsBack(t *testing.T) {
	r, rows, clock := fixture(t)
	for _, mutate := range []func([]retrieval.Candidate){
		func(v []retrieval.Candidate) { v[0].ID = "unknown" }, func(v []retrieval.Candidate) { v[0].Text += "changed" },
		func(v []retrieval.Candidate) { v[0].Metadata = []byte("{}") }, func(v []retrieval.Candidate) { v[0].Score = math.NaN() },
	} {
		v := []retrieval.Candidate{rows[0]}
		mutate(v)
		if _, err := CompleteOrSearch(context.Background(), r, rows, v, clock, 200); err == nil {
			t.Fatal("accepted corrupt advice")
		}
	}
	if _, err := CompleteOrSearch(context.Background(), r, rows, []retrieval.Candidate{rows[0], rows[0]}, clock, 200); err == nil {
		t.Fatal("accepted duplicate advice")
	}
	if _, err := CompleteOrSearch(context.Background(), r, rows[:2], rows[2:], clock, 200); err == nil {
		t.Fatal("accepted outside-frontier source")
	}
	if _, err := CompleteOrSearch(context.Background(), r, rows, nil, clock.Add(-time.Second), 200); err == nil {
		t.Fatal("accepted future source")
	}
}
func TestCompleteRankCapsAndCancellation(t *testing.T) {
	r, rows, clock := fixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := CompleteOrSearch(ctx, r, rows, nil, clock, 200); err == nil {
		t.Fatal("accepted canceled")
	}
	if _, err := CompleteOrSearch(context.Background(), r, rows, nil, clock, 2); err == nil {
		t.Fatal("accepted nomination over cap")
	}
	if _, err := CompleteOrSearch(context.Background(), r, rows, rows, clock, 201); err == nil {
		t.Fatal("accepted wrong cap")
	}
}
