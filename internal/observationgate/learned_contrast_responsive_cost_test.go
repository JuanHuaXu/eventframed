package observationgate

import (
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"
)

func TestLearnedContrastResponsiveCost(t *testing.T) {
	if os.Getenv("EVENTFRAME_RESPONSIVE_COST") != "1" {
		t.Skip("opt-in component cost sample")
	}
	var base [512]float64
	for x := range base {
		base[x] = .05 + .9*float64(x%31)/30
	}
	state, monitor := newLearnedState(), newResponsiveMonitor()
	rng := rand.New(rand.NewSource(20261027))
	const samples = 20000
	updates, selects := make([]time.Duration, samples), make([]time.Duration, samples)
	for i := 0; i < samples; i++ {
		if i%512 == 0 {
			state, monitor = newLearnedState(), newResponsiveMonitor()
		}
		clock := i % 512
		x := uint16((i * 37) % 512)
		start := time.Now()
		p := state.forecast(x, base[x])
		if p <= 0 || p >= 1 {
			t.Fatalf("invalid forecast %g", p)
		}
		rate := .20
		if monitor.AlertAt >= 0 && clock > monitor.AlertAt && clock <= monitor.AlertAt+64 {
			rate = .60
		}
		responsiveNominate("learned_disagreement", x, base[x], .6, &state,
			128, 512-clock, rate, rng)
		selects[i] = time.Since(start)
		start = time.Now()
		y := i%2 == 0
		if err := state.observe(x, y, &base); err != nil {
			t.Fatal(err)
		}
		if err := monitor.observe(clock, x, y, base[x], clock); err != nil {
			t.Fatal(err)
		}
		updates[i] = time.Since(start)
	}
	sort.Slice(updates, func(i, j int) bool { return updates[i] < updates[j] })
	sort.Slice(selects, func(i, j int) bool { return selects[i] < selects[j] })
	index := samples*99/100 - 1
	t.Logf("samples=%d update_monitor_p50=%s update_monitor_p99=%s forecast_select_p50=%s forecast_select_p99=%s",
		samples, updates[samples/2], updates[index], selects[samples/2], selects[index])
}
