// Package researchpartition is an isolated finite Bayesian partition model.
// Its group hypotheses confer no provenance or Anti-Pigeon sharing authority.
package researchpartition

import (
	"errors"
	"math"
	"math/rand"
)

const Trees = 26

type partition struct {
	node  [8]uint8
	prior float64
}
type Model struct {
	base                []float64
	bin                 []uint8
	seen, useful        []bool
	partitions          [Trees]partition
	weights             [Trees + 1]float64
	successes, failures [15]uint64
	count               int
}

func enumerate(lo, hi, depth, node int) []partition {
	p := partition{prior: 1}
	for b := lo; b < hi; b++ {
		p.node[b] = uint8(node)
	}
	if depth == 3 {
		return []partition{p}
	}
	p.prior = .5
	out := []partition{p}
	mid := (lo + hi) / 2
	for _, left := range enumerate(lo, mid, depth+1, node*2+1) {
		for _, right := range enumerate(mid, hi, depth+1, node*2+2) {
			q := partition{prior: .5 * left.prior * right.prior}
			for b := lo; b < mid; b++ {
				q.node[b] = left.node[b]
			}
			for b := mid; b < hi; b++ {
				q.node[b] = right.node[b]
			}
			out = append(out, q)
		}
	}
	return out
}

// Coordinates are supplied before evidence, not learned from hidden labels.
// The cold prior is .99*baseline+.01*.5, not an exact identity baseline.
func New(base, coordinates []float64) (*Model, error) {
	if len(base) < 2 || len(base) > 200 || len(base) != len(coordinates) {
		return nil, errors.New("unsupported frontier")
	}
	m := &Model{base: append([]float64(nil), base...), bin: make([]uint8, len(base)), seen: make([]bool, len(base)), useful: make([]bool, len(base))}
	for i, b := range base {
		r := coordinates[i]
		if math.IsNaN(b) || math.IsInf(b, 0) || b <= 0 || b >= 1 || math.IsNaN(r) || math.IsInf(r, 0) || r < 0 || r > 1 {
			return nil, errors.New("invalid baseline or coordinate")
		}
		m.bin[i] = uint8(math.Min(7, math.Floor(8*r)))
	}
	parts := enumerate(0, 8, 0, 0)
	if len(parts) != Trees {
		return nil, errors.New("invalid partition family")
	}
	copy(m.partitions[:], parts)
	m.weights[0] = .99
	for k, p := range parts {
		m.weights[k+1] = .01 * p.prior
	}
	return m, nil
}

func (m *Model) conditional(i, k int) float64 {
	p := m.base[i]
	if k > 0 {
		node := m.partitions[k-1].node[m.bin[i]]
		p = (1 + float64(m.successes[node])) / (2 + float64(m.successes[node]+m.failures[node]))
	}
	if m.seen[i] {
		y := 0.0
		if m.useful[i] {
			y = 1
		}
		p = (2*p + y) / 3
	}
	return p
}
func (m *Model) Predict(i int) (float64, error) {
	if i < 0 || i >= len(m.base) {
		return 0, errors.New("unknown member")
	}
	q := 0.0
	for k, w := range m.weights {
		q += w * m.conditional(i, k)
	}
	return q, nil
}

// Given a partition, psi_leaf~Beta(1,1), phi_i|psi~Beta(2*psi,2*(1-psi)),
// and training/future labels share phi_i. With one training label per member,
// integrating phi_i gives a Bernoulli(psi_leaf) likelihood. The posterior
// future mean for an observed member is (2*E[psi_leaf|all labels]+y_i)/3.
func (m *Model) Observe(i int, useful bool) error {
	if i < 0 || i >= len(m.base) || m.seen[i] {
		return errors.New("unknown or duplicate evidence")
	}
	var logs [Trees + 1]float64
	maximum := math.Inf(-1)
	for k, w := range m.weights {
		p := m.conditional(i, k)
		if !useful {
			p = 1 - p
		}
		logs[k] = math.Log(w) + math.Log(p)
		maximum = math.Max(maximum, logs[k])
	}
	mass := 0.0
	for k := range logs {
		logs[k] = math.Exp(logs[k] - maximum)
		mass += logs[k]
	}
	if mass <= 0 || math.IsNaN(mass) || math.IsInf(mass, 0) {
		return errors.New("invalid posterior mass")
	}
	for k, v := range logs {
		m.weights[k] = v / mass
	}
	node := int(m.bin[i]) + 7
	for {
		if useful {
			m.successes[node]++
		} else {
			m.failures[node]++
		}
		if node == 0 {
			break
		}
		node = (node - 1) / 2
	}
	m.seen[i], m.useful[i] = true, useful
	m.count++
	return nil
}

func (m *Model) Select(policy string, rng *rand.Rand) (int, error) {
	if m.count == len(m.base) {
		return -1, errors.New("frontier exhausted")
	}
	if policy == "random" {
		if rng == nil {
			return -1, errors.New("missing RNG")
		}
		ordinal := rng.Intn(len(m.base) - m.count)
		for i, seen := range m.seen {
			if !seen {
				if ordinal == 0 {
					return i, nil
				}
				ordinal--
			}
		}
	}
	if policy == "stratified" {
		size, bits := 1, 0
		for size < len(m.base) {
			size *= 2
			bits++
		}
		for k := 0; k < size; k++ {
			index, x := 0, k
			for j := 0; j < bits; j++ {
				index = (index << 1) | (x & 1)
				x >>= 1
			}
			if index < len(m.base) && !m.seen[index] {
				return index, nil
			}
		}
	}
	if policy != "head" && policy != "uncertainty" {
		return -1, errors.New("unsupported policy")
	}
	best, value := -1, math.Inf(-1)
	for i, seen := range m.seen {
		if seen {
			continue
		}
		if policy == "head" {
			return i, nil
		}
		q, _ := m.Predict(i)
		v := -q*math.Log(q) - (1-q)*math.Log1p(-q)
		if v > value {
			best, value = i, v
		}
	}
	return best, nil
}
