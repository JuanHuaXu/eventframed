package rerankexperiment

import (
	"context"
	"reflect"
	"testing"
)

func TestGenerateIsDeterministicAndSeedSensitive(t *testing.T) {
	config := BlockConfig{Name: "confirmation", Seed: 17, BidirectionalRepeats: 1, RetentionRepeats: 1, EnvelopeRepeats: 2}
	left, err := Generate(config)
	if err != nil {
		t.Fatal(err)
	}
	right, err := Generate(config)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(left, right) {
		t.Fatal("same seed generated different datasets")
	}
	config.Seed++
	different, err := Generate(config)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(left.Cases[0].ContractOrder, different.Cases[0].ContractOrder) {
		t.Fatal("different seed did not change filler ordering")
	}
}

func TestActualServiceBidirectionalRepairAndBoundedNegativeControl(t *testing.T) {
	dataset, err := Generate(BlockConfig{Name: "confirmation", Seed: 23, BidirectionalRepeats: 1, RetentionRepeats: 1, EnvelopeRepeats: 5})
	if err != nil {
		t.Fatal(err)
	}
	report, err := Run(context.Background(), dataset)
	if err != nil {
		t.Fatal(err)
	}
	if report.Promotion.Value < .8 || report.Demotion.Value < .8 || report.JointRepair.Value < .75 {
		t.Fatalf("bidirectional mechanism failed: promotion=%+v demotion=%+v joint=%+v", report.Promotion, report.Demotion, report.JointRepair)
	}
	if report.Retention.Value != 1 {
		t.Fatalf("known-useful retention failed: %+v", report.Retention)
	}
	if report.EnvelopePromotion.Value > .05 {
		t.Fatalf("bounded correction crossed wide score gap: %+v", report.EnvelopePromotion)
	}
	for _, result := range report.Results {
		if result.Kind != "envelope" && result.ActiveBayesianApplied == 0 {
			t.Fatalf("case %s did not exercise committed Bayesian scoring", result.ID)
		}
	}
}
