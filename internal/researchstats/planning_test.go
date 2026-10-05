package researchstats

import (
	"math"
	"math/big"
	"testing"
)

func TestNormalPlanSubnormalCancellation(t *testing.T) {
	for _, se := range []float64{math.SmallestNonzeroFloat64, 2 * math.SmallestNonzeroFloat64, 1e-200, .01} {
		units, err := FixedSampleNormalPlan(32, se, se, 3.5, .84162123357)
		if err != nil || units != 604 {
			t.Fatalf("equal SE/gain %.17g: got %d, want 604, err=%v", se, units, err)
		}
	}
	for _, c := range []struct {
		se, gain float64
		want     uint64
	}{
		{math.SmallestNonzeroFloat64, 2 * math.SmallestNonzeroFloat64, 151},
		{2 * math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64, 2413},
	} {
		units, err := FixedSampleNormalPlan(32, c.se, c.gain, 3.5, .84162123357)
		if err != nil || units != c.want {
			t.Fatalf("subnormal ratio %g/%g: got %d, want %d, err=%v", c.se, c.gain, units, c.want, err)
		}
	}
	// Exact rational arithmetic checks this regression's real-valued formula
	// for the supplied float inputs. It is not a certification of the helper.
	z := new(big.Rat).Add(new(big.Rat).SetFloat64(3.5), new(big.Rat).SetFloat64(.84162123357))
	exactUnits := new(big.Rat).Mul(new(big.Rat).Mul(z, z), big.NewRat(32, 1))
	ceiling := new(big.Int).Quo(exactUnits.Num(), exactUnits.Denom())
	if new(big.Int).Mod(exactUnits.Num(), exactUnits.Denom()).Sign() != 0 {
		ceiling.Add(ceiling, big.NewInt(1))
	}
	if ceiling.Uint64() != 604 {
		t.Fatalf("exact regression reference changed: %s", ceiling)
	}
	actualUnits, err := FixedSampleNormalPlan(32, math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64, 3.5, .84162123357)
	if err != nil {
		t.Fatal(err)
	}
	logJSONRecord(t, "planning_numeric_regression", map[string]any{
		"scope":       "floating-point utility regression, not scientific confirmation",
		"pilot_units": 32, "equal_se_and_gain": "SmallestNonzeroFloat64",
		"critical_z": 3.5, "power_z": .84162123357, "expected_units": 604, "actual_units": actualUnits,
		"formal_outward_rounded_ceiling_claim": false,
	})
}

func TestNormalPlanNumericDomainControls(t *testing.T) {
	for _, c := range []struct {
		name string
		args [4]float64
	}{
		{"relative ratio overflow", [4]float64{.01, math.SmallestNonzeroFloat64, 3.5, .84}},
		{"relative ratio underflow", [4]float64{math.SmallestNonzeroFloat64, 2, 3.5, .84}},
		{"subnormal relative ratio", [4]float64{math.SmallestNonzeroFloat64, 1, math.MaxFloat64, 0}},
		{"critical sum overflow", [4]float64{.01, .01, math.MaxFloat64, math.MaxFloat64}},
		{"scaled ratio overflow", [4]float64{.1, 0x1p-1022, 1000, 0}},
		{"scaled ratio underflow", [4]float64{.01, 1, math.SmallestNonzeroFloat64, 0}},
		{"subnormal scaled ratio", [4]float64{.01, .01, math.SmallestNonzeroFloat64, 0}},
		{"finite count outside range", [4]float64{.01, .01, 1e10, 0}},
		{"count overflow", [4]float64{.01, .01, 1e200, 0}},
	} {
		t.Run(c.name, func(t *testing.T) {
			units, err := FixedSampleNormalPlan(32, c.args[0], c.args[1], c.args[2], c.args[3])
			if err == nil || units != 0 {
				t.Fatalf("unsupported numeric regime returned a plan: %d %v", units, err)
			}
		})
	}
	// The result is known to be below the floor without underflowing its square.
	if units, err := FixedSampleNormalPlan(32, .01, 1, 1e-200, 0); err != nil || units != 2 {
		t.Fatalf("safe tiny-plan floor: %d %v", units, err)
	}
	zAtCap := float64(uint64(1) << 24)
	if units, err := FixedSampleNormalPlan(32, .01, .01, zAtCap, 0); err != nil || units != maxUnits {
		t.Fatalf("exact count cap rejected: %d %v", units, err)
	}
	if units, err := FixedSampleNormalPlan(32, .01, .01, math.Nextafter(zAtCap, 0), 0); err != nil || units > maxUnits {
		t.Fatalf("below-cap control rejected: %d %v", units, err)
	}
	if units, err := FixedSampleNormalPlan(32, .01, .01, math.Nextafter(zAtCap, math.Inf(1)), 0); err == nil || units != 0 {
		t.Fatalf("above-cap control returned a plan: %d %v", units, err)
	}
}

func TestFixedSampleNormalPlan(t *testing.T) {
	// Archived summary numbers are planning inputs only, not a replay or test
	// of the old verdict. SE is recovered from its frozen z=3.5 half-width.
	effect := .00859583394601159
	se := (.01485294340663565 - .0023387244853875296) / 7
	for _, c := range []struct {
		gain float64
		want uint64
	}{
		{effect, 27},
		{effect / 2, 105},
		{effect / 4, 418},
	} {
		got, err := FixedSampleNormalPlan(32, se, c.gain, 3.5, .8416212335729143)
		if err != nil || got != c.want {
			t.Fatalf("gain %.9g: units=%d want=%d err=%v", c.gain, got, c.want, err)
		}
	}
	got, err := FixedSampleNormalPlan(32, se, effect, 3.5, 1.2815515655446004)
	if err != nil || got != 32 {
		t.Fatalf("90%% directional power planning: %d %v", got, err)
	}
	got, err = FixedSampleNormalPlan(32, se, 1, 1, 0)
	if err != nil || got != 2 {
		t.Fatalf("minimum unit count: %d %v", got, err)
	}
}

func TestAPMemberShiftNormalPlan(t *testing.T) {
	// Exact consumed summary contrast: confirmation/member_shift/post,
	// mix_mmm_no_ap minus mix_mmm_ap. This is unresolved pilot evidence.
	mean := .0005635187818819783
	lower, upper := -.0003752428971604797, .0015022804609244361
	se := (upper - lower) / 7
	units, err := FixedSampleNormalPlan(32, se, mean, 3.5, .8416212335729143)
	if err != nil || units != 137 || lower >= 0 || upper <= 0 {
		t.Fatalf("AP pilot plan or unresolved interval changed: %d %v", units, err)
	}
	roundedSE := (.001502 - (-.000375)) / 7
	roundedUnits, err := FixedSampleNormalPlan(32, roundedSE, .000564, 3.5, .8416212335729143)
	if err != nil || roundedUnits != 137 {
		t.Fatalf("rounded AP plan: %d %v", roundedUnits, err)
	}
	logJSONRecord(t, "ap_member_planning", map[string]any{
		"scope": "approximate pilot planning only, not scientific confirmation or CS power",
		"split": "confirmation", "scenario": "member_shift", "window": "post",
		"control": "mix_mmm_no_ap", "candidate": "mix_mmm_ap",
		"pilot_units": 32, "mean": mean, "lower": lower, "upper": upper,
		"observed_se": se, "critical_z": 3.5, "power_z": .8416212335729143,
		"directional_power": .8, "approximate_total_fresh_units": units,
		"rounded_input_units": roundedUnits, "old_interval_excludes_zero": false,
	})
}

func TestRecurringAPIsNotPositivePlanningTarget(t *testing.T) {
	mean := -.0003482400977983611
	lower, upper := -.0018459539459258038, .0011494737503290817
	se := (upper - lower) / 7
	_, err := FixedSampleNormalPlan(32, se, mean, 3.5, .8416212335729143)
	if err == nil || lower >= 0 || upper <= 0 {
		t.Fatal("negative recurring AP contrast treated as a positive-benefit target")
	}
	logJSONRecord(t, "ap_recurring_context", map[string]any{
		"scope": "consumed pilot context only, not scientific confirmation",
		"split": "confirmation", "scenario": "recurring", "window": "post",
		"control": "mix_mmm_no_ap", "candidate": "mix_mmm_ap",
		"mean": mean, "lower": lower, "upper": upper,
		"positive_benefit_planning_target": false, "old_interval_excludes_zero": false,
	})
}

func TestInvalidNormalPlan(t *testing.T) {
	for _, n := range []uint64{0, 1, maxUnits + 1} {
		if _, err := FixedSampleNormalPlan(n, .01, .01, 3.5, .84); err == nil {
			t.Errorf("accepted pilot count %d", n)
		}
	}
	for i := 0; i < 4; i++ {
		for _, bad := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), -1} {
			args := [4]float64{.01, .01, 3.5, .84}
			args[i] = bad
			if _, err := FixedSampleNormalPlan(32, args[0], args[1], args[2], args[3]); err == nil {
				t.Errorf("accepted bad argument %d = %g", i, bad)
			}
		}
	}
	for _, args := range [][4]float64{
		{0, .01, 3.5, .84}, {.01, 0, 3.5, .84}, {.01, .01, 0, .84},
		{1, .01, 3.5, .84}, {.01, 3, 3.5, .84},
		{.01, math.SmallestNonzeroFloat64, 3.5, .84},
		{.01, .01, math.MaxFloat64, math.MaxFloat64},
	} {
		if _, err := FixedSampleNormalPlan(32, args[0], args[1], args[2], args[3]); err == nil {
			t.Errorf("accepted invalid or overflowing plan: %v", args)
		}
	}
}
