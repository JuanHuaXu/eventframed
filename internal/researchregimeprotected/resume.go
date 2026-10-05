package researchregimeprotected

import (
	"errors"
	"math"
)

const CacheStride = 64
const CacheCount = 4

// A prefix excludes row count; the contained component arrays are immutable.
type prefix struct {
	count int
	state Result
}

func appendPrefix(cache []prefix, count int, r Result) []prefix {
	r.parts = append([]component(nil), r.parts...)
	r.Support, r.Forecast = nil, nil
	if len(cache) == CacheCount {
		copy(cache, cache[1:])
		cache = cache[:CacheCount-1]
	}
	return append(cache, prefix{count, r})
}

// Resume preserves cold arithmetic order and every retained-mass normalizer.
// Temporary component buffers are reused; published states are never mutated.
func resume(base []float64, cfg Config, rows []Row, support [][]Key, start prefix, cache []prefix, collect bool) (Result, []prefix, error) {
	if err := validate(base, cfg, rows); err != nil {
		return Result{}, nil, err
	}
	if len(support) != len(rows) || start.count < 0 || start.count > len(rows) || start.state.members != len(base) {
		return Result{}, nil, errors.New("regime resume binding")
	}
	capacity := cfg.Cap + Classes
	if cfg.Cap == 0 {
		capacity = Classes * (len(rows) + 1)
	}
	cur := make([]component, len(start.state.parts), capacity)
	copy(cur, start.state.parts)
	next := make([]component, 0, capacity)
	prior := initial(base)
	sources := make(map[Key]int, capacity)
	seen := make(map[Key]bool, capacity)
	r := start.state
	r.Support, r.Forecast = nil, nil
	for t := start.count; t < len(rows); t++ {
		row := rows[t]
		next = next[:0]
		if cfg.Reset < 1 {
			for _, c := range cur {
				c.weight *= 1 - cfg.Reset
				c.hi[row.Member] = (1-cfg.Hazard)*c.hi[row.Member] + cfg.Hazard*high(base[row.Member], c.h)
				next = append(next, c)
			}
		}
		if cfg.Reset > 0 {
			for _, c := range prior {
				c.weight *= cfg.Reset
				c.start = t
				next = append(next, c)
			}
		}
		z, minFactor, maxFactor := 0., 1., 0.
		for j := range next {
			c := &next[j]
			low, hi := likelihood(atom(0), c.h, row), likelihood(atom(1), c.h, row)
			minFactor, maxFactor = math.Min(minFactor, math.Min(low, hi)), math.Max(maxFactor, math.Max(low, hi))
			f := (1-c.hi[row.Member])*low + c.hi[row.Member]*hi
			c.weight *= f
			if f > 0 {
				c.hi[row.Member] *= hi / f
			}
			z += c.weight
		}
		if !finite(z) || z <= 0 {
			return Result{}, nil, errors.New("regime evidence support")
		}
		r.LogEvidence += math.Log(z)
		for j := range next {
			next[j].weight /= z
		}
		e := (1 - cfg.Reset) * r.Envelope
		if e > 0 {
			if minFactor == 0 {
				e = 1
			} else {
				e = math.Min(1, 2*maxFactor*e/minFactor)
			}
		}
		r.MaxComponents = max(r.MaxComponents, len(next))
		keys := support[t]
		if len(keys) == 0 || (cfg.Cap > 0 && len(keys) > cfg.Cap) {
			return Result{}, nil, errors.New("regime support count")
		}
		clear(sources)
		clear(seen)
		for j, c := range next {
			sources[Key{c.h, c.start}] = j
		}
		cur = cur[:0]
		retained := 0.
		for _, key := range keys {
			if seen[key] || key.Class < 0 || key.Class >= Classes || key.Start < -1 || key.Start > t {
				return Result{}, nil, errors.New("regime support identity")
			}
			seen[key] = true
			j, ok := sources[key]
			if !ok {
				return Result{}, nil, errors.New("regime support path binding")
			}
			cur = append(cur, next[j])
			retained += next[j].weight
		}
		if retained <= 0 || !finite(retained) {
			return Result{}, nil, errors.New("regime conditional evidence support")
		}
		r.LogEvidence += math.Log(retained)
		for j := range cur {
			cur[j].weight /= retained
		}
		d := math.Max(0, 1-retained)
		r.Discard += d
		r.Envelope = math.Min(1, e+d)
		r.parts, r.Components = cur, len(cur)
		if collect && (t+1)%CacheStride == 0 {
			cache = appendPrefix(cache, t+1, r)
		}
	}
	r.parts, r.Components = cur, len(cur)
	r.Support = support // Private owned immutable masks; Snapshot deep-copies them.
	r.Forecast = make([]float64, len(base))
	for i := range base {
		for _, c := range cur {
			p := (1-cfg.Hazard)*c.hi[i] + cfg.Hazard*high(base[i], c.h)
			r.Forecast[i] += (1 - cfg.Reset) * c.weight * atomProbability(p)
		}
		for h, w := range classPrior {
			r.Forecast[i] += cfg.Reset * w * atomProbability(high(base[i], h))
		}
	}
	return r, cache, nil
}

// Changed row target may use a prefix with count<=target, never a prefix that
// already contains that row. Old observations outside the cache replay fully.
func (m *Ledger) replay(rows []Row, support [][]Key, target int, collect bool) (Result, []prefix, error) {
	initialState, err := Run(m.base, m.cfg, nil)
	if err != nil {
		return Result{}, nil, err
	}
	start := prefix{0, initialState}
	var kept []prefix
	if collect {
		kept = make([]prefix, 0, CacheCount)
	}
	for _, p := range m.cache {
		if p.count <= target {
			if p.count > start.count {
				start = p
			}
			if collect {
				kept = append(kept, p)
			}
		}
	}
	if target == len(m.rows) && len(rows) == len(m.rows)+1 {
		start = prefix{len(m.rows), m.state}
	}
	return resume(m.base, m.cfg, rows, support, start, kept, collect)
}

func (m *Ledger) rebuild(rows []Row, support [][]Key) (Result, []prefix, error) {
	r, err := Run(m.base, m.cfg, nil)
	if err != nil {
		return Result{}, nil, err
	}
	return resume(m.base, m.cfg, rows, support, prefix{0, r}, make([]prefix, 0, CacheCount), true)
}
