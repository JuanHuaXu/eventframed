package researchstats

import "testing"

var benchmarkInterval Interval
var benchmarkMean float64

func BenchmarkConfidenceSequence(b *testing.B) {
	cs, err := NewConfidenceSequence(.05, []string{"gain", "guard"})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := cs.AddStream("gain", .01); err != nil {
			b.Fatal(err)
		}
		benchmarkInterval, err = cs.Anytime("gain")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPairedBrierMean(b *testing.B) {
	pairs := make([]BrierPair, 512)
	for i := range pairs {
		pairs[i] = BrierPair{Control: .5, Candidate: .75, Outcome: i%2 == 0}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var err error
		benchmarkMean, err = PairedBrierMean(pairs)
		if err != nil {
			b.Fatal(err)
		}
	}
}
