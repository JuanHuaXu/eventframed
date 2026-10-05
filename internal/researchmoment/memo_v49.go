package researchmoment

import "errors"

// NewMemoV49 is a research-only exact constructor ablation. Its bounded cache
// is discarded before returning; no evidence, rounded means, or shared state
// enters it. Runtime updates and BeginEpoch remain the original Model methods.
func NewMemoV49(base []float64, epoch uint64, pendingCap int, cfg Config) (*Model, error) {
	if len(base) < 2 || len(base) > 200 || epoch == 0 || pendingCap < 1 || pendingCap > len(base)*MaxTrials || (cfg.Family != "narrow" && cfg.Family != "rich") || (cfg.Prior != "density" && cfg.Prior != "moment") || !finite(cfg.Strength) || cfg.Strength <= 0 || cfg.Strength > 32 || !finite(cfg.Hazard) || cfg.Hazard < 0 || cfg.Hazard >= 1 {
		return nil, errors.New("invalid moment contract")
	}
	for _, b := range base {
		if !finite(b) || b < .25 || b > .925+1e-12 {
			return nil, errors.New("invalid moment baseline")
		}
	}
	h := 3
	if cfg.Family == "rich" {
		h = MaxFamilies
	}
	m := &Model{base: append([]float64(nil), base...), cfg: cfg, epoch: epoch, h: h, cap: pendingCap, issued: make([]int, len(base)), trials: make([]trial, len(base)*MaxTrials), prior: make([]float64, len(base)*h*Atoms), latest: make([]float64, len(base)*h*Atoms), logs: make([]float64, len(base)*h), weights: familyWeights(h)}
	// Strength and prior mode are fixed for this invocation, so the exact mean
	// is the complete remaining key. FIFO eviction can only cause recomputation.
	var cache [128]struct {
		mu float64
		p  [Atoms]float64
	}
	used, next := 0, 0
	for i, b := range base {
		for j := 0; j < h; j++ {
			mu := familyMean(b, i, len(base), j)
			var p [Atoms]float64
			found := false
			for k := 0; k < used; k++ {
				if cache[k].mu == mu {
					p, found = cache[k].p, true
					break
				}
			}
			if !found {
				var e error
				p, e = DiscretePrior(mu, cfg.Strength, cfg.Prior)
				if e != nil {
					return nil, e
				}
				cache[next].mu, cache[next].p = mu, p
				next = (next + 1) % len(cache)
				used = min(used+1, len(cache))
			}
			copy(m.prior[(i*h+j)*Atoms:], p[:])
		}
	}
	copy(m.latest, m.prior)
	return m, nil
}
