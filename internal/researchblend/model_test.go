package researchblend

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

type treeRef struct {
	nodes []int
	prior float64
}

// Independent flat-tree enumeration, deliberately not using child internals.
func treeReferences(lo, hi, depth, node int) []treeRef {
	p := treeRef{nodes: make([]int, 8), prior: .5}
	for j := lo; j < hi; j++ {
		p.nodes[j] = node
	}
	if depth == 3 {
		p.prior = 1
		return []treeRef{p}
	}
	out := []treeRef{p}
	mid := (lo + hi) / 2
	for _, l := range treeReferences(lo, mid, depth+1, 2*node+1) {
		for _, r := range treeReferences(mid, hi, depth+1, 2*node+2) {
			q := treeRef{nodes: make([]int, 8), prior: .5 * l.prior * r.prior}
			copy(q.nodes[lo:mid], l.nodes[lo:mid])
			copy(q.nodes[mid:hi], r.nodes[mid:hi])
			out = append(out, q)
		}
	}
	return out
}

func reference(base, coordinate []float64, evidence map[int]bool) ([]float64, [3]float64) {
	n := len(base)
	forecast := make([]float64, n)
	var family [3]float64
	add := func(k int, prior, likelihood float64, p []float64) {
		mass := prior * likelihood
		family[k] += mass
		for i, q := range p {
			if y, ok := evidence[i]; ok {
				v := 0.0
				if y {
					v = 1
				}
				q = (2*q + v) / 3
			}
			forecast[i] += mass * q
		}
	}
	likelihood := func(p []float64) float64 {
		v := 1.0
		for i, y := range evidence {
			if y {
				v *= p[i]
			} else {
				v *= 1 - p[i]
			}
		}
		return v
	}
	add(0, .98, likelihood(base), base)
	add(1, .01*.1, likelihood(base), base)
	converted := make([]float64, n)
	for i, b := range base {
		converted[i] = (.9*b - .05) / .8
	}
	add(1, .01*.8, likelihood(converted), converted)
	for _, a := range []float64{.1, .3, .5, .7, .9} {
		for _, c := range []float64{-.8, -.4, 0, .4, .8} {
			p := make([]float64, n)
			for i := range p {
				p[i] = math.Max(.02, math.Min(.98, a+c*float64(i)/float64(n-1)))
			}
			add(1, .01*.004, likelihood(p), p)
		}
	}
	add(2, .01*.99, likelihood(base), base)
	for _, tree := range treeReferences(0, 8, 0, 0) {
		var s, f [15]int
		for i, y := range evidence {
			bin := min(7, int(8*coordinate[i]))
			node := tree.nodes[bin]
			if y {
				s[node]++
			} else {
				f[node]++
			}
		}
		integral := 1.0
		for node := 0; node < 15; node++ {
			// Beta(1+s,1+f) / Beta(1,1), using independent products.
			for j := 1; j <= s[node]; j++ {
				integral *= float64(j) / float64(j+f[node])
			}
			integral /= float64(s[node] + f[node] + 1)
		}
		p := make([]float64, n)
		for i := range p {
			node := tree.nodes[min(7, int(8*coordinate[i]))]
			p[i] = float64(1+s[node]) / float64(2+s[node]+f[node])
		}
		add(2, .01*.01*tree.prior, integral, p)
	}
	mass := family[0] + family[1] + family[2]
	for i := range forecast {
		forecast[i] /= mass
	}
	for k := range family {
		family[k] /= mass
	}
	return forecast, family
}

func fixture(n int) ([]float64, []float64) {
	b, r := make([]float64, n), make([]float64, n)
	for i := range b {
		r[i] = float64(i) / float64(n-1)
		b[i] = .925 - .675*r[i]
	}
	return b, r
}

func TestJointModelEnumeration(t *testing.T) {
	for _, n := range []int{2, 11, 150} {
		b, r := fixture(n)
		m, err := New(b, r)
		if err != nil {
			t.Fatal(err)
		}
		evidence := map[int]bool{}
		order := rand.New(rand.NewSource(981213)).Perm(n)
		for count := 0; count <= min(n, 32); count++ {
			q, w := reference(b, r, evidence)
			for i, want := range q {
				got, err := m.Predict(i)
				if err != nil || math.Abs(got-want) > 2e-13 {
					t.Fatalf("joint forecast n=%d count=%d i=%d: %.17g != %.17g", n, count, i, got, want)
				}
			}
			for k, want := range w {
				if math.Abs(m.weights[k]-want) > 2e-13 {
					t.Fatal("flat integrated family mass differs")
				}
			}
			if count < min(n, 32) {
				i, y := order[count], count%3 != 0
				if err := m.Observe(i, y); err != nil {
					t.Fatal(err)
				}
				evidence[i] = y
			}
		}
	}
}

func predictions(t *testing.T, m *Model) []float64 {
	t.Helper()
	q := make([]float64, len(m.base))
	for i := range q {
		var err error
		q[i], err = m.Predict(i)
		if err != nil {
			t.Fatal(err)
		}
	}
	return q
}

func TestOwnershipAndInvalidEvidence(t *testing.T) {
	b, r := fixture(11)
	m, _ := New(b, r)
	initial := predictions(t, m)
	for i, q := range initial {
		if math.Abs(q-(.9999*b[i]+.00005)) > 2e-15 {
			t.Fatal("cold law changed")
		}
	}
	b[0], r[0] = .3, 1
	if !reflect.DeepEqual(initial, predictions(t, m)) {
		t.Fatal("caller mutated model")
	}
	if err := m.Observe(3, true); err != nil {
		t.Fatal(err)
	}
	q, w := predictions(t, m), m.weights
	for _, i := range []int{-1, 11, 3} {
		if m.Observe(i, false) == nil || !reflect.DeepEqual(q, predictions(t, m)) || m.weights != w || m.count != 1 {
			t.Fatal("invalid evidence mutated state")
		}
	}
	for _, p := range []string{"unknown", "random", "family_information", "stratified_random"} {
		if _, err := m.Select(p, nil); err == nil {
			t.Fatal("invalid policy accepted")
		}
	}
	for _, bad := range []float64{math.NaN(), math.Inf(1), .2, 1} {
		b[1] = bad
		if _, err := New(b, r); err == nil {
			t.Fatal("unsupported baseline accepted")
		}
	}
	for _, bad := range []float64{math.NaN(), math.Inf(1), -1, 2} {
		b, r = fixture(11)
		r[1] = bad
		if _, err := New(b, r); err == nil {
			t.Fatal("unsupported coordinate accepted")
		}
	}
}

func TestPartialFailureQuarantines(t *testing.T) {
	b, r := fixture(11)
	m, _ := New(b, r)
	// Fault injection: make child evidence diverge from its parent's guard.
	if err := m.partition.Observe(0, true); err != nil {
		t.Fatal(err)
	}
	if m.Observe(0, false) == nil || !m.broken {
		t.Fatal("partial failure not quarantined")
	}
	if _, err := m.Predict(1); err == nil {
		t.Fatal("partial law exposed")
	}
	if _, err := m.Select("random", rand.New(rand.NewSource(1))); err == nil {
		t.Fatal("quarantined model selected")
	}
}

func TestSelectionCoverageAndProbabilities(t *testing.T) {
	for _, n := range []int{2, 11, 150, 200} {
		for _, policy := range []string{"random", "stratified_random", "uncertainty", "family_information"} {
			b, r := fixture(n)
			m, _ := New(b, r)
			rng := rand.New(rand.NewSource(91379))
			buckets := map[int]bool{}
			for count := 0; count < n; count++ {
				s, err := m.Select(policy, rng)
				if err != nil || s.Index < 0 || s.Index >= n || m.seen[s.Index] || s.Probability <= 0 || s.Probability > 1 {
					t.Fatal("invalid selection")
				}
				if policy == "random" && s.Probability != 1/float64(n-count) {
					t.Fatal("wrong uniform probability")
				}
				if policy == "family_information" {
					floor := .2 / float64(n-count)
					if math.Abs(s.Probability-floor) > 1e-15 && math.Abs(s.Probability-floor-.8) > 1e-15 {
						t.Fatal("wrong mixed probability")
					}
				}
				if policy == "stratified_random" && count < min(32, n) {
					bucket := m.order[count]
					lo, hi := bucket*n/len(m.order), (bucket+1)*n/len(m.order)
					if s.Index < lo || s.Index >= hi || buckets[bucket] || s.Probability != 1/float64(hi-lo) {
						t.Fatal("stratum or conditional probability differs")
					}
					buckets[bucket] = true
				}
				if err := m.Observe(s.Index, count%2 == 0); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := m.Select(policy, rng); err == nil {
				t.Fatal("exhausted model selected")
			}
		}
	}
}

func TestStratifiedMarginalCoverage(t *testing.T) {
	b, r := fixture(150)
	counts := make([]int, 150)
	for seed := 0; seed < 1000; seed++ {
		m, _ := New(b, r)
		rng := rand.New(rand.NewSource(int64(seed)))
		for k := 0; k < 32; k++ {
			s, err := m.Select("stratified_random", rng)
			if err != nil {
				t.Fatal(err)
			}
			counts[s.Index]++
			if err := m.Observe(s.Index, true); err != nil {
				t.Fatal(err)
			}
		}
	}
	for bucket := 0; bucket < 32; bucket++ {
		lo, hi := bucket*150/32, (bucket+1)*150/32
		for i := lo; i < hi; i++ {
			if math.Abs(float64(counts[i])/1000-1/float64(hi-lo)) > .065 {
				t.Fatal("member missing or nonuniform within stratum")
			}
		}
	}
}

func TestFamilyInformationLookahead(t *testing.T) {
	b, r := fixture(11)
	evidence := map[int]bool{2: false, 7: true, 9: false}
	m, _ := New(b, r)
	for _, i := range []int{2, 7, 9} {
		if err := m.Observe(i, evidence[i]); err != nil {
			t.Fatal(err)
		}
	}
	entropy := func(w [3]float64) float64 {
		v := 0.0
		for _, p := range w {
			if p > 0 {
				v -= p * math.Log(p)
			}
		}
		return v
	}
	q, w := reference(b, r, evidence)
	maximum := math.Inf(-1)
	for i := range b {
		if _, seen := evidence[i]; seen {
			continue
		}
		evidence[i] = true
		_, positive := reference(b, r, evidence)
		evidence[i] = false
		_, negative := reference(b, r, evidence)
		delete(evidence, i)
		information := entropy(w) - q[i]*entropy(positive) - (1-q[i])*entropy(negative)
		maximum = math.Max(maximum, information)
	}
	for seed := 0; seed < 100; seed++ {
		s, err := m.Select("family_information", rand.New(rand.NewSource(int64(seed))))
		if err != nil || math.Abs(s.Score-maximum) > 2e-13 {
			t.Fatal("family score does not equal exact entropy lookahead")
		}
		if s.Probability > .8 {
			i := s.Index
			evidence[i] = true
			_, positive := reference(b, r, evidence)
			evidence[i] = false
			_, negative := reference(b, r, evidence)
			delete(evidence, i)
			value := entropy(w) - q[i]*entropy(positive) - (1-q[i])*entropy(negative)
			if math.Abs(value-maximum) > 2e-13 {
				t.Fatal("exploit probability attached to nonmaximizer")
			}
		}
	}
}

func BenchmarkBlend(b *testing.B) {
	base, r := fixture(150)
	b.Run("Construct", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := New(base, r); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Forecast150", func(b *testing.B) {
		m, _ := New(base, r)
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
	for _, policy := range []string{"random", "stratified_random", "uncertainty", "family_information"} {
		b.Run(policy, func(b *testing.B) {
			m, _ := New(base, r)
			rng := rand.New(rand.NewSource(135))
			for i := 0; i < 16; i++ {
				if err := m.Observe(i, i%2 == 0); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := m.Select(policy, rng); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
