package researchwindowjournal

import (
	"reflect"
	"testing"
)

func TestJournalChildAuthorityAndCanonicalMetadata(t *testing.T) {
	b, e := NewBank([]float64{.3, .7}, 1, 32, 1, [3]int{1, 4, 12})
	if e != nil {
		t.Fatal(e)
	}
	for n := 0; n < 8; n++ {
		x, e := b.Issue(n%2, int64(n))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = b.Resolve(x, n < 4, int64(n)); e != nil {
			t.Fatal(e)
		}
	}
	x, e := b.Issue(0, 8)
	if e != nil {
		t.Fatal(e)
	}
	row := b.models[0].trials[x.base[0].slot]
	near(t, row.clean, x.Forecast())
	near(t, row.observed, b.selector.members[0].rows[x.selector.ordinal].issuedFirst)
	if _, e = b.Resolve(x, true, 9); e != nil {
		t.Fatal(e)
	}
	for k, m := range b.models {
		before := bankSnapshot(b)
		_, a := m.Issue(1, 10)
		_, z := m.Resolve(x.base[k], false, 10)
		c := m.Cancel(x.base[k], 10)
		_, d := m.RequestAudit(x.base[k], 10)
		e := m.BeginEpoch(2, 10)
		if a == nil || z == nil || c == nil || d == nil || e == nil {
			t.Fatal("child mutation authority", k)
		}
		if !reflect.DeepEqual(before, bankSnapshot(b)) {
			t.Fatal("child mutated shared or private state", k)
		}
		if _, _, e := m.Predict(0); e != nil {
			t.Fatal(e)
		}
		if _, e := m.LatentJoint(0); e != nil {
			t.Fatal(e)
		}
		if !reflect.DeepEqual(before, bankSnapshot(b)) {
			t.Fatal("readonly mutated state", k)
		}
	}
	// Standalone models retain their independent mutating API.
	m, e := New([]float64{.3, .7}, 1, 32, Config{Depth: 1, Window: 4})
	if e != nil {
		t.Fatal(e)
	}
	y, e := m.Issue(0, 0)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.Resolve(y, true, 1); e != nil {
		t.Fatal(e)
	}
	z, e := m.RequestAudit(y, 2)
	if e != nil {
		t.Fatal(e)
	}
	if e = m.Cancel(z, 3); e != nil {
		t.Fatal(e)
	}
	if e = m.BeginEpoch(2, 4); e != nil {
		t.Fatal(e)
	}
}

func TestJournalCorruptionRejectsWithoutPublication(t *testing.T) {
	b, _ := NewBank([]float64{.3, .7}, 1, 32, 1, [3]int{1, 4, 12})
	x, _ := b.Issue(0, 0)
	b.models[0].trials[x.base[0].slot].first = 3
	before := bankSnapshot(b)
	if _, e := b.Resolve(x, true, 1); e == nil {
		t.Fatal("corrupt journal accepted")
	}
	if !reflect.DeepEqual(before, bankSnapshot(b)) {
		t.Fatal("partial corruption repair")
	}
	b, _ = NewBank([]float64{.3, .7}, 1, 32, 1, [3]int{1, 4, 12})
	b.models[2].issued[0]++
	before = bankSnapshot(b)
	if _, e := b.Issue(0, 0); e == nil {
		t.Fatal("desynchronized issue accepted")
	}
	if !reflect.DeepEqual(before, bankSnapshot(b)) {
		t.Fatal("partial synchronized issue")
	}
}
