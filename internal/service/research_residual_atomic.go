package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// researchLoadResidualAtomic assigns one unique index per atomic increment.
// Workers take another index when ready, retaining dynamic load balancing without
// a channel handoff per lookup. It remains unwired; cancellation joins workers.
func (s *Service) researchLoadResidualAtomic(ctx context.Context, request model.RecallRequest, queryDigest string, candidates []model.Candidate, decisions []model.BayesianDecision, decisionIndexes map[string]int) ([]model.ResidualCandidates, error) {
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
	var next atomic.Int64
	for worker := 0; worker < workerCount; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				if workCtx.Err() != nil {
					return
				}
				index := int(next.Add(1) - 1)
				if index >= len(candidates) {
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
		}()
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
