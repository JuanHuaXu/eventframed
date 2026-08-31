package store

import (
	"context"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
)

var ErrIdempotencyConflict = errors.New("event id already exists with different content")
var ErrJournalConflict = errors.New("Bayesian journal id already exists with different content")
var ErrJournalNotFound = errors.New("Bayesian journal not found")
var ErrCertificateNotFound = errors.New("Bayesian certificate not found")
var ErrCertificateConflict = errors.New("Bayesian certificate id already exists with different content")
var ErrOutcomeConflict = errors.New("Bayesian outcome id already exists with different content")
var ErrPosteriorNotFound = errors.New("Bayesian posterior not found")
var ErrResidualNotFound = errors.New("residual record not found")
var ErrEventNotFound = errors.New("event not found")
var ErrStaleSnapshot = errors.New("runtime snapshot changed before Bayesian journal commit")
var ErrSnapNotFound = errors.New("predictive snap not found")
var ErrSnapConflict = errors.New("predictive snap conflicts with current graph")
var ErrAgencyConflict = errors.New("agency proposal id already exists with different content or resolution")
var ErrAgencyNotFound = errors.New("agency proposal not found")
var ErrAgencyLease = errors.New("agency proposal is not held by the resolving consumer")
var ErrAgencyExpired = errors.New("agency proposal expired before resolution")
var ErrAgencyChainBudget = errors.New("agency causal-chain budget is exhausted or invalid")
var ErrAgencyEvidence = errors.New("agency proposal references missing, cross-tenant, or unavailable evidence")
var ErrCompositionAuthority = errors.New("higher-order composition lacks current Anti-Pigeon authority")

// JournalSnapshotCompatible permits only fully-accounted future-available
// ingestion after a recall's fixed as_of. Unclassified runtime motion remains
// decision-relevant and therefore fails closed.
func JournalSnapshotCompatible(captured, current model.Snapshot, asOf time.Time, ingestMotion map[uint64]time.Time) bool {
	semanticVersionsMatch := captured.PolicyVersion == current.PolicyVersion &&
		captured.ContractVersion == current.ContractVersion &&
		captured.GraphVersion == current.GraphVersion &&
		captured.PosteriorVersion == current.PosteriorVersion &&
		captured.ResidualVersion == current.ResidualVersion &&
		captured.AbstractionVersion == current.AbstractionVersion &&
		captured.AgencyVersion == current.AgencyVersion
	if !semanticVersionsMatch || current.RuntimeVersion < captured.RuntimeVersion {
		return false
	}
	for version := captured.RuntimeVersion + 1; version <= current.RuntimeVersion; version++ {
		availableAt, ok := ingestMotion[version]
		if !ok || !availableAt.After(asOf) {
			return false
		}
	}
	return true
}

type PutResult struct {
	Duplicate bool
	Snapshot  model.Snapshot
}

type DeleteResult struct {
	Deleted  bool
	Snapshot model.Snapshot
}

type CompositionDeleteResult struct {
	Deleted        bool
	MemberEventIDs []string
	Snapshot       model.Snapshot
}
type RetentionResult struct {
	DeletedIDs []string
	Snapshot   model.Snapshot
}

type BayesianOutcomeResult struct {
	Duplicate   bool
	ChangePoint bool
	Revision    model.BayesianRevision
	Posterior   model.BayesianPosterior
	Snapshot    model.Snapshot
}

type AgencyPutResult struct {
	Duplicate bool
	Record    model.AgencyProposalRecord
	Snapshot  model.Snapshot
}

type AgencyResolveResult struct {
	Duplicate bool
	Record    model.AgencyProposalRecord
	Snapshot  model.Snapshot
}

type SearchResult struct {
	Event             model.Event
	Similarity        float64
	RetrievalMetadata []byte
}

type Stats struct {
	Backend     string `json:"backend"`
	Tenants     int    `json:"tenants"`
	Events      int    `json:"events"`
	MemoryBytes int64  `json:"memory_bytes"`
}

type EventStore interface {
	BindBayesianPolicy(ctx context.Context, digest string) (model.Snapshot, error)
	Put(ctx context.Context, event model.Event, vector []float32, digest string) (PutResult, error)
	PutComposition(ctx context.Context, event model.Event, vector []float32, digest string, base model.Snapshot) (PutResult, error)
	Delete(ctx context.Context, tenantID, eventID string) (DeleteResult, error)
	DeleteComposition(ctx context.Context, tenantID, eventID, reason string, decomposedAt time.Time) (CompositionDeleteResult, error)
	GetCompositionTombstone(ctx context.Context, tenantID, eventID string) (model.CompositionTombstone, error)
	DeleteBefore(ctx context.Context, tenantID string, before time.Time, limit int) (RetentionResult, error)
	Backup(ctx context.Context, destination string) error
	Compact(ctx context.Context) error
	Search(ctx context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]SearchResult, error)
	GetEvents(ctx context.Context, tenantID string, eventIDs []string, availableBy time.Time) ([]model.Event, error)
	PutBayesianJournal(ctx context.Context, entry model.BayesianJournalEntry) error
	GetBayesianJournal(ctx context.Context, tenantID, journalID string) (model.BayesianJournalEntry, error)
	PublishSelectionCertificate(ctx context.Context, certificate model.SelectionSupportCertificate) (model.Snapshot, error)
	GetSelectionCertificate(ctx context.Context, tenantID string) (model.SelectionSupportCertificate, error)
	PublishAntiPigeonCertificate(ctx context.Context, certificate model.AntiPigeonCertificate) (model.Snapshot, error)
	GetAntiPigeonCertificate(ctx context.Context, tenantID string, eventIDs []string) (model.AntiPigeonCertificate, error)
	PublishOmittedInfluenceCertificate(ctx context.Context, certificate model.OmittedInfluenceCertificate) (model.Snapshot, error)
	GetOmittedInfluenceCertificate(ctx context.Context, tenantID string) (model.OmittedInfluenceCertificate, error)
	ApplyBayesianOutcome(ctx context.Context, request model.BayesianOutcomeRequest, posteriorKey, parentPosteriorKey, digest string, weight float64, changePolicy bayes.ChangePolicy, groupPolicy bayes.GroupPolicy, residualObservation model.ResidualObservation, residualPolicy residual.Policy) (BayesianOutcomeResult, error)
	GetBayesianPosterior(ctx context.Context, tenantID, posteriorKey string) (model.BayesianPosterior, error)
	GetResidualCandidates(ctx context.Context, tenantID, actionKey, generalKey string) (model.ResidualCandidates, error)
	GetPredictiveGraph(ctx context.Context, tenantID string) (model.PredictiveGraph, error)
	PublishPredictiveSnap(ctx context.Context, record model.PredictiveSnapRecord) (model.PredictiveGraph, model.Snapshot, error)
	RollbackPredictiveSnap(ctx context.Context, tenantID, snapID, reason string) (model.PredictiveGraph, model.Snapshot, error)
	PutAgencyProposal(ctx context.Context, record model.AgencyProposalRecord, digest string, maxPerChain, maxPending int, evidenceAvailableBy time.Time) (AgencyPutResult, error)
	ClaimAgencyProposals(ctx context.Context, tenantID, consumerID string, now time.Time, limit int, lease time.Duration) ([]model.AgencyProposalRecord, model.Snapshot, error)
	ResolveAgencyProposal(ctx context.Context, request model.ResolveAgencyProposalRequest, now time.Time) (AgencyResolveResult, error)
	Stats(ctx context.Context) (Stats, error)
	Snapshot(ctx context.Context) model.Snapshot
	Close() error
}

// VectorEventStore is an optional capability for bounded analytical paths that
// need durable semantic vectors. Keeping it outside EventStore preserves the
// ordinary retrieval contract and third-party store compatibility.
type VectorEventStore interface {
	GetEventsWithVectors(ctx context.Context, tenantID string, eventIDs []string, availableBy time.Time) ([]model.Event, error)
}
