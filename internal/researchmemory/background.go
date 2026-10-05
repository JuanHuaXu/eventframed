package researchmemory

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Background is an opt-in research learner, not a serving policy. One worker
// owns fitting; predictions use immutable snapshots and retain their ORIGINAL
// expert forecasts for feedback. No text, self-labeling or publication authority
// is provided here. A new epoch requires a new worker.
type Background struct {
	mu                sync.Mutex
	adapter           *Adapter
	current           atomic.Pointer[Frozen]
	pending           map[uint64]pending
	next              uint64
	lastAvailable     time.Time
	jobs              chan backgroundLabel
	stop, done        chan struct{}
	closed            bool
	completed, failed uint64
	changed           chan struct{}
}

type backgroundLabel struct {
	record    pending
	useful    bool
	available time.Time
}

func NewBackground(epoch uint64, seed int64, capacity int) (*Background, error) {
	return backgroundFromOwnedAdapter(New(epoch, seed), capacity)
}

// Ownership transfers here before the worker starts. Restored unresolved records
// belong to the background journal, not simultaneously to the fitting adapter.
func backgroundFromOwnedAdapter(a *Adapter, capacity int) (*Background, error) {
	if capacity < 1 || capacity > 256 {
		return nil, errors.New("invalid background capacity")
	}
	b := &Background{adapter: a, pending: a.pending, next: a.next,
		lastAvailable: a.lastFeedback, completed: uint64(a.total),
		jobs: make(chan backgroundLabel, capacity), stop: make(chan struct{}), done: make(chan struct{})}
	a.pending = make(map[uint64]pending)
	f := b.adapter.Freeze()
	b.current.Store(&f)
	go b.run()
	return b, nil
}

// Predict journals against one snapshot. Holding this short mutex orders IDs
// and feedback admission, but it is never held during model fitting.
func (b *Background) Predict(features uint16, baseline float64, epoch uint64, at time.Time) (Prediction, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed || len(b.pending) >= 256 || b.next == ^uint64(0) {
		return Prediction{}, errors.New("closed or saturated background learner")
	}
	f := b.current.Load()
	p, err := f.Score(features, baseline, epoch, at)
	if err != nil {
		return Prediction{}, err
	}
	sp, lp, tp := .5, .5, .5
	ready := f.short != nil
	if ready {
		sp, _ = f.short.ForecastObserved(511, features)
		lp, _ = f.long.ForecastObserved(511, features)
		tp = f.forest.Predict(features)
	}
	inner := [4]float64{sp, tp, tp, tp}
	outer := [4]float64{baseline, f.inner.Forecast(inner), lp, .5}
	b.next++
	r := Prediction{ID: b.next, Probability: p, Features: features, Epoch: epoch}
	b.pending[r.ID] = pending{Prediction: r, At: at, Outer: outer, Inner: inner, Ready: ready}
	return r, nil
}

// Feedback admits a verified label without waiting for fitting. A full queue
// leaves the journal intact for retry; accepted feedback cannot be submitted
// twice. Availability order is explicit, independent of processing wall time.
func (b *Background) Feedback(id uint64, useful bool, epoch uint64, available time.Time) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	r, ok := b.pending[id]
	if b.closed || !ok || epoch != r.Epoch || available.IsZero() || available.Before(r.At) || available.Before(b.lastAvailable) {
		return errors.New("invalid background feedback")
	}
	select {
	case b.jobs <- backgroundLabel{r, useful, available}:
		delete(b.pending, id)
		b.lastAvailable = available
		return nil
	default:
		return errors.New("background queue full")
	}
}

func (b *Background) run() {
	defer close(b.done)
	for {
		select {
		case <-b.stop:
			return
		case job := <-b.jobs:
			b.mu.Lock()
			closed := b.closed
			b.mu.Unlock()
			if closed {
				return
			}
			// Preserve the forecast made before the outcome. Recomputing Predict
			// here would leak intervening feedback into mixture weight updates.
			b.adapter.mu.Lock()
			b.adapter.pending[job.record.ID] = job.record
			b.adapter.mu.Unlock()
			err := b.adapter.Feedback(job.record.ID, job.useful, job.record.Epoch, job.available)
			if err != nil {
				b.adapter.Discard(job.record.ID)
			}
			f := b.adapter.Freeze()
			b.mu.Lock()
			if err != nil {
				b.failed++
			} else {
				b.completed++
				if !b.closed {
					b.current.Store(&f)
				}
			}
			b.notifyLocked()
			b.mu.Unlock()
		}
	}
}

func (b *Background) Snapshot() Frozen { return *b.current.Load() }

// Discard releases an unlabelled journal without inventing evidence. It cannot
// retract a label that Feedback already accepted into the worker queue.
func (b *Background) Discard(id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.pending, id)
}

func (b *Background) Counts() (completed, failed uint64, pending, queued int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.completed, b.failed, len(b.pending), len(b.jobs)
}

// Close waits for at most the current bounded fit; queued/unlabeled work is
// abandoned, never converted to negative evidence. Existing snapshots survive.
func (b *Background) Close() {
	b.mu.Lock()
	if !b.closed {
		b.closed = true
		b.notifyLocked()
		close(b.stop)
		clear(b.pending)
	}
	b.mu.Unlock()
	<-b.done
	for {
		select {
		case <-b.jobs:
		default:
			return
		}
	}
}
