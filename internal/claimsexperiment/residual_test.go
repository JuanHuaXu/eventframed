package claimsexperiment

import (
	"context"
	"testing"
)

func TestResidualUtilityExperimentDetectsRecurringCorrection(t *testing.T) {
	report, err := RunResidualUtility(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if report.AbsoluteGain <= 0 || report.RelativeReduction < .1 {
		t.Fatalf("residual failed to improve held-out loss: %+v", report)
	}
	if report.ResidualAppliedRate < .9 {
		t.Fatalf("residual did not become reliably eligible: %+v", report)
	}
}
