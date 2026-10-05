package researchdispersion

import (
	"math"
	"math/bits"
	"math/rand"
	"reflect"
	"testing"
)

func copyRolling(m *RollingShape) RollingShape {
	x := *m
	y := cloneShape(m.model)
	x.model = &y
	x.seen = append([]uint64(nil), m.seen...)
	x.retained = append([]uint64(nil), m.retained...)
	x.labels = append([]uint64(nil), m.labels...)
	return x
}

func rollingBatchCheck(t *testing.T, m *RollingShape) {
	t.Helper()
	// Recompute newest arrived identities without trusting retained/counts.
	ref := cloneShape(m.model)
	for i, seen := range m.seen {
		mask, count := uint64(0), 0
		for ord := 63; ord >= 0 && count < m.width; ord-- {
			bit := uint64(1) << ord
			if seen&bit != 0 {
				mask |= bit
				count++
			}
		}
		if mask != m.retained[i] {
			t.Fatal("late arrivals displaced newer retained evidence")
		}
		ref.n[i] = uint16(bits.OnesCount64(mask))
		ref.success[i] = uint16(bits.OnesCount64(mask & m.labels[i]))
		if ref.n[i] != m.model.n[i] || ref.success[i] != m.model.success[i] {
			t.Fatal("retained counts disagree with identities")
		}
	}
	w, q := shapeReference(&ref)
	for j, value := range w {
		near(t, m.model.w[j], value, 3e-10)
	}
	for i, value := range q {
		got, err := m.Predict(i)
		if err != nil {
			t.Fatal(err)
		}
		near(t, got, value, 3e-10)
	}
}

func TestRollingShapeIndependentBatchAndOutOfOrder(t *testing.T) {
	for _, n := range []int{2, 11, 150, 200} {
		for _, width := range []int{1, 4, 8, 16, 64} {
			for _, reverse := range []bool{false, true} {
				m, err := NewRollingShape(bases(n), width)
				if err != nil {
					t.Fatal(err)
				}
				rng := rand.New(rand.NewSource(int64(3700 + n + width)))
				order := rng.Perm(n * 64)
				if reverse {
					for j := range order {
						order[j] = len(order) - j - 1
					}
				}
				for j, trial := range order {
					i, ordinal := trial%n, trial/n+1
					if err := m.Observe(i, ordinal, trial%7 < 3); err != nil {
						t.Fatal(err)
					}
					if j == 0 || j == n*2 || j == len(order)/2 || j == len(order)-1 {
						rollingBatchCheck(t, m)
					}
				}
			}
		}
	}
}

func TestRollingShapeEvictionVersusFullAndFailureAtomicity(t *testing.T) {
	m, _ := NewRollingShape(bases(2), 4)
	full, _ := NewShape(bases(2))
	for ordinal := 1; ordinal <= 8; ordinal++ {
		useful := ordinal > 4
		if m.Observe(0, ordinal, useful) != nil || full.Observe(0, ordinal, useful) != nil {
			t.Fatal("valid evidence failed")
		}
		rollingBatchCheck(t, m)
	}
	if m.model.n[0] != 4 || m.model.success[0] != 4 || full.n[0] != 8 || full.success[0] != 4 {
		t.Fatal("old-label removal did not change the evidence set")
	}
	q, _ := m.Predict(0)
	p, _ := full.Predict(0)
	if q-p <= .1 {
		t.Fatal("eviction control was vacuous")
	}
	before := copyRolling(m)
	for _, v := range [][2]int{{-1, 1}, {2, 1}, {0, 0}, {0, 65}, {0, 8}} {
		if m.Observe(v[0], v[1], true) == nil || !reflect.DeepEqual(before, copyRolling(m)) {
			t.Fatal("invalid/replayed label mutated window")
		}
	}
	m.model.logs[0] = math.Inf(1)
	before = copyRolling(m)
	if m.Observe(0, 9, false) == nil || !reflect.DeepEqual(before, copyRolling(m)) {
		t.Fatal("failed normalization partially evicted evidence")
	}
	for _, width := range []int{0, 65} {
		if _, err := NewRollingShape(bases(2), width); err == nil {
			t.Fatal("invalid width")
		}
	}
}

func TestWindowObserverPrivateOriginalAndExpertArithmetic(t *testing.T) {
	m, _ := NewWindowObserver(bases(11), 1, 704, "adaptive")
	tickets := make([]DelayedTicket, 704)
	for j := range tickets {
		var err error
		tickets[j], err = m.Issue(j%11, int64(j))
		if err != nil {
			t.Fatal(err)
		}
	}
	rng := rand.New(rand.NewSource(3737))
	for step, trial := range rng.Perm(len(tickets)) {
		y := trial%5 < 3
		old := m.weights
		original := m.original[tickets[trial].slot]
		var expected [4]float64
		sum := 0.
		for j, q := range original {
			if !y {
				q = 1 - q
			}
			expected[j] = old[j] * q
			sum += expected[j]
		}
		for j := range expected {
			expected[j] = (599.0/600)*expected[j]/sum + windowPrior[j]/600
		}
		q := tickets[trial].Forecast()
		tickets[trial].q = .001
		r, err := m.Resolve(tickets[trial], y, int64(704+step))
		if err != nil || r.Forecast != q || r.Member != trial%11 || r.TrialOrdinal != trial/11+1 {
			t.Fatal("original forecast was replaced", err)
		}
		for j, w := range m.weights {
			near(t, w, expected[j], 1e-14)
		}
		if step == 0 || step == 31 || step == 703 {
			for _, child := range m.children {
				rollingBatchCheck(t, child)
			}
		}
	}
	if m.Pending() != 0 {
		t.Fatal("undrained observer")
	}
}

func TestWindowObserverAtomicOwnerEpochCancel(t *testing.T) {
	m, _ := NewWindowObserver(bases(2), 1, 2, "adaptive")
	a, _ := m.Issue(0, 1)
	b, _ := m.Issue(1, 1)
	weights := m.weights
	first := copyRolling(m.children[0])
	m.children[1].model.logs[0] = math.Inf(1)
	if _, err := m.Resolve(a, true, 2); err == nil || m.Pending() != 2 || m.ledger.clock != 1 || m.weights != weights || !reflect.DeepEqual(first, copyRolling(m.children[0])) {
		t.Fatal("later child failure committed an earlier child or receipt")
	}
	if m.Cancel(b, 2) != nil {
		t.Fatal("cancellation failed")
	}
	if _, err := m.Resolve(b, true, 2); err == nil {
		t.Fatal("cancelled trial learned")
	}
	if err := m.BeginEpoch(2, 3); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Resolve(a, true, 3); err == nil || m.Pending() != 0 || m.weights != windowPrior {
		t.Fatal("old epoch/owner affected reset state")
	}
	for _, child := range m.children {
		rollingBatchCheck(t, child)
	}
	if _, err := NewWindowObserver(bases(2), 1, 1, "unknown"); err == nil {
		t.Fatal("invalid mode")
	}
}

func TestWindowObserverChangingExpertRows(t *testing.T) {
	m, _ := NewWindowObserver(bases(11), 1, 11, "adaptive")
	changed := false
	for ordinal := 1; ordinal <= 16; ordinal++ {
		for i := 0; i < 11; i++ {
			at := int64((ordinal-1)*11 + i)
			ticket, err := m.Issue(i, at)
			if err != nil {
				t.Fatal(err)
			}
			old, row := m.weights, m.original[ticket.slot]
			y := ordinal > 8
			var expected [4]float64
			sum := 0.
			for j, q := range row {
				if !y {
					q = 1 - q
				}
				expected[j] = old[j] * q
				sum += expected[j]
			}
			for j := range expected {
				expected[j] = (599.0/600)*expected[j]/sum + windowPrior[j]/600
			}
			if _, err := m.Resolve(ticket, y, at); err != nil {
				t.Fatal(err)
			}
			for j, w := range m.weights {
				near(t, w, expected[j], 1e-14)
				changed = changed || math.Abs(w-windowPrior[j]) > .001
			}
		}
	}
	if !changed {
		t.Fatal("expert adaptation test remained at its prior")
	}
}

func BenchmarkRollingShapeReplace(b *testing.B) {
	m, _ := NewRollingShape(bases(150), 4)
	for ord := 1; ord <= 4; ord++ {
		if err := m.Observe(0, ord, ord%2 == 0); err != nil {
			b.Fatal(err)
		}
	}
	original := copyRolling(m)
	b.ReportAllocs()
	for b.Loop() {
		m.model.logs, m.model.w = original.model.logs, original.model.w
		m.model.n[0], m.model.success[0] = 4, 2
		m.seen[0], m.retained[0], m.labels[0] = original.seen[0], original.retained[0], original.labels[0]
		copy(m.model.cache[:Hypotheses*ShapeKernels], original.model.cache[:Hypotheses*ShapeKernels])
		if err := m.Observe(0, 5, true); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWindowObserverNew(b *testing.B) {
	base := bases(150)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := NewWindowObserver(base, 1, 2400, "adaptive"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWindowObserverIssue(b *testing.B) {
	m, _ := NewWindowObserver(bases(150), 1, 2400, "adaptive")
	b.ReportAllocs()
	for b.Loop() {
		m.ledger.issued[0], m.ledger.pending = 0, 0
		if _, err := m.Issue(0, 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWindowObserverResolve(b *testing.B) {
	benchmarkWindowResolve(b, false)
}

func BenchmarkWindowObserverWarmResolve(b *testing.B) {
	benchmarkWindowResolve(b, true)
}

func benchmarkWindowResolve(b *testing.B, warm bool) {
	m, _ := NewWindowObserver(bases(150), 1, 2400, "adaptive")
	if warm {
		for ordinal := 1; ordinal <= 16; ordinal++ {
			for i := 0; i < 150; i++ {
				at := int64((ordinal-1)*150 + i)
				ticket, err := m.Issue(i, at)
				if err != nil {
					b.Fatal(err)
				}
				if _, err := m.Resolve(ticket, (ordinal+i)%3 != 0, at); err != nil {
					b.Fatal(err)
				}
			}
		}
	}
	at := m.ledger.clock + 1
	a, _ := m.Issue(0, at)
	weights := m.weights
	originals := make([]RollingShape, 4)
	for j, child := range m.children {
		originals[j] = copyRolling(child)
	}
	b.ReportAllocs()
	for b.Loop() {
		for j, child := range m.children {
			o := &originals[j]
			child.model.logs, child.model.w = o.model.logs, o.model.w
			child.model.n[0], child.model.success[0] = o.model.n[0], o.model.success[0]
			child.seen[0], child.retained[0], child.labels[0] = o.seen[0], o.retained[0], o.labels[0]
			copy(child.model.cache[:Hypotheses*ShapeKernels], o.model.cache[:Hypotheses*ShapeKernels])
		}
		m.weights = weights
		m.ledger.trials[a.slot].status, m.ledger.pending, m.ledger.clock = 1, 1, at
		if _, err := m.Resolve(a, true, at+1); err != nil {
			b.Fatal(err)
		}
	}
}
