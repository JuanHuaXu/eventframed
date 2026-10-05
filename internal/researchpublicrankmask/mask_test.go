package researchpublicrankmask

import (
	"context"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
	"math"
	"reflect"
	"testing"
)

func TestSourceOnlyExactOwnedCopy(t *testing.T) {
	rows := []r.Row{{ID: "a", Base: .3, Features: r.Vector{.1, .2, .3, .4, .5, .6, .7, .8}}, {ID: "b", Base: .8, Features: r.Vector{.9, .8, .7, .6, .5, .4, .3, .2}}}
	out, e := SourceOnly(context.Background(), rows)
	if e != nil || len(out) != 2 {
		t.Fatal(out, e)
	}
	for i, v := range out {
		if v.ID != rows[i].ID || v.Base != rows[i].Base {
			t.Fatal(v)
		}
		for j := 0; j < 6; j++ {
			if v.Features[j] != rows[i].Features[j] {
				t.Fatal(v)
			}
		}
		if v.Features[6] != 0 || v.Features[7] != 0 {
			t.Fatal(v)
		}
	}
	out[0].Features[0] = 0
	out[0].ID = "changed"
	if rows[0].Features[0] != .1 || rows[0].ID != "a" {
		t.Fatal("alias")
	}
}
func TestSourceOnlyNativeInvariance(t *testing.T) {
	a := []r.Row{{ID: "a", Base: .4, Features: r.Vector{.2, .1, .8, 0, 0, 0, 1, .9}}}
	b := append([]r.Row(nil), a...)
	b[0].Features[6] = 0
	b[0].Features[7] = .1
	x, e := SourceOnly(context.Background(), a)
	if e != nil {
		t.Fatal(e)
	}
	y, e := SourceOnly(context.Background(), b)
	if e != nil || !reflect.DeepEqual(x, y) {
		t.Fatal(x, y, e)
	}
}
func TestSourceOnlyInvalidMaskedCueAndCancel(t *testing.T) {
	for _, v := range []float64{math.NaN(), math.Inf(1), -1, 2} {
		rows := []r.Row{{ID: "a", Base: .4}}
		rows[0].Features[6] = v
		if _, e := SourceOnly(context.Background(), rows); e == nil {
			t.Fatal(v)
		}
	}
	c, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := SourceOnly(c, []r.Row{{ID: "a"}}); e == nil {
		t.Fatal("cancel")
	}
	if _, e := SourceOnly(nil, []r.Row{{ID: "a"}}); e == nil {
		t.Fatal("nil")
	}
	if _, e := SourceOnly(context.Background(), nil); e == nil {
		t.Fatal("empty")
	}
}
