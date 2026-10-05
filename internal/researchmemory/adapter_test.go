package researchmemory

import (
	"math"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestFeatureBoundary(t *testing.T) {
	e := model.Event{Who: model.Field{Value: "Voyager 2"}, What: model.Field{Value: "closest approach"}, Where: model.Field{Value: "Jupiter"}, When: model.Field{Value: "1979-07-09"}}
	q := "What is Voyager 2 closest approach date near Jupiter?"
	a, err := Extract(q, e, .8)
	if err != nil {
		t.Fatal(err)
	}
	e.Content = "unrelated full text"
	e.ID = "hidden-answer"
	e.Attributes = map[string]string{"oracle": "true"}
	e.Provenance.SourceEventIDs = []string{"future-label"}
	b, err := Extract(q, e, .8)
	if err != nil || a != b {
		t.Fatal("excluded data influenced features")
	}
	if a&(1<<7) == 0 || a&(1<<8) == 0 {
		t.Fatal("declared features missing")
	}
	e.Who.Value = "Voyager 1"
	b, _ = Extract(q, e, .8)
	if b&(1<<7) != 0 {
		t.Fatal("numeric constraint ignored")
	}
	for _, bad := range []string{strings.Repeat("x", 1025), string([]byte{255}), "the"} {
		if _, err = Extract(bad, e, .8); err == nil {
			t.Fatal("accepted invalid query")
		}
	}
	if _, err = Extract(q, e, math.NaN()); err == nil {
		t.Fatal("invalid baseline")
	}
}
func TestJournalTimingAndDuplicate(t *testing.T) {
	a := New(1, 1)
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	p, e := a.Predict(0, .3, 1, now)
	if e != nil || p.Probability != .3 {
		t.Fatal("warmup changed baseline")
	}
	if a.Feedback(p.ID, true, 1, now.Add(-time.Second)) == nil {
		t.Fatal("accepted early label")
	}
	if a.Feedback(p.ID, true, 2, now) == nil {
		t.Fatal("accepted wrong epoch")
	}
	if n, k := a.Counts(); n != 0 || k != 1 {
		t.Fatal("rejection consumed record")
	}
	if e = a.Feedback(p.ID, true, 1, now.Add(time.Second)); e != nil {
		t.Fatal(e)
	}
	if a.Feedback(p.ID, true, 1, now.Add(time.Second)) == nil {
		t.Fatal("duplicate label")
	}
	if _, e = a.Predict(0, .3, 1, now); e == nil {
		t.Fatal("backdated prediction used future feedback")
	}
}
func TestBoundedPendingAndDiscard(t *testing.T) {
	a := New(1, 1)
	now := time.Now()
	for i := 0; i < 256; i++ {
		if _, e := a.Predict(0, .2, 1, now); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := a.Predict(0, .2, 1, now); e == nil {
		t.Fatal("unbounded journal")
	}
	a.Discard(1)
	if n, k := a.Counts(); n != 0 || k != 255 {
		t.Fatal("discard learned negative")
	}
}
func TestFitAndConcurrentRead(t *testing.T) {
	a := New(1, 3)
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 80; i++ {
		p, e := a.Predict(uint16(i%2), .5, 1, now.Add(time.Duration(i)*time.Second))
		if e != nil {
			t.Fatal(e)
		}
		if e = a.Feedback(p.ID, i%2 == 1, 1, now.Add(time.Duration(i)*time.Second)); e != nil {
			t.Fatal(e)
		}
	}
	if a.short == nil || a.forest.Nodes() > 155 {
		t.Fatal("fit or bound missing")
	}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, e := a.Predict(1, .5, 1, now.Add(100*time.Second))
			if e != nil {
				t.Error(e)
				return
			}
			if p.Probability <= 0 || p.Probability >= 1 {
				t.Error("invalid law")
			}
			a.Discard(p.ID)
		}()
	}
	wg.Wait()
}
