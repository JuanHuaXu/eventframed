package observationgate

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

const subsetDelayCapacity = 64

type subsetDelayEntry struct {
	active, ready, censored, y, authorized bool
	origin, generation                     int
	forecast                               subsetPending
}

// Research-only, single-owner journal. Models are immutable after publication.
// Conservative generation scoping deliberately discards stale selector losses;
// it makes no claim that discarding them is the best learning policy.
type subsetDelayJournal struct {
	state                    subsetState
	models                   observationpreserved.Models
	model                    *observationlearners.ConditionalForest
	generation, head         int
	entries                  [subsetDelayCapacity]subsetDelayEntry
	applied, stale, censored int
}

func (j *subsetDelayJournal) publish(models observationpreserved.Models, model *observationlearners.ConditionalForest) error {
	if j.state.base == nil || j.state.pending != nil || (j.generation > 0 && models.Version <= j.models.Version) {
		return fmt.Errorf("invalid delayed publication")
	}
	j.models, j.model = models, model
	j.generation++
	return nil
}

func (j *subsetDelayJournal) predict(origin int, reader observation.Reader, seed int64) (observationpreserved.Prediction, error) {
	if origin != j.state.next || origin < 0 || j.generation == 0 || j.entries[origin%subsetDelayCapacity].active {
		return observationpreserved.Prediction{}, fmt.Errorf("invalid delayed issuance or capacity")
	}
	next := j.state
	p, err := next.predict(reader, j.models, j.model, origin, seed)
	if err != nil {
		return observationpreserved.Prediction{}, err
	}
	j.entries[origin%subsetDelayCapacity] = subsetDelayEntry{active: true, origin: origin, generation: j.generation, forecast: *next.pending}
	next.pending = nil
	next.next++
	j.state = next
	return p, nil
}

func (j *subsetDelayJournal) deliver(origin int, y, authorized bool) error {
	if origin < j.head || origin >= j.state.next {
		return fmt.Errorf("unknown delayed origin")
	}
	i := origin % subsetDelayCapacity
	e := j.entries[i]
	if !e.active || e.origin != origin || e.ready || e.censored {
		return fmt.Errorf("duplicate or censored delayed outcome")
	}
	next := *j
	next.entries[i].ready, next.entries[i].y, next.entries[i].authorized = true, y, authorized
	if err := next.drain(); err != nil {
		return err
	}
	*j = next
	return nil
}

// Only an external, frozen expiry policy can censor a missing label. A ready
// label behind that missing prefix remains evidence, never an implicit failure.
func (j *subsetDelayJournal) expireBefore(cutoff int) error {
	if cutoff < j.head || cutoff > j.state.next {
		return fmt.Errorf("invalid delayed expiry")
	}
	next := *j
	for i, e := range next.entries {
		if e.active && !e.ready && e.origin < cutoff {
			next.entries[i].censored = true
		}
	}
	if err := next.drain(); err != nil {
		return err
	}
	*j = next
	return nil
}

func (j *subsetDelayJournal) drain() error {
	for j.head < j.state.next {
		i := j.head % subsetDelayCapacity
		e := j.entries[i]
		if !e.active || e.origin != j.head {
			return fmt.Errorf("delayed journal gap")
		}
		if !e.ready && !e.censored {
			break
		}
		switch {
		case e.censored:
			j.censored++
		case e.generation != j.generation:
			j.stale++
		default:
			issued, wasSplit := j.state.next, j.state.split
			p := e.forecast
			j.state.pending, j.state.next = &p, e.origin
			if err := j.state.observe(e.origin, e.y, e.authorized); err != nil {
				return err
			}
			j.state.next = issued
			j.applied++
			// A pooled-to-local role change also invalidates outstanding advice.
			if j.state.split != wasSplit {
				j.generation++
			}
		}
		j.entries[i] = subsetDelayEntry{}
		j.head++
	}
	return nil
}

func subsetDelayFixture(t testing.TB) (*subsetDelayJournal, subsetState) {
	t.Helper()
	base, err := observationpreserved.Base(false)
	if err != nil {
		t.Fatal(err)
	}
	j := &subsetDelayJournal{state: subsetState{base: base, enabled: true}}
	samples := make([]observation.Sample, 64)
	for i := range samples {
		samples[i] = observation.Sample{Bits: uint16(i * 7 % 512), Outcome: i%2 == 0}
	}
	count, err := observation.Fit(samples)
	if err != nil {
		t.Fatal(err)
	}
	trial := newSubsetTrial(base, 1)
	if err := forestInputFit(trial, samples); err != nil {
		t.Fatal(err)
	}
	if err := j.publish(observationpreserved.Models{Short: count, Pooled: count, Local: count, Version: 1}, trial.model); err != nil {
		t.Fatal(err)
	}
	return j, j.state
}

func TestSubsetDelayImmediateParity(t *testing.T) {
	j, ref := subsetDelayFixture(t)
	for n := 0; n < 192; n++ {
		bits := uint16(n * 13 % 512)
		a, err := j.predict(n, observationexperiment.Frames(bits, "delay"), int64(n))
		if err != nil {
			t.Fatal(err)
		}
		b, err := ref.predict(observationexperiment.Frames(bits, "delay"), j.models, j.model, n, int64(n))
		if err != nil || a != b {
			t.Fatal("issued parity", n, err)
		}
		if err := j.deliver(n, n%3 == 0, n == 80); err != nil {
			t.Fatal(err)
		}
		if err := ref.observe(n, n%3 == 0, n == 80); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(j.state, ref) {
			t.Fatal("state parity", n)
		}
	}
	if j.applied != 192 || j.stale != 0 || j.censored != 0 || j.head != 192 {
		t.Fatal("accounting")
	}
}

func TestSubsetDelayLifecycle(t *testing.T) {
	j, _ := subsetDelayFixture(t)
	issue := func(n int) {
		t.Helper()
		if _, err := j.predict(n, observationexperiment.Frames(uint16(n), "delay"), int64(n)); err != nil {
			t.Fatal(err)
		}
	}
	for n := 0; n < 64; n++ {
		issue(n)
	}
	before := *j
	if _, err := j.predict(64, observationexperiment.Frames(0, "delay"), 0); err == nil || !reflect.DeepEqual(before, *j) {
		t.Fatal("capacity mutation")
	}
	if err := j.deliver(1, true, false); err != nil {
		t.Fatal(err)
	}
	if j.applied != 0 {
		t.Fatal("prefix bypass")
	}
	before = *j
	if err := j.deliver(1, false, false); err == nil || !reflect.DeepEqual(before, *j) {
		t.Fatal("duplicate mutation")
	}
	if err := j.expireBefore(1); err != nil {
		t.Fatal(err)
	}
	if j.censored != 1 || j.applied != 1 || j.head != 2 {
		t.Fatal("ready evidence lost")
	}
	issue(64)
	models := j.models
	models.Version++
	if err := j.publish(models, j.model); err != nil {
		t.Fatal(err)
	}
	weights := j.state.mix
	if err := j.deliver(2, false, true); err != nil {
		t.Fatal(err)
	}
	if j.stale != 1 || j.state.split || j.state.mix != weights {
		t.Fatal("stale advice applied")
	}
	if err := j.expireBefore(65); err != nil {
		t.Fatal(err)
	}
	if j.applied+j.stale+j.censored != 65 || j.head != 65 {
		t.Fatal("settlement")
	}
	before = *j
	if err := j.deliver(0, true, false); err == nil || !reflect.DeepEqual(before, *j) {
		t.Fatal("late censored mutation")
	}
}

func TestSubsetDelaySplitInvalidatesOutstanding(t *testing.T) {
	j, _ := subsetDelayFixture(t)
	for n := 0; n < 3; n++ {
		if _, err := j.predict(n, observationexperiment.Frames(uint16(n), "delay"), 0); err != nil {
			t.Fatal(err)
		}
	}
	if err := j.deliver(1, true, false); err != nil {
		t.Fatal(err)
	}
	if err := j.deliver(0, false, true); err != nil {
		t.Fatal(err)
	}
	if !j.state.split || j.applied != 1 || j.stale != 1 {
		t.Fatal("split generation leak")
	}
	if err := j.deliver(2, true, false); err != nil {
		t.Fatal(err)
	}
	if j.stale != 2 || j.head != 3 {
		t.Fatal("split settlement")
	}
}

func TestSubsetDelayCapturedAdvice(t *testing.T) {
	j, ref := subsetDelayFixture(t)
	var issued [3]subsetPending
	for n := range issued {
		if _, err := j.predict(n, observationexperiment.Frames(uint16(n*113), "captured"), int64(n)); err != nil {
			t.Fatal(err)
		}
		issued[n] = j.entries[n].forecast
	}
	// Arrival order differs from origin order. Reference losses are applied to
	// captured expert vectors, never forecasts recomputed with updated weights.
	for _, n := range []int{2, 1, 0} {
		if err := j.deliver(n, n != 1, false); err != nil {
			t.Fatal(err)
		}
	}
	for n := range issued {
		p := issued[n]
		ref.pending = &p
		if err := ref.observe(n, n != 1, false); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(j.state, ref) {
		t.Fatal("historical advice changed")
	}
	before := *j
	if err := j.publish(j.models, j.model); err == nil || !reflect.DeepEqual(before, *j) {
		t.Fatal("republication mutation")
	}
	if err := j.expireBefore(4); err == nil || !reflect.DeepEqual(before, *j) {
		t.Fatal("future expiry mutation")
	}
}

func BenchmarkSubsetDelayJournal(b *testing.B) {
	j, ref := subsetDelayFixture(b)
	rd := observationexperiment.Frames(173, "bench")
	for _, journal := range []bool{false, true} {
		name := "direct"
		if journal {
			name = "journal"
		}
		b.Run(name, func(b *testing.B) {
			initial := *j
			j := &initial
			ref := ref
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				if journal {
					if _, err := j.predict(n, rd, 0); err != nil {
						b.Fatal(err)
					}
					if err := j.deliver(n, n%2 == 0, false); err != nil {
						b.Fatal(err)
					}
				} else {
					if _, err := ref.predict(rd, j.models, j.model, n, 0); err != nil {
						b.Fatal(err)
					}
					if err := ref.observe(n, n%2 == 0, false); err != nil {
						b.Fatal(err)
					}
				}
			}
		})
	}
}
