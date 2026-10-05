package observationgate

import (
	"reflect"
	"testing"
)

func checkInnerArrivalEndpoints(t testing.TB, got innerArrivalResult, want coupledArrivalResult) {
	t.Helper()
	if !reflect.DeepEqual(got.Fits, want.Fits) {
		t.Fatal("fit drift")
	}
	for a, b := range map[int]int{0: 0, 1: 1, 3: 2} {
		if got.Arms[a].Full != want.Arms[b].Full || got.Arms[a].Post != want.Arms[b].Post || got.Arms[a].SplitAt != want.Arms[b].SplitAt {
			t.Fatal("endpoint score drift", a)
		}
		for i, f := range got.Frames {
			g := want.Frames[i]
			if f.Predictions[a] != g.Predictions[b] || f.X != g.X || f.Y != g.Y || f.RX != g.RX || f.RY != g.RY || f.Arrival != g.Arrival || f.Missing != g.Missing || f.Audit != g.Audit {
				t.Fatal("endpoint tape drift", a, i)
			}
		}
	}
}

func TestInnerArrivalContracts(t *testing.T) {
	cfg := forestDelayCases()[0]
	base, err := forestDependenceBase(cfg, 2026092191)
	if err != nil {
		t.Fatal(err)
	}
	for _, delay := range []bool{false, true} {
		want, err := coupledArrivalRun(base, cfg, 0, 0, 2026092192, delay)
		if err != nil {
			t.Fatal(err)
		}
		got, err := innerArrivalRun(base, cfg, 0, 0, 2026092192, delay)
		if err != nil {
			t.Fatal(err)
		}
		checkInnerArrivalEndpoints(t, got, want)
		if !delay {
			for i, f := range got.Frames {
				if f.Predictions[2].P != f.Predictions[0].P || f.Predictions[2].Mask != f.Predictions[0].Mask {
					t.Fatal("immediate full drift", i)
				}
			}
		}
	}
}
