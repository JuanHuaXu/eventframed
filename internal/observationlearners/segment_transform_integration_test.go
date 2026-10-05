package observationlearners

// Reuse the frozen independent enumeration and lifecycle checks against the candidate.
import (
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"reflect"
	"sync"
	"testing"
)

func TestTransformSegmentPartitionEnumeration(t *testing.T) {
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
			m, err := fitSegmentPosteriorTransform(0, 4, history, 64, h, .5)
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

func TestTransformSegmentAsOfAndBounds(t *testing.T) {
	history := make([]segmentPacket, 48)
	for i := range history {
		history[i] = segmentPacket{Bits: uint16(i * 7), Outcome: i%3 == 0, Arrives: i + 10}
	}
	before := append([]segmentPacket(nil), history...)
	m, err := fitSegmentPosteriorTransform(0, 48, history, 16, .01, .95)
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
	u, err := fitSegmentPosteriorTransform(0, 48, history, 16, .01, .95)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m, u) {
		t.Fatal("excluded outcome leaked")
	}
	// Refitting a snapshot is idempotent, not a second admission of its labels.
	v, err := fitSegmentPosteriorTransform(0, 48, history, 16, .01, .95)
	if err != nil || !reflect.DeepEqual(u, v) {
		t.Fatal("duplicate fit")
	}
	history[38].Outcome = !history[38].Outcome
	v, err = fitSegmentPosteriorTransform(0, 48, history, 16, .01, .95)
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
		if v, e := fitSegmentPosteriorTransform(0, 48, history, 16, h, .5); e == nil || v != nil {
			t.Fatal("hazard accepted")
		}
	}
	for _, mass := range []float64{-1, 2, math.NaN()} {
		if v, e := fitSegmentPosteriorTransform(0, 48, history, 16, .1, mass); e == nil || v != nil {
			t.Fatal("mass accepted")
		}
	}
	for _, cap := range []int{0, 65} {
		if _, e := fitSegmentPosteriorTransform(0, 48, history, cap, .1, .5); e == nil {
			t.Fatal("cap accepted")
		}
	}
	history[0].Bits = 512
	if _, e := fitSegmentPosteriorTransform(0, 48, history, 16, .1, .5); e == nil {
		t.Fatal("bits accepted")
	}
	history[0].Bits = 0
	history[30].Arrives = 29
	if _, e := fitSegmentPosteriorTransform(0, 48, history, 16, .1, .5); e == nil {
		t.Fatal("early label accepted")
	}
	if _, e := fitSegmentPosteriorTransform(-17, 48, history, 16, .1, .5); e == nil {
		t.Fatal("left bound")
	}
	if _, e := fitSegmentPosteriorTransform(0, 257, history, 16, .1, .5); e == nil {
		t.Fatal("clock bound")
	}
	if _, e := fitSegmentPosteriorTransform(0, 47, history, 16, .1, .5); e == nil {
		t.Fatal("length mismatch")
	}
}

func TestTransformSegmentEmptyAndMaximum(t *testing.T) {
	for _, size := range []int{0, 272} {
		history := make([]segmentPacket, size)
		for i := range history {
			history[i].Arrives = -1
		}
		left, clock := 0, 0
		if size > 0 {
			left, clock = -16, 256
		}
		m, e := fitSegmentPosteriorTransform(left, clock, history, 64, .01, .95)
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
		m, e = fitSegmentPosteriorTransform(left, clock, history, 64, .01, .95)
		if e != nil {
			t.Fatal(e)
		}
		if size > 0 && (len(m.origins) != 64 || m.origins[0] != 192 || len(m.lastStart) != 272) {
			t.Fatal("maximum scope")
		}
		if size > 0 {
			short, e := fitSegmentPosteriorTransform(192, 256, history[208:], 64, .01, .95)
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

func TestTransformSegmentConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	var models [8]*segmentPosterior
	var errs [8]error
	history := []segmentPacket{{Bits: 1, Outcome: true, Arrives: 0}, {Bits: 5, Outcome: false, Arrives: 1}}
	for i := range models {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			models[i], errs[i] = fitSegmentPosteriorTransform(0, 2, history, 64, .01, .95)
		}(i)
	}
	wg.Wait()
	for i, m := range models {
		if errs[i] != nil || !reflect.DeepEqual(m, models[0]) {
			t.Fatal("concurrent fit", i, errs[i])
		}
	}
}
