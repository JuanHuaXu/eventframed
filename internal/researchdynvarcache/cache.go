package researchdynvarcache

import (
	"errors"
	"math"
)

// Derived objects are private, updated only with their source belief, and never
// lazily mutated by queries. A pending issue still changes the next-rate law.
type derived struct {
	next       [States]vector
	noise      [States]float64
	individual [States]float64
	marginal   [3]float64
}
type odds struct {
	joint     [States]float64
	finiteSum [States]float64
	zero      [States]int
}

func (m *Model) derive(i int, b belief) (out derived, err error) {
	for k := range b.p {
		out.next[k], _, err = normalize(m.transition(i, k/3, b.p[k]))
		if err != nil {
			return out, err
		}
	}
	for a := 0; a < 3; a++ {
		w, z, e := noiseWeights([3]float64{b.log[3*a], b.log[3*a+1], b.log[3*a+2]})
		if e != nil {
			return out, e
		}
		copy(out.noise[3*a:3*a+3], w[:])
		out.marginal[a] = z
	}
	var prior [States]float64
	for k := range prior {
		prior[k] = m.familyPrior[k/3] * noisePrior[k%3]
	}
	var z float64
	out.individual, z, err = weights(b.log, prior)
	if err == nil && math.IsInf(z, -1) {
		err = errors.New("cached unsupported individual")
	}
	return
}

// This cheap guard retains numeric fault detection on active source evidence.
// It is not a memory-tamper certificate: arbitrary private-array mutation is
// outside the public API, and restored state would need a full revalidation.
func (m *Model) validLogs(i int, b *belief) error {
	for j := range m.members {
		logs := &m.members[j].latest.log
		if j == i && b != nil {
			logs = &b.log
		}
		for k, x := range logs {
			if m.familyPrior[k/3] > 0 && (math.IsNaN(x) || math.IsInf(x, 1)) {
				return errors.New("cached invalid active evidence")
			}
		}
	}
	return nil
}

// Rebuild at publication in the original member order. No infinity subtraction
// or irreversible evidence pruning is used; zero support has an explicit count.
func (m *Model) rebuildOdds(i int, replacement *belief) (out odds, err error) {
	if err = m.validLogs(i, replacement); err != nil {
		return out, err
	}
	if m.cfg.Mode == "individual" {
		return out, nil
	}
	var logs, prior [States]float64
	if m.cfg.Mode == "noise" {
		for j := range m.members {
			logs := &m.members[j].latest.log
			if j == i && replacement != nil {
				logs = &replacement.log
			}
			for k, v := range logs {
				if math.IsInf(v, -1) {
					out.zero[k]++
				} else {
					out.finiteSum[k] += v
				}
			}
		}
		for k := range prior {
			prior[k] = m.familyPrior[k/3] * noisePrior[k%3]
		}
	} else {
		for a, p := range m.familyPrior {
			if p == 0 {
				continue
			}
			for j := range m.members {
				z := m.members[j].derived.marginal[a]
				if j == i && replacement != nil {
					_, z, err = noiseWeights([3]float64{replacement.log[3*a], replacement.log[3*a+1], replacement.log[3*a+2]})
					if err != nil {
						return out, err
					}
				}
				if math.IsInf(z, -1) {
					return out, errors.New("cached unsupported member")
				}
				out.finiteSum[3*a] += z
			}
			prior[3*a] = p
		}
	}
	for k, z := range out.finiteSum {
		logs[k] = z
		if out.zero[k] > 0 {
			logs[k] = math.Inf(-1)
		}
	}
	var z float64
	out.joint, z, err = weights(logs, prior)
	if err == nil && math.IsInf(z, -1) {
		err = errors.New("cached unsupported model")
	}
	return
}
