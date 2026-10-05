package researchblend

import (
	"errors"
	"math/rand"
)

// Crossfit is an immediate-feedback, single-owner research combiner. Its
// validation rows omit their own arrived label, not all later arrived labels.
// It is not a prequential ticket implementation or an AP sharing authority.
type Crossfit struct {
	child   *Model
	weights [3]float64
	broken  bool
}

func NewCrossfit(base, coordinates []float64) (*Crossfit, error) {
	child, err := New(base, coordinates)
	if err != nil {
		return nil, err
	}
	return &Crossfit{child: child, weights: [3]float64{.98, .01, .01}}, nil
}

func (m *Crossfit) Predict(i int) (float64, error) {
	if m.broken {
		return 0, errors.New("quarantined crossfit")
	}
	p, err := m.child.conditional(i)
	if err != nil {
		return 0, err
	}
	q := 0.
	for k, w := range m.weights {
		q += w * p[k]
	}
	return q, nil
}

func (m *Crossfit) Weights() [3]float64 { return m.weights }

func (m *Crossfit) trainingRow(i int) ([3]float64, error) {
	var row [3]float64
	if m.broken || i < 0 || i >= len(m.child.base) || !m.child.seen[i] {
		return row, errors.New("unavailable leave-one-out row")
	}
	row[0] = m.child.base[i]
	var err error
	if row[1], err = m.child.affine.TrainingLOO(i); err != nil {
		return row, err
	}
	row[2], err = m.child.partition.TrainingLOO(i)
	return row, err
}

func (m *Crossfit) fit() ([3]float64, error) {
	gram, response := [3][3]float64{}, [3]float64{.98, .01, .01}
	for i := range gram {
		gram[i][i] = 1
	}
	for i, seen := range m.child.seen {
		if !seen {
			continue
		}
		row, err := m.trainingRow(i)
		if err != nil {
			return [3]float64{}, err
		}
		y := 0.
		if m.child.useful[i] {
			y = 1
		}
		for j := range row {
			response[j] += row[j] * y
			for k := range row {
				gram[j][k] += row[j] * row[k]
			}
		}
	}
	return simplexRidge(gram, response)
}

func (m *Crossfit) Observe(i int, useful bool) error {
	if m.broken || i < 0 || i >= len(m.child.base) || m.child.seen[i] {
		return errors.New("unavailable or duplicate evidence")
	}
	if err := m.child.Observe(i, useful); err != nil {
		m.broken = true
		return err
	}
	weights, err := m.fit()
	if err != nil {
		m.broken = true
		return err
	}
	m.weights = weights
	return nil
}

// Adaptive nomination is intentionally unsupported: removing a likelihood
// does not remove that label's influence on later nomination decisions.
func (m *Crossfit) Select(policy string, rng *rand.Rand) (Selection, error) {
	if m.broken || (policy != "random" && policy != "stratified_random") {
		return Selection{Index: -1}, errors.New("unsupported crossfit nomination")
	}
	return m.child.Select(policy, rng)
}
