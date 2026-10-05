package observationgate

import (
	"math/rand"
	"testing"
)

func BenchmarkCreditLearningFixture(b *testing.B) {
	base, m, err := arrivalSwitchSetup(1, 3, 0)
	if err != nil {
		b.Fatal(err)
	}
	cfg := forestDependenceCase{subsetBreadthCase{Name: arrivalSwitchNames[3], Mode: "stable", Target: "bit"}, 0}
	outcome := func(x uint16, c int, ref bool, r *rand.Rand) bool {
		y := arrivalSwitchTruth(x, m, 3, c, ref)
		if r.Float64() < .05 {
			y = !y
		}
		return y
	}
	for _, delay := range []bool{false, true} {
		for _, credit := range []bool{false, true} {
			name := "immediate"
			if delay {
				name = "delayed"
			}
			if credit {
				name += "/credit"
			} else {
				name += "/original"
			}
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if credit {
						if _, _, e := creditLearningRun(base, cfg, 3, 0, 2026092531, delay, outcome, true); e != nil {
							b.Fatal(e)
						}
					} else {
						if _, e := innerArrivalRunOutcome(base, cfg, 3, 0, 2026092531, delay, outcome); e != nil {
							b.Fatal(e)
						}
					}
				}
			})
		}
	}
}
