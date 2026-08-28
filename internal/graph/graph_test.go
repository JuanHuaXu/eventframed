package graph

import (
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func testPolicy() Policy {
	return Policy{
		MaxNodes: 16, MaxEdges: 32, MaxCandidateFamily: 8, ClosureRadius: 1,
		MinNetPriorityGain: .01, MaxProperRiskIncrease: .01,
		MaxUnresolvedBurden: 0, MinSimultaneousCoverage: .95, MinBucketSupport: 30,
	}
}

func TestDependencyClosureIncludesChangedNodeAndUndirectedNeighbors(t *testing.T) {
	current := model.PredictiveGraph{
		Nodes: []model.CompatibilityNode{
			{ID: "a", Kind: "bucket", MemberEventIDs: []string{"event-a"}, PosteriorKeys: []string{"posterior-a"}, LawSpace: model.RetrievalUsefulnessHorizon},
			{ID: "b", Kind: "bucket", MemberEventIDs: []string{"event-b"}, PosteriorKeys: []string{"posterior-b"}, LawSpace: model.RetrievalUsefulnessHorizon},
			{ID: "c", Kind: "predictor", LawSpace: model.RetrievalUsefulnessHorizon},
		},
		Edges: []model.CompatibilityEdge{
			{ID: "ab", From: "a", To: "b", ComparisonMap: "identity_bernoulli", Weight: 1},
			{ID: "bc", From: "b", To: "c", ComparisonMap: "identity_bernoulli", Weight: 1},
		},
	}
	candidate := current
	candidate.Nodes = append([]model.CompatibilityNode(nil), current.Nodes...)
	candidate.Nodes[0].MemberEventIDs = []string{"event-a", "event-new"}

	closure := DependencyClosure(current, candidate, 1)
	if !reflect.DeepEqual(closure.NodeIDs, []string{"a", "b"}) {
		t.Fatalf("node closure = %v", closure.NodeIDs)
	}
	if !reflect.DeepEqual(closure.EdgeIDs, []string{"ab", "bc"}) {
		t.Fatalf("edge closure = %v", closure.EdgeIDs)
	}
	if !reflect.DeepEqual(closure.EventIDs, []string{"event-a", "event-b", "event-new"}) {
		t.Fatalf("event closure = %v", closure.EventIDs)
	}
	if !reflect.DeepEqual(closure.PosteriorKeys, []string{"posterior-a", "posterior-b"}) {
		t.Fatalf("posterior closure = %v", closure.PosteriorKeys)
	}
}

func TestCandidateValidationAndUnresolvedBurden(t *testing.T) {
	candidate := model.PredictiveGraph{
		Nodes: []model.CompatibilityNode{
			{ID: "a", Kind: "bucket", LawSpace: model.RetrievalUsefulnessHorizon},
			{ID: "b", Kind: "predictor", LawSpace: model.RetrievalUsefulnessHorizon},
			{ID: "c", Kind: "resolution", LawSpace: model.RetrievalUsefulnessHorizon},
		},
		Edges: []model.CompatibilityEdge{{ID: "ab", From: "a", To: "b", ComparisonMap: "identity_bernoulli", Weight: 1}},
	}
	if err := ValidateCandidate(candidate, testPolicy()); err != nil {
		t.Fatal(err)
	}
	burden, err := UnresolvedBurden(candidate, []model.ComparisonObligation{{From: "a", To: "b", Weight: 2}, {From: "a", To: "c", Weight: 3}})
	if err != nil || burden != 3 {
		t.Fatalf("burden = %v, %v", burden, err)
	}

	bad := candidate
	bad.Edges = append([]model.CompatibilityEdge(nil), candidate.Edges...)
	bad.Edges[0].ComparisonMap = "undeclared_projection"
	if err := ValidateCandidate(bad, testPolicy()); err == nil {
		t.Fatal("expected undeclared comparison map rejection")
	}
}

func TestSplitClosureCarriesOldAndNewPosteriorDependencies(t *testing.T) {
	current := model.PredictiveGraph{Nodes: []model.CompatibilityNode{{
		ID: "merged", Kind: "bucket", MemberEventIDs: []string{"event-a", "event-b"},
		PosteriorKeys: []string{"posterior-merged"}, LawSpace: model.RetrievalUsefulnessHorizon,
	}}}
	candidate := model.PredictiveGraph{
		Nodes: []model.CompatibilityNode{
			{ID: "split-a", Kind: "bucket", MemberEventIDs: []string{"event-a"}, PosteriorKeys: []string{"posterior-a"}, LawSpace: model.RetrievalUsefulnessHorizon},
			{ID: "split-b", Kind: "bucket", MemberEventIDs: []string{"event-b"}, PosteriorKeys: []string{"posterior-b"}, LawSpace: model.RetrievalUsefulnessHorizon},
		},
		Edges: []model.CompatibilityEdge{{ID: "split-compatibility", From: "split-a", To: "split-b", ComparisonMap: "identity_bernoulli", Weight: 1}},
	}

	closure := DependencyClosure(current, candidate, 1)
	if !reflect.DeepEqual(closure.EventIDs, []string{"event-a", "event-b"}) {
		t.Fatalf("split event closure = %v", closure.EventIDs)
	}
	if !reflect.DeepEqual(closure.PosteriorKeys, []string{"posterior-a", "posterior-b", "posterior-merged"}) {
		t.Fatalf("split posterior closure = %v", closure.PosteriorKeys)
	}
}
