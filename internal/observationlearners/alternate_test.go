package observationlearners

import (
	"math"
	"testing"
)

func TestAlternatePaths(t *testing.T) {
	base, e := Base(Scenarios[2], 2, 0)
	if e != nil {
		t.Fatal(e)
	}
	r, e := RunDependentStream(base, "design", 2, 0, 0, 2026104202, 2)
	if e != nil {
		t.Fatal(e)
	}
	out, e := RunAlternateRecord(r)
	if e != nil {
		t.Fatal(e)
	}
	if len(out.Ticks) == 0 {
		t.Fatal("empty diagnostic")
	}
	for _, tick := range out.Ticks {
		for a, v := range tick.Views {
			if v.Cost > 6 || v.Observed > 6 || v.Probability != tick.Predictions[a] || math.IsNaN(v.Probability) {
				t.Fatal("view bounds")
			}
			for _, s := range v.Trace {
				if s.Values != r.Inputs[tick.Step]&s.Observed {
					t.Fatal("hidden input mismatch")
				}
			}
		}
	}
	r.Ticks[511].Outcome = !r.Ticks[511].Outcome
	changed, e := RunAlternateRecord(r)
	if e != nil {
		t.Fatal(e)
	}
	for i := range out.Ticks {
		if out.Ticks[i].Predictions != changed.Ticks[i].Predictions {
			t.Fatal("current outcome influenced forecast")
		}
	}
}
