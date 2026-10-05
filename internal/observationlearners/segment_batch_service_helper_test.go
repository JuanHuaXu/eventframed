package observationlearners

import (
	"context"
	"errors"
	"math"
	"sync/atomic"
)

// New labelled workload per group; precomputed reference vectors are verification
// oracles only, never supplied to the fitter or used as returned predictions.
func ResearchBatchChangingFixture(distinct, batch bool) (func(context.Context, int) (float64, error), func() (uint64, uint64, uint64), error) {
	n := 4
	if distinct {
		n = 64
	}
	keys := make([]transformWork, n)
	refs := make([]float64, n)
	for g := 0; g < n; g++ {
		keys[g] = fixtureWork()
		keys[g].History[271].Bits = uint16(g)
		keys[g].History[271].Outcome = g%2 == 0
		keys[g].History[271].Arrives = 255
		m, e := fitSegmentPosteriorContext(context.Background(), keys[g].Left, keys[g].Clock, keys[g].History[:272], 64, .01, .95)
		if e != nil {
			return nil, nil, e
		}
		refs[g] = m.predictions[17]
	}

	var checked atomic.Uint64
	process := func(ctx context.Context, id int) (float64, error) {
		if id < 0 || id >= 64 {
			return 0, errors.New("invalid request identity")
		}
		g := (id % 16) / 4
		if distinct {
			g = id
		}
		var p float64
		var e error
		if batch {
			var m *segmentPosterior
			m, e = fitSegmentPosteriorBatch(ctx, -16, 256, keys[g].History[:272], 64, .01, .95)
			if e == nil {
				p = m.predictions[17]
			}
		} else {
			var m *segmentPosterior
			m, e = fitSegmentPosteriorContext(ctx, -16, 256, keys[g].History[:272], 64, .01, .95)
			if e == nil {
				p = m.predictions[17]
			}
		}
		if e != nil {
			return 0, e
		}
		if math.IsNaN(p) || math.Abs(p-refs[g]) > 1e-12 {
			return 0, errors.New("wrong work forecast")
		}
		checked.Add(1)
		return p, nil
	}
	return process, func() (uint64, uint64, uint64) { return checked.Load(), 0, checked.Load() }, nil
}
