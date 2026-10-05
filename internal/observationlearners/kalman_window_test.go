package observationlearners

import (
	"math"
	"testing"
)

func TestKalmanResidualFiniteSymmetric(t *testing.T) {
	k := NewKalmanResidual()
	for n := 0; n < 1000; n++ {
		k.Advance()
		x := uint16(n % 512)
		b := .05
		if n%2 == 0 {
			b = .95
		}
		h := kalmanFeatures(x, b)
		p := k.Predict(b, h)
		if p < .01 || p > .99 || math.IsNaN(p) {
			t.Fatalf("invalid probability at %d: %g", n, p)
		}
		if err := k.Observe(b, h, n%3 == 0); err != nil {
			t.Fatal(err)
		}
		for i := range k.Cov {
			if k.Cov[i][i] <= 0 {
				t.Fatalf("nonpositive variance at %d, %d", n, i)
			}
			for j := range k.Cov[i] {
				if k.Cov[i][j] != k.Cov[j][i] {
					t.Fatalf("asymmetric covariance at %d, %d, %d", n, i, j)
				}
			}
		}
	}
}

func TestKalmanWindowFeedbackIsAfterForecast(t *testing.T) {
	r, err := RunKalmanWindowStream("design", 7, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Ticks) != 512 || r.StateBytes > 4096 || r.Updates != r.Audits {
		t.Fatalf("invalid stream accounting: %+v", r)
	}
	for i := 0; i <= 16; i++ {
		if r.Ticks[i].Predictions[3] != r.Ticks[i].Predictions[0] {
			t.Fatalf("future feedback changed forecast at tick %d", i)
		}
		for _, origin := range r.Ticks[i].Delivered {
			if origin+16 != i {
				t.Fatalf("wrong delivery at tick %d from %d", i, origin)
			}
		}
	}
	if r.Pending > 16 || r.Available+r.Pending > 512 {
		t.Fatalf("invalid missing/delay accounting: %+v", r)
	}
}
