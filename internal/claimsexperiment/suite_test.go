package claimsexperiment

import (
	"context"
	"testing"
)

func TestConfirmationRunRequiresDistinctDesignSeed(t *testing.T) {
	seed := int64(10)
	if _, err := RunSuiteWithOptions(context.Background(), RunOptions{SeedBase: seed, Role: RunRoleConfirmation}); err == nil {
		t.Fatal("confirmation without design seed succeeded")
	}
	if _, err := RunSuiteWithOptions(context.Background(), RunOptions{SeedBase: seed, Role: RunRoleConfirmation, DesignSeedBase: &seed}); err == nil {
		t.Fatal("confirmation reused design seed")
	}
}

func TestV5SuiteEmbedsProtocolAndIntegrationControl(t *testing.T) {
	designSeed := int64(982451653)
	report, err := RunSuiteWithOptions(context.Background(), RunOptions{SeedBase: 67867967, Role: RunRoleConfirmation, DesignSeedBase: &designSeed})
	if err != nil {
		t.Fatal(err)
	}
	if report.SchemaVersion != SchemaVersion || report.Protocol.ID != ProtocolID {
		t.Fatalf("missing v5 protocol: %+v", report)
	}
	if report.Run.ConfirmationSeedDistinct == nil || !*report.Run.ConfirmationSeedDistinct {
		t.Fatalf("confirmation provenance = %+v", report.Run)
	}
	if !report.IntegrationControls.GroupAuthority.Passed {
		t.Fatalf("group authority control = %+v", report.IntegrationControls.GroupAuthority)
	}
}
