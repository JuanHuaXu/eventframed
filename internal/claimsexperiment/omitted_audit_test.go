package claimsexperiment

import "testing"

func TestOmittedAuditCoverageExperimentIsAuditable(t *testing.T) {
	report := RunOmittedAuditCoverage(49979687)
	if report.Errors != 0 || report.CoveredTrials != report.Trials || report.MeanUpperBound < report.MeanTrueInfluence {
		t.Fatalf("coverage report = %+v", report)
	}
}
