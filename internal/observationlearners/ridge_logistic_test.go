package observationlearners

import (
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/transfergenerator"
)

func ridgeTestSamples(n int) []observation.Sample {
	out := make([]observation.Sample, n)
	for i := range out {
		out[i] = observation.Sample{Bits: uint16((i*73 + i*i*11) % 512), Outcome: i%3 == 0 || i%7 == 1}
	}
	return out
}

func TestRidgeDerivatives(t *testing.T) {
	samples := ridgeTestSamples(32)
	var beta [ridgeDimension]float64
	for i := range beta {
		beta[i] = float64(i-4) * .12
	}
	// Independent likelihood expression at moderate logits, where direct exp
	// is safe, checks the stable implementation rather than repeating it.
	reference := func(b [ridgeDimension]float64) float64 {
		f := 0.
		for _, v := range b {
			f += v * v / 2
		}
		for _, s := range samples {
			z := b[0]
			for j := 0; j < 9; j++ {
				if s.Bits&(1<<j) != 0 {
					z += b[j+1]
				} else {
					z -= b[j+1]
				}
			}
			f += math.Log1p(math.Exp(z))
			if s.Outcome {
				f -= z
			}
		}
		return f
	}
	if math.Abs(reference(beta)-ridgeObjective(samples, beta)) > 1e-12 {
		t.Fatal("objective convention")
	}
	g, h := ridgeDerivatives(samples, beta)
	const step = 1e-5
	for i := range beta {
		plus, minus := beta, beta
		plus[i] += step
		minus[i] -= step
		if math.Abs((reference(plus)-reference(minus))/(2*step)-g[i]) > 1e-7 {
			t.Fatal("gradient", i)
		}
		gp, _ := ridgeDerivatives(samples, plus)
		gm, _ := ridgeDerivatives(samples, minus)
		for j := range beta {
			if math.Abs((gp[j]-gm[j])/(2*step)-h[j][i]) > 2e-8 {
				t.Fatal("Hessian", i, j)
			}
		}
	}
	d, err := ridgeSolve(h, g)
	if err != nil {
		t.Fatal(err)
	}
	for i := range g {
		v := -g[i]
		for j := range d {
			v += h[i][j] * d[j]
		}
		if math.Abs(v) > 1e-11 {
			t.Fatal("linear solve residual", i, v)
		}
	}
	if _, err := ridgeSolve([ridgeDimension][ridgeDimension]float64{}, g); err == nil {
		t.Fatal("singular solve accepted")
	}
}

func TestRidgeAnalyticDegenerateFit(t *testing.T) {
	for _, n := range []int{1, 16, 256} {
		for _, x := range []uint16{0, 17, 511} {
			for _, y := range []bool{false, true} {
				samples := make([]observation.Sample, n)
				for i := range samples {
					samples[i] = observation.Sample{Bits: x, Outcome: y}
				}
				m, err := fitRidgeLogistic(samples)
				if err != nil {
					t.Fatal(n, x, y, err)
				}
				// Identical rows force beta=c*x (or its negative). The unique
				// scalar root is c=N/(1+exp(10c)), solved independently by bisection.
				lo, hi := 0., float64(n)
				for i := 0; i < 100; i++ {
					mid := (lo + hi) / 2
					if mid-float64(n)/(1+math.Exp(10*mid)) > 0 {
						hi = mid
					} else {
						lo = mid
					}
				}
				c := (lo + hi) / 2
				if !y {
					c = -c
				}
				features := ridgeFeatures(x)
				for i, v := range m.beta {
					if math.Abs(v-c*features[i]) > 1e-8 {
						t.Fatal("analytic optimum", n, x, y, i, v, c)
					}
				}
				if m.stats.GradientInf > ridgeTolerance || m.stats.Iterations > 32 || m.stats.Evaluations > 769 || m.stats.Objective > float64(n)*math.Log(2) {
					t.Fatal("fit bounds")
				}
			}
		}
	}
}

func TestRidgeSymmetryAndOwnership(t *testing.T) {
	samples := ridgeTestSamples(64)
	m, err := fitRidgeLogistic(samples)
	if err != nil {
		t.Fatal(err)
	}
	again, err := fitRidgeLogistic(samples)
	if err != nil || *m != *again {
		t.Fatal("deterministic fit")
	}
	neg := append([]observation.Sample(nil), samples...)
	permuted := append([]observation.Sample(nil), samples...)
	rotate := func(x uint16) uint16 { return ((x << 1) & 511) | (x >> 8) }
	for i := range neg {
		neg[i].Outcome = !neg[i].Outcome
		permuted[i].Bits = rotate(permuted[i].Bits)
	}
	n, err := fitRidgeLogistic(neg)
	if err != nil {
		t.Fatal(err)
	}
	p, err := fitRidgeLogistic(permuted)
	if err != nil {
		t.Fatal(err)
	}
	for x := uint16(0); x < 512; x++ {
		a, _ := m.predict(x)
		b, _ := n.predict(x)
		c, _ := p.predict(rotate(x))
		if math.Abs(a+b-1) > 1e-8 || math.Abs(a-c) > 1e-8 {
			t.Fatal("symmetry", x)
		}
	}
	saved := *m
	samples[0] = observation.Sample{Bits: 511, Outcome: !samples[0].Outcome}
	if *m != saved {
		t.Fatal("training storage retained")
	}
}

func TestRidgeConditionalCoherence(t *testing.T) {
	m, err := fitRidgeLogistic(ridgeTestSamples(64))
	if err != nil {
		t.Fatal(err)
	}
	for _, weighted := range []bool{false, true} {
		var w, full [512]float64
		for x := range w {
			w[x] = 1
			if weighted {
				w[x] += float64(x * 17 % 13)
			}
			full[x], err = m.predict(uint16(x))
			if err != nil {
				t.Fatal(err)
			}
		}
		law, err := m.conditional(w)
		if err != nil {
			t.Fatal(err)
		}
		for mask := uint16(0); mask < 512; mask++ {
			for values := mask; ; values = (values - 1) & mask {
				sum, mass := 0., 0.
				for x, p := range full {
					if uint16(x)&mask == values {
						sum += w[x] * p
						mass += w[x]
					}
				}
				got, err := law.Forecast(mask, values)
				if err != nil || math.Abs(got-sum/mass) > 1e-13 {
					t.Fatal("conditional law", mask, values, got, sum/mass, err)
				}
				if values == 0 {
					break
				}
			}
		}
		before, _ := law.Forecast(0, 0)
		saved := m.beta
		m.beta[0] += 99
		if after, _ := law.Forecast(0, 0); before != after {
			t.Fatal("compiled alias")
		}
		m.beta = saved
	}
}

func TestRidgeSmallStepObjectiveDelta(t *testing.T) {
	data, _, err := transfergenerator.Generate(transfergenerator.Spec{SeedBase: transfergenerator.ValidationSeedBase, Index: 21})
	if err != nil {
		t.Fatal(err)
	}
	samples := make([]observation.Sample, 256)
	for i, row := range data.Frames {
		samples[i] = observation.Sample{Bits: row.X, Outcome: row.Y}
	}
	m, err := fitRidgeLogistic(samples)
	if err != nil {
		t.Fatal("roundoff regression", err)
	}
	beta := m.beta
	for i := range beta {
		beta[i] += float64(i+1) * 1e-7
	}
	g, h := ridgeDerivatives(samples, beta)
	d, err := ridgeSolve(h, g)
	if err != nil {
		t.Fatal(err)
	}
	for i := range d {
		d[i] = -d[i]
	}
	// Independent line integral of the already finite-difference-checked
	// gradient. Simpson error is negligible for this tiny smooth step.
	points := [3]float64{0, .5, 1}
	weights := [3]float64{1. / 6, 4. / 6, 1. / 6}
	want := 0.
	for j, at := range points {
		b := beta
		for i := range b {
			b[i] += at * d[i]
		}
		grad, _ := ridgeDerivatives(samples, b)
		want += weights[j] * ridgeDot(grad, d)
	}
	got := ridgeObjectiveDelta(samples, beta, d)
	if got >= 0 || math.Abs(got-want) > 1e-17 {
		t.Fatal("small-step difference", got, want)
	}
	for i := range d {
		d[i] = float64(i-3) * .1
	}
	trial := beta
	for i := range trial {
		trial[i] += d[i]
	}
	if math.Abs(ridgeObjectiveDelta(samples, beta, d)-(ridgeObjective(samples, trial)-ridgeObjective(samples, beta))) > 1e-11 {
		t.Fatal("large-step difference")
	}
}

func TestRidgeBoundsAndExtremeLogits(t *testing.T) {
	for _, samples := range [][]observation.Sample{nil, make([]observation.Sample, 257), {{Bits: 512}}} {
		if m, err := fitRidgeLogistic(samples); err == nil || m != nil {
			t.Fatal("invalid fit accepted")
		}
	}
	for _, limit := range []int{0, 1, 33} {
		if m, err := fitRidgeLogisticLimit([]observation.Sample{{Outcome: true}}, limit); err == nil || m != nil {
			t.Fatal("invalid/unconverged fit accepted", limit)
		}
	}
	for _, z := range []float64{-1000, 0, 1000} {
		for _, y := range []bool{false, true} {
			v := ridgeNLL(z, y)
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
				t.Fatal("unstable loss")
			}
		}
		m := &ridgeLogistic{ready: true}
		m.beta[0] = z
		p, err := m.predict(0)
		if err != nil || p < ridgeProbabilityFloor || p > 1-ridgeProbabilityFloor {
			t.Fatal("forecast floor")
		}
	}
	if _, err := (*ridgeLogistic)(nil).predict(0); err == nil {
		t.Fatal("nil forecast")
	}
	if _, err := new(ridgeLogistic).predict(0); err == nil {
		t.Fatal("uninitialized forecast")
	}
	m, err := fitRidgeLogistic(ridgeTestSamples(32))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.predict(512); err == nil {
		t.Fatal("invalid query")
	}
	for _, bad := range []float64{0, -1, math.NaN(), math.Inf(1), 1e6 + 1} {
		var w [512]float64
		for i := range w {
			w[i] = 1
		}
		w[17] = bad
		if _, err := m.conditional(w); err == nil {
			t.Fatal("invalid mass")
		}
	}
	m.beta[0] = math.NaN()
	if _, err := m.predict(0); err == nil {
		t.Fatal("non-finite logit")
	}
}

func TestRidgeConsumedGeneratorQA(t *testing.T) {
	for f := transfergenerator.Additive; f <= transfergenerator.LocalTable; f++ {
		for mode := transfergenerator.Stationary; mode <= transfergenerator.Gradual; mode++ {
			for index := 0; index < 32; index++ {
				data, _, err := transfergenerator.Generate(transfergenerator.Spec{SeedBase: transfergenerator.ValidationSeedBase, Family: f, Mode: mode, Index: index})
				if err != nil {
					t.Fatal(err)
				}
				for _, size := range []int{16, 32, 64, 256} {
					samples := make([]observation.Sample, size)
					for i := range samples {
						samples[i] = observation.Sample{Bits: data.Frames[i].X, Outcome: data.Frames[i].Y}
					}
					before := append([]observation.Sample(nil), samples...)
					m, err := fitRidgeLogistic(samples)
					if err != nil {
						t.Fatal("consumed QA convergence", f, mode, index, size, err)
					}
					if !reflect.DeepEqual(before, samples) || m.stats.GradientInf > ridgeTolerance {
						t.Fatal("mutation or residual")
					}
				}
			}
		}
	}
}
