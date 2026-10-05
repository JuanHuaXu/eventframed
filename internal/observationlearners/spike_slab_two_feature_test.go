package observationlearners

import (
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type spikeExactTwoResult struct {
	states, predictions [4]float64
}

// Enumerate inclusion states and integrate the original likelihood, not VI factors.
func spikeExactTwo(samples []observation.Sample, panels int) spikeExactTwoResult {
	var counts [4][2]int
	for _, s := range samples {
		y := 0
		if s.Outcome {
			y = 1
		}
		counts[s.Bits&3][y]++
	}
	x, w := make([]float64, panels), make([]float64, panels)
	step := 20 / float64(panels)
	for i := range x {
		x[i] = -10 + (float64(i)+.5)*step
		w[i] = step * math.Exp(-x[i]*x[i]/2) / math.Sqrt(2*math.Pi)
	}
	logSig := func(z float64) float64 {
		if z >= 0 {
			return -math.Log1p(math.Exp(-z))
		}
		return z - math.Log1p(math.Exp(z))
	}
	var result spikeExactTwoResult
	const pi = 1.0 / 255
	for state := 0; state < 4; state++ {
		x1, w1, x2, w2 := []float64{0}, []float64{1}, []float64{0}, []float64{1}
		prior := 1.
		if state&1 != 0 {
			x1, w1 = x, w
			prior *= pi
		} else {
			prior *= 1 - pi
		}
		if state&2 != 0 {
			x2, w2 = x, w
			prior *= pi
		} else {
			prior *= 1 - pi
		}
		for i, a := range x {
			for j, b := range x1 {
				for k, c := range x2 {
					logits := [4]float64{a - b - c, a + b - c, a - b + c, a + b + c}
					logL := 0.
					for q, z := range logits {
						if counts[q][0] != 0 {
							logL += float64(counts[q][0]) * logSig(-z)
						}
						if counts[q][1] != 0 {
							logL += float64(counts[q][1]) * logSig(z)
						}
					}
					mass := prior * w[i] * w1[j] * w2[k] * math.Exp(logL)
					result.states[state] += mass
					for q, z := range logits {
						result.predictions[q] += mass * ridgeSigmoid(z)
					}
				}
			}
		}
	}
	z := 0.
	for _, mass := range result.states {
		z += mass
	}
	for q := range result.states {
		result.states[q] /= z
		result.predictions[q] /= z
	}
	return result
}

func TestSpikeTwoFeaturePosterior(t *testing.T) {
	for _, n := range []int{16, 64} {
		for _, duplicate := range []bool{true, false} {
			s := make([]observation.Sample, n)
			for i := range s {
				x := uint16(i % 4)
				if duplicate {
					x = uint16(i%2) * 3
				}
				s[i] = observation.Sample{Bits: x, Outcome: x&1 != 0}
			}
			a, b := spikeExactTwo(s, 128), spikeExactTwo(s, 192)
			for q := 0; q < 4; q++ {
				if !spikeFinite(b.predictions[q]) || math.Abs(a.predictions[q]-b.predictions[q]) > 2e-7 || math.Abs(a.states[q]-b.states[q]) > 2e-7 {
					t.Fatal("quadrature refinement", n, duplicate, a, b)
				}
			}
			if duplicate && (math.Abs(b.states[1]-b.states[2]) > 1e-10 || math.Abs(b.predictions[1]-.5) > 1e-10 || math.Abs(b.predictions[2]-.5) > 1e-10) {
				t.Fatal("reference exchange symmetry")
			}
			if !duplicate && b.predictions[1] <= .6 {
				t.Fatal("independent design lost information")
			}
			f, err := fitSpikeFixed(s, []uint16{1, 2}, piForSpikeDiagnostic, 1, 1024)
			if err != nil {
				t.Fatal(err)
			}
			r, err := fitSpikeFixed(s, []uint16{2, 1}, piForSpikeDiagnostic, 1, 1024)
			if err != nil {
				t.Fatal(err)
			}
			var forward, reverse, ensemble [4]float64
			maxForward, maxEnsemble := 0., 0.
			for q := uint16(0); q < 4; q++ {
				forward[q], _, err = f.predict(q)
				if err != nil {
					t.Fatal(err)
				}
				reverse[q], _, err = r.predict(q)
				if err != nil {
					t.Fatal(err)
				}
				ensemble[q] = (forward[q] + reverse[q]) / 2
				maxForward = math.Max(maxForward, math.Abs(forward[q]-b.predictions[q]))
				maxEnsemble = math.Max(maxEnsemble, math.Abs(ensemble[q]-b.predictions[q]))
			}
			t.Logf("n=%d duplicate=%t states=%v reference=%v forward=%v reverse=%v ensemble=%v maxForwardError=%g maxEnsembleError=%g", n, duplicate, b.states, b.predictions, forward, reverse, ensemble, maxForward, maxEnsemble)
		}
	}
}

const piForSpikeDiagnostic = 1.0 / 255
