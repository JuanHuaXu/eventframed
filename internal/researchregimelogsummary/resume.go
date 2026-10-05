package researchregimelogsummary

import (
	"errors"
	"math"
)

const CacheStride = 64
const CacheCount = 4

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

// Cold and resumed inference share arithmetic, not a second pruning law.
func resume(base []float64, cfg Config, rows []Row, support [][]Key, start prefix, cache []prefix, collect bool) (Result, []prefix, error) {
	if e := validate(base, cfg, rows); e != nil {
		return Result{}, nil, e
	}
	if (support != nil && len(support) != len(rows)) || start.count < 0 || start.count > len(rows) || start.state.members != len(base) {
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
	var dynamic [][]Key
	if support == nil {
		dynamic = make([][]Key, 0, len(rows))
	}
	for t := start.count; t < len(rows); t++ {
		row := rows[t]
		next = next[:0]
		if cfg.Reset < 1 {
			for _, c := range cur {
				c.logWeight += math.Log1p(-cfg.Reset)
				c.odds[row.Member] = movedOdds(c.odds[row.Member], high(base[row.Member], c.h), cfg.Hazard)
				next = append(next, c)
			}
		}
		if cfg.Reset > 0 {
			for _, c := range prior {
				c.logWeight += math.Log(cfg.Reset)
				c.start = t
				next = append(next, c)
			}
		}
		minFactor, maxFactor := 1., 0.
		if row.First < 0 {
			maxFactor = 1
		}
		for j := range next {
			if row.First < 0 {
				continue
			} // An unavailable factor is exactly one.
			c := &next[j]
			low, hi := likelihood(atom(0), c.h, row), likelihood(atom(1), c.h, row)
			minFactor, maxFactor = math.Min(minFactor, math.Min(low, hi)), math.Max(maxFactor, math.Max(low, hi))
			l, h := logRates(c.odds[row.Member])
			ll, lh := math.Log(low), math.Log(hi)
			lf := logAdd(l+ll, h+lh)
			c.logWeight += lf
			if !math.IsInf(lf, -1) {
				c.odds[row.Member] += lh - ll
			}
		}
		z := logMass(next)
		if !finite(z) {
			return Result{}, nil, errors.New("regime evidence support")
		}
		r.addEvidence(z)
		for j := range next {
			next[j].logWeight -= z
		}
		e := (1 - cfg.Reset) * r.Envelope
		if e > 0 && row.First >= 0 {
			if minFactor == 0 {
				e = 1
			} else {
				e = math.Min(1, 2*maxFactor*e/minFactor)
			}
		}
		r.MaxComponents = max(r.MaxComponents, len(next))
		var keys []Key
		if support == nil {
			selected := protectedTop(next, cfg, t)
			keys = make([]Key, len(selected))
			for i, c := range selected {
				keys[i] = Key{c.h, c.start}
			}
			dynamic = append(dynamic, keys)
		} else {
			keys = support[t]
		}
		if len(keys) == 0 || (cfg.Cap > 0 && len(keys) > cfg.Cap) {
			return Result{}, nil, errors.New("regime support count")
		}
		clear(sources)
		clear(seen)
		for j, c := range next {
			sources[Key{c.h, c.start}] = j
		}
		cur = cur[:0]
		for _, k := range keys {
			if seen[k] || k.Class < 0 || k.Class >= Classes || k.Start < -1 || k.Start > t {
				return Result{}, nil, errors.New("regime support identity")
			}
			seen[k] = true
			j, ok := sources[k]
			if !ok {
				return Result{}, nil, errors.New("regime support path binding")
			}
			cur = append(cur, next[j])
		}
		kept := logMass(cur)
		if !finite(kept) {
			return Result{}, nil, errors.New("regime conditional evidence support")
		}
		r.addEvidence(kept)
		for j := range cur {
			cur[j].logWeight -= kept
		}
		d := 0.
		if len(cur) < len(next) {
			d = math.Max(0, -math.Expm1(kept))
		}
		r.Discard += d
		r.Envelope = math.Min(1, e+d)
		r.parts, r.Components = cur, len(cur)
		if collect && (t+1)%CacheStride == 0 {
			cache = appendPrefix(cache, t+1, r)
		}
	}
	r.parts, r.Components = cur, len(cur)
	r.Support = support
	if support == nil {
		r.Support = dynamic
	}
	r.Forecast = make([]float64, len(base))
	for i := range base {
		for _, c := range cur {
			p := (1-cfg.Hazard)*rate(c.odds[i]) + cfg.Hazard*high(base[i], c.h)
			r.Forecast[i] += (1 - cfg.Reset) * math.Exp(c.logWeight) * atomProbability(p)
		}
		for h, w := range classPrior {
			r.Forecast[i] += cfg.Reset * w * atomProbability(high(base[i], h))
		}
	}
	return r, cache, nil
}
func (m *Ledger) replay(rows []Row, support [][]Key, target int, collect bool) (Result, []prefix, error) {
	initialState, e := Run(m.base, m.cfg, nil)
	if e != nil {
		return Result{}, nil, e
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
	r, e := Run(m.base, m.cfg, nil)
	if e != nil {
		return Result{}, nil, e
	}
	return resume(m.base, m.cfg, rows, support, prefix{0, r}, nil, true)
}
