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

func BenchmarkEvidenceOccupancyPacking50(b *testing.B) {
	candidates := fixtureCandidates(50)
	policy := DefaultPolicy()
	for b.Loop() {
		_ = Select(candidates, nil, 10, 50, 2_000, policy)
	}
}
