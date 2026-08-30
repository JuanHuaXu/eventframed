package claimsexperiment

import "testing"

func TestChangepointExperimentUsesBoundedProductionUpdater(t *testing.T) {
	report := RunChangepointAdaptation()
	if len(report.Scenarios) != 6 || report.Trajectories != changepointTrajectories {
		t.Fatalf("incomplete report: %+v", report)
	}
	stable := report.Scenarios["stable"]
	if stable.ExpectedChanges != 0 || stable.DetectedChanges != 0 {
		t.Fatalf("stable scenario has expected changes: %+v", stable)
	}
	if stable.DetectionInterval != nil || stable.MeanDelayBasis != DetectedChangeDelayBasis {
		t.Fatalf("stable scenario has invalid evidence semantics: %+v", stable)
	}
	for name, scenario := range report.Scenarios {
		if scenario.DetectedChanges > scenario.ExpectedChanges {
			t.Fatalf("%s matched too many changes: %+v", name, scenario)
		}
		if scenario.MissRate < 0 || scenario.MissRate > 1 {
			t.Fatalf("%s invalid miss rate: %+v", name, scenario)
		}
		if scenario.TotalTriggers != scenario.DetectedChanges+scenario.FalseAlarms {
			t.Fatalf("%s trigger denominator mismatch: %+v", name, scenario)
		}
		if scenario.DelaySampleCount != scenario.DetectedChanges {
			t.Fatalf("%s delay denominator mismatch: %+v", name, scenario)
		}
		if !scenario.Acceptance.Evaluated {
			t.Fatalf("%s missing frozen acceptance evaluation: %+v", name, scenario)
		}
	}
}
