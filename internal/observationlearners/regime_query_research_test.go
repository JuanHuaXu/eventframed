package observationlearners

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sync"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type regimeQueryState struct {
	left, clock int
	history     []segmentPacket
	base        *segmentPosterior
}

type regimeQueryValue struct {
	Origin            int
	Mass, LogEvidence [2]float64
	Base              []float64
	Conditional       [2][]float64
	Gain              float64
}

// Own a sanitized snapshot with one evidence slot reserved for conditioning.
// A query must add evidence, not change the base sample set through eviction.
func newRegimeQueryState(ctx context.Context, left, clock int, history []segmentPacket) (*regimeQueryState, error) {
	if left < -16 || clock < 0 || clock > 256 || left > clock || len(history) != clock-left {
		return nil, fmt.Errorf("invalid query history")
	}
	owned := append([]segmentPacket(nil), history...)
	known := 0
	for i, p := range owned {
		if p.Bits >= 512 || p.Arrives < -1 || p.Arrives > 287 || (p.Arrives >= 0 && p.Arrives < max(0, left+i)) {
			return nil, fmt.Errorf("invalid query packet")
		}
		if p.Arrives >= 0 && p.Arrives <= clock {
			known++
		} else {
			owned[i].Arrives = -1
			owned[i].Outcome = false
		}
	}
	if known > 63 {
		return nil, fmt.Errorf("conditioning requires reserved evidence slot")
	}
	base, err := fitSegmentPosteriorBatch(ctx, left, clock, owned, 64, .01, .95)
	if err != nil {
		return nil, err
	}
	return &regimeQueryState{left, clock, owned, base}, nil
}

// Conditioning refers to the queried past origin's regime. Reusing its input
// as if it were a current-time label would silently assert a different joint law.
func (s *regimeQueryState) value(ctx context.Context, origin int, probes []uint16) (regimeQueryValue, error) {
	r := regimeQueryValue{Origin: origin}
	if err := ctx.Err(); err != nil {
		return r, err
	}
	if origin < s.left || origin >= s.clock || s.history[origin-s.left].Arrives >= 0 {
		return r, fmt.Errorf("query must be an unknown past origin")
	}
	if len(probes) < 1 || len(probes) > 8 {
		return r, fmt.Errorf("invalid probe count")
	}
	for _, x := range probes {
		if x >= 512 {
			return r, fmt.Errorf("invalid probe")
		}
		r.Base = append(r.Base, .005+.99*s.base.predictions[x])
	}
	for y := 0; y < 2; y++ {
		history := append([]segmentPacket(nil), s.history...)
		history[origin-s.left].Outcome = y == 1
		history[origin-s.left].Arrives = s.clock
		m, err := fitSegmentPosteriorBatch(ctx, s.left, s.clock, history, 64, .01, .95)
		if err != nil {
			return r, err
		}
		if len(m.origins) != len(s.base.origins)+1 {
			return r, fmt.Errorf("conditioning evicted evidence")
		}
		pos := 0
		for _, j := range m.origins {
			if j == origin {
				continue
			}
			if pos >= len(s.base.origins) || s.base.origins[pos] != j {
				return r, fmt.Errorf("conditioning changed base evidence")
			}
			pos++
		}
		if pos != len(s.base.origins) {
			return r, fmt.Errorf("conditioning lost base evidence")
		}
		r.LogEvidence[y] = m.logEvidence
		r.Mass[y] = math.Exp(m.logEvidence - s.base.logEvidence)
		for _, x := range probes {
			r.Conditional[y] = append(r.Conditional[y], .005+.99*m.predictions[x])
		}
	}
	if math.Abs(r.Mass[0]+r.Mass[1]-1) > 1e-10 {
		return r, fmt.Errorf("query marginal incoherent")
	}
	riskReduction := 0.
	for i, p := range r.Base {
		mean, remaining := 0., 0.
		for y := 0; y < 2; y++ {
			q := r.Conditional[y][i]
			mean += r.Mass[y] * q
			remaining += r.Mass[y] * q * (1 - q)
			r.Gain += r.Mass[y] * (q - p) * (q - p) / float64(len(probes))
		}
		if math.Abs(mean-p) > 1e-10 {
			return r, fmt.Errorf("predictive conditioning incoherent")
		}
		riskReduction += (p*(1-p) - remaining) / float64(len(probes))
	}
	if math.IsNaN(r.Gain) || math.Abs(r.Gain-riskReduction) > 1e-10 {
		return r, fmt.Errorf("risk identity incoherent")
	}
	return r, nil
}

// Exhaustive partition sum for tiny histories, with direct Beta integrals.
// The extra hazard transition advances the virtual probe to clock+1.
func regimeQueryEnumerate(history []segmentPacket, probe uint16) (float64, float64) {
	n := len(history)
	z, num := 0., 0.
	for cuts := 0; cuts < 1<<(n-1); cuts++ {
		prior, product, start, tail := 1., 1., 0, 0.
		for end := 0; end < n; end++ {
			cut := end < n-1 && cuts&(1<<end) != 0
			if end < n-1 {
				if cut {
					prior *= .01
				} else {
					prior *= .99
				}
			}
			if cut || end == n-1 {
				var samples []observation.Sample
				for j := start; j <= end; j++ {
					if history[j].Arrives >= 0 {
						samples = append(samples, observation.Sample{Bits: history[j].Bits, Outcome: history[j].Outcome})
					}
				}
				m := segmentTestMarginal(samples, .95)
				product *= m
				if end == n-1 {
					tail = segmentTestMarginal(append(samples, observation.Sample{Bits: probe, Outcome: true}), .95) / m
				}
				start = end + 1
			}
		}
		z += prior * product
		num += prior * product * tail
	}
	return z, .005 + .99*(.005+.99*num/z)
}

func TestRegimeQueryJoint(t *testing.T) {
	checks := 0
	for seen := 0; seen < 15; seen++ {
		history := []segmentPacket{{Bits: 0, Outcome: true}, {Bits: 1, Outcome: false}, {Bits: 3, Outcome: true}, {Bits: 7, Outcome: true}}
		for i := range history {
			history[i].Arrives = -1
			if seen&(1<<i) != 0 {
				history[i].Arrives = i
			}
		}
		before := append([]segmentPacket(nil), history...)
		s, err := newRegimeQueryState(context.Background(), 0, 4, history)
		if err != nil {
			t.Fatal(err)
		}
		for j := range history {
			if history[j].Arrives >= 0 {
				continue
			}
			probes := []uint16{0, 5, 511}
			r, err := s.value(context.Background(), j, probes)
			if err != nil {
				t.Fatal(err)
			}
			for i, x := range probes {
				z, p := regimeQueryEnumerate(history, x)
				if math.Abs(math.Log(z)-s.base.logEvidence) > 1e-12 || math.Abs(p-r.Base[i]) > 1e-12 {
					t.Fatal("base partition")
				}
				for y := 0; y < 2; y++ {
					changed := append([]segmentPacket(nil), history...)
					changed[j].Arrives = 4
					changed[j].Outcome = y == 1
					zy, py := regimeQueryEnumerate(changed, x)
					if math.Abs(r.Mass[y]-zy/z) > 1e-12 || math.Abs(r.Conditional[y][i]-py) > 1e-12 {
						t.Fatal("query joint partition")
					}
				}
				checks++
			}
		}
		if !reflect.DeepEqual(before, history) {
			t.Fatal("caller history changed")
		}
		poison := append([]segmentPacket(nil), history...)
		for i := range poison {
			if poison[i].Arrives < 0 {
				poison[i].Outcome = !poison[i].Outcome
			}
		}
		other, err := newRegimeQueryState(context.Background(), 0, 4, poison)
		if err != nil || !reflect.DeepEqual(other, s) {
			t.Fatal("hidden outcomes in state")
		}
	}
	if checks != 96 {
		t.Fatal("enumeration count", checks)
	}
	t.Log("96 probe-query checks against exhaustive partitions; total probability/expectation and Brier identity PASS")
}

func regimeQueryCapacityHistory() []segmentPacket {
	h := make([]segmentPacket, 80)
	for i := range h {
		h[i] = segmentPacket{Bits: uint16(i * 17 % 512), Outcome: i%3 == 0, Arrives: -1}
		if i < 63 {
			h[i].Arrives = i
		}
	}
	return h
}

func TestRegimeQueryCapacityAndOwnership(t *testing.T) {
	h := regimeQueryCapacityHistory()
	s, err := newRegimeQueryState(context.Background(), 0, 80, h)
	if err != nil {
		t.Fatal(err)
	}
	before := append([]segmentPacket(nil), s.history...)
	r, err := s.value(context.Background(), 79, []uint16{0, 17, 511})
	if err != nil {
		t.Fatal(err)
	}
	if !(r.Gain >= 0 && r.Gain <= .25) {
		t.Fatal("gain bounds")
	}
	if !reflect.DeepEqual(s.history, before) {
		t.Fatal("shared snapshot mutated")
	}
	for i := range h {
		h[i].Bits = 511
		h[i].Outcome = !h[i].Outcome
	}
	again, err := s.value(context.Background(), 79, []uint16{0, 17, 511})
	if err != nil || !reflect.DeepEqual(r, again) {
		t.Fatal("caller mutation reached owned model")
	}
	full := regimeQueryCapacityHistory()
	full[63].Arrives = 63
	if _, err := newRegimeQueryState(context.Background(), 0, 80, full); err == nil {
		t.Fatal("silent eviction")
	}
	for _, j := range []int{-1, 0, 80} {
		if _, err := s.value(context.Background(), j, []uint16{0}); err == nil {
			t.Fatal("invalid origin")
		}
	}
	for _, probes := range [][]uint16{nil, make([]uint16, 9), {512}} {
		if _, err := s.value(context.Background(), 79, probes); err == nil {
			t.Fatal("invalid probes")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.value(ctx, 79, []uint16{0}); err != context.Canceled {
		t.Fatal("query cancellation")
	}
	if _, err := newRegimeQueryState(ctx, 0, 80, regimeQueryCapacityHistory()); err != context.Canceled {
		t.Fatal("base cancellation")
	}
	for _, bad := range []segmentPacket{{Bits: 512, Arrives: -1}, {Arrives: -2}, {Arrives: 288}} {
		if _, err := newRegimeQueryState(context.Background(), 0, 1, []segmentPacket{bad}); err == nil {
			t.Fatal("invalid packet")
		}
	}
	t.Log("63+1 evidence reservation, no eviction, caller ownership, invalid inputs and cancellation PASS")
}

func TestRegimeQueryConcurrent(t *testing.T) {
	h := []segmentPacket{{Bits: 0, Outcome: true, Arrives: 0}, {Bits: 1, Arrives: -1}, {Bits: 7, Arrives: -1}}
	s, err := newRegimeQueryState(context.Background(), 0, 3, h)
	if err != nil {
		t.Fatal(err)
	}
	want, err := s.value(context.Background(), 2, []uint16{0, 1})
	if err != nil {
		t.Fatal(err)
	}
	var got [4]regimeQueryValue
	var errs [4]error
	var wg sync.WaitGroup
	for i := range got {
		wg.Add(1)
		go func(i int) { defer wg.Done(); got[i], errs[i] = s.value(context.Background(), 2, []uint16{0, 1}) }(i)
	}
	wg.Wait()
	for i := range got {
		if errs[i] != nil || !reflect.DeepEqual(want, got[i]) {
			t.Fatal("concurrent query drift")
		}
	}
}

var regimeQuerySink regimeQueryValue
var regimeQueryStateSink *regimeQueryState

func BenchmarkRegimeQuery(b *testing.B) {
	h := regimeQueryCapacityHistory()
	ctx := context.Background()
	probes := []uint16{0, 1, 3, 7, 15, 31, 63, 127}
	s, err := newRegimeQueryState(ctx, 0, 80, h)
	if err != nil {
		b.Fatal(err)
	}
	b.Run("base63", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var err error
			regimeQueryStateSink, err = newRegimeQueryState(ctx, 0, 80, h)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("query_two_refits", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var err error
			regimeQuerySink, err = s.value(ctx, 79, probes)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("pool_eight_queries", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for j := 72; j < 80; j++ {
				var err error
				regimeQuerySink, err = s.value(ctx, j, probes)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}
