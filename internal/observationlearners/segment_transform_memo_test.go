package observationlearners

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

// Exact work identity, not an abstraction key. History includes unavailable
// packets conservatively; equal keys cannot hide changed evidence or timing.
type transformWork struct {
	Tenant                   string
	Epoch                    uint64
	Left, Clock, Cap, Length int
	Hazard, Mass             float64
	History                  [272]segmentPacket
}

type transformMemo struct {
	token       chan struct{}
	valid       bool
	key         transformWork
	predictions [512]float64
	fits, hits  atomic.Uint64
}

func newTransformMemo() *transformMemo {
	m := &transformMemo{token: make(chan struct{}, 1)}
	m.token <- struct{}{}
	return m
}

func (m *transformMemo) score(ctx context.Context, key transformWork, x int) (float64, error) {
	if key.Tenant == "" || key.Length < 0 || key.Length > 272 || x < 0 || x >= 512 {
		return 0, errors.New("invalid research work")
	}
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-m.token:
	}
	defer func() { m.token <- struct{}{} }()
	if e := ctx.Err(); e != nil {
		return 0, e
	}
	if m.valid && m.key == key {
		m.hits.Add(1)
		return m.predictions[x], nil
	}
	fitted, e := fitSegmentPosteriorContext(ctx, key.Left, key.Clock, key.History[:key.Length], key.Cap, key.Hazard, key.Mass)
	if e != nil {
		return 0, e
	}
	if e := ctx.Err(); e != nil {
		return 0, e
	}
	m.key, m.predictions, m.valid = key, fitted.predictions, true
	m.fits.Add(1)
	return m.predictions[x], nil
}

func fixtureWork() transformWork {
	k := transformWork{Tenant: "research-fixture", Epoch: 1, Left: -16, Clock: 256, Cap: 64, Length: 272, Hazard: .01, Mass: .95}
	copy(k.History[:], transformHistory(272))
	return k
}

// Only exported in the test binary. Query index selects from the exact fitted
// predictive vector; it is not evidence and does not require another fit.
func ResearchMemoFixture() func(context.Context) (float64, error) {
	m := newTransformMemo()
	k := fixtureWork()
	return func(ctx context.Context) (float64, error) { return m.score(ctx, k, 17) }
}

func TestTransformMemoDependencies(t *testing.T) {
	m := newTransformMemo()
	k := fixtureWork()
	ctx := context.Background()
	a, e := m.score(ctx, k, 17)
	if e != nil {
		t.Fatal(e)
	}
	b, e := m.score(ctx, k, 17)
	if e != nil || a != b || m.fits.Load() != 1 || m.hits.Load() != 1 {
		t.Fatal("exact reuse", e)
	}
	changes := []func(*transformWork){
		func(k *transformWork) { k.Tenant = "other" }, func(k *transformWork) { k.Epoch++ },
		func(k *transformWork) { k.History[250].Outcome = !k.History[250].Outcome },
		func(k *transformWork) { k.History[250].Bits ^= 1 }, func(k *transformWork) { k.History[250].Arrives++ },
		func(k *transformWork) { k.Cap = 16 }, func(k *transformWork) { k.Hazard = .02 }, func(k *transformWork) { k.Mass = .5 },
		func(k *transformWork) { k.Clock--; k.Length-- },
	}
	for i, change := range changes {
		next := k
		change(&next)
		before := m.fits.Load()
		if _, e := m.score(ctx, next, 17); e != nil || m.fits.Load() != before+1 {
			t.Fatal("dependency collision", i, e)
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	before := m.fits.Load()
	if _, e := m.score(cancelled, k, 17); !errors.Is(e, context.Canceled) || m.fits.Load() != before {
		t.Fatal("cancelled work admitted")
	}
	// A blocked waiter honors its deadline without changing the cached result.
	<-m.token
	if _, e := m.score(cancelled, k, 17); !errors.Is(e, context.Canceled) {
		t.Fatal("wait cancellation")
	}
	m.token <- struct{}{}
	if _, e := m.score(ctx, k, -1); e == nil {
		t.Fatal("invalid query")
	}
}
