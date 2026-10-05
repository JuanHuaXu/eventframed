package researchpublichybrid

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sync"
	"testing"
	"time"
)

var at = time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)

func fixture(t testing.TB) *Index {
	t.Helper()
	x, e := New(context.Background(), []Document{{"a", "alpha alpha beta", at}, {"b", "beta gamma", at}, {"c", "gamma", at}}, at)
	if e != nil {
		t.Fatal(e)
	}
	return x
}
func TestHybridFormula(t *testing.T) {
	x := fixture(t)
	h, e := x.Search(context.Background(), "ALPHA alpha", 200)
	if e != nil || len(h) != 1 || h[0].ID != "a" {
		t.Fatal(h, e)
	}
	want := math.Log1p(2.5/1.5) * 2 * 2.2 / (2 + 1.2*(.25+.75*3/2))
	if math.Abs(h[0].Score-want) > 1e-14 {
		t.Fatal(h, want)
	}
	if s := x.Stats(); s.Documents != 3 || s.Tokens != 6 || s.Postings != 5 || s.Terms != 3 {
		t.Fatal(s)
	}
	empty, e := x.Search(context.Background(), "absent", 200)
	if e != nil || len(empty) != 0 {
		t.Fatal(empty, e)
	}
	u, e := New(context.Background(), []Document{{"u", "CAFÉ λ2", at}}, at)
	if e != nil {
		t.Fatal(e)
	}
	h, e = u.Search(context.Background(), "café Λ2", 200)
	if e != nil || len(h) != 1 {
		t.Fatal(h, e)
	}
}
func TestHybridFutureAndOwnership(t *testing.T) {
	d := []Document{{"a", "alpha alpha beta", at}, {"b", "beta gamma", at}, {"c", "gamma", at}, {"future", "alpha alpha alpha", at.Add(time.Second)}}
	x, e := New(context.Background(), d, at)
	if e != nil {
		t.Fatal(e)
	}
	d[0].Text = "corrupt"
	a, _ := x.Search(context.Background(), "alpha", 200)
	b, _ := fixture(t).Search(context.Background(), "alpha", 200)
	if !reflect.DeepEqual(a, b) {
		t.Fatal(a, b)
	}
	a[0].ID = "changed"
	fresh, _ := x.Search(context.Background(), "alpha", 200)
	if fresh[0].ID != "a" {
		t.Fatal(fresh)
	}
	if _, e = x.Fuse(context.Background(), []Hit{{"future", 1}}, nil, 200); e == nil {
		t.Fatal("future admitted")
	}
}
func TestHybridFullFrontierAndFusion(t *testing.T) {
	x := fixture(t)
	a := []Hit{{"c", 100}, {"a", -10}, {"b", 0}}
	h, e := x.Rerank(context.Background(), "alpha", a)
	if e != nil || len(h) != 3 || h[0].ID != "a" || h[1].ID != "b" || h[2].ID != "c" {
		t.Fatal(h, e)
	}
	f, e := x.Fuse(context.Background(), []Hit{{"a", 1}, {"b", 0}}, []Hit{{"b", 100}, {"c", 99}}, 200)
	if e != nil || len(f) != 3 || f[0].ID != "b" || math.Abs(f[0].Score-(1./61+1./62)) > 1e-15 {
		t.Fatal(f, e)
	}
	cut, e := x.Fuse(context.Background(), []Hit{{"a", 0}}, []Hit{{"b", 0}}, 1)
	if e != nil || len(cut) != 1 || cut[0].ID != "a" {
		t.Fatal(cut, e)
	}
	if !reflect.DeepEqual(a, []Hit{{"c", 100}, {"a", -10}, {"b", 0}}) {
		t.Fatal("input changed")
	}
}
func TestHybridRejectAndCancel(t *testing.T) {
	x := fixture(t)
	c, cancel := context.WithCancel(context.Background())
	cancel()
	for _, ctx := range []context.Context{nil, c} {
		if _, e := x.Search(ctx, "alpha", 200); e == nil {
			t.Fatal("context")
		}
		if _, e := x.Fuse(ctx, nil, nil, 200); e == nil {
			t.Fatal("context")
		}
	}
	for _, h := range [][]Hit{{{"unknown", 1}}, {{"a", 1}, {"a", 2}}, {{"a", math.Inf(1)}}, {{"a", math.NaN()}}} {
		if _, e := x.Rerank(context.Background(), "alpha", h); e == nil {
			t.Fatal(h)
		}
		if _, e := x.Fuse(context.Background(), nil, h, 200); e == nil {
			t.Fatal(h)
		}
	}
	for _, q := range []string{"", "!!!", string([]byte{255})} {
		if _, e := x.Search(context.Background(), q, 200); e == nil {
			t.Fatal("query")
		}
	}
	for _, cap := range []int{0, 201} {
		if _, e := x.Search(context.Background(), "alpha", cap); e == nil {
			t.Fatal("cap")
		}
	}
	for _, d := range [][]Document{nil, {{"a", "alpha", at}, {"a", "beta", at}}, {{"a", "", at}}, {{"a", "alpha", time.Time{}}}, {{"a", "alpha", at.Add(time.Second)}}} {
		if _, e := New(context.Background(), d, at); e == nil {
			t.Fatal("corpus")
		}
	}
}
func TestHybridConcurrentDeterminism(t *testing.T) {
	x := fixture(t)
	want, _ := x.Search(context.Background(), "beta gamma", 200)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				h, e := x.Search(context.Background(), "beta gamma", 200)
				if e != nil || !reflect.DeepEqual(h, want) {
					t.Error(h, e)
				}
			}
		}()
	}
	wg.Wait()
}
func BenchmarkHybridSearch(b *testing.B) {
	d := make([]Document, 5183)
	for i := range d {
		d[i] = Document{fmt.Sprintf("doc-%05d", i), fmt.Sprintf("alpha beta variable%d gamma", i), at}
	}
	x, e := New(context.Background(), d, at)
	if e != nil {
		b.Fatal(e)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e := x.Search(context.Background(), "alpha variable20", 200); e != nil {
			b.Fatal(e)
		}
	}
}
