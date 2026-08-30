package ranking

import "testing"

func TestElasticDeltaScaleLearnsFasterUnderUncertainty(t *testing.T) {
	policy := DefaultElasticDeltaPolicy()
	lowConfidence := policy.Scale(.2, 1)
	highConfidence := policy.Scale(.9, 1)
	if lowConfidence <= highConfidence {
		t.Fatalf("low-confidence scale %v did not exceed high-confidence scale %v", lowConfidence, highConfidence)
	}
	if got := policy.Scale(.75, 1); got != 1 {
		t.Fatalf("reference scale = %v, want 1", got)
	}
	if got := policy.Scale(0, .2); got != .5 {
		t.Fatalf("low-reliability uncertainty scale = %v, want .5", got)
	}
	if got := policy.Scale(0, 0); got != 0 {
		t.Fatalf("uncertified correction scale = %v, want 0", got)
	}
}

func TestDisabledElasticDeltaKeepsReliabilityGate(t *testing.T) {
	policy := ElasticDeltaPolicy{Enabled: false, MinScale: .5, MaxScale: 2.5}
	for _, confidence := range []float64{0, .5, 1} {
		if got := policy.Scale(confidence, 1); got != 1 {
			t.Fatalf("disabled reliable scale(%v) = %v, want 1", confidence, got)
		}
		if got := policy.Scale(confidence, 0); got != 0 {
			t.Fatalf("disabled uncertified scale(%v) = %v, want 0", confidence, got)
		}
	}
}
