package main

import (
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/evaluation"
	"github.com/JuanHuaXu/eventframed/internal/productioneval"
)

func TestExtractCalibratesOnlySinkVisibleNominations(t *testing.T) {
	now := time.Now().UTC()
	dataset := productioneval.Artifact{Cases: []evaluation.Case{{
		ID: "case-a", RelevantEventIDs: []string{"relevant"},
		Variants: map[string]evaluation.VariantForecast{"full": {Candidates: []evaluation.CandidateForecast{
			{EventID: "relevant", Probability: .7, Nominated: true},
			{EventID: "omitted", Probability: 0, Nominated: false},
		}}},
		PredictedAt: now,
	}}}

	observations, err := extract(dataset, "full")
	if err != nil {
		t.Fatal(err)
	}
	if len(observations) != 1 || observations[0].Probability != .7 || observations[0].Outcome != 1 {
		t.Fatalf("observations = %+v", observations)
	}
}
