// Operational counterfactual audit. The working window mixture is not assumed
// to be a common stationary Bayesian posterior or to satisfy a tower identity.
package researchretentioncoherence

import (
	"encoding/json"
	bank "github.com/JuanHuaXu/eventframed/internal/researchwindowjournal"
	"math"
	"os"
	"testing"
)

type row struct {
	Origin, Member                                                            int
	ObservationProbability, Before, AfterOne, AfterZero, Mixture, TowerDefect float64
}

func TestOperationalMixtureTowerDiagnostic(t *testing.T) {
	path := os.Getenv("EVENTFRAME_RETENTION_TOWER_OUT")
	if path == "" {
		t.Skip("explicit isolated artifact")
	}
	base := []float64{.3, .47, .7, .9}
	windows := [3]int{2, 8, 24}
	var rows []row
	for _, origin := range []int{0, 4, 12, 20, 27} {
		var models [3]*bank.Bank
		var targets [3]bank.BankTicket
		for k := range models {
			var e error
			models[k], e = bank.NewBank(base, 1, 128, 2, windows)
			if e != nil {
				t.Fatal(e)
			}
			for n := 0; n < 28; n++ {
				x, e := models[k].Issue(n%4, int64(n))
				if e != nil {
					t.Fatal(e)
				}
				if _, e = models[k].Resolve(x, (n < 12 && n%3 != 0) || (n >= 12 && n%5 == 0), int64(n)); e != nil {
					t.Fatal(e)
				}
				if n == origin {
					targets[k] = x
				}
			}
		}
		p, e := models[0].QuerySecond(targets[0])
		if e != nil {
			t.Fatal(e)
		}
		for k := 1; k < 3; k++ {
			x, e := models[k].RequestSecond(targets[k], 28)
			if e != nil {
				t.Fatal(e)
			}
			if x.Forecast() != p {
				t.Fatal("prebranch laws differ")
			}
			if _, e = models[k].Resolve(x, k == 1, 28); e != nil {
				t.Fatal(e)
			}
		}
		for i := range base {
			before, _, e := models[0].Predict(i)
			if e != nil {
				t.Fatal(e)
			}
			a, _, e := models[1].Predict(i)
			if e != nil {
				t.Fatal(e)
			}
			b, _, e := models[2].Predict(i)
			if e != nil {
				t.Fatal(e)
			}
			mix := p*a + (1-p)*b
			rows = append(rows, row{origin, i, p, before, a, b, mix, mix - before})
		}
	}
	max := 0.
	for _, r := range rows {
		if !isFinite(r.TowerDefect) {
			t.Fatal("nonfinite")
		}
		max = math.Max(max, math.Abs(r.TowerDefect))
	}
	b, e := json.MarshalIndent(struct {
		Rows              []row
		MaxAbsoluteDefect float64
		Scope             string
	}{rows, max, "public API counterfactual diagnostic; no target truth, no theorem or acquisition gain"}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if _, e = f.Write(append(b, '\n')); e != nil {
		t.Fatal(e)
	}
	if e = f.Sync(); e != nil {
		t.Fatal(e)
	}
	t.Log("counterfactual checks", len(rows), "maximum tower defect", max)
}
func isFinite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
