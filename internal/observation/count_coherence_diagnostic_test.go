package observation

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"testing"
)

// Diagnostic of the existing estimator, not a change to its contract. Any joint
// law requires each parent conditional mean to lie in the convex hull of its
// two child conditional means, regardless of the unknown input marginal.
func TestCountCoherenceDiagnostic(t *testing.T) {
	type row struct {
		Name                       string
		N, Partitions, OutsideHull int
		MaxDefect                  float64
	}
	var rows []row
	fixtures := map[string][]Sample{
		"two_positive_cells": {{Bits: 0, Outcome: true}, {Bits: 1, Outcome: true}},
	}
	var balanced []Sample
	for x := uint16(0); x < 512; x++ {
		balanced = append(balanced, Sample{Bits: x, Outcome: false}, Sample{Bits: x, Outcome: true})
	}
	fixtures["balanced_negative_control"] = balanced
	for _, n := range []int{64, 4096} {
		rng := rand.New(rand.NewSource(int64(2026092161 + n)))
		samples := make([]Sample, n)
		for i := range samples {
			x := uint16(rng.Intn(512))
			y := ((x&64 != 0) != (x&128 != 0)) != (x&256 != 0)
			if rng.Float64() < .05 {
				y = !y
			}
			samples[i] = Sample{Bits: x, Outcome: y}
		}
		name := "noisy_parity64"
		if n == 4096 {
			name = "noisy_parity4096"
		}
		fixtures[name] = samples
	}
	for _, name := range []string{"two_positive_cells", "balanced_negative_control", "noisy_parity64", "noisy_parity4096"} {
		samples := fixtures[name]
		m, err := Fit(samples)
		if err != nil {
			t.Fatal(err)
		}
		r := row{Name: name, N: len(samples)}
		for mask := uint16(0); mask < 512; mask++ {
			for values := mask; ; values = (values - 1) & mask {
				p, err := m.ForecastObserved(mask, values)
				if err != nil {
					t.Fatal(err)
				}
				for bit := uint16(1); bit < 512; bit <<= 1 {
					if mask&bit != 0 {
						continue
					}
					a, _ := m.ForecastObserved(mask|bit, values)
					b, _ := m.ForecastObserved(mask|bit, values|bit)
					defect := math.Max(0, math.Max(math.Min(a, b)-p, p-math.Max(a, b)))
					r.Partitions++
					if defect > 1e-12 {
						r.OutsideHull++
					}
					r.MaxDefect = math.Max(r.MaxDefect, defect)
				}
				if values == 0 {
					break
				}
			}
		}
		if r.Partitions != 59049 {
			t.Fatal("partition enumeration", r.Partitions)
		}
		if name == "balanced_negative_control" && r.OutsideHull != 0 {
			t.Fatal("negative control")
		}
		if name == "two_positive_cells" {
			p, _ := m.ForecastObserved(0, 0)
			a, _ := m.ForecastObserved(1, 0)
			b, _ := m.ForecastObserved(1, 1)
			if p != .75 || a != 2./3 || b != 2./3 || r.OutsideHull == 0 {
				t.Fatal("counterexample")
			}
		}
		rows = append(rows, r)
	}
	b, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
	if path := os.Getenv("EVENTFRAME_COUNT_COHERENCE"); path != "" {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if _, err = f.Write(append(b, '\n')); err != nil {
			t.Fatal(err)
		}
		if err = f.Sync(); err != nil {
			t.Fatal(err)
		}
	}
}
