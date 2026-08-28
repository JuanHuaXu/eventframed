package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type Config struct {
	DefaultRecallK      int
	DefaultPackK        int
	DefaultTokenBudget  int
	OverfetchMultiplier int
	Quantization        string
}

type Service struct {
	store          store.EventStore
	embedder       embed.Embedder
	config         Config
	runtimeVersion atomic.Uint64
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
	}
	digest, err := eventDigest(request.Event)
	if err != nil {
		return model.ObserveResponse{}, err
	}
	result, err := s.store.Put(ctx, request.Event, vector, digest)
	if err != nil {
		return model.ObserveResponse{}, err
	}
	if !result.Duplicate {
		s.runtimeVersion.Add(1)
	}
	return model.ObserveResponse{
		ProtocolVersion: model.ProtocolVersion,
		EventID:         request.Event.ID,
		Duplicate:       result.Duplicate,
		Snapshot:        s.snapshot(),
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
		candidate.Score = scoreCandidate(candidate, request.SessionID, request.AsOf)
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
		Snapshot:        s.snapshot(),
	}, nil
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
		Snapshot:        s.snapshot(),
	}, nil
}

func (s *Service) Close() error { return s.store.Close() }

func (s *Service) snapshot() model.Snapshot {
	return model.Snapshot{RuntimeVersion: s.runtimeVersion.Load(), PolicyVersion: 1, ContractVersion: 1}
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
	event.Embedding = nil
	payload, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("digest event: %w", err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
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

func clamp(value, low, high float64) float64 {
	return math.Min(high, math.Max(low, value))
}
