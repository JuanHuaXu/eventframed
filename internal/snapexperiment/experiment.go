package snapexperiment

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	graphpolicy "github.com/JuanHuaXu/eventframed/internal/graph"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"github.com/JuanHuaXu/eventframed/internal/ranking"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/synthetictext"
)

const ProtocolVersion = "eventframe.public-fact-sheaf-snap-eval.v1"

type Options struct {
	EmbeddingDimension int     `json:"embedding_dimension"`
	RecallK            int     `json:"recall_k"`
	PackK              int     `json:"pack_k"`
	GraphWeight        float64 `json:"graph_weight"`
}

func DefaultOptions() Options {
	return Options{EmbeddingDimension: 256, RecallK: 16, PackK: 3, GraphWeight: .05}
}

type QueryResult struct {
	CaseID          string   `json:"case_id"`
	Split           string   `json:"split"`
	Variant         string   `json:"variant"`
	QueryEventID    string   `json:"query_event_id"`
	RelevantIDs     []string `json:"relevant_ids"`
	ObsoleteIDs     []string `json:"obsolete_ids,omitempty"`
	PackedIDs       []string `json:"packed_ids"`
	RelevantHit     bool     `json:"relevant_hit"`
	ObsoleteHit     bool     `json:"obsolete_hit"`
	ReciprocalRank  float64  `json:"reciprocal_rank"`
	GraphApplied    int      `json:"graph_applied_candidates"`
	AvailableEvents int      `json:"available_events"`
}

type Aggregate struct {
	Split                string  `json:"split"`
	Variant              string  `json:"variant"`
	Queries              int     `json:"queries"`
	HitRate              float64 `json:"hit_rate"`
	ObsoleteHitRate      float64 `json:"obsolete_hit_rate"`
	MeanReciprocalRank   float64 `json:"mean_reciprocal_rank"`
	GraphApplicationRate float64 `json:"graph_application_rate"`
}

type Report struct {
	ProtocolVersion string        `json:"protocol_version"`
	Method          string        `json:"method"`
	Options         Options       `json:"options"`
	Cases           int           `json:"cases"`
	Queries         int           `json:"queries"`
	Results         []QueryResult `json:"results"`
	Aggregates      []Aggregate   `json:"aggregates"`
	RescueVerdict   RescueVerdict `json:"rescue_verdict"`
}

type RescueVerdict struct {
	Status                   string  `json:"status"`
	Comparison               string  `json:"comparison"`
	Split                    string  `json:"split"`
	RelevantHitNonInferior   bool    `json:"relevant_hit_non_inferior"`
	ObsoleteHitNonInferior   bool    `json:"obsolete_hit_non_inferior"`
	StrictRateImprovement    bool    `json:"strict_rate_improvement"`
	MRRWithinMargin          bool    `json:"mrr_within_margin"`
	AllowedMRRDecrease       float64 `json:"allowed_mrr_decrease"`
	ObservedRelevantHitDelta float64 `json:"observed_relevant_hit_delta"`
	ObservedObsoleteHitDelta float64 `json:"observed_obsolete_hit_delta"`
	ObservedMRRDelta         float64 `json:"observed_mrr_delta"`
}

type topologyVariant struct {
	name  string
	graph model.PredictiveGraph
}

func Run(ctx context.Context, records []synthetictext.Record, cases []synthetictext.SnapCase, options Options) (Report, error) {
	if options.EmbeddingDimension <= 0 || options.RecallK <= 0 || options.PackK <= 0 || options.PackK > options.RecallK || options.GraphWeight < 0 || options.GraphWeight > 1 {
		return Report{}, errors.New("invalid experiment options")
	}
	bySession := make(map[string][]synthetictext.Record)
	for _, record := range records {
		sessionID := record.Capture.Turn.SessionID
		bySession[sessionID] = append(bySession[sessionID], record)
	}
	results := make([]QueryResult, 0, len(cases)*12)
	for _, testCase := range cases {
		sessionID := sessionForCase(testCase.ID)
		session := append([]synthetictext.Record(nil), bySession[sessionID]...)
		sort.Slice(session, func(i, j int) bool { return session[i].Capture.Turn.Sequence < session[j].Capture.Turn.Sequence })
		if len(session) != 12 {
			return Report{}, fmt.Errorf("case %s has %d source turns, want 12", testCase.ID, len(session))
		}
		variants := []topologyVariant{
			{name: "no_graph", graph: model.PredictiveGraph{TenantID: testCase.TenantID}},
			{name: "current", graph: testCase.CurrentGraph},
		}
		for _, candidate := range testCase.CandidateFamily {
			variants = append(variants, topologyVariant{name: candidate.ID, graph: candidate.Graph})
		}
		for _, variant := range variants {
			variantResults, err := runVariant(ctx, testCase, session, variant, options)
			if err != nil {
				return Report{}, fmt.Errorf("case %s variant %s: %w", testCase.ID, variant.name, err)
			}
			results = append(results, variantResults...)
		}
	}
	report := Report{
		ProtocolVersion: ProtocolVersion,
		Method:          "Fresh in-memory runtime per case and topology; capture turns 1-9, install fixed topology, recall before capturing turns 10 and 12, and capture turn 11 between them. Oracle and source annotations are never ingested.",
		Options:         options, Cases: len(cases), Queries: len(cases) * 2, Results: results,
	}
	report.Aggregates = aggregate(results)
	report.RescueVerdict = rescueVerdict(report.Aggregates)
	return report, nil
}

func runVariant(ctx context.Context, testCase synthetictext.SnapCase, session []synthetictext.Record, variant topologyVariant, options Options) ([]QueryResult, error) {
	memory := memorystore.New()
	embedder, err := embed.NewHashEmbedder(options.EmbeddingDimension)
	if err != nil {
		return nil, err
	}
	rankingPolicy := ranking.DefaultPolicy()
	rankingPolicy.GraphWeight = options.GraphWeight
	runtime, err := service.New(memory, embedder, service.Config{
		DefaultRecallK: options.RecallK, DefaultPackK: options.PackK, DefaultTokenBudget: 100_000,
		OverfetchMultiplier: 1, RankingPolicy: rankingPolicy,
		PackingPolicy: packing.Policy{MaxPack: options.PackK}, ResidualMode: service.ResidualModeDisabled,
	})
	if err != nil {
		return nil, err
	}
	for index := 0; index < 9; index++ {
		if _, err := runtime.CaptureTurn(ctx, session[index].Capture); err != nil {
			return nil, fmt.Errorf("capture design turn %d: %w", index+1, err)
		}
	}
	previous, err := memory.GetPredictiveGraph(ctx, testCase.TenantID)
	if err != nil {
		return nil, err
	}
	published := variant.graph
	published.TenantID = testCase.TenantID
	published.Version = 0
	published.PublishedAt = session[8].Capture.Turn.AvailableAt
	published.SourceSnapID = "fixed-eval:" + testCase.ID + ":" + variant.name
	closure := graphpolicy.DependencyClosure(previous, published, 1)
	_, _, err = memory.PublishPredictiveSnap(ctx, model.PredictiveSnapRecord{
		ID: published.SourceSnapID, TenantID: testCase.TenantID, PreviousGraph: previous,
		PublishedGraph: published, Closure: closure, SimultaneousCoverage: 1,
		Procedure: "fixed-topology benchmark; no publication-admission inference", Issuer: "eventframe-synthetic-snap-eval",
		PublishedAt: session[8].Capture.Turn.AvailableAt,
	})
	if err != nil {
		return nil, fmt.Errorf("install topology: %w", err)
	}
	results := make([]QueryResult, 0, 2)
	for index := 9; index < len(session); index++ {
		record := session[index]
		if record.Oracle != nil {
			result, err := recall(ctx, runtime, testCase, variant.name, record, options, index)
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		}
		if index < len(session)-1 {
			if _, err := runtime.CaptureTurn(ctx, record.Capture); err != nil {
				return nil, fmt.Errorf("capture turn %d: %w", index+1, err)
			}
		}
	}
	if len(results) != 2 {
		return nil, fmt.Errorf("got %d confirmation queries, want 2", len(results))
	}
	return results, nil
}

func recall(ctx context.Context, runtime *service.Service, testCase synthetictext.SnapCase, variant string, record synthetictext.Record, options Options, available int) (QueryResult, error) {
	packet, err := runtime.Recall(ctx, model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: record.Capture.Turn.TenantID,
		SessionID: record.Capture.Turn.SessionID, Query: record.Oracle.QueryBeforeCapture,
		AsOf: record.Capture.Turn.OccurredAt, RecallK: options.RecallK, PackK: options.PackK, TokenBudget: 100_000,
	})
	if err != nil {
		return QueryResult{}, err
	}
	relevant := stringSet(record.Oracle.RelevantPriorIDs)
	obsolete := stringSet(record.Oracle.ObsoletePriorIDs)
	result := QueryResult{
		CaseID: testCase.ID, Split: testCase.Split, Variant: variant, QueryEventID: record.Capture.Turn.ID,
		RelevantIDs: append([]string(nil), record.Oracle.RelevantPriorIDs...), ObsoleteIDs: append([]string(nil), record.Oracle.ObsoletePriorIDs...),
		AvailableEvents: available,
	}
	for index, candidate := range packet.Candidates {
		result.PackedIDs = append(result.PackedIDs, candidate.Event.ID)
		if candidate.GraphApplied {
			result.GraphApplied++
		}
		if _, ok := relevant[candidate.Event.ID]; ok {
			result.RelevantHit = true
			if result.ReciprocalRank == 0 {
				result.ReciprocalRank = 1 / float64(index+1)
			}
		}
		if _, ok := obsolete[candidate.Event.ID]; ok {
			result.ObsoleteHit = true
		}
	}
	return result, nil
}

func aggregate(results []QueryResult) []Aggregate {
	type sums struct {
		queries, hits, obsolete, graphApplied, packed int
		rr                                            float64
	}
	values := make(map[string]*sums)
	for _, result := range results {
		for _, split := range []string{result.Split, "all"} {
			key := split + "\x00" + result.Variant
			if values[key] == nil {
				values[key] = &sums{}
			}
			value := values[key]
			value.queries++
			if result.RelevantHit {
				value.hits++
			}
			if result.ObsoleteHit {
				value.obsolete++
			}
			value.rr += result.ReciprocalRank
			value.graphApplied += result.GraphApplied
			value.packed += len(result.PackedIDs)
		}
	}
	aggregates := make([]Aggregate, 0, len(values))
	for key, value := range values {
		split, variant := splitKey(key)
		aggregate := Aggregate{Split: split, Variant: variant, Queries: value.queries}
		aggregate.HitRate = float64(value.hits) / float64(value.queries)
		aggregate.ObsoleteHitRate = float64(value.obsolete) / float64(value.queries)
		aggregate.MeanReciprocalRank = value.rr / float64(value.queries)
		if value.packed > 0 {
			aggregate.GraphApplicationRate = float64(value.graphApplied) / float64(value.packed)
		}
		aggregates = append(aggregates, aggregate)
	}
	sort.Slice(aggregates, func(i, j int) bool {
		if aggregates[i].Split == aggregates[j].Split {
			return aggregates[i].Variant < aggregates[j].Variant
		}
		return aggregates[i].Split < aggregates[j].Split
	})
	return aggregates
}

func rescueVerdict(aggregates []Aggregate) RescueVerdict {
	const margin = .05
	verdict := RescueVerdict{
		Status: "not_evaluated", Comparison: "split_and_rewire-current",
		Split: "confirmation", AllowedMRRDecrease: margin,
	}
	var current, candidate *Aggregate
	for index := range aggregates {
		value := &aggregates[index]
		if value.Split != verdict.Split {
			continue
		}
		switch value.Variant {
		case "current":
			current = value
		case "split_and_rewire":
			candidate = value
		}
	}
	if current == nil || candidate == nil {
		return verdict
	}
	verdict.ObservedRelevantHitDelta = candidate.HitRate - current.HitRate
	verdict.ObservedObsoleteHitDelta = candidate.ObsoleteHitRate - current.ObsoleteHitRate
	verdict.ObservedMRRDelta = candidate.MeanReciprocalRank - current.MeanReciprocalRank
	verdict.RelevantHitNonInferior = verdict.ObservedRelevantHitDelta >= 0
	verdict.ObsoleteHitNonInferior = verdict.ObservedObsoleteHitDelta <= 0
	verdict.StrictRateImprovement = verdict.ObservedRelevantHitDelta > 0 || verdict.ObservedObsoleteHitDelta < 0
	verdict.MRRWithinMargin = verdict.ObservedMRRDelta >= -margin
	if verdict.RelevantHitNonInferior && verdict.ObsoleteHitNonInferior && verdict.StrictRateImprovement && verdict.MRRWithinMargin {
		verdict.Status = "passed"
	} else {
		verdict.Status = "failed"
	}
	return verdict
}

func sessionForCase(caseID string) string {
	const prefix = "snap-"
	if len(caseID) > len(prefix) && caseID[:len(prefix)] == prefix {
		return caseID[len(prefix):]
	}
	return caseID
}

func stringSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func splitKey(value string) (string, string) {
	for index := range value {
		if value[index] == 0 {
			return value[:index], value[index+1:]
		}
	}
	return value, ""
}
