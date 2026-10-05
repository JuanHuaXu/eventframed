package observationlearners

import (
	"compress/gzip"
	"encoding/json"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"os"
	"reflect"
	"testing"
)

func TestSubsetArtifactReplay(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SUBSET_ARTIFACT")
	if path == "" {
		t.Skip("explicit artifact required")
	}
	f, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	var a struct{ Records []AcquisitionRecord }
	if e := json.NewDecoder(z).Decode(&a); e != nil {
		t.Fatal(e)
	}
	b, e := RunSubset()
	if e != nil {
		t.Fatal(e)
	}
	if len(a.Records) != len(b) {
		t.Fatal("record count")
	}
	for i := range b {
		a.Records[i].FitNS, a.Records[i].TreeNS, b[i].FitNS, b[i].TreeNS = 0, 0, 0, 0
		if !reflect.DeepEqual(a.Records[i], b[i]) {
			t.Fatalf("replay%d", i)
		}
	}
}

var subsetBenchmarkSink *ConditionalForest

func BenchmarkSubsetBuild(b *testing.B) {
	samples := make([]observation.Sample, 64)
	var w [512]float64
	for i := range samples {
		samples[i] = observation.Sample{Bits: uint16(i), Outcome: i%3 == 0}
	}
	for i := range w {
		w[i] = 1
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m, e := NewSubsetConditional(samples, w)
		if e != nil {
			b.Fatal(e)
		}
		subsetBenchmarkSink = m
	}
}

func TestSubsetRunnerControl(t *testing.T) {
	base, e := Base(Scenarios[7], 7, 0)
	if e != nil {
		t.Fatal(e)
	}
	a, e := RunAcquisitionStream(base, "design", 7, 0, 0, 2026106202, 2, 0)
	if e != nil {
		t.Fatal(e)
	}
	b, e := RunSubsetStream(base, "design", 7, 0, 0, 2026106202, 2, 0)
	if e != nil {
		t.Fatal(e)
	}
	a.FitNS, a.TreeNS, b.FitNS, b.TreeNS = 0, 0, 0, 0
	a.Mode = b.Mode
	if !reflect.DeepEqual(a, b) {
		t.Fatal("forest control changed")
	}
	c, e := RunSubsetStream(base, "design", 7, 0, 0, 2026106202, 2, 1)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(a.Inputs, c.Inputs) {
		t.Fatal("unpaired inputs")
	}
	for step, tick := range c.Ticks {
		if tick.Outcome != a.Ticks[step].Outcome || tick.Audit != a.Ticks[step].Audit || !reflect.DeepEqual(tick.Delivered, a.Ticks[step].Delivered) {
			t.Fatal("unpaired feedback")
		}
		if tick.Predictions[0] != a.Ticks[step].Predictions[0] {
			t.Fatal("fixed control changed")
		}
		for _, origin := range tick.Delivered {
			if origin > step-16 {
				t.Fatal("early delayed label")
			}
		}
	}
}

func logBeta(a, b float64) float64 {
	x, _ := math.Lgamma(a)
	y, _ := math.Lgamma(b)
	z, _ := math.Lgamma(a + b)
	return x + y - z
}
func TestSubsetEvidence(t *testing.T) {
	s := []observation.Sample{{Bits: 0, Outcome: true}, {Bits: 3, Outcome: false}, {Bits: 1, Outcome: true}, {Bits: 0, Outcome: false}, {Bits: 511, Outcome: true}}
	m, e := fitSubset(s)
	if e != nil {
		t.Fatal(e)
	}
	for mask := uint16(0); mask < 512; mask++ {
		var c [512]subsetCell
		for _, v := range s {
			c[v.Bits&mask].n++
			if v.Outcome {
				c[v.Bits&mask].yes++
			}
		}
		evidence := 0.
		for _, v := range c {
			if v.n > 0 {
				evidence += logBeta(float64(v.yes)+.5, float64(v.n-v.yes)+.5) - logBeta(.5, .5)
			}
		}
		if math.Abs(evidence-m.evidence[mask]) > 1e-12 {
			t.Fatal("Beta evidence", mask)
		}
	}
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	r, e := fitSubset(s)
	if e != nil {
		t.Fatal(e)
	}
	total := 0.
	for i, w := range m.weights {
		total += w
		if math.Abs(w-r.weights[i]) > 1e-12 || math.Abs(m.predictions[i]-r.predictions[i]) > 1e-12 {
			t.Fatal("order changed posterior")
		}
	}
	if math.Abs(total-1) > 1e-12 {
		t.Fatal("weight normalization")
	}
	if _, e := fitSubset(nil); e == nil {
		t.Fatal("empty accepted")
	}
	if _, e := fitSubset([]observation.Sample{{Bits: 512}}); e == nil {
		t.Fatal("invalid bits")
	}
}

func TestSubsetConditional(t *testing.T) {
	s := []observation.Sample{{Bits: 0, Outcome: true}, {Bits: 1, Outcome: false}}
	var w [512]float64
	for x := range w {
		w[x] = 1
	}
	f, e := fitSubset(s)
	if e != nil {
		t.Fatal(e)
	}
	m, e := NewSubsetConditional(s, w)
	if e != nil {
		t.Fatal(e)
	}
	for mask := uint16(0); mask < 512; mask++ {
		sum, n := 0., 0.
		for x, p := range f.predictions {
			if uint16(x)&mask == 0 {
				sum += p
				n++
			}
		}
		p, e := m.Forecast(mask, 0)
		if e != nil || math.Abs(p-sum/n) > 1e-12 {
			t.Fatal("partial integral")
		}
	}
	before, _ := m.Forecast(0, 0)
	s[0].Outcome = false
	w[0] = 1e5
	after, _ := m.Forecast(0, 0)
	if before != after {
		t.Fatal("snapshot changed")
	}
}
