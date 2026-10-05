package researchcalendar

import (
	"context"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

// These are counterexample witnesses, not acceptance tests for hard exclusion.
// A passing test confirms the existing parser lacks the authority to discard
// evidence solely because its date fails a substring predicate.
func TestCalendarPredicateAuthorityCounterexamples(t *testing.T) {
	cases := []struct{ name, query, text string }{
		{"polar-question", "Was Rosetta's arrival before 2010?", "Rosetta arrived at its comet on 2014-08-06."},
		{"falsification", "Which record disproves the claim that Rosetta arrived before 2010?", "Rosetta arrived at its comet on 2014-08-06."},
		{"contraction", "Which events aren't before 2010?", "Rosetta arrived at its comet on 2014-08-06."},
		{"disjunction", "Which events occurred before 2000 or involved Rosetta?", "Rosetta launched on 2004-03-02."},
		{"quotation", "Explain why the claim 'Rosetta arrived before 2010' is false.", "Rosetta arrived at its comet on 2014-08-06."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			text := (model.Event{What: model.Field{Value: c.text}}).FrameText()
			req := retrieval.RankRequest{UserID: "research", QueryText: frame.QueryText(Focus(c.query)), K1: 1, K2: 1, Candidates: []retrieval.Candidate{{ID: "relevant-evidence", Text: text, Score: .8}}}
			plan, err := PlanPriority(Bind(context.Background(), "research", c.query), req)
			if err != nil {
				t.Fatal(err)
			}
			if plan.Decisions[0].State != "contradicted" || !plan.AllContradicted {
				t.Fatal("counterexample changed; reassess authority", plan)
			}
			t.Log("relevant evidence misclassified as contradicted; hard exclusion would be unsafe")
		})
	}
}

func TestCalendarPredicateAuthorityAdjacentControls(t *testing.T) {
	for _, q := range []string{"Which event occurred before 2010?", "Which event did not occur before 2010?"} {
		req := retrieval.RankRequest{UserID: "research", QueryText: frame.QueryText(q), K1: 1, K2: 1, Candidates: []retrieval.Candidate{{ID: "arrival", Text: (model.Event{What: model.Field{Value: "Rosetta arrived at its comet on 2014-08-06."}}).FrameText(), Score: .8}}}
		p, err := PlanPriority(Bind(context.Background(), "research", q), req)
		if err != nil {
			t.Fatal(err)
		}
		want := "contradicted"
		if q == "Which event did not occur before 2010?" {
			want = "unknown"
		}
		if p.Decisions[0].State != want {
			t.Fatal(q, p)
		}
	}
}
