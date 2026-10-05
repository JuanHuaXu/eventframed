package researchindex

import (
	"context"
	"errors"
	"math"
)

// ImmutableRun binds a copied complete manifest to its matching private graph.
// This is a static research object, not a durable publication/retirement owner.
type ImmutableRun struct {
	index     *HNSWBase
	view      View
	manifest  []RunEntry
	dimension int
}

func BuildImmutableRun(ctx context.Context, mutations []Mutation, dimension int, directory string) (*ImmutableRun, error) {
	if dimension < 1 {
		return nil, errors.New("invalid run dimension")
	}
	base := make(map[string]record, len(mutations))
	seen := make(map[string]bool, len(mutations))
	manifest := make([]RunEntry, 0, len(mutations))
	for _, m := range mutations {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if m.ID == "" || seen[m.ID] {
			return nil, errors.New("invalid run ID")
		}
		seen[m.ID] = true
		manifest = append(manifest, RunEntry{ID: m.ID, Deleted: m.Delete})
		if m.Delete {
			continue
		}
		if len(m.Vector) != dimension {
			return nil, errors.New("invalid run vector dimension")
		}
		var norm float64
		for _, v := range m.Vector {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				return nil, errors.New("nonfinite run vector")
			}
			norm += float64(v) * float64(v)
		}
		if norm == 0 {
			return nil, errors.New("zero run vector")
		}
		base[m.ID] = record{vector: append([]float32(nil), m.Vector...)}
	}
	v := View{g: &generation{base: &baseGeneration{records: base}, delta: map[string]record{}}}
	h, err := BuildHNSWBase(ctx, v, dimension, directory)
	if err != nil {
		return nil, err
	}
	return &ImmutableRun{index: h, view: v, manifest: manifest, dimension: dimension}, nil
}

func (r *ImmutableRun) Close() error { return r.index.Close() }

// ImmutableRunSearch captures graph/manifest pairs newest first, with no mutable
// delta. Owners must retain all runs until readers finish. No cross-run lease or
// atomic publication is supplied here. Closed runs cause an error, never fallback.
type ImmutableRunSearch struct {
	runs      []*ImmutableRun
	plan      *RunMergePlan
	dimension int
}

func NewImmutableRunSearch(runs []*ImmutableRun, k int) (*ImmutableRunSearch, error) {
	if len(runs) < 1 || len(runs) > 16 {
		return nil, errors.New("invalid run count")
	}
	manifests := make([][]RunEntry, len(runs))
	dim := 0
	for i, r := range runs {
		if r == nil {
			return nil, errors.New("nil run")
		}
		if i == 0 {
			dim = r.dimension
		}
		if dim != r.dimension {
			return nil, errors.New("run dimension mismatch")
		}
		manifests[i] = r.manifest
	}
	// HNSWBase.Search supports at most200. Do not silently truncate a larger
	// shadow-compensated prefix allowed by the independent merge kernel.
	p, err := NewRunMergePlan(manifests, k, 200)
	if err != nil {
		return nil, err
	}
	return &ImmutableRunSearch{runs: append([]*ImmutableRun(nil), runs...), plan: p, dimension: dim}, nil
}

func (s *ImmutableRunSearch) Search(ctx context.Context, q []float32) ([]Candidate, error) {
	if len(q) != s.dimension {
		return nil, errors.New("invalid query dimension")
	}
	var norm float64
	for _, v := range q {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return nil, errors.New("invalid query vector")
		}
		norm += float64(v) * float64(v)
	}
	if norm == 0 {
		return nil, errors.New("zero query")
	}
	prefixes := make([][]Candidate, len(s.runs))
	for i, r := range s.runs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		limit := s.plan.limits[i]
		if limit == 0 {
			// Empty/tombstone-only runs still participate in lifecycle checks.
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
		prefixes[i] = p
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.plan.Merge(prefixes)
}
