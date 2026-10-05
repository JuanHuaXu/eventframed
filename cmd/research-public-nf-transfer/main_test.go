package main

import (
	"context"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
	"math"
	"testing"
)

func TestNFFrozenModel(t *testing.T) {
	if _, e := model(weights{make([]float64, 8)}); e != nil {
		t.Fatal(e)
	}
	for _, n := range []int{0, 7, 9} {
		if _, e := model(weights{make([]float64, n)}); e == nil {
			t.Fatal("accepted dimension", n)
		}
	}
	for _, x := range []float64{math.NaN(), math.Inf(1), 4.01} {
		w := make([]float64, 8)
		w[0] = x
		if _, e := model(weights{w}); e == nil {
			t.Fatal("accepted invalid weight")
		}
	}
	for _, j := range []int{6, 7} {
		w := make([]float64, 8)
		w[j] = .1
		if _, e := model(weights{w}); e == nil {
			t.Fatal("accepted native cue")
		}
	}
}
func TestNFFullShortAndEmptyFrontiers(t *testing.T) {
	for _, n := range []int{0, 1, 50, 200} {
		rows := make([]r.Row, n)
		for j := range rows {
			rows[j] = r.Row{ID: string(rune('a' + j)), Base: .5}
		}
		a, b, e := rank(context.Background(), r.Model{}, rows)
		if e != nil || len(a) != n || len(b) != n {
			t.Fatal(n, e)
		}
	}
	rows := []r.Row{{ID: "a"}, {ID: "a"}}
	if _, _, e := rank(context.Background(), r.Model{}, rows); e == nil {
		t.Fatal("duplicate accepted")
	}
	c, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, e := rank(c, r.Model{}, nil); e == nil {
		t.Fatal("empty cancellation accepted")
	}
}
func TestNFQueryScope(t *testing.T) {
	p := projection{Partition: "nfcorpus-official-test", Queries: make([]query, 323)}
	for i := range p.Queries {
		p.Queries[i] = query{ID: string(rune('a' + i)), Text: "science"}
	}
	if e := checkQueries(p); e != nil {
		t.Fatal(e)
	}
	p.Queries[1] = p.Queries[0]
	if e := checkQueries(p); e == nil {
		t.Fatal("duplicate accepted")
	}
	p.Partition = "fit"
	if e := checkQueries(p); e == nil {
		t.Fatal("wrong partition accepted")
	}
}
