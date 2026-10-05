package observationgate

import (
	"math/rand"
	"sort"
	"testing"
	"time"
	"unsafe"
)

func TestLearnedContrastV3ComponentCost(t *testing.T) {
	var base [512]float64
	for x := range base {
		base[x] = v3Truth(0, uint16(x), 0)
	}
	if bytes := unsafe.Sizeof(v3State{}); bytes >= 4096 {
		t.Fatalf("state exceeds frozen 4 KiB cap: %d", bytes)
	}
	const n = 100000
	for repeat := 0; repeat < 3; repeat++ {
		state := newV3State()
		for x := uint16(0); x < 32; x++ {
			if err := state.observe(x, x&(1<<2) != 0, &base); err != nil {
				t.Fatal(err)
			}
		}
		updates := make([]int64, n)
		for i := range updates {
			x := uint16(i % 512)
			start := time.Now()
			if err := state.observe(x, x&(1<<2) != 0, &base); err != nil {
				t.Fatal(err)
			}
			updates[i] = time.Since(start).Nanoseconds()
		}
		rng := rand.New(rand.NewSource(int64(103 + repeat)))
		selections := make([]int64, n)
		selected := 0
		for i := range selections {
			x := uint16(i % 512)
			start := time.Now()
			p := state.forecast(x, base[x])
			if v3Nominate("learned_disagreement", x, base[x], .19, &state, 64, 256, rng) {
				selected++
			}
			if p <= 0 || p >= 1 {
				t.Fatal("invalid benchmark forecast")
			}
			selections[i] = time.Since(start).Nanoseconds()
		}
		null := newV3State()
		xRNG := rand.New(rand.NewSource(17))
		yRNG := rand.New(rand.NewSource(29))
		for i := 0; i < 32; i++ {
			if err := null.observe(uint16(xRNG.Intn(512)), yRNG.Intn(2) == 1, &base); err != nil {
				t.Fatal(err)
			}
		}
		if null.Weight[11] <= .5 {
			t.Fatal("benchmark null fallback did not activate")
		}
		fallbackRNG := rand.New(rand.NewSource(int64(211 + repeat)))
		fallback := make([]int64, n)
		for i := range fallback {
			x := uint16(i % 512)
			start := time.Now()
			p := null.forecast(x, base[x])
			_ = v3Nominate("learned_disagreement", x, base[x], .19, &null, 64, 256, fallbackRNG)
			if p <= 0 || p >= 1 {
				t.Fatal("invalid fallback forecast")
			}
			fallback[i] = time.Since(start).Nanoseconds()
		}
		sort.Slice(updates, func(i, j int) bool { return updates[i] < updates[j] })
		sort.Slice(selections, func(i, j int) bool { return selections[i] < selections[j] })
		sort.Slice(fallback, func(i, j int) bool { return fallback[i] < fallback[j] })
		updateP99 := updates[98999]
		selectionP99 := selections[98999]
		fallbackP99 := fallback[98999]
		t.Logf("repeat=%d stateBytes=%d updateP99NS=%d forecastSelectP99NS=%d fallbackP99NS=%d selected=%d",
			repeat, unsafe.Sizeof(v3State{}), updateP99, selectionP99, fallbackP99, selected)
		if updateP99 >= 10000 || selectionP99 >= 1000 || fallbackP99 >= 1000 {
			t.Fatalf("frozen component cost exceeded at repeat %d", repeat)
		}
	}
}
