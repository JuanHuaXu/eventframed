package observationlearners

import (
	"math"
	"math/bits"
	"sync"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func TestFamilyEvidence(t *testing.T) {
	s := booleanFixture(32)
	m, weights, err := fitFamilyConditional(s)
	if err != nil {
		t.Fatal(err)
	}
	// Independent integrals: generic per-assignment rates versus one parity
	// agreement rate. Sequence evidence has no unordered-count coefficient.
	var evidence [2]float64
	for mask := uint16(0); mask < 512; mask++ {
		var counts [512][2]int
		agree := 0
		for _, v := range s {
			c := &counts[v.Bits&mask]
			c[0]++
			if v.Outcome {
				c[1]++
			}
			if (bits.OnesCount16(v.Bits&mask)%2 == 1) == v.Outcome {
				agree++
			}
		}
		generic := 0.
		for _, c := range counts {
			if c[0] > 0 {
				generic += logBeta(float64(c[1])+.5, float64(c[0]-c[1])+.5) - logBeta(.5, .5)
			}
		}
		parity := logBeta(float64(agree)+.5, float64(len(s)-agree)+.5) - logBeta(.5, .5)
		k := float64(bits.OnesCount16(mask))
		prior := math.Pow(1./3, k) * math.Pow(2./3, 9-k)
		evidence[0] += prior * math.Exp(generic)
		evidence[1] += prior * math.Exp(parity)
	}
	if math.Abs(weights[0]+weights[1]-1) > 1e-12 {
		t.Fatal("normalization")
	}
	for i, z := range evidence {
		if math.Abs(weights[i]-z/(evidence[0]+evidence[1])) > 1e-12 {
			t.Fatal("family integral", i)
		}
	}
	var uniform [512]float64
	for i := range uniform {
		uniform[i] = 1
	}
	g, _ := NewSubsetConditional(s, uniform)
	p, _ := NewBooleanConditional(s)
	for mask := uint16(0); mask < 512; mask++ {
		for values := mask; ; values = (values - 1) & mask {
			actual, e := m.Forecast(mask, values)
			gp, _ := g.Forecast(mask, values)
			pp, _ := p.Forecast(mask, values)
			want := weights[0]*gp + weights[1]*pp
			if e != nil || math.Abs(actual-want) > 1e-12 {
				t.Fatal("partial mixture identity", mask, values)
			}
			if values == 0 {
				break
			}
		}
	}
}

func TestFamilyContracts(t *testing.T) {
	s := booleanFixture(64)
	m, w, err := fitFamilyConditional(s)
	if err != nil {
		t.Fatal(err)
	}
	perm := func(x uint16) uint16 { return ((x << 3) | (x >> 6)) & 511 }
	for mode := 0; mode < 3; mode++ {
		x := append([]observation.Sample(nil), s...)
		for i := range x {
			switch mode {
			case 0:
				x[i] = s[len(s)-1-i]
			case 1:
				x[i].Outcome = !x[i].Outcome
			case 2:
				x[i].Bits = perm(x[i].Bits)
			}
		}
		other, ow, e := fitFamilyConditional(x)
		if e != nil || math.Abs(w[0]-ow[0]) > 1e-12 {
			t.Fatal("family weight symmetry", mode, e)
		}
		for bits := uint16(0); bits < 512; bits++ {
			v := bits
			if mode == 2 {
				v = perm(v)
			}
			p, _ := m.Forecast(511, bits)
			q, _ := other.Forecast(511, v)
			if mode == 1 {
				q = 1 - q
			}
			if math.Abs(p-q) > 1e-12 {
				t.Fatal("family prediction symmetry", mode)
			}
		}
	}
	before := *m
	for i := range s {
		s[i].Outcome = !s[i].Outcome
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for x := uint16(0); x < 512; x++ {
				p, e := m.Forecast(511, x)
				if e != nil || p <= 0 || p >= 1 {
					t.Error("forecast bounds", p, e)
				}
			}
		}()
	}
	wg.Wait()
	if before != *m {
		t.Fatal("snapshot mutation")
	}
	for _, bad := range [][]observation.Sample{nil, make([]observation.Sample, 257), {{Bits: 512}}} {
		if _, e := NewFamilyConditional(bad); e == nil {
			t.Fatal("invalid samples accepted")
		}
	}
	if _, e := NewFamilyConditional(booleanFixture(256)); e != nil {
		t.Fatal("valid cap rejected", e)
	}
}

func TestFamilySeedIsolation(t *testing.T) {
	seen := map[int64]bool{}
	for _, base := range []int64{2026119000, 2030119100} {
		for phase := int64(0); phase < 2; phase++ {
			for scenario := int64(0); scenario < 10; scenario++ {
				for index := int64(0); index < 64; index++ {
					for role := int64(0); role < 3; role++ {
						seed := base + phase*1000000 + scenario*10000 + index*10 + role
						if seen[seed] {
							t.Fatal("reused role seed", seed)
						}
						seen[seed] = true
					}
				}
			}
		}
	}
}

func BenchmarkFamilyComparison(b *testing.B) {
	for _, n := range []int{16, 64, 256} {
		for _, arm := range []string{"subset", "family"} {
			name := map[int]string{16: "n16", 64: "n64", 256: "n256"}[n] + "/" + arm
			b.Run(name, func(b *testing.B) {
				s := booleanFixture(n)
				var w [512]float64
				for i := range w {
					w[i] = 1
				}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					var m *ConditionalForest
					var err error
					if arm == "subset" {
						m, err = NewSubsetConditional(s, w)
					} else {
						m, err = NewFamilyConditional(s)
					}
					if err != nil {
						b.Fatal(err)
					}
					booleanBenchmarkSink = m
				}
			})
		}
	}
}

func BenchmarkFamilyForecast(b *testing.B) {
	m, err := NewFamilyConditional(booleanFixture(64))
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
