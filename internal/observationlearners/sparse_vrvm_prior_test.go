package observationlearners

import (
	"math"
	"testing"
)

// The hyperbolic substitution resolves the narrow central peak without
// sampling the extremely heavy prior tail or drawing Gamma underflows.
func sparsePriorCentral(a, b, radius float64, panels int) float64 {
	u := math.Asinh(radius / math.Sqrt(2*b))
	l1, _ := math.Lgamma(a + .5)
	l0, _ := math.Lgamma(a)
	c := 2 * math.Exp(l1-l0-.5*math.Log(math.Pi))
	sum := 0.
	for i := 0; i < panels; i++ {
		x := (float64(i) + .5) * u / float64(panels)
		sum += math.Exp(-2 * a * math.Log(math.Cosh(x)))
	}
	return c * sum * u / float64(panels)
}

func TestSparsePriorCentralMass(t *testing.T) {
	// With a=b=1/2, the marginal is standard Cauchy, whose CDF is exact.
	for _, radius := range []float64{.1, 1, 5, 10} {
		got := sparsePriorCentral(.5, .5, radius, 65536)
		want := 2 * math.Atan(radius) / math.Pi
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("Cauchy identity: %.15g vs %.15g", got, want)
		}
	}
	l1, _ := math.Lgamma(sparsePrior + .5)
	l0, _ := math.Lgamma(sparsePrior)
	for _, radius := range []float64{1, 5, 10} {
		got := sparsePriorCentral(sparsePrior, sparsePrior, radius, 65536)
		coarse := sparsePriorCentral(sparsePrior, sparsePrior, radius, 4096)
		upper := 2 * math.Exp(l1-l0-.5*math.Log(math.Pi)) * math.Asinh(radius/math.Sqrt(2*sparsePrior))
		if got <= 0 || got > upper || math.Abs(got-coarse) > 1e-12 {
			t.Fatalf("central mass or bound: p=%g upper=%g coarse=%g", got, upper, coarse)
		}
		t.Logf("radius=%g central_probability=%.12g analytic_upper=%.12g", radius, got, upper)
	}
}
