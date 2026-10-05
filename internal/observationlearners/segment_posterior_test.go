package observationlearners

import (
	"math"
	"math/bits"
	"reflect"
	"sync"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// Independent half-integer beta integral for an ordered binary sequence.
func segmentTestBeta(no, yes int) float64 {
	f := func(n int) float64 {
		v := 1.
		for j := 2; j <= n; j++ {
			v *= float64(j)
		}
		return v
	}
	return f(2*no) * f(2*yes) / (math.Pow(4, float64(no+yes)) * f(no) * f(yes) * f(no+yes))
}

func segmentTestMarginal(samples []observation.Sample, mass float64) float64 {
	g, b := 0., 0.
	for mask := 0; mask < 512; mask++ {
		k := bits.OnesCount(uint(mask))
		prior := math.Pow(1./3, float64(k)) * math.Pow(2./3, float64(9-k))
		cells := map[uint16][2]int{}
		agreement := 0
		for _, s := range samples {
			c := cells[s.Bits&uint16(mask)]
			y := 0
			if s.Outcome {
				y = 1
			}
			c[y]++
			cells[s.Bits&uint16(mask)] = c
			if (bits.OnesCount16(s.Bits&uint16(mask))%2 == 1) == s.Outcome {
				agreement++
			}
		}
		v := 1.
		for _, c := range cells {
			v *= segmentTestBeta(c[0], c[1])
		}
		g += prior * v
		b += prior * segmentTestBeta(len(samples)-agreement, agreement)
	}
	return mass*g + (1-mass)*b
}

func TestSegmentLikelihoodReference(t *testing.T) {
	samples := ridgeTestSamples(6)
	for _, mass := range []float64{0, .5, 1} {
		m, err := buildSegmentLikelihoods(samples, mass)
		if err != nil {
			t.Fatal(err)
		}
		for start := 0; start <= len(samples); start++ {
			for end := start; end <= len(samples); end++ {
				want := segmentTestMarginal(samples[start:end], mass)
				if math.Abs(math.Exp(m.logM[start][end])-want) > 1e-13 {
					t.Fatal("marginal", start, end, mass)
				}
			}
			for _, x := range []uint16{0, 1, 17, 73, 511} {
				segment := append([]observation.Sample(nil), samples[start:]...)
				den := segmentTestMarginal(segment, mass)
				segment = append(segment, observation.Sample{Bits: x, Outcome: true})
				want := segmentTestMarginal(segment, mass) / den
				if math.Abs(m.tail[start][x]-want) > 1e-12 {
					t.Fatal("tail predictive", start, x, mass, m.tail[start][x], want)
				}
			}
		}
	}
}

func TestSegmentExistingModelAgreement(t *testing.T) {
	s := ridgeTestSamples(16)
	for _, mass := range []float64{0, .5, 1} {
		m, err := buildSegmentLikelihoods(s, mass)
		if err != nil {
			t.Fatal(err)
		}
		g, err := fitSubset(s)
		if err != nil {
			t.Fatal(err)
		}
		b, err := fitBooleanSpecialist(s)
		if err != nil {
			t.Fatal(err)
		}
		gz, bz := math.Inf(-1), math.Inf(-1)
		for mask := range g.evidence {
			gz = segmentLogAdd(gz, segmentPrior()[mask]+g.evidence[mask])
			bz = segmentLogAdd(bz, segmentPrior()[mask]+b.evidence[mask])
		}
		z := segmentLogAdd(math.Log(mass)+gz, math.Log1p(-mass)+bz)
		wg, wb := math.Exp(math.Log(mass)+gz-z), math.Exp(math.Log1p(-mass)+bz-z)
		if math.Abs(z-m.logM[0][16]) > 1e-12 {
			t.Fatal("evidence agreement")
		}
		for x := range g.predictions {
			if math.Abs(m.tail[0][x]-wg*g.predictions[x]-wb*b.predictions[x]) > 1e-12 {
				t.Fatal("existing predictive", x)
			}
		}
	}
}

func TestSegmentPartitionEnumeration(t *testing.T) {
	// Every cut pattern and availability pattern for a four-frame history.
	// Marginals use factorial integrals, not the implementation's fit builder.
	for seen := 0; seen < 16; seen++ {
		history := []segmentPacket{{Bits: 0, Outcome: true}, {Bits: 1, Outcome: false}, {Bits: 3, Outcome: true}, {Bits: 7, Outcome: true}}
		for i := range history {
			history[i].Arrives = -1
			if seen&(1<<i) != 0 {
				history[i].Arrives = i
			}
		}
		for _, h := range []float64{.01, .5} {
			m, err := fitSegmentPosterior(0, 4, history, 64, h, .5)
			if err != nil {
				t.Fatal(err)
			}
			for _, query := range []uint16{0, 5, 511} {
				z, num := 0., 0.
				var last [4]float64
				for cuts := 0; cuts < 8; cuts++ {
					prior, likelihood := 1., 1.
					start := 0
					tail := 0.
					for end := 0; end < 4; end++ {
						cut := end < 3 && cuts&(1<<end) != 0
						if end < 3 {
							if cut {
								prior *= h
							} else {
								prior *= 1 - h
							}
						}
						if cut || end == 3 {
							var s []observation.Sample
							for j := start; j <= end; j++ {
								if seen&(1<<j) != 0 {
									s = append(s, observation.Sample{Bits: history[j].Bits, Outcome: history[j].Outcome})
								}
							}
							d := segmentTestMarginal(s, .5)
							likelihood *= d
							if end == 3 {
								tail = segmentTestMarginal(append(s, observation.Sample{Bits: query, Outcome: true}), .5) / d
								last[start] += prior * likelihood
							}
							start = end + 1
						}
					}
					z += prior * likelihood
					num += prior * likelihood * tail
				}
				want := h*.5 + (1-h)*num/z
				if math.Abs(m.predictions[query]-want) > 1e-12 || math.Abs(math.Exp(m.logEvidence)-z) > 1e-12 {
					t.Fatal("partition", seen, h, query)
				}
				for i, w := range last {
					if math.Abs(w/z-m.lastStart[i]) > 1e-12 {
						t.Fatal("last-start posterior")
					}
				}
			}
		}
	}
}

func TestSegmentAsOfAndBounds(t *testing.T) {
	history := make([]segmentPacket, 48)
	for i := range history {
		history[i] = segmentPacket{Bits: uint16(i * 7), Outcome: i%3 == 0, Arrives: i + 10}
	}
	before := append([]segmentPacket(nil), history...)
	m, err := fitSegmentPosterior(0, 48, history, 16, .01, .95)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(history, before) {
		t.Fatal("mutated history")
	}
	var want []int
	for i := 23; i <= 38; i++ {
		want = append(want, i)
	}
	if !reflect.DeepEqual(m.origins, want) {
		t.Fatal("eligible cap", m.origins)
	}
	for i := range history {
		if i < 23 || i > 38 {
			history[i].Outcome = !history[i].Outcome
		}
	}
	u, err := fitSegmentPosterior(0, 48, history, 16, .01, .95)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m, u) {
		t.Fatal("excluded outcome leaked")
	}
	// Refitting a snapshot is idempotent, not a second admission of its labels.
	v, err := fitSegmentPosterior(0, 48, history, 16, .01, .95)
	if err != nil || !reflect.DeepEqual(u, v) {
		t.Fatal("duplicate fit")
	}
	history[38].Outcome = !history[38].Outcome
	v, err = fitSegmentPosterior(0, 48, history, 16, .01, .95)
	if err != nil {
		t.Fatal(err)
	}
	if m.predictions == v.predictions {
		t.Fatal("eligible evidence ignored")
	}
	if m.predictions != u.predictions {
		t.Fatal("snapshot alias")
	}
	for _, h := range []float64{0, 1, math.NaN(), math.Inf(1)} {
		if v, e := fitSegmentPosterior(0, 48, history, 16, h, .5); e == nil || v != nil {
			t.Fatal("hazard accepted")
		}
	}
	for _, mass := range []float64{-1, 2, math.NaN()} {
		if v, e := fitSegmentPosterior(0, 48, history, 16, .1, mass); e == nil || v != nil {
			t.Fatal("mass accepted")
		}
	}
	for _, cap := range []int{0, 65} {
		if _, e := fitSegmentPosterior(0, 48, history, cap, .1, .5); e == nil {
			t.Fatal("cap accepted")
		}
	}
	history[0].Bits = 512
	if _, e := fitSegmentPosterior(0, 48, history, 16, .1, .5); e == nil {
		t.Fatal("bits accepted")
	}
	history[0].Bits = 0
	history[30].Arrives = 29
	if _, e := fitSegmentPosterior(0, 48, history, 16, .1, .5); e == nil {
		t.Fatal("early label accepted")
	}
	if _, e := fitSegmentPosterior(-17, 48, history, 16, .1, .5); e == nil {
		t.Fatal("left bound")
	}
	if _, e := fitSegmentPosterior(0, 257, history, 16, .1, .5); e == nil {
		t.Fatal("clock bound")
	}
	if _, e := fitSegmentPosterior(0, 47, history, 16, .1, .5); e == nil {
		t.Fatal("length mismatch")
	}
}

func TestSegmentEmptyAndMaximum(t *testing.T) {
	for _, size := range []int{0, 272} {
		history := make([]segmentPacket, size)
		for i := range history {
			history[i].Arrives = -1
		}
		left, clock := 0, 0
		if size > 0 {
			left, clock = -16, 256
		}
		m, e := fitSegmentPosterior(left, clock, history, 64, .01, .95)
		if e != nil {
			t.Fatal(e)
		}
		if m.logEvidence != 0 || len(m.origins) != 0 {
			t.Fatal("empty evidence")
		}
		for _, p := range m.predictions {
			if p != .5 {
				t.Fatal("prior forecast")
			}
		}
		for i := range history {
			history[i].Arrives = max(0, left+i)
			history[i].Bits = uint16(i % 512)
			history[i].Outcome = i%3 == 0
		}
		m, e = fitSegmentPosterior(left, clock, history, 64, .01, .95)
		if e != nil {
			t.Fatal(e)
		}
		if size > 0 && (len(m.origins) != 64 || m.origins[0] != 192 || len(m.lastStart) != 272) {
			t.Fatal("maximum scope")
		}
		if size > 0 {
			short, e := fitSegmentPosterior(192, 256, history[208:], 64, .01, .95)
			if e != nil {
				t.Fatal(e)
			}
			// With constant hazard and prior-distributed segment parameters,
			// an entirely unobserved prefix cannot change the predictive law.
			for x, p := range m.predictions {
				if math.Abs(p-short.predictions[x]) > 1e-12 {
					t.Fatal("unobserved prefix", x)
				}
			}
		}
	}
}

func TestSegmentConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	var models [8]*segmentPosterior
	var errs [8]error
	history := []segmentPacket{{Bits: 1, Outcome: true, Arrives: 0}, {Bits: 5, Outcome: false, Arrives: 1}}
	for i := range models {
		wg.Add(1)
		go func(i int) { defer wg.Done(); models[i], errs[i] = fitSegmentPosterior(0, 2, history, 64, .01, .95) }(i)
	}
	wg.Wait()
	for i, m := range models {
		if errs[i] != nil || !reflect.DeepEqual(m, models[0]) {
			t.Fatal("concurrent fit", i, errs[i])
		}
	}
}
