package ranking

import (
	"math"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

type Policy struct {
	ContextualEnabled   bool
	HierarchicalEnabled bool
	BaselineWeight      float64
	PosteriorWeight     float64
	RecencyWeight       float64
	GraphWeight         float64
	ParentStrength      float64
	MaxParentWeight     float64
}

type Features struct {
	Baseline     float64
	Posterior    float64
	Recency      float64
	Graph        float64
	HasPosterior bool
	HasGraph     bool
}

func DefaultPolicy() Policy {
	return Policy{BaselineWeight: .85, PosteriorWeight: .10, RecencyWeight: .05, GraphWeight: .05, ParentStrength: 2, MaxParentWeight: .10}
}

func Score(features Features, policy Policy) float64 {
	if !policy.ContextualEnabled {
		posteriorWeight, graphWeight := 0.0, 0.0
		if features.HasPosterior {
			posteriorWeight = .10
		}
		if features.HasGraph {
			graphWeight = policy.GraphWeight
		}
		baselineWeight := math.Max(0, 1-posteriorWeight-graphWeight)
		return clamp(baselineWeight*features.Baseline + posteriorWeight*features.Posterior + graphWeight*features.Graph)
	}
	baselineWeight := policy.BaselineWeight
	posteriorWeight := policy.PosteriorWeight
	if !features.HasPosterior {
		baselineWeight += posteriorWeight
		posteriorWeight = 0
	}
	graphWeight := policy.GraphWeight
	if !features.HasGraph {
		baselineWeight += graphWeight
		graphWeight = 0
	}
	total := baselineWeight + posteriorWeight + policy.RecencyWeight + graphWeight
	if total <= 0 {
		return clamp(features.Baseline)
	}
	return clamp((baselineWeight*features.Baseline + posteriorWeight*features.Posterior + policy.RecencyWeight*features.Recency + graphWeight*features.Graph) / total)
}

func HierarchicalMean(child, parent model.BayesianPosterior, policy Policy) float64 {
	childSupport := math.Max(0, child.Alpha+child.Beta-2)
	childUseful := math.Max(0, child.Alpha-1)
	if child.Alpha <= 0 || child.Beta <= 0 {
		childSupport, childUseful = 0, 0
	}
	parentMean := parent.Mean()
	if !policy.HierarchicalEnabled || parent.Alpha <= 0 || parent.Beta <= 0 || policy.ParentStrength <= 0 || policy.MaxParentWeight <= 0 {
		return child.Mean()
	}
	strength := policy.ParentStrength
	if childSupport > 0 {
		strength = math.Min(strength, policy.MaxParentWeight*childSupport/(1-policy.MaxParentWeight))
	}
	return clamp((childUseful + strength*parentMean + 1) / (childSupport + strength + 2))
}

func clamp(value float64) float64 {
	return math.Min(1, math.Max(0, value))
}
