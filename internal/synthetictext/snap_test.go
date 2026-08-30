package synthetictext

import (
	"reflect"
	"testing"

	graphpolicy "github.com/JuanHuaXu/eventframed/internal/graph"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestBuildSnapCasesMatchesRuntimeGraphContract(t *testing.T) {
	records, _ := Build()
	cases, manifest, err := BuildSnapCases(records)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Cases != 32 || manifest.CandidateGraphs != 128 || manifest.SplitCases["design"] != 24 || manifest.SplitCases["confirmation"] != 8 {
		t.Fatalf("snap manifest = %+v", manifest)
	}
	policy := graphpolicy.Policy{MaxNodes: 8, MaxEdges: 8, MaxCandidateFamily: 8, ClosureRadius: 1, MinNetPriorityGain: .01, MaxProperRiskIncrease: .01, MaxUnresolvedBurden: 1, MinSimultaneousCoverage: .95, MinBucketSupport: 1}
	for _, testCase := range cases {
		if len(testCase.CandidateFamily) != 4 || len(testCase.DesignQueryIDs) != 2 || len(testCase.ConfirmationQueryIDs) != 2 {
			t.Fatalf("case %s shape is incomplete", testCase.ID)
		}
		for _, candidate := range testCase.CandidateFamily {
			if err := graphpolicy.ValidateCandidate(candidate.Graph, policy); err != nil {
				t.Fatalf("case %s candidate %s: %v", testCase.ID, candidate.ID, err)
			}
			for _, node := range candidate.Graph.Nodes {
				for _, eventID := range node.MemberEventIDs {
					if eventID[len(eventID)-2:] > "09" {
						t.Fatalf("case %s candidate %s leaks confirmation member %s", testCase.ID, candidate.ID, eventID)
					}
				}
			}
		}
		good := testCase.CandidateFamily[2].Graph
		foundSupersession := false
		for _, edge := range good.Edges {
			if edge.Effect == model.CompatibilityEffectSupersedes {
				foundSupersession = true
			}
		}
		if !foundSupersession {
			t.Fatalf("case %s has no directional supersession edge", testCase.ID)
		}
		closure := graphpolicy.DependencyClosure(testCase.CurrentGraph, good, policy.ClosureRadius)
		if !reflect.DeepEqual(closure, testCase.Oracle.ExpectedClosure) {
			t.Fatalf("case %s closure changed: got %+v want %+v", testCase.ID, closure, testCase.Oracle.ExpectedClosure)
		}
		if burden, err := graphpolicy.UnresolvedBurden(good, testCase.Obligations); err != nil || burden != 0 {
			t.Fatalf("case %s good candidate burden=%v err=%v", testCase.ID, burden, err)
		}
		if burden, err := graphpolicy.UnresolvedBurden(testCase.CandidateFamily[1].Graph, testCase.Obligations); err != nil || burden == 0 {
			t.Fatalf("case %s split-only candidate should retain unresolved burden: burden=%v err=%v", testCase.ID, burden, err)
		}
	}
}
