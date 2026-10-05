package researchblend

import (
	"errors"
	"math"
	"math/rand"

	"github.com/JuanHuaXu/eventframed/internal/researchcalibration"
)

// Stack retains ordinary Bayesian child models but fits predictive mixture
// weights from sealed pre-outcome forecasts. These are NOT family posteriors.
type Stack struct {
	child        *Model
	weights      [3]float64
	gram         [3][3]float64
	response     [3]float64
	pending      []stackPending
	pendingCount int
	serial       uint64
	broken       bool
}

type stackPending struct {
	serial uint64
	p      [3]float64
}

// Ticket is an opaque single-owner evidence binding. Mutable/replayed ticket
// payloads cannot replace the predictions retained privately by the model.
type Ticket struct {
	owner    *Stack
	index    int
	serial   uint64
	forecast float64
}

func (t Ticket) Forecast() float64 { return t.forecast }

func NewStack(base, coordinates []float64) (*Stack, error) {
	child, err := New(base, coordinates)
	if err != nil {
		return nil, err
	}
	m := &Stack{child: child, weights: [3]float64{.98, .01, .01}, response: [3]float64{.98, .01, .01}, pending: make([]stackPending, len(base))}
	// Fixed lambda=1 ridge penalty toward the original weights. Gram updates
	// use SUM losses, so this regularizer is not repeated once per outcome.
	for i := range m.gram {
		m.gram[i][i] = 1
	}
	return m, nil
}

func (m *Stack) Predict(i int) (float64, error) {
	if m.broken {
		return 0, errors.New("quarantined stack")
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

func (m *Stack) Weights() [3]float64 { return m.weights }

func (m *Stack) Issue(i int) (Ticket, error) {
	if m.broken || i < 0 || i >= len(m.pending) || m.child.seen[i] || m.pending[i].serial != 0 {
		return Ticket{}, errors.New("unavailable or previously nominated member")
	}
	p, err := m.child.conditional(i)
	if err != nil {
		return Ticket{}, err
	}
	q := 0.
	for k, w := range m.weights {
		q += w * p[k]
	}
	m.serial++
	m.pending[i] = stackPending{m.serial, p}
	m.pendingCount++
	return Ticket{m, i, m.serial, q}, nil
}

func (m *Stack) Resolve(ticket Ticket, useful bool) error {
	if m.broken || ticket.owner != m || ticket.index < 0 || ticket.index >= len(m.pending) || ticket.serial == 0 {
		return errors.New("invalid evidence ticket")
	}
	i := ticket.index
	p := m.pending[i]
	if p.serial != ticket.serial || m.child.seen[i] {
		return errors.New("stale or duplicate evidence ticket")
	}
	gram, response := m.gram, m.response
	y := 0.
	if useful {
		y = 1
	}
	for a := 0; a < 3; a++ {
		response[a] += p.p[a] * y
		for b := 0; b < 3; b++ {
			gram[a][b] += p.p[a] * p.p[b]
		}
	}
	weights, err := simplexRidge(gram, response)
	if err != nil {
		m.broken = true
		return err
	}
	if err = m.child.Observe(i, useful); err != nil {
		m.broken = true
		return err
	}
	m.gram, m.response, m.weights = gram, response, weights
	m.pending[i] = stackPending{}
	m.pendingCount--
	return nil
}

// Observe is the immediate-feedback convenience path. Delayed feedback must
// resolve its original ticket rather than recalculating today's prediction.
func (m *Stack) Observe(i int, useful bool) error {
	ticket, err := m.Issue(i)
	if err != nil {
		return err
	}
	return m.Resolve(ticket, useful)
}

// Seven nonempty simplex faces cover all optima of this strictly convex QP.
// Each face solves its equality-constrained KKT system (at most 4x4).
func simplexRidge(a [3][3]float64, c [3]float64) ([3]float64, error) {
	best := [3]float64{}
	cost := math.Inf(1)
	for mask := 1; mask < 8; mask++ {
		var indices [3]int
		k := 0
		for i := 0; i < 3; i++ {
			if mask&(1<<i) != 0 {
				indices[k] = i
				k++
			}
		}
		var system [4][5]float64
		for i, member := range indices[:k] {
			for j, other := range indices[:k] {
				system[i][j] = 2 * a[member][other]
			}
			system[i][k] = 1
			system[i][k+1] = 2 * c[member]
			system[k][i] = 1
		}
		system[k][k+1] = 1
		solution, ok := solveSmall(system, k+1)
		if !ok {
			continue
		}
		w := [3]float64{}
		sum := 0.
		valid := true
		for i, member := range indices[:k] {
			if solution[i] < -1e-12 || math.IsNaN(solution[i]) || math.IsInf(solution[i], 0) {
				valid = false
				break
			}
			w[member] = math.Max(0, solution[i])
			sum += w[member]
		}
		if !valid || sum <= 0 {
			continue
		}
		for i := range w {
			w[i] /= sum
		}
		value := 0.
		for i := range w {
			value -= 2 * c[i] * w[i]
			for j := range w {
				value += w[i] * a[i][j] * w[j]
			}
		}
		if value < cost {
			best, cost = w, value
		}
	}
	if math.IsInf(cost, 0) || math.IsNaN(cost) {
		return best, errors.New("no feasible mixture")
	}
	return best, nil
}

func solveSmall(matrix [4][5]float64, n int) ([4]float64, bool) {
	for col := 0; col < n; col++ {
		pivot := col
		for row := col + 1; row < n; row++ {
			if math.Abs(matrix[row][col]) > math.Abs(matrix[pivot][col]) {
				pivot = row
			}
		}
		if math.Abs(matrix[pivot][col]) < 1e-14 {
			return [4]float64{}, false
		}
		matrix[col], matrix[pivot] = matrix[pivot], matrix[col]
		v := matrix[col][col]
		for j := col; j <= n; j++ {
			matrix[col][j] /= v
		}
		for row := 0; row < n; row++ {
			if row == col {
				continue
			}
			v = matrix[row][col]
			for j := col; j <= n; j++ {
				matrix[row][j] -= v * matrix[col][j]
			}
		}
	}
	var result [4]float64
	for i := 0; i < n; i++ {
		result[i] = matrix[i][n]
	}
	return result, true
}

func (m *Stack) eligible(i int) bool { return !m.child.seen[i] && m.pending[i].serial == 0 }
func (m *Stack) randomIn(lo, hi int, rng *rand.Rand) (int, float64) {
	count := 0
	for i := lo; i < hi; i++ {
		if m.eligible(i) {
			count++
		}
	}
	if count == 0 {
		return -1, 0
	}
	ordinal := rng.Intn(count)
	for i := lo; i < hi; i++ {
		if m.eligible(i) {
			if ordinal == 0 {
				return i, 1 / float64(count)
			}
			ordinal--
		}
	}
	panic("inconsistent eligibility")
}

// Disagreement is a heuristic spread of child forecasts under predictive
// weights, NOT mutual information under a family posterior or an AP test.
func (m *Stack) Select(policy string, rng *rand.Rand) (Selection, error) {
	bad := Selection{Index: -1}
	count := m.child.count + m.pendingCount
	n := len(m.pending)
	if m.broken || m.child.broken || count == n {
		return bad, errors.New("unavailable or exhausted candidates")
	}
	if policy != "random" && policy != "stratified_random" && policy != "uncertainty" && policy != "disagreement" {
		return bad, errors.New("unknown stack policy")
	}
	if policy != "uncertainty" && rng == nil {
		return bad, errors.New("missing RNG")
	}
	if policy == "random" {
		i, p := m.randomIn(0, n, rng)
		return Selection{Index: i, Probability: p}, nil
	}
	if policy == "stratified_random" {
		buckets := len(m.child.order)
		for attempt := 0; attempt < buckets; attempt++ {
			bucket := m.child.order[(count+attempt)%buckets]
			if i, p := m.randomIn(bucket*n/buckets, (bucket+1)*n/buckets, rng); i >= 0 {
				return Selection{Index: i, Probability: p}, nil
			}
		}
	}
	best, score := -1, math.Inf(-1)
	for i := 0; i < n; i++ {
		if !m.eligible(i) {
			continue
		}
		p, err := m.child.conditional(i)
		if err != nil {
			return bad, err
		}
		q := 0.
		for k, w := range m.weights {
			q += w * p[k]
		}
		value := researchcalibration.Entropy(q)
		if policy == "disagreement" {
			value = 0
			for k, w := range m.weights {
				value += w * (p[k] - q) * (p[k] - q)
			}
		}
		if value > score {
			best, score = i, value
		}
	}
	chosen, probability := best, 1.
	if policy == "disagreement" {
		if rng.Float64() < .2 {
			chosen, _ = m.randomIn(0, n, rng)
		}
		probability = .2 / float64(n-count)
		if chosen == best {
			probability += .8
		}
	}
	return Selection{Index: chosen, Probability: probability, Score: score}, nil
}
