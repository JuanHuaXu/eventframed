package libravdbstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/store"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

type Config struct {
	Path           string
	Dimension      int
	Quantization   string
	MemoryMapping  bool
	EmbeddingModel string
}

type Store struct {
	db           *libra.Database
	config       Config
	mu           sync.Mutex
	writeMu      sync.Mutex
	collections  map[string]*libra.Collection
	bayesian     *libra.Collection
	policyDigest string
	snapshot     model.Snapshot
}

const systemCollection = "_eventframe_system"
const bayesianCollection = "_eventframe_bayesian"

type persistentState struct {
	SchemaVersion        uint64         `json:"schema_version"`
	EmbeddingModel       string         `json:"embedding_model"`
	Dimension            int            `json:"dimension"`
	Quantization         string         `json:"quantization"`
	Snapshot             model.Snapshot `json:"snapshot"`
	BayesianPolicyDigest string         `json:"bayesian_policy_digest,omitempty"`
}

func Open(config Config) (*Store, error) {
	if config.Path == "" {
		return nil, errors.New("database path is required")
	}
	if config.Dimension <= 0 {
		return nil, errors.New("dimension must be positive")
	}
	if config.EmbeddingModel == "" {
		return nil, errors.New("embedding model key is required")
	}
	db, err := libra.Open(
		libra.WithStoragePath(config.Path),
		libra.WithMaxConcurrentWrites(2),
		libra.WithMaxWriteQueueDepth(64),
	)
	if err != nil {
		return nil, fmt.Errorf("open libravdb: %w", err)
	}
	s := &Store{db: db, config: config, collections: make(map[string]*libra.Collection)}
	if err := s.loadOrInitializeState(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Put(ctx context.Context, event model.Event, vector []float32, digest string) (store.PutResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	collectionKey := collectionName(event.TenantID, s.config.EmbeddingModel)
	collection, err := s.collection(ctx, event.TenantID)
	if err != nil {
		return store.PutResult{}, err
	}
	record, err := collection.Get(ctx, event.ID)
	if err == nil {
		if current, _ := record.Metadata["content_digest"].(string); current != digest {
			return store.PutResult{}, store.ErrIdempotencyConflict
		}
		return store.PutResult{Duplicate: true, Snapshot: s.snapshot}, nil
	}
	if !errors.Is(err, libra.ErrRecordNotFound) {
		return store.PutResult{}, fmt.Errorf("check existing event: %w", err)
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return store.PutResult{}, fmt.Errorf("encode event: %w", err)
	}
	metadata := map[string]interface{}{
		"event_json":     string(payload),
		"content_digest": digest,
		"session_id":     event.SessionID,
		"kind":           event.Kind,
		"available_at":   event.AvailableAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		"priority":       event.Priority,
	}
	next := s.snapshot
	next.RuntimeVersion++
	next.EvidenceEpoch++
	stateMetadata, err := s.stateMetadata(next)
	if err != nil {
		return store.PutResult{}, err
	}
	err = s.db.WithTx(ctx, func(tx libra.Tx) error {
		if err := tx.Insert(ctx, collectionKey, event.ID, vector, metadata); err != nil {
			return err
		}
		return tx.Upsert(ctx, systemCollection, "runtime", nil, stateMetadata)
	})
	if err != nil {
		// Resolve the timeout/racing-insert ambiguity through the durable record.
		if current, getErr := collection.Get(ctx, event.ID); getErr == nil {
			if currentDigest, _ := current.Metadata["content_digest"].(string); currentDigest == digest {
				return store.PutResult{Duplicate: true, Snapshot: s.snapshot}, nil
			}
			return store.PutResult{}, store.ErrIdempotencyConflict
		}
		return store.PutResult{}, fmt.Errorf("insert event: %w", err)
	}
	s.snapshot = next
	return store.PutResult{Snapshot: next}, nil
}

func (s *Store) BindBayesianPolicy(ctx context.Context, digest string) (model.Snapshot, error) {
	if strings.TrimSpace(digest) == "" {
		return model.Snapshot{}, errors.New("Bayesian policy digest is required")
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if digest == s.policyDigest {
		return s.snapshot, nil
	}
	next := s.snapshot
	next.RuntimeVersion++
	next.PolicyVersion++
	metadata, err := s.stateMetadataWithPolicy(next, digest)
	if err != nil {
		return model.Snapshot{}, err
	}
	if err := s.db.WithTx(ctx, func(tx libra.Tx) error {
		return tx.Upsert(ctx, systemCollection, "runtime", nil, metadata)
	}); err != nil {
		return model.Snapshot{}, fmt.Errorf("bind Bayesian policy: %w", err)
	}
	s.policyDigest = digest
	s.snapshot = next
	return next, nil
}

func (s *Store) Search(ctx context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]store.SearchResult, error) {
	collection, err := s.collection(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	// LibraVDB's metadata filter is applied after ANN limiting in this release.
	// Expand deterministically until enough availability-eligible candidates are
	// found, or every live record has been examined. This prevents a dense set
	// of future records from crowding valid history out before reranking.
	live := collection.Stats(ctx).LiveRecordCount
	probe := min(max(limit, 1), max(live, 1))
	for {
		results, searchErr := collection.Search(ctx, vector, probe)
		if searchErr != nil {
			return nil, fmt.Errorf("search events: %w", searchErr)
		}
		out := decodeEligible(results.Results, availableBy, limit)
		if len(out) >= limit || probe >= live {
			return out, nil
		}
		probe = min(live, probe*2)
	}
}

func (s *Store) PutBayesianJournal(ctx context.Context, entry model.BayesianJournalEntry) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	encoded, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("encode Bayesian journal: %w", err)
	}
	recordID := bayesianJournalRecordID(entry.TenantID, entry.ID)
	record, err := s.bayesian.Get(ctx, recordID)
	if err == nil {
		current, _ := record.Metadata["journal_json"].(string)
		if current != string(encoded) {
			return store.ErrJournalConflict
		}
		return nil
	}
	if !errors.Is(err, libra.ErrRecordNotFound) {
		return fmt.Errorf("check Bayesian journal: %w", err)
	}
	if entry.Snapshot != s.snapshot {
		return store.ErrStaleSnapshot
	}
	metadata := map[string]interface{}{
		"record_type":  "frontier_journal",
		"tenant_id":    entry.TenantID,
		"journal_id":   entry.ID,
		"as_of":        entry.AsOf.UTC().Format(time.RFC3339Nano),
		"journal_json": string(encoded),
	}
	if err := s.bayesian.Insert(ctx, recordID, nil, metadata); err != nil {
		return fmt.Errorf("insert Bayesian journal: %w", err)
	}
	return nil
}

func (s *Store) GetBayesianJournal(ctx context.Context, tenantID, journalID string) (model.BayesianJournalEntry, error) {
	record, err := s.bayesian.Get(ctx, bayesianJournalRecordID(tenantID, journalID))
	if errors.Is(err, libra.ErrRecordNotFound) {
		return model.BayesianJournalEntry{}, store.ErrJournalNotFound
	}
	if err != nil {
		return model.BayesianJournalEntry{}, fmt.Errorf("read Bayesian journal: %w", err)
	}
	encoded, ok := record.Metadata["journal_json"].(string)
	if !ok {
		return model.BayesianJournalEntry{}, errors.New("Bayesian journal is malformed")
	}
	var entry model.BayesianJournalEntry
	if err := json.Unmarshal([]byte(encoded), &entry); err != nil {
		return model.BayesianJournalEntry{}, fmt.Errorf("decode Bayesian journal: %w", err)
	}
	if entry.TenantID != tenantID || entry.ID != journalID {
		return model.BayesianJournalEntry{}, errors.New("Bayesian journal identity mismatch")
	}
	return entry, nil
}

func (s *Store) PublishSelectionCertificate(ctx context.Context, certificate model.SelectionSupportCertificate) (model.Snapshot, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if certificate.PolicyVersion != s.snapshot.PolicyVersion || certificate.EvidenceEpoch != s.snapshot.EvidenceEpoch {
		return model.Snapshot{}, store.ErrStaleSnapshot
	}
	encoded, err := json.Marshal(certificate)
	if err != nil {
		return model.Snapshot{}, fmt.Errorf("encode selection certificate: %w", err)
	}
	next := s.snapshot
	next.RuntimeVersion++
	stateMetadata, err := s.stateMetadata(next)
	if err != nil {
		return model.Snapshot{}, err
	}
	recordID := selectionCertificateRecordID(certificate.TenantID)
	metadata := map[string]interface{}{"record_type": "selection_certificate", "tenant_id": certificate.TenantID, "certificate_json": string(encoded)}
	if err := s.db.WithTx(ctx, func(tx libra.Tx) error {
		if err := tx.Upsert(ctx, bayesianCollection, recordID, nil, metadata); err != nil {
			return err
		}
		return tx.Upsert(ctx, systemCollection, "runtime", nil, stateMetadata)
	}); err != nil {
		return model.Snapshot{}, fmt.Errorf("publish selection certificate: %w", err)
	}
	s.snapshot = next
	return next, nil
}

func (s *Store) GetSelectionCertificate(ctx context.Context, tenantID string) (model.SelectionSupportCertificate, error) {
	record, err := s.bayesian.Get(ctx, selectionCertificateRecordID(tenantID))
	if errors.Is(err, libra.ErrRecordNotFound) {
		return model.SelectionSupportCertificate{}, store.ErrCertificateNotFound
	}
	if err != nil {
		return model.SelectionSupportCertificate{}, fmt.Errorf("read selection certificate: %w", err)
	}
	var certificate model.SelectionSupportCertificate
	if err := decodeCertificate(record.Metadata, &certificate); err != nil {
		return model.SelectionSupportCertificate{}, err
	}
	if certificate.TenantID != tenantID {
		return model.SelectionSupportCertificate{}, errors.New("selection certificate tenant mismatch")
	}
	return certificate, nil
}

func (s *Store) PublishOmittedInfluenceCertificate(ctx context.Context, certificate model.OmittedInfluenceCertificate) (model.Snapshot, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if certificate.PolicyVersion != s.snapshot.PolicyVersion || certificate.EvidenceEpoch != s.snapshot.EvidenceEpoch {
		return model.Snapshot{}, store.ErrStaleSnapshot
	}
	encoded, err := json.Marshal(certificate)
	if err != nil {
		return model.Snapshot{}, fmt.Errorf("encode omitted-influence certificate: %w", err)
	}
	next := s.snapshot
	next.RuntimeVersion++
	stateMetadata, err := s.stateMetadata(next)
	if err != nil {
		return model.Snapshot{}, err
	}
	metadata := map[string]interface{}{"record_type": "omitted_influence_certificate", "tenant_id": certificate.TenantID, "certificate_json": string(encoded)}
	if err := s.db.WithTx(ctx, func(tx libra.Tx) error {
		if err := tx.Upsert(ctx, bayesianCollection, omittedInfluenceCertificateRecordID(certificate.TenantID), nil, metadata); err != nil {
			return err
		}
		return tx.Upsert(ctx, systemCollection, "runtime", nil, stateMetadata)
	}); err != nil {
		return model.Snapshot{}, fmt.Errorf("publish omitted-influence certificate: %w", err)
	}
	s.snapshot = next
	return next, nil
}

func (s *Store) GetOmittedInfluenceCertificate(ctx context.Context, tenantID string) (model.OmittedInfluenceCertificate, error) {
	record, err := s.bayesian.Get(ctx, omittedInfluenceCertificateRecordID(tenantID))
	if errors.Is(err, libra.ErrRecordNotFound) {
		return model.OmittedInfluenceCertificate{}, store.ErrCertificateNotFound
	}
	if err != nil {
		return model.OmittedInfluenceCertificate{}, fmt.Errorf("read omitted-influence certificate: %w", err)
	}
	var certificate model.OmittedInfluenceCertificate
	if err := decodeCertificate(record.Metadata, &certificate); err != nil {
		return model.OmittedInfluenceCertificate{}, err
	}
	if certificate.TenantID != tenantID {
		return model.OmittedInfluenceCertificate{}, errors.New("omitted-influence certificate tenant mismatch")
	}
	return certificate, nil
}

func (s *Store) PublishAntiPigeonCertificate(ctx context.Context, certificate model.AntiPigeonCertificate) (model.Snapshot, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if certificate.GraphVersion != s.snapshot.GraphVersion || certificate.EvidenceEpoch != s.snapshot.EvidenceEpoch {
		return model.Snapshot{}, store.ErrStaleSnapshot
	}
	encoded, err := json.Marshal(certificate)
	if err != nil {
		return model.Snapshot{}, fmt.Errorf("encode Anti-Pigeon certificate: %w", err)
	}
	certificateRecordID := antiPigeonCertificateRecordID(certificate.TenantID, certificate.ID)
	if current, getErr := s.bayesian.Get(ctx, certificateRecordID); getErr == nil {
		currentJSON, _ := current.Metadata["certificate_json"].(string)
		if currentJSON != string(encoded) {
			return model.Snapshot{}, store.ErrCertificateConflict
		}
		return s.snapshot, nil
	} else if !errors.Is(getErr, libra.ErrRecordNotFound) {
		return model.Snapshot{}, fmt.Errorf("check Anti-Pigeon certificate: %w", getErr)
	}
	next := s.snapshot
	next.RuntimeVersion++
	next.AbstractionVersion++
	stateMetadata, err := s.stateMetadata(next)
	if err != nil {
		return model.Snapshot{}, err
	}
	metadata := map[string]interface{}{"record_type": "anti_pigeon_certificate", "tenant_id": certificate.TenantID, "certificate_json": string(encoded)}
	if err := s.db.WithTx(ctx, func(tx libra.Tx) error {
		if err := tx.Insert(ctx, bayesianCollection, certificateRecordID, nil, metadata); err != nil {
			return err
		}
		for _, eventID := range certificate.MemberEventIDs {
			index := map[string]interface{}{"record_type": "anti_pigeon_index", "tenant_id": certificate.TenantID, "certificate_id": certificate.ID}
			if err := tx.Upsert(ctx, bayesianCollection, antiPigeonIndexRecordID(certificate.TenantID, eventID), nil, index); err != nil {
				return err
			}
		}
		return tx.Upsert(ctx, systemCollection, "runtime", nil, stateMetadata)
	}); err != nil {
		return model.Snapshot{}, fmt.Errorf("publish Anti-Pigeon certificate: %w", err)
	}
	s.snapshot = next
	return next, nil
}

func (s *Store) GetAntiPigeonCertificate(ctx context.Context, tenantID string, eventIDs []string) (model.AntiPigeonCertificate, error) {
	var certificateID string
	for _, eventID := range eventIDs {
		record, err := s.bayesian.Get(ctx, antiPigeonIndexRecordID(tenantID, eventID))
		if errors.Is(err, libra.ErrRecordNotFound) {
			return model.AntiPigeonCertificate{}, store.ErrCertificateNotFound
		}
		if err != nil {
			return model.AntiPigeonCertificate{}, fmt.Errorf("read Anti-Pigeon index: %w", err)
		}
		current, _ := record.Metadata["certificate_id"].(string)
		if current == "" || (certificateID != "" && current != certificateID) {
			return model.AntiPigeonCertificate{}, store.ErrCertificateNotFound
		}
		certificateID = current
	}
	if certificateID == "" {
		return model.AntiPigeonCertificate{}, store.ErrCertificateNotFound
	}
	record, err := s.bayesian.Get(ctx, antiPigeonCertificateRecordID(tenantID, certificateID))
	if errors.Is(err, libra.ErrRecordNotFound) {
		return model.AntiPigeonCertificate{}, store.ErrCertificateNotFound
	}
	if err != nil {
		return model.AntiPigeonCertificate{}, fmt.Errorf("read Anti-Pigeon certificate: %w", err)
	}
	var certificate model.AntiPigeonCertificate
	if err := decodeCertificate(record.Metadata, &certificate); err != nil {
		return model.AntiPigeonCertificate{}, err
	}
	return certificate, nil
}

func (s *Store) ApplyBayesianOutcome(ctx context.Context, request model.BayesianOutcomeRequest, posteriorKey, digest string, weight float64, changePolicy bayes.ChangePolicy, residualObservation model.ResidualObservation, residualPolicy residual.Policy) (store.BayesianOutcomeResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	outcomeRecordID := bayesianOutcomeRecordID(request.TenantID, request.IdempotencyKey)
	if record, err := s.bayesian.Get(ctx, outcomeRecordID); err == nil {
		currentDigest, _ := record.Metadata["content_digest"].(string)
		currentKey, _ := record.Metadata["posterior_key"].(string)
		if currentDigest != digest || currentKey != posteriorKey {
			return store.BayesianOutcomeResult{}, store.ErrOutcomeConflict
		}
		posterior, getErr := s.getBayesianPosterior(ctx, request.TenantID, posteriorKey)
		if getErr != nil {
			return store.BayesianOutcomeResult{}, getErr
		}
		return store.BayesianOutcomeResult{Duplicate: true, Posterior: posterior, Snapshot: s.snapshot}, nil
	} else if !errors.Is(err, libra.ErrRecordNotFound) {
		return store.BayesianOutcomeResult{}, fmt.Errorf("check Bayesian outcome: %w", err)
	}
	posterior, err := s.getBayesianPosterior(ctx, request.TenantID, posteriorKey)
	if errors.Is(err, store.ErrPosteriorNotFound) || posterior.EvidenceEpoch != s.snapshot.EvidenceEpoch {
		posterior = model.BayesianPosterior{TenantID: request.TenantID, PosteriorKey: posteriorKey, Alpha: 1, Beta: 1, EvidenceEpoch: s.snapshot.EvidenceEpoch, Certified: true}
	} else if err != nil {
		return store.BayesianOutcomeResult{}, err
	}
	posterior, changePoint := bayes.ApplyOutcome(posterior, request.Useful, weight, changePolicy)
	updateCalibration(&posterior, residualObservation.CommittedProbability, request.Useful, weight)
	posterior.UpdatedAt = request.AvailableAt
	outcomeJSON, err := json.Marshal(request)
	if err != nil {
		return store.BayesianOutcomeResult{}, fmt.Errorf("encode Bayesian outcome: %w", err)
	}
	next := s.snapshot
	if changePoint {
		next = invalidatedSnapshot(next)
		posterior.EvidenceEpoch = next.EvidenceEpoch
	} else {
		next.RuntimeVersion++
		next.PosteriorVersion++
		next.ResidualVersion++
	}
	exact, err := s.getResidualRecord(ctx, request.TenantID, model.ResidualExact, residualObservation.ActionKey)
	if err != nil && !errors.Is(err, store.ErrResidualNotFound) {
		return store.BayesianOutcomeResult{}, err
	}
	general, err := s.getResidualRecord(ctx, request.TenantID, model.ResidualGeneral, residualObservation.GeneralKey)
	if err != nil && !errors.Is(err, store.ErrResidualNotFound) {
		return store.BayesianOutcomeResult{}, err
	}
	exact = residual.Update(exact, residualObservation, model.ResidualExact, residualObservation.ActionKey, request.TenantID, weight, next, residualPolicy)
	general = residual.Update(general, residualObservation, model.ResidualGeneral, residualObservation.GeneralKey, request.TenantID, weight, next, residualPolicy)
	posteriorJSON, err := json.Marshal(posterior)
	if err != nil {
		return store.BayesianOutcomeResult{}, fmt.Errorf("encode Bayesian posterior: %w", err)
	}
	exactJSON, err := json.Marshal(exact)
	if err != nil {
		return store.BayesianOutcomeResult{}, fmt.Errorf("encode exact residual: %w", err)
	}
	generalJSON, err := json.Marshal(general)
	if err != nil {
		return store.BayesianOutcomeResult{}, fmt.Errorf("encode general residual: %w", err)
	}
	stateMetadata, err := s.stateMetadata(next)
	if err != nil {
		return store.BayesianOutcomeResult{}, err
	}
	posteriorMetadata := map[string]interface{}{"record_type": "posterior", "tenant_id": request.TenantID, "posterior_key": posteriorKey, "posterior_json": string(posteriorJSON)}
	exactMetadata := map[string]interface{}{"record_type": "residual", "tenant_id": request.TenantID, "scope": string(model.ResidualExact), "key": residualObservation.ActionKey, "residual_json": string(exactJSON)}
	generalMetadata := map[string]interface{}{"record_type": "residual", "tenant_id": request.TenantID, "scope": string(model.ResidualGeneral), "key": residualObservation.GeneralKey, "residual_json": string(generalJSON)}
	outcomeMetadata := map[string]interface{}{"record_type": "outcome", "tenant_id": request.TenantID, "posterior_key": posteriorKey, "content_digest": digest, "outcome_json": string(outcomeJSON)}
	if err := s.db.WithTx(ctx, func(tx libra.Tx) error {
		if err := tx.Insert(ctx, bayesianCollection, outcomeRecordID, nil, outcomeMetadata); err != nil {
			return err
		}
		if err := tx.Upsert(ctx, bayesianCollection, bayesianPosteriorRecordID(request.TenantID, posteriorKey), nil, posteriorMetadata); err != nil {
			return err
		}
		if err := tx.Upsert(ctx, bayesianCollection, residualRecordID(request.TenantID, model.ResidualExact, residualObservation.ActionKey), nil, exactMetadata); err != nil {
			return err
		}
		if err := tx.Upsert(ctx, bayesianCollection, residualRecordID(request.TenantID, model.ResidualGeneral, residualObservation.GeneralKey), nil, generalMetadata); err != nil {
			return err
		}
		return tx.Upsert(ctx, systemCollection, "runtime", nil, stateMetadata)
	}); err != nil {
		return store.BayesianOutcomeResult{}, fmt.Errorf("apply Bayesian outcome: %w", err)
	}
	s.snapshot = next
	return store.BayesianOutcomeResult{ChangePoint: changePoint, Posterior: posterior, Snapshot: next}, nil
}

func (s *Store) GetResidualCandidates(ctx context.Context, tenantID, actionKey, generalKey string) (model.ResidualCandidates, error) {
	var candidates model.ResidualCandidates
	exact, err := s.getResidualRecord(ctx, tenantID, model.ResidualExact, actionKey)
	if err == nil {
		candidates.Exact = &exact
	} else if !errors.Is(err, store.ErrResidualNotFound) {
		return model.ResidualCandidates{}, err
	}
	general, err := s.getResidualRecord(ctx, tenantID, model.ResidualGeneral, generalKey)
	if err == nil {
		candidates.General = &general
	} else if !errors.Is(err, store.ErrResidualNotFound) {
		return model.ResidualCandidates{}, err
	}
	return candidates, nil
}

func (s *Store) getResidualRecord(ctx context.Context, tenantID string, scope model.ResidualScope, key string) (model.ResidualRecord, error) {
	record, err := s.bayesian.Get(ctx, residualRecordID(tenantID, scope, key))
	if errors.Is(err, libra.ErrRecordNotFound) {
		return model.ResidualRecord{}, store.ErrResidualNotFound
	}
	if err != nil {
		return model.ResidualRecord{}, fmt.Errorf("read residual record: %w", err)
	}
	encoded, ok := record.Metadata["residual_json"].(string)
	if !ok {
		return model.ResidualRecord{}, errors.New("residual record is malformed")
	}
	var residualRecord model.ResidualRecord
	if err := json.Unmarshal([]byte(encoded), &residualRecord); err != nil {
		return model.ResidualRecord{}, fmt.Errorf("decode residual record: %w", err)
	}
	if residualRecord.TenantID != tenantID || residualRecord.Scope != scope || residualRecord.Key != key {
		return model.ResidualRecord{}, errors.New("residual record identity mismatch")
	}
	return residualRecord, nil
}

func (s *Store) GetBayesianPosterior(ctx context.Context, tenantID, posteriorKey string) (model.BayesianPosterior, error) {
	return s.getBayesianPosterior(ctx, tenantID, posteriorKey)
}

func (s *Store) getBayesianPosterior(ctx context.Context, tenantID, posteriorKey string) (model.BayesianPosterior, error) {
	record, err := s.bayesian.Get(ctx, bayesianPosteriorRecordID(tenantID, posteriorKey))
	if errors.Is(err, libra.ErrRecordNotFound) {
		return model.BayesianPosterior{}, store.ErrPosteriorNotFound
	}
	if err != nil {
		return model.BayesianPosterior{}, fmt.Errorf("read Bayesian posterior: %w", err)
	}
	encoded, ok := record.Metadata["posterior_json"].(string)
	if !ok {
		return model.BayesianPosterior{}, errors.New("Bayesian posterior is malformed")
	}
	var posterior model.BayesianPosterior
	if err := json.Unmarshal([]byte(encoded), &posterior); err != nil {
		return model.BayesianPosterior{}, fmt.Errorf("decode Bayesian posterior: %w", err)
	}
	if posterior.TenantID != tenantID || posterior.PosteriorKey != posteriorKey {
		return model.BayesianPosterior{}, errors.New("Bayesian posterior identity mismatch")
	}
	return posterior, nil
}

func decodeEligible(results []*libra.SearchResult, availableBy time.Time, limit int) []store.SearchResult {
	out := make([]store.SearchResult, 0, min(limit, len(results)))
	for _, result := range results {
		payload, ok := result.Metadata["event_json"].(string)
		if !ok {
			continue
		}
		var event model.Event
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			continue
		}
		event.Embedding = nil
		if event.AvailableAt.After(availableBy) {
			continue
		}
		out = append(out, store.SearchResult{Event: event, Similarity: float64(result.Score)})
		if len(out) >= limit {
			break
		}
	}
	return out
}

func (s *Store) Stats(ctx context.Context) (store.Stats, error) {
	databaseStats := s.db.Stats(ctx)
	stats := store.Stats{Backend: "libravdb", MemoryBytes: databaseStats.MemoryUsage}
	for name, collection := range databaseStats.Collections {
		if !strings.HasPrefix(name, "events_v2_") {
			continue
		}
		stats.Tenants++
		stats.Events += collection.LiveRecordCount
	}
	return stats, nil
}

func (s *Store) Delete(ctx context.Context, tenantID, eventID string) (store.DeleteResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	collection, err := s.collection(ctx, tenantID)
	if err != nil {
		return store.DeleteResult{}, err
	}
	if _, err := collection.Get(ctx, eventID); errors.Is(err, libra.ErrRecordNotFound) {
		return store.DeleteResult{Snapshot: s.snapshot}, nil
	} else if err != nil {
		return store.DeleteResult{}, err
	}
	next := invalidatedSnapshot(s.snapshot)
	metadata, err := s.stateMetadata(next)
	if err != nil {
		return store.DeleteResult{}, err
	}
	key := collectionName(tenantID, s.config.EmbeddingModel)
	err = s.db.WithTx(ctx, func(tx libra.Tx) error {
		if err := tx.Delete(ctx, key, eventID); err != nil {
			return err
		}
		return tx.Upsert(ctx, systemCollection, "runtime", nil, metadata)
	})
	if err != nil {
		return store.DeleteResult{}, fmt.Errorf("delete event: %w", err)
	}
	s.snapshot = next
	return store.DeleteResult{Deleted: true, Snapshot: next}, nil
}

func (s *Store) DeleteBefore(ctx context.Context, tenantID string, before time.Time, limit int) (store.RetentionResult, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	collection, err := s.collection(ctx, tenantID)
	if err != nil {
		return store.RetentionResult{}, err
	}
	records, err := collection.ListAll(ctx)
	if err != nil {
		return store.RetentionResult{}, err
	}
	type datedID struct {
		id string
		at time.Time
	}
	eligible := make([]datedID, 0)
	for _, record := range records {
		payload, ok := record.Metadata["event_json"].(string)
		if !ok {
			continue
		}
		var event model.Event
		if json.Unmarshal([]byte(payload), &event) != nil {
			continue
		}
		if event.AvailableAt.Before(before) {
			eligible = append(eligible, datedID{id: event.ID, at: event.AvailableAt})
		}
	}
	sort.Slice(eligible, func(i, j int) bool {
		if eligible[i].at.Equal(eligible[j].at) {
			return eligible[i].id < eligible[j].id
		}
		return eligible[i].at.Before(eligible[j].at)
	})
	if len(eligible) > limit {
		eligible = eligible[:limit]
	}
	ids := make([]string, len(eligible))
	for i := range eligible {
		ids[i] = eligible[i].id
	}
	if len(ids) == 0 {
		return store.RetentionResult{Snapshot: s.snapshot}, nil
	}
	next := invalidatedSnapshot(s.snapshot)
	metadata, err := s.stateMetadata(next)
	if err != nil {
		return store.RetentionResult{}, err
	}
	key := collectionName(tenantID, s.config.EmbeddingModel)
	err = s.db.WithTx(ctx, func(tx libra.Tx) error {
		if err := tx.DeleteBatch(ctx, key, ids); err != nil {
			return err
		}
		return tx.Upsert(ctx, systemCollection, "runtime", nil, metadata)
	})
	if err != nil {
		return store.RetentionResult{}, fmt.Errorf("retention delete: %w", err)
	}
	s.snapshot = next
	return store.RetentionResult{DeletedIDs: ids, Snapshot: next}, nil
}

func (s *Store) Backup(ctx context.Context, destination string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.db.Backup(ctx, destination)
}

func (s *Store) Compact(ctx context.Context) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.db.Vacuum(ctx)
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Snapshot(_ context.Context) model.Snapshot {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.snapshot
}

func (s *Store) collection(ctx context.Context, tenantID string) (*libra.Collection, error) {
	name := collectionName(tenantID, s.config.EmbeddingModel)
	s.mu.Lock()
	defer s.mu.Unlock()
	if collection := s.collections[name]; collection != nil {
		return collection, nil
	}
	names, err := s.db.ListCollectionsWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	for _, existingName := range names {
		if existingName != name {
			continue
		}
		collection, err := s.db.GetCollection(name)
		if err != nil {
			return nil, fmt.Errorf("load tenant collection: %w", err)
		}
		if collection.Dimension() != s.config.Dimension {
			return nil, fmt.Errorf("tenant collection dimension %d does not match %d", collection.Dimension(), s.config.Dimension)
		}
		s.collections[name] = collection
		return collection, nil
	}
	options := []libra.CollectionOption{
		libra.WithMetric(libra.CosineDistance),
		libra.WithHNSW(16, 200, 100),
		libra.WithMemoryMapping(s.config.MemoryMapping),
	}
	switch s.config.Quantization {
	case "", "none":
	case "sq8":
		options = append(options, libra.WithScalarQuantization(8, 0.10))
	case "fsq6":
		options = append(options, libra.WithFSQQuantization(6, 0.10))
	case "pq8":
		options = append(options, libra.WithProductQuantization(8, 8, 0.10))
	default:
		return nil, fmt.Errorf("unsupported quantization %q", s.config.Quantization)
	}
	collection, err := s.db.EnsureCollection(ctx, name, s.config.Dimension, options...)
	if err != nil {
		return nil, fmt.Errorf("ensure tenant collection: %w", err)
	}
	s.collections[name] = collection
	return collection, nil
}

func collectionName(tenantID, embeddingModel string) string {
	digest := sha256.Sum256([]byte(tenantID + "\x00" + embeddingModel))
	return "events_v2_" + hex.EncodeToString(digest[:12])
}

func (s *Store) loadOrInitializeState(ctx context.Context) error {
	collection, err := s.db.EnsureCollection(ctx, systemCollection, 0, libra.WithMetadataOnly())
	if err != nil {
		return fmt.Errorf("ensure system collection: %w", err)
	}
	s.bayesian, err = s.db.EnsureCollection(ctx, bayesianCollection, 0, libra.WithMetadataOnly())
	if err != nil {
		return fmt.Errorf("ensure Bayesian collection: %w", err)
	}
	record, err := collection.Get(ctx, "runtime")
	if err == nil {
		encoded, ok := record.Metadata["state_json"].(string)
		if !ok {
			return errors.New("durable runtime state is malformed")
		}
		var state persistentState
		if err := json.Unmarshal([]byte(encoded), &state); err != nil {
			return fmt.Errorf("decode durable runtime state: %w", err)
		}
		if state.SchemaVersion != 2 {
			return fmt.Errorf("database schema %d requires migration", state.SchemaVersion)
		}
		if state.EmbeddingModel != s.config.EmbeddingModel || state.Dimension != s.config.Dimension || state.Quantization != s.config.Quantization {
			return fmt.Errorf("database embedding contract %q/d%d does not match active %q/d%d", state.EmbeddingModel, state.Dimension, s.config.EmbeddingModel, s.config.Dimension)
		}
		if state.Snapshot.ContractVersion > model.ContractVersion {
			return fmt.Errorf("database contract version %d is newer than runtime contract %d", state.Snapshot.ContractVersion, model.ContractVersion)
		}
		s.policyDigest = state.BayesianPolicyDigest
		if state.Snapshot.ContractVersion < model.ContractVersion {
			state.Snapshot.ContractVersion = model.ContractVersion
			state.Snapshot.RuntimeVersion++
			metadata, metadataErr := s.stateMetadata(state.Snapshot)
			if metadataErr != nil {
				return metadataErr
			}
			if upsertErr := collection.Upsert(ctx, "runtime", nil, metadata); upsertErr != nil {
				return fmt.Errorf("upgrade durable contract marker: %w", upsertErr)
			}
		}
		s.snapshot = state.Snapshot
		return nil
	}
	if !errors.Is(err, libra.ErrRecordNotFound) {
		return fmt.Errorf("read durable runtime state: %w", err)
	}
	names, listErr := s.db.ListCollectionsWithContext(ctx)
	if listErr != nil {
		return fmt.Errorf("inspect database schema: %w", listErr)
	}
	for _, name := range names {
		if isLegacyCollection(name) {
			return errors.New("legacy Phase 1 collections detected; run with -migrate-v1 and -migration-backup before starting")
		}
	}
	s.snapshot = model.Snapshot{PolicyVersion: 1, ContractVersion: model.ContractVersion, GraphVersion: 1, PosteriorVersion: 1, ResidualVersion: 1, AbstractionVersion: 1}
	metadata, err := s.stateMetadata(s.snapshot)
	if err != nil {
		return err
	}
	if err := collection.Insert(ctx, "runtime", nil, metadata); err != nil {
		return fmt.Errorf("initialize durable runtime state: %w", err)
	}
	return nil
}

func bayesianJournalRecordID(tenantID, journalID string) string {
	digest := sha256.Sum256([]byte(tenantID + "\x00" + journalID))
	return "journal_" + hex.EncodeToString(digest[:])
}

func selectionCertificateRecordID(tenantID string) string {
	return hashedBayesianRecordID("selection", tenantID)
}

func omittedInfluenceCertificateRecordID(tenantID string) string {
	return hashedBayesianRecordID("omitted_influence", tenantID)
}

func antiPigeonCertificateRecordID(tenantID, certificateID string) string {
	return hashedBayesianRecordID("anti_pigeon", tenantID+"\x00"+certificateID)
}

func antiPigeonIndexRecordID(tenantID, eventID string) string {
	return hashedBayesianRecordID("anti_pigeon_index", tenantID+"\x00"+eventID)
}

func hashedBayesianRecordID(kind, value string) string {
	digest := sha256.Sum256([]byte(kind + "\x00" + value))
	return kind + "_" + hex.EncodeToString(digest[:])
}

func bayesianOutcomeRecordID(tenantID, idempotencyKey string) string {
	return hashedBayesianRecordID("outcome", tenantID+"\x00"+idempotencyKey)
}

func bayesianPosteriorRecordID(tenantID, posteriorKey string) string {
	return hashedBayesianRecordID("posterior", tenantID+"\x00"+posteriorKey)
}

func residualRecordID(tenantID string, scope model.ResidualScope, key string) string {
	return hashedBayesianRecordID("residual", tenantID+"\x00"+string(scope)+"\x00"+key)
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

func decodeCertificate(metadata map[string]interface{}, target any) error {
	encoded, ok := metadata["certificate_json"].(string)
	if !ok {
		return errors.New("Bayesian certificate is malformed")
	}
	if err := json.Unmarshal([]byte(encoded), target); err != nil {
		return fmt.Errorf("decode Bayesian certificate: %w", err)
	}
	return nil
}

func isLegacyCollection(name string) bool {
	return strings.HasPrefix(name, "events_") && !strings.HasPrefix(name, "events_v2_")
}

func (s *Store) stateMetadata(snapshot model.Snapshot) (map[string]interface{}, error) {
	return s.stateMetadataWithPolicy(snapshot, s.policyDigest)
}

func (s *Store) stateMetadataWithPolicy(snapshot model.Snapshot, policyDigest string) (map[string]interface{}, error) {
	encoded, err := json.Marshal(persistentState{SchemaVersion: 2, EmbeddingModel: s.config.EmbeddingModel, Dimension: s.config.Dimension, Quantization: s.config.Quantization, Snapshot: snapshot, BayesianPolicyDigest: policyDigest})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"state_json": string(encoded)}, nil
}

func invalidatedSnapshot(snapshot model.Snapshot) model.Snapshot {
	snapshot.RuntimeVersion++
	snapshot.EvidenceEpoch++
	snapshot.GraphVersion++
	snapshot.PosteriorVersion++
	snapshot.ResidualVersion++
	snapshot.AbstractionVersion++
	return snapshot
}
