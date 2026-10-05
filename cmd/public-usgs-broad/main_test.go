//go:build researchpriority

package main

import (
	"context"
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchcalendar"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

func TestTaskLexicalRankerPreservesFrontier(t *testing.T) {
	question := "What magnitude did USGS report for the earthquake at 2026-09-26T12:00:00Z?"
	events := []model.Event{
		{ID: "a", What: model.Field{Value: "USGS earthquake at 2026-09-26T11:00:00Z; magnitude 4.5"}},
		{ID: "b", What: model.Field{Value: "USGS earthquake at 2026-09-26T12:00:00Z; magnitude 4.6"}},
		{ID: "c", What: model.Field{Value: "USGS earthquake at 2026-09-26T13:00:00Z; magnitude 4.7"}},
	}
	input := make([]retrieval.Candidate, len(events))
	for i, event := range events {
		input[i] = retrieval.Candidate{ID: event.ID, Text: event.FrameText(), Score: .2 + float64(i)*.1}
	}
	request := retrieval.RankRequest{UserID: "research", QueryText: frame.QueryText(question), K1: 3, K2: 3, Candidates: input}
	ctx := researchcalendar.Bind(context.Background(), "research", question)
	ranker := &traceRanker{mode: "task-lexical"}
	out, err := ranker.RankCandidates(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 3 || len(ranker.before) != 3 || len(ranker.after) != 3 || out[0].ID != "b" {
		t.Fatalf("unexpected ranker frontier: %+v", out)
	}
	for i, item := range out {
		if math.Abs(item.Score-(1-float64(i)/4)) > 1e-12 {
			t.Fatalf("rank %d score = %f", i, item.Score)
		}
	}
	if !reflect.DeepEqual(request.Candidates, input) {
		t.Fatal("ranker mutated the input frontier")
	}
	if _, err := (&traceRanker{mode: "task-lexical"}).RankCandidates(context.Background(), request); err == nil {
		t.Fatal("unbound task context accepted")
	}
}
