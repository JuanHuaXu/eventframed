package claimsexperiment

import (
	"math"
	"testing"
)

func TestWilsonIntervalMatchesPublishedV4Bounds(t *testing.T) {
	tests := []struct {
		successes, trials int
		lower, upper      float64
	}{
		{64, 64, .9434, 1},
		{56, 64, .7723, .9353},
		{0, 64, 0, .0566},
		{99, 128, .6936, .8374},
		{60, 64, .8500, .9754},
	}
	for _, test := range tests {
		interval := wilsonInterval(test.successes, test.trials)
		if math.Abs(interval.Lower-test.lower) > .0001 || math.Abs(interval.Upper-test.upper) > .0001 {
			t.Fatalf("wilson(%d,%d) = [%f,%f], want [%f,%f]", test.successes, test.trials, interval.Lower, interval.Upper, test.lower, test.upper)
		}
	}
}

func TestGroupComparisonCarriesIntervalsAndFrozenCriteria(t *testing.T) {
	report := RunGroupComparison(69867970)
	strong := report.Scenarios["split_0.9_0.1"]
	if strong.ExpectedDecisionCount != 64 || !strong.Acceptance.Passed {
		t.Fatalf("strong split report = %+v", strong)
	}
	shared := report.Scenarios["shared_0.8_0.8"]
	if shared.ExpectedDecisionRate < .8 || !shared.Acceptance.Passed || shared.WrongCount != 0 {
		t.Fatalf("equivalence replacement did not rescue compatible recognition: %+v", shared)
	}
}
