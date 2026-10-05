package observationlearners

import (
	"math"
	"testing"
)

// Cross-check the independent artifact verifier's cheaper integration rule
// against the dense component reference before using it on fresh quality data.
func TestVariationalArtifactIntegral(t *testing.T) {
	maxError := 0.
	for i := -80; i <= 80; i++ {
		for _, v := range []float64{0, 1e-6, .01, .1, .5, 1, 2, 5, 10} {
			m := float64(i) / 2
			sum := 0.
			for j := 0; j <= 512; j++ {
				z := -12 + float64(j)*24/512
				w := 2.
				if j%2 == 1 {
					w = 4
				}
				if j == 0 || j == 512 {
					w = 1
				}
				sum += w * (1 + math.Tanh((m+math.Sqrt(v)*z)/2)) / 2 * math.Exp(-z*z/2) / math.Sqrt(2*math.Pi)
			}
			e := math.Abs(sum*24/512/3 - variationalReferenceIntegral(m, v))
			maxError = math.Max(maxError, e)
			if e > 1e-10 {
				t.Fatal("artifact integral", m, v, e)
			}
		}
	}
	t.Logf("1449 artifact-reference checks; max error %.17g", maxError)
}
