package main

import (
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/evaluation"
)

func TestSyntheticClaimsV2FrozenResult(t *testing.T) {
	dataset, err := runExperiment()
	if err != nil {
		t.Fatal(err)
	}
	report, err := evaluation.Evaluate(dataset)
	if err != nil {
		t.Fatal(err)
	}
	if report.Cases != 120 || report.Trajectories != 4 {
		t.Fatalf("experiment dimensions = %+v", report)
	}
	updateAll := report.Comparisons["update_all"]
	frontierAllDeepSelective := report.Comparisons["frontier_all_deep_selective"]
	selective := report.Comparisons["selective"]
	eventframe := report.Comparisons["eventframe"]
	if updateAll.PriorityWeightedBrierGain < .03 {
		t.Fatalf("update-all mechanism gain disappeared: %+v", updateAll)
	}
	if math.Abs(frontierAllDeepSelective.PriorityWeightedBrierGain-updateAll.PriorityWeightedBrierGain) > 1e-15 || math.Abs(frontierAllDeepSelective.RecallAt10Gain-updateAll.RecallAt10Gain) > 1e-15 {
		t.Fatalf("deep-selective replacement diverged from frontier-all: replacement=%+v update_all=%+v", frontierAllDeepSelective, updateAll)
	}
	if selective.PriorityWeightedBrierGain <= 0 || selective.PriorityWeightedBrierGain >= .001 {
		t.Fatalf("selective result moved outside frozen v2 bounds: %+v", selective)
	}
	if math.Abs(eventframe.PriorityWeightedBrierGain-selective.PriorityWeightedBrierGain) > 1e-15 {
		t.Fatalf("residual ablation unexpectedly diverged: eventframe=%+v selective=%+v", eventframe, selective)
	}
}
