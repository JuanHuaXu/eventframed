// Package researchretentionlaw binds clean and measurement predictions to one
// declared joint law. It does not authenticate that law against the real world.
package researchretentionlaw

import (
	"errors"
	retention "github.com/JuanHuaXu/eventframed/internal/researchretention"
	"math"
)

// Joint indexes 4*Y+2*W1+W2. W1 and W2 concern the same latent outcome.
type Joint [8]float64

func Forecasts(joints [retention.Experts]Joint) (retention.Forecasts, error) {
	var result retention.Forecasts
	for k, joint := range joints {
		total := 0.
		for _, mass := range joint {
			if math.IsNaN(mass) || math.IsInf(mass, 0) || mass < 0 || mass > 1 {
				return retention.Forecasts{}, errors.New("retention invalid latent joint")
			}
			total += mass
		}
		if math.Abs(total-1) > 2e-12 {
			return retention.Forecasts{}, errors.New("retention latent joint normalization")
		}
		for z, mass := range joint {
			p := mass / total
			result[k].Joint[z%4] += p
			if z >= 4 {
				result[k].Clean += p
			}
		}
	}
	return result, nil
}
