package researchdispersion

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func delayedCopy(d *DelayedShape) DelayedShape {
	v := *d
	m := cloneShape(d.model)
	v.model = &m
	v.base = append([]float64(nil), d.base...)
	v.trials = append([]delayedTrial(nil), d.trials...)
	v.issued = append([]uint16(nil), d.issued...)
	return v
}

func delayedBatchCheck(t *testing.T, d *DelayedShape) {
	t.Helper()
	w, q := shapeReference(d.model)
	for z, p := range w {
		near(t, d.model.w[z], p, 3e-10)
	}
	for i, p := range q {
		got, err := d.Predict(i)
		if err != nil {
			t.Fatal(err)
		}
		near(t, got, p, 3e-10)
	}
}

func TestDelayedShapeArrivalSetAndOriginalForecast(t *testing.T) {
	for _, n := range []int{2, 11, 150, 200} {
		for _, reverse := range []bool{false, true} {
			d, err := NewDelayedShape(bases(n), 1, n*16)
			if err != nil {
				t.Fatal(err)
			}
			tickets := make([]DelayedTicket, n*16)
			outcomes := make([]bool, len(tickets))
			rng := rand.New(rand.NewSource(int64(3600 + n)))
			initial := cloneShape(d.model)
			for j := range tickets {
				i := j % n
				tickets[j], err = d.Issue(i, int64(j))
				if err != nil {
					t.Fatal(err)
				}
				outcomes[j] = rng.Float64() < .2+.6*float64(i)/float64(n-1)
			}
			if !reflect.DeepEqual(initial, cloneShape(d.model)) || d.Pending() != len(tickets) {
				t.Fatal("pending nominations changed the learned model")
			}
			order := rng.Perm(len(tickets))
			if reverse {
				for j := range order {
					order[j] = len(order) - j - 1
				}
			}
			different := false
			for j, k := range order {
				current, _ := d.Predict(k % n)
				different = different || math.Abs(current-tickets[k].Forecast()) > .001
				r, err := d.Resolve(tickets[k], outcomes[k], int64(len(tickets)+j))
				if err != nil || r.Member != k%n || r.TrialOrdinal != k/n+1 || r.Epoch != 1 || r.IssuedAt != int64(k) || r.ArrivedAt != int64(len(tickets)+j) || r.Forecast != tickets[k].Forecast() || r.Useful != outcomes[k] {
					t.Fatal("original evidence binding failed", r, err)
				}
				if j == 0 || j == 31 || j == len(order)-1 {
					delayedBatchCheck(t, d)
				}
			}
			if d.Pending() != 0 || !different {
				t.Fatal("missing resolutions or trivial forecast-binding control")
			}
			for i, count := range d.model.n {
				if count != 16 || d.issued[i] != 16 {
					t.Fatal("arrival and issue counts disagree")
				}
			}
		}
	}
}

func TestDelayedShapeRejectedTransitionsAtomic(t *testing.T) {
	d, _ := NewDelayedShape(bases(2), 1, 2)
	other, _ := NewDelayedShape(bases(2), 1, 2)
	a, _ := d.Issue(0, 3)
	b, _ := d.Issue(0, 3)
	before := delayedCopy(d)
	for _, member := range []int{-1, 0, 2} {
		if _, err := d.Issue(member, 4); err == nil {
			t.Fatal("pending cap/invalid member accepted")
		}
	}
	foreign, _ := other.Issue(0, 3)
	for _, ticket := range []DelayedTicket{{}, foreign, {owner: d, epoch: 1, slot: -1}, {owner: d, epoch: 1, slot: len(d.trials)}, {owner: d, epoch: 1, slot: MaxTrials}} {
		if _, err := d.Resolve(ticket, true, 3); err == nil || d.Cancel(ticket, 3) == nil {
			t.Fatal("invalid or unissued ticket accepted")
		}
	}
	if _, err := d.Resolve(a, true, 2); err == nil || d.Cancel(a, 2) == nil || d.BeginEpoch(1, 4) == nil || d.BeginEpoch(2, 2) == nil {
		t.Fatal("backward time/nonincreasing epoch accepted")
	}
	if !reflect.DeepEqual(before, delayedCopy(d)) {
		t.Fatal("invalid transition mutated state")
	}
	// Even an internal test altering the exported copy cannot replace the
	// private issued forecast used by the receipt.
	original := a.Forecast()
	a.q = .001
	r, err := d.Resolve(a, true, 4)
	if err != nil || r.Forecast != original {
		t.Fatal("receipt trusted a mutable ticket forecast")
	}
	if err := d.Cancel(b, 4); err != nil {
		t.Fatal(err)
	}
	before = delayedCopy(d)
	for _, ticket := range []DelayedTicket{a, b} {
		if _, err := d.Resolve(ticket, false, 4); err == nil || d.Cancel(ticket, 4) == nil {
			t.Fatal("resolved/cancelled trial replay accepted")
		}
	}
	if !reflect.DeepEqual(before, delayedCopy(d)) || d.model.n[0] != 1 || d.model.success[0] != 1 || d.Pending() != 0 {
		t.Fatal("replay/cancellation altered evidence")
	}
	c, _ := d.Issue(1, 4)
	if err := d.BeginEpoch(2, 5); err != nil {
		t.Fatal(err)
	}
	before = delayedCopy(d)
	if _, err := d.Resolve(c, true, 5); err == nil || d.Cancel(c, 5) == nil || !reflect.DeepEqual(before, delayedCopy(d)) {
		t.Fatal("old epoch ticket affected new model")
	}
	delayedBatchCheck(t, d)
}

func TestDelayedShapeCapOwnershipAndModelFailure(t *testing.T) {
	for _, config := range []struct {
		epoch uint64
		cap   int
	}{{0, 1}, {1, 0}, {1, 129}} {
		if _, err := NewDelayedShape(bases(2), config.epoch, config.cap); err == nil {
			t.Fatal("invalid constructor accepted")
		}
	}
	base := bases(2)
	d, _ := NewDelayedShape(base, 1, 1)
	base[0] = .25
	for j := 0; j < MaxTrials; j++ {
		a, err := d.Issue(0, int64(j))
		if err != nil || d.Cancel(a, int64(j)) != nil {
			t.Fatal("valid cancelled trial failed", err)
		}
	}
	before := delayedCopy(d)
	if _, err := d.Issue(0, 65); err == nil || !reflect.DeepEqual(before, delayedCopy(d)) || d.model.n[0] != 0 {
		t.Fatal("cancelled trials were reused or learned as negatives")
	}
	if err := d.BeginEpoch(2, 65); err != nil || d.base[0] != .925 {
		t.Fatal("epoch retained caller-owned input")
	}
	a, _ := d.Issue(0, 65)
	d.model.logs[0] = math.Inf(1)
	before = delayedCopy(d)
	if _, err := d.Resolve(a, true, 66); err == nil || !reflect.DeepEqual(before, delayedCopy(d)) {
		t.Fatal("failed model normalization consumed ticket or changed state")
	}
}

func TestDelayedShapeFutureOutcomesCannotChangeEarlierLaws(t *testing.T) {
	for _, cut := range []int{0, 1, 17, 149} {
		for _, flip := range []bool{false, true} {
			d, _ := NewDelayedShape(bases(11), 1, 150)
			e, _ := NewDelayedShape(bases(11), 1, 150)
			a, b := make([]DelayedTicket, 150), make([]DelayedTicket, 150)
			for j := range a {
				a[j], _ = d.Issue(j%11, int64(j))
				b[j], _ = e.Issue(j%11, int64(j))
				if a[j].Forecast() != b[j].Forecast() {
					t.Fatal("unseen outcome affected issue")
				}
			}
			order := rand.New(rand.NewSource(3636)).Perm(150)
			for j, k := range order {
				if j == cut && !reflect.DeepEqual(cloneShape(d.model), cloneShape(e.model)) {
					t.Fatal("future outcome altered earlier learned law")
				}
				y, z := k%3 == 0, k%3 == 0
				if flip && j >= cut {
					z = !z
				}
				if _, err := d.Resolve(a[k], y, int64(150+j)); err != nil {
					t.Fatal(err)
				}
				if _, err := e.Resolve(b[k], z, int64(150+j)); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func BenchmarkDelayedShapeIssue(b *testing.B) {
	d, _ := NewDelayedShape(bases(150), 1, 32)
	b.ReportAllocs()
	for b.Loop() {
		d.issued[0], d.pending = 0, 0
		if _, err := d.Issue(0, 0); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDelayedShapeResolve(b *testing.B) {
	d, _ := NewDelayedShape(bases(150), 1, 32)
	a, _ := d.Issue(0, 1)
	m := cloneShape(d.model)
	b.ReportAllocs()
	for b.Loop() {
		d.model.logs, d.model.w = m.logs, m.w
		d.model.n[0], d.model.success[0] = 0, 0
		copy(d.model.cache[:Hypotheses*ShapeKernels], m.cache[:Hypotheses*ShapeKernels])
		d.trials[0].status, d.pending, d.clock = 1, 1, 1
		if _, err := d.Resolve(a, true, 2); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDelayedShapeNew(b *testing.B) {
	base := bases(150)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := NewDelayedShape(base, 1, 32); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDelayedShapeEpoch(b *testing.B) {
	d, _ := NewDelayedShape(bases(150), 1, 32)
	b.ReportAllocs()
	for b.Loop() {
		if err := d.BeginEpoch(d.epoch+1, d.clock); err != nil {
			b.Fatal(err)
		}
	}
}
