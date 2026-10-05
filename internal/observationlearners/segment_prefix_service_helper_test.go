package observationlearners

import (
	"context"
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"slices"
	"sync/atomic"
)

// Single-worker ownership as in the service shadow queue. Preparation is cold
// and performed inside the request context, never from the reference model.
func ResearchPrefixFixture(invalidate, usePrefix bool) (func(context.Context, int) (float64, error), func() (uint64, uint64, uint64), error) {
	keys := make([]transformWork, 64)
	refs := make([]float64, 64)
	for g := range keys {
		keys[g] = fixtureWork()
		keys[g].History[271].Bits = uint16(g)
		keys[g].History[271].Outcome = g%2 == 0
		keys[g].History[271].Arrives = 255
		if invalidate {
			keys[g].History[250].Bits = uint16(g)
		}
		m, e := fitSegmentPosteriorTable(context.Background(), -16, 256, keys[g].History[:272], 64, .01, .95)
		if e != nil {
			return nil, nil, e
		}
		refs[g] = m.predictions[17]
	}
	var prepared *segmentPrefix
	var preparations, reuses, checked atomic.Uint64
	process := func(ctx context.Context, id int) (float64, error) {
		if id < 0 || id >= 64 {
			return 0, errors.New("invalid fixture ID")
		}
		h := keys[id].History[:272]
		var m *segmentPosterior
		var e error
		if usePrefix {
			// Select eligible data from this request, not its reference prediction.
			samples := make([]observation.Sample, 0, 272)
			for _, p := range h {
				if p.Arrives >= 0 && p.Arrives <= 256 {
					samples = append(samples, observation.Sample{Bits: p.Bits, Outcome: p.Outcome})
				}
			}
			if len(samples) > 64 {
				samples = samples[len(samples)-64:]
			}
			prefix := samples[:len(samples)-1]
			if prepared == nil || !slices.Equal(prepared.samples, prefix) {
				next, err := prepareSegmentPrefix(ctx, prefix, .95)
				if err != nil {
					return 0, err
				}
				prepared = next
				preparations.Add(1)
			} else {
				reuses.Add(1)
			}
			m, e = fitSegmentPosteriorPrefix(ctx, prepared, -16, 256, h, 64, .01, .95)
		} else {
			m, e = fitSegmentPosteriorTable(ctx, -16, 256, h, 64, .01, .95)
		}
		if e != nil {
			return 0, e
		}
		p := m.predictions[17]
		if math.IsNaN(p) || math.Abs(p-refs[id]) > 1e-12 {
			return 0, errors.New("prefix wrong-history forecast")
		}
		checked.Add(1)
		return p, nil
	}
	return process, func() (uint64, uint64, uint64) { return preparations.Load(), reuses.Load(), checked.Load() }, nil
}
