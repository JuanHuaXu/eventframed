package observationgate

import (
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

func TestFixedMatchesOriginal(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for stream := 0; stream < 64; stream++ {
		var g Gate
		var old observationpreserved.Evidence
		for i := 0; i < 512; i++ {
			ref, live := r.Float64() < .9, r.Float64() < .5
			d := 0.
			if ref {
				d++
			}
			if live {
				d--
			}
			a, e := g.Observe(d)
			if e != nil || a[0] != old.Observe(ref, live) {
				t.Fatal("baseline drift")
			}
		}
	}
}
func TestPredictableAndInvalid(t *testing.T) {
	var g Gate
	_, e := g.Observe(1)
	if e != nil {
		t.Fatal(e)
	}
	want := math.Log1p(.05 * .85)
	if math.Abs(g.Starts[0][0].Adaptive-want) > 1e-14 {
		t.Fatal("bet used revealing observation")
	}
	before := g
	for _, d := range []float64{2, -2, math.NaN(), math.Inf(1)} {
		if _, e = g.Observe(d); e == nil || !reflect.DeepEqual(g, before) {
			t.Fatal("invalid input changed state")
		}
	}
}
func TestConditionalNullFactors(t *testing.T) {
	// Extreme mean-.15 boundary distributions and all allowed rates. This
	// checks factor positivity and the algebra used in the written proof.
	for _, rate := range []float64{0, .05, .25, .8} {
		for _, mean := range []float64{-.15, 0, .15} {
			for _, sign := range []float64{-1, 1} {
				if 1+rate*(-1-.15) < 0 || 1+rate*(sign*mean-.15) > 1+1e-14 {
					t.Fatal("invalid null factor")
				}
			}
		}
	}
}
func TestNoPrematureStartAndLatch(t *testing.T) {
	var g Gate
	for i := 0; i < 64; i++ {
		g.Observe(1)
	}
	if g.Starts[1] != [2]signed{} {
		t.Fatal("early restart")
	}
	for i := 64; i < 512; i++ {
		g.Observe(1)
	}
	if g.Alert != [4]bool{true, true, true, true} {
		t.Fatal("positive control failed")
	}
	for i := 0; i < 512; i++ {
		g.Observe(0)
	}
	if g.Alert != [4]bool{true, true, true, true} {
		t.Fatal("revocation forgot evidence")
	}
}

var benchmarkGate [4]bool

func BenchmarkObserve(b *testing.B) {
	var g Gate
	for i := 0; i < 512; i++ {
		g.Observe(0)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkGate, _ = g.Observe(float64(i%3 - 1))
	}
}
