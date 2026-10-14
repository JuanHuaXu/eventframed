package agency

import (
	"github.com/JuanHuaXu/eventframed/internal/model"
	"math"
	"testing"
	"time"
)

// Enumerate finite reference laws, not just the optimizer's own arithmetic.
func TestTauUtilityBoundAgainstPerturbedLaws(t *testing.T) {
	f, p, now := tauFixture()
	offsets := []int64{10000}
	p.CandidateDelaysMS = &offsets
	p.MinProbability = 0
	p.Margin = .001
	checks := 0
	for qa := 0; qa <= 10; qa++ {
		for qb := 0; qb <= 10-qa; qb++ {
			q := []float64{float64(qa) / 10, float64(qb) / 10, float64(10-qa-qb) / 10}
			f.Points = []model.TauPoint{{Mark: "complete", DelayMS: 15000, Probability: q[0]}, {Mark: "complete", DelayMS: 17000, Probability: q[1]}}
			f.NoEvent = q[2]
			for pa := 0; pa <= 10; pa++ {
				for pb := 0; pb <= 10-pa; pb++ {
					truth := []float64{float64(pa) / 10, float64(pb) / 10, float64(10-pa-pb) / 10}
					f.TVRadius = (math.Abs(q[0]-truth[0]) + math.Abs(q[1]-truth[1]) + math.Abs(q[2]-truth[2])) / 2
					d, err := PlanTau(f, p, now)
					if err != nil {
						t.Fatal(err)
					}
					if !d.Schedule {
						continue
					}
					for _, offset := range []int64{10000, 11500, 13000} {
						useful := 0.0
						for i, point := range f.Points {
							if point.DelayMS >= offset+p.LeadMS && point.DelayMS <= offset+p.LeadMS+p.WindowMS {
								useful += truth[i]
							}
						}
						actual := (p.Benefit+p.FalseCost)*useful - p.FalseCost - p.TurnCost
						if actual+1e-14 < d.LowerUtility {
							t.Fatalf("bound %.17g exceeds actual %.17g q=%v p=%v d=%v", d.LowerUtility, actual, q, truth, now.Add(time.Duration(offset)*time.Millisecond))
						}
						checks++
					}
				}
			}
		}
	}
	if checks < 1000 {
		t.Fatalf("too few independently evaluated bounds: %d", checks)
	}
	t.Logf("checked %d law/delivery bounds", checks)
}
