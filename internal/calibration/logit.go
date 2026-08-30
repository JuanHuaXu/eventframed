package calibration

import (
	"errors"
	"math"
)

type Logit struct {
	Scale float64 `json:"scale"`
	Bias  float64 `json:"bias"`
	Floor float64 `json:"floor"`
}

type Observation struct {
	Probability float64
	Outcome     float64
	Weight      float64
}

func Identity() Logit {
	return Logit{Scale: 1, Floor: 1e-6}
}

// Compose returns a single map equivalent to applying inner and then outer,
// apart from probability-floor clipping at the extreme tails.
func Compose(outer, inner Logit) Logit {
	if !outer.Valid() {
		outer = Identity()
	}
	if !inner.Valid() {
		inner = Identity()
	}
	return Logit{Scale: outer.Scale * inner.Scale, Bias: outer.Scale*inner.Bias + outer.Bias, Floor: math.Max(outer.Floor, inner.Floor)}
}

func (calibrator Logit) Valid() bool {
	return calibrator.Scale > 0 && finite(calibrator.Scale) && finite(calibrator.Bias) && calibrator.Floor > 0 && calibrator.Floor < .5 && finite(calibrator.Floor)
}

func (calibrator Logit) Apply(probability float64) float64 {
	if !calibrator.Valid() {
		calibrator = Identity()
	}
	probability = clamp(probability, calibrator.Floor, 1-calibrator.Floor)
	linear := calibrator.Scale*math.Log(probability/(1-probability)) + calibrator.Bias
	if linear >= 0 {
		decay := math.Exp(-linear)
		return 1 / (1 + decay)
	}
	growth := math.Exp(linear)
	return growth / (1 + growth)
}

// Inverse maps an emitted calibrated probability back to the score domain.
// Values beyond floating-point resolution recover to the declared input floor.
func (calibrator Logit) Inverse(probability float64) float64 {
	if !calibrator.Valid() {
		calibrator = Identity()
	}
	probability = clamp(probability, 1e-300, 1-1e-15)
	linear := (math.Log(probability/(1-probability)) - calibrator.Bias) / calibrator.Scale
	if linear >= 0 {
		decay := math.Exp(-linear)
		return clamp(1/(1+decay), calibrator.Floor, 1-calibrator.Floor)
	}
	growth := math.Exp(linear)
	return clamp(growth/(1+growth), calibrator.Floor, 1-calibrator.Floor)
}

func Fit(observations []Observation) (Logit, error) {
	if len(observations) == 0 {
		return Logit{}, errors.New("calibration observations are empty")
	}
	calibrator := Identity()
	const ridge = 1e-6
	for range 100 {
		var gradientScale, gradientBias float64
		var hScaleScale, hScaleBias, hBiasBias float64
		for _, observation := range observations {
			if observation.Outcome < 0 || observation.Outcome > 1 || !finite(observation.Outcome) || observation.Weight <= 0 || !finite(observation.Weight) {
				return Logit{}, errors.New("calibration observations must have binary outcomes and positive finite weights")
			}
			probability := clamp(observation.Probability, calibrator.Floor, 1-calibrator.Floor)
			x := math.Log(probability / (1 - probability))
			predicted := calibrator.Apply(observation.Probability)
			errorValue := observation.Weight * (predicted - observation.Outcome)
			curvature := observation.Weight * predicted * (1 - predicted)
			gradientScale += errorValue * x
			gradientBias += errorValue
			hScaleScale += curvature * x * x
			hScaleBias += curvature * x
			hBiasBias += curvature
		}
		gradientScale += ridge * calibrator.Scale
		gradientBias += ridge * calibrator.Bias
		hScaleScale += ridge
		hBiasBias += ridge
		determinant := hScaleScale*hBiasBias - hScaleBias*hScaleBias
		if determinant <= 0 || !finite(determinant) {
			return Logit{}, errors.New("calibration Hessian is singular")
		}
		deltaScale := (hBiasBias*gradientScale - hScaleBias*gradientBias) / determinant
		deltaBias := (-hScaleBias*gradientScale + hScaleScale*gradientBias) / determinant
		before := logisticLoss(observations, calibrator, ridge)
		accepted := false
		step := 1.0
		var next Logit
		for range 40 {
			next = calibrator
			next.Scale = clamp(calibrator.Scale-step*deltaScale, .01, 20)
			next.Bias = clamp(calibrator.Bias-step*deltaBias, -20, 20)
			if logisticLoss(observations, next, ridge) < before {
				accepted = true
				break
			}
			step *= .5
		}
		if !accepted {
			break
		}
		movement := math.Abs(next.Scale-calibrator.Scale) + math.Abs(next.Bias-calibrator.Bias)
		calibrator = next
		if movement < 1e-9 {
			break
		}
	}
	return calibrator, nil
}

func logisticLoss(observations []Observation, calibrator Logit, ridge float64) float64 {
	loss := .5 * ridge * (calibrator.Scale*calibrator.Scale + calibrator.Bias*calibrator.Bias)
	for _, observation := range observations {
		predicted := clamp(calibrator.Apply(observation.Probability), calibrator.Floor, 1-calibrator.Floor)
		loss -= observation.Weight * (observation.Outcome*math.Log(predicted) + (1-observation.Outcome)*math.Log(1-predicted))
	}
	return loss
}

func Brier(observations []Observation, calibrator Logit) float64 {
	var total, weight float64
	for _, observation := range observations {
		errorValue := calibrator.Apply(observation.Probability) - observation.Outcome
		total += observation.Weight * errorValue * errorValue
		weight += observation.Weight
	}
	if weight == 0 {
		return 0
	}
	return total / weight
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
func clamp(value, lower, upper float64) float64 {
	return math.Min(math.Max(value, lower), upper)
}
