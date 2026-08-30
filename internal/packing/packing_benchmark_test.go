package packing

import "testing"

func BenchmarkAdaptiveDiversityPacking100(b *testing.B) {
	candidates := fixtureCandidates(100)
	policy := DefaultPolicy()
	policy.AdaptiveEnabled, policy.DiversityEnabled = true, true
	for b.Loop() {
		_ = Select(candidates, nil, 10, 100, 2_000, policy)
	}
}
