package researchpublicpairrank

import (
	"context"
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func pair() (Row, Row) { return Row{"a", .4, Vector{1, 0, .5}}, Row{"b", .5, Vector{0, 1, .1}} }
func TestPairRankGradient(t *testing.T) {
	a, b := pair()
	m := Model{Weights: Vector{.2, -.3, .4}}
	_, g, e := PairGradient(m, a, b)
	if e != nil {
		t.Fatal(e)
	}
	for j := range g {
		u, v := m, m
		u.Weights[j] += 1e-6
		v.Weights[j] -= 1e-6
		lu, _, _ := PairGradient(u, a, b)
		lv, _, _ := PairGradient(v, a, b)
		if math.Abs((lu-lv)/2e-6-g[j]) > 1e-8 {
			t.Fatal(j, g)
		}
	}
}
func TestPairRankIdentityBoundAndFullSet(t *testing.T) {
	a, b := pair()
	r, e := (Model{}).Rank(context.Background(), []Row{a, b})
	if e != nil || r[0].ID != "b" || r[0].Score != b.Base || len(r) != 2 {
		t.Fatal(r, e)
	}
	for _, w := range []float64{-4, 4} {
		m := Model{Weights: Vector{w, w, w, w, w, w, w, w}}
		for _, v := range []Row{a, b} {
			s, e := m.Score(v)
			if e != nil || math.Abs(s-v.Base) > .25 {
				t.Fatal(s, e)
			}
		}
	}
	if !reflect.DeepEqual(a, Row{"a", .4, Vector{1, 0, .5}}) {
		t.Fatal("input changed")
	}
}
func TestPairRankFitAndCancel(t *testing.T) {
	a, b := pair()
	cases := []Case{{"q", "family", []Row{a, b}, map[string]bool{"a": true}}, {"missing", "other", []Row{a, b}, map[string]bool{"outside": true}}}
	m, s, e := Fit(context.Background(), cases)
	if e != nil || s.Pairs != 1 || s.NoPairCases != 1 || s.Families != 2 {
		t.Fatal(m, s, e)
	}
	before, _, _ := PairGradient(Model{}, a, b)
	after, _, _ := PairGradient(m, a, b)
	if after >= before {
		t.Fatal(before, after)
	}
	c, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, e = Fit(c, cases); e == nil {
		t.Fatal("cancel")
	}
	if _, e = m.Rank(c, []Row{a, b}); e == nil {
		t.Fatal("cancel")
	}
	for _, v := range []Row{{"", 0, Vector{}}, {"x", math.NaN(), Vector{}}, {"x", 0, Vector{2}}} {
		if _, e = m.Score(v); e == nil {
			t.Fatal(v)
		}
	}
	if _, e = m.Rank(context.Background(), []Row{a, a}); e == nil {
		t.Fatal("duplicate")
	}
	if _, _, e = Fit(context.Background(), append(cases, cases[0])); e == nil {
		t.Fatal("duplicate query")
	}
}
func TestPairRankSourceEpochAndFeatures(t *testing.T) {
	at := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	epoch := strings.Repeat("a", 64)
	s, e := NewSources(context.Background(), []Source{{"doc", "ALPHA beta", "alpha beta 2026", at}}, at, epoch)
	if e != nil {
		t.Fatal(e)
	}
	f, e := s.Features(context.Background(), "alpha beta 2026", "doc", epoch, 1)
	if e != nil || f[0] != 2./3 || f[1] != 1 || f[2] != .5 || f[3] != 1 || f[6] != 1 || f[7] != 60./61 {
		t.Fatal(f, e)
	}
	for _, ep := range []string{"", strings.Repeat("b", 64)} {
		if _, e = s.Features(context.Background(), "alpha", "doc", ep, 0); e == nil {
			t.Fatal("epoch")
		}
	}
	if _, e = s.Features(context.Background(), "alpha", "unknown", epoch, 0); e == nil {
		t.Fatal("id")
	}
	if _, e = NewSources(context.Background(), []Source{{"future", "alpha", "beta", at.Add(time.Second)}}, at, epoch); e == nil {
		t.Fatal("future")
	}
}
func TestPairRankConcurrent(t *testing.T) {
	a, b := pair()
	m := Model{Weights: Vector{.2, .4}}
	want, _ := m.Rank(context.Background(), []Row{a, b})
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 30; j++ {
				got, e := m.Rank(context.Background(), []Row{a, b})
				if e != nil || !reflect.DeepEqual(got, want) {
					t.Error(got, e)
				}
			}
		}()
	}
	wg.Wait()
}
func BenchmarkPairRank200(b *testing.B) {
	r := make([]Row, 200)
	for i := range r {
		r[i] = Row{ID: string(rune(1000 + i)), Base: float64(i) / 200, Features: Vector{.1, .5, .4}}
	}
	m := Model{Weights: Vector{.3, .2, .5}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e := m.Rank(context.Background(), r); e != nil {
			b.Fatal(e)
		}
	}
}
