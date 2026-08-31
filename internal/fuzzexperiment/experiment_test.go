package fuzzexperiment_test

import (
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/fuzzexperiment"
)

func TestPublicConfirmationSeparatesEnvelopeParaphraseAndCrossDomainSensitivity(t *testing.T) {
	result, err := fuzzexperiment.Run("../../testdata/text-public-facts/corpus.jsonl", "confirmation")
	if err != nil {
		t.Fatal(err)
	}
	if result.Queries != 32 || result.Trials != 88 || len(result.Properties) != 3 {
		t.Fatalf("experiment shape = %+v", result)
	}
	reports := make(map[string]struct {
		mean      float64
		stable    float64
		invariant bool
	})
	for _, report := range result.Properties {
		reports[report.PropertyID] = struct {
			mean      float64
			stable    float64
			invariant bool
		}{report.MeanTotalVariation, report.StableFraction, report.ConditionalInvariant}
	}
	if reports["context-envelope"].invariant || reports["same-answer-paraphrase-bundle"].invariant || reports["cross-domain-semantic-bundle"].invariant {
		t.Fatalf("confirmation family should remain provisional under simultaneous bounds: %+v", reports)
	}
	if reports["context-envelope"].stable != 1 || reports["same-answer-paraphrase-bundle"].stable != 1 || reports["cross-domain-semantic-bundle"].stable >= 1 {
		t.Fatalf("point stability = %+v", reports)
	}
	if reports["context-envelope"].mean >= reports["same-answer-paraphrase-bundle"].mean || reports["same-answer-paraphrase-bundle"].mean >= reports["cross-domain-semantic-bundle"].mean {
		t.Fatalf("sensitivity ordering = %+v", reports)
	}
}
