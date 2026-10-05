package researchdispersion

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/researchbaseline"
)

// The policy is prospective; this audit reuses only the consumed development cohort.
func TestResearchBaselinePolicy20261005(t *testing.T) {
	output := os.Getenv("EVENTFRAME_BASELINE_AUDIT_OUT")
	if output == "" {
		t.Skip("explicit aggregate audit output required")
	}
	type cell struct {
		Geometry, Regime, Schedule, Arm                         string
		Risk, Priority, TerminalRisk, Utility, CoreMS, Recovery float64
	}
	rows := []cell{}
	for g := 0; g < 2; g++ {
		for r := range regimesPairedV60 {
			p := makePairedV60(2026105407, g, r, 0)
			for _, schedule := range schedulesV39 {
				for _, mode := range []string{researchbaseline.SpeedControl, researchbaseline.PrimaryArm} {
					a, e := runMomentV41(p.World, mode, schedule)
					if e != nil {
						t.Fatal(e)
					}
					scoreWindowV37(p.World, &a.windowArmV37)
					// Recompute expected loss directly from issued forecasts, not summary prose.
					independent := 0.
					for k, q := range a.Issued {
						v := p.World.Rates[k/150][k%150]
						independent += ((q-v)*(q-v) + v*(1-v)) / 2400
					}
					if math.Abs(independent-a.IssuedBrier) > 1e-12 {
						t.Fatal("independent loss mismatch")
					}
					last := a.Snapshots[len(a.Snapshots)-1]
					rows = append(rows, cell{p.World.Geometry, p.World.Regime, schedule, mode,
						a.IssuedBrier, a.IssuedPriority, last.Brier, last.PacketUsefulness,
						float64(a.Costs.ElapsedNS) / 1e6, a.Recovery})
				}
			}
		}
	}
	for mode, want := range map[string]float64{"full": .2246990964333006, "adaptive": .2147478547980669} {
		mean, n := 0., 0
		for _, r := range rows {
			if r.Arm == mode {
				mean += r.Risk
				n++
			}
		}
		if n != 120 || math.Abs(mean/float64(n)-want) > 1e-12 {
			t.Fatalf("frozen control mismatch %s: %g %d", mode, mean/float64(n), n)
		}
	}
	b, e := json.MarshalIndent(struct {
		Scope                 string
		DevelopmentSeed       int64
		Primary, SpeedControl string
		IndependentLossChecks int
		UntouchedConfirmation bool
		Cells                 []cell
	}{"Consumed 40-world/120-cell control replication; not serving or confirmation", 2026105407,
		researchbaseline.PrimaryArm, researchbaseline.SpeedControl, 240 * 2400, false, rows}, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	f, e := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.Write(append(b, '\n')); e != nil {
		t.Fatal(e)
	}
	if e = f.Sync(); e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	t.Log("240 control arms; 576000 independently recomputed losses; frozen risks retained")
}
