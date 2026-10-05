package observationgate

import (
	"math/rand"
	"testing"
)

var learnedBenchProbability float64
var learnedBenchSelected bool

func BenchmarkLearnedContrastV2(b *testing.B) {
	var base [512]float64
	for x := range base {
		bit := ((x&(1<<6) != 0) != (x&(1<<7) != 0)) != (x&(1<<8) != 0)
		if bit {
			base[x] = .95
		} else {
			base[x] = .05
		}
	}
	state := newLearnedState()
	for x := uint16(0); x < 32; x++ {
		if err := state.observe(x, x&(1<<2) != 0, &base); err != nil {
			b.Fatal(err)
		}
	}
	b.Run("rolling-update", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			x := uint16(i % 512)
			if err := state.observe(x, x&(1<<2) != 0, &base); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("forecast-and-select", func(b *testing.B) {
		rng := rand.New(rand.NewSource(101))
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			x := uint16(i % 512)
			learnedBenchProbability = state.forecast(x, base[x])
			learnedBenchSelected = learnedNominate("learned_disagreement", x, base[x], .19,
				&state, 64, 256, rng)
		}
	})
}
