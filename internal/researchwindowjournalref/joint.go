package researchwindowjournalref

import "errors"

// The reference's top-down forecast recursion and independently reconstructed
// local posteriors produce this joint; no candidate kernel is consulted.
func (r *Reference) LatentJoint(i int) ([8]float64, error) {
	var out [8]float64
	if i < 0 || i >= len(r.base) {
		return out, errors.New("reference joint member")
	}
	for h, w := range r.weights {
		if w == 0 {
			continue
		}
		var one [3]float64
		one[h] = 1
		mean, _ := r.forecast(i, &r.nodes, one)
		eta := float64(h) / 10
		for y := 0; y < 2; y++ {
			for a := 0; a < 2; a++ {
				for b := 0; b < 2; b++ {
					p := mean
					if y == 0 {
						p = 1 - mean
					}
					p *= w
					if a == y {
						p *= 1 - eta
					} else {
						p *= eta
					}
					if b == y {
						p *= 1 - eta
					} else {
						p *= eta
					}
					out[4*y+2*a+b] += p
				}
			}
		}
	}
	return out, nil
}
