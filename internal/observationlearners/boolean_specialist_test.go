package observationlearners

import (
	"math"
	"math/bits"
	"math/rand"
	"sync"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func booleanFixture(n int) []observation.Sample {
	r := rand.New(rand.NewSource(2026119099))
	s := make([]observation.Sample, n)
	for i := range s {
		x := uint16(r.Intn(512))
		y := bits.OnesCount16(x&150)%2 == 1
		if r.Float64() < .05 {
			y = !y
		}
		s[i] = observation.Sample{Bits: x, Outcome: y}
	}
	return s
}

func TestBooleanSpecialistContracts(t *testing.T) {
	s := booleanFixture(64)
	m, err := fitBooleanSpecialist(s)
	if err != nil {
		t.Fatal(err)
	}
	total := 0.
	for mask, w := range m.weights {
		count := 0
		for _, v := range s {
			if (bits.OnesCount16(v.Bits&uint16(mask))%2 == 1) == v.Outcome {
				count++
			}
		}
		want := logBeta(float64(count)+.5, float64(len(s)-count)+.5) - logBeta(.5, .5)
		if math.Abs(want-m.evidence[mask]) > 1e-11 || w < 0 || w > 1 || math.IsNaN(w) {
			t.Fatal("sequence evidence or posterior weight", mask)
		}
		total += w
	}
	if math.Abs(total-1) > 1e-12 {
		t.Fatal("normalization")
	}
	complement := append([]observation.Sample(nil), s...)
	reverse := append([]observation.Sample(nil), s...)
	permuted := append([]observation.Sample(nil), s...)
	perm := [9]uint{5, 2, 8, 0, 4, 7, 1, 6, 3}
	mapBits := func(x uint16) uint16 {
		v := uint16(0)
		for i, j := range perm {
			v |= ((x >> i) & 1) << j
		}
		return v
	}
	for i := range s {
		complement[i].Outcome = !s[i].Outcome
		reverse[i] = s[len(s)-1-i]
		permuted[i].Bits = mapBits(s[i].Bits)
	}
	c, _ := fitBooleanSpecialist(complement)
	r, _ := fitBooleanSpecialist(reverse)
	p, _ := fitBooleanSpecialist(permuted)
	for x, q := range m.predictions {
		if !(q > 0 && q < 1) || math.Abs(q+c.predictions[x]-1) > 1e-12 || math.Abs(q-r.predictions[x]) > 1e-12 || math.Abs(q-p.predictions[mapBits(uint16(x))]) > 1e-12 {
			t.Fatal("prediction symmetry or order", x)
		}
		if math.Abs(m.weights[x]-c.weights[x]) > 1e-12 || math.Abs(m.weights[x]-p.weights[mapBits(uint16(x))]) > 1e-12 {
			t.Fatal("weight symmetry", x)
		}
	}
	for _, bad := range [][]observation.Sample{nil, make([]observation.Sample, 257), {{Bits: 512}}, {{Bits: 65535}}} {
		if _, err := NewBooleanConditional(bad); err == nil {
			t.Fatal("invalid evidence accepted")
		}
	}
	if _, err := NewBooleanConditional(booleanFixture(256)); err != nil {
		t.Fatal("valid cap rejected", err)
	}
}

func TestBooleanSpecialistMarginals(t *testing.T) {
	s := booleanFixture(32)
	fitted, _ := fitBooleanSpecialist(s)
	m, err := NewBooleanConditional(s)
	if err != nil {
		t.Fatal(err)
	}
	// Independent analytic integration: any missing parity bit makes that rule's
	// outcome probability one half under the declared uniform input law.
	for mask := uint16(0); mask < 512; mask++ {
		for values := mask; ; values = (values - 1) & mask {
			want := 0.
			for rule, w := range fitted.weights {
				p := .5
				if uint16(rule)&^mask == 0 {
					p = fitted.agreement[rule]
					if bits.OnesCount16(values&uint16(rule))%2 == 0 {
						p = 1 - p
					}
				}
				want += w * p
			}
			got, err := m.Forecast(mask, values)
			if err != nil || math.Abs(got-want) > 1e-12 {
				t.Fatal("partial law", mask, values, got, want, err)
			}
			if values == 0 {
				break
			}
		}
	}
	before := *m
	for i := range s {
		s[i].Bits, s[i].Outcome = 511, !s[i].Outcome
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for x := uint16(0); x < 512; x++ {
				if _, err := m.Forecast(511, x); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()
	if before != *m {
		t.Fatal("snapshot mutation")
	}
	for _, v := range [][2]uint16{{512, 0}, {1, 2}, {0, 65535}} {
		if _, err := m.Forecast(v[0], v[1]); err == nil {
			t.Fatal("invalid observation accepted")
		}
	}
}

func TestBooleanSpecialistControls(t *testing.T) {
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i / 2), Outcome: i%2 == 1}
	}
	m, _ := NewBooleanConditional(s)
	for x := uint16(0); x < 512; x++ {
		p, err := m.Forecast(511, x)
		if err != nil || math.Abs(p-.5) > 1e-12 {
			t.Fatal("balanced evidence", p, err)
		}
	}
	for i := range s {
		s[i].Outcome = false
	}
	c, _ := NewBooleanConditional(s)
	p, _ := c.Forecast(0, 0)
	if p <= .25 {
		t.Fatal("unseen-coordinate alternatives discarded", p)
	}
	// A constant on an impoverished training support is not evidence of a
	// globally constant rule. Vary all coordinates for the positive control.
	s = booleanFixture(64)
	for i := range s {
		s[i].Outcome = false
	}
	c, _ = NewBooleanConditional(s)
	p, _ = c.Forecast(0, 0)
	if p >= .1 {
		t.Fatal("empty-subset constant not learned", p)
	}
}

var booleanBenchmarkSink *ConditionalForest
var booleanForecastSink float64

func BenchmarkBooleanSpecialistBuild(b *testing.B) {
	for _, n := range []int{16, 64, 256} {
		name := map[int]string{16: "n16", 64: "n64", 256: "n256"}[n]
		b.Run(name, func(b *testing.B) {
			s := booleanFixture(n)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m, err := NewBooleanConditional(s)
				if err != nil {
					b.Fatal(err)
				}
				booleanBenchmarkSink = m
			}
		})
	}
}

func BenchmarkBooleanSpecialistForecast(b *testing.B) {
	m, err := NewBooleanConditional(booleanFixture(64))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, err := m.Forecast(63, uint16(i)&63)
		if err != nil {
			b.Fatal(err)
		}
		booleanForecastSink = p
	}
}
