package main

import (
	"context"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
	"math"
	"testing"
)

func TestFeedbackFrozenModels(t *testing.T) {
	v := make([]frozen, 5)
	for i := range v {
		v[i] = frozen{Fold: i, Weights: make([]float64, 8)}
	}
	if _, e := models(v); e != nil {
		t.Fatal(e)
	}
	v[0].Weights[7] = .1
	if _, e := models(v); e == nil {
		t.Fatal("native channel accepted")
	}
	v[0].Weights[7] = 0
	v[0].Weights[0] = math.Inf(1)
	if _, e := models(v); e == nil {
		t.Fatal("nonfinite accepted")
	}
	if _, e := models(v[:4]); e == nil {
		t.Fatal("missing fold accepted")
	}
}
func TestFeedbackRankEmptyAndCancel(t *testing.T) {
	v, e := rank(context.Background(), r.Model{}, nil)
	if e != nil || len(v) != 0 {
		t.Fatal(e)
	}
	c, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e = rank(c, r.Model{}, nil); e == nil {
		t.Fatal("cancel accepted")
	}
}
