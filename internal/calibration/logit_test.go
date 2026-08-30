package calibration_test

import (
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/calibration"
)

func TestFitRepairsOverconfidentProbabilitiesWithoutChangingOrder(t *testing.T) {
	observations := []calibration.Observation{
		{Probability: .9, Outcome: 0, Weight: 1}, {Probability: .8, Outcome: 0, Weight: 1},
		{Probability: .7, Outcome: 1, Weight: 1}, {Probability: .6, Outcome: 0, Weight: 1},
		{Probability: .55, Outcome: 1, Weight: 1}, {Probability: .5, Outcome: 0, Weight: 1},
	}
	before := calibration.Brier(observations, calibration.Identity())
	fitted, err := calibration.Fit(observations)
	if err != nil {
		t.Fatal(err)
	}
	if after := calibration.Brier(observations, fitted); after >= before {
		t.Fatalf("Brier did not improve: before=%f after=%f calibrator=%+v", before, after, fitted)
	}
	if fitted.Apply(.8) <= fitted.Apply(.7) {
		t.Fatal("monotonic calibrator changed ranking order")
	}
}

func TestFitDoesNotRunAwayOnImbalancedObservations(t *testing.T) {
	observations := make([]calibration.Observation, 0, 101)
	for range 100 {
		observations = append(observations, calibration.Observation{Probability: .2, Outcome: 0, Weight: 1})
	}
	observations = append(observations, calibration.Observation{Probability: .8, Outcome: 1, Weight: 1})

	fitted, err := calibration.Fit(observations)
	if err != nil {
		t.Fatal(err)
	}
	if !fitted.Valid() || fitted.Scale == .01 && fitted.Bias == 20 {
		t.Fatalf("runaway calibrator: %+v", fitted)
	}
	if fitted.Apply(.2) >= fitted.Apply(.8) {
		t.Fatalf("monotonic order changed: %+v", fitted)
	}
}

func TestComposePreservesSequentialCalibration(t *testing.T) {
	inner := calibration.Logit{Scale: 2, Bias: -1, Floor: 1e-6}
	outer := calibration.Logit{Scale: .5, Bias: .25, Floor: 1e-6}
	combined := calibration.Compose(outer, inner)
	for _, probability := range []float64{.01, .2, .5, .8, .99} {
		want := outer.Apply(inner.Apply(probability))
		if got := combined.Apply(probability); math.Abs(got-want) > 1e-12 {
			t.Fatalf("compose(%v) = %v, want %v", probability, got, want)
		}
	}
}

func TestInverseRecoversCalibratorInput(t *testing.T) {
	calibrator := calibration.Logit{Scale: 2.5, Bias: -1.25, Floor: 1e-6}
	for _, probability := range []float64{.01, .2, .5, .8, .99} {
		if got := calibrator.Inverse(calibrator.Apply(probability)); math.Abs(got-probability) > 1e-12 {
			t.Fatalf("inverse(apply(%v)) = %v", probability, got)
		}
	}
}
