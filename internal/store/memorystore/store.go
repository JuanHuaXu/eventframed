package memorystore

import (
	"context"
	"errors"
	"fmt"
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
	outcomeResults         map[string]map[string]store.BayesianOutcomeResult
	posteriors             map[string]map[string]model.BayesianPosterior
	residuals              map[string]map[string]model.ResidualRecord
	graphs                 map[string]model.PredictiveGraph
	snaps                  map[string]map[string]model.PredictiveSnapRecord
	agencyRecords          map[string]map[string]model.AgencyProposalRecord
	agencyDigests          map[string]map[string]string
	compositionTombstones  map[string]map[string]model.CompositionTombstone
	policyDigest           string
	snapshot               model.Snapshot
	ingestMotion           map[uint64]time.Time
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
		outcomeDigests: make(map[string]map[string]string), outcomeResults: make(map[string]map[string]store.BayesianOutcomeResult), posteriors: make(map[string]map[string]model.BayesianPosterior),
		residuals: make(map[string]map[string]model.ResidualRecord),
		graphs:    make(map[string]model.PredictiveGraph), snaps: make(map[string]map[string]model.PredictiveSnapRecord),
		agencyRecords: make(map[string]map[string]model.AgencyProposalRecord), agencyDigests: make(map[string]map[string]string),
		compositionTombstones: make(map[string]map[string]model.CompositionTombstone),
		snapshot:              initialSnapshot(),
		ingestMotion:          make(map[uint64]time.Time),
	}
}

func (s *Store) PutAgencyProposal(_ context.Context, record model.AgencyProposalRecord, digest string, maxPerChain, maxPending int, evidenceAvailableBy time.Time) (store.AgencyPutResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	records := s.agencyRecords[record.Proposal.TenantID]
	if records == nil {
		records = make(map[string]model.AgencyProposalRecord)
		s.agencyRecords[record.Proposal.TenantID] = records
		s.agencyDigests[record.Proposal.TenantID] = make(map[string]string)
	}
	if existing, ok := records[record.Proposal.ID]; ok {
		if s.agencyDigests[record.Proposal.TenantID][record.Proposal.ID] != digest {
			return store.AgencyPutResult{}, store.ErrAgencyConflict
		}
		return store.AgencyPutResult{Duplicate: true, Record: existing, Snapshot: s.snapshot}, nil
	}
	for _, id := range record.Proposal.EvidenceIDs {
		entry, ok := s.entries[record.Proposal.TenantID][id]
		if !ok || entry.event.AvailableAt.After(evidenceAvailableBy) {
			return store.AgencyPutResult{}, store.ErrAgencyEvidence
		}
	}
	if !validAgencyParent(records, record.Proposal) || countAgencyChain(records, record.Proposal.CausalChainID) >= maxPerChain || countPendingAgency(records, evidenceAvailableBy) >= maxPending {
		return store.AgencyPutResult{}, store.ErrAgencyChainBudget
	}
	records[record.Proposal.ID] = record
	s.agencyDigests[record.Proposal.TenantID][record.Proposal.ID] = digest
	s.snapshot.RuntimeVersion++
	s.snapshot.AgencyVersion++
	return store.AgencyPutResult{Record: record, Snapshot: s.snapshot}, nil
}

func (s *Store) ClaimAgencyProposals(_ context.Context, tenantID, consumerID string, now time.Time, limit int, lease time.Duration) ([]model.AgencyProposalRecord, model.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	records := s.agencyRecords[tenantID]
	eligible := make([]model.AgencyProposalRecord, 0)
	changed := false
	for id, record := range records {
		if (record.Status == model.AgencyPending || record.Status == model.AgencyClaimed) && !now.Before(record.Proposal.ExpiresAt) {
			record.Status, record.ClaimedBy, record.LeaseUntil = model.AgencyExpired, "", time.Time{}
			record.ResolutionReason, record.ResolvedAt = "proposal expired before authorization", now
			records[id], changed = record, true
			continue
		}
		if now.Before(record.Proposal.NotBefore) || record.Status != model.AgencyPending && (record.Status != model.AgencyClaimed || now.Before(record.LeaseUntil)) {
			continue
		}
		eligible = append(eligible, record)
	}
	sort.Slice(eligible, func(i, j int) bool {
		if eligible[i].Proposal.Priority != eligible[j].Proposal.Priority {
			return eligible[i].Proposal.Priority > eligible[j].Proposal.Priority
		}
		if !eligible[i].Proposal.NotBefore.Equal(eligible[j].Proposal.NotBefore) {
			return eligible[i].Proposal.NotBefore.Before(eligible[j].Proposal.NotBefore)
		}
		return eligible[i].Proposal.ID < eligible[j].Proposal.ID
	})
	if len(eligible) > limit {
		eligible = eligible[:limit]
	}
	for index := range eligible {
		record := eligible[index]
		record.Status, record.ClaimedBy, record.LeaseUntil = model.AgencyClaimed, consumerID, now.Add(lease)
		records[record.Proposal.ID], eligible[index], changed = record, record, true
	}
	if changed {
		s.snapshot.RuntimeVersion++
		s.snapshot.AgencyVersion++
	}
	return eligible, s.snapshot, nil
}

func (s *Store) ResolveAgencyProposal(_ context.Context, request model.ResolveAgencyProposalRequest, now time.Time) (store.AgencyResolveResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.agencyRecords[request.TenantID][request.ProposalID]
	if !ok {
		return store.AgencyResolveResult{}, store.ErrAgencyNotFound
	}
	if record.Status == model.AgencyApproved || record.Status == model.AgencyRejected || record.Status == model.AgencyExpired {
		if record.Status == request.Decision && record.ResolutionReason == request.Reason && record.ExecutionRef == request.ExecutionRef {
			return store.AgencyResolveResult{Duplicate: true, Record: record, Snapshot: s.snapshot}, nil
		}
		return store.AgencyResolveResult{}, store.ErrAgencyConflict
	}
	if record.Status != model.AgencyClaimed || record.ClaimedBy != request.ConsumerID {
		return store.AgencyResolveResult{}, store.ErrAgencyLease
	}
	if !now.Before(record.Proposal.ExpiresAt) {
		record.Status, record.ResolutionReason, record.ResolvedAt = model.AgencyExpired, "proposal expired before authorization completed", now
		record.ClaimedBy, record.LeaseUntil, record.ExecutionRef = "", time.Time{}, ""
		s.agencyRecords[request.TenantID][request.ProposalID] = record
		s.snapshot.RuntimeVersion++
		s.snapshot.AgencyVersion++
		return store.AgencyResolveResult{Record: record, Snapshot: s.snapshot}, store.ErrAgencyExpired
	}
	if !now.Before(record.LeaseUntil) {
		return store.AgencyResolveResult{}, store.ErrAgencyLease
	}
	record.Status, record.ResolutionReason, record.ExecutionRef, record.ResolvedAt = request.Decision, request.Reason, request.ExecutionRef, now
	record.ClaimedBy, record.LeaseUntil = "", time.Time{}
	s.agencyRecords[request.TenantID][request.ProposalID] = record
	s.snapshot.RuntimeVersion++
	s.snapshot.AgencyVersion++
	return store.AgencyResolveResult{Record: record, Snapshot: s.snapshot}, nil
}

func validAgencyParent(records map[string]model.AgencyProposalRecord, proposal model.AgencyProposal) bool {
	if proposal.CausalChainDepth == 0 {
		return proposal.ParentProposalID == ""
	}
	parent, ok := records[proposal.ParentProposalID]
	return ok && parent.Proposal.CausalChainID == proposal.CausalChainID && parent.Proposal.CausalChainDepth+1 == proposal.CausalChainDepth
}

func countAgencyChain(records map[string]model.AgencyProposalRecord, chainID string) int {
	count := 0
	for _, record := range records {
		if record.Proposal.CausalChainID == chainID {
			count++
		}
	}
	return count
}

func countPendingAgency(records map[string]model.AgencyProposalRecord, now time.Time) int {
	count := 0
	for _, record := range records {
		if (record.Status == model.AgencyPending || record.Status == model.AgencyClaimed) && record.Proposal.ExpiresAt.After(now) {
			count++
		}
	}
	return count
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

func (s *Store) ApplyBayesianOutcome(_ context.Context, request model.BayesianOutcomeRequest, posteriorKey, parentPosteriorKey, digest string, weight float64, changePolicy bayes.ChangePolicy, groupPolicy bayes.GroupPolicy, residualObservation model.ResidualObservation, residualPolicy residual.Policy) (store.BayesianOutcomeResult, error) {
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
		result := s.outcomeResults[request.TenantID][request.IdempotencyKey]
		result.Duplicate, result.Snapshot = true, s.snapshot
		result.Posterior = copyWorkingBelief(result.Posterior)
		return result, nil
	}
	tenant := s.posteriors[request.TenantID]
	if tenant == nil {
		tenant = make(map[string]model.BayesianPosterior)
		s.posteriors[request.TenantID] = tenant
	}
	posterior := tenant[posteriorKey]
	if posterior.EvidenceEpoch != s.snapshot.EvidenceEpoch || posterior.Alpha <= 0 || posterior.Beta <= 0 || (changePolicy.EvidenceTrust != "" && posterior.EvidenceTrust != changePolicy.EvidenceTrust) {
		posterior = model.BayesianPosterior{TenantID: request.TenantID, PosteriorKey: posteriorKey, Alpha: 1, Beta: 1, EvidenceEpoch: s.snapshot.EvidenceEpoch, Certified: true}
	}
	validationEligible := request.Source == model.OutcomeFullStream || request.Source == model.OutcomeIndependentAudit
	resetAuthorized := validationEligible || len(posteriorKey) < 3 || posteriorKey[:3] != "ap:"
	certificate, memberIDs := s.antiPigeonGroup(request.TenantID, posteriorKey)
	pooledWeight := bayes.SharedOutcomeWeight(groupPolicy, len(memberIDs) >= 2, weight)
	posterior, changePoint := bayes.ApplyOutcomeAuthorized(posterior, request.Useful, pooledWeight, changePolicy, resetAuthorized)
	bayes.UpdateMemberEvidence(&posterior, request.EventID, request.Useful, weight)
	updateCalibration(&posterior, residualObservation.CommittedProbability, request.Useful, weight)
	posterior.UpdatedAt = request.AvailableAt
	revision := bayes.AssessRevision(posterior, memberIDs, request.EventID, changePoint, validationEligible, groupPolicy)
	if changePoint {
		s.invalidate()
		posterior.EvidenceEpoch = s.snapshot.EvidenceEpoch
		if !bayes.RevisionSplits(revision.Action) {
			posterior.MemberEvidence = nil
		}
	} else if bayes.RevisionSplits(revision.Action) {
		s.snapshot.RuntimeVersion++
		s.snapshot.AbstractionVersion++
		s.snapshot.PosteriorVersion++
		s.snapshot.ResidualVersion++
	} else {
		s.snapshot.RuntimeVersion++
		s.snapshot.PosteriorVersion++
		s.snapshot.ResidualVersion++
	}
	tenant[posteriorKey] = posterior
	resultPosterior := posterior
	if bayes.RevisionSplits(revision.Action) {
		posterior.Certified = false
		tenant[posteriorKey] = posterior
		s.revokeAntiPigeon(certificate)
		for _, eventID := range memberIDs {
			evidence := posterior.MemberEvidence[eventID]
			child := materializedMemberPosterior(request.TenantID, eventID, evidence, s.snapshot.EvidenceEpoch, request.AvailableAt)
			child.EvidenceTrust = changePolicy.EvidenceTrust
			if changePoint && eventID == request.EventID {
				child = resetMemberPosterior(child, request.Useful, weight)
				child.WorkingBelief = bayes.UpdateWorking(nil, request.Useful, weight, true, changePolicy.Working)
			}
			tenant[eventID] = child
			if eventID == request.EventID {
				resultPosterior = child
			}
		}
	}
	if parentPosteriorKey != "" && parentPosteriorKey != posteriorKey {
		parent := tenant[parentPosteriorKey]
		if parent.EvidenceEpoch != s.snapshot.EvidenceEpoch || parent.Alpha <= 0 || parent.Beta <= 0 || (changePolicy.EvidenceTrust != "" && parent.EvidenceTrust != changePolicy.EvidenceTrust) {
			parent = model.BayesianPosterior{TenantID: request.TenantID, PosteriorKey: parentPosteriorKey, Alpha: 1, Beta: 1, EvidenceEpoch: s.snapshot.EvidenceEpoch, Certified: true}
		}
		if request.Useful {
			parent.Alpha += weight
		} else {
			parent.Beta += weight
		}
		parent.EffectiveSupport += weight
		parent.EvidenceTrust = changePolicy.EvidenceTrust
		parent.UpdatedAt = request.AvailableAt
		tenant[parentPosteriorKey] = parent
	}
	residualTenant := s.residuals[request.TenantID]
	if residualTenant == nil {
		residualTenant = make(map[string]model.ResidualRecord)
		s.residuals[request.TenantID] = residualTenant
	}
	exactID := residualRecordMapKey(model.ResidualExact, residualObservation.ActionKey)
	generalID := residualRecordMapKey(model.ResidualGeneral, residualObservation.GeneralKey)
	residualTenant[exactID] = residual.Update(residualTenant[exactID], residualObservation, model.ResidualExact, residualObservation.ActionKey, request.TenantID, weight, s.snapshot, residualPolicy)
	residualTenant[generalID] = residual.Update(residualTenant[generalID], residualObservation, model.ResidualGeneral, residualObservation.GeneralKey, request.TenantID, weight, s.snapshot, residualPolicy)
	if bayes.RevisionSplits(revision.Action) {
		exact, general := residualTenant[exactID], residualTenant[generalID]
		exact.Active, general.Active = false, false
		residualTenant[exactID], residualTenant[generalID] = exact, general
	}
	digests[request.IdempotencyKey] = digest
	if s.outcomeResults[request.TenantID] == nil {
		s.outcomeResults[request.TenantID] = make(map[string]store.BayesianOutcomeResult)
	}
	result := store.BayesianOutcomeResult{ChangePoint: changePoint, Revision: revision, Posterior: resultPosterior, Snapshot: s.snapshot}
	s.outcomeResults[request.TenantID][request.IdempotencyKey] = result
	result.Posterior = copyWorkingBelief(result.Posterior)
	return result, nil
}

func (s *Store) antiPigeonGroup(tenantID, posteriorKey string) (model.AntiPigeonCertificate, []string) {
	if len(posteriorKey) <= 3 || posteriorKey[:3] != "ap:" {
		return model.AntiPigeonCertificate{}, nil
	}
	certificate := s.antiPigeonCertificates[posteriorKey[3:]]
	if certificate.TenantID != tenantID {
		return model.AntiPigeonCertificate{}, nil
	}
	for _, eventID := range certificate.MemberEventIDs {
		if s.antiPigeonIndex[tenantID][eventID] != certificate.ID {
			return model.AntiPigeonCertificate{}, nil
		}
	}
	return certificate, append([]string(nil), certificate.MemberEventIDs...)
}

func (s *Store) revokeAntiPigeon(certificate model.AntiPigeonCertificate) {
	if certificate.ID == "" {
		return
	}
	delete(s.antiPigeonCertificates, certificate.ID)
	for _, eventID := range certificate.MemberEventIDs {
		if s.antiPigeonIndex[certificate.TenantID][eventID] == certificate.ID {
			delete(s.antiPigeonIndex[certificate.TenantID], eventID)
		}
	}
}

func materializedMemberPosterior(tenantID, eventID string, evidence model.BayesianMemberEvidence, epoch uint64, updatedAt time.Time) model.BayesianPosterior {
	return model.BayesianPosterior{
		TenantID: tenantID, PosteriorKey: eventID, Alpha: 1 + evidence.UsefulWeight, Beta: 1 + evidence.NotUsefulWeight,
		EffectiveSupport: evidence.UsefulWeight + evidence.NotUsefulWeight, EvidenceEpoch: epoch, Certified: true, UpdatedAt: updatedAt,
		MemberEvidence: map[string]model.BayesianMemberEvidence{eventID: evidence},
	}
}

func resetMemberPosterior(posterior model.BayesianPosterior, useful bool, weight float64) model.BayesianPosterior {
	posterior.Alpha, posterior.Beta, posterior.EffectiveSupport = 1, 1, 0
	posterior.MemberEvidence = make(map[string]model.BayesianMemberEvidence)
	if useful {
		posterior.Alpha += weight
	} else {
		posterior.Beta += weight
	}
	posterior.EffectiveSupport = weight
	bayes.UpdateMemberEvidence(&posterior, posterior.PosteriorKey, useful, weight)
	return posterior
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
	return copyWorkingBelief(posterior), nil
}

// Keep the externally returned working state from aliasing durable memory.
func copyWorkingBelief(p model.BayesianPosterior) model.BayesianPosterior {
	if p.WorkingBelief != nil {
		state := *p.WorkingBelief
		p.WorkingBelief = &state
	}
	return p
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
	if !store.JournalSnapshotCompatible(entry.Snapshot, s.snapshot, entry.AsOf, s.ingestMotion) {
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
	recordIngestMotion(s.ingestMotion, s.snapshot.RuntimeVersion, event.AvailableAt)
	return store.PutResult{Snapshot: s.snapshot}, nil
}

func (s *Store) PutComposition(_ context.Context, event model.Event, vector []float32, digest string, base model.Snapshot) (store.PutResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.snapshot != base {
		return store.PutResult{}, store.ErrStaleSnapshot
	}
	if event.Composition == nil {
		return store.PutResult{}, store.ErrCompositionAuthority
	}
	certificate, members := s.antiPigeonGroup(event.TenantID, "ap:"+event.Composition.AntiPigeonCertificateID)
	if certificate.ID != event.Composition.AntiPigeonCertificateID || !sameIDSet(members, event.Composition.MemberEventIDs) {
		return store.PutResult{}, store.ErrCompositionAuthority
	}
	tenant := s.entries[event.TenantID]
	if tenant == nil {
		return store.PutResult{}, store.ErrCompositionAuthority
	}
	for _, memberID := range event.Composition.MemberEventIDs {
		member, ok := tenant[memberID]
		if !ok || member.event.AvailableAt.After(event.AvailableAt) {
			return store.PutResult{}, store.ErrCompositionAuthority
		}
	}
	if current, ok := tenant[event.ID]; ok {
		if current.digest != digest {
			return store.PutResult{}, store.ErrIdempotencyConflict
		}
		return store.PutResult{Duplicate: true, Snapshot: s.snapshot}, nil
	}
	tenant[event.ID] = entry{event: event, vector: append([]float32(nil), vector...), digest: digest}
	s.snapshot.RuntimeVersion++
	s.snapshot.AbstractionVersion++
	recordIngestMotion(s.ingestMotion, s.snapshot.RuntimeVersion, event.AvailableAt)
	return store.PutResult{Snapshot: s.snapshot}, nil
}

func recordIngestMotion(motion map[uint64]time.Time, version uint64, availableAt time.Time) {
	motion[version] = availableAt
	const retainedVersions = uint64(4096)
	if version > retainedVersions {
		delete(motion, version-retainedVersions)
	}
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

func (s *Store) GetEvents(_ context.Context, tenantID string, eventIDs []string, availableBy time.Time) ([]model.Event, error) {
	return s.getEvents(tenantID, eventIDs, availableBy, false)
}

func (s *Store) GetEventsWithVectors(_ context.Context, tenantID string, eventIDs []string, availableBy time.Time) ([]model.Event, error) {
	return s.getEvents(tenantID, eventIDs, availableBy, true)
}

func (s *Store) getEvents(tenantID string, eventIDs []string, availableBy time.Time, hydrateVectors bool) ([]model.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tenant := s.entries[tenantID]
	results := make([]model.Event, 0, len(eventIDs))
	for _, eventID := range eventIDs {
		item, ok := tenant[eventID]
		if !ok {
			return nil, fmt.Errorf("%w: id=%s absent", store.ErrEventNotFound, eventID)
		}
		if item.event.AvailableAt.After(availableBy) {
			return nil, fmt.Errorf("%w: id=%s is not available as of request", store.ErrEventNotFound, eventID)
		}
		event := item.event
		if hydrateVectors {
			event.Embedding = append([]float32(nil), item.vector...)
		} else {
			event.Embedding = nil
		}
		results = append(results, event)
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
	agencyChanged := s.cancelAgencyByEvidence(tenantID, []string{eventID}, time.Now().UTC())
	s.invalidate()
	if agencyChanged {
		s.snapshot.AgencyVersion++
	}
	return store.DeleteResult{Deleted: true, Snapshot: s.snapshot}, nil
}

func (s *Store) DeleteComposition(_ context.Context, tenantID, eventID, reason string, decomposedAt time.Time) (store.CompositionDeleteResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.entries[tenantID][eventID]
	if !ok {
		return store.CompositionDeleteResult{Snapshot: s.snapshot}, nil
	}
	if item.event.Composition == nil {
		return store.CompositionDeleteResult{}, store.ErrCompositionAuthority
	}
	members := append([]string(nil), item.event.Composition.MemberEventIDs...)
	tombstones := s.compositionTombstones[tenantID]
	if tombstones == nil {
		tombstones = make(map[string]model.CompositionTombstone)
		s.compositionTombstones[tenantID] = tombstones
	}
	tombstones[eventID] = model.CompositionTombstone{TenantID: tenantID, EventID: eventID, MemberEventIDs: members, AntiPigeonCertificateID: item.event.Composition.AntiPigeonCertificateID, Reason: reason, DecomposedAt: decomposedAt.UTC()}
	delete(s.entries[tenantID], eventID)
	s.snapshot.RuntimeVersion++
	s.snapshot.AbstractionVersion++
	return store.CompositionDeleteResult{Deleted: true, MemberEventIDs: members, Snapshot: s.snapshot}, nil
}

func (s *Store) GetCompositionTombstone(_ context.Context, tenantID, eventID string) (model.CompositionTombstone, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tombstone, ok := s.compositionTombstones[tenantID][eventID]
	if !ok {
		return model.CompositionTombstone{}, store.ErrEventNotFound
	}
	return tombstone, nil
}

func sameIDSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[string]struct{}, len(left))
	for _, id := range left {
		seen[id] = struct{}{}
	}
	for _, id := range right {
		if _, ok := seen[id]; !ok {
			return false
		}
	}
	return true
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
		agencyChanged := s.cancelAgencyByEvidence(tenantID, ids, time.Now().UTC())
		s.invalidate()
		if agencyChanged {
			s.snapshot.AgencyVersion++
		}
	}
	return store.RetentionResult{DeletedIDs: ids, Snapshot: s.snapshot}, nil
}

func (s *Store) cancelAgencyByEvidence(tenantID string, eventIDs []string, now time.Time) bool {
	deleted := make(map[string]struct{}, len(eventIDs))
	for _, id := range eventIDs {
		deleted[id] = struct{}{}
	}
	changed := false
	for id, record := range s.agencyRecords[tenantID] {
		if record.Status != model.AgencyPending && record.Status != model.AgencyClaimed {
			continue
		}
		affected := false
		for _, evidenceID := range record.Proposal.EvidenceIDs {
			if _, ok := deleted[evidenceID]; ok {
				affected = true
				break
			}
		}
		if !affected {
			continue
		}
		record.Status, record.ClaimedBy, record.LeaseUntil = model.AgencyRejected, "", time.Time{}
		record.ResolutionReason, record.ResolvedAt = "supporting evidence was deleted", now
		s.agencyRecords[tenantID][id], changed = record, true
	}
	return changed
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
	return model.Snapshot{PolicyVersion: 1, ContractVersion: model.ContractVersion, GraphVersion: 1, PosteriorVersion: 1, ResidualVersion: 1, AbstractionVersion: 1, AgencyVersion: 1}
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
