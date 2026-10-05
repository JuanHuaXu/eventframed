package researchstats

import (
	"errors"
	"math"
)

// FixedSampleNormalPlan is ONLY an approximate fixed-sample planning calculation:
// ceil(pilotUnits*((criticalZ+powerZ)*observedSE/targetGain)^2), at least 2 units.
// It assumes stable pilot variance, independent identical units and a normal
// approximation. Supply a prespecified multiplicity-adjusted criticalZ and the
// standard-normal power quantile (e.g. .841621 for 80% directional power).
// targetGain is the positive distance above the null/benefit threshold, not
// necessarily the raw effect. Units must match the pilot SE's independence unit.
// This is not CS power, guaranteed power, fitted-model generalization, or a
// license to use the consumed pilot as fresh confirmation. See the contract.
// Float64 evaluation is not a formally outward-rounded ceiling of the
// real-valued formula. Unsupported numeric intermediates are rejected.
func FixedSampleNormalPlan(pilotUnits uint64, observedSE, targetGain, criticalZ, powerZ float64) (uint64, error) {
	if pilotUnits < 2 || pilotUnits > maxUnits || !finite(observedSE) || observedSE <= 0 ||
		observedSE > 1/math.Sqrt(float64(pilotUnits-1)) ||
		!finite(targetGain) || targetGain <= 0 || targetGain > 2 ||
		!finite(criticalZ) || criticalZ <= 0 || !finite(powerZ) || powerZ < 0 {
		return 0, errors.New("invalid bounded paired-mean normal planning inputs")
	}
	// Cancel matching scales before multiplication, including equal subnormals.
	relativeSE := observedSE / targetGain
	z := criticalZ + powerZ
	const minNormal = 0x1p-1022
	if !finite(relativeSE) || relativeSE < minNormal || !finite(z) {
		return 0, errors.New("unsupported normal planning ratio or critical sum")
	}
	ratio := relativeSE * z
	if !finite(ratio) || ratio < minNormal {
		return 0, errors.New("unsupported scaled normal planning ratio")
	}
	// This region is safely below the two-unit floor; do not square a tiny ratio.
	if ratio <= 1/math.Sqrt(float64(pilotUnits)) {
		return 2, nil
	}
	units := float64(pilotUnits) * ratio * ratio
	if !finite(units) || units > float64(maxUnits) {
		return 0, errors.New("normal plan exceeds supported count range")
	}
	return uint64(math.Max(2, math.Ceil(units))), nil
}
