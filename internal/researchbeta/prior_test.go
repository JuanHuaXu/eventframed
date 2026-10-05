package researchbeta

import (
	"math"
	"testing"
)

func TestPredictRationalPosteriors(t *testing.T) {
	for _, test := range []struct {
		mean, strength   float64
		success, failure uint64
		want             float64
	}{
		{.5, 2, 1, 0, 2.0 / 3}, {.5, 2, 0, 1, 1.0 / 3},
		{.75, 4, 2, 1, 5.0 / 7}, {.25, 4, 0, 2, 1.0 / 6},
		{.75, 4, 0, 0, .75},
	} {
		got, err := Predict(test.mean, test.strength, test.success, test.failure)
		if err != nil || math.Abs(got-test.want) > 1e-14 {
			t.Fatalf("got=%g want=%g err=%v", got, test.want, err)
		}
	}
}

func TestPredictRejectsInvalidModel(t *testing.T) {
	for _, test := range []struct {
		mean, strength   float64
		success, failure uint64
	}{
		{0, 2, 0, 0}, {1, 2, 0, 0}, {math.NaN(), 2, 0, 0},
		{.5, math.Inf(1), 0, 0}, {.5, 0, 0, 0}, {.5, -1, 0, 0},
		{math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64, 0, 0},
		{.5, 2, maxExactCount, 1}, {.5, 2, ^uint64(0), 0},
	} {
		if _, err := Predict(test.mean, test.strength, test.success, test.failure); err == nil {
			t.Fatal("accepted an invalid prior/count contract", test)
		}
	}
}

var benchmarkSink float64

func BenchmarkPredict32(b *testing.B) {
	b.ReportAllocs()
	var sum float64
	for i := 0; i < b.N; i++ {
		for j := 0; j < 32; j++ {
			success, failure := uint64(j%2), uint64(1-j%2)
			mean := .6 + .01*float64(j)
			prediction, err := Predict(mean, 2, success, failure)
			if err != nil {
				b.Fatal(err)
			}
			sum += prediction
		}
	}
	benchmarkSink = sum
}
