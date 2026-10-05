package observationgate

import (
	"reflect"
	"testing"
)

func TestCoupledArrivalContracts(t *testing.T) {
	cfg := forestDelayCases()[0]
	base, err := forestDependenceBase(cfg, 2026092191)
	if err != nil {
		t.Fatal(err)
	}
	for _, delay := range []bool{false, true} {
		want, err := forestDelayRun(base, cfg, 0, 0, 2026092192, delay)
		if err != nil {
			t.Fatal(err)
		}
		got, err := coupledArrivalRun(base, cfg, 0, 0, 2026092192, delay)
		if err != nil {
			t.Fatal(err)
		}
		if got.Arms[0].Full != want.Arms[0].Full || got.Arms[0].Post != want.Arms[0].Post || got.Arms[0].SplitAt != want.Arms[0].SplitAt || !reflect.DeepEqual(got.Fits, want.Fits) {
			t.Fatal("control or fit drift")
		}
		for i, f := range got.Frames {
			g := want.Frames[i]
			if f.Predictions[0] != g.Predictions[0] || f.X != g.X || f.Y != g.Y || f.RX != g.RX || f.RY != g.RY || f.Arrival != g.Arrival || f.Missing != g.Missing {
				t.Fatal("control tape drift", i)
			}
			if f.Predictions[2].Mask != f.Predictions[0].Mask || f.Predictions[2].Cost != f.Predictions[0].Cost {
				t.Fatal("fixed view changed")
			}
			if !delay && (f.Predictions[1].P != f.Predictions[0].P || f.Predictions[2].P != f.Predictions[0].P || f.Predictions[1].Mask != f.Predictions[0].Mask) {
				t.Fatal("immediate behavior changed", i)
			}
		}
	}
}
