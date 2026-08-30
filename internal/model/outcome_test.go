package model

import "testing"

func TestOutcomeSignalsUseConservativePrecedence(t *testing.T) {
	yes := true
	tests := []struct {
		name                     string
		signals                  OutcomeSignals
		wantUseful, wantEvidence bool
	}{
		{name: "citation", signals: OutcomeSignals{Cited: true}, wantUseful: true, wantEvidence: true},
		{name: "rejection dominates", signals: OutcomeSignals{Cited: true, Rejected: true}, wantEvidence: true},
		{name: "correction dominates", signals: OutcomeSignals{SuccessfulDownstream: true, Correction: true}, wantEvidence: true},
		{name: "explicit", signals: OutcomeSignals{ExplicitUseful: &yes}, wantUseful: true, wantEvidence: true},
		{name: "packed only", signals: OutcomeSignals{Packed: true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useful, evidence := test.signals.Resolve(false)
			if useful != test.wantUseful || evidence != test.wantEvidence {
				t.Fatalf("resolve=(%v,%v), want (%v,%v)", useful, evidence, test.wantUseful, test.wantEvidence)
			}
		})
	}
}
