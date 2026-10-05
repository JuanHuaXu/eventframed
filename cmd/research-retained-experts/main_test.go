package main

import (
	"math"
	"testing"
)

func TestScoreAccumulator(t *testing.T) {
	var one scores
	if err := one.add(0.2, [4]float64{0.1, 0.3, 0.4, 0.5},
		[4]float64{0.6, 0.7, 0.7, 0.7}, 4, true); err != nil {
		t.Fatal(err)
	}
	if one.N != 1 || math.Abs(one.Final-0.64) > 1e-12 ||
		math.Abs(one.Tree-0.09) > 1e-12 || one.ObservedBits != 4 {
		t.Fatalf("unexpected score: %+v", one)
	}
	var both scores
	both.accumulate(one)
	both.accumulate(one)
	normalized := both.normalized()
	if normalized.N != 2 || math.Abs(normalized.Final-0.64) > 1e-12 ||
		math.Abs(normalized.Tree-0.09) > 1e-12 || normalized.ObservedBits != 4 {
		t.Fatalf("unexpected aggregate: %+v", normalized)
	}
}
