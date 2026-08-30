package claimsexperiment

import "testing"

func TestSafeResidualReplacementReportsAbstentionAndHarmBudget(t *testing.T) {
	report := RunSafeResidualReplacement(32452843)
	if report.AppliedCases == 0 || report.AbstainedCases == 0 {
		t.Fatalf("replacement did not exercise apply and abstain paths: %+v", report)
	}
	if report.AppliedCases+report.AbstainedCases != report.EvaluationCases {
		t.Fatalf("evaluation accounting = %+v", report)
	}
}
