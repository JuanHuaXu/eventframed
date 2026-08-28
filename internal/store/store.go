package store

import (
	"context"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

var ErrIdempotencyConflict = errors.New("event id already exists with different content")
var ErrJournalConflict = errors.New("Bayesian journal id already exists with different content")
var ErrJournalNotFound = errors.New("Bayesian journal not found")
var ErrCertificateNotFound = errors.New("Bayesian certificate not found")
var ErrCertificateConflict = errors.New("Bayesian certificate id already exists with different content")
var ErrOutcomeConflict = errors.New("Bayesian outcome id already exists with different content")
var ErrPosteriorNotFound = errors.New("Bayesian posterior not found")
var ErrStaleSnapshot = errors.New("runtime snapshot changed before Bayesian journal commit")

type PutResult struct {
	Duplicate bool
	Snapshot  model.Snapshot
}

type DeleteResult struct {
	Deleted  bool
	Snapshot model.Snapshot
}
type RetentionResult struct {
	DeletedIDs []string
	Snapshot   model.Snapshot
}

type BayesianOutcomeResult struct {
	Duplicate   bool
	ChangePoint bool
	Posterior   model.BayesianPosterior
	Snapshot    model.Snapshot
}

type SearchResult struct {
	Event      model.Event
	Similarity float64
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
	Delete(ctx context.Context, tenantID, eventID string) (DeleteResult, error)
	DeleteBefore(ctx context.Context, tenantID string, before time.Time, limit int) (RetentionResult, error)
	Backup(ctx context.Context, destination string) error
	Compact(ctx context.Context) error
	Search(ctx context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]SearchResult, error)
	PutBayesianJournal(ctx context.Context, entry model.BayesianJournalEntry) error
	GetBayesianJournal(ctx context.Context, tenantID, journalID string) (model.BayesianJournalEntry, error)
	PublishSelectionCertificate(ctx context.Context, certificate model.SelectionSupportCertificate) (model.Snapshot, error)
	GetSelectionCertificate(ctx context.Context, tenantID string) (model.SelectionSupportCertificate, error)
	PublishAntiPigeonCertificate(ctx context.Context, certificate model.AntiPigeonCertificate) (model.Snapshot, error)
	GetAntiPigeonCertificate(ctx context.Context, tenantID string, eventIDs []string) (model.AntiPigeonCertificate, error)
	PublishOmittedInfluenceCertificate(ctx context.Context, certificate model.OmittedInfluenceCertificate) (model.Snapshot, error)
	GetOmittedInfluenceCertificate(ctx context.Context, tenantID string) (model.OmittedInfluenceCertificate, error)
	ApplyBayesianOutcome(ctx context.Context, request model.BayesianOutcomeRequest, posteriorKey, digest string, weight float64, changePolicy bayes.ChangePolicy) (BayesianOutcomeResult, error)
	GetBayesianPosterior(ctx context.Context, tenantID, posteriorKey string) (model.BayesianPosterior, error)
	Stats(ctx context.Context) (Stats, error)
	Snapshot(ctx context.Context) model.Snapshot
	Close() error
}
