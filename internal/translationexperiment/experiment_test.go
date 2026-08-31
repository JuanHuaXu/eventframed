package translationexperiment

import "testing"

func TestPublicChainControls(t *testing.T) {
	for _, split := range []string{"design", "confirmation"} {
		result, err := Run(split)
		if err != nil {
			t.Fatal(err)
		}
		if result.Cases != 24 || result.Passed != result.Cases {
			t.Fatalf("%s result: passed %d/%d: %+v", split, result.Passed, result.Cases, result.Results)
		}
	}
}
