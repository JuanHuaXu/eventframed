package observationlearners

import (
	"errors"
	"math"
	"math/bits"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func jointFixture(uniform, neutral bool) [4]*ConditionalForest {
	var models [4]*ConditionalForest
	masks := [4]uint16{1, 10, 60, 160}
	for j := range models {
		m := new(ConditionalForest)
		for x := uint16(0); x < 512; x++ {
			mass := 1.
			if !uniform {
				mass = float64(1 + x%7)
			}
			p := .05 + .9*float64(bits.OnesCount16(x&masks[j])%2)
			if neutral {
				p = .5
			}
			m.cells[partialIndex(511, x)] = conditionalCell{mass: mass, weighted: mass * p}
		}
		for i := len(m.cells) - 1; i >= 0; i-- {
			v, p := i, 1
			for bit := 0; bit < 9; bit++ {
				if v%3 == 0 {
					a, b := m.cells[i+p], m.cells[i+2*p]
					m.cells[i] = conditionalCell{a.mass + b.mass, a.weighted + b.weighted}
					break
				}
				v /= 3
				p *= 3
			}
		}
		models[j] = m
	}
	return models
}

func TestJointObservationEnumeration(t *testing.T) {
	for _, uniform := range []bool{true, false} {
		models := jointFixture(uniform, false)
		e, err := newObservationExperts(models)
		if err != nil {
			t.Fatal(err)
		}
		w := [5]float64{.1, .2, .3, .15, .25}
		m, err := e.snapshot(w)
		if err != nil {
			t.Fatal(err)
		}
		for mask := uint16(0); mask < 512; mask++ {
			for value := mask; ; value = (value - 1) & mask {
				mass, weighted := 0., 0.
				for x := uint16(0); x < 512; x++ {
					if x&mask != value {
						continue
					}
					c := models[0].cells[partialIndex(511, x)]
					p := .5 * w[0]
					for j := 0; j < 4; j++ {
						cell := models[j].cells[partialIndex(511, x)]
						p += w[j+1] * cell.weighted / cell.mass
					}
					mass += c.mass
					weighted += c.mass * p
				}
				got, err := m.Forecast(mask, value)
				if err != nil || math.Abs(got-weighted/mass) > 1e-12 {
					t.Fatal("literal marginal", mask, value, got, weighted/mass, err)
				}
				if value == 0 {
					break
				}
			}
		}
		for j := 0; j < 5; j++ {
			one := [5]float64{}
			one[j] = 1
			pure, err := e.snapshot(one)
			if err != nil {
				t.Fatal(err)
			}
			for mask := uint16(0); mask < 512; mask++ {
				want := .5
				if j > 0 {
					want, _ = models[j-1].Forecast(mask, mask)
				}
				got, _ := pure.Forecast(mask, mask)
				if math.Abs(got-want) > 1e-12 {
					t.Fatal("one component")
				}
			}
		}
		before, _ := m.Forecast(511, 511)
		*models[0] = ConditionalForest{}
		w = [5]float64{1}
		after, _ := m.Forecast(511, 511)
		if before != after {
			t.Fatal("mutable source leaked into snapshot")
		}
	}
}

func TestJointObservationValidation(t *testing.T) {
	if _, err := newObservationExperts([4]*ConditionalForest{}); err == nil {
		t.Fatal("nil models")
	}
	var zero observationExperts
	if _, err := zero.snapshot([5]float64{1}); err == nil {
		t.Fatal("zero publication accepted")
	}
	m := jointFixture(true, false)
	m[0].cells[0].weighted *= .9
	if _, err := newObservationExperts(m); err == nil {
		t.Fatal("broken marginal accepted")
	}
	m = jointFixture(true, false)
	m[0] = jointFixture(false, false)[0]
	if _, err := newObservationExperts(m); err == nil {
		t.Fatal("different input laws accepted")
	}
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []float64{-1, 0, 1.1, math.NaN(), math.Inf(1)} {
		if _, err := e.snapshot([5]float64{bad}); err == nil {
			t.Fatal("bad weight")
		}
	}
	s, _ := e.snapshot([5]float64{1})
	if _, err := s.Forecast(512, 0); err == nil {
		t.Fatal("mask")
	}
	if _, err := s.Forecast(0, 1); err == nil {
		t.Fatal("hidden value")
	}
}

type jointReader struct {
	x             uint16
	epoch         uint64
	calls, failAt int
	hook          func()
	change        bool
}

func (r *jointReader) Epoch() uint64 { return r.epoch }
func (r *jointReader) Read(v observation.View) (uint16, uint16, error) {
	r.calls++
	if r.hook != nil {
		r.hook()
	}
	if r.failAt == r.calls {
		return 0, 0, errors.New("read failure")
	}
	if r.change {
		r.epoch++
	}
	return v.Mask(), r.x & v.Mask(), nil
}

func compareJointTrace(t *testing.T, a, b observation.Result) {
	t.Helper()
	if a.Cost != b.Cost || a.Observed != b.Observed || a.Stop != b.Stop || len(a.Trace) != len(b.Trace) || math.Abs(a.Probability-b.Probability) > 1e-12 {
		t.Fatal("observer result mismatch")
	}
	for i, x := range a.Trace {
		y := b.Trace[i]
		if x.View != y.View || x.Attempted != y.Attempted || x.Observed != y.Observed || x.Values != y.Values || math.Abs(x.Probability-y.Probability) > 1e-12 {
			t.Fatal("trace mismatch")
		}
	}
}

func TestJointObservationObserverParity(t *testing.T) {
	models := jointFixture(false, false)
	e, err := newObservationExperts(models)
	if err != nil {
		t.Fatal(err)
	}
	m, _ := e.snapshot([5]float64{0, 1})
	for x := uint16(0); x < 512; x++ {
		a, err := RunConditionalObserver(models[0], &jointReader{x: x, epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		b, err := runJointObserver(m, &jointReader{x: x, epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		compareJointTrace(t, a, b)
		last := b.Trace[len(b.Trace)-1]
		hidden := uint16(511) &^ last.Observed
		c, err := runJointObserver(m, &jointReader{x: x ^ hidden, epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		compareJointTrace(t, b, c)
	}
}

func TestJointObservationBinding(t *testing.T) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	s := newRoutedObservationState()
	bank, _ := newBrierBank([]float64{.95, .05 / 3, .05 / 3, .05 / 3}, .001)
	var gate evidenceRouting
	for step := uint64(0); step < 256; step++ {
		before := *s
		m, err := s.preview(step, e)
		if err != nil || *s != before {
			t.Fatal("preview mutation", err)
		}
		mask := uint16(63)
		if step%2 == 0 {
			mask = 511
		}
		value := uint16(step*11) & mask
		var raw [4]float64
		for j := range raw {
			raw[j], _ = e.models[j].Forecast(mask, value)
		}
		bf, err := bank.predict(step, raw[:])
		if err != nil {
			t.Fatal(err)
		}
		want, err := gate.predict(step, raw, bf.Weights)
		if err != nil {
			t.Fatal(err)
		}
		got, err := s.finish(step, m, mask, value)
		if err != nil || got != want {
			t.Fatal("fixed-view parity", err)
		}
		y := step%3 == 0
		if err = s.observe(step, y); err != nil {
			t.Fatal(err)
		}
		if err = bank.observe(step, y); err != nil {
			t.Fatal(err)
		}
		if err = gate.observe(step, y); err != nil {
			t.Fatal(err)
		}
		if s.bank != *bank || s.gate != gate {
			t.Fatal("feedback parity")
		}
		before = *s
		if err = s.observe(step, y); err == nil || *s != before {
			t.Fatal("duplicate mutation")
		}
	}
	// Identical scalar forecasts must not excuse different acquisition weights.
	e, err = newObservationExperts(jointFixture(true, true))
	if err != nil {
		t.Fatal(err)
	}
	s = newRoutedObservationState()
	before := *s
	m, _ := s.preview(0, e)
	m.weights = [5]float64{1}
	if _, err = s.finish(0, m, 1, 0); err == nil || *s != before {
		t.Fatal("unbound mixture accepted")
	}
}

func TestJointObservationReaderAtomicity(t *testing.T) {
	e, err := newObservationExperts(jointFixture(true, true))
	if err != nil {
		t.Fatal(err)
	}
	s := newRoutedObservationState()
	before := *s
	r := &jointReader{epoch: 1, failAt: 2}
	r.hook = func() {
		if _, err := s.preview(0, e); err == nil {
			t.Fatal("reentrant preview")
		}
		if err := s.observe(0, false); err == nil {
			t.Fatal("reentrant feedback")
		}
	}
	if _, err = s.predict(0, e, r, 1); err == nil || !reflect.DeepEqual(before, *s) {
		t.Fatal("failed reader committed")
	}
	if _, err = s.predict(0, e, &jointReader{epoch: 1, change: true}, 1); err == nil || *s != before {
		t.Fatal("epoch mutation")
	}
	got, err := s.predict(0, e, &jointReader{x: 17, epoch: 1}, 1)
	if err != nil || got.Cost > 6 {
		t.Fatal("valid observer", err)
	}
	last := got.Trace[len(got.Trace)-1]
	for j, p := range s.bank.pending.Experts {
		want, _ := e.models[j].Forecast(last.Observed, last.Values)
		if p != want {
			t.Fatal("unobserved forecast binding")
		}
	}
	if err = s.observe(0, true); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkJointObservationSnapshot(b *testing.B) {
	e, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		b.Fatal(err)
	}
	s := newRoutedObservationState()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m, err := s.preview(0, e)
		if err != nil {
			b.Fatal(err)
		}
		p, err := m.Forecast(63, uint16(i)&63)
		if err != nil || p < 0 || p > 1 {
			b.Fatal(err)
		}
	}
}

func BenchmarkJointObservationPublish(b *testing.B) {
	models := jointFixture(true, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := newObservationExperts(models); err != nil {
			b.Fatal(err)
		}
	}
}

// Paired acquisition boundary: immutable models and weights, no refitting or
// feedback. Neutral predictions force both observers through the full budget.
func BenchmarkJointObservationAcquisition(b *testing.B) {
	e, err := newObservationExperts(jointFixture(true, true))
	if err != nil {
		b.Fatal(err)
	}
	m, err := e.snapshot([5]float64{.1, .2, .3, .15, .25})
	if err != nil {
		b.Fatal(err)
	}
	for _, mixed := range []bool{false, true} {
		name := "single"
		if mixed {
			name = "mixture"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r := &jointReader{x: uint16(i) & 511, epoch: 1}
				var got observation.Result
				var err error
				if mixed {
					got, err = runJointObserver(m, r, 1)
				} else {
					got, err = RunConditionalObserver(&e.models[0], r, 1)
				}
				if err != nil || got.Cost != 6 {
					b.Fatal("acquisition", got.Cost, err)
				}
			}
		})
	}
}
