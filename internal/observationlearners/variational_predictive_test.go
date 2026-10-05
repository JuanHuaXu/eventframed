package observationlearners

import (
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/transfergenerator"
)

// Independent dense Simpson integration, wider domain, and tanh logistic.
// This numerical reference is not a posterior-approximation correctness proof.
func variationalReferenceIntegral(mean, variance float64) float64 {
	const n = 8192
	const h = 24. / n
	sum := 0.
	for i := 0; i <= n; i++ {
		z := -12 + float64(i)*h
		w := 2.
		if i%2 == 1 {
			w = 4
		}
		if i == 0 || i == n {
			w = 1
		}
		sum += w * (1 + math.Tanh((mean+math.Sqrt(variance)*z)/2)) / 2 * math.Exp(-z*z/2) / math.Sqrt(2*math.Pi)
	}
	return sum * h / 3
}

func TestVariationalIntegralReference(t *testing.T) {
	sum, second := 0., 0.
	for i, w := range variationalNormalWeights {
		if !(w > 0) {
			t.Fatal("nonpositive weight")
		}
		z := -10 + float64(i)*20/256
		sum += w
		second += w * z * z
	}
	if math.Abs(sum-1) > 2e-15 || math.Abs(second-1) > 2e-15 {
		t.Fatal("quadrature measure", sum, second)
	}
	if math.Erfc(10/math.Sqrt2) >= 1.6e-23 {
		t.Fatal("tail claim")
	}
	means := []float64{-1280, -512, -100, -50, 50, 100, 512, 1280}
	for i := -80; i <= 80; i++ {
		means = append(means, float64(i)/2)
	}
	maxErr := 0.
	for _, m := range means {
		for _, v := range []float64{0, 1e-6, .01, .1, .5, 1, 2, 5, 10} {
			got, err := variationalIntegral(m, v)
			if err != nil {
				t.Fatal(err)
			}
			want := variationalReferenceIntegral(m, v)
			e := math.Abs(got - want)
			maxErr = math.Max(maxErr, e)
			if e > 1e-10 {
				t.Fatal("integral", m, v, got, want)
			}
			opposite, err := variationalIntegral(-m, v)
			if err != nil || math.Abs(got+opposite-1) > 2e-15 {
				t.Fatal("complement", m, v)
			}
		}
	}
	t.Logf("%d reference comparisons; max absolute error %.17g", len(means)*9, maxErr)
	for _, v := range []float64{.01, 1, 10} {
		p, _ := variationalIntegral(2, v)
		if p <= .5 || p >= ridgeSigmoid(2) {
			t.Fatal("uncertainty averaging not present", v, p)
		}
	}
}

func TestVariationalConditionalCoherence(t *testing.T) {
	m, err := fitVariationalLogistic(ridgeTestSamples(64))
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
					t.Fatal("partial law", mask, values, err)
				}
				if values == 0 {
					break
				}
			}
		}
		before, _ := law.Forecast(0, 0)
		saved := m.mean
		m.mean[0] += 99
		if after, _ := law.Forecast(0, 0); before != after {
			t.Fatal("compiled alias")
		}
		m.mean = saved
	}
}

func TestVariationalPredictionBounds(t *testing.T) {
	for _, v := range []float64{-1, math.NaN(), math.Inf(1), 10 + 1e-8} {
		if _, err := variationalIntegral(0, v); err == nil {
			t.Fatal("invalid variance")
		}
	}
	for _, m := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := variationalIntegral(m, 1); err == nil {
			t.Fatal("invalid mean")
		}
	}
	if _, err := (*variationalLogistic)(nil).predict(0); err == nil {
		t.Fatal("nil model")
	}
	if _, err := new(variationalLogistic).predict(0); err == nil {
		t.Fatal("unfitted model")
	}
	m, err := fitVariationalLogistic(nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.predict(512); err == nil {
		t.Fatal("invalid query")
	}
	for _, z := range []float64{-1280, 1280} {
		m.mean[0] = z
		p, err := m.predict(0)
		if err != nil || p < ridgeProbabilityFloor || p > 1-ridgeProbabilityFloor {
			t.Fatal("floor", p, err)
		}
	}
	m.mean[0] = 0
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
	m.cov[0][0] = math.NaN()
	if _, err := m.predict(0); err == nil {
		t.Fatal("nonfinite covariance")
	}
}

func TestVariationalConsumedGeneratorQA(t *testing.T) {
	maxIterations := 0
	for f := transfergenerator.Additive; f <= transfergenerator.LocalTable; f++ {
		for mode := transfergenerator.Stationary; mode <= transfergenerator.Gradual; mode++ {
			for index := 0; index < 32; index++ {
				data, _, err := transfergenerator.Generate(transfergenerator.Spec{SeedBase: transfergenerator.ValidationSeedBase, Family: f, Mode: mode, Index: index})
				if err != nil {
					t.Fatal(err)
				}
				for _, size := range []int{16, 32, 64, 256} {
					s := make([]observation.Sample, size)
					for i := range s {
						s[i] = observation.Sample{Bits: data.Frames[i].X, Outcome: data.Frames[i].Y}
					}
					m, err := fitVariationalLogistic(s)
					if err != nil {
						t.Fatal("consumed QA", f, mode, index, size, err)
					}
					maxIterations = max(maxIterations, m.iterations)
					if m.residual > 1e-10 {
						t.Fatal("fixed point")
					}
					for bits := uint16(0); bits < 512; bits++ {
						p, err := m.predict(bits)
						if err != nil || p < 0 || p > 1 {
							t.Fatal("QA forecast", f, mode, index, size, bits, err)
						}
					}
				}
			}
		}
	}
	t.Logf("1152 consumed fits and 589824 predictions; maximum iterations %d", maxIterations)
}
