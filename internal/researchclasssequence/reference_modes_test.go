package researchclasssequence

import (
	"math"
	"testing"

	ref "github.com/JuanHuaXu/eventframed/internal/researchclasssequenceref"
)

func TestIndependentReferenceModes(t *testing.T) {
	checks := 0
	for _, mode := range []string{"local", "noise", "shared", "hybrid"} {
		r, e := ref.New([]float64{.3, .5, .8}, mode, 1./16)
		if e != nil {
			t.Fatal(e)
		}
		for k := 0; k < 18; k++ {
			if e = r.Issue(k % 3); e != nil {
				t.Fatal(e)
			}
			if e = r.Observe(k%3, k/3+1, 1, k%4 == 0); e != nil {
				t.Fatal(e)
			}
		}
		for i := 0; i < 3; i++ {
			full, e := r.Query(i, 3)
			if e != nil {
				t.Fatal(e)
			}
			for _, field := range []string{"uncertainty", "information", "falsification", "predictive"} {
				x, e := r.QueryMode(i, 3, field)
				if e != nil {
					t.Fatal(e)
				}
				if math.Abs(x.Observed-full.Observed) > 2e-10 || math.Abs(x.Uncertainty-full.Uncertainty) > 2e-10 {
					t.Fatal("reference marginal", mode, field)
				}
				if (field == "information" || field == "falsification") && (math.Abs(x.Information-full.Information) > 2e-10 || math.Abs(x.EdgeCut-full.EdgeCut) > 2e-10) {
					t.Fatal("reference information", mode, field)
				}
				if field == "predictive" && math.Abs(x.Value-full.Value) > 2e-10 {
					t.Fatal("reference value", mode)
				}
				checks++
			}
		}
		if _, e = r.QueryMode(0, 3, "invalid"); e == nil {
			t.Fatal("mode guard")
		}
	}
	t.Log("mode-specific independent reference comparisons", checks)
}
