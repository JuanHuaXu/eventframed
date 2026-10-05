package observationlearners

import (
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
)

func TestRetainedControlsAndJournal(t *testing.T) {
	for _, j := range []int{0, 2, 6, 7, 8} {
		b, e := Base(Scenarios[j], j, 0)
		if e != nil {
			t.Fatal(e)
		}
		old, e := RunPartialStream(b, "unit", j, 0, 0, 2026100101)
		if e != nil {
			t.Fatal(e)
		}
		r, e := RunRetainedStream(b, "unit", j, 0, 0, 2026100101)
		if e != nil {
			t.Fatal(e)
		}
		var inner bayes.ForecastMix
		for i, tick := range r.Ticks {
			for a := 0; a < 2; a++ {
				if tick.Predictions[a] != old.Ticks[i].Predictions[a] || !reflect.DeepEqual(r.Views[i][a], old.Views[i][a]) {
					t.Fatal("control drift")
				}
			}
			if tick.Predictions[2].Experts[1] != inner.Forecast(r.Inner[i][2]) {
				t.Fatal("inner forecast is not pre-feedback")
			}
			for _, origin := range tick.Delivered {
				if origin+Scenarios[j].Delay != i || r.Ticks[origin].Missing {
					t.Fatal("availability violation")
				}
				inner = inner.Observe(r.Inner[origin][2], r.Ticks[origin].Outcome, 1)
			}
			for _, v := range r.Views[i] {
				if v.Cost > 6 || v.Observed > 6 {
					t.Fatal("live budget")
				}
			}
		}
	}
}
