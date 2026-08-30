package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JuanHuaXu/eventframed/internal/agency"
	"github.com/JuanHuaXu/eventframed/internal/audit"
	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/calibration"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	graphpolicy "github.com/JuanHuaXu/eventframed/internal/graph"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"github.com/JuanHuaXu/eventframed/internal/rankdelta"
	"github.com/JuanHuaXu/eventframed/internal/ranking"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

const (
	maxAgencyIdentifierBytes = 256
	maxRecallK               = 1000
	maxPackK                 = 100
	ResidualModeApply        = "apply"
	ResidualModeShadow       = "shadow"
	ResidualModeDisabled     = "disabled"
)

type Config struct {
	DefaultRecallK             int
	DefaultPackK               int
	DefaultTokenBudget         int
	OverfetchMultiplier        int
	Quantization               string
	BayesianPolicy             bayes.Policy
	BaselineCalibration        calibration.Logit
	PredictiveCalibration      calibration.Logit
	BayesianScoreWeight        float64
	BayesianChangePolicy       bayes.ChangePolicy
	BayesianGroupPolicy        bayes.GroupPolicy
	RankingPolicy              ranking.Policy
	PackingPolicy              packing.Policy
	CandidateRanker            retrieval.CandidateRanker
	CandidateRankerRequired    bool
	CandidateRetriever         retrieval.CandidateRetriever
	CandidateRetrieverRequired bool
	CandidateIndex             retrieval.CandidateIndex
	CandidateCollectionPrefix  string
	RankDeltaStore             rankdelta.Store
	RankDeltaStoreRequired     bool
	RankDeltaTTL               time.Duration
	MaxRankDelta               float64
	ElasticRankDelta           ranking.ElasticDeltaPolicy
	ResidualPolicy             residual.Policy
	ResidualMode               string
	SnapPolicy                 graphpolicy.Policy
	AgencyPolicy               agency.Policy
	AgencySigner               *agency.Signer
	AgencyIssuerToken          string
	AgencyAuthorityToken       string
}

type Service struct {
	store     store.EventStore
	embedder  embed.Embedder
	config    Config
	ranker    retrieval.CandidateRanker
	retriever retrieval.CandidateRetriever
	index     retrieval.CandidateIndex
	rankDelta rankdelta.Store
}

func New(eventStore store.EventStore, embedder embed.Embedder, config Config) (*Service, error) {
	if eventStore == nil || embedder == nil {
		return nil, errors.New("store and embedder are required")
	}
	if config.DefaultRecallK <= 0 || config.DefaultPackK <= 0 || config.DefaultTokenBudget <= 0 {
		return nil, errors.New("default recall, packing, and token budgets must be positive")
	}
	if config.DefaultPackK > config.DefaultRecallK {
		return nil, errors.New("default pack budget cannot exceed recall budget")
	}
	if config.OverfetchMultiplier < 1 {
		config.OverfetchMultiplier = 3
	}
	if config.BayesianPolicy.MaxActive <= 0 {
		// Retrieval already bounds the frontier by recall_k. Update every
		// evidence-ready frontier member by default; explicit BayesianPolicy
		// values retain selective activation for ablations or tighter budgets.
		config.BayesianPolicy = bayes.Policy{VectorWeight: .6, NeighborWeight: .15, NoveltyWeight: .15, IndependenceWeight: .1, Threshold: .65, CriticalThreshold: .45, AuditProbability: .02, MaxActive: 32, AuditSeed: "eventframe-frontier-all-v1", CheapUpdateAll: true}
	}
	if !config.BaselineCalibration.Valid() {
		config.BaselineCalibration = calibration.Identity()
	}
	if !config.PredictiveCalibration.Valid() {
		config.PredictiveCalibration = config.BaselineCalibration
	}
	if config.BayesianScoreWeight == 0 {
		config.BayesianScoreWeight = 0.10
	}
	if config.BayesianScoreWeight < 0 || config.BayesianScoreWeight > 0.25 {
		return nil, errors.New("Bayesian score weight must be in [0,0.25]")
	}
	if !config.BayesianChangePolicy.Valid() {
		config.BayesianChangePolicy = bayes.ChangePolicy{
			Hazard: .05, Threshold: .30, MaxRun: 64,
			FastRate: .25, SlowRate: .025, DriftThreshold: .30, DriftPersistence: 12, MinSamples: 20,
			CUSUMSlack: .10, CUSUMThreshold: 8,
			CooldownSamples: 20,
		}
	}
	if !config.BayesianGroupPolicy.Valid() {
		config.BayesianGroupPolicy = bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 8, MaxMembers: 64, EquivalenceMargin: .15, EquivalenceThreshold: .80, MaxUncertainBorrowing: .10, SharedEvidenceWeight: .5}
	}
	if config.PackingPolicy.MaxPack <= 0 {
		defaults := packing.DefaultPolicy()
		defaults.AdaptiveEnabled = config.PackingPolicy.AdaptiveEnabled
		defaults.DiversityEnabled = config.PackingPolicy.DiversityEnabled
		config.PackingPolicy = defaults
	}
	if config.CandidateRanker == nil {
		config.CandidateRanker = retrieval.PassthroughRanker{}
	}
	if config.CandidateIndex != nil {
		if config.CandidateRetriever == nil {
			return nil, errors.New("candidate index requires a candidate retriever")
		}
		if strings.TrimSpace(config.CandidateCollectionPrefix) == "" {
			return nil, errors.New("candidate index requires a collection prefix")
		}
	}
	if config.RankDeltaStoreRequired && config.RankDeltaStore == nil {
		return nil, errors.New("required rank-delta store is unavailable")
	}
	if config.RankDeltaTTL <= 0 {
		config.RankDeltaTTL = 5 * time.Minute
	}
	if config.MaxRankDelta == 0 {
		config.MaxRankDelta = .25
	}
	if config.MaxRankDelta < 0 || config.MaxRankDelta > 1 {
		return nil, errors.New("maximum rank delta must be in (0,1]")
	}
	if config.ElasticRankDelta.MinScale == 0 && config.ElasticRankDelta.MaxScale == 0 {
		config.ElasticRankDelta = ranking.DefaultElasticDeltaPolicy()
	}
	if !config.ElasticRankDelta.Valid() {
		return nil, errors.New("elastic rank-delta scales must satisfy 0 <= min <= max <= 10")
	}
	if !config.ResidualPolicy.Valid() {
		config.ResidualPolicy = residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .10, MaxAge: 30 * 24 * time.Hour, ImprovementDelta: .001, MinMeanGain: .001, MaxConsecutiveHarm: 1, ShrinkByConfidence: true}
	}
	if config.ResidualMode == "" {
		config.ResidualMode = ResidualModeApply
	}
	if config.ResidualMode != ResidualModeApply && config.ResidualMode != ResidualModeShadow && config.ResidualMode != ResidualModeDisabled {
		return nil, errors.New("residual mode must be apply, shadow, or disabled")
	}
	if !config.SnapPolicy.Valid() {
		config.SnapPolicy = graphpolicy.Policy{MaxNodes: 256, MaxEdges: 512, MaxCandidateFamily: 32, ClosureRadius: 2, MinNetPriorityGain: .01, MaxProperRiskIncrease: .01, MaxUnresolvedBurden: 0, MinSimultaneousCoverage: .95, MinBucketSupport: 30}
	}
	if !config.AgencyPolicy.Valid() {
		config.AgencyPolicy = agency.DefaultPolicy(config.AgencyPolicy.Enabled)
	}
	if config.AgencyPolicy.Enabled && (config.AgencySigner == nil || len(config.AgencyIssuerToken) < 32 || len(config.AgencyAuthorityToken) < 32) {
		return nil, errors.New("enabled agency policy requires a proposal signer, issuer token, and authority token")
	}
	policyDigest, err := bayesianPolicyDigest(config)
	if err != nil {
		return nil, err
	}
	if _, err := eventStore.BindBayesianPolicy(context.Background(), policyDigest); err != nil {
		return nil, fmt.Errorf("bind Bayesian policy: %w", err)
	}
	return &Service{store: eventStore, embedder: embedder, config: config, ranker: config.CandidateRanker, retriever: config.CandidateRetriever, index: config.CandidateIndex, rankDelta: config.RankDeltaStore}, nil
}

func (s *Service) Observe(ctx context.Context, request model.ObserveRequest) (model.ObserveResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.ObserveResponse{}, err
	}
	if request.IdempotencyKey == "" {
		return model.ObserveResponse{}, errors.New("idempotency_key is required")
	}
	if request.IdempotencyKey != request.Event.ID {
		return model.ObserveResponse{}, errors.New("idempotency_key must equal event id in v1alpha1")
	}
	if err := request.Event.Validate(s.embedder.Dimension()); err != nil {
		return model.ObserveResponse{}, err
	}
	vector := request.Event.Embedding
	if len(vector) == 0 {
		var err error
		vector, err = embed.Document(s.embedder, request.Event.EmbeddingText())
		if err != nil {
			return model.ObserveResponse{}, fmt.Errorf("embed event: %w", err)
		}
		request.Event.EmbeddingModel = s.embedder.ModelKey()
	} else if request.Event.EmbeddingModel != s.embedder.ModelKey() {
		return model.ObserveResponse{}, fmt.Errorf("embedding_model %q does not match active model %q", request.Event.EmbeddingModel, s.embedder.ModelKey())
	}
	digest, err := eventDigest(request.Event)
	if err != nil {
		return model.ObserveResponse{}, err
	}
	result, err := s.store.Put(ctx, request.Event, vector, digest)
	if err != nil {
		return model.ObserveResponse{}, err
	}
	if s.index != nil {
		collection := s.candidateCollection(request.Event.TenantID)
		metadata, metadataErr := eventIndexMetadata(request.Event, collection, digest)
		if metadataErr != nil {
			return model.ObserveResponse{}, metadataErr
		}
		candidate := retrieval.Candidate{ID: request.Event.ID, Text: request.Event.Content, Metadata: metadata}
		if indexErr := s.index.EnsureText(ctx, collection, candidate, "eventframe_digest", digest); indexErr != nil {
			return model.ObserveResponse{}, fmt.Errorf("index event through LibraVDB contract: %w", indexErr)
		}
	}
	return model.ObserveResponse{
		ProtocolVersion: model.ProtocolVersion,
		EventID:         request.Event.ID,
		Duplicate:       result.Duplicate,
		Snapshot:        result.Snapshot,
	}, nil
}

func (s *Service) Recall(ctx context.Context, request model.RecallRequest) (model.ContextPacket, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.ContextPacket{}, err
	}
	if strings.TrimSpace(request.TenantID) == "" || strings.TrimSpace(request.SessionID) == "" {
		return model.ContextPacket{}, errors.New("tenant_id and session_id are required")
	}
	if strings.TrimSpace(request.Query) == "" && len(request.Embedding) == 0 {
		return model.ContextPacket{}, errors.New("query or embedding is required")
	}
	if request.AsOf.IsZero() {
		return model.ContextPacket{}, errors.New("as_of is required to enforce availability-time filtering")
	}
	recallK, packK, tokenBudget, err := s.resolveBudgets(request)
	if err != nil {
		return model.ContextPacket{}, err
	}
	vector := request.Embedding
	if len(vector) == 0 {
		vector, err = embed.Query(s.embedder, request.Query)
		if err != nil {
			return model.ContextPacket{}, fmt.Errorf("embed query: %w", err)
		}
	} else if len(vector) != s.embedder.Dimension() {
		return model.ContextPacket{}, fmt.Errorf("query embedding dimension %d does not match %d", len(vector), s.embedder.Dimension())
	} else if request.EmbeddingModel != s.embedder.ModelKey() {
		return model.ContextPacket{}, errors.New("query embedding_model does not match active model")
	}
	searchLimit := recallK * s.config.OverfetchMultiplier
	results, nominationContract, err := s.nominateCandidates(ctx, request, vector, searchLimit)
	if err != nil {
		return model.ContextPacket{}, err
	}
	candidates := make([]model.Candidate, 0, len(results))
	eligible := 0
	for _, result := range results {
		if result.Event.AvailableAt.After(request.AsOf) {
			continue
		}
		eligible++
		candidate := model.Candidate{
			Event:             result.Event,
			Similarity:        result.Similarity,
			RetrievalMetadata: append([]byte(nil), result.RetrievalMetadata...),
			RecencyScore:      candidateRecency(result.Event, request.AsOf),
			EstimatedTokens:   estimateTokens(result.Event.Content),
		}
		candidate.BaselineScore = scoreCandidate(candidate, request.SessionID, request.AsOf)
		candidate.PredictiveScore = ranking.Score(ranking.Features{Baseline: candidate.BaselineScore, Recency: candidate.RecencyScore}, s.config.RankingPolicy)
		candidate.Score = candidate.PredictiveScore
		candidates = append(candidates, candidate)
	}
	predictiveGraph, graphErr := s.store.GetPredictiveGraph(ctx, request.TenantID)
	if graphErr != nil {
		return model.ContextPacket{}, graphErr
	}
	baseScores := make(map[string]float64, len(candidates))
	for _, candidate := range candidates {
		baseScores[candidate.Event.ID] = candidate.BaselineScore
	}
	graphScores := graphpolicy.CandidateCompatibility(predictiveGraph, baseScores)
	for index := range candidates {
		graphScore, hasGraph := graphScores[candidates[index].Event.ID]
		candidates[index].GraphCompatibility = graphScore
		candidates[index].GraphApplied = hasGraph
		candidates[index].PredictiveScore = ranking.Score(ranking.Features{Baseline: candidates[index].BaselineScore, Recency: candidates[index].RecencyScore, Graph: graphScore, HasGraph: hasGraph}, s.config.RankingPolicy)
		candidates[index].Score = candidates[index].PredictiveScore
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Event.AvailableAt.After(candidates[j].Event.AvailableAt)
		}
		return candidates[i].Score > candidates[j].Score
	})
	shadowInputs := make([]bayes.Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		shadowInputs = append(shadowInputs, bayes.Candidate{EventID: candidate.Event.ID, VectorRelevance: clamp((candidate.Similarity+1)/2, 0, 1), NeighborCompatibility: candidate.GraphCompatibility, Novelty: 1 - candidate.Event.MeanFieldConfidence(), SourceIndependence: sourceIndependence(candidate.Event), Priority: candidate.Event.Priority, EvidenceReady: !candidate.Event.AvailableAt.After(request.AsOf)})
	}
	snapshot := s.store.Snapshot(ctx)
	shadow := bayes.Evaluate(shadowInputs, snapshot.EvidenceEpoch, s.config.BayesianPolicy)
	queryDigest, err := recallQueryDigest(request, vector, s.embedder.ModelKey())
	if err != nil {
		return model.ContextPacket{}, err
	}
	selectionCertificate, certificateErr := s.store.GetSelectionCertificate(ctx, request.TenantID)
	if certificateErr != nil && !errors.Is(certificateErr, store.ErrCertificateNotFound) {
		return model.ContextPacket{}, certificateErr
	}
	omittedCertificate, omittedErr := s.store.GetOmittedInfluenceCertificate(ctx, request.TenantID)
	if omittedErr != nil && !errors.Is(omittedErr, store.ErrCertificateNotFound) {
		return model.ContextPacket{}, omittedErr
	}
	now := time.Now().UTC()
	omittedCertified := omittedErr == nil && omittedInfluenceCertificateActive(omittedCertificate, snapshot, now, queryDigest)
	shadow.OmittedInfluenceCertified = omittedCertified
	if certificateErr == nil && selectionCertificateActive(selectionCertificate, snapshot, now) {
		shadow.Mode = "selection-certified-shadow"
		shadow.SelectionSupportCertified = true
		if omittedCertified {
			shadow.Mode = "certified-shadow"
		}
		for index := range shadow.Decisions {
			decision := &shadow.Decisions[index]
			decision.NominationProbabilityLowerBound = selectionCertificate.MinSelectionProbability
			if decision.Activated {
				decision.ActivationProbability = 1
				decision.TotalSelectionProbabilityLowerBound = selectionCertificate.MinSelectionProbability
				if certificate, getErr := s.store.GetAntiPigeonCertificate(ctx, request.TenantID, []string{decision.EventID}); getErr == nil && antiPigeonCertificateActive(certificate, snapshot, now) {
					decision.PosteriorKey = "ap:" + certificate.ID
				} else if getErr != nil && !errors.Is(getErr, store.ErrCertificateNotFound) {
					return model.ContextPacket{}, getErr
				}
			}
		}
	}
	if shadow.SelectionSupportCertified && shadow.OmittedInfluenceCertified {
		decisions := make(map[string]int, len(shadow.Decisions))
		for index := range shadow.Decisions {
			decisions[shadow.Decisions[index].EventID] = index
		}
		applied := false
		for index := range candidates {
			decision := &shadow.Decisions[decisions[candidates[index].Event.ID]]
			if !decision.Activated || decision.TotalSelectionProbabilityLowerBound <= 0 {
				continue
			}
			posterior, getErr := s.store.GetBayesianPosterior(ctx, request.TenantID, decision.PosteriorKey)
			if errors.Is(getErr, store.ErrPosteriorNotFound) && s.config.RankingPolicy.HierarchicalEnabled {
				posterior = model.BayesianPosterior{TenantID: request.TenantID, PosteriorKey: decision.PosteriorKey, Alpha: 1, Beta: 1, EvidenceEpoch: snapshot.EvidenceEpoch, Certified: true}
				getErr = nil
			} else if errors.Is(getErr, store.ErrPosteriorNotFound) {
				continue
			}
			if getErr != nil {
				return model.ContextPacket{}, getErr
			}
			if !posterior.Certified || posterior.EvidenceEpoch != snapshot.EvidenceEpoch || posterior.UpdatedAt.After(request.AsOf) {
				continue
			}
			probability := clamp(posterior.Mean(), 0, 1)
			if s.config.RankingPolicy.HierarchicalEnabled {
				decision.ParentPosteriorKey = parentPosteriorKey()
				parent, parentErr := s.store.GetBayesianPosterior(ctx, request.TenantID, decision.ParentPosteriorKey)
				if parentErr == nil && parent.Certified && parent.EvidenceEpoch == snapshot.EvidenceEpoch && !parent.UpdatedAt.After(request.AsOf) {
					probability = ranking.HierarchicalMean(posterior, parent, s.config.RankingPolicy)
				} else if parentErr != nil && !errors.Is(parentErr, store.ErrPosteriorNotFound) {
					return model.ContextPacket{}, parentErr
				}
			}
			candidates[index].BayesianProbability = probability
			candidates[index].BayesianApplied = true
			if s.config.RankingPolicy.ContextualEnabled || s.config.RankingPolicy.HierarchicalEnabled {
				_, hasGraph := graphScores[candidates[index].Event.ID]
				features := ranking.Features{Baseline: candidates[index].BaselineScore, Posterior: probability, Recency: candidates[index].RecencyScore, Graph: candidates[index].GraphCompatibility, HasPosterior: true, HasGraph: hasGraph}
				candidates[index].PredictiveScore = ranking.Score(features, s.config.RankingPolicy)
				candidates[index].Score = candidates[index].PredictiveScore
			} else {
				candidates[index].PredictiveScore = (1-s.config.BayesianScoreWeight)*candidates[index].BaselineScore + s.config.BayesianScoreWeight*probability
				candidates[index].Score = candidates[index].PredictiveScore
			}
			applied = true
		}
		if applied {
			shadow.Mode = "production"
			sort.SliceStable(candidates, func(i, j int) bool {
				if candidates[i].Score == candidates[j].Score {
					return candidates[i].Event.AvailableAt.After(candidates[j].Event.AvailableAt)
				}
				return candidates[i].Score > candidates[j].Score
			})
		}
	}
	decisionIndexes := make(map[string]int, len(shadow.Decisions))
	for index := range shadow.Decisions {
		decisionIndexes[shadow.Decisions[index].EventID] = index
	}
	for index := range candidates {
		candidate := &candidates[index]
		decisionIndex, ok := decisionIndexes[candidate.Event.ID]
		if !ok {
			return model.ContextPacket{}, errors.New("Bayesian frontier omitted a recalled candidate")
		}
		decision := &shadow.Decisions[decisionIndex]
		baseProbability := s.config.BaselineCalibration.Apply(candidate.BaselineScore)
		preResidualProbability := baseProbability
		if candidate.BayesianApplied {
			preResidualProbability = s.config.PredictiveCalibration.Apply(candidate.PredictiveScore)
		}
		preResidual := bernoulliLaw(preResidualProbability)
		forecast := model.ForecastBundle{
			ModelKind: "plugin-bernoulli-retrieval-usefulness", RankScore: candidate.Score, HorizonKey: model.RetrievalUsefulnessHorizon,
			BaseLaw: bernoulliLaw(baseProbability), PreResidualLaw: preResidual, CorrectedLaw: preResidual,
			Template:     model.ForecastTemplate{EventID: candidate.Event.ID, PredictedUseful: preResidual.Useful >= .5, Confidence: math.Max(preResidual.Useful, preResidual.NotUseful)},
			PosteriorKey: decision.PosteriorKey, PosteriorVersion: snapshot.PosteriorVersion,
		}
		if candidate.BayesianApplied {
			belief := bernoulliLaw(candidate.BayesianProbability)
			forecast.BeliefLaw = &belief
		}
		actionKey := residualActionKey(queryDigest, candidate.Event.ID, forecast.HorizonKey)
		generalKey := residualGeneralKey(decision.PosteriorKey, forecast.HorizonKey)
		var cached model.ResidualCandidates
		var getErr error
		if s.config.ResidualMode != ResidualModeDisabled {
			cached, getErr = s.store.GetResidualCandidates(ctx, request.TenantID, actionKey, generalKey)
			if getErr != nil {
				return model.ContextPacket{}, getErr
			}
		}
		var selected *model.ResidualRecord
		if cached.Exact != nil && residual.Eligible(*cached.Exact, preResidual.Useful, snapshot, request.AsOf.UTC(), s.config.ResidualPolicy) {
			selected = cached.Exact
		} else if cached.General != nil && residual.Eligible(*cached.General, preResidual.Useful, snapshot, request.AsOf.UTC(), s.config.ResidualPolicy) {
			selected = cached.General
		}
		if selected != nil {
			forecast.ResidualShadowEligible = true
			forecast.ResidualRecordID = selected.ID
			shadow.ResidualShadowEligible++
			if s.config.ResidualMode == ResidualModeApply {
				forecast.CorrectedLaw = residual.Apply(preResidual, *selected, s.config.ResidualPolicy)
				forecast.ResidualApplied = true
				forecast.Template = model.ForecastTemplate{EventID: candidate.Event.ID, PredictedUseful: forecast.CorrectedLaw.Useful >= .5, Confidence: math.Max(forecast.CorrectedLaw.Useful, forecast.CorrectedLaw.NotUseful)}
				candidate.Score = clamp(candidate.Score+forecast.CorrectedLaw.Useful-preResidual.Useful, 0, 1)
				forecast.RankScore = candidate.Score
				shadow.ResidualApplied++
			}
		}
		candidate.Forecast = forecast
		decision.Forecast = forecast
	}
	if shadow.ResidualApplied > 0 {
		if shadow.Mode != "production" {
			shadow.Mode = "residual-production"
		}
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].Score == candidates[j].Score {
				return candidates[i].Event.AvailableAt.After(candidates[j].Event.AvailableAt)
			}
			return candidates[i].Score > candidates[j].Score
		})
	}
	deltas, deltaErr := s.materializeRankDeltas(ctx, request, queryDigest, snapshot, candidates)
	if deltaErr != nil {
		if s.config.RankDeltaStoreRequired {
			return model.ContextPacket{}, deltaErr
		}
		deltas = liveRankDeltas(request, queryDigest, snapshot, candidates, s.config)
	}
	rankedCandidates, rankErr := s.rankCandidates(ctx, request, candidates, recallK)
	if rankErr != nil {
		if s.config.CandidateRankerRequired {
			return model.ContextPacket{}, rankErr
		}
		for index := range candidates {
			candidates[index].RetrievalScore = contractBaseScore(candidates[index])
			candidates[index].Score = candidates[index].RetrievalScore
		}
	} else {
		candidates = rankedCandidates
	}
	packetAnswerCertainty := s.applyRankDeltas(candidates, deltas, queryDigest, packK)
	for index := range candidates {
		candidates[index].RetrievalContract = s.ranker.ContractName()
		candidates[index].Forecast.RankScore = candidates[index].Score
		if decisionIndex, ok := decisionIndexes[candidates[index].Event.ID]; ok {
			shadow.Decisions[decisionIndex].Forecast = candidates[index].Forecast
		}
	}
	if len(candidates) > recallK {
		candidates = candidates[:recallK]
	}
	journal := model.BayesianJournalEntry{
		TenantID: request.TenantID, SessionID: request.SessionID, AsOf: request.AsOf.UTC(),
		QueryDigest: queryDigest, Snapshot: snapshot, Report: shadow,
	}
	journal.ID, err = bayesianJournalID(journal)
	if err != nil {
		return model.ContextPacket{}, err
	}
	journal.Report.JournalID = journal.ID
	journal.Report.JournalDurable = true
	if err := s.store.PutBayesianJournal(ctx, journal); err != nil {
		return model.ContextPacket{}, fmt.Errorf("persist Bayesian frontier journal: %w", err)
	}
	shadow = journal.Report
	posteriorKeys := make(map[string]string, len(shadow.Decisions))
	for _, decision := range shadow.Decisions {
		posteriorKeys[decision.EventID] = decision.PosteriorKey
	}
	packingResult := packing.Select(candidates, posteriorKeys, packK, recallK, tokenBudget, s.config.PackingPolicy)
	return model.ContextPacket{
		ProtocolVersion:       model.ProtocolVersion,
		Candidates:            packingResult.Candidates,
		Recalled:              len(results),
		Eligible:              eligible,
		Packed:                len(packingResult.Candidates),
		UsedTokens:            packingResult.UsedTokens,
		AdaptiveExpanded:      packingResult.Expanded,
		PacketConfidence:      packetAnswerCertainty,
		PacketAnswerCertainty: packetAnswerCertainty,
		RetrievalContract:     s.ranker.ContractName(),
		NominationContract:    nominationContract,
		Snapshot:              snapshot,
		BayesianShadow:        shadow,
	}, nil
}

func (s *Service) nominateCandidates(ctx context.Context, request model.RecallRequest, vector []float32, limit int) ([]store.SearchResult, string, error) {
	if s.retriever == nil {
		results, err := s.store.Search(ctx, request.TenantID, vector, request.AsOf, limit)
		return results, "embedded-vector-search", err
	}
	collections := append([]string(nil), request.RetrievalCollections...)
	exclusions := request.RetrievalExcludeByCollection
	if s.config.CandidateCollectionPrefix != "" {
		reserved := s.candidateCollection(request.TenantID)
		if len(collections) == 0 {
			collections = []string{reserved}
		} else if len(collections) != 1 || collections[0] != reserved {
			return nil, "", errors.New("retrieval collections cannot override the tenant-isolated EventFrame collection")
		}
		for collection := range exclusions {
			if collection != reserved {
				return nil, "", errors.New("retrieval exclusions cannot reference another EventFrame collection")
			}
		}
	}
	if len(collections) == 0 {
		if s.config.CandidateRetrieverRequired {
			return nil, "", errors.New("retrieval_collections are required for contract-native nomination")
		}
		results, err := s.store.Search(ctx, request.TenantID, vector, request.AsOf, limit)
		return results, "embedded-vector-search", err
	}
	if len(collections) > 16 {
		return nil, "", errors.New("retrieval collection count exceeds safety cap")
	}
	excluded := 0
	for collection, ids := range exclusions {
		if strings.TrimSpace(collection) == "" || len(ids) > maxRecallK*4 {
			return nil, "", errors.New("retrieval exclusions exceed safety caps")
		}
		excluded += len(ids)
	}
	if excluded > maxRecallK*8 {
		return nil, "", errors.New("total retrieval exclusions exceed safety cap")
	}
	contractResults, err := s.retriever.SearchTextCollections(ctx, retrieval.SearchRequest{
		Collections: collections, QueryText: request.Query, K: limit,
		ExcludeByCollection: exclusions,
	})
	if err != nil {
		return nil, "", err
	}
	ids := make([]string, 0, len(contractResults))
	scores := make(map[string]float64, len(contractResults))
	metadataByID := make(map[string][]byte, len(contractResults))
	allowedCollections := make(map[string]struct{}, len(collections))
	for _, collection := range collections {
		allowedCollections[collection] = struct{}{}
	}
	for _, candidate := range contractResults {
		if _, duplicate := scores[candidate.ID]; duplicate {
			return nil, "", fmt.Errorf("nomination contract returned duplicate candidate %q", candidate.ID)
		}
		var metadata struct {
			Collection string `json:"collection"`
		}
		if err := json.Unmarshal(candidate.Metadata, &metadata); err != nil || metadata.Collection == "" {
			return nil, "", fmt.Errorf("nomination contract returned candidate %q without valid collection metadata", candidate.ID)
		}
		if _, ok := allowedCollections[metadata.Collection]; !ok {
			return nil, "", fmt.Errorf("nomination contract returned candidate %q from undeclared collection %q", candidate.ID, metadata.Collection)
		}
		ids = append(ids, candidate.ID)
		scores[candidate.ID] = candidate.Score
		metadataByID[candidate.ID] = append([]byte(nil), candidate.Metadata...)
	}
	events, err := s.store.GetEvents(ctx, request.TenantID, ids, request.AsOf)
	if err != nil {
		if !errors.Is(err, store.ErrEventNotFound) || s.index == nil {
			return nil, "", fmt.Errorf("resolve nominated events: %w", err)
		}
		events, err = s.resolveAndRepairNominations(ctx, request, ids, collections[0])
		if err != nil {
			return nil, "", err
		}
	}
	results := make([]store.SearchResult, 0, len(events))
	for _, event := range events {
		results = append(results, store.SearchResult{Event: event, Similarity: scores[event.ID], RetrievalMetadata: metadataByID[event.ID]})
	}
	return results, s.retriever.RetrievalContractName(), nil
}

func (s *Service) resolveAndRepairNominations(ctx context.Context, request model.RecallRequest, ids []string, collection string) ([]model.Event, error) {
	events := make([]model.Event, 0, len(ids))
	stale := make([]string, 0)
	for _, id := range ids {
		resolved, err := s.store.GetEvents(ctx, request.TenantID, []string{id}, request.AsOf)
		if err == nil {
			events = append(events, resolved[0])
			continue
		}
		if !errors.Is(err, store.ErrEventNotFound) {
			return nil, fmt.Errorf("resolve nominated event %q: %w", id, err)
		}
		future, futureErr := s.store.GetEvents(ctx, request.TenantID, []string{id}, time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC))
		if futureErr == nil && len(future) == 1 {
			continue
		}
		if futureErr != nil && !errors.Is(futureErr, store.ErrEventNotFound) {
			return nil, fmt.Errorf("check stale nominated event %q: %w", id, futureErr)
		}
		stale = append(stale, id)
	}
	if len(stale) > 0 {
		if err := s.index.DeleteTextBatch(ctx, collection, stale); err != nil {
			return nil, fmt.Errorf("repair stale LibraVDB nominations: %w", err)
		}
	}
	return events, nil
}

func sourceIndependence(event model.Event) float64 {
	if len(event.Provenance.RetrievedIDs) > 0 {
		return 0.25
	}
	if event.Who.Source == model.SourceObserved || event.What.Source == model.SourceObserved {
		return 1
	}
	return 0.5
}

func (s *Service) rankCandidates(ctx context.Context, request model.RecallRequest, candidates []model.Candidate, recallK int) ([]model.Candidate, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	contractCandidates := make([]retrieval.Candidate, 0, len(candidates))
	byID := make(map[string]model.Candidate, len(candidates))
	for _, candidate := range candidates {
		metadata, err := retrievalMetadata(candidate, request)
		if err != nil {
			return nil, err
		}
		contractCandidates = append(contractCandidates, retrieval.Candidate{
			ID: candidate.Event.ID, Text: candidate.Event.Content, Score: contractBaseScore(candidate), Metadata: metadata,
		})
		byID[candidate.Event.ID] = candidate
	}
	k2 := min(recallK, len(contractCandidates))
	ranked, err := s.ranker.RankCandidates(ctx, retrieval.RankRequest{
		Candidates: contractCandidates, QueryText: request.Query, SessionID: request.SessionID,
		UserID: request.TenantID, K1: len(contractCandidates), K2: k2,
	})
	if err != nil {
		return nil, err
	}
	if len(ranked) > k2 {
		return nil, fmt.Errorf("retrieval contract returned %d candidates; maximum is %d", len(ranked), k2)
	}
	result := make([]model.Candidate, 0, len(candidates))
	seen := make(map[string]struct{}, len(ranked))
	for _, rankedCandidate := range ranked {
		candidate, ok := byID[rankedCandidate.ID]
		if !ok {
			return nil, fmt.Errorf("retrieval contract returned unknown candidate %q", rankedCandidate.ID)
		}
		if _, duplicate := seen[rankedCandidate.ID]; duplicate {
			return nil, fmt.Errorf("retrieval contract returned duplicate candidate %q", rankedCandidate.ID)
		}
		seen[rankedCandidate.ID] = struct{}{}
		candidate.RetrievalScore = clamp(rankedCandidate.Score, 0, 1)
		candidate.Score = candidate.RetrievalScore
		result = append(result, candidate)
	}
	// RankCandidates is a second-pass selector and may omit its low-scored tail.
	// Preserve that bounded nominated frontier so EventFrame deltas can still
	// recover a candidate that LibraVDB did not return from the second pass.
	for _, candidate := range candidates {
		if _, ok := seen[candidate.Event.ID]; ok {
			continue
		}
		candidate.RetrievalScore = contractBaseScore(candidate)
		candidate.Score = candidate.RetrievalScore
		result = append(result, candidate)
	}
	return result, nil
}

func contractBaseScore(candidate model.Candidate) float64 {
	if len(candidate.RetrievalMetadata) > 0 {
		return clamp(candidate.Similarity, 0, 1)
	}
	return clamp(candidate.BaselineScore, 0, 1)
}

func (s *Service) materializeRankDeltas(ctx context.Context, request model.RecallRequest, queryDigest string, snapshot model.Snapshot, candidates []model.Candidate) (map[string]rankdelta.Record, error) {
	records := liveRankDeltaRecords(request, queryDigest, snapshot, candidates, s.config)
	if s.rankDelta == nil {
		return indexRankDeltas(records), nil
	}
	keys := make([]string, 0, len(records))
	for _, record := range records {
		keys = append(keys, record.Key)
	}
	loaded, err := s.rankDelta.GetBatch(ctx, request.TenantID, keys, snapshot, request.AsOf.UTC())
	if err != nil {
		return nil, fmt.Errorf("load EventFrame rank deltas: %w", err)
	}
	pending := make([]rankdelta.Record, 0, len(records))
	for _, record := range records {
		if cached, ok := loaded[record.Key]; ok && sameRankDelta(cached, record) {
			continue
		}
		pending = append(pending, record)
		loaded[record.Key] = record
	}
	if err := s.rankDelta.PutBatch(ctx, pending); err != nil {
		return nil, fmt.Errorf("persist EventFrame rank deltas: %w", err)
	}
	return loaded, nil
}

func liveRankDeltas(request model.RecallRequest, queryDigest string, snapshot model.Snapshot, candidates []model.Candidate, config Config) map[string]rankdelta.Record {
	return indexRankDeltas(liveRankDeltaRecords(request, queryDigest, snapshot, candidates, config))
}

func liveRankDeltaRecords(request model.RecallRequest, queryDigest string, snapshot model.Snapshot, candidates []model.Candidate, config Config) []rankdelta.Record {
	records := make([]rankdelta.Record, 0, len(candidates))
	expiresAt := request.AsOf.UTC().Add(config.RankDeltaTTL)
	for _, candidate := range candidates {
		// Delta is the EventFrame correction only. LibraVDB's returned score is
		// the base to which this correction is applied after the contract call.
		delta := clamp(candidate.Score-candidate.BaselineScore, -config.MaxRankDelta, config.MaxRankDelta)
		if math.Abs(delta) <= 1e-12 {
			continue
		}
		records = append(records, rankdelta.Record{
			TenantID: request.TenantID, Key: rankDeltaKey(queryDigest, candidate.Event.ID), EventID: candidate.Event.ID,
			Delta: delta, Reliability: correctionReliability(candidate),
			PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch, GraphVersion: snapshot.GraphVersion,
			PosteriorVersion: snapshot.PosteriorVersion, ResidualVersion: snapshot.ResidualVersion, AbstractionVersion: snapshot.AbstractionVersion,
			UpdatedAt: request.AsOf.UTC(), ExpiresAt: expiresAt,
		})
	}
	return records
}

func sameRankDelta(left, right rankdelta.Record) bool {
	return left.TenantID == right.TenantID && left.Key == right.Key && left.EventID == right.EventID &&
		left.Delta == right.Delta && left.Reliability == right.Reliability &&
		left.PolicyVersion == right.PolicyVersion && left.EvidenceEpoch == right.EvidenceEpoch &&
		left.GraphVersion == right.GraphVersion && left.PosteriorVersion == right.PosteriorVersion &&
		left.ResidualVersion == right.ResidualVersion && left.AbstractionVersion == right.AbstractionVersion
}

func indexRankDeltas(records []rankdelta.Record) map[string]rankdelta.Record {
	indexed := make(map[string]rankdelta.Record, len(records))
	for _, record := range records {
		indexed[record.Key] = record
	}
	return indexed
}

func (s *Service) applyRankDeltas(candidates []model.Candidate, deltas map[string]rankdelta.Record, queryDigest string, packK int) float64 {
	answerCertainty := rankBoundaryCertainty(candidates, packK)
	for index := range candidates {
		record, ok := deltas[rankDeltaKey(queryDigest, candidates[index].Event.ID)]
		if !ok {
			continue
		}
		scale := s.config.ElasticRankDelta.Scale(answerCertainty, record.Reliability)
		candidates[index].RankDeltaScale = scale
		candidates[index].RankDeltaAnswerCertainty = answerCertainty
		candidates[index].RankDeltaCorrectionReliability = record.Reliability
		candidates[index].RankDeltaConfidence = answerCertainty
		candidates[index].RankDeltaBasis = "rank-boundary+correction-reliability"
		candidates[index].RankDelta = clamp(record.Delta*scale, -s.config.MaxRankDelta, s.config.MaxRankDelta)
		candidates[index].Score = clamp(candidates[index].RetrievalScore+candidates[index].RankDelta, 0, 1)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})
	return answerCertainty
}

func rankBoundaryCertainty(candidates []model.Candidate, packK int) float64 {
	if len(candidates) == 0 || packK <= 0 {
		return 0
	}
	if packK >= len(candidates) {
		return 1
	}
	inside := candidates[packK-1].RetrievalScore
	outside := candidates[packK].RetrievalScore
	scale := math.Max(math.Max(math.Abs(inside), math.Abs(outside)), 1e-12)
	return clamp((inside-outside)/scale, 0, 1)
}

func correctionReliability(candidate model.Candidate) float64 {
	if candidate.BayesianApplied || candidate.Forecast.ResidualApplied || candidate.GraphApplied {
		return 1
	}
	return 0
}

func rankDeltaKey(queryDigest, eventID string) string {
	digest := sha256.Sum256([]byte("eventframe-rank-delta-v2\x00" + queryDigest + "\x00" + eventID))
	return hex.EncodeToString(digest[:])
}

func retrievalMetadata(candidate model.Candidate, request model.RecallRequest) ([]byte, error) {
	attributes := candidate.Event.Attributes
	collection := "user:" + request.TenantID
	if candidate.Event.SessionID == request.SessionID {
		collection = "session:" + request.SessionID
	}
	metadata := make(map[string]any)
	if len(candidate.RetrievalMetadata) > 0 {
		if err := json.Unmarshal(candidate.RetrievalMetadata, &metadata); err != nil {
			return nil, fmt.Errorf("decode LibraVDB nomination metadata: %w", err)
		}
		if storedCollection, _ := metadata["collection"].(string); storedCollection == "" {
			return nil, errors.New("LibraVDB nomination metadata has no collection")
		}
	} else {
		metadata["collection"] = collection
	}
	defaults := map[string]any{
		"ts":         candidate.Event.AvailableAt.UnixMilli(),
		"session_id": candidate.Event.SessionID, "user_id": request.TenantID,
		"node_kind": candidate.Event.Kind, "token_estimate": candidate.EstimatedTokens,
		"priority": candidate.Event.Priority, "role": attributes["role"],
		"authored": false, "access_count": 0,
		"authority": candidate.Event.Priority, "salience": candidate.Event.Priority,
		"compaction_confidence": 1.0,
	}
	for key, value := range defaults {
		if _, exists := metadata[key]; !exists {
			metadata[key] = value
		}
	}
	for _, key := range []string{"authored", "authority", "salience", "access_count", "compaction_confidence", "summary_confidence", "tier"} {
		if value, ok := attributes[key]; ok {
			metadata[key] = parseRetrievalAttribute(value)
		}
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("encode LibraVDB retrieval metadata: %w", err)
	}
	return encoded, nil
}

func eventIndexMetadata(event model.Event, collection, digest string) ([]byte, error) {
	metadata := map[string]any{
		"collection": collection, "ts": event.AvailableAt.UnixMilli(),
		"session_id": event.SessionID, "user_id": event.TenantID,
		"node_kind": event.Kind, "priority": event.Priority,
		"role": event.Attributes["role"], "authored": false, "access_count": 0,
		"authority": event.Priority, "salience": event.Priority,
		"compaction_confidence": 1.0, "eventframe_digest": digest,
		"eventframe_available_at": event.AvailableAt.UTC().Format(time.RFC3339Nano),
	}
	for _, key := range []string{"authored", "authority", "salience", "access_count", "compaction_confidence", "summary_confidence", "tier"} {
		if value, ok := event.Attributes[key]; ok {
			metadata[key] = parseRetrievalAttribute(value)
		}
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("encode LibraVDB index metadata: %w", err)
	}
	return encoded, nil
}

func (s *Service) candidateCollection(tenantID string) string {
	digest := sha256.Sum256([]byte(tenantID))
	return s.config.CandidateCollectionPrefix + hex.EncodeToString(digest[:12])
}

func parseRetrievalAttribute(value string) any {
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}
	var number float64
	if _, err := fmt.Sscan(value, &number); err == nil {
		return number
	}
	return value
}

func (s *Service) Health(ctx context.Context) (model.HealthResponse, error) {
	stats, err := s.store.Stats(ctx)
	if err != nil {
		return model.HealthResponse{}, err
	}
	return model.HealthResponse{
		ProtocolVersion:        model.ProtocolVersion,
		Status:                 "ok",
		Store:                  stats.Backend,
		Dimension:              s.embedder.Dimension(),
		Quantization:           s.config.Quantization,
		NominationContract:     nominationContractName(s.retriever),
		RetrievalContract:      s.ranker.ContractName(),
		ExternalCandidateIndex: s.index != nil,
		Snapshot:               s.store.Snapshot(ctx),
	}, nil
}

func (s *Service) Delete(ctx context.Context, request model.DeleteRequest) (model.DeleteResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.DeleteResponse{}, err
	}
	if strings.TrimSpace(request.TenantID) == "" || strings.TrimSpace(request.EventID) == "" {
		return model.DeleteResponse{}, errors.New("tenant_id and event_id are required")
	}
	result, err := s.store.Delete(ctx, request.TenantID, request.EventID)
	if err != nil {
		return model.DeleteResponse{}, err
	}
	if s.index != nil {
		if indexErr := s.index.DeleteText(ctx, s.candidateCollection(request.TenantID), request.EventID); indexErr != nil {
			return model.DeleteResponse{}, fmt.Errorf("delete event through LibraVDB contract: %w", indexErr)
		}
	}
	return model.DeleteResponse{ProtocolVersion: model.ProtocolVersion, EventID: request.EventID, Deleted: result.Deleted, Snapshot: result.Snapshot}, nil
}

func (s *Service) Retain(ctx context.Context, request model.RetentionRequest) (model.RetentionResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.RetentionResponse{}, err
	}
	if request.TenantID == "" || request.Before.IsZero() {
		return model.RetentionResponse{}, errors.New("tenant_id and before are required")
	}
	if request.Limit == 0 {
		request.Limit = 1000
	}
	if request.Limit < 1 || request.Limit > 10000 {
		return model.RetentionResponse{}, errors.New("retention limit must be in [1,10000]")
	}
	result, err := s.store.DeleteBefore(ctx, request.TenantID, request.Before, request.Limit)
	if err != nil {
		return model.RetentionResponse{}, err
	}
	if s.index != nil && len(result.DeletedIDs) > 0 {
		if indexErr := s.index.DeleteTextBatch(ctx, s.candidateCollection(request.TenantID), result.DeletedIDs); indexErr != nil {
			return model.RetentionResponse{}, fmt.Errorf("retain events through LibraVDB contract: %w", indexErr)
		}
	}
	return model.RetentionResponse{ProtocolVersion: model.ProtocolVersion, DeletedIDs: result.DeletedIDs, Snapshot: result.Snapshot}, nil
}

func (s *Service) Backup(ctx context.Context, request model.BackupRequest) (model.MaintenanceResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.MaintenanceResponse{}, err
	}
	if !filepath.IsAbs(request.Destination) {
		return model.MaintenanceResponse{}, errors.New("backup destination must be an absolute path")
	}
	if err := s.store.Backup(ctx, request.Destination); err != nil {
		return model.MaintenanceResponse{}, err
	}
	return model.MaintenanceResponse{ProtocolVersion: model.ProtocolVersion, Operation: "backup", Snapshot: s.store.Snapshot(ctx)}, nil
}

func (s *Service) Compact(ctx context.Context) (model.MaintenanceResponse, error) {
	if err := s.store.Compact(ctx); err != nil {
		return model.MaintenanceResponse{}, err
	}
	return model.MaintenanceResponse{ProtocolVersion: model.ProtocolVersion, Operation: "compact", Snapshot: s.store.Snapshot(ctx)}, nil
}

func (s *Service) PublishSelectionCertificate(ctx context.Context, request model.PublishSelectionCertificateRequest) (model.CertificateResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.CertificateResponse{}, err
	}
	snapshot := s.store.Snapshot(ctx)
	if err := validateSelectionCertificate(request.Certificate, snapshot, time.Now().UTC()); err != nil {
		return model.CertificateResponse{}, err
	}
	next, err := s.store.PublishSelectionCertificate(ctx, request.Certificate)
	if err != nil {
		return model.CertificateResponse{}, err
	}
	return model.CertificateResponse{ProtocolVersion: model.ProtocolVersion, CertificateID: request.Certificate.ID, Snapshot: next}, nil
}

func (s *Service) PublishAntiPigeonCertificate(ctx context.Context, request model.PublishAntiPigeonCertificateRequest) (model.CertificateResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.CertificateResponse{}, err
	}
	snapshot := s.store.Snapshot(ctx)
	if err := validateAntiPigeonCertificate(request.Certificate, snapshot, time.Now().UTC()); err != nil {
		return model.CertificateResponse{}, err
	}
	next, err := s.store.PublishAntiPigeonCertificate(ctx, request.Certificate)
	if err != nil {
		return model.CertificateResponse{}, err
	}
	return model.CertificateResponse{ProtocolVersion: model.ProtocolVersion, CertificateID: request.Certificate.ID, Snapshot: next}, nil
}

func (s *Service) PublishOmittedInfluenceCertificate(ctx context.Context, request model.PublishOmittedInfluenceCertificateRequest) (model.CertificateResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.CertificateResponse{}, err
	}
	snapshot := s.store.Snapshot(ctx)
	if err := validateOmittedInfluenceCertificate(request.Certificate, snapshot, time.Now().UTC()); err != nil {
		return model.CertificateResponse{}, err
	}
	if math.Abs(request.Certificate.AuditProbability-s.config.BayesianPolicy.AuditProbability) > 1e-9 {
		return model.CertificateResponse{}, errors.New("omitted-influence certificate audit_probability does not match the active policy")
	}
	next, err := s.store.PublishOmittedInfluenceCertificate(ctx, request.Certificate)
	if err != nil {
		return model.CertificateResponse{}, err
	}
	return model.CertificateResponse{ProtocolVersion: model.ProtocolVersion, CertificateID: request.Certificate.ID, Snapshot: next}, nil
}

func (s *Service) EstimateOmittedInfluence(ctx context.Context, request model.EstimateOmittedInfluenceRequest) (model.CertificateResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.CertificateResponse{}, err
	}
	if strings.TrimSpace(request.ID) == "" || strings.TrimSpace(request.TenantID) == "" || strings.TrimSpace(request.JournalID) == "" || len(request.PopulationEventIDs) == 0 || len(request.PopulationEventIDs) > 10_000 {
		return model.CertificateResponse{}, errors.New("audit id, tenant, journal, and a population of at most 10000 events are required")
	}
	journal, err := s.store.GetBayesianJournal(ctx, request.TenantID, request.JournalID)
	if err != nil {
		return model.CertificateResponse{}, err
	}
	nominated := make(map[string]struct{}, len(journal.Report.Decisions))
	for _, decision := range journal.Report.Decisions {
		nominated[decision.EventID] = struct{}{}
	}
	for _, eventID := range request.PopulationEventIDs {
		if _, present := nominated[eventID]; present {
			return model.CertificateResponse{}, errors.New("runtime omitted-influence population contains a nominated event")
		}
	}
	expanded := make(map[string]model.BernoulliLaw, len(request.Observations))
	for _, observation := range request.Observations {
		if _, duplicate := expanded[observation.EventID]; duplicate {
			return model.CertificateResponse{}, errors.New("audit observations must be unique")
		}
		expanded[observation.EventID] = observation.ExpandedLaw
	}
	snapshot := s.store.Snapshot(ctx)
	if journal.Snapshot.PolicyVersion != snapshot.PolicyVersion || journal.Snapshot.EvidenceEpoch != snapshot.EvidenceEpoch {
		return model.CertificateResponse{}, errors.New("runtime audit journal is stale")
	}
	estimate, err := audit.EstimateInfluence(request.PopulationEventIDs, request.LocalLaw, expanded, s.config.BayesianPolicy.AuditSeed, snapshot.EvidenceEpoch, request.AuditSequence, s.config.BayesianPolicy.AuditProbability, request.ConfidenceDelta)
	if err != nil {
		return model.CertificateResponse{}, err
	}
	populationPayload, err := json.Marshal(request.PopulationEventIDs)
	if err != nil {
		return model.CertificateResponse{}, err
	}
	populationHash := sha256.Sum256(populationPayload)
	certificate := model.OmittedInfluenceCertificate{
		ID: request.ID, TenantID: request.TenantID, PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch,
		DivergenceUCB: estimate.UpperBound, DivergenceLimit: request.DivergenceLimit,
		AuditProbability: s.config.BayesianPolicy.AuditProbability, SimultaneousCoverage: 1 - request.ConfidenceDelta,
		Procedure: "runtime-horvitz-thompson-js-anytime-v1", Issuer: "eventframed", RuntimeEstimated: true,
		QueryDigest: journal.QueryDigest, PopulationDigest: hex.EncodeToString(populationHash[:]), ValidUntil: request.ValidUntil,
	}
	if err := validateOmittedInfluenceCertificate(certificate, snapshot, time.Now().UTC()); err != nil {
		return model.CertificateResponse{}, err
	}
	next, err := s.store.PublishOmittedInfluenceCertificate(ctx, certificate)
	if err != nil {
		return model.CertificateResponse{}, err
	}
	return model.CertificateResponse{ProtocolVersion: model.ProtocolVersion, CertificateID: certificate.ID, Snapshot: next}, nil
}

func (s *Service) ObserveBayesianOutcome(ctx context.Context, request model.BayesianOutcomeRequest) (model.BayesianOutcomeResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.BayesianOutcomeResponse{}, err
	}
	if strings.TrimSpace(request.IdempotencyKey) == "" || strings.TrimSpace(request.TenantID) == "" || strings.TrimSpace(request.JournalID) == "" || strings.TrimSpace(request.EventID) == "" {
		return model.BayesianOutcomeResponse{}, errors.New("idempotency_key, tenant_id, journal_id, and event_id are required")
	}
	resolvedUseful, evidence := request.Signals.Resolve(request.Useful)
	if !evidence {
		return model.BayesianOutcomeResponse{}, errors.New("packed-only exposure is not usefulness evidence")
	}
	request.Useful = resolvedUseful
	now := time.Now().UTC()
	if request.ObservedAt.IsZero() || request.AvailableAt.IsZero() || request.AvailableAt.Before(request.ObservedAt) || request.AvailableAt.After(now) {
		return model.BayesianOutcomeResponse{}, errors.New("outcome times must satisfy observed_at <= available_at <= current time")
	}
	journal, err := s.store.GetBayesianJournal(ctx, request.TenantID, request.JournalID)
	if err != nil {
		return model.BayesianOutcomeResponse{}, err
	}
	if request.ObservedAt.Before(journal.AsOf) {
		return model.BayesianOutcomeResponse{}, errors.New("outcome predates the recall decision")
	}
	var decision *model.BayesianDecision
	for index := range journal.Report.Decisions {
		if journal.Report.Decisions[index].EventID == request.EventID {
			decision = &journal.Report.Decisions[index]
			break
		}
	}
	if decision == nil {
		return model.BayesianOutcomeResponse{}, errors.New("outcome event was not nominated by the referenced journal")
	}
	if request.Source == model.OutcomeSelected {
		current := s.store.Snapshot(ctx)
		if current.PolicyVersion != journal.Snapshot.PolicyVersion || current.EvidenceEpoch != journal.Snapshot.EvidenceEpoch {
			return model.BayesianOutcomeResponse{}, errors.New("selective outcome journal is stale for the current policy or evidence epoch")
		}
	}
	weight, err := bayesianOutcomeWeight(request, *decision, journal.Report)
	if err != nil {
		return model.BayesianOutcomeResponse{}, err
	}
	digest, err := bayesianOutcomeDigest(request)
	if err != nil {
		return model.BayesianOutcomeResponse{}, err
	}
	if decision.Forecast.HorizonKey != model.RetrievalUsefulnessHorizon || decision.Forecast.ModelKind == "" {
		return model.BayesianOutcomeResponse{}, errors.New("referenced journal lacks a Phase 4 forecast commitment")
	}
	residualObservation := model.ResidualObservation{
		ActionKey:  residualActionKey(journal.QueryDigest, request.EventID, decision.Forecast.HorizonKey),
		GeneralKey: residualGeneralKey(decision.PosteriorKey, decision.Forecast.HorizonKey),
		HorizonKey: decision.Forecast.HorizonKey, BaseProbability: decision.Forecast.PreResidualLaw.Useful, CommittedProbability: decision.Forecast.CorrectedLaw.Useful,
		Useful: request.Useful, ValidationEligible: request.Source == model.OutcomeFullStream || request.Source == model.OutcomeIndependentAudit,
		EventID: request.EventID, JournalID: request.JournalID, AvailableAt: request.AvailableAt,
		PosteriorKey: decision.PosteriorKey,
	}
	result, err := s.store.ApplyBayesianOutcome(ctx, request, decision.PosteriorKey, decision.ParentPosteriorKey, digest, weight, s.config.BayesianChangePolicy, s.config.BayesianGroupPolicy, residualObservation, s.config.ResidualPolicy)
	if err != nil {
		return model.BayesianOutcomeResponse{}, err
	}
	return model.BayesianOutcomeResponse{ProtocolVersion: model.ProtocolVersion, Duplicate: result.Duplicate, ChangePoint: result.ChangePoint, Revision: result.Revision, Posterior: result.Posterior, Snapshot: result.Snapshot}, nil
}

// CompareBayesianGroup is a proposal-only slow-path operation. It never
// publishes an Anti-Pigeon certificate or changes the active posterior keys.
func (s *Service) CompareBayesianGroup(ctx context.Context, request model.BayesianGroupComparisonRequest) (model.BayesianGroupComparison, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.BayesianGroupComparison{}, err
	}
	if strings.TrimSpace(request.TenantID) == "" || len(request.MemberEventIDs) < 2 || len(request.MemberEventIDs) > s.config.BayesianGroupPolicy.MaxMembers {
		return model.BayesianGroupComparison{}, errors.New("tenant_id and between two and the configured maximum member_event_ids are required")
	}
	ids := append([]string(nil), request.MemberEventIDs...)
	sort.Strings(ids)
	for index, id := range ids {
		if strings.TrimSpace(id) == "" || (index > 0 && id == ids[index-1]) {
			return model.BayesianGroupComparison{}, errors.New("member_event_ids must be non-empty and unique")
		}
	}
	snapshot := s.store.Snapshot(ctx)
	now := time.Now().UTC()
	members := make([]model.BayesianGroupMember, 0, len(ids))
	for _, eventID := range ids {
		posteriorKey := eventID
		certificate, certificateErr := s.store.GetAntiPigeonCertificate(ctx, request.TenantID, []string{eventID})
		if certificateErr == nil && antiPigeonCertificateActive(certificate, snapshot, now) {
			posteriorKey = "ap:" + certificate.ID
		} else if certificateErr != nil && !errors.Is(certificateErr, store.ErrCertificateNotFound) {
			return model.BayesianGroupComparison{}, certificateErr
		}
		member := model.BayesianGroupMember{EventID: eventID, CurrentPosteriorKey: posteriorKey}
		posterior, posteriorErr := s.store.GetBayesianPosterior(ctx, request.TenantID, posteriorKey)
		if posteriorErr == nil && posterior.Certified && posterior.EvidenceEpoch == snapshot.EvidenceEpoch {
			evidence := posterior.MemberEvidence[eventID]
			member.UsefulWeight, member.NotUsefulWeight = evidence.UsefulWeight, evidence.NotUsefulWeight
		} else if posteriorErr != nil && !errors.Is(posteriorErr, store.ErrPosteriorNotFound) {
			return model.BayesianGroupComparison{}, posteriorErr
		}
		members = append(members, member)
	}
	if current := s.store.Snapshot(ctx); current != snapshot {
		return model.BayesianGroupComparison{}, store.ErrStaleSnapshot
	}
	comparison := bayes.CompareGroup(members, s.config.BayesianGroupPolicy)
	comparison.ProtocolVersion = model.ProtocolVersion
	comparison.TenantID = request.TenantID
	comparison.HorizonKey = model.RetrievalUsefulnessHorizon
	comparison.Snapshot = snapshot
	return comparison, nil
}

func (s *Service) PublishPredictiveSnap(ctx context.Context, request model.PredictiveSnapRequest) (model.PredictiveSnapResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.PredictiveSnapResponse{}, err
	}
	if strings.TrimSpace(request.ID) == "" || strings.TrimSpace(request.TenantID) == "" || strings.TrimSpace(request.Issuer) == "" || strings.TrimSpace(request.Procedure) == "" {
		return model.PredictiveSnapResponse{}, errors.New("snap id, tenant_id, issuer, and procedure are required")
	}
	if !request.ExternalAudit {
		return model.PredictiveSnapResponse{}, errors.New("predictive snap requires an external confirmation audit")
	}
	currentSnapshot := s.store.Snapshot(ctx)
	if request.BaseSnapshot != currentSnapshot {
		return model.PredictiveSnapResponse{}, store.ErrStaleSnapshot
	}
	if request.Candidate.TenantID != request.TenantID {
		return model.PredictiveSnapResponse{}, errors.New("candidate graph tenant does not match request")
	}
	if err := graphpolicy.ValidateCandidate(request.Candidate, s.config.SnapPolicy); err != nil {
		return model.PredictiveSnapResponse{}, err
	}
	now := time.Now().UTC()
	if request.DesignStart.IsZero() || !request.DesignStart.Before(request.DesignEnd) || !request.DesignEnd.Before(request.ConfirmationStart) || !request.ConfirmationStart.Before(request.ConfirmationEnd) || request.ConfirmationEnd.After(now) {
		return model.PredictiveSnapResponse{}, errors.New("snap requires disjoint ordered design and confirmation windows available by the current time")
	}
	if request.CandidateFamilySize < 1 || request.CandidateFamilySize > s.config.SnapPolicy.MaxCandidateFamily || !request.UnchangedCandidateIncluded {
		return model.PredictiveSnapResponse{}, errors.New("snap candidate family must be bounded and include the unchanged graph")
	}
	if !unitOpen(request.SimultaneousCoverage) || request.SimultaneousCoverage < s.config.SnapPolicy.MinSimultaneousCoverage {
		return model.PredictiveSnapResponse{}, errors.New("snap simultaneous coverage is below policy")
	}
	currentGraph, err := s.store.GetPredictiveGraph(ctx, request.TenantID)
	if err != nil {
		return model.PredictiveSnapResponse{}, err
	}
	if currentGraph.Version != request.BaseSnapshot.GraphVersion {
		return model.PredictiveSnapResponse{}, store.ErrStaleSnapshot
	}
	closure := graphpolicy.DependencyClosure(currentGraph, request.Candidate, s.config.SnapPolicy.ClosureRadius)
	if len(closure.NodeIDs) == 0 && len(closure.EdgeIDs) == 0 {
		return model.PredictiveSnapResponse{ProtocolVersion: model.ProtocolVersion, Reason: "candidate is identical to the published graph", Graph: currentGraph, Closure: closure, Snapshot: currentSnapshot}, nil
	}
	burden, err := graphpolicy.UnresolvedBurden(request.Candidate, request.Obligations)
	if err != nil {
		return model.PredictiveSnapResponse{}, err
	}
	if err := validateSnapCertificates(request, closure, s.config.SnapPolicy); err != nil {
		return model.PredictiveSnapResponse{}, err
	}
	netGain := request.PriorityGainLCB - request.ResourceCostUCB
	if !finiteNonnegative(request.ResourceCostUCB) || math.IsNaN(request.PriorityGainLCB) || math.IsInf(request.PriorityGainLCB, 0) || !finiteNonnegative(request.ProperRiskIncreaseUCB) {
		return model.PredictiveSnapResponse{}, errors.New("snap risk, gain, and cost bounds must be finite")
	}
	if netGain <= s.config.SnapPolicy.MinNetPriorityGain || request.ProperRiskIncreaseUCB > s.config.SnapPolicy.MaxProperRiskIncrease || burden > s.config.SnapPolicy.MaxUnresolvedBurden {
		return model.PredictiveSnapResponse{ProtocolVersion: model.ProtocolVersion, Reason: "candidate failed confirmation acceptance gates", Graph: currentGraph, Closure: closure, Snapshot: currentSnapshot}, nil
	}
	candidate := request.Candidate
	candidate.Version = currentSnapshot.GraphVersion + 1
	candidate.PublishedAt = now
	candidate.SourceSnapID = request.ID
	record := model.PredictiveSnapRecord{ID: request.ID, TenantID: request.TenantID, PreviousGraph: currentGraph, PublishedGraph: candidate, Closure: closure, UnresolvedBurden: burden, SimultaneousCoverage: request.SimultaneousCoverage, Procedure: request.Procedure, Issuer: request.Issuer, PublishedAt: now}
	published, snapshot, err := s.store.PublishPredictiveSnap(ctx, record)
	if err != nil {
		return model.PredictiveSnapResponse{}, err
	}
	return model.PredictiveSnapResponse{ProtocolVersion: model.ProtocolVersion, Accepted: true, Reason: "confirmed predictive snap published", Graph: published, Closure: closure, Snapshot: snapshot}, nil
}

func (s *Service) GetPredictiveGraph(ctx context.Context, tenantID string) (model.PredictiveGraphResponse, error) {
	if strings.TrimSpace(tenantID) == "" {
		return model.PredictiveGraphResponse{}, errors.New("tenant_id is required")
	}
	graph, err := s.store.GetPredictiveGraph(ctx, tenantID)
	if err != nil {
		return model.PredictiveGraphResponse{}, err
	}
	snapshot := s.store.Snapshot(ctx)
	if graph.Version != snapshot.GraphVersion {
		return model.PredictiveGraphResponse{}, store.ErrStaleSnapshot
	}
	return model.PredictiveGraphResponse{ProtocolVersion: model.ProtocolVersion, Graph: graph, Snapshot: snapshot}, nil
}

func (s *Service) IssueAgencyProposal(ctx context.Context, request model.IssueAgencyProposalRequest) (model.IssueAgencyProposalResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.IssueAgencyProposalResponse{}, err
	}
	if !s.config.AgencyPolicy.Enabled || s.config.AgencySigner == nil {
		return model.IssueAgencyProposalResponse{}, errors.New("agency is disabled")
	}
	if subtle.ConstantTimeCompare([]byte(request.IssuerToken), []byte(s.config.AgencyIssuerToken)) != 1 {
		return model.IssueAgencyProposalResponse{}, errors.New("agency issuer authentication failed")
	}
	now := time.Now().UTC()
	proposal, err := agency.BuildProposal(request.Proposal, now, s.config.AgencyPolicy)
	if err != nil {
		return model.IssueAgencyProposalResponse{}, err
	}
	signed, err := s.config.AgencySigner.Sign(proposal)
	if err != nil {
		return model.IssueAgencyProposalResponse{}, err
	}
	digest, err := agencyProposalDigest(proposal)
	if err != nil {
		return model.IssueAgencyProposalResponse{}, err
	}
	record := model.AgencyProposalRecord{Proposal: proposal, Signed: signed, Status: model.AgencyPending}
	result, err := s.store.PutAgencyProposal(ctx, record, digest, s.config.AgencyPolicy.MaxProposalsPerChain, s.config.AgencyPolicy.MaxPendingPerTenant, now)
	if err != nil {
		return model.IssueAgencyProposalResponse{}, err
	}
	return model.IssueAgencyProposalResponse{ProtocolVersion: model.ProtocolVersion, Duplicate: result.Duplicate, Record: result.Record, Snapshot: result.Snapshot}, nil
}

func (s *Service) ClaimAgencyProposals(ctx context.Context, request model.ClaimAgencyProposalsRequest) (model.ClaimAgencyProposalsResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.ClaimAgencyProposalsResponse{}, err
	}
	if !s.config.AgencyPolicy.Enabled {
		return model.ClaimAgencyProposalsResponse{}, errors.New("agency is disabled")
	}
	if subtle.ConstantTimeCompare([]byte(request.AuthorityToken), []byte(s.config.AgencyAuthorityToken)) != 1 {
		return model.ClaimAgencyProposalsResponse{}, errors.New("agency authority authentication failed")
	}
	request.TenantID, request.ConsumerID = strings.TrimSpace(request.TenantID), strings.TrimSpace(request.ConsumerID)
	if !boundedAgencyIdentifier(request.TenantID) || !boundedAgencyIdentifier(request.ConsumerID) {
		return model.ClaimAgencyProposalsResponse{}, errors.New("tenant_id and consumer_id are required")
	}
	if request.Limit == 0 {
		request.Limit = min(10, s.config.AgencyPolicy.MaxClaims)
	}
	if request.Limit < 1 || request.Limit > s.config.AgencyPolicy.MaxClaims {
		return model.ClaimAgencyProposalsResponse{}, errors.New("agency claim limit exceeds policy")
	}
	records, snapshot, err := s.store.ClaimAgencyProposals(ctx, request.TenantID, request.ConsumerID, time.Now().UTC(), request.Limit, s.config.AgencyPolicy.LeaseDuration)
	if err != nil {
		return model.ClaimAgencyProposalsResponse{}, err
	}
	return model.ClaimAgencyProposalsResponse{ProtocolVersion: model.ProtocolVersion, Records: records, Snapshot: snapshot}, nil
}

func (s *Service) ResolveAgencyProposal(ctx context.Context, request model.ResolveAgencyProposalRequest) (model.ResolveAgencyProposalResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.ResolveAgencyProposalResponse{}, err
	}
	if !s.config.AgencyPolicy.Enabled {
		return model.ResolveAgencyProposalResponse{}, errors.New("agency is disabled")
	}
	if subtle.ConstantTimeCompare([]byte(request.AuthorityToken), []byte(s.config.AgencyAuthorityToken)) != 1 {
		return model.ResolveAgencyProposalResponse{}, errors.New("agency authority authentication failed")
	}
	request.TenantID, request.ProposalID, request.ConsumerID = strings.TrimSpace(request.TenantID), strings.TrimSpace(request.ProposalID), strings.TrimSpace(request.ConsumerID)
	request.Reason, request.ExecutionRef = strings.TrimSpace(request.Reason), strings.TrimSpace(request.ExecutionRef)
	if !boundedAgencyIdentifier(request.TenantID) || !boundedAgencyIdentifier(request.ProposalID) || !boundedAgencyIdentifier(request.ConsumerID) || request.Reason == "" || len(request.Reason) > 1024 || len(request.ExecutionRef) > 512 {
		return model.ResolveAgencyProposalResponse{}, errors.New("bounded tenant, proposal, consumer, and resolution reason are required")
	}
	if request.Decision != model.AgencyApproved && request.Decision != model.AgencyRejected {
		return model.ResolveAgencyProposalResponse{}, errors.New("agency resolution must be approved or rejected")
	}
	if request.Decision == model.AgencyApproved && request.ExecutionRef == "" {
		return model.ResolveAgencyProposalResponse{}, errors.New("approved agency resolution requires an execution_ref")
	}
	if request.Decision == model.AgencyRejected && request.ExecutionRef != "" {
		return model.ResolveAgencyProposalResponse{}, errors.New("rejected agency resolution cannot carry an execution_ref")
	}
	result, err := s.store.ResolveAgencyProposal(ctx, request, time.Now().UTC())
	if err != nil {
		return model.ResolveAgencyProposalResponse{}, err
	}
	return model.ResolveAgencyProposalResponse{ProtocolVersion: model.ProtocolVersion, Duplicate: result.Duplicate, Record: result.Record, Snapshot: result.Snapshot}, nil
}

func boundedAgencyIdentifier(value string) bool {
	return value != "" && len([]byte(value)) <= maxAgencyIdentifierBytes
}

func (s *Service) RollbackPredictiveSnap(ctx context.Context, request model.RollbackSnapRequest) (model.PredictiveSnapResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.PredictiveSnapResponse{}, err
	}
	if request.TenantID == "" || request.SnapID == "" || strings.TrimSpace(request.Reason) == "" {
		return model.PredictiveSnapResponse{}, errors.New("tenant_id, snap_id, and reason are required")
	}
	graph, snapshot, err := s.store.RollbackPredictiveSnap(ctx, request.TenantID, request.SnapID, request.Reason)
	if err != nil {
		return model.PredictiveSnapResponse{}, err
	}
	return model.PredictiveSnapResponse{ProtocolVersion: model.ProtocolVersion, Accepted: true, Reason: "predictive snap rolled back; dependent state remains invalidated", Graph: graph, Snapshot: snapshot}, nil
}

func validateSnapCertificates(request model.PredictiveSnapRequest, closure model.DependencyClosure, policy graphpolicy.Policy) error {
	nodes, edges := make(map[string]model.CompatibilityNode), make(map[string]model.CompatibilityEdge)
	for _, node := range request.Candidate.Nodes {
		nodes[node.ID] = node
	}
	for _, edge := range request.Candidate.Edges {
		edges[edge.ID] = edge
	}
	buckets := make(map[string]model.SnapBucketCertificate)
	for _, certificate := range request.BucketCertificates {
		buckets[certificate.NodeID] = certificate
	}
	edgeCertificates := make(map[string]model.SnapEdgeCertificate)
	for _, certificate := range request.EdgeCertificates {
		edgeCertificates[certificate.EdgeID] = certificate
	}
	for _, id := range closure.NodeIDs {
		node, retained := nodes[id]
		if !retained || len(node.MemberEventIDs) == 0 {
			continue
		}
		certificate, ok := buckets[id]
		if !ok || !finiteNonnegative(certificate.FutureDiameterUCB) || !finiteNonnegative(certificate.DiameterLimit) || certificate.FutureDiameterUCB > certificate.DiameterLimit || !finiteNonnegative(certificate.EffectiveSupport) || certificate.EffectiveSupport < policy.MinBucketSupport {
			return errors.New("every affected active bucket requires a passing external future-diameter certificate")
		}
	}
	for _, id := range closure.EdgeIDs {
		if _, retained := edges[id]; !retained {
			continue
		}
		certificate, ok := edgeCertificates[id]
		if !ok || !finiteNonnegative(certificate.DefectUCB) || !finiteNonnegative(certificate.DefectLimit) || certificate.DefectUCB > certificate.DefectLimit {
			return errors.New("every affected retained or new edge requires a passing compatibility certificate")
		}
	}
	return nil
}

func (s *Service) Close() error {
	var rankDeltaErr error
	if s.rankDelta != nil {
		rankDeltaErr = s.rankDelta.Close()
	}
	return errors.Join(rankDeltaErr, s.store.Close())
}

func (s *Service) resolveBudgets(request model.RecallRequest) (int, int, int, error) {
	recallK, packK, tokenBudget := request.RecallK, request.PackK, request.TokenBudget
	if recallK == 0 {
		recallK = s.config.DefaultRecallK
	}
	if packK == 0 {
		packK = s.config.DefaultPackK
	}
	if tokenBudget == 0 {
		tokenBudget = s.config.DefaultTokenBudget
	}
	if recallK <= 0 || packK <= 0 || tokenBudget <= 0 {
		return 0, 0, 0, errors.New("recall_k, pack_k, and token_budget must be positive")
	}
	if recallK > maxRecallK || packK > maxPackK {
		return 0, 0, 0, errors.New("requested budgets exceed v1alpha1 safety caps")
	}
	if packK > recallK {
		return 0, 0, 0, errors.New("pack_k cannot exceed recall_k")
	}
	return recallK, packK, tokenBudget, nil
}

func checkProtocol(version string) error {
	if version != model.ProtocolVersion {
		return fmt.Errorf("unsupported protocol_version %q; expected %q", version, model.ProtocolVersion)
	}
	return nil
}

func eventDigest(event model.Event) (string, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("digest event: %w", err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}

func recallQueryDigest(request model.RecallRequest, vector []float32, activeModel string) (string, error) {
	payload, err := json.Marshal(struct {
		Query          string    `json:"query"`
		Embedding      []float32 `json:"embedding"`
		EmbeddingModel string    `json:"embedding_model"`
	}{Query: request.Query, Embedding: vector, EmbeddingModel: activeModel})
	if err != nil {
		return "", fmt.Errorf("digest recall query: %w", err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}

func bayesianJournalID(entry model.BayesianJournalEntry) (string, error) {
	entry.ID = ""
	entry.Report.JournalID = ""
	entry.Report.JournalDurable = false
	payload, err := json.Marshal(entry)
	if err != nil {
		return "", fmt.Errorf("digest Bayesian journal: %w", err)
	}
	digest := sha256.Sum256(payload)
	return "bj_" + hex.EncodeToString(digest[:16]), nil
}

func validateSelectionCertificate(certificate model.SelectionSupportCertificate, snapshot model.Snapshot, now time.Time) error {
	if strings.TrimSpace(certificate.ID) == "" || strings.TrimSpace(certificate.TenantID) == "" || strings.TrimSpace(certificate.Issuer) == "" || strings.TrimSpace(certificate.Procedure) == "" {
		return errors.New("selection certificate id, tenant_id, issuer, and procedure are required")
	}
	if !certificate.ExternalAudit {
		return errors.New("selection certificate must be issued by an external audit")
	}
	if certificate.PolicyVersion != snapshot.PolicyVersion || certificate.EvidenceEpoch != snapshot.EvidenceEpoch {
		return errors.New("selection certificate does not bind the current policy and evidence epoch")
	}
	if !unitOpen(certificate.MinSelectionProbability) || !unitOpen(certificate.SimultaneousCoverage) {
		return errors.New("selection probability and simultaneous coverage must be finite values in (0,1]")
	}
	if certificate.ValidFrom.IsZero() || certificate.ValidUntil.IsZero() || certificate.ValidUntil.Before(certificate.ValidFrom) || now.Before(certificate.ValidFrom) || !now.Before(certificate.ValidUntil) {
		return errors.New("selection certificate validity window must contain the current time")
	}
	return nil
}

func selectionCertificateActive(certificate model.SelectionSupportCertificate, snapshot model.Snapshot, now time.Time) bool {
	return validateSelectionCertificate(certificate, snapshot, now) == nil
}

func validateAntiPigeonCertificate(certificate model.AntiPigeonCertificate, snapshot model.Snapshot, now time.Time) error {
	if strings.TrimSpace(certificate.ID) == "" || strings.TrimSpace(certificate.TenantID) == "" || strings.TrimSpace(certificate.HorizonKey) == "" || strings.TrimSpace(certificate.Issuer) == "" || strings.TrimSpace(certificate.Procedure) == "" {
		return errors.New("Anti-Pigeon certificate id, tenant_id, horizon_key, issuer, and procedure are required")
	}
	if certificate.HorizonKey != "retrieval-usefulness-v1" {
		return errors.New("Anti-Pigeon certificate horizon_key must be retrieval-usefulness-v1")
	}
	if !certificate.ExternalAudit {
		return errors.New("Anti-Pigeon certificate must be issued by an external audit")
	}
	if certificate.GraphVersion != snapshot.GraphVersion || certificate.EvidenceEpoch != snapshot.EvidenceEpoch {
		return errors.New("Anti-Pigeon certificate does not bind the current graph and evidence epoch")
	}
	if len(certificate.MemberEventIDs) < 2 {
		return errors.New("Anti-Pigeon certificate requires at least two members")
	}
	seen := make(map[string]struct{}, len(certificate.MemberEventIDs))
	for _, eventID := range certificate.MemberEventIDs {
		if strings.TrimSpace(eventID) == "" {
			return errors.New("Anti-Pigeon member event ids cannot be empty")
		}
		if _, duplicate := seen[eventID]; duplicate {
			return errors.New("Anti-Pigeon member event ids must be unique")
		}
		seen[eventID] = struct{}{}
	}
	if !finiteNonnegative(certificate.TargetDiameterUCB) || !finiteNonnegative(certificate.DiameterLimit) || certificate.TargetDiameterUCB > certificate.DiameterLimit {
		return errors.New("Anti-Pigeon target diameter UCB must be finite and within its declared limit")
	}
	if !finiteNonnegative(certificate.EffectiveSupport) || !finiteNonnegative(certificate.MinEffectiveSupport) || certificate.EffectiveSupport < certificate.MinEffectiveSupport || certificate.MinEffectiveSupport <= 0 {
		return errors.New("Anti-Pigeon effective support must meet a positive declared minimum")
	}
	if !unitOpen(certificate.SimultaneousCoverage) || certificate.ValidUntil.IsZero() || !now.Before(certificate.ValidUntil) {
		return errors.New("Anti-Pigeon coverage must be in (0,1] and the certificate must be unexpired")
	}
	return nil
}

func antiPigeonCertificateActive(certificate model.AntiPigeonCertificate, snapshot model.Snapshot, now time.Time) bool {
	return validateAntiPigeonCertificate(certificate, snapshot, now) == nil
}

func validateOmittedInfluenceCertificate(certificate model.OmittedInfluenceCertificate, snapshot model.Snapshot, now time.Time) error {
	if strings.TrimSpace(certificate.ID) == "" || strings.TrimSpace(certificate.TenantID) == "" || strings.TrimSpace(certificate.Issuer) == "" || strings.TrimSpace(certificate.Procedure) == "" {
		return errors.New("omitted-influence certificate id, tenant_id, issuer, and procedure are required")
	}
	if certificate.ExternalAudit == certificate.RuntimeEstimated {
		return errors.New("omitted-influence certificate requires exactly one external or runtime estimator provenance")
	}
	if certificate.RuntimeEstimated && (strings.TrimSpace(certificate.QueryDigest) == "" || strings.TrimSpace(certificate.PopulationDigest) == "") {
		return errors.New("runtime omitted-influence certificate requires query and population bindings")
	}
	if certificate.PolicyVersion != snapshot.PolicyVersion || certificate.EvidenceEpoch != snapshot.EvidenceEpoch {
		return errors.New("omitted-influence certificate does not bind the current policy and evidence epoch")
	}
	if !finiteNonnegative(certificate.DivergenceUCB) || !finiteNonnegative(certificate.DivergenceLimit) || certificate.DivergenceUCB > certificate.DivergenceLimit {
		return errors.New("omitted-influence divergence UCB must be finite and within its declared limit")
	}
	if !unitOpen(certificate.AuditProbability) || !unitOpen(certificate.SimultaneousCoverage) || certificate.ValidUntil.IsZero() || !now.Before(certificate.ValidUntil) {
		return errors.New("omitted-influence audit probability and coverage must be in (0,1] and the certificate must be unexpired")
	}
	return nil
}

func omittedInfluenceCertificateActive(certificate model.OmittedInfluenceCertificate, snapshot model.Snapshot, now time.Time, queryDigest string) bool {
	if validateOmittedInfluenceCertificate(certificate, snapshot, now) != nil {
		return false
	}
	return !certificate.RuntimeEstimated || subtle.ConstantTimeCompare([]byte(certificate.QueryDigest), []byte(queryDigest)) == 1
}

func unitOpen(value float64) bool {
	return value > 0 && value <= 1 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func finiteNonnegative(value float64) bool {
	return value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func bayesianOutcomeWeight(request model.BayesianOutcomeRequest, decision model.BayesianDecision, report model.BayesianShadowReport) (float64, error) {
	var probability float64
	switch request.Source {
	case model.OutcomeFullStream:
		probability = 1
	case model.OutcomeIndependentAudit:
		if !decision.AuditSelected || !unitOpen(decision.AuditProbability) {
			return 0, errors.New("outcome is not backed by an independently selected audit")
		}
		probability = decision.AuditProbability
	case model.OutcomeSelected:
		if !report.SelectionSupportCertified || !decision.Activated || !unitOpen(decision.TotalSelectionProbabilityLowerBound) {
			return 0, errors.New("selective outcome lacks a valid selection-support certificate")
		}
		probability = decision.TotalSelectionProbabilityLowerBound
	default:
		return 0, errors.New("outcome source must be full_stream, independent_audit, or selected")
	}
	if !unitOpen(request.InclusionProbability) || math.Abs(request.InclusionProbability-probability) > 1e-9 {
		return 0, errors.New("outcome inclusion_probability does not match the durable journal")
	}
	return math.Min(20, 1/probability), nil
}

func bayesianOutcomeDigest(request model.BayesianOutcomeRequest) (string, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("digest Bayesian outcome: %w", err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}

func bayesianPolicyDigest(config Config) (string, error) {
	payload, err := json.Marshal(struct {
		Frontier              bayes.Policy               `json:"frontier"`
		BaselineCalibration   calibration.Logit          `json:"baseline_calibration"`
		PredictiveCalibration calibration.Logit          `json:"predictive_calibration"`
		ScoreWeight           float64                    `json:"score_weight"`
		Change                bayes.ChangePolicy         `json:"change"`
		Group                 bayes.GroupPolicy          `json:"group"`
		Ranking               ranking.Policy             `json:"ranking"`
		MaxRankDelta          float64                    `json:"max_rank_delta"`
		ElasticRankDelta      ranking.ElasticDeltaPolicy `json:"elastic_rank_delta"`
		Packing               packing.Policy             `json:"packing"`
		NominationContract    string                     `json:"nomination_contract"`
		RetrievalContract     string                     `json:"retrieval_contract"`
		Residual              residual.Policy            `json:"residual"`
		ResidualMode          string                     `json:"residual_mode"`
		Snap                  graphpolicy.Policy         `json:"snap"`
		Agency                agency.Policy              `json:"agency"`
	}{Frontier: config.BayesianPolicy, BaselineCalibration: config.BaselineCalibration, PredictiveCalibration: config.PredictiveCalibration, ScoreWeight: config.BayesianScoreWeight, Change: config.BayesianChangePolicy, Group: config.BayesianGroupPolicy, Ranking: config.RankingPolicy, MaxRankDelta: config.MaxRankDelta, ElasticRankDelta: config.ElasticRankDelta, Packing: config.PackingPolicy, NominationContract: nominationContractName(config.CandidateRetriever), RetrievalContract: config.CandidateRanker.ContractName(), Residual: config.ResidualPolicy, ResidualMode: config.ResidualMode, Snap: config.SnapPolicy, Agency: config.AgencyPolicy})
	if err != nil {
		return "", fmt.Errorf("encode Bayesian policy: %w", err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}

func nominationContractName(retriever retrieval.CandidateRetriever) string {
	if retriever == nil {
		return "embedded-vector-search"
	}
	return retriever.RetrievalContractName()
}

func agencyProposalDigest(proposal model.AgencyProposal) (string, error) {
	proposal.CreatedAt = time.Time{}
	payload, err := json.Marshal(proposal)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}

func residualActionKey(queryDigest, eventID, horizonKey string) string {
	digest := sha256.Sum256([]byte(queryDigest + "\x00" + eventID + "\x00" + horizonKey))
	return hex.EncodeToString(digest[:16])
}

func residualGeneralKey(posteriorKey, horizonKey string) string {
	digest := sha256.Sum256([]byte(posteriorKey + "\x00" + horizonKey))
	return hex.EncodeToString(digest[:16])
}

func scoreCandidate(candidate model.Candidate, sessionID string, asOf time.Time) float64 {
	similarity := clamp((candidate.Similarity+1)/2, 0, 1)
	recency := candidateRecency(candidate.Event, asOf)
	session := 0.0
	if candidate.Event.SessionID == sessionID {
		session = 1
	}
	return 0.65*similarity + 0.15*candidate.Event.MeanFieldConfidence() + 0.10*recency + 0.05*candidate.Event.Priority + 0.05*session
}

func candidateRecency(event model.Event, asOf time.Time) float64 {
	age := asOf.Sub(event.AvailableAt)
	if age < 0 {
		age = 0
	}
	return math.Exp(-age.Hours() / (24 * 30))
}

func parentPosteriorKey() string {
	return "global:" + model.RetrievalUsefulnessHorizon
}

func estimateTokens(content string) int {
	count := utf8.RuneCountInString(content)
	return max(1, (count+3)/4)
}

func bernoulliLaw(useful float64) model.BernoulliLaw {
	useful = clamp(useful, 0, 1)
	return model.BernoulliLaw{Useful: useful, NotUseful: 1 - useful}
}

func clamp(value, low, high float64) float64 {
	return math.Min(high, math.Max(low, value))
}
