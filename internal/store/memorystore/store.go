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
		snapshot: initialSnapshot(),
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

func (s *Store) ApplyBayesianOutcome(_ context.Context, request model.BayesianOutcomeRequest, posteriorKey, digest string, weight float64, changePolicy bayes.ChangePolicy) (store.BayesianOutcomeResult, error) {
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
	posterior.UpdatedAt = request.AvailableAt
	if changePoint {
		s.invalidate()
		posterior.EvidenceEpoch = s.snapshot.EvidenceEpoch
	} else {
		s.snapshot.RuntimeVersion++
		s.snapshot.PosteriorVersion++
	}
	tenant[posteriorKey] = posterior
	digests[request.IdempotencyKey] = digest
	return store.BayesianOutcomeResult{ChangePoint: changePoint, Posterior: posterior, Snapshot: s.snapshot}, nil
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
