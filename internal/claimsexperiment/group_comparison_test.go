package claimsexperiment

import (
	"context"
	"encoding/json"
	"testing"
)

func TestBayesianGroupComparisonProposesWithoutChangingAuthority(t *testing.T) {
	control, err := RunGroupAuthorityControl(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !control.Passed {
		t.Fatalf("group authority control failed: %+v", control)
	}
	record, err := json.Marshal(control)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(record))
}
