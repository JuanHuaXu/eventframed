package researchblend

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func TestCrossfitRebuildWeightsAndRows(t *testing.T) {
	for _, n := range []int{2, 11, 150, 200} {
		base, coordinates := fixture(n)
		m, err := NewCrossfit(base, coordinates)
		if err != nil {
			t.Fatal(err)
		}
		order := rand.New(rand.NewSource(831)).Perm(n)
		for count := 0; count <= min(n, 32); count++ {
			rows := []scoreRow{}
			for j, heldout := range order[:count] {
				child, _ := New(base, coordinates)
				for k, i := range order[:count] {
					if i != heldout {
						if e := child.Observe(i, k%3 != 0); e != nil {
							t.Fatal(e)
						}
					}
				}
				want, _ := child.conditional(heldout)
				got, e := m.trainingRow(heldout)
				if e != nil {
					t.Fatal(e)
				}
				for k := range want {
					if math.Abs(want[k]-got[k]) > 2e-13 {
						t.Fatal("training row disagrees with full omitted-member refit")
					}
				}
				row := scoreRow{p: want}
				if j%3 != 0 {
					row.y = 1
				}
				rows = append(rows, row)
			}
			a, c := stackStats(rows)
			w, e := simplexRidge(a, c)
			if e != nil {
				t.Fatal(e)
			}
			for k := range w {
				if math.Abs(w[k]-m.weights[k]) > 5e-12 {
					t.Fatal("crossfit weights disagree with full refits")
				}
			}
			for i := range base {
				p, _ := m.child.conditional(i)
				q, e := m.Predict(i)
				if e != nil || math.Abs(q-(w[0]*p[0]+w[1]*p[1]+w[2]*p[2])) > 5e-12 {
					t.Fatal("actual law differs from full-data child combination")
				}
			}
			if count < min(n, 32) {
				if err = m.Observe(order[count], count%3 != 0); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestCrossfitOwnershipInvalidAndQuarantine(t *testing.T) {
	base, coordinate := fixture(11)
	m, _ := NewCrossfit(base, coordinate)
	q, _ := m.Predict(0)
	base[0], coordinate[0] = .3, 1
	after, _ := m.Predict(0)
	if q != after {
		t.Fatal("caller-owned input leaked")
	}
	if err := m.Observe(0, true); err != nil {
		t.Fatal(err)
	}
	weights, count := m.weights, m.child.count
	before := make([]float64, 11)
	for i := range before {
		before[i], _ = m.Predict(i)
	}
	for _, i := range []int{-1, 11, 0} {
		if m.Observe(i, false) == nil || m.broken || weights != m.weights || count != m.child.count {
			t.Fatal("invalid evidence corrupted a valid object")
		}
	}
	for _, policy := range []string{"uncertainty", "disagreement", "family_information", "unknown"} {
		if _, err := m.Select(policy, rand.New(rand.NewSource(1))); err == nil {
			t.Fatal("label-adaptive LOO nomination silently accepted")
		}
	}
	if _, err := m.Select("random", nil); err == nil {
		t.Fatal("nil RNG accepted")
	}
	for i, want := range before {
		got, _ := m.Predict(i)
		if got != want {
			t.Fatal("invalid call mutated the law")
		}
	}
	if err := m.child.partition.Observe(1, true); err != nil {
		t.Fatal(err)
	}
	if m.Observe(1, false) == nil || !m.broken || weights != m.weights {
		t.Fatal("partial child failure not quarantined")
	}
	if _, err := m.Predict(2); err == nil {
		t.Fatal("partial law exposed")
	}
	if _, err := m.Select("random", rand.New(rand.NewSource(1))); err == nil {
		t.Fatal("quarantined object nominated evidence")
	}
}

func TestCrossfitPrefixNoFutureAndNomination(t *testing.T) {
	base, coordinate := fixture(150)
	for _, policy := range []string{"random", "stratified_random"} {
		labels := make([]bool, 150)
		for i := range labels {
			labels[i] = i%3 != 0
		}
		type tape struct {
			trace []traceV30
			q     []float64
			w     [3]float64
		}
		run := func(labels []bool) tape {
			m, _ := NewCrossfit(base, coordinate)
			rng := rand.New(rand.NewSource(637))
			result := tape{}
			for k := 0; k < 32; k++ {
				s, e := m.Select(policy, rng)
				if e != nil {
					t.Fatal(e)
				}
				q, e := m.Predict(s.Index)
				if e != nil {
					t.Fatal(e)
				}
				y := labels[s.Index]
				result.trace = append(result.trace, traceV30{s.Index, y, q, s.Probability, s.Score})
				if e = m.Observe(s.Index, y); e != nil {
					t.Fatal(e)
				}
			}
			for i := range base {
				q, e := m.Predict(i)
				if e != nil {
					t.Fatal(e)
				}
				result.q = append(result.q, q)
			}
			result.w = m.Weights()
			return result
		}
		a := run(labels)
		seen := make([]bool, 150)
		for _, v := range a.trace {
			seen[v.Index] = true
		}
		for i := range labels {
			if !seen[i] {
				labels[i] = !labels[i]
			}
		}
		if !reflect.DeepEqual(a, run(labels)) {
			t.Fatal("unobserved future labels entered LOO state")
		}
		// The frozen runner adds 303 to its seed and only the blend supports
		// both policies; match its RNG, not its independently forecast law.
		control := runV30(base, coordinate, labels, "blend", policy, 637-303)
		for i, row := range a.trace {
			if row.Index != control.Trace[i].Index || row.Useful != control.Trace[i].Useful || row.Probability != control.Trace[i].Probability {
				t.Fatal("nonadaptive evidence control differs")
			}
		}
	}
}

func BenchmarkCrossfit(b *testing.B) {
	base, coordinate := fixture(150)
	b.Run("Construct", func(b *testing.B) {
		b.ReportAllocs()
		for j := 0; j < b.N; j++ {
			if _, err := NewCrossfit(base, coordinate); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Forecast150", func(b *testing.B) {
		m, _ := NewCrossfit(base, coordinate)
		for i := 0; i < 32; i++ {
			if err := m.Observe(i, i%3 != 0); err != nil {
				b.Fatal(err)
			}
		}
		b.ReportAllocs()
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			for i := range base {
				if _, err := m.Predict(i); err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	for _, count := range []int{1, 32, 64, 150} {
		b.Run(fmt.Sprintf("Refit%d", count), func(b *testing.B) {
			m, _ := NewCrossfit(base, coordinate)
			for i := 0; i < count; i++ {
				if err := m.Observe(i, i%3 != 0); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				if _, err := m.fit(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
