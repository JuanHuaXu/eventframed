package service

import (
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/rankdelta"
	"github.com/JuanHuaXu/eventframed/internal/ranking"
)

func TestElasticPromotionUsesRankBoundaryCertainty(t *testing.T) {
	uncertain := applyTestDelta(t, .51, .5, .1, 1, 1, .25)
	certain := applyTestDelta(t, .9, .5, .1, 1, 1, .25)
	if uncertain.RankDelta <= certain.RankDelta {
		t.Fatalf("uncertain-boundary promotion %v did not exceed certain-boundary promotion %v", uncertain.RankDelta, certain.RankDelta)
	}
	if uncertain.RankDeltaBasis != "rank-boundary+correction-reliability" || uncertain.RankDeltaAnswerCertainty <= 0 || uncertain.RankDeltaCorrectionReliability != 1 {
		t.Fatalf("promotion modulation = %+v", uncertain)
	}
}

func TestCorrectionReliabilityGatesPlasticity(t *testing.T) {
	reliable := applyTestDelta(t, .51, .5, -.1, 1, 1, .25)
	weak := applyTestDelta(t, .51, .5, -.1, .2, 1, .25)
	if math.Abs(reliable.RankDelta) <= math.Abs(weak.RankDelta) {
		t.Fatalf("reliable correction %v did not exceed weak correction %v", reliable.RankDelta, weak.RankDelta)
	}
	uncertified := applyTestDelta(t, .51, .5, -.1, 0, 1, .25)
	if uncertified.RankDelta != 0 {
		t.Fatalf("uncertified correction was applied: %+v", uncertified)
	}
}

func TestElasticDeltaPreservesHardCapAndDisabledControl(t *testing.T) {
	capped := applyTestDelta(t, .51, .5, .2, 1, 1, .25)
	if capped.RankDelta != .25 {
		t.Fatalf("elastic delta = %v, want hard cap .25", capped.RankDelta)
	}
	disabled := applyTestDelta(t, .51, .5, .1, 1, 0, .25)
	if disabled.RankDelta != .1 || disabled.RankDeltaScale != 1 {
		t.Fatalf("disabled elastic delta = %+v", disabled)
	}
	uncertifiedDisabled := applyTestDelta(t, .51, .5, .1, 0, 0, .25)
	if uncertifiedDisabled.RankDelta != 0 {
		t.Fatalf("disabled elasticity bypassed reliability gate: %+v", uncertifiedDisabled)
	}
}

func TestRankBoundaryCertaintyIgnoresForecastProbability(t *testing.T) {
	candidates := []model.Candidate{
		{RetrievalScore: .8, Forecast: model.ForecastBundle{CorrectedLaw: bernoulliLaw(.99)}},
		{RetrievalScore: .4, Forecast: model.ForecastBundle{CorrectedLaw: bernoulliLaw(.01)}},
	}
	want := rankBoundaryCertainty(candidates, 1)
	candidates[0].Forecast.CorrectedLaw = bernoulliLaw(.01)
	candidates[1].Forecast.CorrectedLaw = bernoulliLaw(.99)
	if got := rankBoundaryCertainty(candidates, 1); got != want || got != .5 {
		t.Fatalf("rank certainty = %v, want %v", got, want)
	}
}

func applyTestDelta(t *testing.T, topScore, candidateScore, delta, reliability, enabledScale, maxDelta float64) model.Candidate {
	t.Helper()
	policy := ranking.ElasticDeltaPolicy{Enabled: enabledScale != 0, MinScale: .5, MaxScale: 2.5}
	service := &Service{config: Config{MaxRankDelta: maxDelta, ElasticRankDelta: policy}}
	queryDigest := "query"
	candidates := []model.Candidate{
		{Event: model.Event{ID: "top"}, RetrievalScore: topScore, Score: topScore, BayesianApplied: true, Forecast: model.ForecastBundle{CorrectedLaw: bernoulliLaw(.5)}},
		{Event: model.Event{ID: "candidate"}, RetrievalScore: candidateScore, Score: candidateScore, BayesianApplied: true, Forecast: model.ForecastBundle{CorrectedLaw: bernoulliLaw(.5)}},
	}
	deltas := map[string]rankdelta.Record{
		rankDeltaKey(queryDigest, "candidate"): {Delta: delta, Reliability: reliability},
	}
	service.applyRankDeltas(candidates, deltas, queryDigest, 1)
	for _, candidate := range candidates {
		if candidate.Event.ID == "candidate" {
			return candidate
		}
	}
	t.Fatal("candidate disappeared")
	return model.Candidate{}
}
