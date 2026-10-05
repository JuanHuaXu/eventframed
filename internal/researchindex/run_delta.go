package researchindex

import (
	"context"
	"errors"
	"math"
	"sort"
)

// RunDeltaSearch is an immutable query snapshot, not a writer or a WAL owner.
// One bounded delta is the newest virtual run. Its complete membership shadows
// all older graphs, including when its replacement score drops below top k.
type RunDeltaSearch struct {
	runs      []*ImmutableRun
	delta     map[string]record
	plan      *RunMergePlan
	dimension int
}

func NewRunDeltaSearch(ctx context.Context, runs []*ImmutableRun, delta []Mutation, dimension, capacity, k int) (*RunDeltaSearch, error) {
	if len(runs) > 15 {
		return nil, ErrCapacity
	} // reserve the sixteenth slot for delta
	core, err := NewGenerations(dimension, capacity)
	if err != nil {
		return nil, err
	}
	owned := map[string]record{}
	if len(delta) > 0 {
		prepared, err := core.Prepare(ctx, delta)
		if err != nil {
			return nil, err
		}
		owned = prepared.next.delta
		if err = prepared.Abort(); err != nil {
			return nil, err
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	manifests := make([][]RunEntry, len(runs)+1)
	for id, r := range owned {
		manifests[0] = append(manifests[0], RunEntry{ID: id, Deleted: r.deleted})
	}
	for i, r := range runs {
		if r == nil || r.dimension != dimension {
			return nil, errors.New("run dimension mismatch")
		}
		manifests[i+1] = r.manifest
	}
	plan, err := NewRunMergePlan(manifests, k, 200)
	if err != nil {
		return nil, err
	}
	return &RunDeltaSearch{runs: append([]*ImmutableRun(nil), runs...), delta: owned, plan: plan, dimension: dimension}, nil
}

func (s *RunDeltaSearch) Search(ctx context.Context, q []float32) ([]Candidate, error) {
	if len(q) != s.dimension {
		return nil, errors.New("query dimension mismatch")
	}
	var norm float64
	for _, v := range q {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return nil, errors.New("nonfinite query")
		}
		norm += float64(v) * float64(v)
	}
	if norm == 0 {
		return nil, errors.New("zero query")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	prefixes := make([][]Candidate, len(s.runs)+1)
	for id, r := range s.delta {
		if !r.deleted {
			prefixes[0] = append(prefixes[0], Candidate{ID: id, Score: cosine(q, r.vector)})
		}
	}
	sort.Slice(prefixes[0], func(i, j int) bool { return runCandidateLess(prefixes[0][i], prefixes[0][j]) })
	prefixes[0] = prefixes[0][:s.plan.limits[0]]
	for i, r := range s.runs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		limit := s.plan.limits[i+1]
		if limit == 0 {
			r.index.mu.RLock()
			closed := r.index.closed
			r.index.mu.RUnlock()
			if closed {
				return nil, errors.New("empty run closed")
			}
			continue
		}
		p, err := r.index.Search(ctx, r.view, q, limit)
		if err != nil {
			return nil, err
		}
		prefixes[i+1] = p
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.plan.Merge(prefixes)
}
