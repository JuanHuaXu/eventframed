package main

import (
	"math"
	"testing"
)

func TestOracleMassAndFullView(t *testing.T) {
	for _, mode := range []string{"uniform", "latent"} {
		o, err := newOracle(mode)
		if err != nil {
			t.Fatal(err)
		}
		for _, local := range []bool{false, true} {
			prior, err := o.selectedView(0, 0, local)
			if err != nil || math.Abs(prior-0.5) > 1e-12 {
				t.Fatalf("prior %s local=%v: %.15g, %v", mode, local, prior, err)
			}
			for x := 0; x < 512; x++ {
				got, err := o.selectedView(511, uint16(x), local)
				if err != nil || math.Abs(got-truthProbability(uint16(x), local)) > 1e-12 {
					t.Fatalf("full view %s local=%v x=%d: %.15g, %v", mode, local, x, got, err)
				}
			}
		}
	}
}

func TestLatentMassCorrelation(t *testing.T) {
	same := 0.0
	for x := 0; x < 512; x++ {
		if bit(uint16(x), 0) == bit(uint16(x), 1) {
			same += inputMass(uint16(x), "latent")
		}
	}
	want := 0.85*0.85 + 0.15*0.15
	if math.Abs(same-want) > 1e-12 {
		t.Fatalf("latent correlation mass %.15g, want %.15g", same, want)
	}
}
