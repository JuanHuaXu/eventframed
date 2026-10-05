package observationlearners

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func TestCadenceBaselineMatchesV5(t *testing.T) {
	for _, j := range []int{0, 2, 7, 8} {
		b, e := Base(Scenarios[j], j, 0)
		if e != nil {
			t.Fatal(e)
		}
		old, e := RunStream(b, "unit", j, 0, 0, 2026092701)
		if e != nil {
			t.Fatal(e)
		}
		got, e := RunCadenceStream(b, "unit", j, 0, 0, 2026092701)
		if e != nil {
			t.Fatal(e)
		}
		for i, tick := range got.Ticks {
			if !reflect.DeepEqual(tick.Delivered, old.Ticks[i].Delivered) || tick.Outcome != old.Ticks[i].Outcome || tick.Audit != old.Ticks[i].Audit || tick.Missing != old.Ticks[i].Missing {
				t.Fatal("generator or availability changed")
			}
			if tick.Predictions[0] != old.Ticks[i].Predictions[3] || tick.Predictions[1] != old.Ticks[i].Predictions[5] {
				t.Fatal("matched control changed")
			}
		}
		if got.TreeUpdates[0] != got.Audits || got.TreeUpdates[1] > got.Audits || got.Audits-got.TreeUpdates[1] >= 32 {
			t.Fatal("audit accounting")
		}
		for _, n := range got.TreeNodes {
			if n > 155 {
				t.Fatal("node cap")
			}
		}
	}
}
func TestBatchedForestIdentity(t *testing.T) {
	a, b := NewForest(90), NewForest(90)
	r := rand.New(rand.NewSource(91))
	var pending []observation.Sample
	for i := 1; i <= 256; i++ {
		x := uint16(r.Intn(512))
		y := r.Intn(2) == 1
		a.Update(x, y)
		pending = append(pending, observation.Sample{Bits: x, Outcome: y})
		if i < 32 || i%16 != 0 {
			continue
		}
		for _, p := range pending {
			b.Update(p.Bits, p.Outcome)
		}
		pending = nil
		if !reflect.DeepEqual(a.Trees, b.Trees) {
			t.Fatal("batch changed training semantics")
		}
		for x := uint16(0); x < 512; x++ {
			if a.Predict(x) != b.Predict(x) {
				t.Fatal("batch forecast drift")
			}
		}
	}
}
