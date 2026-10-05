package observationlearners

import (
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/researchinput"
	"math/rand"
	"testing"
)

func TestForestAdapterParity(t *testing.T) {
	for seed := int64(0); seed < 20; seed++ {
		rng := rand.New(rand.NewSource(2026092290 + seed))
		samples := make([]observation.Sample, 64)
		for i := range samples {
			samples[i] = observation.Sample{Bits: uint16(rng.Intn(512)), Outcome: rng.Intn(2) == 1}
		}
		a, err := fitHeldoutInputForest(samples)
		if err != nil {
			t.Fatal(err)
		}
		b, err := researchinput.Fit(samples)
		if err != nil || a.weights != b {
			t.Fatal("adapter changed candidate", err)
		}
	}
}
