package researchregimelogsummary

import (
	"math"
	"testing"

	old "github.com/JuanHuaXu/eventframed/internal/researchregimelog"
)

func summaryNear(t *testing.T, a, b float64) {
	t.Helper()
	near(t, a, b)
	report.SummaryParityChecks++
	report.MaxSummaryParity = math.Max(report.MaxSummaryParity, math.Abs(a-b))
}
func TestPublicSummaryParityAgainstFrozenLogParent(t *testing.T) {
	for _, members := range []int{2, 150, 200} {
		base := make([]float64, members)
		for i := range base {
			base[i] = .25 + .675*float64(i)/float64(members-1)
		}
		for _, cfg := range []Config{{0, 1, 9}, {.0625, .25, 36}, {.5, 0, 18}, {1, 1, 9}} {
			a, e := New(base, cfg)
			if e != nil {
				t.Fatal(e)
			}
			b, e := old.New(base, old.Config{Reset: cfg.Reset, Hazard: cfg.Hazard, Cap: cfg.Cap})
			if e != nil {
				t.Fatal(e)
			}
			for i := 0; i < 64; i++ {
				x, e := a.Issue(i%members, int64(i))
				if e != nil {
					t.Fatal(e)
				}
				y, e := b.Issue(i%members, int64(i))
				if e != nil {
					t.Fatal(e)
				}
				if x.LawID != y.LawID || x.Version != y.Version || x.SupportEpoch != y.SupportEpoch {
					t.Fatal("issue identity changed")
				}
				summaryNear(t, x.Clean, y.Clean)
				summaryNear(t, x.First, y.First)
				if i%4 == 3 {
					target := i - 2
					if e = a.Reveal(a.token, target, 1, target%2, int64(i)); e != nil {
						t.Fatal(e)
					}
					if e = b.Reveal(b.Snapshot().Token, target, 1, target%2, int64(i)); e != nil {
						t.Fatal(e)
					}
					q, e := a.Pending(a.token, target, 2)
					if e != nil {
						t.Fatal(e)
					}
					r, e := b.Pending(b.Snapshot().Token, target, 2)
					if e != nil {
						t.Fatal(e)
					}
					if q.LawID != r.LawID {
						t.Fatal("query identity changed")
					}
					for v := 0; v < 2; v++ {
						summaryNear(t, q.Branches[v].Probability, r.Branches[v].Probability)
						summaryNear(t, q.Branches[v].LogProbability, r.Branches[v].LogProbability)
						if q.Branches[v].Underflow != r.Branches[v].Underflow {
							t.Fatal("underflow metadata")
						}
						for j, p := range q.Branches[v].NextClean {
							summaryNear(t, p, r.Branches[v].NextClean[j])
						}
					}
					if e = a.Reveal(a.token, target, 2, 1-target%2, int64(i)); e != nil {
						t.Fatal(e)
					}
					if e = b.Reveal(b.Snapshot().Token, target, 2, 1-target%2, int64(i)); e != nil {
						t.Fatal(e)
					}
				}
			}
			s, u := a.Snapshot(), b.Snapshot()
			if s.LawID != u.LawID {
				t.Fatal("final identity changed")
			}
			summaryNear(t, s.LogEvidence, u.LogEvidence)
			for j, p := range s.NextClean {
				summaryNear(t, p, u.NextClean[j])
			}
		}
	}
}
