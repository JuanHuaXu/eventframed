package observationlearners

import (
	"math"
	"math/bits"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type literalJointRow struct {
	x       uint16
	mass, p float64
}

// Independent reference: filter literal full assignments at every node. No
// partial-cell lookup, ternary memoization or planner recursion is reused.
func literalLookahead(rows []literalJointRow, mask uint16, budget int) (float64, uint8) {
	mass, weighted := 0., 0.
	for _, r := range rows {
		mass += r.mass
		weighted += r.mass * r.p
	}
	p := weighted / mass
	h := -p*math.Log(p) - (1-p)*math.Log(1-p)
	used := bits.OnesCount16(mask)
	if used >= budget || (used > 0 && (p <= .1 || p >= .9)) {
		return h, 1
	}
	best, action := math.Inf(1), uint8(1)
	for s := 0; s < 3; s++ {
		for d := 0; d < 3; d++ {
			view := uint16((1<<(d+1))-1) << (3 * s)
			added := view &^ mask
			if added == 0 || bits.OnesCount16(mask|view) > budget || (mask == 0 && (s != 0 || d != 0)) {
				continue
			}
			groups := map[uint16][]literalJointRow{}
			for _, r := range rows {
				groups[r.x&added] = append(groups[r.x&added], r)
			}
			expected := 0.
			// Numeric order avoids nondeterministic floating-point map iteration.
			for value := uint16(0); value < 512; value++ {
				group := groups[value]
				if len(group) == 0 {
					continue
				}
				gm := 0.
				for _, r := range group {
					gm += r.mass
				}
				v, _ := literalLookahead(group, mask|view, budget)
				expected += gm / mass * v
			}
			if expected < best-1e-12 {
				best, action = expected, uint8(2+3*s+d)
			}
		}
	}
	if action == 1 {
		return h, action
	}
	return best, action
}

func literalRows(m observationMixture) []literalJointRow {
	rows := make([]literalJointRow, 512)
	for x := uint16(0); x < 512; x++ {
		prob := .5 * m.weights[0]
		for j := 0; j < 4; j++ {
			c := m.experts.models[j].cells[partialIndex(511, x)]
			prob += m.weights[j+1] * c.weighted / c.mass
		}
		rows[x] = literalJointRow{x: x, mass: m.experts.models[0].cells[partialIndex(511, x)].mass, p: prob}
	}
	return rows
}

func TestJointLookaheadLiteral(t *testing.T) {
	for _, uniform := range []bool{true, false} {
		e, err := newObservationExperts(jointFixture(uniform, false))
		if err != nil {
			t.Fatal(err)
		}
		m, err := e.snapshot([5]float64{.1, .2, .3, .15, .25})
		if err != nil {
			t.Fatal(err)
		}
		for budget := 1; budget <= 4; budget++ {
			p, err := newJointPlanner(m, budget)
			if err != nil {
				t.Fatal(err)
			}
			got, action := p.solve(0, 0)
			want, wa := literalLookahead(literalRows(m), 0, budget)
			if math.Abs(got-want) > 1e-12 || action != wa {
				t.Fatal("literal", uniform, budget, got, want, action, wa)
			}
			nodes, branches := p.nodes, p.branches
			p.solve(0, 0)
			if p.nodes != nodes || p.branches != branches {
				t.Fatal("memo not reused")
			}
			// The forced root action alone cannot validate action selection.
			// Check every visited interior state against the independent tree.
			rows := literalRows(m)
			for i, a := range p.action {
				if a == 0 {
					continue
				}
				var mask, value uint16
				v := i
				for bit := 0; bit < 9; bit++ {
					digit := v % 3
					if digit != 0 {
						mask |= 1 << bit
					}
					if digit == 2 {
						value |= 1 << bit
					}
					v /= 3
				}
				var selected []literalJointRow
				for _, row := range rows {
					if row.x&mask == value {
						selected = append(selected, row)
					}
				}
				want, action := literalLookahead(selected, mask, budget)
				if math.Abs(want-p.value[i]) > 1e-12 || action != a {
					t.Fatal("interior literal", uniform, budget, mask, value, want, p.value[i], action, a)
				}
			}
		}
	}
}

func lookaheadParity(t testing.TB) observationMixture {
	t.Helper()
	model := new(ConditionalForest)
	// Direct enumeration builds every partial marginal independently.
	for mask := uint16(0); mask < 512; mask++ {
		for value := mask; ; value = (value - 1) & mask {
			c := conditionalCell{}
			for x := uint16(0); x < 512; x++ {
				if x&mask == value {
					c.mass++
					c.weighted += .05 + .9*float64(bits.OnesCount16(x&144)%2)
				}
			}
			model.cells[partialIndex(mask, value)] = c
			if value == 0 {
				break
			}
		}
	}
	e, err := newObservationExperts([4]*ConditionalForest{model, model, model, model})
	if err != nil {
		t.Fatal(err)
	}
	m, err := e.snapshot([5]float64{0, 1})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestJointLookaheadJointGain(t *testing.T) {
	m := lookaheadParity(t)
	base, _ := m.Forecast(1, 0)
	for scope := 0; scope < 3; scope++ {
		for depth := 0; depth < 3; depth++ {
			added := observation.View{Scope: scope, Depth: depth}.Mask() &^ uint16(1)
			for sub := added; ; sub = (sub - 1) & added {
				p, _ := m.Forecast(1|added, sub)
				if math.Abs(p-base) > 1e-12 {
					t.Fatal("nonzero one-step gain")
				}
				if sub == 0 {
					break
				}
			}
		}
	}
	p, _ := newJointPlanner(m, 6)
	v, _ := p.solve(0, 0)
	if math.Abs(v-jointObservationEntropy(.05)) > 1e-12 {
		t.Fatal("joint value", v)
	}
	t.Logf("parity planner: %d distinct states, %d branch evaluations", p.nodes, p.branches)
	for x := uint16(0); x < 512; x++ {
		greedy, err := runJointObserver(m, &jointReader{x: x, epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		got, err := runJointLookahead(m, &jointReader{x: x, epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		want, _ := m.Forecast(511, x)
		if math.Abs(greedy.Probability-.5) > 1e-12 || math.Abs(got.Probability-want) > 1e-12 || got.Cost > 6 {
			t.Fatal("joint rescue", x, greedy.Probability, got.Probability, want)
		}
		last := got.Trace[len(got.Trace)-1]
		other, err := runJointLookahead(m, &jointReader{x: x ^ (511 &^ last.Observed), epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		compareJointTrace(t, got.Result, other.Result)
	}
}

func TestJointLookaheadLifecycle(t *testing.T) {
	m := lookaheadParity(t)
	s := newRoutedObservationState()
	before := *s
	for _, reader := range []*jointReader{{epoch: 1, failAt: 2}, {epoch: 1, change: true}} {
		if _, err := s.predictLookahead(0, m.experts, reader, 1); err == nil || !reflect.DeepEqual(before, *s) {
			t.Fatal("failed issue mutated")
		}
	}
	reader := &jointReader{epoch: 1, x: 17}
	reader.hook = func() {
		if err := s.observe(0, true); err == nil {
			t.Fatal("reentrant feedback")
		}
	}
	got, err := s.predictLookahead(0, m.experts, reader, 1)
	if err != nil {
		t.Fatal(err)
	}
	last := got.Trace[len(got.Trace)-1]
	for j, v := range s.bank.pending.Experts {
		want, _ := m.experts.models[j].Forecast(last.Observed, last.Values)
		if v != want {
			t.Fatal("wrong acquired mask")
		}
	}
	if err := s.observe(0, true); err != nil {
		t.Fatal(err)
	}
	before = *s
	if err := s.observe(0, true); err == nil || !reflect.DeepEqual(before, *s) {
		t.Fatal("duplicate feedback")
	}
	if _, err := newJointPlanner(m, 7); err == nil {
		t.Fatal("unbounded budget")
	}
	if _, err := newJointPlanner(observationMixture{}, 6); err == nil {
		t.Fatal("empty law")
	}
	if _, err := runJointLookahead(m, nil, 1); err == nil {
		t.Fatal("nil reader")
	}
}

func BenchmarkJointLookahead(b *testing.B) {
	m := lookaheadParity(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, err := runJointLookahead(m, &jointReader{x: uint16(i) & 511, epoch: 1}, 1)
		if err != nil || r.Cost > 6 {
			b.Fatal(err)
		}
	}
}
