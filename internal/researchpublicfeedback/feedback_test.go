package researchpublicfeedback

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sync"
	"testing"
	"time"
)

func fixture(t *testing.T) *Index {
	t.Helper()
	at := time.Unix(100, 0)
	docs := []Document{}
	for j := 0; j < 20; j++ {
		text := "background"
		if j == 0 {
			text = "needle bridge bridge"
		}
		if j == 1 {
			text = "bridge destination"
		}
		docs = append(docs, Document{ID: fmt.Sprint(j), Text: text, AvailableAt: at})
	}
	x, e := New(context.Background(), docs, at)
	if e != nil {
		t.Fatal(e)
	}
	return x
}
func TestFeedbackExactWeightsAndNomination(t *testing.T) {
	x := fixture(t)
	ctx := context.Background()
	h, e := x.Search(ctx, "needle", 200)
	if e != nil {
		t.Fatal(e)
	}
	w, e := x.Feedback(ctx, "needle", h)
	if e != nil {
		t.Fatal(e)
	}
	want := []Term{{"bridge", 1. / 3}, {"needle", 2. / 3}}
	if len(w) != 2 {
		t.Fatal(w)
	}
	for j := range w {
		if w[j].Text != want[j].Text || math.Abs(w[j].Weight-want[j].Weight) > 1e-14 {
			t.Fatal(w)
		}
	}
	expanded, e := x.SearchWeighted(ctx, w, 200)
	if e != nil || len(expanded) != 2 {
		t.Fatal(e, expanded)
	}
	if _, e = x.Feedback(ctx, "needle", nil); e == nil {
		t.Fatal("lost nominee accepted")
	}
	h[0].Score++
	if _, e = x.Feedback(ctx, "needle", h); e == nil {
		t.Fatal("changed score accepted")
	}
}
func TestFeedbackUnknownFutureAndCancel(t *testing.T) {
	x := fixture(t)
	ctx := context.Background()
	h, e := x.Search(ctx, "unknown", 200)
	if e != nil || len(h) != 0 {
		t.Fatal(e)
	}
	w, e := x.Feedback(ctx, "unknown", h)
	if e != nil || !reflect.DeepEqual(w, []Term{{"unknown", 1}}) {
		t.Fatal(e, w)
	}
	hits, e := x.SearchWeighted(ctx, w, 200)
	if e != nil || len(hits) != 0 {
		t.Fatal(e)
	}
	at := time.Unix(100, 0)
	a, e := New(ctx, []Document{{"a", "needle", at}, {"future", "needle poison", at.Add(time.Second)}}, at)
	if e != nil {
		t.Fatal(e)
	}
	if a.Stats().Documents != 1 || len(a.postings["poison"]) != 0 {
		t.Fatal("future affected corpus")
	}
	for _, v := range [][]Term{{{"bad", math.NaN()}}, {{"needle", .3}}, {{"needle", .5}, {"needle", .5}}, {{"UPPER", 1}}} {
		if _, e = x.SearchWeighted(ctx, v, 200); e == nil {
			t.Fatal("invalid mass/term accepted")
		}
	}
	c, cancel := context.WithCancel(ctx)
	cancel()
	if _, e = x.Feedback(c, "unknown", nil); e == nil {
		t.Fatal("cancel accepted")
	}
	if _, e = x.SearchWeighted(nil, w, 200); e == nil {
		t.Fatal("nil accepted")
	}
}
func TestFeedbackConcurrentOwnedAndBaseline(t *testing.T) {
	x := fixture(t)
	ctx := context.Background()
	h, _ := x.Search(ctx, "needle", 200)
	want, _ := x.Feedback(ctx, "needle", h)
	var wg sync.WaitGroup
	for j := 0; j < 32; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w, e := x.Feedback(ctx, "needle", h)
			if e != nil || !reflect.DeepEqual(w, want) {
				t.Error(e, w)
			}
			w[0].Weight = 99
		}()
	}
	wg.Wait()
	got, _ := x.Feedback(ctx, "needle", h)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("output alias mutated index")
	}
	a, _ := x.Search(ctx, "needle", 200)
	if !reflect.DeepEqual(a, h) {
		t.Fatal("baseline moved")
	}
}
