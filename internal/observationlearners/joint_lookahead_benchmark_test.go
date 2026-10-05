package observationlearners

import "testing"

// Paired policy cost on the same known parity law, without fitting or feedback.
// Quality on learned streams is measured by v103, not by this timing fixture.
func BenchmarkJointPlannerPair(b *testing.B) {
	m := lookaheadParity(b)
	for _, planner := range []bool{false, true} {
		name := "one_step"
		if planner {
			name = "lookahead"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r := &jointReader{x: uint16(i) & 511, epoch: 1}
				if planner {
					got, err := runJointLookahead(m, r, 1)
					if err != nil || got.Cost > 6 {
						b.Fatal(err)
					}
				} else {
					got, err := runJointObserver(m, r, 1)
					if err != nil || got.Cost > 6 {
						b.Fatal(err)
					}
				}
			}
		})
	}
}
