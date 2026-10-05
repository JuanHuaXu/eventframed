package researchcalendar

import (
	"context"
	"fmt"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

func BenchmarkCalendarFrontier(b *testing.B) {
	for _, n := range []int{50, 200} {
		q := "Which retained mission event occurred before 2010?"
		ctx := Bind(context.Background(), "research", q)
		r := retrieval.RankRequest{UserID: "research", QueryText: frame.QueryText(Focus(q)), K1: n, K2: n}
		for i := 0; i < n; i++ {
			text := "Rosetta launched on 2004-03-02."
			if i%2 == 0 {
				text = "Rosetta arrived at its comet on 2014-08-06."
			}
			r.Candidates = append(r.Candidates, retrieval.Candidate{ID: fmt.Sprintf("fixture-%d", i), Text: (model.Event{What: model.Field{Value: text}}).FrameText(), Score: float64(n-i) / float64(n+1)})
		}
		for _, arm := range []string{"passthrough", "calendar"} {
			b.Run(fmt.Sprintf("%s/%d", arm, n), func(b *testing.B) {
				var ranker retrieval.CandidateRanker = retrieval.PassthroughRanker{}
				if arm == "calendar" {
					ranker = Ranker{}
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					out, err := ranker.RankCandidates(ctx, r)
					if err != nil || len(out) != n {
						b.Fatal(len(out), err)
					}
				}
			})
		}
	}
}
