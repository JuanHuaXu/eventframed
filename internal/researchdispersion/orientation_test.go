package researchdispersion

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

// Enumerate every hidden path, independently of the candidate recursion.
func orientationPaths(m *OrientationFilter, end int) float64 {
	total, flipped := 0., 0.
	for path := 0; path < 1<<end; path++ {
		mass, old := 1., 0
		for j, x := range m.nodes[:end] {
			state := (path >> j) & 1
			if state == old {
				mass *= 1 - m.hazard
			} else {
				mass *= m.hazard
			}
			if x.observed {
				p := m.template[x.member]
				if state == 1 {
					p = 1 - p
				}
				if x.useful {
					mass *= p
				} else {
					mass *= 1 - p
				}
			}
			old = state
		}
		total += mass
		if path>>(end-1)&1 == 1 {
			flipped += mass
		}
	}
	return flipped / total
}

func TestOrientationFilterPathsAndOrder(t *testing.T) {
	for _, hazard := range []float64{0, orientationHazard, .1, .49} {
		for _, order := range [][]int{{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, {9, 0, 8, 1, 7, 2, 6, 3, 5, 4}} {
			m, err := NewOrientationFilter([]float64{.15, .75, .6}, hazard, 10)
			if err != nil {
				t.Fatal(err)
			}
			for j := 0; j < 10; j++ {
				if _, err := m.nominate(j % 3); err != nil {
					t.Fatal(err)
				}
			}
			for _, j := range order {
				if err := m.Observe(j, j%4 == 0); err != nil {
					t.Fatal(err)
				}
				for end := 1; end <= 10; end++ {
					near(t, m.nodes[end-1].forward, orientationPaths(m, end), 4e-14)
				}
				f := orientationPaths(m, 10)
				next := hazard + (1-2*hazard)*f
				for i, p := range m.template {
					q, err := m.Predict(i)
					if err != nil {
						t.Fatal(err)
					}
					near(t, q, (1-next)*p+next*(1-p), 5e-14)
				}
			}
		}
	}
}

func TestOrientationFilterAtomicityAndCaps(t *testing.T) {
	for _, p := range [][]float64{nil, {0}, {1}, {math.NaN()}, {math.Inf(1)}} {
		if _, err := NewOrientationFilter(p, .1, 2); err == nil {
			t.Fatal("invalid template")
		}
	}
	m, _ := NewOrientationFilter([]float64{.25, .8}, .1, 3)
	for i := 0; i < 3; i++ {
		m.nominate(i % 2)
	}
	if _, err := m.nominate(0); err == nil {
		t.Fatal("cap")
	}
	if _, err := m.Predict(-1); err == nil {
		t.Fatal("member")
	}
	if err := m.Observe(2, true); err != nil {
		t.Fatal(err)
	}
	before := append([]orientationNode(nil), m.nodes...)
	m.template[0] = math.NaN()
	if err := m.Observe(0, false); err == nil || !reflect.DeepEqual(before, m.nodes) {
		t.Fatal("failed replay partially published")
	}
	m.template[0] = .25
	if err := m.Observe(2, false); err == nil || !reflect.DeepEqual(before, m.nodes) {
		t.Fatal("duplicate changed state")
	}
}

func TestOrientationObserverAnchorAndLateEvidence(t *testing.T) {
	for _, mode := range []string{"anchor", "orientation"} {
		m, _ := NewOrientationObserver(bases(11), 7, 704, mode)
		tickets := make([]DelayedTicket, 88)
		for trial := range tickets {
			var err error
			tickets[trial], err = m.Issue(trial%11, int64(trial))
			if err != nil {
				t.Fatal(err)
			}
		}
		order := rand.New(rand.NewSource(999)).Perm(88)
		anchorCount := 0
		for j, trial := range order {
			ticket := tickets[trial]
			original := ticket.Forecast()
			ticket.q = .999
			r, err := m.Resolve(ticket, trial%4 == 0, int64(88+j))
			if err != nil || r.Forecast != original || r.TrialOrdinal != trial/11+1 {
				t.Fatal("original receipt", r, err)
			}
			if trial/11 < orientationWarm {
				anchorCount++
			}
			if (m.filter != nil) != (anchorCount == 44) {
				t.Fatal("premature or absent anchor")
			}
			if m.filter != nil {
				_, template := shapeReference(m.anchor.model)
				for i, p := range template {
					near(t, p, m.filter.template[i], 3e-10)
				}
				// Separate forward matrix multiplication over arrived emissions.
				f := orientationReference(m.filter.template, m.filter.hazard, m.filter.nodes)
				near(t, m.filter.nodes[len(m.filter.nodes)-1].forward, f, 2e-13)
			}
		}
		if m.Pending() != 0 || len(m.filter.nodes) != 44 {
			t.Fatal("coverage")
		}
	}
}

// Independent two-entry matrix propagation avoids the candidate scalar map.
func orientationReference(template []float64, hazard float64, nodes []orientationNode) float64 {
	a, b := 1., 0.
	for _, node := range nodes {
		x, y := a*(1-hazard)+b*hazard, a*hazard+b*(1-hazard)
		if node.observed {
			p := template[node.member]
			if node.useful {
				x, y = x*p, y*(1-p)
			} else {
				x, y = x*(1-p), y*p
			}
		}
		a, b = x/(x+y), y/(x+y)
	}
	return b
}

func TestOrientationObserverOwnerCancelEpochFreezeFailure(t *testing.T) {
	m, _ := NewOrientationObserver(bases(2), 1, 128, "orientation")
	other, _ := NewOrientationObserver(bases(2), 1, 128, "orientation")
	foreign, _ := other.Issue(0, 0)
	if _, err := m.Resolve(foreign, true, 0); err == nil {
		t.Fatal("foreign owner")
	}
	var last DelayedTicket
	for j := 0; j < 8; j++ {
		last, _ = m.Issue(j%2, int64(j))
		if j < 7 {
			if _, err := m.Resolve(last, j%2 == 0, int64(j)); err != nil {
				t.Fatal(err)
			}
		}
	}
	// Inject a bad untouched row. Freeze must be validated before the last
	// anchor label, ledger clock, pending count or template is committed.
	row := m.anchor.model.cache[:Hypotheses*ShapeKernels]
	backup := append([]float64(nil), row...)
	row[0] = math.NaN()
	before := cloneShape(m.anchor.model)
	if _, err := m.Resolve(last, true, 8); err == nil || m.anchorLabels != 7 || m.Pending() != 1 || m.ledger.clock != 7 || !reflect.DeepEqual(before.n, m.anchor.model.n) {
		t.Fatal("non-atomic freeze")
	}
	copy(row, backup)
	if _, err := m.Resolve(last, true, 8); err != nil || m.filter == nil {
		t.Fatal("freeze recovery", err)
	}
	post, _ := m.Issue(0, 9)
	if err := m.Cancel(post, 10); err != nil || m.filter.nodes[0].observed {
		t.Fatal("cancellation supplied evidence", err)
	}
	if err := m.BeginEpoch(2, 11); err != nil || m.filter != nil || m.anchorLabels != 0 {
		t.Fatal("epoch", err)
	}
	if _, err := m.Resolve(post, true, 11); err == nil {
		t.Fatal("old epoch")
	}
	if _, err := m.Issue(0, 10); err == nil {
		t.Fatal("backward issue")
	}
	// A cancelled anchor prevents freezing instead of fabricating its label.
	c, _ := NewOrientationObserver(bases(2), 1, 128, "orientation")
	for j := 0; j < 8; j++ {
		ticket, _ := c.Issue(j%2, int64(j))
		if j == 0 {
			c.Cancel(ticket, int64(j))
		} else {
			c.Resolve(ticket, true, int64(j))
		}
	}
	if c.filter != nil || c.anchorLabels != 7 {
		t.Fatal("censoring falsely completed anchor")
	}
}

func BenchmarkOrientationFilterReplay150(b *testing.B) {
	m, _ := NewOrientationFilter(bases(150), orientationHazard, 2400)
	for j := 0; j < 1800; j++ {
		m.nominate(j % 150)
		if j != 1650 {
			m.Observe(j, j%3 == 0)
		}
	}
	copyNodes := append([]orientationNode(nil), m.nodes...)
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		copy(m.nodes[1650:], copyNodes[1650:])
		if err := m.Observe(1650, j%2 == 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOrientationFilterReplay1800(b *testing.B) {
	m, _ := NewOrientationFilter(bases(150), orientationHazard, 2400)
	for j := 0; j < 1800; j++ {
		m.nominate(j % 150)
		if j != 0 {
			m.Observe(j, j%3 == 0)
		}
	}
	copyNodes := append([]orientationNode(nil), m.nodes...)
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		copy(m.nodes, copyNodes)
		if err := m.Observe(0, j%2 == 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOrientationObserverNew(b *testing.B) {
	base := bases(150)
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		if _, err := NewOrientationObserver(base, 1, 2400, "orientation"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOrientationObserverFreezePlan(b *testing.B) {
	m, _ := NewOrientationObserver(bases(150), 1, 2400, "orientation")
	for j := 0; j < 600; j++ {
		ticket, _ := m.Issue(j%150, int64(j))
		if j < 599 {
			m.Resolve(ticket, j%3 == 0, int64(j))
		}
	}
	plan, err := m.anchor.prepare(149, 4, false)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		if _, _, err := m.prepareFreeze(&plan); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOrientationFilterPredict150(b *testing.B) {
	m, _ := NewOrientationFilter(bases(150), orientationHazard, 2400)
	b.ReportAllocs()
	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		for i := 0; i < 150; i++ {
			if _, err := m.Predict(i); err != nil {
				b.Fatal(err)
			}
		}
	}
}
