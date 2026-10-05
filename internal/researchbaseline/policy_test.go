package researchbaseline

import (
	"math"
	"testing"
)

func TestSelectAccuracyUnderBudget(t *testing.T) {
	c := []Candidate{
		{Name: "full", Cohort: "native", Cells: 120, ExpectedBrier: .2246990964333006, MeanCoreMS: 50.307, WorstCoreMS: 68.356, Comparable: true},
		{Name: "adaptive", Cohort: "native", Cells: 120, ExpectedBrier: .2147478547980669, MeanCoreMS: 201.944, WorstCoreMS: 211.952, Comparable: true},
		{Name: "prototype", Cohort: "component", Cells: 120, ExpectedBrier: .01, MeanCoreMS: 1, WorstCoreMS: 2, Comparable: false},
	}
	w, e := Select(c, "native")
	if e != nil || w.Name != PrimaryArm {
		t.Fatalf("selection: %+v %v", w, e)
	}
	c[1].WorstCoreMS = 401
	w, e = Select(c, "native")
	if e != nil || w.Name != SpeedControl {
		t.Fatalf("budget: %+v %v", w, e)
	}
}

func TestSelectRejectsIncompleteOrInvalidEvidence(t *testing.T) {
	base := Candidate{Name: "x", Cohort: "native", Cells: 120, ExpectedBrier: .2, MeanCoreMS: 1, WorstCoreMS: 2, Comparable: true}
	for _, mutate := range []func(*Candidate){
		func(c *Candidate) { c.Cells = 119 }, func(c *Candidate) { c.ExpectedBrier = math.NaN() },
		func(c *Candidate) { c.ExpectedBrier = math.Inf(1) }, func(c *Candidate) { c.ExpectedBrier = -.1 },
		func(c *Candidate) { c.MeanCoreMS = 0 }, func(c *Candidate) { c.WorstCoreMS = math.NaN() },
		func(c *Candidate) { c.WorstCoreMS = .5 },
	} {
		c := base
		mutate(&c)
		if _, e := Select([]Candidate{c}, "native"); e == nil {
			t.Fatalf("accepted invalid evidence: %+v", c)
		}
	}
	if _, e := Select([]Candidate{base}, ""); e == nil {
		t.Fatal("accepted unspecified scope")
	}
}
