package ranking

import "testing"

func BenchmarkContextualScore(b *testing.B) {
	policy := DefaultPolicy()
	policy.ContextualEnabled = true
	features := Features{Baseline: .7, Posterior: .8, Recency: .9, HasPosterior: true}
	for b.Loop() {
		_ = Score(features, policy)
	}
}
