package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"github.com/JuanHuaXu/eventframed/internal/rankdelta"
	"github.com/JuanHuaXu/eventframed/internal/researchcalendar"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

// This injects admissible correction records; it does not claim they were learned
// in these fixtures. It exercises the actual delta and packing implementations.
func TestCalendarMarginSurvivesBoundedDeltas(t *testing.T) {
	for _, n := range []int{2, 50, 200} {
		q := "Which event before 2010?"
		r := retrieval.RankRequest{UserID: "research", QueryText: frame.QueryText(q), K1: n, K2: n}
		for i := 0; i < n; i++ {
			text := "Rosetta arrived at its comet on 2014-08-06."
			if i == n-1 {
				text = "Rosetta launched on 2004-03-02."
			}
			r.Candidates = append(r.Candidates, retrieval.Candidate{ID: fmt.Sprintf("event-%d", i), Text: (model.Event{What: model.Field{Value: text}}).FrameText(), Score: float64(n-i) / float64(n+1)})
		}
		out, err := (researchcalendar.MarginRanker{MaxDelta: .25}).RankCandidates(researchcalendar.Bind(context.Background(), "research", q), r)
		if err != nil {
			t.Fatal(err)
		}
		good := fmt.Sprintf("event-%d", n-1)
		if out[0].ID != good {
			t.Fatal("calendar stage failed")
		}
		var candidates []model.Candidate
		deltas := map[string]rankdelta.Record{}
		for _, c := range out {
			candidates = append(candidates, model.Candidate{Event: model.Event{ID: c.ID}, Score: c.Score, RetrievalScore: c.Score, EstimatedTokens: 1, Forecast: model.ForecastBundle{CorrectedLaw: bernoulliLaw(.5)}})
			d := .25
			if c.ID == good {
				d = -.25
			}
			deltas[rankDeltaKey("query", c.ID)] = rankdelta.Record{Delta: d, Reliability: 1}
		}
		s := Service{config: Config{MaxRankDelta: .25}}
		s.applyRankDeltas(candidates, deltas, "query", 1)
		packed := packing.Select(candidates, nil, 1, n, 100, packing.DefaultPolicy())
		if len(packed.Candidates) != 1 || packed.Candidates[0].Event.ID != good {
			t.Fatal("margin rescue failed", n, packed)
		}
		for _, c := range candidates {
			if c.Forecast.CorrectedLaw != bernoulliLaw(.5) {
				t.Fatal("law changed")
			}
		}
		t.Logf("n=%d: temporal winner %s preserved as %s after allowed deltas", n, good, packed.Candidates[0].Event.ID)
	}
}
