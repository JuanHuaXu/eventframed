package observationlearners

import (
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type spikeExactResult struct{ inclusion, prediction, logBF float64 }

// Diagnostic only: direct original-likelihood integration, not the JJ surrogate.
func spikeExactOne(samples []observation.Sample, panels int) spikeExactResult {
	var counts [2][2]int
	for _, s := range samples {
		x := int(s.Bits & 1)
		y := 0
		if s.Outcome {
			y = 1
		}
		counts[x][y]++
	}
	logSig := func(x float64) float64 {
		if x >= 0 {
			return -math.Log1p(math.Exp(-x))
		}
		return x - math.Log1p(math.Exp(x))
	}
	likelihood := func(a, b float64) float64 {
		minus, plus := a-b, a+b
		return math.Exp(float64(counts[0][1])*logSig(minus) + float64(counts[0][0])*logSig(-minus) + float64(counts[1][1])*logSig(plus) + float64(counts[1][0])*logSig(-plus))
	}
	step := 20 / float64(panels)
	x, w := make([]float64, panels), make([]float64, panels)
	for i := range x {
		x[i] = -10 + (float64(i)+.5)*step
		w[i] = step * math.Exp(-x[i]*x[i]/2) / math.Sqrt(2*math.Pi)
	}
	z0, z1, p0, p1 := 0., 0., 0., 0.
	for i, a := range x {
		null := w[i] * likelihood(a, 0)
		z0 += null
		p0 += null * ridgeSigmoid(a)
		for j, b := range x {
			weight := w[i] * w[j] * likelihood(a, b)
			z1 += weight
			p1 += weight * ridgeSigmoid(a+b)
		}
	}
	const pi = 1.0 / 255
	z := (1-pi)*z0 + pi*z1
	return spikeExactResult{pi * z1 / z, ((1-pi)*p0 + pi*p1) / z, math.Log(z1) - math.Log(z0)}
}

func TestSpikeExactPosteriorDiagnostic(t *testing.T) {
	for _, n := range []int{16, 64} {
		for _, flips := range []int{0, 2} {
			s := make([]observation.Sample, n)
			for i := range s {
				x := uint16(i % 2)
				y := x == 1
				if i/2 < flips {
					y = !y
				}
				s[i] = observation.Sample{Bits: x, Outcome: y}
			}
			a, b := spikeExactOne(s, 512), spikeExactOne(s, 1024)
			if math.Abs(a.inclusion-b.inclusion) > 1e-8 || math.Abs(a.prediction-b.prediction) > 1e-8 || math.Abs(a.logBF-b.logBF) > 1e-8 {
				t.Fatal("quadrature resolution", n, flips, a, b)
			}
			f, err := fitSpikeFixed(s, []uint16{1}, 1.0/255, 1, 1024)
			if err != nil {
				t.Fatal(err)
			}
			p, _, err := f.predict(1)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("n=%d flips_per_class=%d exactBFlog=%.12g exactInclusion=%.12g VIInclusion=%.12g exactP=%.12g VIP=%.12g", n, flips, b.logBF, b.inclusion, f.q.probability(0), b.prediction, p)
			for i := range s {
				s[i].Outcome = !s[i].Outcome
			}
			c := spikeExactOne(s, 512)
			if math.Abs(c.prediction+a.prediction-1) > 1e-8 || math.Abs(c.inclusion-a.inclusion) > 1e-8 {
				t.Fatal("exact posterior complement")
			}
		}
	}
}
