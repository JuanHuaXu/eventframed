package api

import (
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// openClawEvent is the deliberately narrow memory projection returned to the
// OpenClaw adapter. Semantic fields and embeddings remain daemon-internal.
type openClawEvent struct {
	ID          string           `json:"id"`
	TenantID    string           `json:"tenant_id"`
	SessionID   string           `json:"session_id"`
	Sequence    uint64           `json:"sequence"`
	Kind        string           `json:"kind"`
	Content     string           `json:"content"`
	OccurredAt  time.Time        `json:"occurred_at"`
	ObservedAt  time.Time        `json:"observed_at"`
	AvailableAt time.Time        `json:"available_at"`
	Priority    float64          `json:"priority"`
	Tags        []string         `json:"tags,omitempty"`
	Provenance  model.Provenance `json:"provenance"`
}

type openClawCandidate struct {
	Event                          openClawEvent        `json:"event"`
	Similarity                     float64              `json:"similarity"`
	RecencyScore                   float64              `json:"recency_score"`
	GraphCompatibility             float64              `json:"graph_compatibility"`
	GraphApplied                   bool                 `json:"graph_applied"`
	BaselineScore                  float64              `json:"baseline_score"`
	PredictiveScore                float64              `json:"predictive_score"`
	RetrievalScore                 float64              `json:"retrieval_score"`
	RankDelta                      float64              `json:"rank_delta"`
	RankDeltaScale                 float64              `json:"rank_delta_scale,omitempty"`
	RankDeltaAnswerCertainty       float64              `json:"rank_delta_answer_certainty,omitempty"`
	RankDeltaCorrectionReliability float64              `json:"rank_delta_correction_reliability,omitempty"`
	RankDeltaConfidence            float64              `json:"rank_delta_confidence,omitempty"`
	RankDeltaBasis                 string               `json:"rank_delta_basis,omitempty"`
	ResolutionRankDelta            float64              `json:"resolution_rank_delta,omitempty"`
	RetrievalContract              string               `json:"retrieval_contract"`
	Score                          float64              `json:"score"`
	BayesianProbability            float64              `json:"bayesian_probability,omitempty"`
	BayesianApplied                bool                 `json:"bayesian_applied"`
	Forecast                       model.ForecastBundle `json:"forecast"`
	EstimatedTokens                int                  `json:"estimated_tokens"`
}

type openClawContextPacket struct {
	ProtocolVersion       string                     `json:"protocol_version"`
	Candidates            []openClawCandidate        `json:"candidates"`
	Recalled              int                        `json:"recalled"`
	Eligible              int                        `json:"eligible"`
	Packed                int                        `json:"packed"`
	UsedTokens            int                        `json:"used_tokens"`
	AdaptiveExpanded      bool                       `json:"adaptive_expanded"`
	CorrelatedSuppressed  int                        `json:"correlated_suppressed"`
	PacketConfidence      float64                    `json:"packet_confidence"`
	PacketAnswerCertainty float64                    `json:"packet_answer_certainty"`
	RetrievalContract     string                     `json:"retrieval_contract"`
	NominationContract    string                     `json:"nomination_contract"`
	Snapshot              model.Snapshot             `json:"snapshot"`
	BayesianShadow        model.BayesianShadowReport `json:"bayesian_shadow"`
}

func projectOpenClawPacket(packet model.ContextPacket) openClawContextPacket {
	candidates := make([]openClawCandidate, 0, len(packet.Candidates))
	for _, candidate := range packet.Candidates {
		event := candidate.Event
		candidates = append(candidates, openClawCandidate{
			Event: openClawEvent{
				ID: event.ID, TenantID: event.TenantID, SessionID: event.SessionID, Sequence: event.Sequence,
				Kind: event.Kind, Content: event.Content, OccurredAt: event.OccurredAt, ObservedAt: event.ObservedAt,
				AvailableAt: event.AvailableAt, Priority: event.Priority, Tags: event.Tags, Provenance: event.Provenance,
			},
			Similarity: candidate.Similarity, RecencyScore: candidate.RecencyScore,
			GraphCompatibility: candidate.GraphCompatibility, GraphApplied: candidate.GraphApplied,
			BaselineScore: candidate.BaselineScore, PredictiveScore: candidate.PredictiveScore,
			RetrievalScore: candidate.RetrievalScore, RankDelta: candidate.RankDelta, RankDeltaScale: candidate.RankDeltaScale,
			RankDeltaAnswerCertainty:       candidate.RankDeltaAnswerCertainty,
			RankDeltaCorrectionReliability: candidate.RankDeltaCorrectionReliability,
			RankDeltaConfidence:            candidate.RankDeltaConfidence, RankDeltaBasis: candidate.RankDeltaBasis,
			ResolutionRankDelta: candidate.ResolutionRankDelta,
			RetrievalContract:   candidate.RetrievalContract, Score: candidate.Score,
			BayesianProbability: candidate.BayesianProbability, BayesianApplied: candidate.BayesianApplied,
			Forecast: candidate.Forecast, EstimatedTokens: candidate.EstimatedTokens,
		})
	}
	return openClawContextPacket{
		ProtocolVersion: packet.ProtocolVersion, Candidates: candidates, Recalled: packet.Recalled, Eligible: packet.Eligible,
		Packed: packet.Packed, UsedTokens: packet.UsedTokens, AdaptiveExpanded: packet.AdaptiveExpanded, CorrelatedSuppressed: packet.CorrelatedSuppressed,
		PacketConfidence: packet.PacketConfidence, PacketAnswerCertainty: packet.PacketAnswerCertainty,
		RetrievalContract: packet.RetrievalContract, NominationContract: packet.NominationContract,
		Snapshot: packet.Snapshot, BayesianShadow: packet.BayesianShadow,
	}
}
