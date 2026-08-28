package memorystore

import (
	"context"
	"errors"
	"math"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type entry struct {
	event  model.Event
	vector []float32
	digest string
}

type Store struct {
	mu                     sync.RWMutex
	entries                map[string]map[string]entry
	journals               map[string]map[string]model.BayesianJournalEntry
	selectionCertificates  map[string]model.SelectionSupportCertificate
	omittedCertificates    map[string]model.OmittedInfluenceCertificate
	antiPigeonCertificates map[string]model.AntiPigeonCertificate
	antiPigeonIndex        map[string]map[string]string
	outcomeDigests         map[string]map[string]string
	posteriors             map[string]map[string]model.BayesianPosterior
	residuals              map[string]map[string]model.ResidualRecord
	graphs                 map[string]model.PredictiveGraph
	snaps                  map[string]map[string]model.PredictiveSnapRecord
	policyDigest           string
	snapshot               model.Snapshot
}

func (s *Store) BindBayesianPolicy(_ context.Context, digest string) (model.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if digest == s.policyDigest {
		return s.snapshot, nil
	}
	s.policyDigest = digest
	s.snapshot.RuntimeVersion++
	s.snapshot.PolicyVersion++
	return s.snapshot, nil
}

func New() *Store {
	return &Store{
		entries: make(map[string]map[string]entry), journals: make(map[string]map[string]model.BayesianJournalEntry),
		selectionCertificates:  make(map[string]model.SelectionSupportCertificate),
		omittedCertificates:    make(map[string]model.OmittedInfluenceCertificate),
		antiPigeonCertificates: make(map[string]model.AntiPigeonCertificate), antiPigeonIndex: make(map[string]map[string]string),
		outcomeDigests: make(map[string]map[string]string), posteriors: make(map[string]map[string]model.BayesianPosterior),
		residuals: make(map[string]map[string]model.ResidualRecord),
		graphs:    make(map[string]model.PredictiveGraph), snaps: make(map[string]map[string]model.PredictiveSnapRecord),
		snapshot: initialSnapshot(),
	}
}

func (s *Store) GetPredictiveGraph(_ context.Context, tenantID string) (model.PredictiveGraph, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	graph := s.graphs[tenantID]
	if graph.TenantID == "" {
		graph = model.PredictiveGraph{TenantID: tenantID}
	}
	graph.Version = s.snapshot.GraphVersion
	return graph, nil
}

func (s *Store) PublishPredictiveSnap(_ context.Context, record model.PredictiveSnapRecord) (model.PredictiveGraph, model.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.graphs[record.TenantID]
	if current.TenantID == "" {
		current = model.PredictiveGraph{TenantID: record.TenantID}
	}
	current.Version = s.snapshot.GraphVersion
	if !reflect.DeepEqual(current, record.PreviousGraph) || s.snapshot.GraphVersion != record.PreviousGraph.Version {
		return model.PredictiveGraph{}, model.Snapshot{}, store.ErrSnapConflict
	}
	if s.snaps[record.TenantID] == nil {
		s.snaps[record.TenantID] = make(map[string]model.PredictiveSnapRecord)
	}
	if existing, ok := s.snaps[record.TenantID][record.ID]; ok {
		if !reflect.DeepEqual(existing, record) {
			return model.PredictiveGraph{}, model.Snapshot{}, store.ErrSnapConflict
		}
		return current, s.snapshot, nil
	}
	s.snapshot.RuntimeVersion++
	s.snapshot.GraphVersion++
	s.snapshot.AbstractionVersion++
	s.snapshot.PosteriorVersion++
	s.snapshot.ResidualVersion++
	graph := record.PublishedGraph
	graph.Version = s.snapshot.GraphVersion
	record.PublishedGraph = graph
	s.invalidateClosure(record.TenantID, record.Closure)
	s.graphs[record.TenantID] = graph
	s.snaps[record.TenantID][record.ID] = record
	return graph, s.snapshot, nil
}

func (s *Store) RollbackPredictiveSnap(_ context.Context, tenantID, snapID, reason string) (model.PredictiveGraph, model.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.snaps[tenantID][snapID]
	if !ok {
		return model.PredictiveGraph{}, model.Snapshot{}, store.ErrSnapNotFound
	}
	current := s.graphs[tenantID]
	if record.RolledBack || current.SourceSnapID != snapID {
		return model.PredictiveGraph{}, model.Snapshot{}, store.ErrSnapConflict
	}
	s.snapshot.RuntimeVersion++
	s.snapshot.GraphVersion++
	s.snapshot.AbstractionVersion++
	s.snapshot.PosteriorVersion++
	s.snapshot.ResidualVersion++
	graph := record.PreviousGraph
	graph.Version = s.snapshot.GraphVersion
	graph.PublishedAt = time.Now().UTC()
	graph.SourceSnapID = "rollback:" + snapID
	record.RolledBack = true
	record.RollbackReason = reason
	s.snaps[tenantID][snapID] = record
	s.invalidateClosure(tenantID, record.Closure)
	s.graphs[tenantID] = graph
	return graph, s.snapshot, nil
}

func (s *Store) invalidateClosure(tenantID string, closure model.DependencyClosure) {
	posteriorKeys, eventIDs := make(map[string]struct{}), make(map[string]struct{})
	for _, key := range closure.PosteriorKeys {
		posteriorKeys[key] = struct{}{}
		if posterior, ok := s.posteriors[tenantID][key]; ok {
			posterior.Certified = false
			s.posteriors[tenantID][key] = posterior
		}
	}
	for _, id := range closure.EventIDs {
		eventIDs[id] = struct{}{}
	}
	for key, record := range s.residuals[tenantID] {
		_, posteriorAffected := posteriorKeys[record.PosteriorKey]
		_, eventAffected := eventIDs[record.SourceEventID]
		if posteriorAffected || eventAffected {
			record.Active = false
			s.residuals[tenantID][key] = record
		}
	}
}

func (s *Store) PublishOmittedInfluenceCertificate(_ context.Context, certificate model.OmittedInfluenceCertificate) (model.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if certificate.PolicyVersion != s.snapshot.PolicyVersion || certificate.EvidenceEpoch != s.snapshot.EvidenceEpoch {
		return model.Snapshot{}, store.ErrStaleSnapshot
	}
	s.omittedCertificates[certificate.TenantID] = certificate
	s.snapshot.RuntimeVersion++
	return s.snapshot, nil
}

func (s *Store) GetOmittedInfluenceCertificate(_ context.Context, tenantID string) (model.OmittedInfluenceCertificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	certificate, ok := s.omittedCertificates[tenantID]
	if !ok {
		return model.OmittedInfluenceCertificate{}, store.ErrCertificateNotFound
	}
	return certificate, nil
}

func (s *Store) ApplyBayesianOutcome(_ context.Context, request model.BayesianOutcomeRequest, posteriorKey, digest string, weight float64, changePolicy bayes.ChangePolicy, residualObservation model.ResidualObservation, residualPolicy residual.Policy) (store.BayesianOutcomeResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	digests := s.outcomeDigests[request.TenantID]
	if digests == nil {
		digests = make(map[string]string)
		s.outcomeDigests[request.TenantID] = digests
	}
	if current, ok := digests[request.IdempotencyKey]; ok {
		if current != digest {
			return store.BayesianOutcomeResult{}, store.ErrOutcomeConflict
		}
		return store.BayesianOutcomeResult{Duplicate: true, Posterior: s.posteriors[request.TenantID][posteriorKey], Snapshot: s.snapshot}, nil
	}
	tenant := s.posteriors[request.TenantID]
	if tenant == nil {
		tenant = make(map[string]model.BayesianPosterior)
		s.posteriors[request.TenantID] = tenant
	}
	posterior := tenant[posteriorKey]
	if posterior.EvidenceEpoch != s.snapshot.EvidenceEpoch || posterior.Alpha <= 0 || posterior.Beta <= 0 {
		posterior = model.BayesianPosterior{TenantID: request.TenantID, PosteriorKey: posteriorKey, Alpha: 1, Beta: 1, EvidenceEpoch: s.snapshot.EvidenceEpoch, Certified: true}
	}
	posterior, changePoint := bayes.ApplyOutcome(posterior, request.Useful, weight, changePolicy)
	updateCalibration(&posterior, residualObservation.CommittedProbability, request.Useful, weight)
	posterior.UpdatedAt = request.AvailableAt
	if changePoint {
		s.invalidate()
		posterior.EvidenceEpoch = s.snapshot.EvidenceEpoch
	} else {
		s.snapshot.RuntimeVersion++
		s.snapshot.PosteriorVersion++
		s.snapshot.ResidualVersion++
	}
	tenant[posteriorKey] = posterior
	residualTenant := s.residuals[request.TenantID]
	if residualTenant == nil {
		residualTenant = make(map[string]model.ResidualRecord)
		s.residuals[request.TenantID] = residualTenant
	}
	exactID := residualRecordMapKey(model.ResidualExact, residualObservation.ActionKey)
	generalID := residualRecordMapKey(model.ResidualGeneral, residualObservation.GeneralKey)
	residualTenant[exactID] = residual.Update(residualTenant[exactID], residualObservation, model.ResidualExact, residualObservation.ActionKey, request.TenantID, weight, s.snapshot, residualPolicy)
	residualTenant[generalID] = residual.Update(residualTenant[generalID], residualObservation, model.ResidualGeneral, residualObservation.GeneralKey, request.TenantID, weight, s.snapshot, residualPolicy)
	digests[request.IdempotencyKey] = digest
	return store.BayesianOutcomeResult{ChangePoint: changePoint, Posterior: posterior, Snapshot: s.snapshot}, nil
}

func (s *Store) GetResidualCandidates(_ context.Context, tenantID, actionKey, generalKey string) (model.ResidualCandidates, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var candidates model.ResidualCandidates
	if record, ok := s.residuals[tenantID][residualRecordMapKey(model.ResidualExact, actionKey)]; ok {
		copy := record
		candidates.Exact = &copy
	}
	if record, ok := s.residuals[tenantID][residualRecordMapKey(model.ResidualGeneral, generalKey)]; ok {
		copy := record
		candidates.General = &copy
	}
	return candidates, nil
}

func residualRecordMapKey(scope model.ResidualScope, key string) string {
	return string(scope) + "\x00" + key
}

func updateCalibration(posterior *model.BayesianPosterior, probability float64, useful bool, weight float64) {
	target := 0.0
	if useful {
		target = 1
	}
	posterior.CalibrationWeight += weight
	posterior.BrierLossSum += weight * (target - probability) * (target - probability)
	posterior.ForecastUsefulSum += weight * probability
	posterior.ObservedUsefulSum += weight * target
}

func (s *Store) GetBayesianPosterior(_ context.Context, tenantID, posteriorKey string) (model.BayesianPosterior, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	posterior, ok := s.posteriors[tenantID][posteriorKey]
	if !ok {
		return model.BayesianPosterior{}, store.ErrPosteriorNotFound
	}
	return posterior, nil
}

func (s *Store) PublishSelectionCertificate(_ context.Context, certificate model.SelectionSupportCertificate) (model.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if certificate.PolicyVersion != s.snapshot.PolicyVersion || certificate.EvidenceEpoch != s.snapshot.EvidenceEpoch {
		return model.Snapshot{}, store.ErrStaleSnapshot
	}
	s.selectionCertificates[certificate.TenantID] = certificate
	s.snapshot.RuntimeVersion++
	return s.snapshot, nil
}

func (s *Store) GetSelectionCertificate(_ context.Context, tenantID string) (model.SelectionSupportCertificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	certificate, ok := s.selectionCertificates[tenantID]
	if !ok {
		return model.SelectionSupportCertificate{}, store.ErrCertificateNotFound
	}
	return certificate, nil
}

func (s *Store) PublishAntiPigeonCertificate(_ context.Context, certificate model.AntiPigeonCertificate) (model.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if certificate.GraphVersion != s.snapshot.GraphVersion || certificate.EvidenceEpoch != s.snapshot.EvidenceEpoch {
		return model.Snapshot{}, store.ErrStaleSnapshot
	}
	s.antiPigeonCertificates[certificate.ID] = certificate
	index := s.antiPigeonIndex[certificate.TenantID]
	if index == nil {
		index = make(map[string]string)
		s.antiPigeonIndex[certificate.TenantID] = index
	}
	for _, eventID := range certificate.MemberEventIDs {
		index[eventID] = certificate.ID
	}
	s.snapshot.RuntimeVersion++
	s.snapshot.AbstractionVersion++
	return s.snapshot, nil
}

func (s *Store) GetAntiPigeonCertificate(_ context.Context, tenantID string, eventIDs []string) (model.AntiPigeonCertificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var certificateID string
	for _, eventID := range eventIDs {
		current := s.antiPigeonIndex[tenantID][eventID]
		if current == "" || (certificateID != "" && current != certificateID) {
			return model.AntiPigeonCertificate{}, store.ErrCertificateNotFound
		}
		certificateID = current
	}
	certificate, ok := s.antiPigeonCertificates[certificateID]
	if !ok {
		return model.AntiPigeonCertificate{}, store.ErrCertificateNotFound
	}
	return certificate, nil
}

func (s *Store) PutBayesianJournal(_ context.Context, entry model.BayesianJournalEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tenant := s.journals[entry.TenantID]
	if tenant == nil {
		tenant = make(map[string]model.BayesianJournalEntry)
		s.journals[entry.TenantID] = tenant
	}
	if current, ok := tenant[entry.ID]; ok {
		if !journalEqual(current, entry) {
			return store.ErrJournalConflict
		}
		return nil
	}
	if entry.Snapshot != s.snapshot {
		return store.ErrStaleSnapshot
	}
	tenant[entry.ID] = entry
	return nil
}

func (s *Store) GetBayesianJournal(_ context.Context, tenantID, journalID string) (model.BayesianJournalEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.journals[tenantID][journalID]
	if !ok {
		return model.BayesianJournalEntry{}, store.ErrJournalNotFound
	}
	return entry, nil
}

func (s *Store) Put(_ context.Context, event model.Event, vector []float32, digest string) (store.PutResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tenant := s.entries[event.TenantID]
	if tenant == nil {
		tenant = make(map[string]entry)
		s.entries[event.TenantID] = tenant
	}
	if current, ok := tenant[event.ID]; ok {
		if current.digest != digest {
			return store.PutResult{}, store.ErrIdempotencyConflict
		}
		return store.PutResult{Duplicate: true, Snapshot: s.snapshot}, nil
	}
	tenant[event.ID] = entry{event: event, vector: append([]float32(nil), vector...), digest: digest}
	s.snapshot.RuntimeVersion++
	s.snapshot.EvidenceEpoch++
	return store.PutResult{Snapshot: s.snapshot}, nil
}

func (s *Store) Search(_ context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]store.SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	results := make([]store.SearchResult, 0, len(s.entries[tenantID]))
	for _, candidate := range s.entries[tenantID] {
		if candidate.event.AvailableAt.After(availableBy) {
			continue
		}
		results = append(results, store.SearchResult{Event: candidate.event, Similarity: cosine(vector, candidate.vector)})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Similarity > results[j].Similarity })
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func (s *Store) Stats(_ context.Context) (store.Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stats := store.Stats{Backend: "memory", Tenants: len(s.entries)}
	for _, tenant := range s.entries {
		stats.Events += len(tenant)
	}
	return stats, nil
}

func (s *Store) Delete(_ context.Context, tenantID, eventID string) (store.DeleteResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.entries[tenantID][eventID]; !ok {
		return store.DeleteResult{Snapshot: s.snapshot}, nil
	}
	delete(s.entries[tenantID], eventID)
	s.invalidate()
	return store.DeleteResult{Deleted: true, Snapshot: s.snapshot}, nil
}

func (s *Store) DeleteBefore(_ context.Context, tenantID string, before time.Time, limit int) (store.RetentionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0)
	for id, item := range s.entries[tenantID] {
		if item.event.AvailableAt.Before(before) && len(ids) < limit {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		delete(s.entries[tenantID], id)
	}
	if len(ids) > 0 {
		s.invalidate()
	}
	return store.RetentionResult{DeletedIDs: ids, Snapshot: s.snapshot}, nil
}

func (s *Store) Backup(context.Context, string) error {
	return errors.New("memory store does not support backup")
}
func (s *Store) Compact(context.Context) error { return nil }

func (s *Store) Close() error { return nil }

func (s *Store) Snapshot(_ context.Context) model.Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot
}

func initialSnapshot() model.Snapshot {
	return model.Snapshot{PolicyVersion: 1, ContractVersion: model.ContractVersion, GraphVersion: 1, PosteriorVersion: 1, ResidualVersion: 1, AbstractionVersion: 1}
}

func (s *Store) invalidate() {
	s.snapshot.RuntimeVersion++
	s.snapshot.EvidenceEpoch++
	s.snapshot.GraphVersion++
	s.snapshot.PosteriorVersion++
	s.snapshot.ResidualVersion++
	s.snapshot.AbstractionVersion++
}

func cosine(left, right []float32) float64 {
	if len(left) != len(right) || len(left) == 0 {
		return 0
	}
	var dot, leftNorm, rightNorm float64
	for index := range left {
		l, r := float64(left[index]), float64(right[index])
		dot += l * r
		leftNorm += l * l
		rightNorm += r * r
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return dot / math.Sqrt(leftNorm*rightNorm)
}

func journalEqual(left, right model.BayesianJournalEntry) bool {
	return reflect.DeepEqual(left, right)
}
