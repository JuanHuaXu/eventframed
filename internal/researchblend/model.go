// Package researchblend averages bounded alternative joint models. It is
// research-only: posterior weights confer no evidence or sharing authority.
package researchblend

import (
	"errors"
	"math"
	"math/rand"

	"github.com/JuanHuaXu/eventframed/internal/researchcalibration"
	"github.com/JuanHuaXu/eventframed/internal/researchpartition"
)

type Model struct {
	base         []float64
	seen, useful []bool
	affine       *researchcalibration.Model
	partition    *researchpartition.Model
	weights      [3]float64
	order        []int
	count        int
	broken       bool
}

type Selection struct {
	Index       int
	Probability float64 // Conditional nomination probability given past labels.
	Score       float64
}

// Priors are frozen, not chosen from the new experiment's outcomes. The
// component families retain their own latent-variable priors and likelihoods.
func New(base, coordinates []float64) (*Model, error) {
	a, err := researchcalibration.New(base)
	if err != nil {
		return nil, err
	}
	p, err := researchpartition.New(base, coordinates)
	if err != nil {
		return nil, err
	}
	m := &Model{base: append([]float64(nil), base...), seen: make([]bool, len(base)), useful: make([]bool, len(base)), affine: a, partition: p, weights: [3]float64{.98, .01, .01}}
	buckets := min(32, len(base))
	size, bits := 1, 0
	for size < buckets {
		size *= 2
		bits++
	}
	for k := 0; k < size; k++ {
		j, x := 0, k
		for bit := 0; bit < bits; bit++ {
			j = (j << 1) | (x & 1)
			x >>= 1
		}
		if j < buckets {
			m.order = append(m.order, j)
		}
	}
	return m, nil
}

func (m *Model) conditional(i int) ([3]float64, error) {
	var p [3]float64
	if m.broken || i < 0 || i >= len(m.base) {
		return p, errors.New("unavailable model or member")
	}
	p[0] = m.base[i]
	if m.seen[i] {
		y := 0.0
		if m.useful[i] {
			y = 1
		}
		p[0] = (2*p[0] + y) / 3
	}
	var err error
	if p[1], err = m.affine.Predict(i); err != nil {
		return p, err
	}
	if p[2], err = m.partition.Predict(i); err != nil {
		return p, err
	}
	return p, nil
}

func (m *Model) Predict(i int) (float64, error) {
	p, err := m.conditional(i)
	if err != nil {
		return 0, err
	}
	q := 0.0
	for k, w := range m.weights {
		q += w * p[k]
	}
	return q, nil
}

// One label updates alternative models, not three independent copies of
// evidence. The outer Bayes factor uses each child's pre-label predictive.
// This implements one joint model with a latent family index.
func (m *Model) Observe(i int, useful bool) error {
	if m.broken || i < 0 || i >= len(m.base) || m.seen[i] {
		return errors.New("unavailable or duplicate evidence")
	}
	p, err := m.conditional(i)
	if err != nil {
		return err
	}
	var next [3]float64
	maximum := math.Inf(-1)
	for k, w := range m.weights {
		likelihood := p[k]
		if !useful {
			likelihood = 1 - likelihood
		}
		next[k] = math.Log(w) + math.Log(likelihood)
		maximum = math.Max(maximum, next[k])
	}
	mass := 0.0
	for k := range next {
		next[k] = math.Exp(next[k] - maximum)
		mass += next[k]
	}
	if mass <= 0 || math.IsNaN(mass) || math.IsInf(mass, 0) {
		return errors.New("invalid posterior mass")
	}
	for k := range next {
		next[k] /= mass
	}
	// Valid child states accept the same unseen index. Unexpected partial
	// failure quarantines this single-owner object; no partial law is served.
	if err = m.affine.Observe(i, useful); err != nil {
		m.broken = true
		return err
	}
	if err = m.partition.Observe(i, useful); err != nil {
		m.broken = true
		return err
	}
	m.weights = next
	m.seen[i], m.useful[i] = true, useful
	m.count++
	return nil
}

func (m *Model) Weights() [3]float64 { return m.weights }

func (m *Model) randomIn(lo, hi int, rng *rand.Rand) (int, float64) {
	left := 0
	for i := lo; i < hi; i++ {
		if !m.seen[i] {
			left++
		}
	}
	if left == 0 {
		return -1, 0
	}
	j := rng.Intn(left)
	for i := lo; i < hi; i++ {
		if !m.seen[i] {
			if j == 0 {
				return i, 1 / float64(left)
			}
			j--
		}
	}
	panic("invalid selection count")
}

// The falsification score discriminates the three family indices, not all
// hypotheses inside each family. A .2 uniform exploration floor preserves
// positive conditional coverage; it does not establish model adequacy.
func (m *Model) Select(policy string, rng *rand.Rand) (Selection, error) {
	bad := Selection{Index: -1}
	if m.broken || m.count == len(m.base) {
		return bad, errors.New("unavailable or exhausted frontier")
	}
	if policy != "random" && policy != "stratified_random" && policy != "uncertainty" && policy != "family_information" {
		return bad, errors.New("unknown policy")
	}
	if policy != "uncertainty" && rng == nil {
		return bad, errors.New("missing RNG")
	}
	if policy == "random" {
		i, probability := m.randomIn(0, len(m.base), rng)
		return Selection{Index: i, Probability: probability}, nil
	}
	if policy == "stratified_random" {
		buckets := len(m.order)
		for attempt := 0; attempt < buckets; attempt++ {
			bucket := m.order[(m.count+attempt)%buckets]
			lo, hi := bucket*len(m.base)/buckets, (bucket+1)*len(m.base)/buckets
			if i, probability := m.randomIn(lo, hi, rng); i >= 0 {
				return Selection{Index: i, Probability: probability}, nil
			}
		}
	}
	best, score := -1, math.Inf(-1)
	for i := range m.base {
		if m.seen[i] {
			continue
		}
		p, err := m.conditional(i)
		if err != nil {
			return bad, err
		}
		q := 0.0
		for k, w := range m.weights {
			q += w * p[k]
		}
		value := researchcalibration.Entropy(q)
		if policy == "family_information" {
			for k, w := range m.weights {
				value -= w * researchcalibration.Entropy(p[k])
			}
			value = math.Max(0, value)
		}
		if value > score {
			best, score = i, value
		}
	}
	probability := 1.0
	chosen := best
	if policy == "family_information" {
		if rng.Float64() < .2 {
			chosen, _ = m.randomIn(0, len(m.base), rng)
		}
		probability = .2 / float64(len(m.base)-m.count)
		if chosen == best {
			probability += .8
		}
	}
	return Selection{Index: chosen, Probability: probability, Score: score}, nil
}
