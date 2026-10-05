package researchswitch

import (
	"math"

	"github.com/JuanHuaXu/eventframed/internal/researchbrier"
)

func validScoreModeV51(mode string) bool {
	return mode == "log_mean" || mode == "log_strong" || mode == "brier_mean" || mode == "brier_strong"
}

// p is the originally issued probability of the REVEALED outcome, not a
// recomputed prediction. Brier weights are strategy-loss weights, not Bayes.
func scoreEmissionV51(mode string, p float64) float64 {
	if mode == "brier_mean" || mode == "brier_strong" {
		d := 1 - p
		return -2 * d * d
	}
	return math.Log(p)
}

func (m *scoreModelV51) scoredForecastV51(advice []float64, arithmetic float64) (float64, error) {
	if m.mode == "log_strong" || m.mode == "brier_strong" {
		logs := m.nextLogs()
		return researchbrier.Forecast(logs[:m.n], advice)
	}
	return arithmetic, nil
}
