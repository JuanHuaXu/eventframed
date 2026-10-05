package researchmemory

import (
	"testing"
	"time"
)

func TestFrozenIsolationAndNoJournal(t *testing.T) {
	a := New(1, 42)
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 32; i++ {
		p, e := a.Predict(uint16(i), .6, 1, now)
		if e != nil {
			t.Fatal(e)
		}
		if e = a.Feedback(p.ID, i%2 == 0, 1, now); e != nil {
			t.Fatal(e)
		}
	}
	f := a.Freeze()
	want, e := f.Score(3, .6, 1, now)
	if e != nil {
		t.Fatal(e)
	}
	p, e := a.Predict(3, .6, 1, now)
	if e != nil || p.Probability != want {
		t.Fatal("frozen prediction differs", e)
	}
	a.Discard(p.ID)
	for i := 0; i < 32; i++ {
		p, e := a.Predict(3, .6, 1, now)
		if e != nil {
			t.Fatal(e)
		}
		if e = a.Feedback(p.ID, true, 1, now); e != nil {
			t.Fatal(e)
		}
	}
	for i := 0; i < 300; i++ {
		got, e := f.Score(3, .6, 1, now)
		if e != nil || got != want {
			t.Fatal("published model mutated", e)
		}
	}
	labels, pending := a.Counts()
	if labels != 64 || pending != 0 {
		t.Fatal("frozen score changed journal")
	}
	if _, e = f.Score(3, .6, 2, now); e == nil {
		t.Fatal("accepted wrong epoch")
	}
	if _, e = f.Score(3, .6, 1, now.Add(-time.Second)); e == nil {
		t.Fatal("accepted future-trained model")
	}
}
