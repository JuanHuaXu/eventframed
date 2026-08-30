package graph

import (
	"fmt"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func BenchmarkCandidateCompatibility200Nodes400Edges(b *testing.B) {
	predictive := model.PredictiveGraph{Nodes: make([]model.CompatibilityNode, 200), Edges: make([]model.CompatibilityEdge, 0, 400)}
	scores := make(map[string]float64, 200)
	for index := 0; index < 200; index++ {
		nodeID, eventID := fmt.Sprintf("node-%03d", index), fmt.Sprintf("event-%03d", index)
		predictive.Nodes[index] = model.CompatibilityNode{ID: nodeID, Kind: "bucket", MemberEventIDs: []string{eventID}, LawSpace: model.RetrievalUsefulnessHorizon}
		scores[eventID] = float64(index%100) / 100
	}
	for index := 0; index < 400; index++ {
		predictive.Edges = append(predictive.Edges, model.CompatibilityEdge{ID: fmt.Sprintf("edge-%03d", index), From: fmt.Sprintf("node-%03d", index%200), To: fmt.Sprintf("node-%03d", (index*37+1)%200), ComparisonMap: "identity_bernoulli", Weight: 1})
	}
	b.ReportAllocs()
	for b.Loop() {
		_ = CandidateCompatibility(predictive, scores)
	}
}
