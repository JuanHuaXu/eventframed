package observationlearners

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"os"
	"reflect"
	"sync"
	"testing"
)

func compareRegimeQuery(t *testing.T, a, b regimeQueryValue) float64 {
	t.Helper()
	if a.Origin != b.Origin || len(a.Base) != len(b.Base) {
		t.Fatal("query shape")
	}
	maxError := 0.
	check := func(x, y float64) {
		t.Helper()
		if math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
			t.Fatal("nonfinite comparison")
		}
		e := math.Abs(x - y)
		maxError = math.Max(maxError, e)
		if e > 1e-10 {
			t.Fatalf("batch disagreement %.17g %.17g error%g", x, y, e)
		}
	}
	check(a.Gain, b.Gain)
	for y := 0; y < 2; y++ {
		check(a.Mass[y], b.Mass[y])
		check(a.LogEvidence[y], b.LogEvidence[y])
		if len(a.Conditional[y]) != len(b.Conditional[y]) {
			t.Fatal("conditional shape")
		}
		for p := range a.Base {
			check(a.Conditional[y][p], b.Conditional[y][p])
		}
	}
	for p := range a.Base {
		check(a.Base[p], b.Base[p])
	}
	return maxError
}

// Numerical ties have a declared deterministic origin rule. This is a test
// selection contract, not a claim that insignificant utility gaps are meaningful.
func regimeQueryWinner(values []regimeQueryValue) int {
	best := math.Inf(-1)
	for _, v := range values {
		best = math.Max(best, v.Gain)
	}
	selected := int(^uint(0) >> 1)
	for _, v := range values {
		if best-v.Gain <= 1e-10 && v.Origin < selected {
			selected = v.Origin
		}
	}
	return selected
}

func TestRegimeQueryBatchTiny(t *testing.T) {
	maxError, queries := 0., 0
	for seen := 0; seen < 15; seen++ {
		history := []segmentPacket{{Bits: 0, Outcome: true}, {Bits: 1, Outcome: false}, {Bits: 3, Outcome: true}, {Bits: 7, Outcome: true}}
		var pool []int
		for i := range history {
			history[i].Arrives = -1
			if seen&(1<<i) != 0 {
				history[i].Arrives = i
			} else {
				pool = append(pool, i)
			}
		}
		probes := []uint16{0, 5, 511}
		batch, err := fitRegimeQueryBatch(context.Background(), 0, 4, history, pool, probes)
		if err != nil {
			t.Fatal(err)
		}
		s, err := newRegimeQueryState(context.Background(), 0, 4, history)
		if err != nil {
			t.Fatal(err)
		}
		var reference []regimeQueryValue
		for i, j := range pool {
			r, err := s.value(context.Background(), j, probes)
			if err != nil {
				t.Fatal(err)
			}
			maxError = math.Max(maxError, compareRegimeQuery(t, batch.Values[i], r))
			reference = append(reference, r)
			for p, x := range probes {
				z, base := regimeQueryEnumerate(history, x)
				if math.Abs(base-batch.Values[i].Base[p]) > 1e-12 || math.Abs(math.Log(z)-batch.BaseLogEvidence) > 1e-12 {
					t.Fatal("exhaustive base")
				}
				for y := 0; y < 2; y++ {
					other := append([]segmentPacket(nil), history...)
					other[j].Arrives = 4
					other[j].Outcome = y == 1
					zy, py := regimeQueryEnumerate(other, x)
					if math.Abs(zy/z-batch.Values[i].Mass[y]) > 1e-12 || math.Abs(py-batch.Values[i].Conditional[y][p]) > 1e-12 {
						t.Fatal("exhaustive conditional")
					}
				}
			}
			queries++
		}
		if regimeQueryWinner(batch.Values) != regimeQueryWinner(reference) {
			t.Fatal("tiny query ranking")
		}
	}
	if queries != 32 {
		t.Fatal("tiny count", queries)
	}
	// Same-cell empty-prior joint moment is3/8. Different-cell generic means
	// factor only within each mask; averaging does not restore independence.
	m, err := buildRegimeQueryMoments(context.Background(), nil, []uint16{0}, []uint16{0, 511}, []int{0})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(m.joint[0][0][0]-.375) > 1e-12 {
		t.Fatal("empty same-cell covariance")
	}
	if !(m.joint[0][0][1] > .25 && m.joint[0][0][1] < .375) {
		t.Fatal("empty different-cell mixture")
	}
	if regimeQueryWinner([]regimeQueryValue{{Origin: 9, Gain: .2}, {Origin: 3, Gain: .2}, {Origin: 1, Gain: .1}}) != 3 {
		t.Fatal("exact tie order")
	}
	t.Logf("32 queries/96 probe comparisons, independent partitions, empty moments and tie ranking PASS; max error%g", maxError)
}

func regimeQueryGapHistory() ([]segmentPacket, []int) {
	h := make([]segmentPacket, 96)
	for i := range h {
		h[i] = segmentPacket{Bits: uint16(i * 37 % 512), Outcome: i%5 < 2, Arrives: -1}
		if i%3 != 1 && i != 95 {
			h[i].Arrives = i
		}
	}
	return h, []int{1, 4, 7, 10, 13, 16, 19, 95}
}

func TestRegimeQueryBatchCapacity(t *testing.T) {
	gap, gapPool := regimeQueryGapHistory()
	probes := []uint16{0, 1, 3, 7, 15, 31, 63, 511}
	maxError := 0.
	for _, fixture := range []struct {
		h    []segmentPacket
		pool []int
	}{{regimeQueryCapacityHistory(), []int{72, 73, 74, 75, 76, 77, 78, 79}}, {gap, gapPool}} {
		before := append([]segmentPacket(nil), fixture.h...)
		s, err := newRegimeQueryState(context.Background(), 0, len(fixture.h), fixture.h)
		if err != nil {
			t.Fatal(err)
		}
		if len(s.base.origins) != 63 {
			t.Fatal("fixture support")
		}
		batch, err := fitRegimeQueryBatch(context.Background(), 0, len(fixture.h), fixture.h, fixture.pool, probes)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(batch.BaseLogEvidence-s.base.logEvidence) > 1e-10 {
			t.Fatal("base evidence")
		}
		var reference []regimeQueryValue
		for i, j := range fixture.pool {
			r, err := s.value(context.Background(), j, probes)
			if err != nil {
				t.Fatal(err)
			}
			maxError = math.Max(maxError, compareRegimeQuery(t, batch.Values[i], r))
			reference = append(reference, r)
		}
		if regimeQueryWinner(batch.Values) != regimeQueryWinner(reference) {
			t.Fatal("full query ranking")
		}
		if !reflect.DeepEqual(before, fixture.h) {
			t.Fatal("batch mutated history")
		}
		for i := range fixture.h {
			if fixture.h[i].Arrives < 0 {
				fixture.h[i].Outcome = !fixture.h[i].Outcome
			}
		}
		other, err := fitRegimeQueryBatch(context.Background(), 0, len(fixture.h), fixture.h, fixture.pool, probes)
		if err != nil || !reflect.DeepEqual(batch, other) {
			t.Fatal("hidden outcome leak")
		}
		reversed := append([]int(nil), fixture.pool...)
		for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
			reversed[i], reversed[j] = reversed[j], reversed[i]
		}
		reordered, err := fitRegimeQueryBatch(context.Background(), 0, len(fixture.h), fixture.h, reversed, probes)
		if err != nil {
			t.Fatal(err)
		}
		for i, v := range reordered.Values {
			compareRegimeQuery(t, v, batch.Values[len(batch.Values)-1-i])
		}
		if regimeQueryWinner(batch.Values) != regimeQueryWinner(reordered.Values) {
			t.Fatal("pool order changed choice")
		}
	}
	t.Logf("two full63 histories,16 queries/128 probes, order and ownership PASS; max error%g", maxError)
}

func TestRegimeQueryBatchGuards(t *testing.T) {
	ctx := context.Background()
	h := []segmentPacket{{Bits: 1, Arrives: 0}, {Bits: 2, Arrives: -1}, {Bits: 3, Arrives: -1}}
	for _, pool := range [][]int{nil, make([]int, 9), {0}, {3}, {-1}, {1, 1}} {
		if _, err := fitRegimeQueryBatch(ctx, 0, 3, h, pool, []uint16{0}); err == nil {
			t.Fatal("invalid pool")
		}
	}
	for _, probes := range [][]uint16{nil, make([]uint16, 9), {512}} {
		if _, err := fitRegimeQueryBatch(ctx, 0, 3, h, []int{1}, probes); err == nil {
			t.Fatal("invalid probes")
		}
	}
	full := regimeQueryCapacityHistory()
	full[63].Arrives = 63
	if _, err := fitRegimeQueryBatch(ctx, 0, 80, full, []int{79}, []uint16{0}); err == nil {
		t.Fatal("evicted support")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := fitRegimeQueryBatch(cancelled, 0, 3, h, []int{1}, []uint16{0}); err != context.Canceled {
		t.Fatal("cancellation")
	}
	for _, bad := range []segmentPacket{{Bits: 512, Arrives: -1}, {Arrives: -2}, {Arrives: 288}} {
		if _, err := fitRegimeQueryBatch(ctx, 0, 1, []segmentPacket{bad}, []int{0}, []uint16{0}); err == nil {
			t.Fatal("invalid packet")
		}
	}
	want, err := fitRegimeQueryBatch(ctx, 0, 3, h, []int{1, 2}, []uint16{0, 511})
	if err != nil {
		t.Fatal(err)
	}
	var got [4]*regimeQueryBatch
	var errs [4]error
	var wg sync.WaitGroup
	for i := range got {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got[i], errs[i] = fitRegimeQueryBatch(ctx, 0, 3, h, []int{1, 2}, []uint16{0, 511})
		}(i)
	}
	wg.Wait()
	for i := range got {
		if errs[i] != nil || !reflect.DeepEqual(want, got[i]) {
			t.Fatal("concurrent batch drift")
		}
	}
}

func TestRegimeQueryBatchTape(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit tape required")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var header softV120Artifact
	if err := d.Decode(&header); err != nil {
		t.Fatal(err)
	}
	views, queries, maxError := 0, 0, 0.
	for {
		var input softV120Record
		err := d.Decode(&input)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if input.Index != 0 || input.Schedule != 1 || (input.Case != 0 && input.Case != 19 && input.Case != 20) {
			continue
		}
		s, pool, probes, err := regimeQueryFromTape(context.Background(), input, 160)
		if err != nil {
			t.Fatal(err)
		}
		batch, err := fitRegimeQueryBatch(context.Background(), s.left, s.clock, s.history, pool, probes)
		if err != nil {
			t.Fatal(err)
		}
		var reference []regimeQueryValue
		for i, j := range pool {
			r, err := s.value(context.Background(), j, probes)
			if err != nil {
				t.Fatal(err)
			}
			maxError = math.Max(maxError, compareRegimeQuery(t, batch.Values[i], r))
			reference = append(reference, r)
			queries++
		}
		if regimeQueryWinner(batch.Values) != regimeQueryWinner(reference) {
			t.Fatal("replay query ranking")
		}
		views++
	}
	if views != 6 {
		t.Fatal("replay view count", views)
	}
	t.Logf("six stationary/transition replay views, %d complete-pool queries PASS; max error%g", queries, maxError)
}

var regimeQueryBatchSink *regimeQueryBatch

func BenchmarkRegimeQueryBatchPaired(b *testing.B) {
	ctx := context.Background()
	h := regimeQueryCapacityHistory()
	pool := []int{72, 73, 74, 75, 76, 77, 78, 79}
	probes := []uint16{0, 1, 3, 7, 15, 31, 63, 127}
	b.Run("reference_cold", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			s, err := newRegimeQueryState(ctx, 0, 80, h)
			if err != nil {
				b.Fatal(err)
			}
			for _, j := range pool {
				regimeQuerySink, err = s.value(ctx, j, probes)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
	b.Run("batch_cold", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var err error
			regimeQueryBatchSink, err = fitRegimeQueryBatch(ctx, 0, 80, h, pool, probes)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
