package claimsexperiment

import (
	"context"
	"testing"
)

func TestAntiPigeonExperimentSeparatesDivergentGroups(t *testing.T) {
	report, err := RunAntiPigeonGranularity(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	naive, oracle, separate := report.Variants["naive_shared"], report.Variants["oracle_ap"], report.Variants["separate"]
	if naive.FalseMergeRate != 1 || oracle.FalseMergeRate != 0 || separate.FalseMergeRate != 0 {
		t.Fatalf("unexpected merge rates: %+v", report.Variants)
	}
	if oracle.Brier >= naive.Brier || separate.Brier >= naive.Brier {
		t.Fatalf("granular posteriors did not improve loss: %+v", report.Variants)
	}
	if oracle.PosteriorKeys != 2 || separate.PosteriorKeys != 4 || naive.PosteriorKeys != 1 {
		t.Fatalf("unexpected posterior granularity: %+v", report.Variants)
	}
}
