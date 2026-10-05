// Package researchbounds implements conditional error budgets, not evidence
// authentication or a proof that a caller's supplied bounds are valid.
package researchbounds

import (
	"errors"
	"math"
)

func unit(x float64) bool { return !math.IsNaN(x) && x >= 0 && x <= 1 }
func up(x float64) float64 {
	// A computed zero can be a positive underflow. Exact zero cases are
	// handled at the call site rather than conflated with rounded products.
	return math.Nextafter(x, math.Inf(1))
}
func down(x float64) float64 { return math.Nextafter(x, math.Inf(-1)) }

// StepTV transports a previous TV bound through a reset transition, Bayesian
// conditioning, and truncation. Inputs must be independently justified bounds:
// likelihood is in [minimum, maximum], approximate evidence is >= evidenceLower,
// and discardedMass bounds mass removed from the approximate updated law.
// Directed arithmetic protects this calculation, not the upstream inference.
func StepTV(previous, reset, minimum, maximum, evidenceLower, discardedMass float64) (float64, error) {
	if !unit(previous) || !unit(reset) || !unit(minimum) || !unit(maximum) || maximum == 0 || minimum > maximum || !unit(evidenceLower) || evidenceLower > maximum || !unit(discardedMass) {
		return 0, errors.New("invalid TV bound inputs")
	}
	prior := math.Min(1, up(up(1-reset)*previous))
	if previous == 0 || reset == 1 {
		prior = 0
	}
	conditioned := prior
	if maximum != minimum && prior > 0 {
		// The exact evidence is at least Z_hat - oscillation(g)*prior.
		denominator := math.Max(minimum, down(evidenceLower-up(up(maximum-minimum)*prior)))
		if denominator <= 0 {
			conditioned = 1
		} else {
			conditioned = math.Min(1, up(up(maximum*prior)/denominator))
		}
	}
	if conditioned == 0 && discardedMass == 0 {
		return 0, nil
	}
	return math.Min(1, up(conditioned+discardedMass)), nil
}

// BinaryBrierShift bounds change in the SAME outcome's single-coordinate Brier
// loss after a common Bernoulli predictive map. It is not a calibration bound,
// nor a bound on log loss, or on two different posterior-dependent decoders.
func BinaryBrierShift(tv float64) (float64, error) {
	if !unit(tv) {
		return 0, errors.New("invalid TV bound")
	}
	if tv == 0 {
		return 0, nil
	}
	return math.Min(1, up(2*tv)), nil
}

// TotalTV uses the triangle inequality; numeric must cover inference error
// propagated to this law, not merely local arithmetic or observed audit parity.
func TotalTV(model, numeric float64) (float64, error) {
	if !unit(model) || !unit(numeric) {
		return 0, errors.New("invalid composed TV bound")
	}
	if model == 0 && numeric == 0 {
		return 0, nil
	}
	return math.Min(1, up(model+numeric)), nil
}

// SafeDiscard is the sufficient uniform per-step truncation budget when
// a=(1-reset)*maximum/minimum < 1 and error must stay <= tolerance. A cap
// count alone supplies no discarded-mass guarantee. No contractive region is
// claimed for zero minimum likelihood or a>=1.
func SafeDiscard(reset, minimum, maximum, tolerance float64) (float64, error) {
	if !unit(reset) || !unit(minimum) || !unit(maximum) || minimum > maximum || minimum == 0 || !unit(tolerance) {
		return 0, errors.New("invalid safe-region inputs")
	}
	a := up(up(up(1-reset)*maximum) / minimum)
	if reset == 1 {
		a = 0
	}
	if a >= 1 {
		return 0, errors.New("no uniform contractive region")
	}
	return math.Max(0, down(down(1-a)*tolerance)), nil
}
