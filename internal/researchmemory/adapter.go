package researchmemory

import (
	"errors"
	"math"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

type Prediction struct {
	ID          uint64
	Probability float64
	Features    uint16
	Epoch       uint64
}
type pending struct {
	Prediction
	At           time.Time
	Outer, Inner [4]float64
	Ready        bool
}

// Adapter serializes a bounded offline/slow-path learner. Predict and Feedback
// are NOT a production hot-path API: fitting currently occurs synchronously.
type Adapter struct {
	mu           sync.Mutex
	epoch, next  uint64
	seed         int64
	pending      map[uint64]pending
	samples      []observation.Sample
	total        int
	short, long  *observation.Model
	forest       *observationlearners.Forest
	outer, inner bayes.ForecastMix
	lastFeedback time.Time
}

func New(epoch uint64, seed int64) *Adapter {
	return &Adapter{epoch: epoch, seed: seed, pending: map[uint64]pending{}}
}

func (a *Adapter) Predict(features uint16, baseline float64, epoch uint64, at time.Time) (Prediction, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if features >= 512 || epoch != a.epoch || at.IsZero() || at.Before(a.lastFeedback) || math.IsNaN(baseline) || math.IsInf(baseline, 0) || baseline < 0 || baseline > 1 || len(a.pending) >= 256 || a.next == ^uint64(0) {
		return Prediction{}, errors.New("invalid or saturated research prediction")
	}
	sp, lp, tp := .5, .5, .5
	ready := a.short != nil
	if ready {
		sp, _ = a.short.ForecastObserved(511, features)
		lp, _ = a.long.ForecastObserved(511, features)
		tp = a.forest.Predict(features)
	}
	i := [4]float64{sp, tp, tp, tp}
	o := [4]float64{baseline, a.inner.Forecast(i), lp, .5}
	p := baseline
	if ready {
		p = a.outer.Forecast(o)
	}
	a.next++
	r := Prediction{ID: a.next, Probability: p, Features: features, Epoch: epoch}
	a.pending[r.ID] = pending{Prediction: r, At: at, Outer: o, Inner: i, Ready: ready}
	return r, nil
}

// Feedback must be explicit externally verified usefulness, never inferred from
// a high rank or the model's own output. Repeated, early and stale labels reject
// without consuming the original record. Epoch changes require a fresh adapter.
func (a *Adapter) Feedback(id uint64, useful bool, epoch uint64, available time.Time) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	r, ok := a.pending[id]
	if !ok || epoch != a.epoch || available.IsZero() || available.Before(r.At) || available.Before(a.lastFeedback) {
		return errors.New("unknown, early, out-of-order or stale feedback")
	}
	// Precompute replacement fits before mutating state, keeping failures atomic.
	samples := append(append([]observation.Sample(nil), a.samples...), observation.Sample{Bits: r.Features, Outcome: useful})
	if len(samples) > 256 {
		samples = samples[len(samples)-256:]
	}
	short, long, forest := a.short, a.long, a.forest
	if a.total+1 >= 32 && (a.total+1)%16 == 0 {
		var e error
		short, e = observation.Fit(samples[max(0, len(samples)-64):])
		if e != nil {
			return e
		}
		long, e = observation.Fit(samples)
		if e != nil {
			return e
		}
		forest = observationlearners.NewForest(a.seed)
		for _, v := range samples[max(0, len(samples)-64):] {
			forest.Update(v.Bits, v.Outcome)
		}
	}
	if r.Ready {
		a.outer = a.outer.Observe(r.Outer, useful, 1)
		a.inner = a.inner.Observe(r.Inner, useful, 1)
	}
	a.short, a.long, a.forest = short, long, forest
	a.samples = samples
	a.total++
	a.lastFeedback = available
	delete(a.pending, id)
	return nil
}

// Discard abandons an unlabeled record without turning absence into a negative.
func (a *Adapter) Discard(id uint64) { a.mu.Lock(); defer a.mu.Unlock(); delete(a.pending, id) }
func (a *Adapter) Counts() (labels, pending int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.total, len(a.pending)
}
