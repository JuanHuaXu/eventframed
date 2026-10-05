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
func TestCalendarPriorityWithDiversityAfterDeltas(t *testing.T) {
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
		out := r.Candidates
		good := fmt.Sprintf("event-%d", n-1)
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
		certainty := s.applyRankDeltas(candidates, deltas, "query", 1)
		originalScores := make(map[string]float64, n)
		sourceText := make(map[string]string, n)
		for _, c := range out {
			sourceText[c.ID] = c.Text
		}
		plannedRequest := r
		plannedRequest.Candidates = nil
		for _, c := range candidates {
			originalScores[c.Event.ID] = c.Score
			plannedRequest.Candidates = append(plannedRequest.Candidates, retrieval.Candidate{ID: c.Event.ID, Text: sourceText[c.Event.ID], Score: c.Score})
		}
		plan, err := researchcalendar.PlanPriority(researchcalendar.Bind(context.Background(), "research", q), plannedRequest)
		if err != nil {
			t.Fatal(err)
		}
		// Earlier modulation used the original unmodified numeric score boundary.
		if certainty != priorityDiversityFixtureCertainty(r.Candidates) {
			t.Fatal("certainty input changed")
		}
		for _, c := range candidates {
			if c.Score != originalScores[c.Event.ID] {
				t.Fatal("score changed")
			}
		}
		if plan.CalibrationStatus != "not_evaluated" {
			t.Fatal("invented calibration")
		}
		policy := packing.DefaultPolicy()
		policy.DiversityEnabled = true
		packed, err := researchcalendar.PackPriority(candidates, plan, nil, 1, n, 100, policy)
		if err != nil {
			t.Fatal(err)
		}
		if len(packed.Candidates) != 1 || packed.Candidates[0].Event.ID != good {
			t.Fatal("priority rescue failed", n, packed)
		}
		for _, c := range candidates {
			if c.Forecast.CorrectedLaw != bernoulliLaw(.5) {
				t.Fatal("law changed")
			}
		}
		t.Logf("n=%d: temporal winner %s preserved as %s after allowed deltas", n, good, packed.Candidates[0].Event.ID)
	}
}

func priorityDiversityFixtureCertainty(rows []retrieval.Candidate) float64 {
	c := make([]model.Candidate, len(rows))
	for i, r := range rows {
		c[i].RetrievalScore = r.Score
	}
	return rankBoundaryCertainty(c, 1)
}
