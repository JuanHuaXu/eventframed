// External API audit of full-length histories. No model-private state is read.
package researchmeanratiocheck

import (
	"fmt"
	"math"
	"testing"

	ref "github.com/JuanHuaXu/eventframed/internal/researchmeanjointref"
	model "github.com/JuanHuaXu/eventframed/internal/researchmeanratio"
)

func near(t *testing.T, label string, a, b float64) {
	t.Helper()
	if math.IsNaN(a) || math.IsNaN(b) || math.IsInf(a, 0) || math.IsInf(b, 0) || math.Abs(a-b) > 2e-11 {
		t.Fatalf("%s %.17g != %.17g", label, a, b)
	}
}
func TestFullJournalOldQueries(t *testing.T) {
	checks := 0
	for _, mode := range []string{"noise", "local", "individual"} {
		for _, family := range []string{"learn", "baseline", "current", "free"} {
			for _, hazard := range []float64{0, 1. / 16, 1} {
				t.Run(fmt.Sprintf("%s/%s/%g", mode, family, hazard), func(t *testing.T) {
					base := []float64{.25, .925}
					m, e := model.New(base, 1, 256, model.Config{Mode: mode, Family: family, Means: "learn", Hazard: hazard})
					if e != nil {
						t.Fatal(e)
					}
					r, e := ref.New(base, mode, family, "learn", hazard)
					if e != nil {
						t.Fatal(e)
					}
					var tickets [128]model.Ticket
					compare := func() {
						w, e := m.ModelWeights()
						if e != nil {
							t.Fatal(e)
						}
						v := r.ModelWeights()
						for k := range w {
							near(t, "long family", w[k], v[k])
							checks++
						}
						for i := 0; i < 2; i++ {
							q, o, e := m.Predict(i)
							if e != nil {
								t.Fatal(e)
							}
							u, v, e := r.Predict(i)
							if e != nil {
								t.Fatal(e)
							}
							near(t, "long clean", q, u)
							near(t, "long W1", o, v)
							checks += 2
						}
					}
					for n := 0; n < 128; n++ {
						compare()
						tickets[n], e = m.Issue(n%2, int64(n))
						if e != nil {
							t.Fatal(e)
						}
						if e = r.Issue(n % 2); e != nil {
							t.Fatal(e)
						}
						if n >= 11 {
							j := n - 11
							value := j%19 < 8
							if _, e = m.Resolve(tickets[j], value, int64(n)); e != nil {
								t.Fatal(e)
							}
							if e = r.Observe(j%2, j/2+1, 1, value); e != nil {
								t.Fatal(e)
							}
						}
					}
					at := int64(128)
					for j := 117; j < 128; j++ {
						if _, e = m.Resolve(tickets[j], j%19 < 8, at); e != nil {
							t.Fatal(e)
						}
						if e = r.Observe(j%2, j/2+1, 1, j%19 < 8); e != nil {
							t.Fatal(e)
						}
						at++
					}
					for _, j := range []int{0, 1, 20, 63, 126, 127} {
						for _, query := range []string{"forecast", "uncertainty", "information", "falsification", "predictive", "model_class", "noise_class"} {
							a, e := m.Query(tickets[j], query)
							if e != nil {
								t.Fatal(e)
							}
							b, e := r.QueryMode(j%2, j/2+1, query)
							if e != nil {
								t.Fatal(e)
							}
							for k, x := range []float64{a.Observed, a.Uncertainty, a.Information, a.EdgeCut, a.Value, a.ClassGain} {
								y := []float64{b.Observed, b.Uncertainty, b.Information, b.EdgeCut, b.Value, b.ClassGain}[k]
								near(t, "long query "+query, x, y)
								checks++
							}
						}
						a, e := m.RequestSecond(tickets[j], at)
						if e != nil {
							t.Fatal(e)
						}
						if e = r.Audit(j%2, j/2+1); e != nil {
							t.Fatal(e)
						}
						// Deliberate disagreement removes zero-noise support from the whole
						// conditional model, even when a rate reset could revive a rate atom.
						value := !(j%19 < 8)
						if _, e = m.Resolve(a, value, at+1); e != nil {
							t.Fatal(e)
						}
						if e = r.Observe(j%2, j/2+1, 2, value); e != nil {
							t.Fatal(e)
						}
						compare()
						at += 2
					}
					if m.Pending() != 0 {
						t.Fatal("terminal pending")
					}
					if _, e = m.Issue(0, at); e == nil {
						t.Fatal("full journal cap")
					}
				})
			}
		}
	}
	t.Logf("%d counted long-history scalar comparisons in 36 configurations", checks)
}
