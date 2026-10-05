package researchstats_test

import (
	"fmt"

	"github.com/JuanHuaXu/eventframed/internal/researchstats"
)

func ExampleConfidenceSequence_Anytime() {
	cs, err := researchstats.NewConfidenceSequence(.05, []string{"synthetic-benefit"})
	if err != nil {
		panic(err)
	}
	// A degenerate synthetic law: the outcome is always true, with constant
	// forecasts declared before it. Each completed stream is one observation.
	stream := []researchstats.BrierPair{{Control: .5, Candidate: .75, Outcome: true}}
	for n := 0; n < 4096; n++ {
		if err := cs.AddPairedStream("synthetic-benefit", stream); err != nil {
			panic(err)
		}
	}
	interval, err := cs.Anytime("synthetic-benefit")
	if err != nil {
		panic(err)
	}
	fmt.Printf("streams=%d mean=%.4f lower>0=%t\n", interval.N, interval.Mean, interval.Lower > 0)
	// Output: streams=4096 mean=0.1875 lower>0=true
}
