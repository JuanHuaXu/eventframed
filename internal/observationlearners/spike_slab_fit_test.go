package observationlearners

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type spikeFit struct {
	masks   []uint16
	q       spikeFactors
	profile *fastSpikeProfile
	trace   [][3]float64
	motion  []float64
	stop    string
}

func fitSpikeFixed(s []observation.Sample, masks []uint16, pi, c2 float64, limit int) (*spikeFit, error) {
	if limit < 1 || limit > 1024 || len(masks) < 1 || len(masks) > 255 {
		return nil, fmt.Errorf("spike fit budget")
	}
	q := spikeFactors{make([]float64, len(masks)), make([]float64, len(masks)), make([]float64, len(masks))}
	for j := range masks {
		q.logOdds[j], q.variance[j] = spikeLogOdds(pi), c2
	}
	if err := q.check(len(masks), pi, c2); err != nil {
		return nil, err
	}
	xi := makeOnes(len(s))
	p, err := fastSpikeSetup(s, masks, xi)
	if err != nil {
		return nil, err
	}
	f := &spikeFit{masks: append([]uint16(nil), masks...)}
	for iteration := 0; iteration < limit; iteration++ {
		before, err := p.bound(q, pi, c2)
		if err != nil {
			return nil, err
		}
		next, err := p.sweep(q, pi, c2)
		if err != nil {
			return nil, err
		}
		afterSweep, err := p.bound(next, pi, c2)
		if err != nil {
			return nil, err
		}
		newXi := make([]float64, len(s))
		motion := 0.
		// These moments belong only to this newly swept state, never another cycle.
		var means, variances [255]float64
		for j := range masks {
			means[j], variances[j] = next.moments(j)
			motion = math.Max(motion, math.Abs(next.probability(j)-q.probability(j)))
			motion = math.Max(motion, math.Abs(next.mean[j]-q.mean[j])/(1+math.Abs(q.mean[j])))
			motion = math.Max(motion, math.Abs(math.Log(next.variance[j])-math.Log(q.variance[j])))
		}
		for i := range s {
			mu, v := p.momentsPrepared(p.x[i], &means, &variances)
			newXi[i] = math.Sqrt(mu*mu + v)
			motion = math.Max(motion, math.Abs(newXi[i]-xi[i])/(1+xi[i]))
		}
		newProfile, err := fastSpikeSetup(s, masks, newXi)
		if err != nil {
			return nil, err
		}
		after, err := newProfile.bound(next, pi, c2)
		if err != nil {
			return nil, err
		}
		if !spikeFinite(motion) || afterSweep < before-1e-9 || after < afterSweep-1e-9 {
			return nil, fmt.Errorf("spike fit decreased or nonfinite")
		}
		f.trace = append(f.trace, [3]float64{before, afterSweep, after})
		f.motion = append(f.motion, motion)
		q, p, xi = next, newProfile, newXi
		f.q, f.profile = q, p
		if math.Abs(after-before) <= 1e-6 && motion <= 1e-6 {
			f.stop = "bound-and-state"
			return f, nil
		}
	}
	f.stop = "iteration-cap"
	return f, nil
}

func TestSpikeSlabFit(t *testing.T) {
	s, masks, _, _ := spikeLargeFixture()
	for _, d := range []int{1, 9, 255} {
		f, err := fitSpikeFixed(s, masks[:d], .1, 1, 1024)
		if err != nil {
			t.Fatal(err)
		}
		short, err := fitSpikeFixed(s, masks[:d], .1, 1, 4)
		if err != nil {
			t.Fatal(err)
		}
		if len(f.trace) < len(short.trace) || !reflect.DeepEqual(f.trace[:len(short.trace)], short.trace) {
			t.Fatal("fit prefix mismatch")
		}
		last, err := f.profile.bound(f.q, .1, 1)
		if err != nil || last != f.trace[len(f.trace)-1][2] {
			t.Fatal("profile/factor mismatch", err)
		}
		if f.stop == "bound-and-state" && (math.Abs(f.trace[len(f.trace)-1][2]-f.trace[len(f.trace)-1][0]) > 1e-6 || f.motion[len(f.motion)-1] > 1e-6) {
			t.Fatal("false convergence")
		}
		complement := append([]observation.Sample(nil), s...)
		for i := range complement {
			complement[i].Outcome = !complement[i].Outcome
		}
		g, err := fitSpikeFixed(complement, masks[:d], .1, 1, 1024)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(f.trace, g.trace) || len(f.q.mean) != len(g.q.mean) {
			t.Fatal("fit complement trace")
		}
		for j := range f.q.mean {
			if f.q.mean[j] != -g.q.mean[j] || f.q.variance[j] != g.q.variance[j] || f.q.logOdds[j] != g.q.logOdds[j] {
				t.Fatal("fit complement factors")
			}
		}
		t.Logf("features=%d iterations=%d stop=%s bound=%g motion=%g", d, len(f.trace), f.stop, last, f.motion[len(f.motion)-1])
	}
}

func TestSpikeSlabFitDenseTrajectory(t *testing.T) {
	s, masks, _, _ := spikeLargeFixture()
	s = s[:7]
	masks = masks[:9]
	original := append([]observation.Sample(nil), s...)
	f, err := fitSpikeFixed(s, masks, .2, 1, 16)
	if err != nil {
		t.Fatal(err)
	}
	q := spikeFactors{make([]float64, 9), make([]float64, 9), make([]float64, 9)}
	for j := range q.mean {
		q.logOdds[j] = spikeLogOdds(.2)
		q.variance[j] = 1
	}
	xi := makeOnes(7)
	for iteration := range f.trace {
		p, err := spikeSetup(s, masks, xi)
		if err != nil {
			t.Fatal(err)
		}
		before, err := p.bound(q, .2, 1)
		if err != nil {
			t.Fatal(err)
		}
		for j := range masks {
			q, err = p.coordinate(q, j, .2, 1)
			if err != nil {
				t.Fatal(err)
			}
		}
		afterSweep, err := p.bound(q, .2, 1)
		if err != nil {
			t.Fatal(err)
		}
		for i := range xi {
			mu, v := p.moments(q, p.x[i])
			xi[i] = math.Sqrt(mu*mu + v)
		}
		p, err = spikeSetup(s, masks, xi)
		if err != nil {
			t.Fatal(err)
		}
		after, err := p.bound(q, .2, 1)
		if err != nil {
			t.Fatal(err)
		}
		for j, v := range []float64{before, afterSweep, after} {
			if math.Abs(v-f.trace[iteration][j]) > 1e-9 {
				t.Fatal("dense trajectory bound")
			}
		}
	}
	for j := range masks {
		if math.Abs(q.mean[j]-f.q.mean[j]) > 1e-9 || math.Abs(q.probability(j)-f.q.probability(j)) > 1e-9 || math.Abs(q.variance[j]-f.q.variance[j]) > 1e-9 {
			t.Fatal("dense final factors")
		}
	}
	if !reflect.DeepEqual(original, s) {
		t.Fatal("fit mutated evidence")
	}
	for _, limit := range []int{-1, 0, 1025} {
		if _, err := fitSpikeFixed(s, masks, .2, 1, limit); err == nil {
			t.Fatal("bad budget")
		}
	}
	for _, prior := range [][2]float64{{0, 1}, {1, 1}, {.2, 0}, {.2, math.NaN()}} {
		if _, err := fitSpikeFixed(s, masks, prior[0], prior[1], 16); err == nil {
			t.Fatal("bad prior")
		}
	}
}

func BenchmarkSpikeSlabFit(b *testing.B) {
	s, masks, _, _ := spikeLargeFixture()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := fitSpikeFixed(s, masks, .1, 1, 1024); err != nil {
			b.Fatal(err)
		}
	}
}
