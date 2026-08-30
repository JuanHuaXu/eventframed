package bayes

import (
	"crypto/sha256"
	"encoding/binary"
	"sort"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

type Policy struct {
	VectorWeight, NeighborWeight, NoveltyWeight, IndependenceWeight float64
	Threshold, CriticalThreshold, AuditProbability                  float64
	MaxActive                                                       int
	AuditSeed                                                       string
	CheapUpdateAll                                                  bool
}

type Candidate struct {
	EventID                                                             string
	VectorRelevance, NeighborCompatibility, Novelty, SourceIndependence float64
	Priority                                                            float64
	EvidenceReady                                                       bool
}

func Evaluate(candidates []Candidate, epoch uint64, policy Policy) model.BayesianShadowReport {
	decisions := make([]model.BayesianDecision, 0, len(candidates))
	for _, candidate := range candidates {
		score := policy.VectorWeight*clamp(candidate.VectorRelevance) + policy.NeighborWeight*clamp(candidate.NeighborCompatibility) + policy.NoveltyWeight*clamp(candidate.Novelty) + policy.IndependenceWeight*clamp(candidate.SourceIndependence)
		threshold := policy.Threshold
		if candidate.Priority >= 0.8 {
			threshold = policy.CriticalThreshold
		}
		deepReview := candidate.EvidenceReady && score >= threshold
		cheapUpdate := deepReview
		if policy.CheapUpdateAll {
			cheapUpdate = candidate.EvidenceReady
		}
		decisions = append(decisions, model.BayesianDecision{EventID: candidate.EventID, ActivationScore: clamp(score), Activated: cheapUpdate, CheapUpdate: cheapUpdate, DeepReview: deepReview, EvidenceReady: candidate.EvidenceReady, AuditSelected: audit(candidate.EventID, epoch, policy), AuditProbability: clamp(policy.AuditProbability), PosteriorKey: candidate.EventID})
	}
	sort.SliceStable(decisions, func(i, j int) bool { return decisions[i].ActivationScore > decisions[j].ActivationScore })
	active, deep := 0, 0
	for index := range decisions {
		if decisions[index].Activated {
			active++
		}
		if !decisions[index].DeepReview {
			continue
		}
		if deep >= policy.MaxActive {
			decisions[index].DeepReview = false
			if !policy.CheapUpdateAll {
				decisions[index].Activated = false
				decisions[index].CheapUpdate = false
				active--
			}
			continue
		}
		deep++
	}
	return model.BayesianShadowReport{Mode: "shadow", Nominated: len(candidates), Activated: active, DeepReviewed: deep, SelectionSupportCertified: false, Decisions: decisions}
}

func audit(eventID string, epoch uint64, policy Policy) bool {
	probability := clamp(policy.AuditProbability)
	if probability <= 0 {
		return false
	}
	if probability >= 1 {
		return true
	}
	digest := sha256.Sum256([]byte(policy.AuditSeed + "\x00" + eventID + "\x00" + string(binary.BigEndian.AppendUint64(nil, epoch))))
	draw := float64(binary.BigEndian.Uint64(digest[:8])) / float64(^uint64(0))
	return draw < probability
}

func clamp(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
