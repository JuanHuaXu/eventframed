package service

import (
	"context"
	"crypto/sha256"
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

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	graphpolicy "github.com/JuanHuaXu/eventframed/internal/graph"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type Config struct {
	DefaultRecallK       int
	DefaultPackK         int
	DefaultTokenBudget   int
	OverfetchMultiplier  int
	Quantization         string
	BayesianPolicy       bayes.Policy
	BayesianScoreWeight  float64
	BayesianChangePolicy bayes.ChangePolicy
	ResidualPolicy       residual.Policy
	SnapPolicy           graphpolicy.Policy
}

type Service struct {
	store    store.EventStore
	embedder embed.Embedder
	config   Config
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
		config.BayesianPolicy = bayes.Policy{VectorWeight: .6, NeighborWeight: .15, NoveltyWeight: .15, IndependenceWeight: .1, Threshold: .72, CriticalThreshold: .55, AuditProbability: .02, MaxActive: 8, AuditSeed: "eventframe-v1"}
	}
	if config.BayesianScoreWeight == 0 {
		config.BayesianScoreWeight = 0.10
	}
	if config.BayesianScoreWeight < 0 || config.BayesianScoreWeight > 0.25 {
		return nil, errors.New("Bayesian score weight must be in [0,0.25]")
	}
	if !config.BayesianChangePolicy.Valid() {
		config.BayesianChangePolicy = bayes.ChangePolicy{Hazard: .05, Threshold: .30, MaxRun: 64}
	}
	if !config.ResidualPolicy.Valid() {
		config.ResidualPolicy = residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .10, MaxAge: 30 * 24 * time.Hour, ImprovementDelta: .001}
	}
	if !config.SnapPolicy.Valid() {
		config.SnapPolicy = graphpolicy.Policy{MaxNodes: 256, MaxEdges: 512, MaxCandidateFamily: 32, ClosureRadius: 2, MinNetPriorityGain: .01, MaxProperRiskIncrease: .01, MaxUnresolvedBurden: 0, MinSimultaneousCoverage: .95, MinBucketSupport: 30}
	}
	policyDigest, err := bayesianPolicyDigest(config)
	if err != nil {
		return nil, err
	}
	if _, err := eventStore.BindBayesianPolicy(context.Background(), policyDigest); err != nil {
		return nil, fmt.Errorf("bind Bayesian policy: %w", err)
	}
	return &Service{store: eventStore, embedder: embedder, config: config}, nil
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
		vector, err = s.embedder.Embed(request.Event.EmbeddingText())
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
		vector, err = s.embedder.Embed(request.Query)
		if err != nil {
			return model.ContextPacket{}, fmt.Errorf("embed query: %w", err)
		}
	} else if len(vector) != s.embedder.Dimension() {
		return model.ContextPacket{}, fmt.Errorf("query embedding dimension %d does not match %d", len(vector), s.embedder.Dimension())
	} else if request.EmbeddingModel != s.embedder.ModelKey() {
		return model.ContextPacket{}, errors.New("query embedding_model does not match active model")
	}
	searchLimit := recallK * s.config.OverfetchMultiplier
	results, err := s.store.Search(ctx, request.TenantID, vector, request.AsOf, searchLimit)
	if err != nil {
		return model.ContextPacket{}, err
	}
	candidates := make([]model.Candidate, 0, min(recallK, len(results)))
	eligible := 0
	for _, result := range results {
		if result.Event.AvailableAt.After(request.AsOf) {
			continue
		}
		eligible++
		candidate := model.Candidate{
			Event:           result.Event,
			Similarity:      result.Similarity,
			EstimatedTokens: estimateTokens(result.Event.Content),
		}
		candidate.BaselineScore = scoreCandidate(candidate, request.SessionID, request.AsOf)
		candidate.Score = candidate.BaselineScore
		candidates = append(candidates, candidate)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Event.AvailableAt.After(candidates[j].Event.AvailableAt)
		}
		return candidates[i].Score > candidates[j].Score
	})
	if len(candidates) > recallK {
		candidates = candidates[:recallK]
	}
	shadowInputs := make([]bayes.Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		shadowInputs = append(shadowInputs, bayes.Candidate{EventID: candidate.Event.ID, VectorRelevance: clamp((candidate.Similarity+1)/2, 0, 1), Novelty: 1 - candidate.Event.MeanFieldConfidence(), SourceIndependence: sourceIndependence(candidate.Event), Priority: candidate.Event.Priority, EvidenceReady: !candidate.Event.AvailableAt.After(request.AsOf)})
	}
	snapshot := s.store.Snapshot(ctx)
	shadow := bayes.Evaluate(shadowInputs, snapshot.EvidenceEpoch, s.config.BayesianPolicy)
	selectionCertificate, certificateErr := s.store.GetSelectionCertificate(ctx, request.TenantID)
	if certificateErr != nil && !errors.Is(certificateErr, store.ErrCertificateNotFound) {
		return model.ContextPacket{}, certificateErr
	}
	omittedCertificate, omittedErr := s.store.GetOmittedInfluenceCertificate(ctx, request.TenantID)
	if omittedErr != nil && !errors.Is(omittedErr, store.ErrCertificateNotFound) {
		return model.ContextPacket{}, omittedErr
	}
	now := time.Now().UTC()
	omittedCertified := omittedErr == nil && omittedInfluenceCertificateActive(omittedCertificate, snapshot, now)
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
		decisions := make(map[string]model.BayesianDecision, len(shadow.Decisions))
		for _, decision := range shadow.Decisions {
			decisions[decision.EventID] = decision
		}
		applied := false
		for index := range candidates {
			decision := decisions[candidates[index].Event.ID]
			if !decision.Activated || decision.TotalSelectionProbabilityLowerBound <= 0 {
				continue
			}
			posterior, getErr := s.store.GetBayesianPosterior(ctx, request.TenantID, decision.PosteriorKey)
			if errors.Is(getErr, store.ErrPosteriorNotFound) {
				continue
			}
			if getErr != nil {
				return model.ContextPacket{}, getErr
			}
			if !posterior.Certified || posterior.EvidenceEpoch != snapshot.EvidenceEpoch {
				continue
			}
			probability := clamp(posterior.Mean(), 0, 1)
			candidates[index].BayesianProbability = probability
			candidates[index].BayesianApplied = true
			candidates[index].Score = (1-s.config.BayesianScoreWeight)*candidates[index].BaselineScore + s.config.BayesianScoreWeight*probability
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
	queryDigest, err := recallQueryDigest(request, vector, s.embedder.ModelKey())
	if err != nil {
		return model.ContextPacket{}, err
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
		preResidual := bernoulliLaw(candidate.Score)
		forecast := model.ForecastBundle{
			ModelKind: "plugin-bernoulli-retrieval-usefulness", HorizonKey: model.RetrievalUsefulnessHorizon,
			BaseLaw: bernoulliLaw(candidate.BaselineScore), PreResidualLaw: preResidual, CorrectedLaw: preResidual,
			Template:     model.ForecastTemplate{EventID: candidate.Event.ID, PredictedUseful: preResidual.Useful >= .5, Confidence: math.Max(preResidual.Useful, preResidual.NotUseful)},
			PosteriorKey: decision.PosteriorKey, PosteriorVersion: snapshot.PosteriorVersion,
		}
		if candidate.BayesianApplied {
			belief := bernoulliLaw(candidate.BayesianProbability)
			forecast.BeliefLaw = &belief
		}
		actionKey := residualActionKey(queryDigest, candidate.Event.ID, forecast.HorizonKey)
		generalKey := residualGeneralKey(decision.PosteriorKey, forecast.HorizonKey)
		cached, getErr := s.store.GetResidualCandidates(ctx, request.TenantID, actionKey, generalKey)
		if getErr != nil {
			return model.ContextPacket{}, getErr
		}
		var selected *model.ResidualRecord
		if cached.Exact != nil && residual.Eligible(*cached.Exact, preResidual.Useful, snapshot, now, s.config.ResidualPolicy) {
			selected = cached.Exact
		} else if cached.General != nil && residual.Eligible(*cached.General, preResidual.Useful, snapshot, now, s.config.ResidualPolicy) {
			selected = cached.General
		}
		if selected != nil {
			forecast.CorrectedLaw = residual.Apply(preResidual, *selected, s.config.ResidualPolicy)
			forecast.ResidualApplied = true
			forecast.ResidualRecordID = selected.ID
			forecast.Template = model.ForecastTemplate{EventID: candidate.Event.ID, PredictedUseful: forecast.CorrectedLaw.Useful >= .5, Confidence: math.Max(forecast.CorrectedLaw.Useful, forecast.CorrectedLaw.NotUseful)}
			candidate.Score = forecast.CorrectedLaw.Useful
			shadow.ResidualApplied++
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
	packed := make([]model.Candidate, 0, min(packK, len(candidates)))
	usedTokens := 0
	for _, candidate := range candidates {
		if len(packed) >= packK {
			break
		}
		if usedTokens+candidate.EstimatedTokens > tokenBudget {
			continue
		}
		packed = append(packed, candidate)
		usedTokens += candidate.EstimatedTokens
	}
	return model.ContextPacket{
		ProtocolVersion: model.ProtocolVersion,
		Candidates:      packed,
		Recalled:        len(results),
		Eligible:        eligible,
		Packed:          len(packed),
		UsedTokens:      usedTokens,
		Snapshot:        snapshot,
		BayesianShadow:  shadow,
	}, nil
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

func (s *Service) Health(ctx context.Context) (model.HealthResponse, error) {
	stats, err := s.store.Stats(ctx)
	if err != nil {
		return model.HealthResponse{}, err
	}
	return model.HealthResponse{
		ProtocolVersion: model.ProtocolVersion,
		Status:          "ok",
		Store:           stats.Backend,
		Dimension:       s.embedder.Dimension(),
		Quantization:    s.config.Quantization,
		Snapshot:        s.store.Snapshot(ctx),
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

func (s *Service) ObserveBayesianOutcome(ctx context.Context, request model.BayesianOutcomeRequest) (model.BayesianOutcomeResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.BayesianOutcomeResponse{}, err
	}
	if strings.TrimSpace(request.IdempotencyKey) == "" || strings.TrimSpace(request.TenantID) == "" || strings.TrimSpace(request.JournalID) == "" || strings.TrimSpace(request.EventID) == "" {
		return model.BayesianOutcomeResponse{}, errors.New("idempotency_key, tenant_id, journal_id, and event_id are required")
	}
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
	result, err := s.store.ApplyBayesianOutcome(ctx, request, decision.PosteriorKey, digest, weight, s.config.BayesianChangePolicy, residualObservation, s.config.ResidualPolicy)
	if err != nil {
		return model.BayesianOutcomeResponse{}, err
	}
	return model.BayesianOutcomeResponse{ProtocolVersion: model.ProtocolVersion, Duplicate: result.Duplicate, ChangePoint: result.ChangePoint, Posterior: result.Posterior, Snapshot: result.Snapshot}, nil
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

func (s *Service) Close() error { return s.store.Close() }

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
	if recallK > 1000 || packK > 100 {
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
	if !certificate.ExternalAudit {
		return errors.New("omitted-influence certificate must be issued by an external audit")
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

func omittedInfluenceCertificateActive(certificate model.OmittedInfluenceCertificate, snapshot model.Snapshot, now time.Time) bool {
	return validateOmittedInfluenceCertificate(certificate, snapshot, now) == nil
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
		Frontier    bayes.Policy       `json:"frontier"`
		ScoreWeight float64            `json:"score_weight"`
		Change      bayes.ChangePolicy `json:"change"`
		Residual    residual.Policy    `json:"residual"`
		Snap        graphpolicy.Policy `json:"snap"`
	}{Frontier: config.BayesianPolicy, ScoreWeight: config.BayesianScoreWeight, Change: config.BayesianChangePolicy, Residual: config.ResidualPolicy, Snap: config.SnapPolicy})
	if err != nil {
		return "", fmt.Errorf("encode Bayesian policy: %w", err)
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
	age := asOf.Sub(candidate.Event.AvailableAt)
	if age < 0 {
		age = 0
	}
	recency := math.Exp(-age.Hours() / (24 * 30))
	session := 0.0
	if candidate.Event.SessionID == sessionID {
		session = 1
	}
	return 0.65*similarity + 0.15*candidate.Event.MeanFieldConfidence() + 0.10*recency + 0.05*candidate.Event.Priority + 0.05*session
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
