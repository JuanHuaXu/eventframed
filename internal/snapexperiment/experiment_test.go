package snapexperiment

import (
	"context"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/synthetictext"
)

func TestRunIsChronologicalAndKeepsUnchangedControlIdentical(t *testing.T) {
	records, _ := synthetictext.Build()
	cases, _, err := synthetictext.BuildSnapCases(records)
	if err != nil {
		t.Fatal(err)
	}
	wantedCaseID := "snap-" + records[0].Capture.Turn.SessionID
	var selected []synthetictext.SnapCase
	for _, testCase := range cases {
		if testCase.ID == wantedCaseID {
			selected = append(selected, testCase)
			break
		}
	}
	if len(selected) != 1 {
		t.Fatalf("case %s not found", wantedCaseID)
	}
	report, err := Run(context.Background(), records[:12], selected, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	if report.Queries != 2 || len(report.Results) != 12 {
		t.Fatalf("report shape = %+v", report)
	}
	var current, unchanged []QueryResult
	for _, result := range report.Results {
		if result.AvailableEvents != 9 && result.AvailableEvents != 11 {
			t.Fatalf("query %s saw %d events", result.QueryEventID, result.AvailableEvents)
		}
		switch result.Variant {
		case "current":
			current = append(current, result)
		case "unchanged":
			unchanged = append(unchanged, result)
		}
	}
	for index := range current {
		current[index].Variant = ""
		unchanged[index].Variant = ""
	}
	if !reflect.DeepEqual(current, unchanged) {
		t.Fatalf("unchanged control diverged:\ncurrent=%+v\nunchanged=%+v", current, unchanged)
	}
}

func TestRescueVerdictAppliesFrozenCriteria(t *testing.T) {
	verdict := rescueVerdict([]Aggregate{
		{Split: "confirmation", Variant: "current", HitRate: .9, ObsoleteHitRate: .4, MeanReciprocalRank: .8},
		{Split: "confirmation", Variant: "split_and_rewire", HitRate: .9, ObsoleteHitRate: .1, MeanReciprocalRank: .76},
	})
	if verdict.Status != "passed" || !verdict.RelevantHitNonInferior || !verdict.ObsoleteHitNonInferior || !verdict.StrictRateImprovement || !verdict.MRRWithinMargin {
		t.Fatalf("passing verdict = %+v", verdict)
	}
	failed := rescueVerdict([]Aggregate{
		{Split: "confirmation", Variant: "current", HitRate: .9, ObsoleteHitRate: .4, MeanReciprocalRank: .8},
		{Split: "confirmation", Variant: "split_and_rewire", HitRate: .8, ObsoleteHitRate: .1, MeanReciprocalRank: .8},
	})
	if failed.Status != "failed" || failed.RelevantHitNonInferior {
		t.Fatalf("failing verdict = %+v", failed)
	}
}
