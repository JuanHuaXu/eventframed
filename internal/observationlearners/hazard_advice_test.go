package observationlearners

import (
	"math"
	"testing"
)

// Sum unnormalized path probabilities across rates before one normalization.
// Normalizing each rate independently would test a different, incorrect model.
func hazardPaths(rows [][4]float64, known, ys []bool, ratesPrior [hazardRateCount]float64) ([5]float64, [hazardRateCount]float64) {
	prior := newAgedAdvice(false, false).prior
	var joint [hazardRateCount][5]float64
	for h, alpha := range hazardRates() {
		var walk func(int, int, float64)
		walk = func(state, step int, mass float64) {
			if step == len(rows) {
				joint[h][state] += mass
				return
			}
			if known[step] {
				p := rows[step][state-1]
				if !ys[step] {
					p = 1 - p
				}
				mass *= p
			}
			for next := 1; next < 5; next++ {
				p := alpha * prior[next]
				if next == state {
					p += 1 - alpha
				}
				walk(next, step+1, mass*p)
			}
		}
		for state := 1; state < 5; state++ {
			walk(state, 0, ratesPrior[h]*prior[state])
		}
	}
	var w [5]float64
	var rw [hazardRateCount]float64
	total := 0.
	for h, row := range joint {
		for j, v := range row {
			w[j] += v
			rw[h] += v
			total += v
		}
	}
	for j := range w {
		w[j] /= total
	}
	for h := range rw {
		rw[h] /= total
	}
	return w, rw
}

func TestHazardAdvicePaths(t *testing.T) {
	rows := [][4]float64{{.05, .95, .4, .6}, {.1, .9, .7, .3}, {.9, .1, .2, .8}, {.8, .2, .6, .4}}
	ys := []bool{true, true, false, true}
	var finals [][hazardRateCount]float64
	for _, order := range [][]int{{0, 1, 2, 3}, {3, 2, 1, 0}, {1, 3, 0, 2}} {
		f, err := newHazardAdvice(hazardPrior())
		if err != nil {
			t.Fatal(err)
		}
		known := make([]bool, len(rows))
		for i, row := range rows {
			if err := f.issue(uint64(i), row); err != nil {
				t.Fatal(err)
			}
		}
		for _, i := range order {
			if err := f.deliver(uint64(i), ys[i]); err != nil {
				t.Fatal(err)
			}
			known[i] = true
			want, rates := hazardPaths(rows, known, ys, hazardPrior())
			got, gotRates := f.current.weights(), f.current.rateWeights()
			for j := range want {
				adviceNear(t, got[j], want[j])
			}
			for h := range rates {
				adviceNear(t, gotRates[h], rates[h])
			}
		}
		finals = append(finals, f.current.rateWeights())
	}
	for _, r := range finals {
		for h := range r {
			adviceNear(t, r[h], finals[0][h])
		}
	}
	if math.Abs(finals[0][1]-.5) < 1e-4 {
		t.Fatal("fixture did not exercise rate learning")
	}
}

func TestHazardAdviceNeutralEvidence(t *testing.T) {
	for _, known := range []bool{false, true} {
		f, _ := newHazardAdvice(hazardPrior())
		for i := uint64(0); i < 64; i++ {
			if err := f.issue(i, [4]float64{.3, .3, .3, .3}); err != nil {
				t.Fatal(err)
			}
			if known {
				if err := f.deliver(i, i%2 == 0); err != nil {
					t.Fatal(err)
				}
			}
		}
		r := f.current.rateWeights()
		for h, p := range hazardPrior() {
			adviceNear(t, r[h], p)
		}
		w := f.current.weights()
		for j, p := range f.current.prior {
			adviceNear(t, w[j], p)
		}
		before := f
		if err := f.issue(65, [4]float64{.5, .5, .5, .5}); err == nil || f != before {
			t.Fatal("invalid issue mutation")
		}
	}
}

func TestHazardAdvicePointMass(t *testing.T) {
	for h, alpha := range hazardRates() {
		var prior [hazardRateCount]float64
		prior[h] = 1
		f, _ := newHazardAdvice(prior)
		ref, _ := newMarkovAdvice(alpha)
		for i := uint64(0); i < 64; i++ {
			row := [4]float64{.1 + .1*float64(i%7), .2, .8, .7}
			if err := f.issue(i, row); err != nil {
				t.Fatal(err)
			}
			if err := ref.issue(i, row); err != nil {
				t.Fatal(err)
			}
		}
		for i := 63; i >= 0; i-- {
			if i%7 == 0 {
				continue
			}
			if err := f.deliver(uint64(i), i%3 == 0); err != nil {
				t.Fatal(err)
			}
			if err := ref.deliver(uint64(i), i%3 == 0); err != nil {
				t.Fatal(err)
			}
			got, want := f.current.weights(), ref.current.weights()
			for j := range got {
				adviceNear(t, got[j], want[j])
			}
		}
		if err := f.expireBefore(64); err != nil {
			t.Fatal(err)
		}
		if err := ref.expireBefore(64); err != nil {
			t.Fatal(err)
		}
		got, want := f.current.weights(), ref.current.weights()
		for j := range got {
			adviceNear(t, got[j], want[j])
		}
		r := f.current.rateWeights()
		for k := range r {
			adviceNear(t, r[k], prior[k])
		}
	}
}

// Dense probability-domain reference with global normalization at every step.
func hazardDense(rows [][4]float64, known, ys []bool) ([5]float64, [hazardRateCount]float64) {
	prior := newAgedAdvice(false, false).prior
	var w [hazardRateCount][5]float64
	for h, p := range hazardPrior() {
		for j, q := range prior {
			w[h][j] = p * q
		}
	}
	for i, row := range rows {
		var next [hazardRateCount][5]float64
		total := 0.
		for h, alpha := range hazardRates() {
			for j := 1; j < 5; j++ {
				mass := w[h][j]
				if known[i] {
					p := row[j-1]
					if !ys[i] {
						p = 1 - p
					}
					mass *= p
				}
				for k := 1; k < 5; k++ {
					p := alpha * prior[k]
					if k == j {
						p += 1 - alpha
					}
					next[h][k] += mass * p
					total += mass * p
				}
			}
		}
		for h := range w {
			for j := range w[h] {
				w[h][j] = next[h][j] / total
			}
		}
	}
	var roles [5]float64
	var rates [hazardRateCount]float64
	for h, row := range w {
		for j, p := range row {
			roles[j] += p
			rates[h] += p
		}
	}
	return roles, rates
}

func TestHazardAdviceCheckpoint(t *testing.T) {
	f, _ := newHazardAdvice(hazardPrior())
	var rows [][4]float64
	var known, ys []bool
	check := func() {
		t.Helper()
		want, rates := hazardDense(rows, known, ys)
		got, r := f.current.weights(), f.current.rateWeights()
		for j := range got {
			adviceNear(t, got[j], want[j])
		}
		for h := range r {
			adviceNear(t, r[h], rates[h])
		}
	}
	for block := 0; block < 4; block++ {
		for i := block * 64; i < (block+1)*64; i++ {
			row := [4]float64{.1 + .1*float64(i%7), .2, .8, .7}
			if err := f.issue(uint64(i), row); err != nil {
				t.Fatal(err)
			}
			rows = append(rows, row)
			known = append(known, false)
			ys = append(ys, i%3 == 0)
		}
		before := f
		if err := f.issue(f.next, [4]float64{.5, .5, .5, .5}); err == nil || f != before {
			t.Fatal("capacity mutation")
		}
		for i := (block+1)*64 - 1; i >= block*64; i-- {
			if i%7 == 0 {
				continue
			}
			if err := f.deliver(uint64(i), ys[i]); err != nil {
				t.Fatal(err)
			}
			known[i] = true
			check()
			before = f
			if err := f.deliver(uint64(i), ys[i]); err == nil || f != before {
				t.Fatal("duplicate mutation")
			}
		}
		if err := f.expireBefore(f.next); err != nil {
			t.Fatal(err)
		}
		check()
		if f.base != f.next {
			t.Fatal("checkpoint")
		}
		before = f
		if err := f.deliver(uint64(block*64), true); err == nil || f != before {
			t.Fatal("settled mutation")
		}
	}
}

func TestHazardAdviceAtomicity(t *testing.T) {
	for _, v := range []float64{-1, 1.1, math.NaN(), math.Inf(1)} {
		p := hazardPrior()
		p[0] = v
		if _, err := newHazardAdvice(p); err == nil {
			t.Fatal("invalid prior")
		}
	}
	if _, err := newHazardAdvice([hazardRateCount]float64{}); err == nil {
		t.Fatal("empty prior")
	}
	f, _ := newHazardAdvice(hazardPrior())
	for _, v := range []float64{0, 1, math.NaN(), math.Inf(1)} {
		before := f
		if err := f.issue(0, [4]float64{.5, .5, .5, v}); err == nil || f != before {
			t.Fatal("invalid emission")
		}
	}
	before := f
	if err := f.deliver(0, true); err == nil || f != before {
		t.Fatal("future label")
	}
	if err := f.expireBefore(1); err == nil || f != before {
		t.Fatal("future expiry")
	}
}

func TestHazardAdviceUnderflow(t *testing.T) {
	var p [hazardRateCount]float64
	p[0] = 1
	f, _ := newHazardAdvice(p)
	for i := uint64(0); i < 240; i++ {
		if err := f.issue(i, [4]float64{.001, .999, .5, .5}); err != nil {
			t.Fatal(err)
		}
		if err := f.deliver(i, i < 120); err != nil {
			t.Fatal(err)
		}
		if i == 119 && f.current.weights()[1] != 0 {
			t.Fatal("underflow not reached")
		}
	}
	if math.Abs(f.current.logs[0][1]-f.current.logs[0][2]-math.Log(57)) > 1e-9 {
		t.Fatal("recoverable evidence erased")
	}
}
