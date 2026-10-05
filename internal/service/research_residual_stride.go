package service

import (
	"context"
	"errors"
	"sync"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// researchLoadResidualStride is an unwired research alternative to the per-item
// channel dispatcher. Each worker owns disjoint output indices; cancellation
// joins all workers before returning, preserving the caller's snapshot lifetime.
func (s *Service) researchLoadResidualStride(ctx context.Context, request model.RecallRequest, queryDigest string, candidates []model.Candidate, decisions []model.BayesianDecision, decisionIndexes map[string]int) ([]model.ResidualCandidates, error) {
	loaded := make([]model.ResidualCandidates, len(candidates))
	if s.config.ResidualMode == ResidualModeDisabled || len(candidates) == 0 {
		return loaded, nil
	}
	for _, candidate := range candidates {
		if _, ok := decisionIndexes[candidate.Event.ID]; !ok {
			return nil, errors.New("Bayesian frontier omitted a recalled candidate")
		}
	}
	workerCount := min(8, len(candidates))
	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var workers sync.WaitGroup
	var firstErr error
	var errOnce sync.Once
	for worker := 0; worker < workerCount; worker++ {
		workers.Add(1)
		go func(worker int) {
			defer workers.Done()
			for index := worker; index < len(candidates); index += workerCount {
				if workCtx.Err() != nil {
					return
				}
				candidate := candidates[index]
				decision := decisions[decisionIndexes[candidate.Event.ID]]
				actionKey := residualActionKey(queryDigest, candidate.Event.ID, model.RetrievalUsefulnessHorizon)
				generalKey := residualGeneralKey(decision.PosteriorKey, model.RetrievalUsefulnessHorizon)
				residuals, err := s.store.GetResidualCandidates(workCtx, request.TenantID, actionKey, generalKey)
				if err != nil {
					errOnce.Do(func() { firstErr = err; cancel() })
					return
				}
				loaded[index] = residuals
			}
		}(worker)
	}
	workers.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return loaded, nil
}
