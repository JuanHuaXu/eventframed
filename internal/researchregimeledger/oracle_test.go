package researchregimeledger

import "math"

// A separate dense transition implementation: no run-length components, cache,
// candidate mean/likelihood helper, pruning or conditional-independence premise.
func oraclePrior(base []float64) []float64 {
	width := 1 << len(base)
	p := make([]float64, 9*width)
	for h := 0; h < 9; h++ {
		for mask := 0; mask < width; mask++ {
			w := []float64{.1, .8, .1}[h/3] * []float64{.8, .1, .1}[h%3]
			for i, b := range base {
				mu := b
				if h/3 == 1 {
					mu = (9*b - .5) / 8
				}
				if h/3 == 2 {
					mu = 1 - b
				}
				u := (math.Min(.98, math.Max(.02, mu)) - .02) / .96
				if mask&(1<<i) == 0 {
					u = 1 - u
				}
				w *= u
			}
			p[h*width+mask] = w
		}
	}
	return p
}

func oracleHigh(base float64, h int) float64 {
	mu := []float64{base, (9*base - .5) / 8, 1 - base}[h/3]
	return (math.Min(.98, math.Max(.02, mu)) - .02) / .96
}

func oracleTransition(from, to, member int, base []float64, cfg Config, prior []float64) float64 {
	width := 1 << len(base)
	value := cfg.Reset * prior[to]
	fh, fm, th, tm := from/width, from%width, to/width, to%width
	if fh == th && fm&^(1<<member) == tm&^(1<<member) {
		p := oracleHigh(base[member], th)
		if tm&(1<<member) == 0 {
			p = 1 - p
		}
		q := cfg.Hazard * p
		if fm == tm {
			q += 1 - cfg.Hazard
		}
		value += (1 - cfg.Reset) * q
	}
	return value
}

func oracleFactor(state, members int, row Row) float64 {
	if row.First < 0 {
		return 1
	}
	width := 1 << members
	e := []float64{0, .1, .2}[(state/width)%3]
	p := .02
	if state%width&(1<<row.Member) != 0 {
		p = .98
	}
	if row.Second < 0 {
		q := e + (1-2*e)*p
		if row.First == 0 {
			q = 1 - q
		}
		return q
	}
	a, b := e, e
	if row.First == 1 {
		a = 1 - e
	}
	if row.Second == 1 {
		b = 1 - e
	}
	return p*a*b + (1-p)*(1-a)*(1-b)
}

func conditionalProduct(p []float64, members int) []float64 {
	width := 1 << members
	out := make([]float64, len(p))
	for h := 0; h < 9; h++ {
		w := 0.
		hi := make([]float64, members)
		for mask := 0; mask < width; mask++ {
			q := p[h*width+mask]
			w += q
			for i := range hi {
				if mask&(1<<i) != 0 {
					hi[i] += q
				}
			}
		}
		if w == 0 {
			continue
		}
		for i := range hi {
			hi[i] /= w
		}
		for mask := 0; mask < width; mask++ {
			q := w
			for i, x := range hi {
				if mask&(1<<i) == 0 {
					x = 1 - x
				}
				q *= x
			}
			out[h*width+mask] = q
		}
	}
	return out
}

func dense(base []float64, cfg Config, rows []Row, project bool) ([]float64, float64) {
	prior := oraclePrior(base)
	p := append([]float64(nil), prior...)
	logZ := 0.
	for _, row := range rows {
		next := make([]float64, len(p))
		for from, q := range p {
			for to := range next {
				next[to] += q * oracleTransition(from, to, row.Member, base, cfg, prior)
			}
		}
		if project {
			next = conditionalProduct(next, len(base))
		}
		z := 0.
		for to := range next {
			next[to] *= oracleFactor(to, len(base), row)
			z += next[to]
		}
		for to := range next {
			next[to] /= z
		}
		logZ += math.Log(z)
		if project {
			next = conditionalProduct(next, len(base))
		}
		p = next
	}
	return p, logZ
}

// Brute latent-path enumeration is independent of the dense forward recursion.
func enumerate(base []float64, cfg Config, rows []Row) ([]float64, float64) {
	prior := oraclePrior(base)
	out := make([]float64, len(prior))
	var visit func(int, int, float64)
	visit = func(t, state int, w float64) {
		if t == len(rows) {
			out[state] += w
			return
		}
		for target := range out {
			q := w * oracleTransition(state, target, rows[t].Member, base, cfg, prior) * oracleFactor(target, len(base), rows[t])
			if q > 0 {
				visit(t+1, target, q)
			}
		}
	}
	for state, w := range prior {
		visit(0, state, w)
	}
	z := 0.
	for _, w := range out {
		z += w
	}
	for state := range out {
		out[state] /= z
	}
	return out, math.Log(z)
}

func denseForecast(base []float64, cfg Config, p []float64, member int) float64 {
	prior, width := oraclePrior(base), 1<<len(base)
	q := 0.
	for from, w := range p {
		for to := range p {
			rate := .02
			if to%width&(1<<member) != 0 {
				rate = .98
			}
			q += w * oracleTransition(from, to, member, base, cfg, prior) * rate
		}
	}
	return q
}
