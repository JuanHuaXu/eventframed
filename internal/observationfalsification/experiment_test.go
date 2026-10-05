package observationfalsification

import (
	"math"
	"testing"

	prev "github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

func TestSeedDomains(t *testing.T) {
	seen := map[int64]bool{FitSeed: true, prev.FitSeed: true}
	for _, base := range []int64{prev.DesignSeed, prev.ConfirmationSeed, DesignSeed, ConfirmationSeed} {
		for scenario := 0; scenario < 5; scenario++ {
			for stream := 0; stream < 32; stream++ {
				for role := 0; role < 14; role++ {
					s := prev.Seed(base, scenario, stream, role)
					if seen[s] {
						t.Fatal("seed reuse")
					}
					seen[s] = true
				}
			}
		}
	}
}
func TestCompetingPlansAndExploration(t *testing.T) {
	p := [4][2]float64{{.01, .01}, {.99, .01}, {.5, .5}, {.1, .1}}
	i, prob := Choice("single_mmm", p, false, .1)
	if i != 2 || prob != .8125 {
		t.Fatal("uncertainty observer not distinct")
	}
	i, prob = Choice("paired_mmm", p, false, .1)
	if i != 1 || prob != .8125 {
		t.Fatal("disagreement observer not distinct")
	}
	i, prob = Choice("paired_mmm", p, true, .9)
	if i != 3 || prob != .0625 {
		t.Fatal("independent exploration removed")
	}
	for j := range p {
		p[j][1] = p[j][0]
	}
	i, prob = Choice("paired_mmm", p, false, .9)
	if i != 3 || prob != .25 {
		t.Fatal("identical observers manufacture disagreement")
	}
	if divergence(.2, .8) != divergence(.8, .2) || math.Abs(divergence(.7, .7)) > 1e-14 {
		t.Fatal("invalid JS divergence")
	}
}
func TestStreamBoundaries(t *testing.T) {
	base, e := Base(false)
	if e != nil {
		t.Fatal(e)
	}
	r, e := RunStream(base, "unit", 0, 0, 2026092105)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Ticks) != 512 {
		t.Fatal("missing stream")
	}
	audits, version := 0, 0
	for _, tick := range r.Ticks {
		for _, p := range tick.Predictions {
			if p.Version != version || p.Cost > 6 {
				t.Fatal("future fitting or budget leak")
			}
		}
		if tick.Audit {
			audits++
			for _, p := range tick.Probes {
				if p.Probability < .0625 || p.Bits != tick.Pool[p.Choice] {
					t.Fatal("wrong admitted evidence")
				}
			}
			if audits >= 32 && (audits-32)%16 == 0 {
				version++
			}
		}
	}
	if r.ProbeCost != 36*audits || r.OutcomeQueries != 2*audits {
		t.Fatal("hidden acquisition cost")
	}
}
