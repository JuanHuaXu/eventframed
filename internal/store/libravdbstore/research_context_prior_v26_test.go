package libravdbstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"sync/atomic"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type priorScopeKeyV26 struct{}

type priorScopeV26 struct {
	tenant, digest string
	asOf           time.Time
}

// The caller supplies scope from the actual request, not from a cached key.
// This isolated adapter is not a new native Store query-context contract.
func priorContextV26(ctx context.Context, request model.RecallRequest) (context.Context, error) {
	payload, err := json.Marshal(struct {
		Query          string    `json:"query"`
		Embedding      []float32 `json:"embedding"`
		EmbeddingModel string    `json:"embedding_model"`
	}{frame.QueryText(request.Query), request.Embedding, request.EmbeddingModel})
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(payload)
	scope := priorScopeV26{request.TenantID, hex.EncodeToString(digest[:]), request.AsOf}
	return context.WithValue(ctx, priorScopeKeyV26{}, scope), nil
}

type priorBindingV26 struct {
	journal model.BayesianJournalEntry
	means   map[string]float64
}

// Only the predictive read is reparameterized. The owned outcome store keeps
// ordinary Beta(1,1) counts, the original journal and its publication ordering.
// Shared, tempered and unknown-context evidence is deliberately unsupported.
type contextPriorStoreV26 struct {
	store.EventStore
	binding atomic.Pointer[priorBindingV26]
}

func (s *contextPriorStoreV26) bind(ctx context.Context, tenant, journalID string) error {
	journal, err := s.EventStore.GetBayesianJournal(ctx, tenant, journalID)
	if err != nil {
		return err
	}
	current := s.EventStore.Snapshot(ctx)
	if journal.TenantID != tenant || journal.ID != journalID || journal.QueryDigest == "" ||
		!journal.Report.JournalDurable || journal.Report.JournalID != journalID ||
		!journal.Report.SelectionSupportCertified || !journal.Report.OmittedInfluenceCertified ||
		current.PolicyVersion != journal.Snapshot.PolicyVersion || current.EvidenceEpoch != journal.Snapshot.EvidenceEpoch ||
		len(journal.Report.Decisions) == 0 || len(journal.Report.Decisions) > 200 {
		return errors.New("invalid durable context-prior origin")
	}
	bound := &priorBindingV26{journal: journal, means: make(map[string]float64, len(journal.Report.Decisions))}
	for _, decision := range journal.Report.Decisions {
		mean := decision.Forecast.BaseLaw.Useful
		if !decision.Activated || decision.PosteriorKey != decision.EventID ||
			decision.Forecast.HorizonKey != model.RetrievalUsefulnessHorizon ||
			decision.Forecast.BeliefLaw != nil || decision.Forecast.ResidualApplied ||
			decision.Forecast.ExpertMixture.Enabled || mean <= 0 || mean >= 1 || math.IsNaN(mean) {
			return errors.New("origin is not an unshared pre-feedback baseline")
		}
		if _, duplicate := bound.means[decision.PosteriorKey]; duplicate {
			return errors.New("duplicate context-prior key")
		}
		bound.means[decision.PosteriorKey] = mean
	}
	if !s.binding.CompareAndSwap(nil, bound) {
		return errors.New("context-prior origin is immutable")
	}
	return nil
}

func (s *contextPriorStoreV26) GetBayesianPosterior(ctx context.Context, tenant, key string) (model.BayesianPosterior, error) {
	bound := s.binding.Load()
	scope, scoped := ctx.Value(priorScopeKeyV26{}).(priorScopeV26)
	if bound == nil || !scoped || scope.tenant != tenant || tenant != bound.journal.TenantID ||
		scope.digest != bound.journal.QueryDigest || scope.asOf.Before(bound.journal.AsOf) {
		return model.BayesianPosterior{}, store.ErrPosteriorNotFound
	}
	mean, present := bound.means[key]
	if !present {
		return model.BayesianPosterior{}, store.ErrPosteriorNotFound
	}
	posterior, err := s.EventStore.GetBayesianPosterior(ctx, tenant, key)
	if err != nil {
		return model.BayesianPosterior{}, err
	}
	current := s.EventStore.Snapshot(ctx)
	if !posterior.Certified || posterior.TenantID != tenant || posterior.PosteriorKey != key ||
		posterior.EvidenceEpoch != bound.journal.Snapshot.EvidenceEpoch || current.EvidenceEpoch != posterior.EvidenceEpoch ||
		current.PolicyVersion != bound.journal.Snapshot.PolicyVersion ||
		posterior.UpdatedAt.Before(bound.journal.AsOf) || posterior.UpdatedAt.After(scope.asOf) || posterior.WorkingBelief != nil {
		return model.BayesianPosterior{}, store.ErrPosteriorNotFound
	}
	u, v := posterior.Alpha-1, posterior.Beta-1
	if u < 0 || v < 0 || !finiteCountV26(u) || !finiteCountV26(v) || u+v > 1000000 ||
		posterior.EffectiveSupport != u+v {
		return model.BayesianPosterior{}, store.ErrPosteriorNotFound
	}
	posterior.Alpha, posterior.Beta = 2*mean+u, 2*(1-mean)+v
	return posterior, nil
}

func finiteCountV26(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && math.Trunc(value) == value
}

func (s *contextPriorStoreV26) ApplyBayesianOutcome(ctx context.Context, request model.BayesianOutcomeRequest,
	posteriorKey, parentPosteriorKey, digest string, weight float64,
	changePolicy bayes.ChangePolicy, groupPolicy bayes.GroupPolicy,
	observation model.ResidualObservation, residualPolicy residual.Policy,
) (store.BayesianOutcomeResult, error) {
	bound := s.binding.Load()
	if bound == nil || request.TenantID != bound.journal.TenantID || request.Source != model.OutcomeFullStream ||
		weight != 1 || parentPosteriorKey != "" || posteriorKey != request.EventID ||
		changePolicy.Working.Enabled || changePolicy.ForecastRescue {
		return store.BayesianOutcomeResult{}, errors.New("context prior requires owned ordinary full-stream outcomes")
	}
	if _, present := bound.means[posteriorKey]; !present {
		return store.BayesianOutcomeResult{}, errors.New("outcome is outside the anchored frontier")
	}
	journal, err := s.EventStore.GetBayesianJournal(ctx, request.TenantID, request.JournalID)
	if err != nil {
		return store.BayesianOutcomeResult{}, err
	}
	if journal.QueryDigest != bound.journal.QueryDigest || journal.AsOf.Before(bound.journal.AsOf) ||
		journal.Snapshot.PolicyVersion != bound.journal.Snapshot.PolicyVersion ||
		journal.Snapshot.EvidenceEpoch != bound.journal.Snapshot.EvidenceEpoch {
		return store.BayesianOutcomeResult{}, errors.New("outcome journal is outside the context-prior origin")
	}
	return s.EventStore.ApplyBayesianOutcome(ctx, request, posteriorKey, parentPosteriorKey, digest,
		weight, changePolicy, groupPolicy, observation, residualPolicy)
}
