// Package researchbeta implements a research-only Beta-Bernoulli predictive
// model. Runtime provenance, selection and epoch validity belong to the caller.
package researchbeta

import (
	"errors"
	"math"
)

const maxExactCount uint64 = 1 << 53

// Predict uses one joint model: theta has the declared Beta prior, and past
// observations and the future outcome are independent Bernoulli(theta) draws
// conditional on theta. Counts must be ordinary, untempered observations.
func Predict(priorMean, priorStrength float64, successes, failures uint64) (float64, error) {
	if math.IsNaN(priorMean) || math.IsInf(priorMean, 0) || priorMean <= 0 || priorMean >= 1 ||
		math.IsNaN(priorStrength) || math.IsInf(priorStrength, 0) || priorStrength <= 0 ||
		successes > maxExactCount || failures > maxExactCount || successes+failures > maxExactCount {
		return 0, errors.New("invalid proper prior or inexact observation counts")
	}
	alpha := priorMean * priorStrength
	beta := (1 - priorMean) * priorStrength
	if alpha <= 0 || beta <= 0 || math.IsInf(alpha+beta, 0) {
		return 0, errors.New("prior parameters are not representable")
	}
	return (alpha + float64(successes)) / (alpha + beta + float64(successes+failures)), nil
}
