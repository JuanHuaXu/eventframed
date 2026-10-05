package researchbrier

import (
	"math"
	"reflect"
	"testing"
)

// Independent general two-outcome positive-part substitution, with linear
// normalized masses and a root search rather than the binary closed form.
func reference(logs, advice []float64) (float64, [2]float64) {
	m := logs[0]
	for _, v := range logs {
		m = math.Max(m, v)
	}
	w := make([]float64, len(logs))
	sum := 0.
	for j, v := range logs {
		w[j] = math.Exp(v - m)
		sum += w[j]
	}
	var g [2]float64
	for y := 0; y < 2; y++ {
		mass := 0.
		for j, p := range advice {
			d := p - float64(y)
			mass += w[j] / sum * math.Exp(-2*d*d)
		}
		g[y] = -math.Log(mass)
	}
	lo, hi := math.Min(g[0], g[1])-2, math.Max(g[0], g[1])+2
	for j := 0; j < 96; j++ {
		s := (lo + hi) / 2
		if math.Max(s-g[0], 0)+math.Max(s-g[1], 0) < 2 {
			lo = s
		} else {
			hi = s
		}
	}
	return math.Max(((lo+hi)/2)-g[1], 0) / 2, g
}

func TestSubstitutionReferenceAndDomination(t *testing.T) {
	checks := 0
	for n := 1; n <= 8; n++ {
		for r := 0; r < 128; r++ {
			logs, p := make([]float64, n), make([]float64, n)
			for j := range p {
				logs[j] = -float64((r*37 + j*23) % 1000)
				p[j] = float64((r*17+j*31)%101) / 100
			}
			q, e := Forecast(logs, p)
			if e != nil {
				t.Fatal(e)
			}
			want, g := reference(logs, p)
			if math.Abs(q-want) > 3e-14 {
				t.Fatal("substitution mismatch", q, want)
			}
			for y := 0; y < 2; y++ {
				if 2*(q-float64(y))*(q-float64(y)) > g[y]+3e-14 {
					t.Fatal("domination failed")
				}
			}
			mirror := make([]float64, n)
			for j := range mirror {
				mirror[j] = 1 - p[j]
			}
			v, e := Forecast(logs, mirror)
			if e != nil || math.Abs(q+v-1) > 3e-14 {
				t.Fatal("symmetry", e)
			}
			checks++
		}
	}
	for _, p := range [][]float64{{0, 0}, {1, 1}, {0, 1}, {1e-300, 2e-300}, {math.Nextafter(1, 0), 1}} {
		q, e := Forecast([]float64{-1e4, 0}, p)
		if e != nil || q < math.Min(p[0], p[1]) || q > math.Max(p[0], p[1]) {
			t.Fatal("endpoint", q, e)
		}
	}
	t.Log("independent substitution/domain/pointwise domination cases", checks)
}

func TestExhaustiveImmediateRegret(t *testing.T) {
	const steps = 12
	for _, prior := range [][]float64{{1, 1}, {.8, .1, .1}, {.01, .99}} {
		for mask := 0; mask < 1<<steps; mask++ {
			m, e := New(prior, 1, steps)
			if e != nil {
				t.Fatal(e)
			}
			expert := make([]float64, len(prior))
			learner := 0.
			past := 0
			for j := 0; j < steps; j++ {
				advice := make([]float64, len(prior))
				for k := range advice {
					advice[k] = float64((j*3+k*7+past)%11) / 10
				}
				x, e := m.Issue(advice, int64(j*2))
				if e != nil {
					t.Fatal(e)
				}
				y := float64((mask >> j) & 1)
				_, e = m.Resolve(x, y == 1, int64(j*2+1))
				if e != nil {
					t.Fatal(e)
				}
				past += int(y)
				d := x.Forecast() - y
				learner += d * d
				for k, p := range advice {
					d := p - y
					expert[k] += d * d
					if learner > expert[k]-.5*m.prior[k]+3e-12 {
						t.Fatal("prefix regret bound failed", mask, j, k, learner, expert[k], m.prior[k])
					}
				}
			}
		}
	}
	t.Log("all12-bit paths under3priors, adaptive past-only advice and every prefix passed")
}

func TestLifecycleAtomicityAndOwnedAdvice(t *testing.T) {
	m, e := New([]float64{.8, .2}, 1, 2)
	if e != nil {
		t.Fatal(e)
	}
	other, e := New([]float64{.8, .2}, 1, 2)
	if e != nil {
		t.Fatal(e)
	}
	p := []float64{.1, .9}
	x, e := m.Issue(p, 1)
	if e != nil {
		t.Fatal(e)
	}
	p[0] = 1
	before := *m
	if _, e = m.Issue([]float64{.2, .8}, 2); e == nil || !reflect.DeepEqual(before, *m) {
		t.Fatal("pending admission mutated")
	}
	y, e := other.Issue([]float64{.1, .9}, 1)
	if e != nil {
		t.Fatal(e)
	}
	for _, bad := range []Ticket{y, {owner: m, epoch: 2, ordinal: 1, q: x.q}, {owner: m, epoch: 1, ordinal: 2, q: x.q}, {owner: m, epoch: 1, ordinal: 1, q: x.q + .1}} {
		if _, e = m.Resolve(bad, true, 2); e == nil || !reflect.DeepEqual(before, *m) {
			t.Fatal("invalid handle mutated")
		}
	}
	if _, e = m.Resolve(x, true, 0); e == nil || !reflect.DeepEqual(before, *m) {
		t.Fatal("time rollback mutated")
	}
	r, e := m.Resolve(x, true, 2)
	if e != nil || r.Forecast != x.q || r.TrialOrdinal != 1 {
		t.Fatal("served receipt", e)
	}
	_, e = other.Resolve(y, true, 2)
	if e != nil || !reflect.DeepEqual(m.logs, other.logs) {
		t.Fatal("advice aliased", e)
	}
	before = *m
	if _, e = m.Resolve(x, true, 3); e == nil || !reflect.DeepEqual(before, *m) {
		t.Fatal("duplicate mutated")
	}
	if e = m.BeginEpoch(1, 3); e == nil || !reflect.DeepEqual(before, *m) {
		t.Fatal("same epoch mutated")
	}
	if e = m.BeginEpoch(2, 3); e != nil {
		t.Fatal(e)
	}
	if _, e = m.Resolve(x, false, 4); e == nil {
		t.Fatal("stale epoch accepted")
	}
	if m.used != 0 || m.pending || m.logs != m.prior {
		t.Fatal("epoch state")
	}
}

func TestInvalidAndRevival(t *testing.T) {
	for _, p := range [][]float64{nil, {0}, {-1}, {math.NaN()}, {math.Inf(1)}, make([]float64, 9)} {
		if _, e := New(p, 1, 1); e == nil {
			t.Fatal("invalid prior accepted")
		}
	}
	for _, c := range []int{0, 4097} {
		if _, e := New([]float64{1}, 1, c); e == nil {
			t.Fatal("invalid cap")
		}
	}
	if _, e := New([]float64{1}, 0, 1); e == nil {
		t.Fatal("zero epoch")
	}
	for _, p := range [][]float64{nil, {-.01}, {1.01}, {math.NaN()}, {math.Inf(-1)}} {
		if _, e := Forecast([]float64{0}, p); e == nil {
			t.Fatal("invalid advice")
		}
	}
	for _, w := range [][]float64{nil, {math.NaN()}, {math.Inf(-1)}} {
		if _, e := Forecast(w, []float64{.5}); e == nil {
			t.Fatal("invalid weight")
		}
	}
	m, e := New([]float64{1, 1}, 1, 2048)
	if e != nil {
		t.Fatal(e)
	}
	for j := 0; j < 2048; j++ {
		x, e := m.Issue([]float64{0, 1}, int64(j*2))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = m.Resolve(x, j < 1024, int64(j*2+1)); e != nil {
			t.Fatal(e)
		}
	}
	q, e := m.Predict([]float64{0, 1})
	if e != nil || math.Abs(q-.5) > 2e-11 {
		t.Fatal("underflow revival", q, e)
	}
	before := *m
	if _, e = m.Issue([]float64{0, 1}, 4096); e == nil || !reflect.DeepEqual(before, *m) {
		t.Fatal("cap overflow")
	}
}

func TestCommonOffsetAndExtremePriors(t *testing.T) {
	advice := []float64{.1, .45, .95}
	base, err := Forecast([]float64{-9, -3, 0}, advice)
	if err != nil {
		t.Fatal(err)
	}
	for _, offset := range []float64{-1024, 1024, 1 << 40, -(1 << 40)} {
		got, err := Forecast([]float64{offset - 9, offset - 3, offset}, advice)
		if err != nil || got != base {
			t.Fatal("common offset changed forecast", offset, got, base, err)
		}
	}
	for _, prior := range [][]float64{{math.MaxFloat64, math.MaxFloat64}, {math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64}, {math.SmallestNonzeroFloat64, math.MaxFloat64}} {
		m, err := New(prior, 1, 1)
		if err != nil {
			t.Fatal("finite prior failed", err)
		}
		q, err := m.Predict([]float64{0, 1})
		if err != nil || !finite(q) || q < 0 || q > 1 {
			t.Fatal("extreme prior forecast", q, err)
		}
		if prior[0] == prior[1] && math.Abs(q-.5) > 3e-14 {
			t.Fatal("equal extreme prior symmetry", q)
		}
	}
}

func TestPredictionCannotReadUnrevealedOutcomes(t *testing.T) {
	for fork := 0; fork < 16; fork++ {
		a, err := New([]float64{.8, .1, .1}, 1, 16)
		if err != nil {
			t.Fatal(err)
		}
		b, err := New([]float64{.8, .1, .1}, 1, 16)
		if err != nil {
			t.Fatal(err)
		}
		for j := 0; j <= fork; j++ {
			advice := []float64{float64(j%7) / 7, .5, float64(j%5) / 5}
			x, err := a.Issue(advice, int64(2*j))
			if err != nil {
				t.Fatal(err)
			}
			y, err := b.Issue(advice, int64(2*j))
			if err != nil || x.Forecast() != y.Forecast() {
				t.Fatal("unrevealed fork affected issued law", fork, j, err)
			}
			label := j%2 == 0
			if _, err = a.Resolve(x, label, int64(2*j+1)); err != nil {
				t.Fatal(err)
			}
			if j == fork {
				label = !label
			}
			if _, err = b.Resolve(y, label, int64(2*j+1)); err != nil {
				t.Fatal(err)
			}
		}
	}
}
