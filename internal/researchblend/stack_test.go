package researchblend

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

type scoreRow struct {
	p [3]float64
	y float64
}

func directStackObjective(w [3]float64, rows []scoreRow) float64 {
	anchor := [3]float64{.98, .01, .01}
	v := 0.
	for k := range w {
		v += (w[k] - anchor[k]) * (w[k] - anchor[k])
	}
	for _, row := range rows {
		q := 0.
		for k := range w {
			q += w[k] * row.p[k]
		}
		v += (q - row.y) * (q - row.y)
	}
	return v
}
func stackStats(rows []scoreRow) ([3][3]float64, [3]float64) {
	var a [3][3]float64
	c := [3]float64{.98, .01, .01}
	for i := range a {
		a[i][i] = 1
	}
	for _, r := range rows {
		for i := range c {
			c[i] += r.p[i] * r.y
			for j := range c {
				a[i][j] += r.p[i] * r.p[j]
			}
		}
	}
	return a, c
}
func TestSimplexRidgeAgainstKKTAndGrid(t *testing.T) {
	rng := rand.New(rand.NewSource(898173))
	for trial := 0; trial < 150; trial++ {
		rows := []scoreRow{}
		for j := 0; j < trial%65; j++ {
			r := scoreRow{y: float64(rng.Intn(2))}
			for k := range r.p {
				r.p[k] = rng.Float64()
			}
			if trial%3 == 0 {
				r.p[1] = r.p[0]
			}
			if trial%5 == 0 {
				r.p = [3]float64{.9, .9, .9}
			}
			rows = append(rows, r)
		}
		a, c := stackStats(rows)
		w, err := simplexRidge(a, c)
		if err != nil {
			t.Fatal(err)
		}
		sum := 0.
		g := [3]float64{}
		active := 0.
		count := 0
		for i, v := range w {
			if v < 0 || v > 1 || math.IsNaN(v) {
				t.Fatal("invalid simplex")
			}
			sum += v
			g[i] = -2 * c[i]
			for j := range w {
				g[i] += 2 * a[i][j] * w[j]
			}
			if v > 1e-10 {
				active += g[i]
				count++
			}
		}
		if math.Abs(sum-1) > 1e-12 || count == 0 {
			t.Fatal("simplex mass differs")
		}
		active /= float64(count)
		for i, v := range w {
			if v > 1e-10 && math.Abs(g[i]-active) > 1e-8 || v <= 1e-10 && g[i] < active-1e-8 {
				t.Fatal("KKT optimality violated")
			}
		}
		value := directStackObjective(w, rows)
		for a := 0; a <= 30; a++ {
			for b := 0; b <= 30-a; b++ {
				candidate := [3]float64{float64(a) / 30, float64(b) / 30, float64(30-a-b) / 30}
				if value > directStackObjective(candidate, rows)+1e-8 {
					t.Fatal("grid point beats claimed optimum")
				}
			}
		}
	}
	for _, n := range []int{0, 1, 32, 200} {
		rows := make([]scoreRow, n)
		for i := range rows {
			rows[i] = scoreRow{[3]float64{.8, .8, .8}, float64(i % 2)}
		}
		a, c := stackStats(rows)
		w, err := simplexRidge(a, c)
		if err != nil {
			t.Fatal(err)
		}
		for i, v := range [3]float64{.98, .01, .01} {
			if math.Abs(w[i]-v) > 1e-11 {
				t.Fatal("identical experts moved ridge anchor")
			}
		}
	}
}

func TestSealedTicketsAndDelayedScoreRows(t *testing.T) {
	b, r := fixture(11)
	m, _ := NewStack(b, r)
	other, _ := NewStack(b, r)
	first, err := m.Issue(0)
	if err != nil {
		t.Fatal(err)
	}
	p0 := m.pending[0].p
	second, err := m.Issue(5)
	if err != nil {
		t.Fatal(err)
	}
	p5 := m.pending[5].p
	if _, err = m.Issue(0); err == nil {
		t.Fatal("duplicate pending nomination")
	}
	foreign, _ := other.Issue(0)
	beforeA, beforeC, beforeW := m.gram, m.response, m.weights
	for _, bad := range []Ticket{{}, foreign, {owner: m, index: 0, serial: first.serial + 1}, {owner: m, index: 11, serial: 1}} {
		if m.Resolve(bad, true) == nil || m.gram != beforeA || m.response != beforeC || m.weights != beforeW || m.pendingCount != 2 {
			t.Fatal("foreign ticket mutated state")
		}
	}
	if err = m.Resolve(second, false); err != nil {
		t.Fatal(err)
	}
	current, err := m.child.conditional(0)
	if err != nil || current == p0 {
		t.Fatal("delayed-control child forecasts did not move")
	}
	// A ticket's forecast is not accepted as a replacement for the sealed row.
	first.forecast = math.NaN()
	if err = m.Resolve(first, true); err != nil {
		t.Fatal(err)
	}
	a, c := stackStats([]scoreRow{{p0, 1}, {p5, 0}})
	for i := range a {
		for j := range a[i] {
			if math.Abs(a[i][j]-m.gram[i][j]) > 1e-14 {
				t.Fatal("post-outcome row replaced original")
			}
		}
		if math.Abs(c[i]-m.response[i]) > 1e-14 {
			t.Fatal("response changed")
		}
	}
	beforeA, beforeC, beforeW = m.gram, m.response, m.weights
	if m.Resolve(first, false) == nil || m.gram != beforeA || m.response != beforeC || m.weights != beforeW || m.pendingCount != 0 {
		t.Fatal("replayed ticket accepted")
	}
	if err = m.Observe(0, false); err == nil {
		t.Fatal("duplicate member evidence")
	}
}

func TestStackQuarantineAndOwnership(t *testing.T) {
	b, r := fixture(11)
	m, _ := NewStack(b, r)
	q, err := m.Predict(0)
	if err != nil {
		t.Fatal(err)
	}
	b[0], r[0] = .3, 1
	after, _ := m.Predict(0)
	if q != after {
		t.Fatal("caller mutation leaked")
	}
	ticket, _ := m.Issue(0)
	a, c, w := m.gram, m.response, m.weights
	if err = m.child.partition.Observe(0, true); err != nil {
		t.Fatal(err)
	}
	if m.Resolve(ticket, false) == nil || !m.broken || a != m.gram || c != m.response || w != m.weights {
		t.Fatal("partial child failure not isolated")
	}
	if _, err = m.Predict(1); err == nil {
		t.Fatal("quarantined law exposed")
	}
	if _, err = m.Select("random", rand.New(rand.NewSource(1))); err == nil {
		t.Fatal("quarantined nomination")
	}
}

func TestStackPendingExclusionAndSelectors(t *testing.T) {
	for _, n := range []int{2, 11, 150, 200} {
		for _, policy := range []string{"random", "stratified_random", "uncertainty", "disagreement"} {
			b, r := fixture(n)
			m, _ := NewStack(b, r)
			rng := rand.New(rand.NewSource(91103))
			tickets := []Ticket{}
			for k := 0; k < n; k++ {
				s, err := m.Select(policy, rng)
				if err != nil || !m.eligible(s.Index) || s.Probability <= 0 || s.Probability > 1 {
					t.Fatal("invalid pending-aware selection")
				}
				if policy == "random" && s.Probability != 1/float64(n-k) {
					t.Fatal("pending mass omitted")
				}
				if policy == "disagreement" {
					floor := .2 / float64(n-k)
					if math.Abs(s.Probability-floor) > 1e-15 && math.Abs(s.Probability-floor-.8) > 1e-15 {
						t.Fatal("wrong exploration probability")
					}
				}
				ticket, err := m.Issue(s.Index)
				if err != nil {
					t.Fatal(err)
				}
				tickets = append(tickets, ticket)
			}
			if _, err := m.Select(policy, rng); err == nil {
				t.Fatal("all-pending frontier selected")
			}
			for i := len(tickets) - 1; i >= 0; i-- {
				if err := m.Resolve(tickets[i], i%2 == 0); err != nil {
					t.Fatal(err)
				}
			}
			if m.pendingCount != 0 || m.child.count != n {
				t.Fatal("pending accounting differs")
			}
			for i := 0; i < n; i++ {
				q, err := m.Predict(i)
				if err != nil || q < 0 || q > 1 {
					t.Fatal("invalid final law")
				}
			}
		}
	}
}

func TestStackPrequentialHistoryReconstruction(t *testing.T) {
	b, r := fixture(150)
	m, _ := NewStack(b, r)
	rng := rand.New(rand.NewSource(937))
	rows := []scoreRow{}
	for j := 0; j < 64; j++ {
		s, err := m.Select("stratified_random", rng)
		if err != nil {
			t.Fatal(err)
		}
		p, _ := m.child.conditional(s.Index)
		ticket, err := m.Issue(s.Index)
		if err != nil {
			t.Fatal(err)
		}
		before, _ := m.Predict(s.Index)
		if math.Abs(ticket.Forecast()-before) > 1e-15 {
			t.Fatal("issued law not actual stack")
		}
		y := j%3 == 0
		row := scoreRow{p: p}
		if y {
			row.y = 1
		}
		rows = append(rows, row)
		if err = m.Resolve(ticket, y); err != nil {
			t.Fatal(err)
		}
		a, c := stackStats(rows)
		want, e := simplexRidge(a, c)
		if e != nil {
			t.Fatal(e)
		}
		for k, v := range want {
			if math.Abs(m.weights[k]-v) > 1e-12 {
				t.Fatal("incremental stats differ from full original history")
			}
		}
	}
	q := make([]float64, 150)
	for i := range q {
		q[i], _ = m.Predict(i)
	}
	a, c, w := m.gram, m.response, m.weights
	if _, err := m.Issue(-1); err == nil {
		t.Fatal("bad member issued")
	}
	if m.Observe(150, false) == nil || a != m.gram || c != m.response || w != m.weights {
		t.Fatal("invalid input mutated stats")
	}
	after := make([]float64, 150)
	for i := range after {
		after[i], _ = m.Predict(i)
	}
	if !reflect.DeepEqual(q, after) {
		t.Fatal("invalid call changed law")
	}
}

func BenchmarkStack(b *testing.B) {
	base, r := fixture(150)
	b.Run("SolveWeights", func(b *testing.B) {
		rows := make([]scoreRow, 32)
		for j := range rows {
			rows[j] = scoreRow{[3]float64{.9, float64(j) / 32, .5}, float64(j % 2)}
		}
		a, c := stackStats(rows)
		b.ReportAllocs()
		b.ResetTimer()
		for k := 0; k < b.N; k++ {
			if _, err := simplexRidge(a, c); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Forecast150", func(b *testing.B) {
		m, _ := NewStack(base, r)
		b.ReportAllocs()
		b.ResetTimer()
		for k := 0; k < b.N; k++ {
			for i := range base {
				if _, err := m.Predict(i); err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	for _, policy := range []string{"random", "stratified_random", "uncertainty", "disagreement"} {
		b.Run(policy, func(b *testing.B) {
			m, _ := NewStack(base, r)
			rng := rand.New(rand.NewSource(731))
			for j := 0; j < 16; j++ {
				if err := m.Observe(j, j%2 == 0); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				if _, err := m.Select(policy, rng); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
