package observationlearners

import (
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func TestSpikeFiniteOddsCoordinate(t *testing.T) {
	s := make([]observation.Sample, 64)
	xi := make([]float64, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i % 2), Outcome: i%2 == 1}
		xi[i] = 10
	}
	p, err := fastSpikeSetup(s, []uint16{1}, xi)
	if err != nil {
		t.Fatal(err)
	}
	q := spikeFactors{[]float64{spikeLogOdds(.5)}, []float64{0}, []float64{1}}
	got, err := p.sweep(q, .1, 1)
	if err != nil {
		t.Fatal("finite-odds coordinate rejected", err)
	}
	r, err := spikeSetup(s, []uint16{1}, xi)
	if err != nil {
		t.Fatal(err)
	}
	want, err := r.coordinate(q, 0, .1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.logOdds[0]-want.logOdds[0]) > 1e-10 || got.logOdds[0] < 100 {
		t.Fatal("dense extreme odds", got.logOdds, want.logOdds)
	}
	g, h, _, _ := spikeBernoulli(got.logOdds[0])
	if g != 1 || h <= 0 {
		t.Fatal("lost representable exclusion tail")
	}
	b, err := p.bound(got, .1, 1)
	if err != nil || !spikeFinite(b) || math.Abs(b-r.uncollapsed(got, .1, 1)) > 1e-9 {
		t.Fatal("extreme bound", err)
	}
	t.Logf("logOdds=%g exclusion=%g bound=%g", got.logOdds[0], h, b)
}

func TestSpikeStableBernoulli(t *testing.T) {
	for _, odds := range []float64{-1000, -700, -100, -40, -36.745813783531418, -1, 0, 1, 36.745813783531418, 40, 100, 700, 1000} {
		g, h, lg, lh := spikeBernoulli(odds)
		rg, rh, rlg, rlh := spikeBernoulli(-odds)
		if g != rh || h != rg || lg != rlh || lh != rlg || math.Abs(g+h-1) > 3e-16 || math.Abs(lg-lh-odds) > 1e-12 {
			t.Fatal("odds symmetry", odds)
		}
		if math.Abs(odds) <= 700 && (g <= 0 || h <= 0) {
			t.Fatal("representable tail lost", odds)
		}
		q := spikeFactors{[]float64{odds}, []float64{0}, []float64{1}}
		if err := q.check(1, .5, 1); err != nil || !spikeFinite(q.kl(.5, 1)) {
			t.Fatal("finite entropy", odds, err)
		}
		if math.Abs(odds) <= 1 {
			want := g*math.Log(2*g) + h*math.Log(2*h)
			if math.Abs(q.kl(.5, 1)-want) > 1e-15 {
				t.Fatal("ordinary KL changed")
			}
		}
	}
	for _, odds := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		q := spikeFactors{[]float64{odds}, []float64{0}, []float64{1}}
		if q.check(1, .5, 1) == nil {
			t.Fatal("nonfinite odds accepted")
		}
	}
	// A representable tiny tail can materially affect variance for a large slab mean.
	q := spikeFactors{[]float64{40}, []float64{1e8}, []float64{1e-20}}
	_, variance := q.moments(0)
	u := math.Exp(-40)
	want := 1e-20/(1+u) + u/((1+u)*(1+u))*1e16
	if math.Abs(variance-want) > 1e-15 || variance < .04 {
		t.Fatal("tail variance", variance, want)
	}
	g, h, _, _ := spikeBernoulli(40)
	l := spikeLaw{inclusion: []float64{g}, exclusion: []float64{h}, mean: []float64{0}, slabVariance: []float64{100}}
	wantCF := h + g*math.Exp(-50)
	if math.Abs(real(l.characteristic(1))-wantCF) > wantCF*1e-14 {
		t.Fatal("characteristic loses atom")
	}
	l.exclusion[0] = .2
	if _, _, _, err := l.validate(); err == nil {
		t.Fatal("invalid mixture mass")
	}
}

func BenchmarkSpikePilotPriorFit(b *testing.B) {
	s, masks, _, _ := spikeLargeFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := fitSpikeFixed(s, masks, 1.0/255, 1, 1024); err != nil {
			b.Fatal(err)
		}
	}
}

func TestSpikePreparedArithmetic(t *testing.T) {
	s, masks, xi, q := spikeLargeFixture()
	p, err := fastSpikeSetup(s, masks, xi)
	if err != nil {
		t.Fatal(err)
	}
	for cycle := 0; cycle < 3; cycle++ {
		got := p.predictors(q)
		var means, variances [255]float64
		for j := range masks {
			means[j], variances[j] = q.moments(j)
		}
		for i, row := range p.x {
			want := 0.
			for j, x := range row {
				want += x * q.probability(j) * q.mean[j]
			}
			if got[i] != want {
				t.Fatal("predictor arithmetic changed")
			}
			a, b := p.moments(q, row)
			c, d := p.momentsPrepared(row, &means, &variances)
			if a != c || b != d {
				t.Fatal("prepared moments changed")
			}
		}
		q, err = p.sweep(q, .1, 1)
		if err != nil {
			t.Fatal(err)
		}
	}
}
