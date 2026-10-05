package researchregimeprotected

import "sort"

// Preserve existing model mass for fresh alternatives. This changes truncation
// and its conditional working law; it does not create evidence or an IS weight.
func protectedTop(parts []component, cfg Config, t int) []component {
	if cfg.Cap == 0 || len(parts) <= cfg.Cap {
		return parts
	}
	selected := make([]component, 0, cfg.Cap)
	rest := make([]component, 0, len(parts))
	if cfg.Reset > 0 {
		for h := 0; h < Classes; h++ {
			for _, c := range parts {
				if c.h == h && c.start == t {
					selected = append(selected, c)
					break
				}
			}
		}
	}
	for _, c := range parts {
		if cfg.Reset == 0 || c.start != t {
			rest = append(rest, c)
		}
	}
	sort.SliceStable(rest, func(i, j int) bool { return rest[i].weight > rest[j].weight })
	return append(selected, rest[:cfg.Cap-len(selected)]...)
}
