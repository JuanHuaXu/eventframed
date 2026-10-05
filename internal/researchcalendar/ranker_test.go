package researchcalendar

import (
	"context"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

func TestCalendarCycle(t *testing.T) {
	valid, invalid := 0, 0
	for y := 1600; y < 2000; y++ {
		for m := time.January; m <= time.December; m++ {
			for d := 1; d <= 31; d++ {
				dt := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
				if dt.Month() != m {
					invalid++
					continue
				}
				valid++
				for _, s := range []string{dt.Format("2006-01-02"), dt.Format("2 January 2006")} {
					got, ok := Date(s)
					if !ok || !got.Equal(dt) {
						t.Fatalf("date %s -> %v,%v", s, got, ok)
					}
				}
			}
		}
	}
	if valid != 146097 || invalid != 2703 {
		t.Fatal(valid, invalid)
	}
	for _, s := range []string{"1900-02-29", "31 April 2016", "2004-03-020", "2004-03-02suffix", "2004-03-02+01:00", "2004-03-02T07:17:00Z", "2004-03-02 and 2014-08-06", "2 March 2004 and 2014-08-06"} {
		if _, ok := Date(s); ok {
			t.Fatal("accepted", s)
		}
	}
}

func fixture(q string) (context.Context, retrieval.RankRequest) {
	text := func(what, when string) string {
		return (model.Event{What: model.Field{Value: what}, When: model.Field{Value: when}}).FrameText()
	}
	return Bind(context.Background(), "research", q), retrieval.RankRequest{UserID: "research", QueryText: frame.QueryText(Focus(q)), K1: 2, K2: 2, Candidates: []retrieval.Candidate{
		{ID: "late", Text: text("Arrived 2014-08-06.", "2000-01-01"), Score: .9},
		{ID: "early", Text: text("Launched 2004-03-02.", "2026-09-01"), Score: .8},
	}}
}

func TestOriginalQueryAndWhatBinding(t *testing.T) {
	ctx, req := fixture("Which event occurred before 2010?")
	before := append([]retrieval.Candidate(nil), req.Candidates...)
	out, err := (Ranker{}).RankCandidates(ctx, req)
	if err != nil || out[0].ID != "early" || out[0].Score <= out[1].Score {
		t.Fatal(out, err)
	}
	if !reflect.DeepEqual(req.Candidates, before) {
		t.Fatal("input mutated")
	}
	for _, q := range []string{"Which event, rather than 6 August 2014?", "Which event, rather than 2 March 2004?"} {
		ctx, req = fixture(q)
		out, err = (Ranker{}).RankCandidates(ctx, req)
		want := "early"
		if q == "Which event, rather than 2 March 2004?" {
			want = "late"
		}
		if err != nil || out[0].ID != want {
			t.Fatal(q, out, err)
		}
	}
}

func TestUnknownAndAllContradictedPreserved(t *testing.T) {
	for _, q := range []string{"Which event before 2000?", "Which event earlier than 2010?", "Which event not before 2010?"} {
		ctx, req := fixture(q)
		out, err := (Ranker{}).RankCandidates(ctx, req)
		if err != nil || !reflect.DeepEqual(out, req.Candidates) {
			t.Fatal(q, out, err)
		}
	}
	ctx, req := fixture("Which event before 2010?")
	req.Candidates[0].Text = (model.Event{What: model.Field{Value: "unknown"}, When: model.Field{Value: "2014-08-06"}}).FrameText()
	out, err := (Ranker{}).RankCandidates(ctx, req)
	if err != nil || out[0].ID != "late" {
		t.Fatal("used when as historical evidence", out, err)
	}
}

func TestFailClosed(t *testing.T) {
	for _, kind := range []string{"missing", "tenant", "query", "cancel", "duplicate", "nan", "cap"} {
		ctx, req := fixture("Which event before 2010?")
		switch kind {
		case "missing":
			ctx = context.Background()
		case "tenant":
			req.UserID = "other"
		case "query":
			req.QueryText = "different"
		case "cancel":
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			cancel()
		case "duplicate":
			req.Candidates[1].ID = req.Candidates[0].ID
		case "nan":
			req.Candidates[0].Score = math.NaN()
		case "cap":
			req.K1 = 201
		}
		if _, err := (Ranker{}).RankCandidates(ctx, req); err == nil {
			t.Fatal(kind)
		}
	}
}
