package observationlearners

import (
	"errors"
	"math"
	"testing"
)

// Retained initial all-log kernel, before transition arithmetic optimization.
func hazardAdviceLogReferenceStep(s *hazardAdviceState, e markovAdviceEntry) error {
	copy := *s
	rates := hazardRates()
	maximum := math.Inf(-1)
	for h, alpha := range rates {
		row := copy.logs[h]
		mass := math.Inf(-1)
		for j := 1; j < 5; j++ {
			if e.known {
				p := e.raw[j-1]
				ll := math.Log1p(-p)
				if e.outcome {
					ll = math.Log(p)
				}
				row[j] += ll
			}
			mass = hazardLogAdd(mass, row[j])
		}
		for j := range row {
			copy.logs[h][j] = hazardLogAdd(math.Log1p(-alpha)+row[j], math.Log(alpha)+math.Log(s.prior[j])+mass)
			maximum = math.Max(maximum, copy.logs[h][j])
		}
	}
	if math.IsInf(maximum, 0) || math.IsNaN(maximum) {
		return errors.New("invalid hazard joint evidence")
	}
	for h := range copy.logs {
		for j := range copy.logs[h] {
			copy.logs[h][j] -= maximum
		}
	}
	*s = copy
	return nil
}

func TestHazardAdviceTransitionReference(t *testing.T) {
	f, _ := newHazardAdvice(hazardPrior())
	a, b := f.current, f.current
	for i := uint64(0); i < 256; i++ {
		e := markovAdviceEntry{active: true, origin: i, known: i%7 != 0, outcome: i%5 < 2, raw: [4]float64{.000001, .999999, .2 + .1*float64(i%6), .5}}
		if err := hazardAdviceStep(&a, e); err != nil {
			t.Fatal(err)
		}
		if err := hazardAdviceLogReferenceStep(&b, e); err != nil {
			t.Fatal(err)
		}
		for h := range a.logs {
			for j := range a.logs[h] {
				x, y := a.logs[h][j], b.logs[h][j]
				if math.IsInf(x, -1) && math.IsInf(y, -1) {
					continue
				}
				if math.IsNaN(x) || math.IsNaN(y) || math.Abs(x-y) > 1e-9 {
					t.Fatal("transition log mismatch", i, h, j, x, y)
				}
			}
		}
	}
}
