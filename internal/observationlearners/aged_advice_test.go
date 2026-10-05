package observationlearners

import (
	"math"
	"math/rand"
	"testing"
)

func adviceNear(t testing.TB, a, b float64) {
	t.Helper()
	if math.IsNaN(a) || math.IsNaN(b) || math.Abs(a-b) > 1e-11 {
		t.Fatalf("advice mismatch %.17g %.17g", a, b)
	}
}

func TestAgedAdviceBatch(t *testing.T) {
	type packet struct {
		origin, arrival uint64
		p               [4]float64
		y               bool
	}
	rng := rand.New(rand.NewSource(8128))
	var packets []packet
	for i := 0; i < 256; i++ {
		p := packet{origin: uint64(i), arrival: uint64(i + rng.Intn(32)), y: rng.Intn(2) == 1}
		for j := range p.p {
			p.p[j] = .01 + .98*rng.Float64()
		}
		if i%5 != 0 {
			packets = append(packets, p)
		}
	}
	for _, neutral := range []bool{false, true} {
		s := newAgedAdvice(neutral, true)
		reverse := s
		var received []packet
		for clock := uint64(0); clock <= 288; clock++ {
			if err := s.advance(clock); err != nil {
				t.Fatal(err)
			}
			if err := reverse.advance(clock); err != nil {
				t.Fatal(err)
			}
			for _, p := range packets {
				if p.arrival == clock {
					if err := s.observe(p.origin, p.p, p.y); err != nil {
						t.Fatal(err)
					}
					received = append(received, p)
				}
			}
			for i := len(packets) - 1; i >= 0; i-- {
				p := packets[i]
				if p.arrival == clock {
					if err := reverse.observe(p.origin, p.p, p.y); err != nil {
						t.Fatal(err)
					}
				}
			}
			var literal [5]float64
			for _, p := range received {
				target := 0.
				if p.y {
					target = 1
				}
				row := [5]float64{.5, p.p[0], p.p[1], p.p[2], p.p[3]}
				for j, x := range row {
					literal[j] += math.Pow(2, -float64(clock-p.origin)/32) * (x - target) * (x - target)
				}
			}
			var w [5]float64
			sum := 0.
			for j, x := range literal {
				adviceNear(t, s.losses[j], x)
				adviceNear(t, reverse.losses[j], x)
				w[j] = s.prior[j] * math.Exp(-.5*x)
				sum += w[j]
			}
			got := s.weights()
			for j := range w {
				adviceNear(t, got[j], w[j]/sum)
			}
		}
	}
}

func TestAgedAdviceLimits(t *testing.T) {
	s := newAgedAdvice(false, false)
	ref, _ := newBrierBank([]float64{.95, .05 / 3, .05 / 3, .05 / 3}, .001)
	for i := uint64(0); i < 256; i++ {
		if err := s.advance(i); err != nil {
			t.Fatal(err)
		}
		p := [4]float64{.1 + .8*float64(i%3)/2, .2, .8, .5}
		f, err := ref.predict(i, p[:])
		if err != nil {
			t.Fatal(err)
		}
		w := s.weights()
		adviceNear(t, w[0], 0)
		for j := range f.Weights {
			adviceNear(t, w[j+1], f.Weights[j])
		}
		if err := ref.observe(i, i%3 == 0); err != nil {
			t.Fatal(err)
		}
		if err := s.observe(i, p, i%3 == 0); err != nil {
			t.Fatal(err)
		}
	}
	a := newAgedAdvice(true, true)
	if err := a.advance(32); err != nil {
		t.Fatal(err)
	}
	if err := a.observe(0, [4]float64{.1, .2, .3, .4}, true); err != nil {
		t.Fatal(err)
	}
	adviceNear(t, a.losses[1], .5*.9*.9)
	if err := a.advance(64); err != nil {
		t.Fatal(err)
	}
	adviceNear(t, a.losses[1], .25*.9*.9)
	before := a
	for _, op := range []func() error{func() error { return a.advance(63) }, func() error { return a.advance(289) }, func() error { return a.observe(65, [4]float64{}, true) }, func() error { return a.observe(0, [4]float64{math.NaN()}, true) }} {
		if err := op(); err == nil || a != before {
			t.Fatal("invalid advice mutated state")
		}
	}
	var zero agedAdvice
	if err := zero.advance(0); err == nil {
		t.Fatal("zero advice accepted")
	}
}

func TestAgedAdviceRejectionAuthority(t *testing.T) {
	for rejected := 1; rejected < 16; rejected++ {
		var gate evidenceRouting
		gate.gate.issued = 1
		for j := 0; j < 4; j++ {
			gate.gate.tests[j].Rejected = rejected&(1<<j) != 0
		}
		p := [4]float64{.1, .3, .7, .9}
		w := [5]float64{.2, .2, .2, .2, .2}
		got, err := adviceRoutedWeights(&gate, 1, p, w)
		if err != nil {
			t.Fatal(err)
		}
		sum := 0.
		for _, v := range got.Weights {
			sum += v
		}
		adviceNear(t, sum, 1)
		for j := 0; j < 4; j++ {
			if rejected&(1<<j) != 0 {
				adviceNear(t, got.Weights[j+1], 0)
			}
		}
		if got.Weights[0] < .2 {
			t.Fatal("neutral mass lost")
		}
		want := .5 * got.Weights[0]
		for j, x := range p {
			want += x * got.Weights[j+1]
		}
		adviceNear(t, want, got.P)
	}
}

func TestAgedAdviceNeutralRoundoff(t *testing.T) {
	var gate evidenceRouting
	gate.gate.issued = 1
	for j := range gate.gate.tests {
		gate.gate.tests[j].Rejected = true
	}
	w := [5]float64{0, .1, .2, .3, .40000000000000013}
	got, err := adviceRoutedWeights(&gate, 1, [4]float64{.1, .3, .7, .9}, w)
	if err != nil || got.Weights != ([5]float64{1}) || got.P != .5 {
		t.Fatal("all-neutral composition failed exact boundary", got, err)
	}
	before := gate
	if _, err := adviceRoutedWeights(&gate, 2, [4]float64{}, [5]float64{math.NaN()}); err == nil || gate != before {
		t.Fatal("invalid mixture mutated gate")
	}
}
