package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/fuzzing"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

const backgroundFuzzProperty = "background-source-event-semantic-bundle"

type BackgroundFuzzPolicy struct {
	Enabled                   bool
	AnswerCertaintyThreshold  float64
	QueueCapacity             int
	WorkerInterval            time.Duration
	JobTimeout                time.Duration
	Cooldown                  time.Duration
	MaxEvents                 int
	MaxPerturbations          int
	StabilityThreshold        float64
	RequiredStableProbability float64
	ConfidenceLevel           float64
	MinTrials                 int
}

func (policy BackgroundFuzzPolicy) withDefaults() BackgroundFuzzPolicy {
	if policy.AnswerCertaintyThreshold == 0 {
		policy.AnswerCertaintyThreshold = .20
	}
	if policy.QueueCapacity == 0 {
		policy.QueueCapacity = 128
	}
	if policy.WorkerInterval == 0 {
		policy.WorkerInterval = 30 * time.Second
	}
	if policy.JobTimeout == 0 {
		policy.JobTimeout = 30 * time.Second
	}
	if policy.Cooldown == 0 {
		policy.Cooldown = 15 * time.Minute
	}
	if policy.MaxEvents == 0 {
		policy.MaxEvents = 8
	}
	if policy.MaxPerturbations == 0 {
		policy.MaxPerturbations = 8
	}
	if policy.StabilityThreshold == 0 {
		policy.StabilityThreshold = .05
	}
	if policy.RequiredStableProbability == 0 {
		policy.RequiredStableProbability = .90
	}
	if policy.ConfidenceLevel == 0 {
		policy.ConfidenceLevel = .95
	}
	if policy.MinTrials == 0 {
		policy.MinTrials = min(8, policy.MaxPerturbations)
	}
	return policy
}

func (policy BackgroundFuzzPolicy) validate() error {
	if !policy.Enabled {
		return nil
	}
	if !finiteUnit(policy.AnswerCertaintyThreshold) || policy.AnswerCertaintyThreshold <= 0 {
		return errors.New("background fuzz answer-certainty threshold must be in (0,1]")
	}
	if policy.QueueCapacity <= 0 || policy.QueueCapacity > 4096 || policy.WorkerInterval <= 0 || policy.JobTimeout <= 0 || policy.Cooldown < 0 {
		return errors.New("background fuzz queue, interval, timeout, and cooldown controls are invalid")
	}
	if policy.MaxEvents < 2 || policy.MaxEvents > 64 || policy.MaxPerturbations <= 0 || policy.MaxPerturbations > 512 || policy.MinTrials <= 0 || policy.MinTrials > policy.MaxPerturbations {
		return errors.New("background fuzz event, perturbation, and trial bounds are invalid")
	}
	if !finiteUnit(policy.StabilityThreshold) || policy.StabilityThreshold <= 0 || !finiteUnit(policy.RequiredStableProbability) || policy.RequiredStableProbability <= 0 {
		return errors.New("background fuzz stability controls must be in (0,1]")
	}
	if policy.ConfidenceLevel != .90 && policy.ConfidenceLevel != .95 && policy.ConfidenceLevel != .99 {
		return errors.New("background fuzz confidence level must be 0.90, 0.95, or 0.99")
	}
	return nil
}

type backgroundFuzzJob struct {
	id              string
	dedupKey        string
	tenantID        string
	asOf            time.Time
	snapshot        model.Snapshot
	queryDigest     string
	queryVector     []float32
	eventIDs        []string
	perturbations   []model.FuzzPerturbation
	answerCertainty float64
	triggerReason   string
}

type backgroundFuzzQueue struct {
	service   *Service
	policy    BackgroundFuzzPolicy
	jobs      chan backgroundFuzzJob
	ctx       context.Context
	cancel    context.CancelFunc
	wait      sync.WaitGroup
	closeOnce sync.Once

	mu                sync.Mutex
	dedup             map[string]time.Time
	running           int
	enqueuedTotal     uint64
	completedTotal    uint64
	failedTotal       uint64
	staleTotal        uint64
	droppedTotal      uint64
	deduplicatedTotal uint64
	lastResult        *model.BackgroundFuzzResultSummary
}

func newBackgroundFuzzQueue(service *Service, policy BackgroundFuzzPolicy) *backgroundFuzzQueue {
	ctx, cancel := context.WithCancel(context.Background())
	queue := &backgroundFuzzQueue{service: service, policy: policy, jobs: make(chan backgroundFuzzJob, policy.QueueCapacity), ctx: ctx, cancel: cancel, dedup: make(map[string]time.Time)}
	queue.wait.Add(1)
	go queue.run()
	return queue
}

func (s *Service) nominateBackgroundFuzz(request model.RecallRequest, queryDigest string, queryVector []float32, candidates []model.Candidate, packet model.ContextPacket) {
	queue := s.backgroundFuzz
	if queue == nil || packet.PacketAnswerCertainty > queue.policy.AnswerCertaintyThreshold || len(candidates) < 2 {
		return
	}
	limit := min(len(candidates), queue.policy.MaxEvents)
	events := make([]model.Event, limit)
	eventIDs := make([]string, limit)
	for index := range limit {
		events[index] = candidates[index].Event
		eventIDs[index] = candidates[index].Event.ID
	}
	perturbations := backgroundSemanticBundles(events, queue.policy.MaxPerturbations)
	if len(perturbations) == 0 {
		return
	}
	dedupKey := backgroundFuzzDedupKey(request.TenantID, queryDigest, packet.Snapshot, eventIDs)
	job := backgroundFuzzJob{
		id: hex.EncodeToString(sha256Digest("background-fuzz-job-v1", dedupKey)[:16]), dedupKey: dedupKey,
		tenantID: request.TenantID, asOf: request.AsOf.UTC(), snapshot: packet.Snapshot, queryDigest: queryDigest,
		queryVector: append([]float32(nil), queryVector...), eventIDs: eventIDs, perturbations: perturbations,
		answerCertainty: packet.PacketAnswerCertainty, triggerReason: "low-packing-boundary-certainty",
	}
	queue.Enqueue(job)
}

func backgroundSemanticBundles(events []model.Event, maximum int) []model.FuzzPerturbation {
	perturbations := make([]model.FuzzPerturbation, 0, min(len(events), maximum))
	for targetIndex := 0; targetIndex < len(events) && len(perturbations) < maximum; targetIndex++ {
		target := events[targetIndex]
		for offset := 1; offset < len(events); offset++ {
			source := events[(targetIndex+offset)%len(events)]
			if target.What.Value == source.What.Value && target.Why.Value == source.Why.Value && target.How.Value == source.How.Value {
				continue
			}
			evidence := "background audit copy from as-of source EventFrame " + source.ID
			perturbations = append(perturbations, model.FuzzPerturbation{
				ID: "background-bundle-" + fmt.Sprint(len(perturbations)+1), PropertyID: backgroundFuzzProperty,
				EventID: target.ID, SourceEventID: source.ID, ValidityRuleID: "source-event-semantic-bundle-v1",
				ValidationKind: model.FuzzValidationSourceEventBundle,
				Replacements: map[model.FuzzField]model.Field{
					model.FuzzWhat: backgroundField(source.What, evidence+": what"),
					model.FuzzWhy:  backgroundField(source.Why, evidence+": why"),
					model.FuzzHow:  backgroundField(source.How, evidence+": how"),
				},
			})
			break
		}
	}
	return perturbations
}

func backgroundField(source model.Field, evidence string) model.Field {
	return model.Field{Value: source.Value, Source: model.SourceSynthetic, Confidence: source.Confidence, Evidence: evidence}
}

func backgroundFuzzDedupKey(tenantID, queryDigest string, snapshot model.Snapshot, eventIDs []string) string {
	parts := []string{"background-fuzz-v1", tenantID, queryDigest, fmt.Sprintf("%+v", snapshot)}
	canonicalIDs := append([]string(nil), eventIDs...)
	sort.Strings(canonicalIDs)
	parts = append(parts, canonicalIDs...)
	return hex.EncodeToString(sha256Digest(parts...))
}

func sha256Digest(parts ...string) []byte {
	hash := sha256.New()
	for _, part := range parts {
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(part))
	}
	return hash.Sum(nil)
}

func (queue *backgroundFuzzQueue) Enqueue(job backgroundFuzzJob) bool {
	now := time.Now().UTC()
	queue.mu.Lock()
	for key, until := range queue.dedup {
		if !until.IsZero() && !now.Before(until) {
			delete(queue.dedup, key)
		}
	}
	if _, exists := queue.dedup[job.dedupKey]; exists {
		queue.deduplicatedTotal++
		queue.mu.Unlock()
		return false
	}
	queue.dedup[job.dedupKey] = time.Time{}
	queue.mu.Unlock()

	select {
	case queue.jobs <- job:
		queue.mu.Lock()
		queue.enqueuedTotal++
		queue.mu.Unlock()
		return true
	default:
		queue.mu.Lock()
		delete(queue.dedup, job.dedupKey)
		queue.droppedTotal++
		queue.mu.Unlock()
		return false
	}
}

func (queue *backgroundFuzzQueue) run() {
	defer queue.wait.Done()
	ticker := time.NewTicker(queue.policy.WorkerInterval)
	defer ticker.Stop()
	for {
		select {
		case <-queue.ctx.Done():
			return
		case <-ticker.C:
			if queue.service.activeRecalls.Load() != 0 {
				continue
			}
			select {
			case job := <-queue.jobs:
				queue.execute(job)
			default:
			}
		}
	}
}

func (queue *backgroundFuzzQueue) execute(job backgroundFuzzJob) {
	queue.mu.Lock()
	queue.running = 1
	queue.mu.Unlock()
	ctx, cancel := context.WithTimeout(queue.ctx, queue.policy.JobTimeout)
	response, err := queue.service.executeBackgroundFuzz(ctx, job, queue.policy)
	cancel()

	result := model.BackgroundFuzzResultSummary{
		JobID: job.id, TriggerReason: job.triggerReason, AnswerCertainty: job.answerCertainty,
		EventCount: len(job.eventIDs), PerturbationCount: len(job.perturbations), CompletedAt: time.Now().UTC(),
	}
	if err != nil {
		result.Status = "failed"
		result.Error = "background fuzz audit failed"
		if errors.Is(err, store.ErrStaleSnapshot) {
			result.Status = "stale"
			result.Error = "snapshot changed before audit"
		} else if errors.Is(err, context.DeadlineExceeded) {
			result.Error = "background fuzz audit timed out"
		} else if errors.Is(err, context.Canceled) {
			result.Error = "background fuzz audit canceled"
		}
	} else {
		result.Status = "completed"
		result.TrialCount = len(response.Trials)
		result.PropertyCount = len(response.Properties)
		for _, property := range response.Properties {
			if property.ConditionalInvariant {
				result.ConditionalInvariants++
			}
		}
	}

	queue.mu.Lock()
	queue.running = 0
	queue.lastResult = &result
	queue.dedup[job.dedupKey] = result.CompletedAt.Add(queue.policy.Cooldown)
	if err == nil {
		queue.completedTotal++
	} else if errors.Is(err, store.ErrStaleSnapshot) {
		queue.staleTotal++
	} else {
		queue.failedTotal++
	}
	queue.mu.Unlock()
}

func (s *Service) executeBackgroundFuzz(ctx context.Context, job backgroundFuzzJob, policy BackgroundFuzzPolicy) (model.FuzzSensitivityResponse, error) {
	if s.store.Snapshot(ctx) != job.snapshot {
		return model.FuzzSensitivityResponse{}, store.ErrStaleSnapshot
	}
	events, err := s.getEventsForAnalysis(ctx, job.tenantID, job.eventIDs, job.asOf)
	if err != nil {
		return model.FuzzSensitivityResponse{}, err
	}
	predictor, err := fuzzing.NewEmbeddingNominationPredictorFromVector(s.embedder, job.queryVector, job.queryDigest, snapshotCacheNamespace(job.snapshot), s.nominationCache)
	if err != nil {
		return model.FuzzSensitivityResponse{}, err
	}
	request := model.FuzzSensitivityRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: job.tenantID, Query: "background-query-vector:" + job.queryDigest,
		AsOf: job.asOf, BaseSnapshot: job.snapshot, EventIDs: append([]string(nil), job.eventIDs...),
		Perturbations:      append([]model.FuzzPerturbation(nil), job.perturbations...),
		StabilityThreshold: policy.StabilityThreshold, RequiredStableProbability: policy.RequiredStableProbability,
		ConfidenceLevel: policy.ConfidenceLevel, MinTrials: policy.MinTrials,
	}
	response, err := fuzzing.Evaluate(ctx, request, events, predictor)
	if err != nil {
		return model.FuzzSensitivityResponse{}, err
	}
	if s.store.Snapshot(ctx) != job.snapshot {
		return model.FuzzSensitivityResponse{}, store.ErrStaleSnapshot
	}
	response.Snapshot = job.snapshot
	return response, nil
}

func (s *Service) BackgroundFuzzStatus() model.BackgroundFuzzQueueStatus {
	if s.backgroundFuzz == nil {
		return model.BackgroundFuzzQueueStatus{ProtocolVersion: model.ProtocolVersion}
	}
	return s.backgroundFuzz.Status()
}

func (queue *backgroundFuzzQueue) Status() model.BackgroundFuzzQueueStatus {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	status := model.BackgroundFuzzQueueStatus{
		ProtocolVersion: model.ProtocolVersion, Enabled: true, Capacity: cap(queue.jobs), QueueDepth: len(queue.jobs), Running: queue.running,
		EnqueuedTotal: queue.enqueuedTotal, CompletedTotal: queue.completedTotal, FailedTotal: queue.failedTotal,
		StaleTotal: queue.staleTotal, DroppedTotal: queue.droppedTotal, DeduplicatedTotal: queue.deduplicatedTotal,
	}
	if queue.lastResult != nil {
		copy := *queue.lastResult
		status.LastResult = &copy
	}
	return status
}

func (queue *backgroundFuzzQueue) Close() error {
	queue.closeOnce.Do(queue.cancel)
	queue.wait.Wait()
	return nil
}

func finiteUnit(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}
