package observationgate

import (
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"
)

func TestLearnedContrastV4Cost(t *testing.T) {
	if os.Getenv("EVENTFRAME_LEARNED_V4_COST") != "1" {
		t.Skip("opt-in component cost sample")
	}
	var base [512]float64
	for x := range base {
		base[x] = .05 + .9*float64(x%31)/30
	}
	state, selector := newV4State(), newLearnedState()
	rng := rand.New(rand.NewSource(20261024))
	const samples = 20000
	updates := make([]time.Duration, samples)
	selects := make([]time.Duration, samples)
	gateCount := 0
	for i := 0; i < samples; i++ {
		x := uint16(i % 512)
		start := time.Now()
		_, gated, err := state.forecast(x, base[x])
		if err != nil {
			t.Fatal(err)
		}
		if gated {
			gateCount++
		}
		v4Nominate("learned_disagreement", x, base[x], .6,
			selector.bestAlternative(), 128, 512-i%512, rng)
		selects[i] = time.Since(start)
		start = time.Now()
		if err := state.observe(x, i%2 == 0, &base); err != nil {
			t.Fatal(err)
		}
		if err := selector.observe(x, i%2 == 0, &base); err != nil {
			t.Fatal(err)
		}
		updates[i] = time.Since(start)
	}
	sort.Slice(updates, func(i, j int) bool { return updates[i] < updates[j] })
	sort.Slice(selects, func(i, j int) bool { return selects[i] < selects[j] })
	index := samples*99/100 - 1
	t.Logf("samples=%d update_p50=%s update_p99=%s forecast_select_p50=%s forecast_select_p99=%s gated=%d",
		samples, updates[samples/2], updates[index], selects[samples/2], selects[index], gateCount)
}
