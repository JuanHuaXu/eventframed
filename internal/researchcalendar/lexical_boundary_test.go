package researchcalendar

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

func lexicalFixture(n int) []retrieval.Candidate {
	out := make([]retrieval.Candidate, n)
	for i := range out {
		text := "Curiosity launched on 2011-11-26."
		if i%2 == 1 {
			text = "Curiosity landed on Mars on 2012-08-06."
		}
		out[i] = retrieval.Candidate{ID: fmt.Sprintf("r-%d", i), Text: (model.Event{What: model.Field{Value: text}}).FrameText(), Score: .5}
	}
	return out
}

func TestLexicalBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func([]retrieval.Candidate) []retrieval.Candidate
	}{
		{"empty", func(c []retrieval.Candidate) []retrieval.Candidate { return nil }},
		{"over-cap", func(c []retrieval.Candidate) []retrieval.Candidate { return lexicalFixture(201) }},
		{"duplicate", func(c []retrieval.Candidate) []retrieval.Candidate { c[1].ID = c[0].ID; return c }},
		{"nan", func(c []retrieval.Candidate) []retrieval.Candidate { c[0].Score = math.NaN(); return c }},
		{"unframed", func(c []retrieval.Candidate) []retrieval.Candidate { c[0].Text = "what: landing"; return c }},
		{"duplicate-what", func(c []retrieval.Candidate) []retrieval.Candidate { c[0].Text += "\nwhat: other"; return c }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := LexicalOrder(context.Background(), "landing", tc.mutate(lexicalFixture(2)), nil); err == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := LexicalOrder(ctx, "landing", lexicalFixture(2), nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := LexicalOrder(context.Background(), strings.Repeat("q", 4097), lexicalFixture(2), nil); err == nil {
		t.Fatal("query cap")
	}
	if _, err := LexicalOrder(context.Background(), "landing", lexicalFixture(2), &PriorityPlan{}); err == nil {
		t.Fatal("plan identity")
	}
	got, err := LexicalOrder(context.Background(), "the and", lexicalFixture(2), nil)
	if err != nil || got.Scores[0] != 0 || got.Scores[1] != 0 || got.Order[0] != 0 {
		t.Fatal("empty terms tie contract", got, err)
	}
}

func BenchmarkWhatLexical(b *testing.B) {
	for _, n := range []int{50, 200} {
		b.Run(fmt.Sprintf("n%d", n), func(b *testing.B) {
			candidates := lexicalFixture(n)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := LexicalOrder(context.Background(), "Which record refutes the claim that Curiosity landed after 2020?", candidates, nil); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
