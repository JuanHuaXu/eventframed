package researchmemory

import (
	"errors"
	"math"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

// Frozen is a read-only research model, with no feedback or journal side effects.
// Adapter refits replace models instead of mutating published instances. Keep
// that invariant if fitting changes: readers may outlive the next refit.
type Frozen struct {
	epoch        uint64
	asOf         time.Time
	short, long  *observation.Model
	forest       *observationlearners.Forest
	outer, inner bayes.ForecastMix
}

func (a *Adapter) Freeze() Frozen {
	a.mu.Lock()
	defer a.mu.Unlock()
	return Frozen{a.epoch, a.lastFeedback, a.short, a.long, a.forest, a.outer, a.inner}
}

// Score refuses use before the model's latest admitted feedback, or across an
// epoch change. It returns a research probability, not a calibrated rank score.
func (f Frozen) Score(features uint16, baseline float64, epoch uint64, at time.Time) (float64, error) {
	if features >= 512 || epoch != f.epoch || at.IsZero() || at.Before(f.asOf) || math.IsNaN(baseline) || math.IsInf(baseline, 0) || baseline < 0 || baseline > 1 {
		return 0, errors.New("invalid or stale frozen prediction")
	}
	if f.short == nil {
		return baseline, nil
	}
	sp, err := f.short.ForecastObserved(511, features)
	if err != nil {
		return 0, err
	}
	lp, err := f.long.ForecastObserved(511, features)
	if err != nil {
		return 0, err
	}
	tp := f.forest.Predict(features)
	return f.outer.Forecast([4]float64{baseline, f.inner.Forecast([4]float64{sp, tp, tp, tp}), lp, .5}), nil
}
