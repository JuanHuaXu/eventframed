package researchpubliclambdarank

import (
	"context"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
	"math"
	"reflect"
	"testing"
)

func TestLambdaDiscountAndSwap(t *testing.T) {
	if Discount(1) != 1 || Discount(11) != 0 || Discount(0) != 0 {
		t.Fatal("discount")
	}
	g := []float64{0, 1}
	before := g[0]*Discount(1) + g[1]*Discount(2)
	after := g[1]*Discount(1) + g[0]*Discount(2)
	if math.Abs(after-before-math.Abs(Discount(1)-Discount(2))) > 1e-15 {
		t.Fatal("swap")
	}
}
func TestLambdaFitFullAndCanceled(t *testing.T) {
	rows := []r.Row{{ID: "a", Base: .4, Features: r.Vector{1, 0}}, {ID: "b", Base: .5, Features: r.Vector{0, 1}}}
	c := []r.Case{{ID: "q", Family: "f", Rows: rows, Positive: map[string]bool{"a": true}}}
	m, s, e := Fit(context.Background(), c)
	if e != nil || s.Pairs != 1 {
		t.Fatal(m, s, e)
	}
	old, _, _ := r.PairGradient(r.Model{}, rows[0], rows[1])
	got, _, _ := r.PairGradient(m, rows[0], rows[1])
	if got >= old {
		t.Fatal(old, got)
	}
	a, e := m.Rank(context.Background(), rows)
	if e != nil || len(a) != 2 {
		t.Fatal(a, e)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, e = Fit(canceled, c); e == nil {
		t.Fatal("cancel")
	}
	if !reflect.DeepEqual(rows, c[0].Rows) {
		t.Fatal("input mutated")
	}
	if _, _, e = Fit(context.Background(), append(c, c[0])); e == nil {
		t.Fatal("duplicate")
	}
}
