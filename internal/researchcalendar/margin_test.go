package researchcalendar

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

func TestMarginBoundGrid(t *testing.T) {
	for _, n := range []int{2, 3, 50, 200} {
		for _, good := range []int{1, n / 2, n - 1} {
			for _, bound := range []float64{0, .1, .25, .49} {
				q := "Which event before 2010?"
				ctx := Bind(context.Background(), "research", q)
				r := retrieval.RankRequest{UserID: "research", QueryText: frame.QueryText(q), K1: n, K2: n}
				for i := 0; i < n; i++ {
					text := "Rosetta arrived on 2014-08-06."
					if i < good {
						text = "Rosetta launched on 2004-03-02."
					}
					r.Candidates = append(r.Candidates, retrieval.Candidate{ID: fmt.Sprint(i), Text: (model.Event{What: model.Field{Value: text}}).FrameText(), Score: .5})
				}
				out, err := (MarginRanker{MaxDelta: bound}).RankCandidates(ctx, r)
				if err != nil {
					t.Fatal(err)
				}
				lowest := out[good-1].Score - bound
				highest := out[good].Score + bound
				if lowest <= highest {
					t.Fatal(n, good, bound, lowest, highest)
				}
			}
		}
	}
}

func TestMarginRejectsUnsupportedContract(t *testing.T) {
	ctx, r := fixture("Which event before 2010?")
	for _, v := range []float64{-.1, .5, math.NaN(), math.Inf(1)} {
		if _, err := (MarginRanker{MaxDelta: v}).RankCandidates(ctx, r); err == nil {
			t.Fatal(v)
		}
	}
	r.K2 = 1
	if _, err := (MarginRanker{MaxDelta: .25}).RankCandidates(ctx, r); err == nil {
		t.Fatal("partial frontier accepted")
	}
}
