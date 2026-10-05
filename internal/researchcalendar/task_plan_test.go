package researchcalendar

import (
	"context"
	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"testing"
)

func TestTaskRolePlan(t *testing.T) {
	for _, tc := range []struct {
		q    string
		want int
		role string
	}{
		{"Which event occurred before 2010?", 0, "selection"},
		{"Which event occurred after 2010?", 1, "selection"},
		{"Was Rosetta's arrival before 2010?", 1, "assessment"},
		{"Which record disproves the claim that Rosetta launched after 2010?", 0, "assessment"},
		{"Which event isn't before 2010?", 1, "selection"},
		{"Which event did not occur after 2010?", 0, "selection"},
		{"Explain why the claim 'Rosetta arrived before 2010' is false.", 1, "assessment"},
		{"Which event occurred before 2000 or involved Rosetta?", 0, "unsupported"},
	} {
		req := retrieval.RankRequest{UserID: "research", QueryText: frame.QueryText(tc.q), K1: 2, K2: 2}
		for i, text := range []string{"Rosetta launched on 2004-03-02.", "Rosetta arrived at its comet on 2014-08-06."} {
			req.Candidates = append(req.Candidates, retrieval.Candidate{ID: []string{"launch", "arrival"}[i], Text: (model.Event{What: model.Field{Value: text}}).FrameText(), Score: .5})
		}
		p, err := PlanTaskRole(Bind(context.Background(), "research", tc.q), req)
		if err != nil || p.Order[0] != tc.want || p.Role != tc.role {
			t.Fatal(tc, p, err)
		}
		if p.Method != "research/task-role-v1" {
			t.Fatal("missing version")
		}
	}
}
