package graph

import (
	"errors"
	"math"
	"reflect"
	"sort"
	"strings"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

type Policy struct {
	MaxNodes, MaxEdges, MaxCandidateFamily, ClosureRadius int
	MinNetPriorityGain, MaxProperRiskIncrease             float64
	MaxUnresolvedBurden, MinSimultaneousCoverage          float64
	MinBucketSupport                                      float64
}

func (policy Policy) Valid() bool {
	return policy.MaxNodes > 0 && policy.MaxEdges >= 0 && policy.MaxCandidateFamily > 0 && policy.ClosureRadius >= 0 && finite(policy.MinNetPriorityGain) && policy.MinNetPriorityGain > 0 && finite(policy.MaxProperRiskIncrease) && policy.MaxProperRiskIncrease >= 0 && finite(policy.MaxUnresolvedBurden) && policy.MaxUnresolvedBurden >= 0 && policy.MinSimultaneousCoverage > 0 && policy.MinSimultaneousCoverage <= 1 && policy.MinBucketSupport > 0
}

func ValidateCandidate(candidate model.PredictiveGraph, policy Policy) error {
	if len(candidate.Nodes) > policy.MaxNodes || len(candidate.Edges) > policy.MaxEdges {
		return errors.New("predictive graph exceeds bounded node or edge capacity")
	}
	nodes := make(map[string]struct{}, len(candidate.Nodes))
	for _, node := range candidate.Nodes {
		if strings.TrimSpace(node.ID) == "" || node.LawSpace != model.RetrievalUsefulnessHorizon {
			return errors.New("every node requires an id and retrieval-usefulness-v1 law space")
		}
		switch node.Kind {
		case "bucket", "predictor", "resolution", "agent":
		default:
			return errors.New("unsupported predictive node kind")
		}
		if _, exists := nodes[node.ID]; exists {
			return errors.New("predictive node ids must be unique")
		}
		nodes[node.ID] = struct{}{}
		if hasDuplicateOrEmpty(node.MemberEventIDs) || hasDuplicateOrEmpty(node.PosteriorKeys) {
			return errors.New("node members and posterior keys must be non-empty when present and unique")
		}
	}
	edges := make(map[string]struct{}, len(candidate.Edges))
	for _, edge := range candidate.Edges {
		if strings.TrimSpace(edge.ID) == "" || edge.From == edge.To || edge.ComparisonMap != "identity_bernoulli" || !finite(edge.Weight) || edge.Weight <= 0 {
			return errors.New("edges require distinct endpoints, positive weight, and identity_bernoulli comparison")
		}
		if _, ok := nodes[edge.From]; !ok {
			return errors.New("edge source is absent from candidate graph")
		}
		if _, ok := nodes[edge.To]; !ok {
			return errors.New("edge target is absent from candidate graph")
		}
		if _, exists := edges[edge.ID]; exists {
			return errors.New("predictive edge ids must be unique")
		}
		edges[edge.ID] = struct{}{}
	}
	return nil
}

func DependencyClosure(current, candidate model.PredictiveGraph, radius int) model.DependencyClosure {
	currentNodes, candidateNodes := nodeMap(current), nodeMap(candidate)
	seed := make(map[string]struct{})
	for id, node := range currentNodes {
		if next, ok := candidateNodes[id]; !ok || !reflect.DeepEqual(node, next) {
			seed[id] = struct{}{}
		}
	}
	for id, node := range candidateNodes {
		if previous, ok := currentNodes[id]; !ok || !reflect.DeepEqual(previous, node) {
			seed[id] = struct{}{}
		}
	}
	currentEdges, candidateEdges := edgeMap(current), edgeMap(candidate)
	for id, edge := range currentEdges {
		if next, ok := candidateEdges[id]; !ok || !reflect.DeepEqual(edge, next) {
			seed[edge.From], seed[edge.To] = struct{}{}, struct{}{}
		}
	}
	for id, edge := range candidateEdges {
		if previous, ok := currentEdges[id]; !ok || !reflect.DeepEqual(previous, edge) {
			seed[edge.From], seed[edge.To] = struct{}{}, struct{}{}
		}
	}
	adjacency := make(map[string][]string)
	for _, graph := range []model.PredictiveGraph{current, candidate} {
		for _, edge := range graph.Edges {
			adjacency[edge.From] = append(adjacency[edge.From], edge.To)
			adjacency[edge.To] = append(adjacency[edge.To], edge.From)
		}
	}
	affected := seed
	frontier := mapKeys(seed)
	for depth := 0; depth < radius; depth++ {
		next := make([]string, 0)
		for _, id := range frontier {
			for _, neighbor := range adjacency[id] {
				if _, seen := affected[neighbor]; seen {
					continue
				}
				affected[neighbor] = struct{}{}
				next = append(next, neighbor)
			}
		}
		frontier = next
	}
	closure := model.DependencyClosure{NodeIDs: mapKeys(affected)}
	edgeIDs, eventIDs, posteriorKeys := make(map[string]struct{}), make(map[string]struct{}), make(map[string]struct{})
	for _, graph := range []model.PredictiveGraph{current, candidate} {
		for _, edge := range graph.Edges {
			if _, left := affected[edge.From]; left {
				edgeIDs[edge.ID] = struct{}{}
			} else if _, right := affected[edge.To]; right {
				edgeIDs[edge.ID] = struct{}{}
			}
		}
		for _, node := range graph.Nodes {
			if _, ok := affected[node.ID]; !ok {
				continue
			}
			for _, id := range node.MemberEventIDs {
				eventIDs[id] = struct{}{}
			}
			for _, key := range node.PosteriorKeys {
				posteriorKeys[key] = struct{}{}
			}
		}
	}
	closure.EdgeIDs, closure.EventIDs, closure.PosteriorKeys = mapKeys(edgeIDs), mapKeys(eventIDs), mapKeys(posteriorKeys)
	return closure
}

func UnresolvedBurden(candidate model.PredictiveGraph, obligations []model.ComparisonObligation) (float64, error) {
	adjacency := make(map[string][]string)
	nodes := nodeMap(candidate)
	for _, edge := range candidate.Edges {
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
		adjacency[edge.To] = append(adjacency[edge.To], edge.From)
	}
	burden := 0.0
	for _, obligation := range obligations {
		if obligation.From == obligation.To || !finite(obligation.Weight) || obligation.Weight <= 0 {
			return 0, errors.New("comparison obligations require distinct endpoints and positive weight")
		}
		if _, left := nodes[obligation.From]; !left || !reachable(adjacency, obligation.From, obligation.To) {
			burden += obligation.Weight
		}
	}
	return burden, nil
}

func reachable(adjacency map[string][]string, from, to string) bool {
	seen, queue := map[string]struct{}{from: {}}, []string{from}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range adjacency[current] {
			if next == to {
				return true
			}
			if _, ok := seen[next]; !ok {
				seen[next] = struct{}{}
				queue = append(queue, next)
			}
		}
	}
	return false
}

func nodeMap(graph model.PredictiveGraph) map[string]model.CompatibilityNode {
	out := make(map[string]model.CompatibilityNode)
	for _, value := range graph.Nodes {
		out[value.ID] = value
	}
	return out
}
func edgeMap(graph model.PredictiveGraph) map[string]model.CompatibilityEdge {
	out := make(map[string]model.CompatibilityEdge)
	for _, value := range graph.Edges {
		out[value.ID] = value
	}
	return out
}
func mapKeys[T any](values map[string]T) []string {
	out := make([]string, 0, len(values))
	for key := range values {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
func hasDuplicateOrEmpty(values []string) bool {
	seen := make(map[string]struct{})
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
		if _, ok := seen[value]; ok {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}
func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
